# Data Dictionary v1 - Majucau Financial Intelligence

- **Status:** Proposto; campos externos serão validados com payloads reais
- **Convenção:** nomes físicos em `snake_case`, domínio em inglês e interface em português
- **Moeda proposta:** BRL
- **Timezone proposto:** America/Sao_Paulo

## 1. Legenda de origem

| Código | Significado |
|---|---|
| `BLING` | Dado obtido diretamente da API Bling |
| `NUVEMSHOP` | Pedido/cliente/status obtido da API Nuvemshop |
| `NUVEM_PAGO` | Dado financeiro futuro de fonte oficial Nuvem Pago ainda não disponível |
| `PAYROLL` | Arquivo ou relatório oficial de folha |
| `MANUAL_VALIDATED` | Entrada manual revisada/autorizada |
| `CALCULATED` | Resultado produzido por regra versionada Majucau |
| `ACCOUNTING_ADJUSTMENT` | Ajuste contábil identificado e aprovado |

`source_system` aceita somente provedores de fatos (`BLING`, `NUVEMSHOP`, `NUVEM_PAGO`, `PAYROLL`, `MANUAL_VALIDATED`). A origem de cálculo é separada em `provenance_code`, por exemplo `CALCULATED_FROM_BLING`, `CALCULATED_FROM_NUVEMSHOP`, `CALCULATED_MIXED` e `ACCOUNTING_ADJUSTMENT`. Na UI, `NUVEMSHOP` e `NUVEM_PAGO` podem aparecer como “Nuvem”, mas nunca compartilham identidade, credencial ou confirmação implícita.

## 2. Convenções de tipos

| Conteúdo | PostgreSQL | Go | TypeScript | Regra |
|---|---|---|---|---|
| ID interno | `uuid` | `uuid.UUID`/tipo próprio | `string` | Nunca usar ID visual externo como PK |
| ID externo | `text` | `string` | `string` | Preservar exatamente como recebido |
| Dinheiro | `numeric(19,4)` | tipo decimal/Money | string decimal | Nunca converter para float |
| Taxa/percentual | `numeric(20,8)` | tipo decimal | string decimal | Arredondar somente na regra aprovada |
| Data de negócio | `date` | tipo LocalDate próprio | `YYYY-MM-DD` | Sem conversão implícita de timezone |
| Instante técnico | `timestamptz` | `time.Time` UTC | ISO 8601 | Persistir em UTC |
| Payload externo | `jsonb` | `json.RawMessage` | não expor integralmente | RAW imutável |
| Enum | `text` + check/lookup | tipo definido | union type | Rejeitar valor desconhecido ou mapear explicitamente |

## 3. Metadados comuns e linhagem

| Campo interno | Tipo | Origem | Regra |
|---|---|---|---|
| `id` | uuid | CALCULATED | PK técnica gerada internamente |
| `connection_id` | uuid nullable | integração | FK da conta externa; obrigatório para registros de API |
| `source_system` | text | integração | Provedor responsável pelo fato |
| `source_entity` | text | integração | Recurso original, como `accounts_receivable` |
| `source_id` | text | API/arquivo | Identificador original |
| `raw_record_id` | uuid | RAW | FK para versão exata do payload |
| `payload_hash` | text | CALCULATED | SHA-256 do payload canônico |
| `sync_batch_id` | uuid | CALCULATED | Lote que importou o registro |
| `source_updated_at` | timestamptz nullable | fonte | Data de atualização informada pelo provedor |
| `imported_at` | timestamptz | CALCULATED | Momento da ingestão |
| `created_at` | timestamptz | CALCULATED | Criação interna |
| `updated_at` | timestamptz | CALCULATED | Última alteração tratada |
| `provenance_code` | text | CALCULATED | Como o valor foi produzido, sem misturar provedor e regra |

A chave de identidade externa normalizada é `(connection_id, source_system, source_entity, source_id)`. O mesmo ID em recursos ou contas distintas nunca será fundido.

### 3.1 Perfis e autorização

| Perfil | Escopo inicial |
|---|---|
| `ADMIN` | Configuração, credenciais, usuários, backup/restore, atualização e todas as operações funcionais |
| `DIRECTOR` | Visão executiva, cenários, aprovações e relatórios; sem administrar secrets ou restore |
| `FINANCE` | Tesouraria, integrações operacionais, conciliação e ajustes financeiros autorizados |
| `ACCOUNTING` | Plano de contas, períodos, DRE/P&L, lançamentos e ajustes contábeis |
| `VIEWER` | Somente leitura de telas e drill-down permitido pela política de PII |

Permissão é validada no worker a partir do SID e do catálogo `role_permissions`; esconder botão no frontend não autoriza nem protege uma operação.

## 4. Integrações

### 4.1 `integration_connections`

| Campo | Tipo | Exposição na UI | Regra |
|---|---|---|---|
| `provider` | text | Bling/Nuvemshop/Nuvem Pago | Unique por provedor no single-user |
| `status` | text | badge com tooltip | Enum de estado da conexão |
| `external_account_id` | text nullable | ID da conta/loja | Retornado após autorização |
| `external_account_name` | text nullable | Nome legível | Não usar como chave |
| `secret_ref` | text nullable | nunca mostrar | Referência opaca ao arquivo DPAPI |
| `authorized_scopes` | text[] | lista legível | Escopos efetivamente concedidos |
| `authorized_at` | timestamptz nullable | data da conexão | Não implica sincronização |
| `last_success_at` | timestamptz nullable | última sincronização | Atualizar somente após commit completo |
| `last_attempt_at` | timestamptz nullable | última tentativa | Pode ser posterior ao último sucesso |
| `last_error_code` | text nullable | mensagem traduzida | Sem token/payload sensível |

Status permitidos:

```text
NOT_CONFIGURED
AUTHORIZING
CONNECTED
TOKEN_EXPIRING
AUTH_ERROR
SCHEMA_MISMATCH
SYNCING
STALE
PARTIALLY_AVAILABLE
UNAVAILABLE
```

O estado da conexão e o estado de frescor serão armazenados separadamente. Padrão operacional proposto: sincronização a cada 15 minutos; `STALE` após duas execuções perdidas ou 30 minutos sem sucesso; `PARTIALLY_AVAILABLE` quando ao menos um recurso obrigatório falhar ou uma fonte não fornecer todos os campos; `UNAVAILABLE` quando a fonte não estiver configurada, não possuir contrato oficial ou tiver falha permanente. Os limiares serão configuráveis e aparecerão no tooltip/status.

### 4.2 `integration_sync_batches`

| Campo | Tipo | Regra |
|---|---|---|
| `connection_id` | uuid | Conexão executada |
| `resource` | text | Recurso sincronizado |
| `started_at` | timestamptz | Início |
| `finished_at` | timestamptz nullable | Ausente enquanto em execução |
| `status` | text | RUNNING/SUCCESS/PARTIAL/FAILED/CANCELLED |
| `records_read` | integer | Quantidade lida |
| `records_created` | integer | Novas versões/entidades |
| `records_updated` | integer | Estado tratado alterado |
| `records_failed` | integer | Rejeitados com erro |
| `retry_count` | integer | Tentativas transitórias |
| `error_code` | text nullable | Código interno estável |
| `error_message_sanitized` | text nullable | Sem segredo nem dado pessoal desnecessário |

`integration_sync_logs` será uma view de auditoria sobre `integration_sync_batches`, com os mesmos IDs, recurso, início/fim, status, contagens, retries e erros sanitizados. Assim, o nome literal exigido pelo handoff permanece disponível sem duplicar a fonte de verdade.

## 5. RAW

### 5.1 `raw_records`

| Campo | Tipo | Regra |
|---|---|---|
| `connection_id` | uuid | Conta externa responsável pelo payload |
| `source_system` | text | BLING/NUVEMSHOP/NUVEM_PAGO/PAYROLL |
| `source_entity` | text | Recurso externo |
| `source_id` | text | ID original |
| `payload` | jsonb | Payload completo permitido pela política de dados |
| `payload_hash` | text | Deduplicação de versão |
| `source_updated_at` | timestamptz nullable | Quando disponível |
| `imported_at` | timestamptz | Momento local |
| `is_current` | boolean | Somente uma versão atual por identidade externa |

Registros RAW não são atualizados por sincronização ou regra de negócio. Quando a fonte muda, inserir nova versão e marcar a anterior como não atual dentro da mesma transação. A única exceção é redação LGPD autorizada: remover PII do payload e manter tombstone auditável com hash anterior, motivo, ator e instante, sem alterar valores financeiros não pessoais necessários à obrigação legal.

## 6. Contatos e contas

### 6.1 `contacts`

| Fonte | Campo original | Campo interno | Tipo | Regra |
|---|---|---|---|---|
| Bling `/contatos` | `id` | `source_id` | text | ID do contato |
| Bling/Nuvemshop | nome | `name` | text | Manter versão original no RAW |
| Bling/Nuvemshop | documento | `document_hash` | text nullable | Hash para correlação; exibição deve respeitar LGPD |
| Regra Majucau | configuração | `kind` | text | CUSTOMER/SUPPLIER/BOTH/EMPLOYEE/OWNER |
| Fonte | situação/status | `status` | text | Mapeamento versionado |

### 6.2 `financial_accounts`

| Fonte | Campo original | Campo interno | Tipo | Regra |
|---|---|---|---|---|
| Bling `/contas-contabeis` | ID | `source_id` | text | Validar nome exato no payload real |
| Bling | descrição/nome | `name` | text | Conta financeira legível |
| Bling | tipo | `account_type` | text | Mapear sem inferência |
| Bling | situação | `status` | text | Ativa/inativa conforme payload |

O saldo conciliado não é um campo agregado comprovado; será calculado a partir das fontes documentadas e reconciliadas.

### 6.3 `financial_balance_snapshots`

| Campo | Tipo | Regra |
|---|---|---|
| `reference_date` | date | Data de fechamento D-1 |
| `connection_id` | uuid | Conexão Bling |
| `financial_account_id` | uuid | Conta financeira incluída |
| `reconciled_amount` | numeric(19,4) | Saldo composto apenas por lançamentos/contas conciliados |
| `source_reference` | uuid[] | RAW das contas e lançamentos usados |
| `rule_version_id` | uuid | Filtro e composição aplicados |
| `status` | text | `PENDING_RECONCILIATION`, `CONFIRMED` ou `DIVERGENT` |

O total confirmado de D-1 deve fechar com a visão conciliada equivalente do Bling no mesmo conjunto de contas/filtros com tolerância de R$ 0,01. Esse total alimenta `daily_balances.opening_balance` em D0; a continuidade interna D para D+1 continua exigindo diferença R$ 0,00.

## 7. Recebíveis

### 7.1 `receivables`

| Fonte | Campo original | Campo interno | Tipo | Regra |
|---|---|---|---|---|
| Bling/Nuvemshop | ID | `source_id` | text | Unicidade com conexão, sistema e entidade |
| Regra Majucau | classificação | `business_type` | text | B2B/B2C/B2C_EXCLUDED_FROM_BLING_FUTURE/UNCLASSIFIED; regra versionada abaixo |
| Fonte | contato/cliente | `customer_id` | uuid | FK para `contacts` |
| Fonte | documento/pedido | `external_document` | text | Referência visível, não PK |
| Fonte | emissão | `issue_date` | date nullable | Data de criação do título/pedido |
| Regra/fonte | competência | `competence_date` | date nullable | Não confundir com caixa |
| Fonte | vencimento | `due_date` | date nullable | Obrigatório quando houver título |
| Nuvem Pago | previsão | `expected_receipt_date` | date nullable | Não preencher a partir de pedido Nuvemshop sem fonte oficial |
| Fonte | bruto | `gross_amount` | numeric(19,4) | Exato |
| Nuvem Pago | taxa | `fee_amount` | numeric(19,4) nullable | Não estimar silenciosamente |
| Nuvem Pago | líquido | `net_amount` | numeric(19,4) nullable | Não derivar de campo ambíguo |
| Fonte | saldo | `open_balance` | numeric(19,4) nullable | Deve reconciliar com a fonte |
| Fonte | moeda | `currency_code` | char(3) | BRL proposta |
| Fonte/regra | situação | `status` | text | Mapeamento controlado |
| Fonte | pagamento | `payment_method` | text nullable | Preservar nomenclatura original no RAW |
| Fonte | parcela | `installment_number` | integer nullable | >= 1 |
| Fonte | total parcelas | `installment_count` | integer nullable | >= número da parcela |
| RAW/lineage | referência da origem | `source_payload_reference` | uuid[] | Campo de API derivado das FKs das tabelas `receivable_sources`; equivale aos registros RAW exatos, não a texto livre |

Regra de composição:

```text
Total a Receber = B2C futuro da Nuvem + B2B futuro do Bling
```

Nunca somar B2C futuro do Bling.

Regra B2B proposta, em prioridade: override auditado; correspondência com pedido Nuvemshop exclui o futuro do Bling; CNPJ válido inclui como B2B; lista aprovada de canal/categoria/contato inclui como B2B; demais ficam `UNCLASSIFIED` e não entram em total confirmado. A correlação usa ID direto ou igualdade exata de conexão/loja, número externo, hash do documento, moeda, valor bruto e data; zero/múltiplos candidatos não são classificados automaticamente. Toda ambiguidade aparece numa fila de revisão e alterações geram nova versão de regra/snapshot.

Aging proposto em D-005 usa faixas mutuamente exclusivas na data de referência: `OVERDUE` (< D0), `TODAY` (= D0), `D1_7`, `D8_15`, `D16_30`, `D31_45` e `D46_60`. Títulos parciais usam somente `open_balance`; cancelados/estornados não entram como abertos. Todas as fronteiras são datas de negócio em `America/Sao_Paulo`. A regra não fica ativa antes da aprovação D-005.

### 7.2 `receipts`

| Campo | Tipo | Fonte/regra |
|---|---|---|
| `source_system` | text | BLING para realizado oficial |
| `source_id` | text | ID de baixa/recebimento a validar no payload |
| `financial_account_id` | uuid nullable | Conta de recebimento |
| `receipt_date` | date | Data efetiva de caixa |
| `amount` | numeric(19,4) | Valor realizado |
| `interest` | numeric(19,4) | Se disponível |
| `penalty` | numeric(19,4) | Se disponível |
| `discount` | numeric(19,4) | Se disponível |
| `status` | text | CONFIRMED após reconciliação |

A Nuvemshop não alimenta recebimentos realizados oficiais.

## 8. Obrigações e pagamentos

### 8.1 `payables`

| Campo | Tipo | Fonte/regra |
|---|---|---|
| `source_system` | text | BLING |
| `source_id` | text | ID da conta a pagar |
| `supplier_id` | uuid nullable | Contato/fornecedor |
| `document` | text nullable | Documento externo |
| `category` | text nullable | Categoria da fonte e posterior mapeamento contábil |
| `competence_date` | date nullable | Competência |
| `due_date` | date | Vencimento |
| `payment_date` | date nullable | Caixa efetivo |
| `original_amount` | numeric(19,4) | Valor original |
| `interest` | numeric(19,4) | Juros |
| `penalty` | numeric(19,4) | Multa |
| `discount` | numeric(19,4) | Desconto |
| `open_balance` | numeric(19,4) | Saldo em aberto |
| `status` | text | Mapeamento de situação Bling |
| `cost_center` | text nullable | Centro de custo quando disponível |
| `source_reference` | uuid[] | Campo de API derivado das FKs de `payable_sources`; aponta para os RAW exatos |

### 8.2 `payments`

| Campo | Tipo | Fonte/regra |
|---|---|---|
| `source_system` | text | BLING |
| `source_id` | text | ID da baixa/pagamento a validar |
| `financial_account_id` | uuid nullable | Conta de saída |
| `payment_date` | date | Data de caixa |
| `amount` | numeric(19,4) | Valor pago |
| `status` | text | CONFIRMED após reconciliação |

## 9. Projeção diária

### 9.1 `daily_balances`

| Campo | Tipo | Regra |
|---|---|---|
| `calculation_run_id` | uuid | Garante histórico e reprodutibilidade |
| `scenario_id` | uuid | BASE/CONSERVATIVE/STRESS/OPTIMISTIC |
| `balance_date` | date | D0 até D+60 |
| `opening_balance` | numeric(19,4) | D0 recebe saldo final conciliado D-1; D+1 recebe closing D |
| `inflows` | numeric(19,4) | Entradas elegíveis daquele dia |
| `outflows` | numeric(19,4) | Saídas elegíveis daquele dia |
| `adjustments` | numeric(19,4) | Somente ajustes autorizados |
| `closing_balance` | numeric(19,4) | opening + inflows - outflows ± adjustments |
| `balance_type` | text | REALIZED/PROVISIONAL/PROJECTED |
| `calculated_at` | timestamptz | Execução do motor |

Regras temporais:

- D-1 e anteriores: realizado/conciliado;
- D0: provisório/em andamento;
- D+1 a D+60: projetado.

## 10. Planejamento

### 10.1 `budget_lines`

| Campo | Tipo | Regra |
|---|---|---|
| `account_or_category` | text/uuid | Linha orçamentária |
| `period` | date | Competência do período |
| `budget_amount` | numeric(19,4) | Orçado |
| `actual_amount` | numeric(19,4) | Realizado calculado |
| `variance_amount` | numeric(19,4) | actual - budget, com semântica por natureza |
| `variance_percent` | numeric(20,8) nullable | Evitar divisão por zero |
| `favorability` | text | FAVORABLE/UNFAVORABLE/NEUTRAL/NOT_APPLICABLE |

Valor maior não é automaticamente favorável.

### 10.2 `forecast_versions` e `forecast_lines`

| Campo | Tipo | Regra |
|---|---|---|
| `name` | text | Ex.: Forecast Agosto v1 |
| `scenario_id` | uuid | Cenário associado |
| `assumptions` | jsonb | Premissas explícitas |
| `created_by` | uuid | Usuário responsável |
| `created_at` | timestamptz | Criação |
| `status` | text | DRAFT/APPROVED/FINAL/ARCHIVED |
| `account_or_category` | text/uuid | Linha da versão |
| `period` | date | Período |
| `amount` | numeric(19,4) | Valor previsto |

Versões são imutáveis após aprovação; correção cria nova versão.

## 11. DRE, P&L e EBITDA

### 11.1 `statement_snapshots`

| Campo | Tipo | Regra |
|---|---|---|
| `statement_type` | text | DRE/PNL/EBITDA/BALANCE_SHEET/TRIAL_BALANCE/WORKING_CAPITAL |
| `period_start` | date | Início |
| `period_end` | date | Fim |
| `scenario_id` | uuid nullable | Realizado ou cenário |
| `calculation_run_id` | uuid | Regra e inputs usados |
| `status` | text | CONFIRMED/PARTIALLY_CONFIRMED/PROVISIONAL/PROJECTED/DIVERGENT |

### 11.2 `statement_lines`

| Campo | Tipo | Regra |
|---|---|---|
| `line_code` | text | Código estável da linha |
| `label` | text | Nomenclatura apresentada |
| `amount` | numeric(19,4) | Valor exato |
| `display_order` | integer | Ordem no demonstrativo |
| `rule_version_id` | uuid | Regra parametrizável aplicada |

Toda linha deve permitir drill-down até contribuições, registros tratados e RAW.

Os nomes literais do handoff serão expostos como views tipadas sobre o modelo comum: `dre_snapshots`, `pnl_snapshots`, `ebitda_snapshots` e `working_capital_snapshots`, filtrando `statement_snapshots.statement_type`. A tabela comum é a fonte física; as views mantêm contratos claros e evitam quatro implementações divergentes de versionamento e lineage.

## 12. Aplicação financeira

### 12.0 `investment_parameters`

| Campo | Tipo | Regra |
|---|---|---|
| `rule_version_id` | uuid | FK para a regra de aplicação aprovada |
| `effective_from` | date | Início de vigência, sem sobrescrever histórico |
| `minimum_reserve` | numeric(19,4) | Reserva mínima autorizada |
| `eligible_profit_cutoff_months` | integer | Valor proposto: 2, representando M-2 |
| `projection_horizon_days` | integer | Valor protegido: 60 |
| `include_expected_receipts` | boolean | Somente quando houver fonte oficial/data válida |
| `approved_by` | uuid | Responsável pela aprovação |
| `approved_at` | timestamptz | Momento da aprovação |

Parâmetros aprovados são imutáveis; mudança cria nova vigência e novos snapshots.

### 12.1 `investment_calculations`

| Campo | Tipo | Regra |
|---|---|---|
| `eligible_accumulated_profit` | numeric(19,4) | Lucro líquido acumulado elegível até M-2 |
| `already_invested` | numeric(19,4) | Valor já investido |
| `available_profit` | numeric(19,4) | max(0, eligible - invested) |
| `worst_week_cash` | numeric(19,4) | Menor saldo diário de fechamento encontrado dentro das semanas ISO dos próximos 60 dias |
| `minimum_reserve` | numeric(19,4) | Reserva mínima aprovada |
| `financial_limit` | numeric(19,4) | worst_week_cash - reserve |
| `maximum_investment` | numeric(19,4) | max(0, min(financial_limit, available_profit)) |
| `rule_version_id` | uuid | Versão da fórmula |
| `status` | text | PROVISIONAL até existir modelo de referência aprovado |

Regra v1 proposta: considerar somente recebimentos `CONFIRMED` ou `PROJECTED` com fonte oficial e data válida; calcular o saldo diário D0..D+60; `worst_week_cash` é o menor `closing_balance` diário do horizonte; `financial_limit = max(0, worst_week_cash - minimum_reserve)`; `maximum_investment = max(0, min(available_profit, financial_limit))`. A regra permanece `PROVISIONAL` e o card fica bloqueado até a aprovação dessa definição ou a entrega da planilha de referência, seguida de casos dourados com tolerância de R$ 0,01.

## 13. Status de dados

| Status | Uso simples na interface |
|---|---|
| `CONFIRMED` | “O dado foi conciliado com a fonte oficial.” |
| `PARTIALLY_CONFIRMED` | “Parte do cálculo foi confirmada; ainda existem componentes pendentes.” |
| `PROVISIONAL` | “O valor pode mudar porque o período ou a fonte ainda não foi conciliado.” |
| `PROJECTED` | “Estimativa futura baseada nas informações disponíveis.” |
| `PENDING_RECONCILIATION` | “O dado foi importado, mas ainda precisa ser comparado com a fonte.” |
| `DIVERGENT` | “Foi encontrada uma diferença que precisa ser investigada.” |

## 14. Lacunas que exigem payload ou aprovação

1. Nomes e tipos exatos de campos retornados por cada endpoint Bling.
2. Regra de classificação B2B.
3. Agenda, taxa e líquido do Nuvem Pago.
4. Formato do arquivo de folha.
5. Política de arredondamento por cálculo.
6. Plano de contas e mapeamento DRE/P&L.
7. Tratamento de chargeback, estorno, cancelamento e recebimento parcial.
8. Retenção e anonimização LGPD.
9. Parâmetros completos do cálculo de aplicação.

Essas lacunas não impedem a fundação técnica, mas impedem marcar os respectivos resultados como confirmados.

## 15. Defaults técnicos submetidos à aprovação

- moeda única do MVP: BRL;
- calendário e datas de negócio: `America/Sao_Paulo`, persistindo instantes técnicos em UTC;
- dinheiro: `numeric(19,4)` internamente, sem `float`; comparação/visualização em centavos;
- arredondamento: `ROUND_HALF_UP` para duas casas somente na fronteira da regra ou reconciliação, nunca em somas intermediárias;
- intervalos de data: início inclusivo, fim inclusivo para relatórios de data de negócio; intervalos técnicos em UTC são `[início, fim)`;
- pagamentos parciais preservam saldo aberto e alocações; cancelamento/estorno/chargeback cria evento compensatório, nunca apaga fato anterior.

## 16. Contrato de importação de folha proposto

Arquivo CSV UTF-8, delimitador `;`, competência `YYYY-MM`, um registro por pessoa/verba/departamento. Colunas obrigatórias: `competence`, `employee_external_id`, `employee_name`, `document`, `department_code`, `department_name`, `earning_code`, `earning_description`, `earning_type` (`SALARY`, `CHARGE`, `BENEFIT`, `PRO_LABORE`, `DEDUCTION`), `amount`, `currency`, `source_document`. O importador calcula hash do arquivo e da linha, mascara documento na UI, impede reimportação, valida totais por competência e rejeita linha inválida sem confirmar a folha. O template e exemplos sanitizados serão gerados na implementação; a folha não entra em DRE/EBITDA até o total do arquivo ser aprovado.

Regras protegidas da folha: departamentos iniciais `PRODUCTION`, `SALES` e `ADMINISTRATIVE`; vínculo `CLT`, `PARTNER_PRO_LABORE` ou `TERMINATED`; desligados permanecem no histórico e não geram competência após a data final; adiantamento salarial é `PAYROLL_ADVANCE_OFFSET`, compensação financeira contra a folha e nunca nova despesa; PLR é `NOT_APPLICABLE` até aprovação expressa.

## 17. Capital de giro e alertas

`working_capital_snapshots` materializa `current_assets`, `current_liabilities` e `working_capital = current_assets - current_liabilities`, todos vinculados à versão do plano de contas e às contribuições. A fórmula é padrão proposto; a classificação de contas circulantes precisa de aprovação contábil antes de `CONFIRMED`.

Códigos iniciais de alerta: `CASH_BELOW_MINIMUM_RESERVE`, `OVERDUE_PAYABLES`, `OVERDUE_RECEIVABLES`, `LARGE_PAYMENT_DUE`, `EBITDA_RELEVANT_DROP`, `SYNC_ERROR`, `ACCOUNTING_DIVERGENCE` e `STALE_DATA`. Cada alerta guarda severidade, limiar/regra versionada, entidade de referência, aberto/resolvido em e usuário da resolução.
