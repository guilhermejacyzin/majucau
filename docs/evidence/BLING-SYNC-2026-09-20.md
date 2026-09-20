# Evidência — sincronização RAW explícita Bling — 2026-09-20

## Contrato entregue

Foi adicionado o método IPC `bling.sync` e o binding Wails `SyncBling`. O worker recebe uma janela/status explícitos, recusa uma requisição vazia e executa, nesta ordem:

1. valida/testa o token autorizado;
2. importa páginas RAW de `contas/receber`;
3. importa páginas RAW de `contas/pagar`;
4. grava `sync_cursors` e finaliza o lote somente na mesma transação confirmada;
5. devolve status, lotes e contagens sanitizadas.

A tela de integrações agora abre um painel explícito com tipo de data (vencimento, recebimento ou pagamento), data inicial/final e situação opcional. O botão de execução permanece bloqueado enquanto o Bling não estiver conectado e a validação local recusa filtro vazio ou período invertido; não existe janela financeira inventada nem sincronização silenciosa.

## Segurança e consistência

- sem payload financeiro ou token atravessando a resposta IPC;
- sem normalização de valor, classificação B2B/B2C ou cálculo de dashboard;
- cursor não avança quando o lote falha ou é cancelado;
- página repetida usa hash e não cria nova versão quando não houve mudança;
- falha em uma fonte não apaga o último RAW válido.
- respostas 408, 429 e 5xx, além de falhas transitórias de transporte, usam até três tentativas com backoff exponencial e jitter; `Retry-After` é respeitado com teto de 30 segundos;
- 401/403, schema incompatível e resposta inválida não são repetidos e chegam à UI por códigos sanitizados.

## Evidência automatizada

- teste de rejeição de filtro vazio;
- teste de passagem de filtros de recebimento/pagamento e contagens sanitizadas;
- testes de retry/backoff, `Retry-After`, limite de tentativas e classificação de erros sanitizados;
- teste de aceitação do método no protocolo versionado;
- `go test ./...` e `go vet ./...`: aprovados;
- frontend lint/typecheck/test (19) e build: aprovados;
- `wails build -clean -trimpath`: aprovado; bindings incluem `SyncBling`.

## Limites

Ainda faltam retry/backoff com jitter, rate limiter por conta, janela incremental com watermark de negócio, normalização idempotente, reconciliação de R$ 0,01 e execução com credencial real. Esses itens permanecem no G2/G3 e não são alegados como concluídos.
