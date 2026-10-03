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
- respostas 408, 429 e 5xx, além de falhas transitórias de transporte, usam até três tentativas totais com backoff exponencial e jitter; `Retry-After` é respeitado com teto de 30 segundos;
- 401/403, schema incompatível e resposta inválida não são repetidos e chegam à UI por códigos sanitizados.

## Evidência automatizada

- teste de rejeição de filtro vazio;
- teste de passagem de filtros de recebimento/pagamento e contagens sanitizadas;
- testes de retry/backoff, `Retry-After`, limite de tentativas e classificação de erros sanitizados;
- teste de aceitação do método no protocolo versionado;
- `go test ./...` e `go vet ./...`: aprovados; os testes de timeout de transporte e cancelamento do chamador também passaram na CI do commit `88cf345` ([run 37079153382](https://github.com/guilhermejacyzin/majucau/actions/runs/37079153382));
- frontend lint/typecheck/test (19) e build: aprovados;
- `wails build -clean -trimpath`: aprovado; bindings incluem `SyncBling`.

## Limites

Ainda faltam rate limiter por conta, definição oficial da janela/marcador incremental, testes de avanço transacional do cursor, normalização idempotente de negócio, reconciliação de R$ 0,01 e execução com credencial real. O retry HTTP limitado descrito acima está implementado no adapter Bling; adapters futuros devem seguir o contrato comum em `AGENTS.md`, respeitando a segurança de replay de cada operação.

## Verificação adicional da documentação de paginação — 2026-10-03

A página oficial [Limites da API](https://developer.bling.com.br/limites) informa que filtros GET por período com intervalo maior que um ano retornam HTTP 400, cita os sufixos `Inicial`/`Final` — incluindo `dataAlteracaoInicial` e `dataAlteracaoFinal` — e limita a conta a 3 requisições por segundo e 120.000 por dia. A [referência interativa](https://developer.bling.com.br/referencia) não expôs, nesta consulta, os parâmetros específicos de `GET /contas/receber` e `GET /contas/pagar`.

Essa página genérica não comprova que cada um desses endpoints aceite filtros por data de alteração, nem define ordenação estável, semântica de páginas durante alterações concorrentes ou como retomar após falha. O valor persistido `page:N` continua sendo apenas um marcador técnico da execução; não é tratado como cursor oficial do Bling. Nenhuma alteração de filtro, janela, checkpoint ou regra financeira fica autorizada por essa descoberta.

Para fechar INT-04 ainda é necessária a seção oficial de cada endpoint com seus filtros e paginação ou uma resposta sanitizada obtida com a credencial autorizada. Depois disso, homologar cada endpoint separadamente, inclusive a retomada após falha e a política `stale`, antes de alterar o sync incremental.
