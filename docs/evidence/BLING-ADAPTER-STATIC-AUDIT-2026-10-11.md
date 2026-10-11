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
- O `NamedPipeServer.Serve` processa uma conexão por vez. Porém,
  `BlingOAuthService.Start` inicia `runSession` em goroutine. `Test` pode renovar
  o token e salvar o bundle enquanto essa sessão OAuth também lê/grava o mesmo
  segredo; `Disconnect` também altera o bundle. A mutex atual protege o mapa de
  sessões, não uma seção crítica comum para essas operações de credencial.
  Portanto, a serialização normal do IPC não prova a exclusão de concorrência
  necessária para fechar BK-041. É uma lacuna técnica a corrigir e verificar
  antes de considerar a renovação concorrente segura.
- A sincronização transacional tem evidência de atomicidade/idempotência; ela
  não homologa retomada incremental real, rate limits de produção ou o cursor
  oficial. Estes permanecem pendentes.

## Próximo passo seguro

Implementar uma proteção de concorrência para leitura/renovação/gravação do
bundle de credenciais, coordenando `Test`, `runSession` e `Disconnect`, e
validar a corrida em CI antes de atualizar a conclusão de BK-041. Homologar
depois os endpoints e campos com uma amostra sanitizada, sem armazenar tokens,
dados pessoais ou valores reais da empresa na evidência.
