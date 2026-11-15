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

Esses resultados comprovam a correção da abertura da projeção e da semana
22–28/11 observada; não aprovam automaticamente as demais semanas nem concluem
a homologação geral da T02. A responsável ainda fará a conferência semana a
semana. T03 continua bloqueada pela aprovação funcional da T02.

O retrato financeiro local não foi copiado para o código nem para o Git.
