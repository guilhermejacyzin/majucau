# MAJUCAU FINANCIAL INTELLIGENCE  
## Handoff para Senior Dev / Tech Lead

**Empresa:** Majucau Chocolate Brasileiro Ltda.  
**Projeto:** Sistema Executivo de Inteligência Financeira e Contábil  
**Versão:** 0.1  
**Status:** Especificação inicial aprovada para desenvolvimento do MVP

---

# 1. OBJETIVO DO PROJETO

Construir um sistema próprio de inteligência financeira e contábil para tomada de decisão da Majucau.

O sistema deverá consolidar dados operacionais, financeiros e contábeis, transformando-os em:

- posição financeira;
- projeção financeira;
- recebíveis;
- recebimentos realizados;
- obrigações a pagar;
- pagamentos realizados;
- fluxo projetado;
- DRE;
- P&L;
- EBITDA;
- Orçado x Realizado;
- Forecast;
- capital de giro;
- rentabilidade;
- valor máximo para aplicação;
- balanço patrimonial;
- balancete contábil;
- indicadores;
- alertas;
- conciliações;
- rastreabilidade.

O sistema deve priorizar:

1. confiabilidade;
2. rastreabilidade;
3. auditabilidade;
4. não duplicidade;
5. separação entre realizado e projetado;
6. separação entre competência e caixa.

---

# 2. PRINCÍPIO CENTRAL DO SISTEMA

## Bling não é o motor de interpretação contábil.

O Bling será tratado prioritariamente como:

**fonte operacional de dados.**

A Majucau Financial Intelligence será:

**fonte de cálculo, classificação e interpretação.**

Portanto:

- não utilizar automaticamente a DRE pronta do Bling;
- não utilizar automaticamente o “balancete” do Bling como balancete contábil;
- não utilizar automaticamente demonstrações gerenciais do Bling como verdade contábil;
- extrair os fatos operacionais;
- aplicar regras próprias documentadas;
- manter rastreabilidade até o dado original.

---

# 3. ESCOPO DO MVP

A primeira versão deverá priorizar:

## Tesouraria e Projeção Financeira

- Saldo Inicial do Dia
- Saldo Final Projetado Hoje
- Saldo Projetado em 30 Dias
- Saldo Projetado em 60 Dias
- Menor Saldo Projetado em 60 Dias
- Entradas previstas
- Saídas previstas
- Reserva mínima
- Alertas de liquidez

## Recebíveis

- B2C a receber
- B2B a receber
- Total a receber
- Aging
- vencidos
- próximos 7 dias
- 15 dias
- 30 dias
- 45 dias
- 60 dias

## Realizado

- Recebimentos realizados
- Pagamentos realizados

## Obrigações

- Contas a pagar
- vencidas
- hoje
- 7 dias
- 15 dias
- 30 dias
- 45 dias
- 60 dias

## Resultado

- DRE
- P&L
- EBITDA

## Planejamento

- Orçado x Realizado
- Forecast

## Gestão

- Capital de Giro
- Valor Máximo para Aplicação
- Alertas

---

# 4. FORA DO ESCOPO DESTE MVP

Não desenvolver agora:

- previsão comercial por SKU;
- previsão de demanda;
- planejamento de produção;
- otimização de estoque;
- recomendação de compras;
- pricing comercial avançado;
- CRM;
- gestão de marketing;
- integração direta com Appmax;
- integração direta com Mercado Pago;
- integração direta com Itaú.

Esses temas poderão virar módulos posteriores.

---

# 5. FONTES OFICIAIS

## 5.1 Bling

Fonte principal para:

- saldo financeiro conciliado;
- contas a pagar;
- pagamentos realizados;
- B2B a receber;
- recebimentos realizados;
- clientes;
- fornecedores;
- notas fiscais;
- faturamento;
- categorias financeiras;
- compras;
- demais registros financeiros;
- informações conciliadas disponíveis.

## 5.2 Nuvem / Nuvem Pago

Fonte exclusiva para:

**B2C futuro a receber.**

Utilizar para:

- valor bruto;
- taxa;
- valor líquido;
- data prevista de recebimento;
- meio de pagamento;
- status;
- parcelas;
- demais informações financeiras disponíveis.

## 5.3 Folha

Fonte:

- relatórios do sistema de folha;
- resumo de encerramento;
- ficha de funcionário;
- provisões;
- pró-labore;
- FGTS;
- encargos;
- departamentos.

## 5.4 Dados calculados

DRE, EBITDA, P&L, forecast, capital de giro, valor máximo para aplicação e demais indicadores serão calculados pelo motor Majucau.

---

# 6. REGRA DE FONTE — RECEBÍVEIS

## FIN-REC-001 — B2C futuro

Fonte exclusiva:

**Nuvem.**

Recebíveis B2C futuros eventualmente existentes no Bling NÃO devem ser somados.

---

## FIN-REC-002 — B2B futuro

Fonte exclusiva:

**Bling.**

Somente registros classificados como B2B.

---

## FIN-REC-003 — Recebimentos realizados

Fonte oficial:

**Bling — situação paga/recebida.**

A Nuvem não alimenta o card de recebido realizado.

---

## FIN-REC-004 — Regra de não duplicidade

Nunca calcular:

**B2C Nuvem + B2C Bling.**

A fórmula do total futuro será:

**Total a Receber = B2C Nuvem + B2B Bling**

---

# 7. REGRA DE CORTE TEMPORAL

## FIN-TES-001 — Saldo Inicial do Dia

O saldo inicial do dia D será:

**Saldo Inicial D = Saldo Final Conciliado D-1**

Exemplo:

Saldo Final 14/08 = Saldo Inicial 15/08.

---

## FIN-TES-002 — Saldo Final

**Saldo Final = Saldo Inicial + Entradas − Saídas ± Ajustes Autorizados**

---

## FIN-TES-003 — Status temporal

### D-1 e anteriores

Status:

**Realizado / Conciliado**

### D0

Status:

**Provisório / Em andamento**

O saldo inicial é realizado.

Movimentos de D0 podem ainda não estar conciliados.

### D+1 até D+60

Status:

**Projetado**

---

# 8. PROJEÇÃO FINANCEIRA

Horizonte máximo inicial:

**60 dias.**

Calcular diariamente.

Nunca utilizar somente:

saldo inicial + total acumulado de entradas − total acumulado de saídas.

O motor deverá calcular:

D0  
D+1  
D+2  
...  
D+60

Cada saldo final alimenta o saldo inicial do dia seguinte.

---

# 9. MENOR SALDO PROJETADO

O sistema deve identificar:

**menor saldo diário dentro de D0 a D+60.**

Guardar:

- data;
- saldo;
- cenário;
- composição das entradas;
- composição das saídas.

Esse dado será utilizado também no módulo de aplicação financeira.

---

# 10. NOMENCLATURA OFICIAL

Não utilizar o termo genérico “Caixa” nos cards principais.

Utilizar:

- Saldo Inicial do Dia
- Saldo Final Projetado Hoje
- Saldo Projetado em 30 Dias
- Saldo Projetado em 60 Dias
- Menor Saldo Projetado — 60 Dias
- Reserva Mínima
- Valor Máximo para Aplicação

Menu:

**Tesouraria e Projeção Financeira**

---

# 11. PRIMEIRA TELA — VISÃO EXECUTIVA

Cards obrigatórios:

## Tesouraria

1. Saldo Inicial do Dia
2. Saldo Final Projetado Hoje
3. Saldo Projetado em 30 Dias
4. Saldo Projetado em 60 Dias
5. Menor Saldo Projetado — 60 Dias
6. Reserva Mínima
7. Valor Máximo para Aplicação

## Recebíveis

8. A Receber B2C — Nuvem
9. A Receber B2B — Bling
10. Total a Receber
11. Recebido no Mês

## Obrigações

12. A Pagar
13. Obrigações Vencidas
14. Pago no Mês

## Resultado

15. DRE / P&L
16. EBITDA

## Planejamento

17. Orçado x Realizado
18. Forecast

## Gestão

19. Capital de Giro
20. Alertas

---

# 12. REGRA DOS CARDS

Cada card deverá possuir:

- nome;
- valor atual;
- status;
- fonte;
- data de atualização;
- comparação quando aplicável;
- botão/área clicável;
- drill-down.

Cada card deve responder:

1. Onde estamos?
2. Para onde estamos indo?
3. Qual a origem do dado?
4. Há risco ou desvio?
5. O usuário consegue chegar ao detalhe?

---

# 13. DRILL-DOWN

Padrão:

**Card → Módulo → Detalhamento → Registro original**

Exemplo:

A Receber B2B

→ Recebíveis B2B

→ Cliente

→ Documento

→ Conta a receber original Bling

Sempre preservar a referência da origem.

---

# 14. STATUS DO DADO

Criar enum equivalente a:

- CONFIRMED
- PARTIALLY_CONFIRMED
- PROVISIONAL
- PROJECTED
- PENDING_RECONCILIATION
- DIVERGENT

Na interface:

- Confirmado
- Parcialmente confirmado
- Provisório
- Projetado
- Pendente de conciliação
- Divergente

---

# 15. ORIGEM DO DADO

Todo indicador deve registrar:

- BLING
- NUVEM
- PAYROLL
- MANUAL_VALIDATED
- CALCULATED
- ACCOUNTING_ADJUSTMENT

Exemplo:

Receita Líquida

Origem:

**CALCULATED_FROM_BLING**

---

# 16. ARQUITETURA SUGERIDA

```text
BLING API
    │
    ▼
BLING CONNECTOR
    │
    ▼
RAW DATABASE
    ▲
    │
NUVEM CONNECTOR
    │
    ▼
RAW DATABASE
    │
    ▼
NORMALIZATION LAYER
    │
    ▼
FINANCIAL ENGINE
    │
    ├── Treasury
    ├── Receivables
    ├── Payables
    ├── Cash Projection
    ├── DRE
    ├── P&L
    ├── EBITDA
    ├── Forecast
    ├── Working Capital
    └── Investment Calculator
    │
    ▼
INTERNAL API
    │
    ▼
FRONTEND / DASHBOARD
```

---

# 17. STACK SUGERIDA

Preferencialmente:

## Frontend

- Next.js
- React
- TypeScript

## Backend

- Node.js
- TypeScript

## Banco

- PostgreSQL

## ORM

Prisma ou equivalente maduro.

## Jobs

Sistema de tarefas agendadas.

## Cache

Somente se necessário.

Não adicionar infraestrutura complexa sem justificativa.

---

# 18. CAMADA RAW

Obrigatório preservar os dados brutos.

Exemplos:

- raw_bling_accounts_receivable
- raw_bling_accounts_payable
- raw_bling_payments
- raw_bling_receipts
- raw_bling_financial_accounts
- raw_nuvem_receivables
- raw_payroll

Nunca alterar silenciosamente um registro RAW.

Dados tratados devem residir em outra camada.

---

# 19. IDEMPOTÊNCIA

Toda sincronização deverá ser idempotente.

A mesma informação importada duas vezes:

**não pode gerar dois registros financeiros.**

Cada fonte deverá ter:

- source_system
- source_id
- imported_at
- updated_at
- payload_hash quando útil
- sync_batch_id

Criar constraint adequada para prevenir duplicidade.

---

# 20. LOG DE SINCRONIZAÇÃO

Criar tabela:

`integration_sync_logs`

Campos mínimos:

- integration
- started_at
- finished_at
- status
- records_read
- records_created
- records_updated
- records_failed
- error_message
- retry_count

---

# 21. TRATAMENTO DE ERROS

Falha de API não pode:

- apagar dados;
- zerar indicador;
- sobrescrever dado válido;
- criar duplicidade.

Se a sincronização falhar:

mostrar:

**Dados desatualizados**

e manter último dado válido.

---

# 22. MODELO DE DADOS — NÚCLEO

Criar inicialmente estruturas equivalentes a:

## Integrações

- raw_bling_*
- raw_nuvem_*
- integration_sync_logs

## Financeiro

- financial_accounts
- receivables
- receipts
- payables
- payments
- daily_balances
- financial_adjustments

## Planejamento

- budgets
- forecasts
- forecast_versions

## Contabilidade

- chart_of_accounts
- accounting_entries
- accounting_periods
- accounting_adjustments

## Gerencial

- dre_snapshots
- pnl_snapshots
- ebitda_snapshots
- working_capital_snapshots

## Tesouraria

- investment_parameters
- investment_calculations

---

# 23. RECEIVABLES

Tabela lógica mínima:

`receivables`

Campos:

- id
- source_system
- source_id
- business_type
- customer_id
- external_document
- due_date
- expected_receipt_date
- gross_amount
- fee_amount
- net_amount
- status
- payment_method
- installment_number
- installment_count
- source_payload_reference
- created_at
- updated_at

`business_type`:

- B2B
- B2C

---

# 24. PAYABLES

Campos:

- id
- source_id
- supplier_id
- document
- category
- competence_date
- due_date
- payment_date
- original_amount
- interest
- penalty
- discount
- open_balance
- status
- cost_center
- source_reference

---

# 25. DAILY BALANCES

Tabela:

`daily_balances`

Campos:

- date
- opening_balance
- inflows
- outflows
- adjustments
- closing_balance
- balance_type
- scenario
- calculated_at

`balance_type`:

- REALIZED
- PROVISIONAL
- PROJECTED

---

# 26. CENÁRIOS

Estruturar desde o início suporte a:

- BASE
- CONSERVATIVE
- STRESS
- OPTIMISTIC

Mesmo que MVP utilize apenas Base inicialmente.

---

# 27. DRE

Não usar DRE pronta do Bling.

Estrutura inicial:

Receita Bruta

(-) Deduções

(-) Tributos sobre vendas

= Receita Líquida

(-) CPV / CMV

= Lucro Bruto

(-) Despesas Comerciais

(-) Despesas Administrativas

(-) Outras Despesas Operacionais

= Resultado Operacional

(+/-) Resultado Financeiro

= Resultado Antes dos Tributos

(-) Tributos sobre Lucro

= Resultado Líquido

Todas as linhas devem ser parametrizáveis por regra de classificação.

---

# 28. P&L

P&L gerencial deverá ser reconciliável com DRE.

Não permitir exclusões silenciosas.

Toda diferença entre DRE e P&L deve possuir:

- adjustment_id
- descrição
- motivo
- valor
- usuário
- data
- aprovação

---

# 29. EBITDA

Fórmula deverá ficar em configuração/regra central.

Mostrar:

- EBITDA
- Margem EBITDA
- Realizado
- Orçado
- Forecast
- Variação

Se houver EBITDA Ajustado:

todo ajuste deve ser rastreável.

---

# 30. BALANCETE CONTÁBIL

Não usar o balancete financeiro do Bling.

Estrutura:

| Conta | Saldo Inicial | Débitos | Créditos | Saldo Final |

O motor contábil será desenvolvido em fase posterior do MVP financeiro.

---

# 31. BALANÇO PATRIMONIAL

Estrutura:

## Ativo

- Circulante
- Não Circulante

## Passivo

- Circulante
- Não Circulante

## Patrimônio Líquido

Validação obrigatória:

**Ativo = Passivo + Patrimônio Líquido**

Se diferença diferente de zero:

status:

**DIVERGENT**

---

# 32. FOLHA

Dados conhecidos incluem departamentos:

- Produção
- Comercial
- Administrativo

Manter separação entre:

- empregados CLT;
- sócios / pró-labore;
- desligados.

Adiantamento salarial:

**não é despesa adicional.**

É compensação financeira contra folha.

PLR:

**não aplicável atualmente.**

---

# 33. VALOR MÁXIMO PARA APLICAÇÃO

Regra oficial deverá reproduzir a calculadora de referência aprovada.

## Lucro disponível

**Lucro Líquido acumulado elegível até M-2  
− Valor já investido**

## Limite financeiro

**Caixa projetado na pior semana  
− Reserva mínima**

A regra atual considera também os recebimentos previstos conforme modelo da calculadora.

## Valor Máximo

**MAX(0; MIN(Limite Financeiro; Lucro Disponível))**

Nenhuma alteração dessa regra poderá ser feita sem aprovação funcional.

---

# 34. ARQUIVO DE REFERÊNCIA

Manter no repositório:

`reference-models/Calculadora_Aplicacao_Lucro_Liquido_Profissional.xlsx`

O sistema deverá possuir testes automatizados comparando casos conhecidos com a planilha.

---

# 35. ORÇADO X REALIZADO

Estrutura mínima:

- account/category
- period
- budget_amount
- actual_amount
- variance_amount
- variance_percent

O sistema deverá mostrar:

**Favorável / Desfavorável**

conforme natureza da linha.

Não assumir que valor maior é sempre favorável.

---

# 36. FORECAST

Manter versões.

Exemplo:

- Forecast Agosto v1
- Forecast Agosto v2
- Forecast Agosto Final

Nunca sobrescrever histórico.

Guardar:

- version
- created_at
- assumptions
- scenario
- user
- values

---

# 37. ALERTAS

Primeiros alertas:

- saldo projetado abaixo da reserva mínima;
- obrigações vencidas;
- recebíveis vencidos;
- grandes pagamentos próximos;
- queda relevante de EBITDA;
- erro de sincronização;
- divergência contábil;
- dado desatualizado.

---

# 38. AUDITORIA

Toda transformação material deverá permitir:

**Dado apresentado  
→ Regra usada  
→ Registro tratado  
→ Registro RAW  
→ Fonte original**

Obrigatório.

---

# 39. SEGURANÇA

Credenciais:

- nunca no código;
- nunca no frontend;
- nunca versionadas no Git;
- utilizar secrets/variáveis de ambiente.

Tokens Bling/Nuvem devem ficar apenas no backend.

---

# 40. ACESSOS

Preparar RBAC.

Perfis iniciais:

- ADMIN
- DIRECTOR
- FINANCE
- ACCOUNTING
- VIEWER

Mesmo que inicialmente todos os perfis não sejam utilizados.

---

# 41. LGPD

Clientes e funcionários poderão conter dados pessoais.

Aplicar:

- mínimo necessário;
- controle de acesso;
- logs;
- criptografia adequada;
- exclusão/anonimização quando aplicável.

---

# 42. PERFORMANCE

Não consultar Bling ou Nuvem diretamente sempre que usuário abrir dashboard.

Fluxo correto:

API externa

→ sincronização

→ banco local

→ cálculo

→ dashboard.

---

# 43. ATUALIZAÇÃO

Proposta inicial:

## Bling

Sincronização incremental periódica.

## Nuvem

Sincronização periódica de recebíveis B2C.

## Dashboard

Exibir:

**Última atualização de cada fonte.**

---

# 44. WIREFRAME

Usar como referência visual o wireframe aprovado da Visão Executiva Majucau.

A implementação deve manter:

- sidebar;
- cards clicáveis;
- estados visuais;
- gráficos;
- alertas;
- drill-down;
- filtros;
- atualização das fontes.

O wireframe é referência de UX, não especificação matemática.

As regras financeiras deste documento prevalecem.

---

# 45. AGENTS.md

Criar `AGENTS.md` na raiz do projeto contendo:

1. não inventar regra financeira;
2. não alterar fórmula sem aprovação;
3. preservar RAW;
4. manter idempotência;
5. criar testes para cálculos;
6. não expor credentials;
7. documentar mudanças;
8. garantir rastreabilidade;
9. priorizar consistência sobre velocidade.

---

# 46. ESTRUTURA DO REPOSITÓRIO

```text
majucau-financial-intelligence/
│
├── AGENTS.md
│
├── README.md
│
├── docs/
│   ├── PRODUCT-SPEC.md
│   ├── ARCHITECTURE.md
│   ├── DATA-SOURCES.md
│   ├── DATA-DICTIONARY.md
│   ├── FINANCIAL-RULES.md
│   ├── ACCOUNTING-RULES.md
│   ├── API-INTEGRATIONS.md
│   ├── UI-SPEC.md
│   └── ACCEPTANCE-TESTS.md
│
├── reference-models/
│   └── Calculadora_Aplicacao_Lucro_Liquido_Profissional.xlsx
│
├── backend/
│
├── frontend/
│
├── database/
│
├── tests/
│
└── scripts/
```

---

# 47. ORDEM DE IMPLEMENTAÇÃO

## FASE 0 — Foundation

- repositório;
- estrutura;
- ambiente;
- banco;
- secrets;
- logs;
- usuários.

## FASE 1 — Bling

- autenticação;
- contas a receber;
- contas a pagar;
- recebimentos;
- pagamentos;
- contas financeiras;
- categorias;
- RAW;
- sincronização.

## FASE 2 — Nuvem

- autenticação;
- recebíveis futuros B2C;
- RAW;
- normalização.

## FASE 3 — Tesouraria

- saldo inicial D-1;
- entradas;
- saídas;
- projeção diária;
- 30 dias;
- 60 dias;
- menor saldo;
- alertas.

## FASE 4 — Dashboard

- Visão Executiva;
- cards;
- gráficos;
- drill-down.

## FASE 5 — Resultado

- DRE;
- P&L;
- EBITDA.

## FASE 6 — Planejamento

- Orçado x Realizado;
- Forecast.

## FASE 7 — Gestão

- Capital de Giro;
- Valor Máximo para Aplicação.

## FASE 8 — Contabilidade

- Plano de contas;
- lançamentos;
- balancete;
- balanço;
- ajustes.

---

# 48. CRITÉRIOS DE ACEITE — MVP FINANCEIRO

## Bling

A soma de contas a pagar do sistema deverá fechar com o Bling para o mesmo filtro/período.

Tolerância:

**R$ 0,01**

---

## B2B

A soma de recebíveis B2B deverá fechar com o Bling filtrado como B2B.

Tolerância:

**R$ 0,01**

---

## B2C

A soma de recebíveis B2C deverá fechar com a Nuvem para o mesmo período.

Tolerância:

**R$ 0,01**

---

## Recebimentos realizados

Deverão fechar com os registros pagos do Bling.

Tolerância:

**R$ 0,01**

---

## Pagamentos realizados

Deverão fechar com os registros pagos do Bling.

Tolerância:

**R$ 0,01**

---

## Saldo diário

Para cada dia:

**Saldo Final = Saldo Inicial + Entradas − Saídas ± Ajustes**

Diferença permitida:

**R$ 0,00**

---

## Continuidade

**Saldo Inicial D+1 = Saldo Final D**

Diferença permitida:

**R$ 0,00**

---

## Investimento

Resultado do sistema deverá reproduzir os casos da calculadora de referência.

Diferença permitida:

**R$ 0,01**

---

# 49. TESTES OBRIGATÓRIOS

Criar testes para:

- duplicidade;
- idempotência;
- datas;
- valores negativos;
- cancelamentos;
- estornos;
- recebível vencido;
- obrigação vencida;
- ausência de sincronização;
- mudança de status;
- projeção diária;
- mudança de cenário;
- cálculo de aplicação;
- arredondamento.

---

# 50. REGRAS QUE O DEV NÃO PODE ALTERAR

Sem aprovação expressa:

- fonte oficial do B2C;
- fonte oficial do B2B;
- fonte dos realizados;
- corte D-1;
- horizonte de 60 dias;
- regra de saldo diário;
- regra de não duplicidade;
- fórmula da calculadora de aplicação;
- estrutura conceitual DRE/P&L;
- nomenclaturas oficiais.

---

# 51. PRIMEIRA ENTREGA SOLICITADA AO SENIOR DEV

Antes de programar integrações, entregar:

### A. Architecture Decision Record

Explicando:

- stack;
- arquitetura;
- banco;
- autenticação;
- sincronização;
- idempotência;
- segurança.

### B. ERD

Diagrama inicial do banco de dados.

### C. API Integration Map

Lista dos endpoints que serão usados no Bling e Nuvem.

### D. Data Dictionary v1

Mapeamento:

Fonte → Campo original → Campo interno → Tipo → Regra.

### E. Plano de implementação

Com:

- fases;
- dependências;
- riscos;
- estimativa relativa;
- ordem.

Nenhum código de produção deve começar antes da revisão desses cinco itens.

---

# 52. DEFINIÇÃO DE PRONTO

Uma funcionalidade só está pronta quando:

- código concluído;
- teste automático;
- teste funcional;
- rastreabilidade;
- fonte identificada;
- regra documentada;
- tratamento de erro;
- logs;
- critério de aceite aprovado.

“Está aparecendo na tela” não significa “está pronto”.

---

# 53. PRINCÍPIO FINAL

O sistema deve ser capaz de responder:

> De onde veio esse número?

e seguir:

**Card  
→ cálculo  
→ regra  
→ registro tratado  
→ registro RAW  
→ sistema de origem.**

Se não for possível fazer esse caminho, a funcionalidade ainda não está pronta.