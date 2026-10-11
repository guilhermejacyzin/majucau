# Auditoria estática do adaptador Bling — 2026-10-11

## Escopo

Leitura do cliente HTTP, OAuth, sincronização RAW e servidor IPC. Não foi feita
chamada à conta Bling da Majucau, nem alteração de tela ou regra financeira.
Esta auditoria não executou testes locais; registra evidências de código e CIs
anteriores.

## Controles já encontrados

- `internal/integrations/bling/api_client.go`: operações GET para contas a
  receber e a pagar; limite configurado de 100 registros por página e 4 MiB por
  corpo; paginação e limite de páginas; cancelamento via contexto; timeout;
  até três tentativas para falhas transitórias de GET, com backoff/jitter e
  `Retry-After` limitado a 30 segundos.
- Erros de autenticação, validação, schema e resposta inválida não são
  repetidos. `internal/integrations/bling/oauth.go` não aplica retry ao POST de
  troca/renovação de token, que pode ser de uso único.
- `internal/integrations/bling/api_sync.go`: uma transação guarda as páginas
  RAW, o cursor e o resultado do lote; falha reverte as páginas/cursor da
  tentativa e preserva uma auditoria sanitizada; o hash da página permite
  replay idempotente. Evidências de integração PostgreSQL anteriores:
  [atomicidade](BLING-SYNC-ATOMICITY-CI-2026-10-03.md),
  [idempotência](BLING-SYNC-IDEMPOTENCY-CI-2026-10-03.md) e
  [classificação de retry](BLING-RETRY-CLASSIFICATION-2026-10-03.md).
- As credenciais são mantidas no worker/armazenamento protegido, fora do
  frontend. Dados financeiros continuam RAW; não foi presumida interpretação
  de campos da conta real.

## Lacunas observadas

- Só foram confirmados no código os recursos de contas a receber e a pagar.
  Não há homologação da conta real nem evidência de cobertura de todos os
  recursos/campos que o Majucau precisa. A documentação existente registra que
  `page:N` é um marcador técnico, não um cursor incremental oficial aprovado
  pelo Bling. Essas lacunas pertencem à descoberta de BK-040.
- A sincronização transacional tem evidência de atomicidade/idempotência; ela
  não homologa retomada incremental real, rate limits de produção ou o cursor
  oficial. Estes permanecem pendentes.

## Proteção de concorrência implementada nesta execução

- `BlingCredentialState` compartilha a mesma trava entre o salvamento da
  configuração e o serviço OAuth no worker. `Save`, `Start`, `Test`, a validação
  de `Sync`, a conclusão de `runSession` e `Disconnect` coordenam leitura e
  gravação do bundle com essa trava.
- `Sync` mantém a seção crítica somente durante a validação/renovação e a leitura
  do token; a trava é liberada antes de percorrer páginas financeiras.
- Depois do callback, `runSession` confirma que a sessão ainda está registrada
  antes de trocar e gravar tokens. Se `Disconnect` já removeu a sessão, o retorno
  atrasado não recria a autorização.
- Testes locais de `internal/integrations/bling` e `cmd/worker`, `go vet` e o
  scanner local de segredos passaram. O detector `go test -race` não pôde rodar
  neste Windows porque `CGO_ENABLED=0` e não há compilador C instalado. A CI
  `38119637844` passou no detector `-race` Ubuntu e no job PostgreSQL; o Windows
  apontou um falso positivo em um valor fictício sem marcador, corrigido no
  commit `cc97182`. A CI completa `38120125096` passou nos jobs Windows e
  PostgreSQL, incluindo os dois cenários `-race`.

## Limite e próximos passos

A trava é compartilhada entre os serviços de um único processo worker. Antes
de executar múltiplos workers/instâncias na hospedagem privada, será necessário
definir e validar uma exclusão compartilhada entre processos. Também faltam a
homologação dos endpoints/campos e do cursor oficial, sem gravar tokens, dados
pessoais ou valores reais da empresa na evidência.

## Proteção de conclusão de página — 2026-10-11

- `BlingReceivablesPage` agora registra separadamente se o fim da paginação é conhecido. A sincronização só persiste e pode avançar quando `hasNext` ou `total` vem informado.
- Sem essa informação, a rotina encerra o lote como `FAILED` com `BLING_API_PAGINATION_UNKNOWN` antes de gravar a página. O rollback do savepoint preserva os dados RAW e o cursor do último lote válido.
- `go test ./internal/integrations/bling`, `go vet ./internal/integrations/bling` e `git diff --check` passaram neste Windows, usando `artifacts/cache/`. A integração PostgreSQL não foi executada localmente porque não há DSN de teste; a CI `38121237550` passou nos jobs Windows e PostgreSQL e executou o teste de rollback/cursor.
- O contrato da API, os campos financeiros, a última página real e o cursor incremental continuam sem homologação. Nenhuma regra financeira ou tela foi alterada.

## Retry único para consultas repetíveis — 2026-10-11

- As duas consultas GET do adapter (`api_client.go` e `api_resources.go`) agora
  usam `withBlingReadRetry`, que concentra classificação de falha transitória,
  máximo de tentativas, backoff/jitter, `Retry-After` e cancelamento.
- O mecanismo recebe o método HTTP e rejeita qualquer chamada que não seja GET
  antes de executar a operação. O teste `TestBlingReadRetryRejectsNonGETBeforeCallingRequest`
  confirma que POST não é chamado nem repetido.
- O POST de troca/renovação OAuth continua fora desse mecanismo e não é repetido,
  pois pode consumir código ou refresh token de uso único.
- `go test -count=1 ./internal/integrations/bling` e
  `go vet ./internal/integrations/bling` passaram no Windows. Os três caches Go
  foram direcionados a `artifacts/cache` e removidos pela rotina segura ao fim.
- Nenhuma chamada foi feita à conta real do Bling; telas, payloads financeiros
  e regras de negócio não mudaram. BK-041 continua em 55% porque a homologação
  dos campos e da paginação oficial ainda falta.
