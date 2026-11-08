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
