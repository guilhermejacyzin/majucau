# Evidência — sincronização RAW explícita Bling — 2026-09-20

## Contrato entregue

Foi adicionado o método IPC `bling.sync` e o binding Wails `SyncBling`. O worker recebe uma janela/status explícitos, recusa uma requisição vazia e executa, nesta ordem:

1. valida/testa o token autorizado;
2. importa páginas RAW de `contas/receber`;
3. importa páginas RAW de `contas/pagar`;
4. grava `sync_cursors` e finaliza o lote somente na mesma transação confirmada;
5. devolve status, lotes e contagens sanitizadas.

O botão da UI ainda não chama esse método automaticamente: a tela não inventa uma janela financeira. A habilitação visual ficará para o momento em que o período explícito e o estado operacional forem apresentados no componente.

## Segurança e consistência

- sem payload financeiro ou token atravessando a resposta IPC;
- sem normalização de valor, classificação B2B/B2C ou cálculo de dashboard;
- cursor não avança quando o lote falha ou é cancelado;
- página repetida usa hash e não cria nova versão quando não houve mudança;
- falha em uma fonte não apaga o último RAW válido.

## Evidência automatizada

- teste de rejeição de filtro vazio;
- teste de passagem de filtros de recebimento/pagamento e contagens sanitizadas;
- teste de aceitação do método no protocolo versionado;
- `go test ./...` e `go vet ./...`: aprovados;
- frontend lint/typecheck/test (18) e build: aprovados;
- `wails build -clean -trimpath`: aprovado; bindings incluem `SyncBling`.

## Limites

Ainda faltam retry/backoff com jitter, rate limiter por conta, janela incremental com watermark de negócio, normalização idempotente, reconciliação de R$ 0,01 e execução com credencial real. Esses itens permanecem no G2/G3 e não são alegados como concluídos.
