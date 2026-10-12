# Revisão da estrutura dos relatórios de CAPEX — 11/10/2026

## Resultado da revisão dos arquivos enviados

Foram revisados os dois PDFs de Contas a Pagar enviados pela responsável. As
tabelas mostram fornecedor, histórico, número do documento, vencimento,
liquidação, situação e valor pago, mas não mostram a categoria de cada
lançamento.

O print separado confirma que o relatório do Bling apresenta a categoria
CAPEX e suas subcategorias, agrupadas. O print não mostra a data do filtro nem
liga cada pagamento à sua categoria.

## Correção da conclusão anterior — documentação oficial da API

A conclusão anterior ficou incompleta ao pedir outro relatório. A categoria
CAPEX existe nos relatórios enviados, e a API oficial do Bling documenta uma
rota que permite consultar movimentações por data e categoria.

Na documentação atual, `GET /caixas` aceita `dataInicial`, `dataFinal` e
`idsCategorias`. Cada resultado da lista tem um identificador; `GET
/caixas/{idCaixa}` expõe data do lançamento, valor, débito/crédito, categoria e
origem. `GET /categorias/receitas-despesas` expõe descrição e relação com a
categoria pai. Isso permite buscar o cadastro da categoria CAPEX e suas
subcategorias, e então associar as movimentações à categoria sem adivinhar por
fornecedor ou histórico.

Referências oficiais consultadas em 11/10/2026:

- [Referência OpenAPI do Bling — movimentações de caixa e bancos](https://developer.bling.com.br/referencia#/Caixas%20e%20Bancos)
- [Referência OpenAPI do Bling — categorias de receitas e despesas](https://developer.bling.com.br/referencia#/Categorias%20-%20Receitas%20e%20Despesas)

## Consequência para os cards aprovados

Os arquivos PDF de contas a pagar continuam insuficientes, isoladamente, para
ligar cada pagamento à categoria: a lista detalhada não mostra a categoria.
Porém, não é mais correto dizer que falta outro relatório como condição para
avançar. O caminho técnico é integrar a API de movimentações e o cadastro de
categorias do Bling, processando páginas e detalhes em fluxo, de modo
idempotente e com limite por resposta.

Antes de preencher os cards, ainda é necessário homologar os IDs e a
hierarquia real de CAPEX, confirmar que as movimentações retornadas conciliam
com o relatório do Bling e integrar a leitura dessa fonte ao armazenamento e
ao snapshot. Até então, os
valores continuam `Dado indisponível`; não são substituídos por zero nem
estimados.

## Limites da revisão

Esta revisão confirma a estrutura dos dois PDFs e do print enviados, e a
existência dos campos e filtros descritos na documentação pública da API. Ela
não confirma os IDs de categoria da conta Majucau, nem que os movimentos
retornados conciliam com os relatórios reais. A situação da BK-067 permanece
em 20%: a interface existe e a fonte técnica foi identificada, mas a integração
com armazenamento/snapshot, o cálculo e a conferência ainda não foram
implementados.

## Atualização de implementação — leitura de movimentações

O adaptador agora possui `ListCashMovements` para ler uma página de `GET
/caixas` com `pagina`, `limite`, `dataInicial` e `dataFinal`, e
`GetCashMovement` para consultar `GET /caixas/{idCaixa}`. Ambos usam o cliente
HTTP existente: Bearer/JWT, cancelamento, limite de resposta e retry seguro
para `GET`; os registros seguem como JSON bruto. A página sem metadados
explícitos continua com término desconhecido, e uma página curta não é tratada
como lista completa.

Essa mudança implementa somente a leitura técnica do recurso. Ela não persiste
movimentações, não mapeia categoria, não calcula CAPEX e não preenche os cards.
Ainda faltam confirmar IDs/árvore CAPEX na conta, o formato do filtro por IDs de
categoria, a paginação real e a conciliação com os relatórios fornecidos. Até
isso, os valores dos cards permanecem indisponíveis.

Nenhum valor financeiro, nome de fornecedor, identificador ou credencial foi
copiado para este registro.

## Revisão da especificação pública de paginação — 12/10/2026

Foi consultado o OpenAPI oficial atual do Bling. Para `GET /caixas`, ele lista
`pagina`, `dataInicial`, `dataFinal` e `idsCategorias` (array de inteiros). A
resposta documentada contém `data` como uma lista de lançamentos; o contrato
publicado não descreve nessa rota o parâmetro `limite`, nem um total de páginas
ou indicador de próxima página. Já `GET /categorias/receitas-despesas` lista
`pagina` e `limite` (padrão 100), mas também documenta a resposta como `data`
sem metadados de continuação.

Isso revela uma diferença que precisa ser resolvida antes da integração: o
adaptador atual envia `limite` para `/caixas` e depende de uma indicação de
continuação para declarar a leitura completa. A especificação não confirma
esses dois detalhes. Não foi alterado o código nem adotada a hipótese de que
página curta ou vazia encerra a leitura, para evitar que uma sincronização
incompleta seja tratada como sucesso. A lista de categorias confirma que há
filtro por IDs de categoria, mas não confirma os IDs e a hierarquia CAPEX da
conta Majucau.

Referências oficiais consultadas em 12/10/2026:

- [OpenAPI oficial atual do Bling](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json)
- [Changelog do Bling — formato de data em `/caixas`](https://developer.bling.com.br/changelogs)
- [Migração para autenticação JWT](https://developer.bling.com.br/migracao-jwt)

O navegador não permitiu consultar a conta: a política salva bloqueou acesso a
`https://www.bling.com.br`. A permissão anteriormente exibida para
`https://developer.bling.com.br` cobre a documentação, que é outro endereço.
Portanto, não foi possível confirmar a resposta real, os IDs de categoria ou a
conciliação dos relatórios. A BK-067 continua em 20% e a integração CAPEX não
deve preencher os cards até a homologação.

Nenhuma tela aprovada, regra financeira, dado real ou credencial foi alterado
ou registrado.

## Correção da consulta de movimentações — 12/10/2026

A revisão do OpenAPI oficial confirmou que `GET /caixas` não documenta o
parâmetro `limite`. O adapter deixou de enviá-lo e agora chama o campo interno
de segurança `MaxRecords`: ele limita a quantidade máxima de registros aceitos
localmente, sem acrescentar um parâmetro não documentado à API. Datas e página
continuam na consulta; o filtro por categoria segue pendente de homologação.

A ausência de metadados de continuação ainda impede declarar a leitura completa.
Um `total` sem o tamanho real da página não é suficiente para inferir o fim;
limite local não é usado como se tivesse sido enviado ao Bling. O código
continua falhando com segurança quando não há prova de fim da paginação. Nada
foi persistido, classificado ou usado para preencher os cards.
O navegador mostra a tela de login do Bling; nenhum dado real foi consultado.

## Nova tentativa de consulta da conta — 12/10/2026

A lista de permissões do Google Chrome enviada pela responsável mostra
`https://www.bling.com.br` como “Sempre permitir”. Mesmo assim, a ferramenta de
navegação recusou a leitura da aba da conta com a mensagem de que uma preferência
salva bloqueia o acesso. A recusa aconteceu antes de carregar ou ler qualquer
dado da conta. Não tentei contornar a barreira por outra aba ou por comandos de
navegador.

Os PDFs foram conferidos novamente em 12/10/2026. Os dois trazem datas de
liquidação e valores pagos, mas não a categoria por lançamento; o print mostra
CAPEX somente agrupado, sem período ou vínculo com os documentos. Assim, a
conclusão operacional permanece: não calcular nem preencher os cards até a
categoria e o pagamento poderem ser conciliados com segurança.
