# Auditoria da fonte de CAPEX nos relatórios — 2026-10-11

## Objetivo e limite

Verificar, em leitura somente, se os relatórios de contas a pagar preservados no material privado têm os campos necessários para alimentar os dois cards de CAPEX já aprovados para a T6. Nenhum arquivo original, tela, fórmula financeira ou registro do Bling foi alterado. Não foram usados dados de acesso à conta.

## O que os arquivos enviados mostram

- O relatório geral cobre 01/01/2026 a 30/09/2026; o outro relatório cobre 01/09/2026 a 30/09/2026.
- As linhas exibem vencimento, liquidação, situação e valor pago, além dos dados de identificação da conta.
- Esses PDFs não exibem a categoria de cada conta na mesma linha. Portanto, sozinhos, permitem observar pagamentos e datas, mas não identificar com segurança quais deles pertencem a CAPEX.
- O print de resumo por categoria enviado na conversa mostra CAPEX e subcategorias, mas não deixa visível o período e não relaciona cada total aos pagamentos e datas dos PDFs. Ele confirma que há uma categoria CAPEX nos relatórios, mas não permite reconciliar os dois arquivos linha a linha.

Esta auditoria não conclui que o Bling não tenha a informação. Conclui somente que os PDFs enviados não contêm, juntos e de forma rastreável, todas as colunas necessárias para calcular os cards.

## Contrato público da API Bling

Na OpenAPI pública consultada em 11/10/2026:

- `GET /contas/pagar` aceita filtros de data de pagamento e é paginado; o contrato da resposta de lista documenta campos básicos, sem a categoria.
- `GET /contas/pagar/{idContaPagar}` documenta a categoria pelo identificador e os IDs dos borderôs ligados à conta.
- `GET /borderos/{idBordero}` documenta a data do borderô, a categoria e os pagamentos com o valor pago.
- `GET /categorias/receitas-despesas` fornece categorias paginadas, inclusive identificadores e descrições para resolver o nome da categoria.

Referências: [OpenAPI oficial — contas a pagar e filtro por pagamento](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), [OpenAPI oficial — categoria no detalhe da conta](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), [OpenAPI oficial — borderôs e valores pagos](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json) e [OpenAPI oficial — categorias de receitas e despesas](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json).

O contrato público mostra um caminho técnico possível, mas ainda falta uma resposta real sanitizada para confirmar como o Bling da Majucau liga pagamentos, borderôs e contas — especialmente em pagamentos parciais ou em parcelas. A documentação pública não comprova os dados históricos existentes na conta.

## Estado no Majucau

- O adaptador atual lê e preserva páginas RAW de contas a pagar; não normaliza categoria e borderôs para gerar os valores de CAPEX.
- A consulta atual do painel não calcula nem preenche `CAPEX PAGO NO MÊS` ou `CAPEX ACUMULADO`. Os cards continuam sem valor confirmado.
- A tela e a regra aprovadas permanecem: pagamentos de CAPEX no período filtrado e acumulado desde o início do histórico confiável, em cards informativos fora da tabela e do lucro da DRE.

## Próximo passo técnico seguro

Implementar a leitura somente de recursos oficiais, com paginação, limite de chamadas, retry seguro, streaming e vínculo auditável de conta, categoria, pagamento e data. Antes de apresentar números, validar o contrato com uma resposta sanitizada da conta, cobrir pagamento parcial e reconciliar um período do sistema com o relatório oficial. Se a associação entre conta e pagamento não puder ser provada, manter o valor como indisponível.
