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
- `GET /contas/pagar/{idContaPagar}` documenta a categoria pelo identificador e uma lista de IDs de borderôs relacionados à conta.
- `GET /borderos/{idBordero}` documenta a data do borderô, a categoria e a lista de pagamentos com valor pago, contato e número do documento. O item de pagamento não documenta o ID da conta a pagar.
- `GET /categorias/receitas-despesas` recebe página e limite, mas o schema público da resposta documenta somente `data[]`, sem `hasNext` nem `total`. A listagem atual, corretamente, não considera uma página sem marcador como a última; a consulta de uma categoria pelo ID já pode ser feita quando esse ID vier confirmado no detalhe da conta.

Referências: [OpenAPI oficial — contas a pagar e filtro por pagamento](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), [OpenAPI oficial — categoria no detalhe da conta](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), [OpenAPI oficial — borderôs e valores pagos](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json) e [OpenAPI oficial — categorias de receitas e despesas](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json).

O contrato público mostra um caminho técnico possível, mas ainda falta uma resposta real sanitizada para confirmar como o Bling da Majucau liga pagamentos, borderôs e contas — especialmente em pagamentos parciais, parcelas ou um borderô com mais de um pagamento. Como cada pagamento do borderô não traz um ID da conta no schema documentado, não distribuir seu valor entre contas por aproximação. A documentação pública também não comprova os dados históricos existentes na conta.

## Estado no Majucau

- O adaptador lê e preserva páginas RAW de contas a pagar. Agora também oferece consultas somente de leitura para o detalhe de uma conta a pagar, de um borderô e de uma categoria, além de páginas de categorias financeiras. Os detalhes são devolvidos como JSON original, com limite de resposta, cancelamento e retry seguro; a lista de categorias exige metadados explícitos para decidir que terminou. Testes locais do pacote Bling cobrem esses contratos sintéticos.
- Esses novos métodos ainda não persistem os detalhes no banco, não ligam pagamentos às contas e não alimentam o snapshot. Portanto, não normalizam categoria nem geram valores de CAPEX.
- A consulta atual do painel não calcula nem preenche `CAPEX PAGO NO MÊS` ou `CAPEX ACUMULADO`. Os cards continuam sem valor confirmado.
- A tela e a regra aprovadas permanecem: pagamentos de CAPEX no período filtrado e acumulado desde o início do histórico confiável, em cards informativos fora da tabela e do lucro da DRE.

## Próximo passo técnico seguro

Persistir em RAW os detalhes e categorias, em blocos limitados e com idempotência/transação; estabelecer vínculo auditável entre conta, categoria, borderô, pagamento e data. Respeitar o limite da API ao consultar muitos detalhes. Antes de apresentar números, validar o contrato com uma resposta sanitizada da conta, cobrir pagamento parcial e reconciliar um período do sistema com o relatório oficial. Se a associação entre conta e pagamento não puder ser provada, manter o valor como indisponível.
