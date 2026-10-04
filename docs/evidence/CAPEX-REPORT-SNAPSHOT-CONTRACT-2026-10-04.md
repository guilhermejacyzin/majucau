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

O leitor existente do painel ainda consulta somente recebíveis, recebimentos, contas a pagar e pagamentos. Ele ainda não extrai as métricas de CAPEX do relatório. Por isso, o contrato e a tela estão preparados para recebê-las, mas os valores reais continuarão como `—` até a ligação no backend.

## Revisão dos arquivos recebidos — 2026-10-04

- Os PDFs de contas a pagar mostram lançamentos pagos e seus períodos, mas não mostram a categoria CAPEX por lançamento.
- O relatório geral cobre `01/01/2026 a 30/09/2026` e informa `R$ 125.481,02` no total de contas a pagar. Ele lista fornecedor, histórico, vencimento, liquidação, situação e valor pago; não traz uma coluna de categoria CAPEX.
- Esse total geral não pode ser usado como CAPEX: inclui pagamentos de várias naturezas e não identifica quais são investimentos em ativos.
- O print do resumo por categoria mostra CAPEX e subcategorias, mas a parte com o período escolhido não aparece.
- Assim, ainda não é possível confirmar se o total do print é CAPEX pago no mês ou acumulado, nem reconciliá-lo com o relatório geral de contas a pagar.
- Não usar o total geral de contas a pagar como CAPEX. Para concluir a ligação, falta o mesmo resumo por categoria com as datas do filtro visíveis.
- Nenhum valor, fornecedor ou dado de lançamento dos anexos foi copiado para o repositório.

## Verificação

- `git diff --check`: aprovado.
- Testes de frontend e backend não foram executados nesta etapa.
