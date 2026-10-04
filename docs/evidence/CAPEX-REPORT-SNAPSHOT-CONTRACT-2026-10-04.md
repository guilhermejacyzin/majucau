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

## Verificação

- `git diff --check`: aprovado.
- Testes de frontend e backend não foram executados nesta etapa.
