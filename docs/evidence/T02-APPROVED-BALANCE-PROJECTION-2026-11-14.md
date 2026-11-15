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
- Auditoria do limite de 60 dias: a data-base 14/11 está incluída e as 60 datas vão até 12/01. A série semanal alcança 13/01 (61 datas, para fechar a semana parcial); o retrato diário confirmou zero de entrada e saída em 13/01, portanto os totais dos cards de 60 dias coincidem com os totais da série exibida.
- A semana 22–28/11 foi a única comparada nesta sessão com relatórios enviados. A equivalência dos demais agrupamentos ainda precisa de conferência pela responsável.
- A sincronização mais recente, em 14/11 às 22:39, atualizou os cards para entradas de R$ 248.274,64 (30 dias) e R$ 289.559,53 (60 dias), saídas de -R$ 307.955,72 e -R$ 506.770,38, mínimo de -R$ 56.191,12 em 05/12 e -R$ 156.695,68 em 06/01. O saldo-base permaneceu R$ 62.548,33; a semana 22–28/11 permaneceu igual aos relatórios conferidos.
- O teste visual automatizado confirmou que o filtro de datas da T02 não encurta o gráfico semanal da T01. O subtítulo da tabela explicita as fontes: Contas a Receber líquido e Contas a Pagar por vencimento.
- Auditoria de integridade: o fluxo agora conserva o estado interno da última carga normalizada de cada relatório (Receber e Pagar) em cards, dias, semanas, fechamentos e mínimos. Os valores continuam visíveis, sem o rótulo “Parcial” na tela, conforme orientação da responsável; o estado interno não transforma uma carga interrompida em dado completo. A escolha de status considera a tentativa normalizada mais recente de cada recurso, não lotes antigos da mesma execução. Teste unitário e regressão PostgreSQL adicionados; execução PostgreSQL aguarda o CI após o envio deste commit.

Esses resultados comprovam a correção da abertura da projeção e da semana
22–28/11 observada; não aprovam automaticamente as demais semanas nem concluem
a homologação geral da T02. A responsável ainda fará a conferência semana a
semana. T03 continua bloqueada pela aprovação funcional da T02.

O retrato financeiro local não foi copiado para o código nem para o Git.
