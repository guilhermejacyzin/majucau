# T02: projeções parciais de 30 dias — 03/11/2026

## O que foi ligado

- O backend agora calcula separadamente títulos em aberto do Bling com
  vencimento de hoje até antes de hoje + 30 dias.
- A entrada prevista combina Bling e Nuvem Pago pela regra já aprovada: nome
  normalizado + valor bruto exato, pareamento um a um e data fora da comparação.
- A saída prevista usa o saldo em aberto dos títulos do Bling, mantendo o sinal
  negativo.
- A T02 lê esses campos novos nos cards “ENTRADAS 30 DIAS” e “SAÍDAS 30 DIAS”.
  Eles são marcados “Parcial · A receber/A pagar” porque são componentes do
  relatório de Fluxo de Caixa, não o relatório consolidado completo.
- Corrigida a ponte Wails que não repassava os dois campos novos do snapshot
  para a interface. Antes dessa correção, o backend calculava as métricas, mas
  elas se perdiam antes de chegar ao navegador. Foi incluído teste para manter
  o valor e o estado `PARTIAL` na passagem backend → tela.

## O que permanece pendente

- A série do gráfico ainda não está ligada ao resumo da T02.
- A “Semana crítica”, o menor saldo em 30/60 dias e o saldo atual bancário ainda
  não são calculados por essa entrega.
- O saldo por banco não pode ser deduzido somando movimentos. A fonte oficial
  continua sendo Caixas e Bancos.
- A referência pública da API lista Caixas e Bancos, Contas a Pagar e Contas a
  Receber, mas não lista uma rota própria para o relatório de Fluxo de Caixa.
  Ainda é necessário reproduzir o relatório com as fontes disponíveis e
  comparar o resultado agregado com a captura aprovada, sem usar a captura como
  dado vivo.
- A integração PostgreSQL não foi executada localmente porque
  `MAJUCAU_TEST_DATABASE_URL` não está configurada. Ela e o restante da CI ainda
  precisam validar o SQL em PostgreSQL.

## Verificações feitas

- `go test -count=1 ./internal/dashboard ./internal/application`: passou.
- `go test -count=1 . ./internal/dashboard ./internal/application`: passou,
  incluindo a regressão da ponte Wails.
- `go vet ./internal/dashboard ./internal/application`: passou.
- Frontend isolado: 19 testes da navegação T01/T02 passaram.
- `tsc --noEmit`: passou.
- `vite build`: passou.

Os arquivos temporários de verificação foram criados em `artifacts/` e devem
ser removidos após terminar a execução local; nenhum dado de cliente foi
incluído neste documento.

## Estado da entrega

Esta é uma ligação parcial dos componentes previstos de 30 dias. Não conclui a
T02, não prova os totais do relatório completo e não homologa o protótipo.
