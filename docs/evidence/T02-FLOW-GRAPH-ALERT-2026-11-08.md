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
- Protótipo desktop atualizado na T02. A conferência da árvore da tela confirmou que o link do gráfico e os dois itens removidos da Semana Crítica não aparecem; os dados parciais de entradas e saídas continuam presentes.
- Resultados monetários negativos na T02 recebem destaque vermelho na tabela, nos cards que os exibem, na semana crítica, nos pontos/valores do gráfico e no tooltip.
- A série do saldo resultante permanece sem valor porque a abertura não foi entregue no snapshot atual. A regra está implementada, mas a tela não inventa abertura nem saldo acumulado.
- A captura visual nativa não foi gerada nesta revisão. O protótipo está aberto na T02 para conferência/aprovação da responsável.
