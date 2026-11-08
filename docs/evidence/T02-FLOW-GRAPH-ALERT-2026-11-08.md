# T02 — gráfico, saldo resultante e alerta — 08/11/2026

## Pedido aplicado

- Remover o link “Ver fluxo completo” do painel do gráfico.
- Remover do quadro Semana Crítica o aviso “Posição de caixa da semana: SEM DADOS” e o link “Ver detalhes”.
- Restaurar o tamanho maior do gráfico: 235 px de área de desenho e painel com altura mínima de 400 px.
- Mostrar o alerta já existente de forma pequena e vermelha.
- Destacar resultados negativos em vermelho.
- Ajustar a sequência visual à referência Bling: Receitas azul, Despesas vermelho, Saldo laranja.

## Regra semanal registrada

`Saldo final = saldo inicial + receitas - despesas`

O saldo final de uma semana inicia a seguinte. A operação é feita em decimal
exato. Se faltar abertura confirmada ou entrada/saída necessária, a tela não
estima nem substitui por zero. O saldo atual de Caixa e Bancos → Itaú permanece
um indicador independente da série do gráfico.

## Arquivos

- `frontend/src/mockupDashboard.tsx`
- `frontend/src/mockup.css`
- `AGENTS.md`
- `G0-DECISION-REGISTER.md`
- `docs/META-PROJETO.md`

## Validação

- Build do frontend (`tsc && vite build`): concluído com sucesso.
- Suíte frontend: 73 testes passaram; lint frontend, `go test ./...`, checksum de migrations e build do worker passaram. A versão desktop Wails foi reconstruída e aberta na T01.
- Testes de interface cobrem a ausência do link removido, as datas na T02 e a margem do texto dentro do SVG. O filtro do relatório começa no dia atual de Brasília (08/11/2026).
- Resultados monetários negativos na T02 recebem destaque vermelho na tabela, nos cards que os exibem, na semana crítica, nos pontos/valores do gráfico e no tooltip.
- A série do saldo resultante permanece sem valor porque a abertura não foi entregue no snapshot atual. A regra está implementada, mas a tela não inventa abertura nem saldo acumulado.
- A captura visual nativa não foi gerada nesta revisão. A janela do protótipo inicia na T01; a captura da janela falhou duas vezes por tempo limite, então a conferência visual T02 ainda está pendente.

## Atualização posterior — 08/11/2026

A responsável enviou uma captura posterior mostrando que os rótulos de data
continuavam cortados. O reposicionamento para y=142 não garantiu margem visual
suficiente; a versão atual amplia o `viewBox` do SVG para 210 e usa y=155 nos
rótulos. Os testes verificam o eixo em T02 e T01. Frontend e worker foram
reconstruídos; a captura da janela Wails falhou duas vezes, então a verificação
visual final do eixo na T02 continua pendente. A migration 000024 e a nova
sincronização real também permanecem pendentes. Ver também
`docs/evidence/RESPONSIVE-LAYOUT-AND-BALANCE-SOURCE-2026-11-08.md`.

## Ajuste adicional após nova conferência da responsável

A responsável informou que as datas ainda apareciam cortadas. A captura de
referência mostra o padrão exato: data inicial DD/MM/AAAA com “a” na primeira
linha e data final DD/MM/AAAA na segunda. O SVG agora exibe essa faixa completa,
com margem inferior adicional para impedir o corte da segunda linha. A suíte
confirma o texto, as duas linhas e o limite inferior do `viewBox`. O build Wails
foi reconstruído e uma nova janela foi aberta; a captura da janela expirou,
portanto a conferência visual e o aceite continuam pendentes. A migration
000024 e a sincronização real não foram executadas nesta atualização.

## Ajuste final dos rótulos

A captura seguinte ainda mostrava apenas o início da data semanal. Removi o
`tspan` e passei a desenhar o dia inicial e o dia final em dois elementos SVG
`text` separados, com posições verticais explícitas. O mesmo componente atende
Visão Executiva e Fluxo de Caixa; os testes conferem as duas linhas nas duas
telas. O Wails foi reconstruído novamente. A captura automatizada falhou por
tempo limite, então não declaro conferência visual concluída.

## Correção da interpretação semanal — 08/11/2026

A responsável apontou que o eixo parecia mostrar dias individuais. A causa foi
a apresentação anterior: início e fim da faixa foram desenhados em linhas ou
elementos separados, embora cada ponto represente um período semanal. A regra
correta é um ponto por faixa semanal e um único rótulo de intervalo abaixo
dele, no formato `DD/MM/AAAA a DD/MM/AAAA` quando houver espaço; em séries
densas o eixo pode abreviar para `DD/MM–DD/MM`, mantendo as datas completas
no detalhe/tooltip. O rótulo fica acima da margem inferior do gráfico. A
primeira faixa pode ter menos de sete dias se o período
selecionado começar no meio da semana; isso reproduz os limites do relatório,
não uma série diária.

O componente compartilhado de T01/T02 agora renderiza um único elemento de
texto por período, e os testes verificam quantidade, conteúdo e posição. O
agrupamento continua semanal: a primeira faixa começa na data selecionada e
termina no sábado; as seguintes seguem domingo a sábado. A aprovação visual
continua pendente até Gisele conferir o protótipo reaberto.

Na sequência, Gisele esclareceu que a linha amarela do print é o marcador
vertical da semana sob o ponteiro, que deve cruzar a área do gráfico e alinhar
com os três pontos daquela faixa. O tooltip já lista Despesas, Receitas e
Saldo; o marcador vertical foi acrescentado atrás dos pontos e antes da caixa
de detalhes. A validação automatizada verifica posição, altura e presença do
tooltip. A captura nativa continua sem funcionar, portanto a conferência visual
e o aceite ficam pendentes no protótipo reaberto.

**Aceite da correção:** em 08/11/2026, Gisele respondeu “Ok, deu certo!” após
conferir a linha de acompanhamento. Este aceite é somente da apresentação
semanal e do marcador vertical; não homologa os valores da API, o saldo bancário
nem conclui BK-062/T02.
