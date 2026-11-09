# Auditoria da atualização completa T02 — 09/11/2026

## O que foi observado

A carga oficial terminou às 02:14:12 BRT, depois de 30 minutos, antes de
publicar o retrato completo:

- `contas.receber`: `SUCCESS`, 15.155 registros.
- `caixas.conciliadas`: `SUCCESS`, 3.162 registros.
- `contas.pagar`: leu 104 páginas/10.301 registros; preparou detalhes até a
  página 19 de 104 às 02:13:31 e terminou sem publicar.
- `caixas`: `PARTIAL`, código `BLING_CASH_MOVEMENT_STATUS_MISSING`.
- O registro sanitizado do worker aponta expiração/cancelamento do prazo total
  de 30 minutos. O lote de Contas a Pagar não teve sucesso completo.

Às 02:18 BRT, trouxe o aplicativo Wails real para a frente e conferi sua árvore
de acessibilidade: ele está na T01. A sincronização exibida é 09/11 às 01:58;
Saldo Atual R$ 62.548,33; A Receber R$ 0,00; A Pagar -R$ 688.071,73 com rótulo
“Parcial”; todas as receitas do gráfico estão R$ 0,00 e os saldos semanais são
“—”. A captura de pixels da janela expirou, então a verificação do conteúdo foi
por acessibilidade. Os números não foram homologados.

## Causa e correção do registro de erro

A sincronização relê as contas e depois busca detalhes em blocos de 100
registros para normalizar os campos usados nas telas. A API limita o ritmo das
chamadas; a execução observada levou cerca de 40 segundos para preparar cada
página de Contas a Pagar. Com 104 páginas, 30 minutos não são suficientes. A
publicação completa é atômica, então o lote interrompido preserva o retrato
anterior e os novos dados não aparecem na tela.

O código também classificava erros de contexto (prazo vencido/cancelamento)
durante essa preparação como falha de normalização. Corrigi essa classificação
para `SYNC_DEADLINE_EXCEEDED` ou `SYNC_CANCELLED`, sem gravar detalhes sensíveis.
`go test ./internal/integrations/bling` passou com os casos de regressão.

## Ajuste realizado no código local

Aumentei `BlingRefreshTimeout` para duas horas no contrato IPC compartilhado
por aplicativo e worker. Os comentários e testes usam a mesma constante. O
prazo estendido vale apenas para a leitura completa e somente de leitura; não
altera os prazos dos demais comandos, os relatórios de origem, o desenho ou as
regras financeiras.

## Estado atual e próximos passos

O commit `9384758` já contém o prazo de duas horas, mas o processo aberto usa a
build anterior. A correção do código de classificação descrita acima ainda é
local. Os dados reais da carga interrompida não foram publicados. Portanto, a
T02 não está concluída nem homologada.

1. Publicar a correção de classificação, reconstruir o worker e o aplicativo,
   e iniciar uma única nova carga com o prazo de duas horas.
2. Confirmar que o retrato completo aparece na tela e comparar os períodos com
   Contas a Receber (líquido), Contas a Pagar e o saldo conciliado do Itaú.
3. Abrir a T02 para a aprovação visual; T03 só começa após essa aprovação.
