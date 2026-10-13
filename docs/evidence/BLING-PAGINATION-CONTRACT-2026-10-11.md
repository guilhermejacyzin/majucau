# Evidência — contrato público de paginação do Bling

Data da consulta: 11/10/2026.

## O que a documentação oficial confirma

- A referência OpenAPI pública descreve `GET /contas/receber` como uma lista paginada e inclui os parâmetros `pagina` e `limite`.
- O FAQ do Bling informa que o padrão é até 100 registros por requisição e orienta informar `pagina` para percorrer os resultados.
- Para `GET /contas/receber`, o schema público da resposta descreve um objeto cuja propriedade `data` é uma lista de contas. O schema não declara `hasNext`, `total`, página atual, nem outra informação para identificar a última página.
- O endpoint declara `tipoFiltroData`, `dataInicial` e `dataFinal`. A referência não declara `dataAlteracaoInicial` nem `dataAlteracaoFinal` para essa rota. A página de limites menciona esses nomes como filtros que existem em APIs do Bling, mas isso não confirma que estejam disponíveis em toda rota.
- O filtro de situação não tem o mesmo nome nas duas rotas: contas a receber usa `situacoes[]` (lista numérica); contas a pagar usa `situacao` (número único). O adapter enviava `situacao` para ambas.
- A data também tem contratos diferentes: contas a receber escolhe um tipo (`E` emissão, `V` vencimento ou `R` recebimento) por `tipoFiltroData` e usa `dataInicial`/`dataFinal`; contas a pagar tem pares separados para vencimento e pagamento.

## O que isso significa para o Majucau

A paginação por número de página está documentada, mas a documentação pública consultada não permite decidir, sem suposição, quando a última página foi alcançada nem estabelecer um cursor incremental de alterações para contas a receber. A regra atual do adapter, que interrompe a sincronização com erro e rollback quando não reconhece um marcador de término, deve continuar ativa. Não substituir por “página com menos de 100 registros” nem assumir metadados ausentes.

O adapter agora envia o nome documentado do filtro de situação por rota: `situacoes[]` em contas a receber e `situacao` em contas a pagar. O teste focal do pacote Bling passou nos dois casos. Isso corrige a forma da consulta conforme o contrato público; não prova o comportamento da conta real.

O filtro de data ainda não foi corrigido. A tela aprovada envia uma opção de data e sincroniza as duas listas. Quando a opção não se aplica à outra rota, é preciso decidir se a lista deve continuar sendo consultada sem filtro ou se deve ser omitida. Essa decisão foi solicitada à responsável; nenhum filtro de data ou comportamento da tela foi alterado.

A inspeção da resposta da conta Bling da Majucau não foi concluída naquela coleta. A preferência que bloqueava o navegador foi liberada depois, conforme captura enviada pela responsável em 12/10/2026; isso permite navegar no site, mas não comprova uma chamada autenticada à API nem fornece um token OAuth. A resposta real ainda precisa ser homologada pela integração autorizada.

## Fontes oficiais

- [Boas práticas — paginação](https://developer.bling.com.br/boas-praticas)
- [Perguntas frequentes — limite padrão e parâmetro `pagina`](https://developer.bling.com.br/perguntas-frequentes)
- [Referência OpenAPI da API Bling v3](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), rota `GET /contas/receber`
- [Limites — filtros por período e limites de requisições](https://developer.bling.com.br/limites)

## Limites desta evidência

Esta é uma leitura da documentação pública, não uma chamada autenticada à conta da empresa. Ela não comprova os campos que a resposta real devolve nem a disponibilidade efetiva dos filtros na conta. Nenhuma tela ou regra financeira foi alterada. A correção técnica do filtro foi verificada pelo teste focal do adapter; a paginação e os filtros da conta real continuam sem homologação.

## Validação defensiva dos metadados — 11/10/2026

- O adapter trata os metadados de paginação recebidos como entrada não confiável. Se a resposta incluir `page`, ele deve ser positivo e corresponder à página solicitada; `limit` deve ficar entre 1 e o limite local de 100; `total` não pode ser negativo nem declarar zero quando a resposta contém registros.
- O cálculo de continuação baseado em `total` usa divisão e resto, sem multiplicar valores fornecidos pelo servidor. Isso evita overflow inteiro que poderia marcar uma página como final incorretamente.
- Testes cobrem página zero/diferente da solicitada, limite acima do teto, total negativo/contraditório e valores que causariam overflow na multiplicação. A suíte focal de Bling e `go vet` passaram localmente. A resposta real da conta continua sem homologação; os testes não tornam `total` um campo oficialmente confirmado.

## Rechecagem da referência e da permissão do navegador — 12/10/2026

- A referência OpenAPI oficial atual confirma que `GET /contas/receber` recebe `pagina` e `limite`; a resposta HTTP 200 descreve somente `data` como lista. Não há no schema divulgado `hasNext`, total de registros, página atual ou número total de páginas.
- Na mesma rota, a referência documenta `tipoFiltroData` (`E` emissão, `V` vencimento, `R` recebimento), `dataInicial` e `dataFinal`. Não lista `dataAlteracaoInicial` nem `dataAlteracaoFinal`; esses parâmetros não podem ser presumidos para essa consulta.
- As páginas oficiais de boas práticas e perguntas frequentes explicam que se percorre a paginação com `pagina` e que o padrão é até 100 registros por chamada. Elas não descrevem um marcador de última página.
- A captura enviada pela responsável mostra `https://developer.bling.com.br` e `https://www.bling.com.br` com navegação permitida no Chrome. Isso corrige o registro antigo de bloqueio do navegador, mas não é uma resposta da API. Nesta rechecagem não foi feita chamada autenticada nem lido dado financeiro da conta.
- A sincronização continua sem declarar sucesso quando não consegue provar que percorreu a lista completa. Não adotei a hipótese de que página curta ou vazia significa fim, nem mudei filtros ou comportamento da tela.

Referências oficiais consultadas em 12/10/2026:

- [OpenAPI do Bling v3 — rota `GET /contas/receber`](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json)
- [Boas práticas — paginação e limite padrão](https://developer.bling.com.br/boas-praticas)
- [Perguntas frequentes — registros por página](https://developer.bling.com.br/perguntas-frequentes)

O endpoint também depende de autorização OAuth da API. A senha de login ao site não substitui essa autorização. Para confirmar o formato real, ainda é necessária uma consulta de leitura pela integração Majucau autorizada ou uma resposta sanitizada da consulta.

## Correção da consulta por data aprovada — 12/10/2026

- Gisele definiu o escopo dos filtros manuais: vencimento consulta Contas a
  Receber e Contas a Pagar; recebimento consulta apenas Contas a Receber; e
  pagamento consulta apenas Contas a Pagar.
- O adapter agora monta os parâmetros de cada rota conforme a OpenAPI: em
  Contas a Receber, `tipoFiltroData=V` ou `R` com `dataInicial`/`dataFinal`; em
  Contas a Pagar, os pares específicos de vencimento ou pagamento. Filtros
  incompatíveis com a rota e seleção simultânea de tipos de data falham antes
  de enviar a requisição.
- O serviço não faz uma consulta de teste em Contas a Receber antes de cada
  sincronização; isso evita uma leitura financeira fora do escopo quando a
  seleção é somente pagamentos. A atualização do token continua protegida e
  pode ocorrer sem consultar uma lista financeira.
- Esta correção não resolve o término da paginação nem o filtro automático de
  alterações desde o checkpoint. A sincronização continua fechando com erro e
  rollback se não consegue provar que recebeu a lista completa. A conta real
  ainda precisa de conexão e homologação.

## Escopo da atualização automática diária — decisão de 12/10/2026

- Gisele aprovou reler diariamente todas as Contas a Receber e Contas a Pagar do Bling. A coleta automática não deve depender de um filtro genérico de “alterado desde ontem” nem se limitar aos próximos 60 dias.
- Essa escolha substitui o escopo anterior de busca incremental para a atualização automática. Não altera os filtros da sincronização manual, nem as regras financeiras ou o período de projeção dos relatórios.
- Uma releitura completa depende de percorrer todas as páginas de cada lista. A documentação pública consultada ainda não define um marcador explícito de última página; por isso, não declarar sucesso nem avançar o estado de atualização até homologar uma forma segura de detectar o término.
