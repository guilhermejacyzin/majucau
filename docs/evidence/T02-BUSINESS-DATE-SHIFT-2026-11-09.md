# T02 — correção de um dia a menos na projeção — 09/11/2026

## Causa

`readStreamSummary` calcula a data de negócio em `America/Sao_Paulo` e a
representa como data simples à meia-noite UTC. `readBlingCashFlow` aplicava
`domain.BusinessDate` de novo sobre esse valor. Como meia-noite UTC ainda é
21h do dia anterior em São Paulo, a segunda conversão recuava um dia.

Efeito: o período da T02 podia começar um dia antes do escolhido, incluir
títulos vencidos fora da janela e deslocar as semanas do relatório. Isso
explica o dia extra que aparecia no início e alterava os totais.

## Correção

`readBlingCashFlow` agora preserva a data comercial já normalizada e a usa como
início da série. A janela segue com 61 datas (hoje e os próximos 60 dias), e os
cards de 30 e 60 dias são calculados a partir da mesma data inicial. As fontes
aprovadas permanecem Contas a Receber (valor líquido), Contas a Pagar (saldo
em aberto) e o saldo conciliado do Itaú como base separada.

## Verificação

- `go test ./internal/dashboard -count=1`: passou.
- `TestReadStreamSummaryPostgresRequiresNormalizedDailyStages` contra o
  PostgreSQL 18 isolado do protótipo: passou.
- O teste confirma a primeira data da série como 19/10/2026 para o cenário de
  19/10 e confirma 61 posições, além dos totais de 30/60 dias após o ajuste.
- Migration 000024 e privilégios de runtime estão aplicados ao banco isolado;
  a verificação `verify_runtime.sql` passou.

## O que ainda não foi comprovado

O aplicativo nativo foi reconstruído e reaberto com esta correção. A janela
está ativa na tela inicial, mas a captura para confirmar visualmente a T02
expirou; portanto, ainda falta a conferência visual da T02. A comparação com
uma nova carga real e com os filtros do relatório do Bling continua pendente.
O último saldo-base autorizado é R$ 62.548,33 em 06/11/2026; a
atualização desse valor depende de uma sincronização real, normalizada e
concluída de Caixa e Bancos com movimentos conciliados do Itaú.
