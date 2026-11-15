# T02 — saldo conciliado como abertura da projeção — 14/11/2026

## Problema observado

O card **Saldo atual** mostrava o checkpoint conciliado aprovado de Itaú, mas
**Menor saldo 30 / 60 dias** ficava em branco. A série semanal podia ter
fechamentos calculados separadamente, deixando os resultados inconsistentes
entre os cards e o gráfico.

## Causa confirmada

`readLatestItauBalance` mantinha o saldo conciliado em `Metric`, mas iniciava
`Projection` como `UNAVAILABLE`. Só preenchia a projeção quando encontrava um
lote normalizado de conciliações mais recente. Assim, um checkpoint aprovado e
válido não servia de abertura para as projeções enquanto não houvesse outra
conciliação processada.

## Correção

A projeção agora começa pelo mesmo checkpoint aprovado de Itaú usado no card.
Quando há um lote posterior de conciliações disponível, o delta conciliado
continua avançando essa base. Sem checkpoint aprovado, o saldo e as projeções
continuam indisponíveis; não há valor presumido.

## Verificação

- Teste da unidade `internal/dashboard`: aprovado.
- Teste de integração PostgreSQL `TestReadStreamSummaryPostgresRequiresNormalizedDailyStages`: aprovado no banco de desenvolvimento isolado; o caso novo confirma que o checkpoint aprovado alimenta os mínimos de 30/60 dias antes da próxima conciliação.
- Protótipo Wails reaberto pelo inicializador oficial e conferido na T02 pela árvore de acessibilidade.
- Saldo atual: R$ 62.548,33.
- Menor saldo projetado em 30 dias: -R$ 57.998,81 em 05/12; folga sobre a reserva: -R$ 107.998,81.
- Menor saldo projetado em 60 dias: -R$ 159.418,96 em 06/01; folga: -R$ 209.418,96.
- Semana 22–28/11 exibida: receita líquida R$ 43.790,66 e despesa R$ 25.275,71, iguais aos relatórios enviados para essa semana.
- **Retificação em 15/11:** a auditoria original do limite de 60 dias estava incorreta. A T02 incluía 14/01/2027 além da janela dos cards, adicionando R$ 109,65 em recebimentos. A data final padrão foi corrigida para 13/01/2027, totalizando as mesmas 60 datas dos cards; o total da tabela agora é R$ 278.455,46 e coincide com o card de 60 dias. Veja [auditoria do intervalo e da tentativa de atualização](T02-60-DAY-WEEKLY-RANGE-AUDIT-2026-11-15.md).
- A semana 22–28/11 foi a única comparada nesta sessão com relatórios enviados. A equivalência dos demais agrupamentos ainda precisa de conferência pela responsável.
- A sincronização mais recente, em 14/11 às 22:39, atualizou os cards para entradas de R$ 248.274,64 (30 dias) e R$ 289.559,53 (60 dias), saídas de -R$ 307.955,72 e -R$ 506.770,38, mínimo de -R$ 56.191,12 em 05/12 e -R$ 156.695,68 em 06/01. O saldo-base permaneceu R$ 62.548,33; a semana 22–28/11 permaneceu igual aos relatórios conferidos.
- O teste visual automatizado confirmou que o filtro de datas da T02 não encurta o gráfico semanal da T01. O subtítulo da tabela explicita as fontes: Contas a Receber líquido e Contas a Pagar por vencimento.
- Auditoria de integridade: o fluxo agora conserva o estado interno da última carga normalizada de cada relatório (Receber e Pagar) em cards, dias, semanas, fechamentos e mínimos. Os valores continuam visíveis, sem o rótulo “Parcial” na tela, conforme orientação da responsável; o estado interno não transforma uma carga interrompida em dado completo. A escolha de status considera a tentativa normalizada mais recente de cada recurso, não lotes antigos da mesma execução. Teste unitário e regressão PostgreSQL adicionados; execução PostgreSQL aguarda o CI após o envio deste commit.

## Atualização posterior — 15/11/2026

Os totais de 22:39 acima são históricos, não os dados mais recentes disponíveis.
Contas a Receber foi normalizado em 14/11 às 23:53. Contas a Pagar não recebeu
uma nova publicação: a última carga publicada continua sendo 08/11 às 11:00; a
tentativa manual mais recente foi interrompida na página 92/104 antes da
publicação. O protótipo mantém o último retrato válido visível quando a leitura
falha ou retorna vazia. Consulte a evidência de 15/11 para a janela atual e a
frescura dos dados.

Esses resultados comprovam a correção da abertura da projeção e da semana
22–28/11 observada; não aprovam automaticamente as demais semanas nem concluem
a homologação geral da T02. A responsável ainda fará a conferência semana a
semana. T03 continua bloqueada pela aprovação funcional da T02.

O retrato financeiro local não foi copiado para o código nem para o Git.
