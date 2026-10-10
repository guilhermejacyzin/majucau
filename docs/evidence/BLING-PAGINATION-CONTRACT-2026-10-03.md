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

## Rechecagem da documentação pública e do código — 2026-10-10

### O que a documentação pública confirma

- A [FAQ oficial do Bling](https://developer.bling.com.br/perguntas-frequentes) e as [boas práticas oficiais](https://developer.bling.com.br/boas-praticas) confirmam paginação por `pagina`, com `limite` configurável e padrão de 100 registros.
- A página oficial de [limites e filtros](https://developer.bling.com.br/limites) explica sufixos `Inicial`/`Final` para filtros de período e cita `dataAlteracaoInicial`/`dataAlteracaoFinal`. Isso não confirma, por si só, que `GET /contas/receber` aceita esses filtros.
- As páginas públicas consultadas não descrevem o campo de resposta que sinaliza a última página nem confirmam `pagination.hasNext` ou `pagination.total` para esse endpoint.

### Risco identificado no adaptador atual

- `api_client.go` só marca `HasNext` quando recebe `pagination.hasNext` ou `pagination.total`. Se a resposta não trouxer esses metadados, `HasNext` permanece falso.
- `api_sync.go` interpreta esse valor como fim normal e marca a sincronização como concluída. Se o endpoint omitir a paginação de resposta, a execução pode terminar após a primeira página sem acusar incompletude.
- O cursor gravado é `page:<n>`, mas a próxima sincronização não o lê para escolher a página inicial. Portanto, ele não é um marcador incremental comprovado nem um mecanismo de retomada entre execuções.
- Os testes atuais usam respostas sintéticas; não provam o envelope real da conta Bling.

### Limite da investigação e decisão

- A referência interativa do Bling foi bloqueada por uma preferência de segurança salva no navegador. Nenhum outro navegador, ferramenta ou rota foi usado para contornar o bloqueio.
- Não alterei o código: adivinhar o campo de resposta ou encerrar pela quantidade de registros poderia pular dados em uma resposta incompleta.
- `DATA-02` permanece `PARTIAL` e a paginação continua bloqueada até confirmar o contrato específico do endpoint por documentação acessível ou resposta sanitizada. Para a resposta, basta preservar nomes dos campos de paginação e quantidades; remover registros, IDs reais, valores financeiros, tokens e dados pessoais.
