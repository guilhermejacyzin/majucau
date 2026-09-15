# ERD v1 - Majucau Financial Intelligence

- **Status:** Proposto para aprovação
- **Banco:** PostgreSQL local dedicado
- **Objetivo:** modelar integração, RAW, normalização, finanças, planejamento, contabilidade, auditoria e credenciais sem perder rastreabilidade.

## 1. Princípios do modelo

1. RAW é append-only para sincronização e regras: uma nova versão do payload cria novo registro; não há sobrescrita silenciosa. A única mutação permitida é redação LGPD autorizada, que substitui PII por tombstone auditável e preserva o hash anterior.
2. Todo registro tratado aponta para um ou mais registros RAW por tabelas de contribuição tipadas, com FKs reais, e para a versão da regra aplicada.
3. `source_system + source_entity + source_id + payload_hash` evita repetição do mesmo estado externo.
4. Realizado, provisório e projetado não compartilham significado implícito.
5. Competência, vencimento, previsão e liquidação possuem datas distintas.
6. Valores monetários são decimais exatos.
7. Snapshots e forecasts são versionados; histórico não é sobrescrito.
8. Credenciais não são armazenadas no banco em texto aberto. O banco mantém somente metadados e um `secret_ref` opaco para o cofre DPAPI.

## 2. Visão por camadas

```mermaid
flowchart TB
  subgraph UI[Desktop]
    USER[system_users]
  end

  subgraph INT[Integrações]
    IC[integration_connections]
    CUR[sync_cursors]
    BATCH[integration_sync_batches]
    EVENT[integration_events]
  end

  subgraph RAW[RAW imutável]
    RR[raw_records]
  end

  subgraph NORM[Financeiro normalizado]
    CONTACT[contacts]
    ACCOUNT[financial_accounts]
    REC[receivables]
    RECEIPT[receipts]
    PAY[payables]
    PAYMENT[payments]
    ADJ[financial_adjustments]
  end

  subgraph ENGINE[Motor e projeções]
    RULE[financial_rule_versions]
    RUN[calculation_runs]
    BAL[daily_balances]
    LINEAGE[record_lineage]
  end

  subgraph PLAN[Planejamento]
    BUD[budgets]
    FV[forecast_versions]
    FL[forecast_lines]
  end

  subgraph ACC[Contabilidade e gerencial]
    COA[chart_of_accounts]
    AE[accounting_entries + lines]
    SNAP[statement_snapshots + lines]
    INV[investment_calculations]
  end

  subgraph AUDIT[Auditoria]
    AU[audit_events]
    ALERT[alerts]
  end

  USER --> IC
  IC --> CUR
  IC --> BATCH
  BATCH --> RR
  EVENT --> RR
  RR --> CONTACT
  RR --> ACCOUNT
  RR --> REC
  RR --> RECEIPT
  RR --> PAY
  RR --> PAYMENT
  RR --> LINEAGE
  RULE --> RUN
  RUN --> BAL
  REC --> BAL
  RECEIPT --> BAL
  PAY --> BAL
  PAYMENT --> BAL
  ADJ --> BAL
  LINEAGE --> BAL
  FV --> FL
  BUD --> SNAP
  COA --> AE
  AE --> SNAP
  BAL --> SNAP
  SNAP --> INV
  AU --> ALERT
```

## 3. ERD do núcleo de integração

```mermaid
erDiagram
  SYSTEM_USERS ||--o{ INTEGRATION_CONNECTIONS : configures
  INTEGRATION_CONNECTIONS ||--o{ SYNC_CURSORS : tracks
  INTEGRATION_CONNECTIONS ||--o{ INTEGRATION_SYNC_BATCHES : runs
  INTEGRATION_CONNECTIONS ||--o{ INTEGRATION_EVENTS : receives
  INTEGRATION_SYNC_BATCHES ||--o{ RAW_RECORDS : imports
  INTEGRATION_EVENTS }o--o| RAW_RECORDS : materializes

  SYSTEM_USERS {
    uuid id PK
    text windows_sid UK
    text display_name
    uuid role_id FK
    boolean active
    timestamptz created_at
  }

  INTEGRATION_CONNECTIONS {
    uuid id PK
    uuid configured_by FK
    text provider UK
    text status
    text external_account_id
    text external_account_name
    text secret_ref
    text_array authorized_scopes
    timestamptz authorized_at
    timestamptz last_success_at
    timestamptz last_attempt_at
    text last_error_code
    timestamptz updated_at
  }

  SYNC_CURSORS {
    uuid id PK
    uuid connection_id FK
    text resource
    text cursor_value
    timestamptz watermark_at
    timestamptz updated_at
  }

  INTEGRATION_SYNC_BATCHES {
    uuid id PK
    uuid connection_id FK
    text resource
    timestamptz started_at
    timestamptz finished_at
    text status
    integer records_read
    integer records_created
    integer records_updated
    integer records_failed
    integer retry_count
    text error_code
    text error_message_sanitized
  }

  INTEGRATION_EVENTS {
    uuid id PK
    uuid connection_id FK
    text provider_event_id
    text event_type
    timestamptz occurred_at
    timestamptz received_at
    text payload_hash
    text processing_status
    uuid raw_record_id FK
  }

  RAW_RECORDS {
    uuid id PK
    uuid sync_batch_id FK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    text payload_hash
    jsonb payload
    timestamptz source_updated_at
    timestamptz imported_at
    boolean is_current
  }
```

Constraints essenciais:

```sql
unique (connection_id, resource) on sync_cursors
unique (connection_id, provider_event_id) on integration_events
unique (connection_id, source_system, source_entity, source_id, payload_hash) on raw_records
```

Um índice parcial garantirá no máximo uma versão `is_current = true` por `connection_id/source_system/source_entity/source_id`.

`integration_sync_logs` será uma view sobre `integration_sync_batches`, preservando o contrato nominal do handoff sem duplicar os dados de execução.

RBAC será efetivo no worker, não somente declarativo. `roles`, `permissions` e `role_permissions` formarão catálogo versionado; toda chamada IPC carrega o SID autenticado do cliente, que deve corresponder a `system_users.windows_sid`. O usuário único inicial recebe `ADMIN`, mas operações materiais (`CREDENTIAL_WRITE`, `INTEGRATION_DISCONNECT`, `MANUAL_ADJUST`, `PERIOD_CLOSE`, `BACKUP_EXPORT`, `RESTORE`, `UPDATE`) passam pela mesma autorização e geram auditoria. Testes negativos usarão SID não cadastrado e permissão ausente.

## 4. ERD financeiro

```mermaid
erDiagram
  CONTACTS ||--o{ RECEIVABLES : debtor
  CONTACTS ||--o{ PAYABLES : supplier
  FINANCIAL_ACCOUNTS ||--o{ RECEIPTS : receives
  FINANCIAL_ACCOUNTS ||--o{ PAYMENTS : pays
  FINANCIAL_ACCOUNTS ||--o{ FINANCIAL_BALANCE_SNAPSHOTS : closes
  RECEIVABLES ||--o{ RECEIPT_ALLOCATIONS : settles
  RECEIPTS ||--o{ RECEIPT_ALLOCATIONS : allocates
  PAYABLES ||--o{ PAYMENT_ALLOCATIONS : settles
  PAYMENTS ||--o{ PAYMENT_ALLOCATIONS : allocates
  FINANCIAL_ADJUSTMENTS }o--|| SYSTEM_USERS : authorized_by

  CONTACTS {
    uuid id PK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    text kind
    text name
    text document_hash
    text status
    uuid raw_record_id FK
  }

  FINANCIAL_ACCOUNTS {
    uuid id PK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    text name
    text account_type
    text status
    uuid raw_record_id FK
  }

  FINANCIAL_BALANCE_SNAPSHOTS {
    uuid id PK
    uuid connection_id FK
    uuid financial_account_id FK
    date reference_date
    numeric reconciled_amount
    uuid rule_version_id FK
    text status
    timestamptz calculated_at
  }

  RECEIVABLES {
    uuid id PK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    text business_type
    uuid customer_id FK
    text external_document
    date issue_date
    date competence_date
    date due_date
    date expected_receipt_date
    numeric gross_amount
    numeric fee_amount
    numeric net_amount
    numeric open_balance
    text currency_code
    text status
    text payment_method
    integer installment_number
    integer installment_count
    uuid raw_record_id FK
    timestamptz created_at
    timestamptz updated_at
  }

  RECEIPTS {
    uuid id PK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    uuid financial_account_id FK
    date receipt_date
    numeric amount
    numeric interest
    numeric penalty
    numeric discount
    text currency_code
    text status
    uuid raw_record_id FK
  }

  RECEIPT_ALLOCATIONS {
    uuid receivable_id FK
    uuid receipt_id FK
    numeric allocated_amount
  }

  PAYABLES {
    uuid id PK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    uuid supplier_id FK
    text document
    text category
    date competence_date
    date due_date
    date payment_date
    numeric original_amount
    numeric interest
    numeric penalty
    numeric discount
    numeric open_balance
    text currency_code
    text status
    text cost_center
    uuid raw_record_id FK
  }

  PAYMENTS {
    uuid id PK
    uuid connection_id FK
    text source_system
    text source_entity
    text source_id
    uuid financial_account_id FK
    date payment_date
    numeric amount
    text currency_code
    text status
    uuid raw_record_id FK
  }

  PAYMENT_ALLOCATIONS {
    uuid payable_id FK
    uuid payment_id FK
    numeric allocated_amount
  }

  FINANCIAL_ADJUSTMENTS {
    uuid id PK
    date effective_date
    numeric amount
    text direction
    text reason
    uuid authorized_by FK
    timestamptz authorized_at
    text status
  }
```

Constraints essenciais:

```sql
unique (connection_id, source_system, source_entity, source_id) on contacts
unique (connection_id, source_system, source_entity, source_id) on financial_accounts
unique (connection_id, source_system, source_entity, source_id) on receivables
unique (connection_id, source_system, source_entity, source_id) on receipts
unique (connection_id, source_system, source_entity, source_id) on payables
unique (connection_id, source_system, source_entity, source_id) on payments
primary key (receivable_id, receipt_id) on receipt_allocations
primary key (payable_id, payment_id) on payment_allocations
unique (connection_id, financial_account_id, reference_date, rule_version_id) on financial_balance_snapshots
check (business_type in ('B2B', 'B2C', 'B2C_EXCLUDED_FROM_BLING_FUTURE', 'UNCLASSIFIED'))
check (installment_number >= 1)
check (installment_count >= installment_number)
```

## 5. Projeções, regras e linhagem

```mermaid
erDiagram
  FINANCIAL_RULE_VERSIONS ||--o{ CALCULATION_RUNS : governs
  CALCULATION_RUNS ||--o{ DAILY_BALANCES : produces
  SCENARIOS ||--o{ DAILY_BALANCES : selects
  RAW_RECORDS ||--o{ NORMALIZED_RECORD_SOURCES : source
  DAILY_BALANCES ||--o{ DAILY_BALANCE_CONTRIBUTIONS : explains

  FINANCIAL_RULE_VERSIONS {
    uuid id PK
    text rule_code
    integer version
    text description
    jsonb parameters
    text status
    uuid approved_by FK
    timestamptz approved_at
  }

  CALCULATION_RUNS {
    uuid id PK
    text calculation_type
    uuid rule_version_id FK
    uuid scenario_id FK
    date reference_date
    timestamptz started_at
    timestamptz finished_at
    text status
    text input_hash
  }

  SCENARIOS {
    uuid id PK
    text code UK
    text name
    boolean active
  }

  DAILY_BALANCES {
    uuid id PK
    uuid calculation_run_id FK
    uuid scenario_id FK
    date balance_date
    numeric opening_balance
    numeric inflows
    numeric outflows
    numeric adjustments
    numeric closing_balance
    text balance_type
    timestamptz calculated_at
  }

  NORMALIZED_RECORD_SOURCES {
    uuid id PK
    uuid raw_record_id FK
    text target_entity
    uuid target_id
    uuid rule_version_id FK
    timestamptz created_at
  }

  DAILY_BALANCE_CONTRIBUTIONS {
    uuid id PK
    uuid daily_balance_id FK
    text source_kind
    uuid source_id
    numeric contribution_amount
    uuid rule_version_id FK
  }
```

Constraints:

```sql
unique (rule_code, version) on financial_rule_versions
unique (calculation_run_id, balance_date) on daily_balances
check (balance_type in ('REALIZED', 'PROVISIONAL', 'PROJECTED'))
check (scenario.code in ('BASE', 'CONSERVATIVE', 'STRESS', 'OPTIMISTIC'))
```

Não criar `unique(date, scenario, balance_type)` global, pois isso apagaria a possibilidade de histórico. A unicidade pertence ao `calculation_run_id`.

O diagrama acima resume as famílias de linhagem. A migration não poderá implementar `target_entity/target_id` ou `source_kind/source_id` como FKs polimórficas sem integridade. Serão criadas tabelas tipadas:

- `contact_sources`, `financial_account_sources`, `receivable_sources`, `receipt_sources`, `payable_sources` e `payment_sources`, cada uma com FK para a entidade normalizada, FK para `raw_records`, FK para `financial_rule_versions` e PK composta;
- `financial_balance_snapshot_sources`, com FKs para o snapshot de saldo, os RAW de conta/lançamento e a regra de composição;
- `daily_balance_receivable_contributions`, `daily_balance_receipt_contributions`, `daily_balance_payable_contributions`, `daily_balance_payment_contributions` e `daily_balance_adjustment_contributions`, cada uma com FKs reais para `daily_balances` e para a entidade contribuinte;
- `statement_line_contributions`, com FK para `statement_lines` e exatamente uma FK de origem entre `accounting_entry_lines` e `daily_balances`, validada por `check`.

`record_lineage` poderá existir somente como view `UNION ALL` de leitura. Não será a fonte de integridade. Testes deverão percorrer card -> snapshot/linha -> contribuição tipada -> entidade -> RAW, falhando se qualquer caminho estiver ausente.

Nos contratos de leitura, `receivables.source_payload_reference` e `payables.source_reference` serão arrays de UUID derivados de `receivable_sources` e `payable_sources`. Não serão colunas textuais nem substituirão as FKs tipadas.

## 6. Planejamento e demonstrações

```mermaid
erDiagram
  FORECAST_VERSIONS ||--o{ FORECAST_LINES : contains
  SCENARIOS ||--o{ FORECAST_VERSIONS : uses
  CHART_OF_ACCOUNTS ||--o{ BUDGET_LINES : classifies
  CHART_OF_ACCOUNTS ||--o{ ACCOUNTING_ENTRY_LINES : classifies
  ACCOUNTING_ENTRIES ||--|{ ACCOUNTING_ENTRY_LINES : contains
  ACCOUNTING_PERIODS ||--o{ ACCOUNTING_ENTRIES : groups
  STATEMENT_SNAPSHOTS ||--|{ STATEMENT_LINES : contains
  STATEMENT_LINES ||--o{ STATEMENT_LINE_CONTRIBUTIONS : explains
  ACCOUNTING_ENTRY_LINES ||--o{ STATEMENT_LINE_CONTRIBUTIONS : contributes
  DAILY_BALANCES ||--o{ STATEMENT_LINE_CONTRIBUTIONS : contributes
  CALCULATION_RUNS ||--o{ STATEMENT_SNAPSHOTS : produces
  STATEMENT_SNAPSHOTS ||--o{ INVESTMENT_CALCULATIONS : informs
  INVESTMENT_PARAMETERS ||--o{ INVESTMENT_CALCULATIONS : configures

  FORECAST_VERSIONS {
    uuid id PK
    text name UK
    uuid scenario_id FK
    jsonb assumptions
    uuid created_by FK
    timestamptz created_at
    text status
  }

  FORECAST_LINES {
    uuid id PK
    uuid forecast_version_id FK
    text account_or_category
    date period
    numeric amount
  }

  BUDGET_LINES {
    uuid id PK
    uuid chart_account_id FK
    date period
    numeric budget_amount
    numeric actual_amount
    numeric variance_amount
    numeric variance_percent
    text favorability
  }

  ACCOUNTING_PERIODS {
    uuid id PK
    date start_date
    date end_date
    text status
    timestamptz closed_at
  }

  CHART_OF_ACCOUNTS {
    uuid id PK
    text code UK
    text name
    text nature
    uuid parent_id FK
    boolean active
  }

  ACCOUNTING_ENTRIES {
    uuid id PK
    uuid period_id FK
    date entry_date
    text description
    text source_system
    text source_id
    text status
  }

  ACCOUNTING_ENTRY_LINES {
    uuid id PK
    uuid entry_id FK
    uuid chart_account_id FK
    numeric debit_amount
    numeric credit_amount
    uuid raw_record_id FK
  }

  STATEMENT_SNAPSHOTS {
    uuid id PK
    uuid calculation_run_id FK
    text statement_type
    date period_start
    date period_end
    uuid scenario_id FK
    text status
    timestamptz calculated_at
  }

  STATEMENT_LINES {
    uuid id PK
    uuid snapshot_id FK
    text line_code
    text label
    numeric amount
    integer display_order
    uuid rule_version_id FK
  }

  STATEMENT_LINE_CONTRIBUTIONS {
    uuid id PK
    uuid statement_line_id FK
    uuid accounting_entry_line_id FK
    uuid daily_balance_id FK
    numeric contribution_amount
    uuid rule_version_id FK
  }

  INVESTMENT_CALCULATIONS {
    uuid id PK
    uuid statement_snapshot_id FK
    uuid daily_balance_run_id FK
    numeric eligible_accumulated_profit
    numeric already_invested
    numeric available_profit
    numeric worst_week_cash
    numeric minimum_reserve
    numeric financial_limit
    numeric maximum_investment
    uuid rule_version_id FK
    text status
  }

  INVESTMENT_PARAMETERS {
    uuid id PK
    uuid rule_version_id FK
    date effective_from
    numeric minimum_reserve
    integer eligible_profit_cutoff_months
    integer projection_horizon_days
    boolean include_expected_receipts
    uuid approved_by FK
    timestamptz approved_at
  }
```

Constraint obrigatória de `statement_line_contributions`: exatamente uma das colunas `accounting_entry_line_id` e `daily_balance_id` deve estar preenchida. A migration deve usar FKs, `check` XOR e teste de deleção/restrição.

`dre_snapshots`, `pnl_snapshots`, `ebitda_snapshots` e `working_capital_snapshots` serão views tipadas sobre `statement_snapshots`, filtradas por `statement_type`. `working_capital_snapshots` expõe também `current_assets`, `current_liabilities` e a diferença, todos com lineage. As views mantêm os nomes literais do handoff; a tabela física comum mantém versionamento e auditoria uniformes.

## 7. Auditoria e alertas

| Entidade | Finalidade | Campos mínimos |
|---|---|---|
| `audit_events` | Ações materiais e alterações de configuração | ator, ação, entidade, ID, before/after sanitizado, timestamp, correlation ID |
| `alerts` | Riscos financeiros ou operacionais | tipo, severidade, status, mensagem simples, referência, aberto/resolvido em |
| `accounting_adjustments` | Diferenças DRE/P&L e ajustes contábeis | descrição, motivo, valor, usuário, data, aprovação |
| `data_quality_issues` | Divergências, ausência, atraso e inconsistência | regra, entidade, registro, severidade, status e resolução |

Segredos, tokens e payloads pessoais completos não entram em `audit_events`.

Tipos iniciais de `alerts`: `CASH_BELOW_MINIMUM_RESERVE`, `OVERDUE_PAYABLES`, `OVERDUE_RECEIVABLES`, `LARGE_PAYMENT_DUE`, `EBITDA_RELEVANT_DROP`, `SYNC_ERROR`, `ACCOUNTING_DIVERGENCE` e `STALE_DATA`.

## 8. Decisões ainda sujeitas à aprovação funcional

1. Moeda BRL como única moeda do MVP.
2. Timezone `America/Sao_Paulo`.
3. Precisão/escala e política de arredondamento.
4. Regras para classificar um contato/recebível Bling como B2B.
5. Status de cancelamentos, estornos, chargebacks e reembolsos.
6. Campos exatos de `receipts`, `payments` e contas financeiras conforme payload real.
7. Plano de contas e regras de classificação DRE/P&L.
8. Fórmula completa da aplicação conforme modelo de referência ainda inexistente.
9. Forma oficial de obter taxas, líquido e agenda do Nuvem Pago.
10. Política LGPD para retenção, pseudonimização e redação auditável de PII em RAW e backups.

Essas decisões bloqueiam os cálculos e módulos dependentes. A fundação técnica pode ser construída isoladamente, mas nenhum resultado afetado poderá receber status `CONFIRMED` nem passar pelos gates G2/G3/G5/G7 antes dos respectivos oráculos e regras serem aprovados.

## 9. Validações estruturais obrigatórias

- débitos = créditos por lançamento contábil;
- Ativo = Passivo + Patrimônio Líquido;
- Saldo Final D = Saldo Inicial D + Entradas - Saídas ± Ajustes;
- Saldo Inicial D+1 = Saldo Final D;
- Total futuro = B2C Nuvem + B2B Bling, sem B2C Bling;
- nenhuma alocação supera saldo do título ou valor da liquidação;
- toda transformação material possui lineage;
- toda sincronização é idempotente;
- snapshots e forecasts nunca são sobrescritos.
