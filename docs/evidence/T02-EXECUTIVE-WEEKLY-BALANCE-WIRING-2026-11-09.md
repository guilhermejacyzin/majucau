# T02 — ligação do saldo semanal na Visão Executiva

**Data:** 09/11/2026 05:40 BRT  
**Situação:** correção de código aplicada; T02 ainda não homologada.

## Falha identificada

A T02 montava uma série semanal com saldo de abertura e fechamento. A T01
(Visão Executiva) recebia a lista crua de períodos do servidor, sem esses dois
campos. Por isso a linha de saldo acumulado podia aparecer na T02 e faltar na
T01, apesar de ambas mostrarem o mesmo gráfico.

## Correção

- T01 e T02 passam a usar a mesma série semanal selecionada e o mesmo cálculo
  de semana crítica.
- Entradas e saídas continuam vindo respectivamente das telas de Contas a
  Receber (valor líquido) e Contas a Pagar.
- O saldo semanal continua sendo saldo inicial + entradas − saídas; o fechamento
  da semana inicia a próxima.
- Foi removido o sufixo visual “Parcial” dos valores e da origem do gráfico,
  conforme pedido. O estado técnico de completude continua armazenado; não é
  convertido em confirmação nem apagado.

## Conferência e limites

- `npm run typecheck --prefix frontend` passou nesta execução.
- `git diff --check` passou.
- A árvore de acessibilidade do protótipo Wails anteriormente conferida mostrou
  valores e a origem “Contas a Receber e Contas a Pagar”. Nesta execução, o
  processo Wails foi localizado em execução, mas a captura visual expirou duas
  vezes; não foi possível conferir a tela por imagem.
- O snapshot local consultado tem 9 períodos e 61 dias. Contas a Receber está
  `PARTIAL` com 14.526 registros e Contas a Pagar `PARTIAL` com 1.273. A
  semana de 22–28/11 mostra R$ 43.790,66 de entradas líquidas e R$ 25.275,71 de
  saídas, fechando em R$ 1.287,90 a partir do saldo-base registrado. Esses
  valores são do snapshot existente e ainda não provam uma leitura integral
  atualizada.

## Próximos passos

1. Completar/publicar a leitura integral das duas fontes.
2. Comparar cada semana selecionada com os relatórios oficiais de Contas a
   Receber e Contas a Pagar.
3. Confirmar o último saldo conciliado Itaú e a atualização automática após
   futuras conciliações.
4. Conferir a tela e os números com a responsável. Não iniciar T03 até a
   aprovação da T02.
