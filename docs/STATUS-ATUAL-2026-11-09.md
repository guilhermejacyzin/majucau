# Status atual — 09/11/2026, 02:36 BRT

## Correção aplicada: receita líquida da T02

O protótipo mostrava entradas R$ 0,00 apesar de Contas a Receber ter dados. A
causa confirmada foi a execução completa anterior ter expirado durante Contas a
Pagar antes de atualizar as formas de pagamento. A consulta exigia taxas da
mesma execução e convertia o valor líquido ausente em zero.

Para 22–28/11, os 146 títulos do snapshot real somam R$ 45.482,22 brutos. A
última configuração oficial válida de taxas cobre os 146 e reproduz R$ 43.790,66
líquidos, igual ao total do relatório do Bling enviado pela responsável. O
código agora usa a tabela de taxas oficial mais recente bem-sucedida da mesma
conta e prioriza a nova consulta de taxas antes da longa preparação de Contas a
Pagar.

Testes focados de dashboard/PostgreSQL, Bling e worker passaram. Às 02:36
reconstruí e abri a T02 no protótipo real: a semana 22–28/11 mostra R$ 43.790,66
de receitas e -R$ 25.275,71 de despesas, os mesmos totais dos relatórios
enviados. O saldo acumulado mostrado para a semana é R$ 898,42, calculado pela
abertura semanal exibida mais as entradas líquidas e as saídas negativas.

A sincronização nova de Contas a Pagar terminou `FAILED`/`RAW_ONLY` às 02:14,
então a tela preserva o último snapshot normalizado dessa fonte. A semana
conferida bate, mas outras semanas e uma nova carga completa ainda precisam ser
verificadas antes de homologar a T02. T03 não começou e aguarda aprovação da T02.

## Em que ponto está

O protótipo Wails real está aberto na T02 — Fluxo de Caixa. O cabeçalho mostra
última sincronização em 09/11 às 02:31 e saldo bancário R$ 62.548,33. A semana
22–28/11 já mostra receitas líquidas R$ 43.790,66 e despesas -R$ 25.275,71,
iguais aos dois relatórios enviados. Isso valida a correção dessa semana, mas
não homologa toda a série: a nova carga de Contas a Pagar não terminou e as
demais semanas ainda precisam ser comparadas.

## Resultado observado na sincronização

- Contas a Receber: `SUCCESS`, 15.155 registros lidos.
- Movimentos conciliados: `SUCCESS`, 3.162 registros lidos.
- Contas a Pagar: 104 páginas/10.301 registros lidos; detalhes preparados até
  19/104. A execução terminou às 02:14:12, sem publicar a carga completa.
- Movimentos gerais do caixa: `PARTIAL`, código
  `BLING_CASH_MOVEMENT_STATUS_MISSING`.

O lote de Contas a Pagar não ficou `RUNNING`: expirou após 30 minutos. A carga
completa não foi publicada, por isso a tela continua com o retrato anterior.

## Correção técnica da carga anterior

O commit `9384758` no GitHub aumenta o prazo compartilhado de atualização para
duas horas. O classificador de cancelamentos e prazos excedidos também foi
corrigido e publicado no commit `a787a29`. A nova build está aberta e mostra a
correção do líquido. A carga que terminou às 02:14 ocorreu antes da build nova e
permanece sem publicação completa.

## Saldo e regras aprovadas

Último valor-base aprovado: R$ 62.548,33 em 06/11/2026, informado a partir do
saldo conciliado do Itaú. O saldo atual usa apenas Itaú. A projeção do fluxo usa
Contas a Receber pelo valor líquido, Contas a Pagar pelas saídas e a base
conciliada do Itaú para abrir o saldo. Receitas − despesas não substituem o
saldo bancário atual.

## O que falta antes de aprovar a T02

1. Publicar a correção atual de seleção/ordem das taxas e reconstruir o worker.
2. Executar uma carga completa atualizada e confirmar publicação de Contas a
   Pagar, Contas a Receber, taxas e movimentos.
3. Conferir as semanas selecionadas contra os relatórios correspondentes de
   Contas a Receber (líquido) e Contas a Pagar.
4. Confirmar o saldo da primeira semana e os fechamentos semanais com a base
   conciliada do Itaú.
5. Aguardar a aprovação visual e numérica de Gisele para encerrar a T02.

Somente depois da aprovação da T02 começa a T03. Percentuais antigos de BK-062
e do projeto não foram recalculados e devem ser tratados como histórico até a
revisão do backlog inteiro.
