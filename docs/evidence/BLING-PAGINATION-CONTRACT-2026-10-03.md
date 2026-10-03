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
