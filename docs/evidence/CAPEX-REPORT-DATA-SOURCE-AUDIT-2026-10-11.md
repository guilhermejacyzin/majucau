# Auditoria da fonte de CAPEX nos relatórios — 2026-10-11

## Objetivo e limite

Verificar, em leitura somente, se os relatórios de contas a pagar preservados no material privado têm os campos necessários para alimentar os dois cards de CAPEX já aprovados para a T6. Nenhum arquivo original, tela, fórmula financeira ou registro do Bling foi alterado. Não foram usados dados de acesso à conta.

## O que os arquivos enviados mostram

- Rechecagem visual das cópias atuais enviadas: `Bling - Relatório de Contas a Pagar.pdf` cobre 01/01/2026 a 30/11/2026; `Bling - Relatório de Contas a Pagar Geral.pdf` cobre 01/01/2026 a 30/09/2026. A anotação anterior de 01/09/2026 a 30/09/2026 para o primeiro arquivo não corresponde à cópia atual e fica substituída por esta leitura.
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

## Rechecagem do contrato oficial — 2026-10-13

Consultei novamente a OpenAPI oficial atual do Bling. Ela confirma uma ligação por identificador entre a conta a pagar e seus borderôs: o detalhe de uma conta contém `categoria.id` e `borderos[]`; o detalhe de cada borderô contém `data`, `categoria.id` e `pagamentos[]`, com `valorPago` em cada pagamento. O cadastro de categorias inclui `idCategoriaPai`, o que permite seguir a hierarquia sem comparar nomes de fornecedor ou número de documento.

Isso corrige uma parte da conclusão anterior: não é necessário adivinhar qual borderô está ligado a uma conta, pois a própria conta lista os IDs dos borderôs. O pagamento individual dentro do borderô ainda não traz o ID da conta. Assim, os totais agregados podem ser investigados usando os IDs de borderô e deduplicando cada borderô uma única vez; não se pode repartir um pagamento do borderô entre contas por aproximação.

O caminho de implementação fica mais claro: percorrer as contas pagas; buscar seus detalhes; obter cada borderô referenciado uma única vez; preservar os registros RAW com identidade estável; e validar a categoria e a data do pagamento contra o relatório do Bling. A hierarquia de categoria pode ser confirmada pelos IDs e `idCategoriaPai`. A lista de categorias não documenta um marcador de última página, portanto não deve ser considerada completa só porque uma página veio curta.

Ainda falta confirmar na conta real se `bordero.data` corresponde à data de liquidação exibida no relatório, se a categoria do borderô coincide com a categoria da conta e como ficam pagamentos parciais ou borderôs ligados a mais de uma conta. A primeira sincronização precisa comparar um período com o relatório oficial. Até lá, os cards permanecem sem valores confirmados. A tela e as regras aprovadas não foram alteradas.

Referência: [OpenAPI oficial atual do Bling](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), esquemas `ContasPagarDadosDTO`, `BorderosDadosDTO`, `BorderosPagamentoDTO` e `CategoriasReceitasDespesasDadosBaseDTO`.

## Rota adicional para CAPEX pago — 2026-10-13

A OpenAPI também oferece `GET /caixas`, com filtros por período e IDs de categoria. Cada movimentação inclui ID, direção débito/crédito, situação, data, valor, categoria e origem. A origem distingue, entre outras opções, pagamento por borderô e lançamento ligado a uma conta a pagar. Isso torna essa rota uma candidata direta para calcular **CAPEX pago no mês**, desde que a categoria, a origem e o valor conciliem com o relatório da Majucau.

O schema da lista documenta `data[]`, mas não documenta limite de página, total ou marcador da última página. O adapter atual já lê uma página limitada localmente e usa os nomes documentados dos filtros; ele corretamente não considera uma página curta como prova de conclusão. Ainda não há persistência RAW dessa rota nem inclusão nos cards.

`/caixas` informa movimentação de caixa e não prova, sozinho, o total de CAPEX reconhecido como investimento. Portanto, a rota de caixa pode apoiar o card pago mensal; o card acumulado continua dependendo da fonte de CAPEX reconhecido, do histórico completo e da reconciliação contábil. Não somar os dois conceitos nem tratar uma transferência ou lançamento genérico como CAPEX.

Referência: [OpenAPI oficial atual do Bling — `GET /caixas`](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json) e [changelog oficial de formato de datas de `/caixas`](https://developer.bling.com.br/changelogs).

## Implementação atual e início aprovado do acumulado — 2026-10-26

- Gisele confirmou **01/01/2026 (inclusive)** como início do card `CAPEX ACUMULADO`. A rotina diária às 09:30 já consulta `/caixas` desde essa data até a data da execução e preserva as páginas recebidas como RAW no PostgreSQL.
- Essa carga ainda não converte as movimentações em valores CAPEX nem atualiza o snapshot. O dado RAW, por si só, não prova que uma movimentação é CAPEX nem que representa investimento reconhecido; nenhum valor deve aparecer no card antes da classificação e conciliação aprovadas.
- A sessão do Bling no Chrome não pôde ser inspecionada: a política de acesso do navegador bloqueou a leitura e vedou qualquer tentativa por outra superfície. Não consultei nem alterei dados da conta.
- Os PDFs e o print anteriormente fornecidos continuam sem uma linha que mostre, ao mesmo tempo, categoria CAPEX, data de pagamento/liquidação e valor. Para concluir a integração com dados reais, é necessária uma exportação sanitizada dessas colunas e a conciliação com o relatório do período. Os cards continuam `—`; nenhuma tela ou regra financeira foi alterada.

## Conferência visual complementar dos PDFs enviados — 13/10/2026

- Revisei visualmente as oito páginas das duas cópias atuais dos relatórios. Nas linhas individuais aparecem fornecedor, histórico, número do documento, vencimento, liquidação, situação e valor pago; não há coluna de categoria CAPEX.
- O print do relatório agrupado mostra CAPEX e subcategorias, mas não liga esse agrupamento a cada linha paga dos PDFs. Isso confirma que o resumo não deve ser usado como se trouxesse data e pagamento por item, nem os PDFs como se trouxessem categoria por item.
- Sem essa ligação, os PDFs não bastam para calcular com segurança nem o CAPEX pago no mês nem o acumulado reconhecido desde o início do histórico confiável. Não extraí nem registrei nomes de fornecedores, números de documento ou valores financeiros.
- A suíte local `go test -count=1 ./internal/integrations/bling` passou em 9,494 s e `go vet ./internal/integrations/bling` passou. A variável de banco de integração foi removida apenas no processo de teste, então o teste PostgreSQL de sincronização não foi executado. Isso valida o código sintético do pacote, não a autorização nem a resposta real da conta.
- As imagens temporárias de revisão foram criadas dentro de `artifacts/tmp/` e removidas com a rotina segura do projeto após a inspeção. Nenhum output de revisão permaneceu.

Nenhuma regra financeira, tela ou dado do Bling foi alterado. Os cards continuam sem valores até a integração ser autorizada e os registros poderem ser conciliados.
