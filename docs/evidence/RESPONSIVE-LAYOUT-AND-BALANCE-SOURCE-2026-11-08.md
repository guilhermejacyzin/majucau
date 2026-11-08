# T02 responsiva e origem do saldo — 08/11/2026

## Solicitação

Gisele mostrou que, em uma janela larga com o menu lateral visível, os valores
nos cards ficam comprimidos. A adaptação anterior dependia da largura total da
janela e não da área restante depois do menu. A solicitação foi corrigir a
disposição em telas menores, preservando as telas aprovadas, e permitir reduzir
o gráfico quando necessário.

## Alteração visual

- A grade passa a usar a largura real da página com consultas de contêiner;
  os cards diminuem de quatro para duas colunas antes de o conteúdo ficar
  estreito, e passam a uma coluna em larguras pequenas.
- Valores, títulos e descrições podem quebrar em linhas sem invadir o card
  vizinho.
- O menu lateral sobe para o topo em janelas de tablet, liberando largura útil.
- A tabela mantém rolagem horizontal quando não cabe; as colunas não se
  sobrepõem. Em telas pequenas, a altura do gráfico também diminui.
- Mudanças são de disposição responsiva; não mudam conteúdo, regra financeira,
  ordem das telas nem elementos aprovados.

## Origem do saldo — auditoria e correção

O saldo-base mostrado anteriormente veio do print enviado/aprovado por Gisele.
Ele foi transcrito para o PostgreSQL local isolado; não foi extraído
automaticamente do extrato pela API. A implementação atual ainda não atualiza
o card automaticamente a partir do último movimento conciliado.

A referência oficial do Bling documenta `situacaoConciliacao` em `GET /caixas`
e paginação `pagina`/`limite`. A auditoria confirmou que `situacao=R` significa
“registrado”, não “conciliado”. A resposta da listagem não contém o estado de
conciliação no item; por isso o fluxo anterior não podia identificar os
movimentos que atualizam o saldo-base.

A implementação agora faz uma segunda leitura paginada com
`situacaoConciliacao=1`, salva os IDs conciliados em um retrato separado da
mesma execução diária e exige os dois retratos para considerar a carga completa.
O card parte do checkpoint Itaú aprovado e avança somente pelos movimentos
conciliados. Para um período iniciado depois da data-base, a abertura da semana
carrega os movimentos conciliados entre a base e o início do período. O estado
permanece **Parcial** até a nova sincronização real e a comparação com o mesmo
período/opções do relatório.

A tabela escrita da tela foi ajustada para as linhas do relatório semanal
aprovadas: saldo inicial, receitas, despesas e saldo acumulado. Linhas de
variação foram removidas da tabela. A referência mostrada cobre 07/11–30/11,
enquanto a captura anterior da tela estava em 08/11–06/01; não se deve comparar
os totais desses períodos como se fossem a mesma consulta. A reprodução via
API segue parcial até conferência de mesmo período e filtros.

Fonte: [referência oficial da API Bling](https://developer.bling.com.br/referencia)
e [changelog oficial sobre o filtro de conciliação](https://developer.bling.com.br/changelogs).

## Estado e validação

- `frontend/src/mockup.css` foi ajustado para larguras reais da área de
  conteúdo.
- A captura posterior mostrou que o primeiro ajuste ainda deixava os rótulos do
  eixo inferior cortados. A correção seguinte ampliou a altura do `viewBox` de
  185 para 210 e moveu as datas para y=155, mantendo margem entre a área do
  gráfico e o fim da legenda. Um teste de interface verifica as datas na T02 e
  na T01 e confere que o texto cabe no `viewBox`.
- A captura de 08/11 mostrou só a data inicial; a data final não aparecia.
  A correção aumenta a área do gráfico de 235 para 270 px no desktop e de 185
  para 220 px em tela pequena. O `viewBox` vai de 210 para 230 unidades; as
  linhas de início e fim ficam em y=158 e y=178, com fonte 10. Isso dá mais
  espaço ao eixo inferior. A captura nativa da janela ainda falhou, então não
  declaro a inspeção visual concluída.
- A suíte frontend passou com 73 testes; o lint/build do frontend, `go test ./...`
  e build do worker passaram. A versão Wails foi reconstruída e aberta na T01.
  A captura nativa da janela falhou
  duas vezes por tempo limite; não foi possível confirmar visualmente o
  resultado na T02 nesta sessão.
- A migration 000024 está no código e nos checksums, mas ainda não foi aplicada.
  A nova carga real de conciliados e a comparação com o relatório ainda não
  ocorreram.
- BK-062 permanece em 30%; projeto em 12,3%. Não houve conclusão nem aceite da
  responsável.

## Correção posterior dos rótulos — 08/11/2026

A referência atualizada do Bling mostra, para cada semana, a data inicial
completa seguida de “a” na primeira linha e a data final completa na segunda.
O eixo do gráfico foi ajustado para mostrar DD/MM/AAAA nas duas linhas e recebeu
mais espaço inferior, com 285 px no desktop e 250 px em telas menores. O mesmo
componente atende T01 e T02. Os testes verificam os rótulos completos e que a
segunda linha termina dentro do SVG. A captura visual da janela Wails expirou;
o aceite visual ainda precisa ocorrer no protótipo reaberto.

## Rastro para auditoria

- Regra e estado do saldo: `docs/evidence/T02-LAST-ITAU-BALANCE-API-2026-11-08.md`.
- Governança e decisões: `docs/META-PROJETO.md` e
  `docs/planejamento/work/backlog_v2.json`.
- Mudança visual: `frontend/src/mockup.css`.
