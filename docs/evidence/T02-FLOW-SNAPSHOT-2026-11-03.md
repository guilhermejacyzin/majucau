# T02 — primeira ligação da série de fluxo ao snapshot — 03/11/2026

## Decisão atendida

A responsável confirmou que a T02 deve usar o resultado do relatório **Fluxo de
Caixa** do Bling. A integração deve consultar apenas as APIs oficiais do Bling;
relatórios CSV/XLS não são fonte de produção. O desenho aprovado da T02 e as
regras financeiras permanecem protegidos.

## O que foi ligado

- O backend lê, por cursor de linhas, candidatos das execuções normalizadas do
  mesmo lote diário: `/caixas`, Contas a Receber, Contas a Pagar, identidade de
  contas e termos de formas de pagamento. A série é limitada a hoje até D+60;
  os cards usam os primeiros 30 dias.
- Só Itaú e Mercado Pago, aprovados como contas do relatório de referência,
  entram na composição. Títulos e movimentos sem conta correspondente deixam
  o resultado parcial e não são somados às cegas.
- Movimentos de caixa só entram quando têm situação regular `R`. Contas em
  aberto usam saldo remanescente; recebimentos também exigem destino, taxa,
  prazo e configuração de dias conhecidos. Saídas mantêm sinal negativo.
- `DashboardSnapshot.cash_flow` carrega os totais de 30 dias e a série diária.
  A T02 liga esses dados aos cards e ao gráfico aprovado; a T01 lê o mesmo
  snapshot compartilhado.

## Limite que continua visível

Esta etapa é uma **composição parcial pela API**, não prova que o resultado já
seja igual ao relatório Bling. A referência pública consultada não documenta
um endpoint pronto para o relatório. Ainda faltam provar a equivalência dos
filtros e do float, tratar prazos em dias úteis com calendário aprovado,
obter o saldo inicial oficial por conta e comparar valores/datas com uma
execução atual do relatório.

Por isso os valores da série são marcados `PARTIAL`. Não se calcula saldo
projetado, menor saldo, saldo transferível ou semana crítica sem o saldo
inicial e sem a regra/entrada oficial necessária. O saldo atual continua
indisponível até vir da fonte aprovada de Caixa e Bancos; não é inferido dos
movimentos.

## Verificações desta etapa

- `go test ./internal/...` — passou localmente.
- `go vet ./internal/dashboard ./internal/financial` — passou.
- `git diff --check` — passou.
- Build TypeScript/Vite e ESLint — passaram numa cópia isolada em
  `artifacts/verification`.
- Testes frontend — 70 testes passaram.
- A cópia de verificação frontend foi mantida em `artifacts/verification/`
  porque a revisão automática bloqueou a exclusão solicitada após a validação.
  O local e o motivo estão listados em
  `docs/evidence/OUTPUT-RETIDO-MANUAL-2026-11-03.md`.
- `TestReadStreamSummaryPostgresRequiresNormalizedDailyStages` — pulado
  localmente; `MAJUCAU_TEST_DATABASE_URL` não está configurada neste ambiente.
- CI 41586134294 — verificadores do backlog/commit, testes Go e integração
  PostgreSQL passaram. O Windows também passou por testes Go, vet e frontend,
  mas falhou no smoke de limpeza do serviço de teste. A causa foi o teste manter
  aberto o identificador do serviço antes de removê-lo; o script de verificação
  agora fecha esse identificador. Aguardar nova CI para confirmar o ajuste.
- Nenhuma chamada nova à conta Bling foi iniciada e o executável Wails aberto
  não foi substituído nesta etapa.

## Próxima sequência

1. Aguardar nova CI confirmar o smoke de serviço corrigido.
2. Atualizar o executável do protótipo, sem concorrer com o worker ativo.
3. Conferir com a responsável o mesmo período, contas Itaú/Mercado Pago,
   float, prazo e opção de incluir pagamentos/recebimentos no relatório Bling.
4. Reexecutar a sincronização diária e comparar a série, mantendo valores
   parciais enquanto houver divergência ou campo não comprovado.
5. Só depois ligar os resultados dependentes de saldo à T02/T01.
