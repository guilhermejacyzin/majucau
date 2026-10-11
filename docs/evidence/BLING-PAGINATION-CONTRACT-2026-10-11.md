# Evidência — contrato público de paginação do Bling

Data da consulta: 11/10/2026.

## O que a documentação oficial confirma

- A referência OpenAPI pública descreve `GET /contas/receber` como uma lista paginada e inclui os parâmetros `pagina` e `limite`.
- O FAQ do Bling informa que o padrão é até 100 registros por requisição e orienta informar `pagina` para percorrer os resultados.
- Para `GET /contas/receber`, o schema público da resposta descreve um objeto cuja propriedade `data` é uma lista de contas. O schema não declara `hasNext`, `total`, página atual, nem outra informação para identificar a última página.
- O endpoint declara `tipoFiltroData`, `dataInicial` e `dataFinal`. A referência não declara `dataAlteracaoInicial` nem `dataAlteracaoFinal` para essa rota. A página de limites menciona esses nomes como filtros que existem em APIs do Bling, mas isso não confirma que estejam disponíveis em toda rota.

## O que isso significa para o Majucau

A paginação por número de página está documentada, mas a documentação pública consultada não permite decidir, sem suposição, quando a última página foi alcançada nem estabelecer um cursor incremental de alterações para contas a receber. A regra atual do adapter, que interrompe a sincronização com erro e rollback quando não reconhece um marcador de término, deve continuar ativa. Não substituir por “página com menos de 100 registros” nem assumir metadados ausentes.

A inspeção da resposta da conta Bling da Majucau não foi concluída: a preferência de segurança salva no Chrome ainda bloqueia a aba do Bling. A autorização anterior da usuária não deve ser contornada. Para homologar o contrato real, será necessário liberar o site nessa preferência e avisar, ou compartilhar uma resposta de exemplo sanitizada que preserve os nomes dos campos e remova valores, IDs reais, dados pessoais e credenciais.

## Fontes oficiais

- [Boas práticas — paginação](https://developer.bling.com.br/boas-praticas)
- [Perguntas frequentes — limite padrão e parâmetro `pagina`](https://developer.bling.com.br/perguntas-frequentes)
- [Referência OpenAPI da API Bling v3](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), rota `GET /contas/receber`
- [Limites — filtros por período e limites de requisições](https://developer.bling.com.br/limites)

## Limites desta evidência

Esta é uma leitura da documentação pública, não uma chamada autenticada à conta da empresa. Ela não comprova os campos que a resposta real devolve nem a disponibilidade efetiva dos filtros na conta. Nenhuma tela, regra financeira, cursor ou comportamento do sistema foi alterado; não foram executados testes.
