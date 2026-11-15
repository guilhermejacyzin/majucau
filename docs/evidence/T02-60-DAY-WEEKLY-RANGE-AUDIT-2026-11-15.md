# T02 — auditoria do intervalo semanal e estado da atualização — 15/11/2026

## O que foi corrigido

A tela da T02 montava o intervalo padrão até 14/01/2027, mas os cards de
projeção usam 60 datas contadas de 15/11/2026 a 13/01/2027. Esse dia extra
continha R$ 109,65 em recebimentos e fazia o total da tabela divergir do card.
O intervalo padrão da tabela e do gráfico agora termina em 13/01/2027. O total
líquido de Contas a Receber na tabela, R$ 278.455,46, corresponde ao card de
60 dias.

A semana 22/11–28/11 foi conferida com os relatórios fornecidos: R$ 43.790,66
líquidos em Contas a Receber e R$ 25.275,71 em Contas a Pagar. Essa conferência
não valida automaticamente as demais semanas; a aprovação semana a semana
continua pendente.

## Dados e limites de atualização

- Saldo-base: último saldo conciliado de Itaú disponível, R$ 62.548,33.
- Contas a Receber: última leitura normalizada registrada em 14/11/2026 às
  23:53, com 15.443 registros.
- Contas a Pagar: última carga publicada continua sendo 08/11/2026 às 11:00,
  com 10.301 registros. Uma nova tentativa manual chegou à página 92 de 104,
  mas foi interrompida antes da publicação; os dados preparados não substituem
  o último retrato publicado.
- Por isso, a interface continua mostrando o último retrato válido, mas não se
  declara que os dados de Contas a Pagar estejam atualizados em 15/11.

## Proteção contra tela vazia

A interface conserva o retrato confirmado mais recente no armazenamento local
do navegador. Uma resposta vazia ou indisponível durante a leitura não apaga os
cards e séries; após recarregar, o último retrato válido pode ser restaurado
enquanto o worker responde. Isso evita que uma recarga durante uma chamada
demorada seja apresentada como se não houvesse dados.

## Verificação de implementação

Os testes focados de restauração do retrato financeiro e do alinhamento do
intervalo padrão passaram: 58/58. A janela inicial segue 15/11/2026–13/01/2027,
e a data 14/01 não integra a tabela nem o total de 60 dias.

## Situação funcional

O protótipo atualizado está aberto na T02 — Fluxo de Caixa. A correção do
intervalo e a conferência de 22/11–28/11 estão registradas, mas a responsável
ainda precisa analisar as outras semanas e aprovar a T02. A atualização atual
de Contas a Pagar precisa ser concluída antes de afirmar que os números refletem
o relatório mais recente. T03 permanece aguardando aprovação da T02.
