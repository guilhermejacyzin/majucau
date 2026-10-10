# CAPEX nos relatórios e cards da DRE — 2026-10-04

## Decisão confirmada

Gisele confirmou que os relatórios existentes informam CAPEX. A T6 mantém dois cards fora da tabela e do lucro da DRE:

- `CAPEX PAGO NO MÊS`: valor de CAPEX pago no mês informado pelo relatório;
- `CAPEX ACUMULADO`: valor acumulado desde o início do histórico confiável.

Os valores devem vir do relatório. Não substituir o dado do relatório por todos os pagamentos do Bling, nem somar CAPEX ao lucro da DRE.

## Ajuste realizado

- O contrato do snapshot Go e o tipo do frontend aceitam as métricas opcionais `capex_paid_month` e `capex_accumulated`.
- Os cards mostram os valores quando o snapshot receber uma métrica `CONFIRMED` ou `PARTIAL` com valor. Ausência, indisponibilidade ou valor projetado continuam aparecendo como `—`.
- Os tooltips e documentos do projeto agora registram que os relatórios têm CAPEX e que a pendência é integração ao snapshot.
- Nenhuma linha da DRE, fórmula ou regra de negócio foi alterada.

## Limite atual

O leitor do painel transmite linhas tipadas de recebíveis, recebimentos, contas a pagar e pagamentos. A sincronização da API Bling atualmente guarda páginas de contas a pagar como evidência RAW; ainda não transforma essas páginas em registros tipados de `payables`/`payments` nem extrai CAPEX. Por isso, o contrato e a tela estão preparados para receber as métricas, mas os valores reais continuarão como `—` até que os campos e a categoria sejam confirmados e o fluxo de integração seja implementado.

## Revisão dos arquivos recebidos — 2026-10-04

- Os PDFs de contas a pagar mostram lançamentos pagos e seus períodos, mas não mostram a categoria CAPEX por lançamento.
- O relatório geral cobre `01/01/2026 a 30/09/2026` e informa `R$ 125.481,02` no total de contas a pagar. Ele lista fornecedor, histórico, vencimento, liquidação, situação e valor pago; não traz uma coluna de categoria CAPEX.
- Esse total geral não pode ser usado como CAPEX: inclui pagamentos de várias naturezas e não identifica quais são investimentos em ativos.
- O print do resumo por categoria mostra CAPEX e subcategorias, mas a parte com o período escolhido não aparece.
- Assim, ainda não é possível confirmar se o total do print é CAPEX pago no mês ou acumulado, nem reconciliá-lo com o relatório geral de contas a pagar.
- Não usar o total geral de contas a pagar como CAPEX. Para concluir a ligação, falta o mesmo resumo por categoria com as datas do filtro visíveis.
- Nenhum valor, fornecedor ou dado de lançamento dos anexos foi copiado para o repositório.

## Rechecagem do relatório e da origem dos dados — 2026-10-04

- O PDF geral recebido nesta continuação confirma o mesmo período (`01/01/2026 a 30/09/2026`) e o mesmo total geral (`R$ 125.481,02`) já registrado acima; ele continua sem coluna de categoria.
- Uma tentativa de abrir o relatório por categoria no Chrome foi bloqueada por uma preferência salva de segurança do navegador. Nenhum contorno foi tentado. Para conferir o período do total `CAPEX`, é necessário que a usuária envie o print do resumo por categoria com o filtro de datas visível, ou libere o acesso diretamente no navegador.
- A leitura local do código confirmou que `ListPayables` e `PayablesAPISyncService` preservam cada página da API como JSON RAW. Não foi encontrado fluxo que normalize páginas da API em `payables` ou `payments`; o painel lê essas tabelas tipadas separadamente. Assim, ainda falta homologar os campos reais do Bling e criar a normalização antes de calcular CAPEX a partir da API.
- O PDF geral não pode substituir essa homologação: ele confirma somente contas pagas no intervalo, não quais pagamentos são CAPEX.

## Verificação

- `git diff --check`: aprovado.
- Testes de frontend e backend não foram executados nesta etapa.

## Rechecagem dos PDFs recebidos — 2026-10-10

- O relatório de contas a pagar do mês cobre `01/09/2026 a 30/09/2026` e informa `R$ 27.902,40` como total pago no período.
- O relatório geral cobre `01/01/2026 a 30/09/2026` e informa `R$ 125.481,02` como total pago no período.
- Os dois relatórios listam data de vencimento, liquidação, situação e valor pago. Nenhum dos dois mostra a categoria CAPEX em cada lançamento.
- Portanto, `R$ 27.902,40` é o total pago de todas as contas no mês, não o CAPEX pago; `R$ 125.481,02` é o total de todas as contas no histórico, não o CAPEX acumulado. Não usar esses valores nos cards.
- Para ligar o relatório ao card, falta um arquivo ou print que mostre, para os mesmos lançamentos, categoria CAPEX, data de pagamento/liquidação e valor; para o acumulado, também é necessário definir o início do histórico confiável conforme a decisão já registrada.
- Nenhum nome de fornecedor ou detalhe de lançamento foi copiado para o repositório. Os PDFs originais não foram alterados.
- A tela continua mantendo `—` sem dado CAPEX confiável; nenhuma tela ou regra de negócio foi alterada.
