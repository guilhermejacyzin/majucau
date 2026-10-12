# Evidência — contrato de paginação do Bling — 2026-10-03

## Fonte oficial consultada

- A [documentação geral de boas práticas do Bling](https://developer.bling.com.br/boas-praticas) diz que leituras GET paginadas usam `pagina` e `limite`, retornam 100 registros por padrão e permitem configurar a quantidade.
- A página oficial de [limites e filtros do Bling](https://developer.bling.com.br/limites) informa que filtros por período com intervalo maior que um ano recebem HTTP 400 e dá como exemplos parâmetros com sufixos `Inicial`/`Final`, incluindo `dataAlteracaoInicial` e `dataAlteracaoFinal`.
- Essas páginas não provam que cada filtro de período está disponível em `GET /contas/receber` ou `GET /contas/pagar`, nem especificam o objeto de paginação devolvido por esses endpoints.

## Comparação com o adapter

- `internal/integrations/bling/api_client.go` envia `pagina` e `limite`. O adapter escolhe 100 registros como limite local por bloco; esse valor não deve ser descrito como máximo documentado pelo provedor.
- O decoder atualmente espera a resposta em `data` e procura `pagination.page`, `pagination.limit`, `pagination.total` e `pagination.hasNext`. Os testes HTTP existentes usam fixtures sintéticas com esse formato; não há amostra sanitizada de uma resposta autorizada da conta Majucau.
- `ReceivablesAPISyncService.Sync` começa pela página pedida, que por padrão é 1, percorre as páginas da execução e grava `page:<n>` no cursor somente no commit final. A próxima chamada não lê esse valor para escolher a página inicial. Portanto, esse valor identifica a última página da execução persistida; não é um cursor incremental comprovado pelo Bling.
- A documentação consultada não determina ordenação estável durante alterações concorrentes nem uma marca d'água oficial que permita avançar incrementalmente sem risco de lacunas.

## Estado e evidência faltante

- Nenhuma tela ou regra de negócio foi alterada. Nenhum filtro de `dataAlteracao` ou envelope de resposta foi inventado.
- `DATA-02` permanece `PARTIAL`. Antes de mudar a sincronização, é necessária a referência específica do endpoint ou uma resposta sanitizada preservando somente os nomes e valores dos campos de paginação. Registros, IDs reais, valores financeiros, tokens e dados pessoais devem ser removidos.
- Ainda é necessário confirmar se os filtros de alteração são aceitos nesses endpoints e se a paginação permanece estável enquanto os registros mudam. A homologação com a conta autorizada deve gerar uma fixture sanitizada e testes de contrato.

## Rechecagem da OpenAPI pública e do código — 2026-10-10

### O que a OpenAPI do endpoint confirma

- A [referência oficial OpenAPI do Bling](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json) descreve `GET /contas/receber` como paginado e lista os parâmetros `pagina` e `limite`.
- A definição oficial desses parâmetros informa `pagina` padrão 1 e mínimo 1; `limite` padrão 100 e mínimo 1. A OpenAPI não publica um limite máximo.
- Para esse endpoint, a OpenAPI documenta filtros `dataInicial` e `dataFinal`, interpretados conforme `tipoFiltroData` (`E` emissão, `V` vencimento, `R` recebimento). Ela não lista `dataAlteracaoInicial` nem `dataAlteracaoFinal` entre os parâmetros de `GET /contas/receber`.
- A resposta HTTP 200 documenta um objeto cuja propriedade `data` é uma lista de contas. Não documenta metadados como `pagination.hasNext`, `pagination.total` ou a quantidade total de páginas. O schema não é prova de que a API nunca envie campos adicionais fora da documentação.
- Portanto, a fonte oficial confirma os parâmetros da paginação e o formato documentado da lista, mas não define como detectar a última página nem uma marca d'água de alteração para sincronização incremental.

### Risco identificado no adaptador atual

- `api_client.go` só marca `HasNext` quando recebe `pagination.hasNext` ou `pagination.total`. Se a resposta não trouxer esses metadados, `HasNext` permanece falso.
- `api_sync.go` interpreta esse valor como fim normal e marca a sincronização como concluída. Como a OpenAPI não documenta esses metadados, existe risco de uma resposta válida conforme o schema ser tratada como completa após uma única página. Isso precisa ser resolvido antes de aceitar a sincronização para produção.
- O cursor gravado é `page:<n>`, mas a próxima sincronização não o lê para escolher a página inicial. Portanto, ele não é um marcador incremental comprovado nem um mecanismo de retomada entre execuções.
- Os testes atuais usam respostas sintéticas; não provam o envelope real da conta Bling.

### Limite da investigação e decisão

- A tentativa de abrir a aba autenticada do Bling foi bloqueada por uma preferência de segurança salva no navegador, mesmo com a autorização já dada pela usuária. Nenhum outro navegador, ferramenta ou rota foi usado para contornar o bloqueio.
- Não alterei o código: adivinhar o campo de resposta ou encerrar pela quantidade de registros poderia pular dados em uma resposta incompleta.
- `DATA-02` permanece `PARTIAL`. Não alterei o código nem presumi que uma página curta ou vazia seja o marcador final. Antes de corrigir o comportamento, precisamos confirmar como a resposta real encerra a paginação ou aprovar um procedimento seguro de leitura página a página.
- Para ajudar nessa confirmação, uma resposta sanitizada precisa manter apenas os nomes dos campos de paginação e quantidades, removendo registros, IDs reais, valores financeiros, tokens e dados pessoais. Outra opção é ajustar a preferência salva do navegador para permitir o acesso ao Bling; a tentativa bloqueada não será contornada.

## Proteção contra metadados de paginação ausentes — 2026-10-11

- O cliente agora diferencia `hasNext=false` confirmado de ausência de evidência sobre a próxima página. `HasNextKnown` só fica verdadeiro quando a resposta fornece `hasNext` ou `total`.
- Antes de persistir uma página, o sincronizador exige essa confirmação. Se ela faltar, o lote registra `BLING_API_PAGINATION_UNKNOWN`, desfaz a gravação e mantém o cursor anterior. Uma resposta sem metadados deixa de poder parecer uma sincronização completa.
- Testes locais cobrem resposta sem `pagination` e resposta que traz somente página/limite. A CI `38121237550` passou nos jobs Windows e PostgreSQL, incluindo o teste que verifica que páginas RAW válidas e cursor não mudam quando a resposta não confirma a última página.
- Esta proteção não define o contrato real de paginação e não cria retomada incremental. A homologação com a conta e a documentação permanece pendente; não inferir fim de lista por página curta/vazia.

## Conferência dos filtros dos dois endpoints — 2026-10-11

Foi conferida novamente a [OpenAPI oficial do Bling](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), incluindo os dois endpoints usados pelo sincronizador:

- `GET /contas/receber` documenta `pagina` e `limite`; os filtros de data são `tipoFiltroData` (`E` emissão, `V` vencimento, `R` recebimento) junto com `dataInicial`/`dataFinal`. A situação usa `situacoes[]`. A resposta 200 descreve `data` como lista, sem metadado documentado de última página.
- `GET /contas/pagar` documenta `pagina` e `limite`, filtros distintos de emissão, vencimento e pagamento, e situação singular `situacao`. A resposta 200 também descreve `data` como lista, sem metadado documentado de última página.
- A documentação não apresenta filtros de `dataAlteracao` para esses dois endpoints. A existência genérica de filtros com esse nome em outros recursos não homologa seu uso aqui.

O cliente atual constrói os filtros dos dois endpoints por uma função compartilhada (`internal/integrations/bling/api_client.go`). Os nomes enviados para Contas a Receber não correspondem aos parâmetros de data e situação publicados na OpenAPI; Contas a Pagar tem parâmetros próprios. A tela atual seleciona Vencimento, Recebimento ou Pagamento e informa que a execução lê os dois grupos. Não está definido se, ao escolher Recebimento ou Pagamento, a execução deve consultar somente o grupo correspondente ou também consultar o outro grupo sem filtro de período. Essa escolha pode mudar o conjunto de lançamentos importados.

Nenhum filtro foi alterado. A sincronização mantém a proteção contra resposta sem confirmação de última página. Antes de corrigir o mapeamento dos filtros, é necessária confirmação do comportamento esperado para os dois grupos. Telas aprovadas e regras financeiras permanecem intactas.

## Janela desejada para atualização diária — 2026-10-12

- Gisele definiu que a atualização automática procure alterações desde a execução anterior e mantenha os dados dos próximos 60 dias.
- A regra de integridade do projeto exige usar o último checkpoint confirmado em commit, e não a última tentativa; uma execução falha não pode avançar o período nem pular dados.
- A OpenAPI oficial já conferida para `GET /contas/receber` e `GET /contas/pagar` não documenta filtros `dataAlteracaoInicial`/`dataAlteracaoFinal` nem metadados de última página. A intenção de coleta, portanto, ainda não tem um contrato técnico comprovado para esses endpoints.
- O sincronizador atual falha com `BLING_API_PAGINATION_UNKNOWN` quando a resposta não confirma se há outra página; essa proteção deve ser mantida. Não declarar coleta completa, não avançar o checkpoint e não apresentar dados antigos como atuais até obter contrato endpoint-específico ou evidência sanitizada válida da resposta autorizada.
- Esta decisão não altera o corte D-1 da análise normal, as regras financeiras nem as telas. A consulta automática Bling permanece pendente de validação de cobertura e paginação.
