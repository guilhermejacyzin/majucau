# Nuvem Pago — tabela de tarifas aprovada

## Proveniência e precedência

- **Fonte:** quatro capturas enviadas e confirmadas pela responsável pelo produto em 2026-09-20.
- **Evidências visuais:** [1x D30](Nuvem-Pago-Taxas-1x-D30-2026-09-20.png), [2x D30](Nuvem-Pago-Taxas-2x-D30-2026-09-20.png), [3x D30](Nuvem-Pago-Taxas-3x-D30-2026-09-20.png) e [boleto/PIX](Nuvem-Pago-Taxas-Boleto-PIX-2026-09-20.png).
- **Regra de precedência:** esta tabela substitui a captura genérica anterior ([arquivo arquivado](Nuvem-Pago-Taxas-2026-09-20.png)); os prazos D+2 e D+14 de cartão da tabela anterior não devem ser usados no produto.
- **Hashes das evidências:** 1x `a32bbac72387f12aeb16d2aa9e8905c21869c96cb99d98b4756c474129c59214`; 2x `f283d0632a14bf633e5417fc06a34f4afd4618ac056888fd3af5eedcbe8c61a8`; 3x `55bc4945c91b00c3ec4d7dfa385d988fb453be7567873cb5bd8847ac5f86cd6e`; boleto/PIX `0b7ae2643539574c5524ae450ed549efb88dfa946a1390aa2b7fe83e7317b329`.
- **Situação:** tarifário aprovado para configuração/estimativa. Ele ainda não é contrato técnico do ledger/API.

## Regras válidas

Todo recebimento de cartão até 3x usa prazo de recebimento D+30. Boleto e PIX mantêm as exceções exibidas nas capturas aprovadas.

| Meio de pagamento | Condição | Prazo de recebimento | Tarifa anunciada | TPV mostrado |
|---|---:|---:|---:|---|
| Cartão de crédito | 1x | D+30 | 2,59% + R$ 0,35 | grátis |
| Cartão de crédito | 2x | D+30 | 4,49% + R$ 0,35 | grátis |
| Cartão de crédito | 3x | D+30 | 5,44% + R$ 0,35 | grátis |
| Boleto | — | D+2 | R$ 2,39 | grátis |
| PIX | — | na hora | 0,99% | grátis |

O painel também exibe Pix com validade personalizável, gestão no painel da loja, checkout transparente e estado **Ativado**.

## Interpretação segura para o produto

- Cartão: `fee = bruto × percentual + R$ 0,35` por transação, com percentual definido pela condição 1x, 2x ou 3x.
- Boleto: tarifa fixa anunciada de R$ 2,39 por transação e prazo D+2.
- PIX: `fee = bruto × 0,0099` e recebimento anunciado “na hora”.
- “TPV grátis” significa apenas que não há cobrança de TPV exibida; não autoriza inferir que antecipação, chargeback, impostos, retenções, ajustes ou outras tarifas sejam zero.
- O prazo anunciado é uma expectativa tarifária. A data efetiva de liquidação de uma transação deve vir do registro financeiro oficial.
- Quando houver ledger oficial com taxa e líquido reais, os valores transacionais prevalecem. A tabela serve como tarifa configurada, estimativa identificada e validação de divergência, nunca como substituto silencioso do fato liquidado.

## O que esta evidência ainda não comprova

As capturas não informam endpoint, autenticação, identificador de transação, parcela, cancelamento, reembolso, chargeback, antecipação, retenção, data efetiva de repasse ou contrato de exportação. Portanto, reduzem a lacuna de **tarifário**, mas não liberam sozinhas o estado financeiro `CONFIRMED` do B2C.

## Contrato técnico a implementar

O adapter Nuvem Pago deve versionar esta tabela por vigência e selecionar meio/condição. Para uma cobrança bruta `B`:

- cartão 1x: `fee = B × 0,0259 + 0,35`, prazo D+30;
- cartão 2x: `fee = B × 0,0449 + 0,35`, prazo D+30;
- cartão 3x: `fee = B × 0,0544 + 0,35`, prazo D+30;
- boleto: `fee = 2,39`, prazo D+2;
- PIX: `fee = B × 0,0099`, prazo imediato;
- `net = B - fee` somente como cálculo tarifário identificado, não como liquidação confirmada.

Toda taxa calculada precisa carregar `pricing_version`, meio, condição, prazo, vigência, origem e estado (`PROVISIONAL` até reconciliação com o ledger). Valores efetivos do ledger prevalecem sobre a tabela.
