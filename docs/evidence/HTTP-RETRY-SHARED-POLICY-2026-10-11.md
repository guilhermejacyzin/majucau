# Política compartilhada de retry HTTP — 2026-10-11

## Objetivo

Aplicar como padrão reutilizável nas integrações do Majucau o retry limitado já aprovado, sem repetir chamadas que possam consumir um código ou alterar estado externo.

## Implementação

- `internal/integrations/httpretry` executa no máximo três chamadas no total e aceita somente `GET` por padrão.
- O adapter informa quais erros são transitórios. A política comum controla backoff exponencial, jitter, `Retry-After` limitado a 30 segundos e cancelamento.
- `BlingAPIClient` usa o componente comum nas consultas GET. A classificação continua específica do Bling: transporte transitório, HTTP 408, 429 e 5xx podem ser repetidos; autenticação, autorização, schema e resposta inválida não.
- A troca de tokens OAuth é POST e permanece fora do retry. Nenhuma chamada mutável foi habilitada para repetição.
- O inventário local de `http.Client.Do` e criação de requests encontrou as consultas GET do Bling e o POST OAuth; não encontrou outro cliente HTTP de API externo de produção no código Go atual. Integrações futuras devem usar `httpretry` depois de confirmar a segurança de replay.

## Verificação local

- `go test -count=1 ./internal/integrations/httpretry ./internal/integrations/bling` passou no Windows.
- `go vet ./internal/integrations/httpretry ./internal/integrations/bling` passou.
- `git diff --check` passou.
- `GOCACHE`, `GOMODCACHE` e `GOTMPDIR` foram direcionados temporariamente a `artifacts/cache`; a pasta foi removida pela rotina segura após os comandos.
- A CI completa [38166703369](https://github.com/guilhermejacyzin/majucau/actions/runs/38166703369), no commit `0856a26`, passou nos quatro jobs: detecção/validação de commits, backlog, Windows e PostgreSQL.

## Limites

Embora a CI do commit tenha passado, isso não substitui a homologação do Bling, o cursor incremental oficial ou os testes de outras integrações quando forem implementadas. A política comum não permite retry de escrita sem contrato oficial e desenho de idempotência revisado.

Nenhuma tela, regra financeira, fórmula, migration ou chamada à conta real foi alterada.

## Correção da classificação de falhas de transporte — 2026-10-12

Uma revisão mais profunda encontrou uma lacuna na implementação anterior: o adapter convertia qualquer erro de transporte em “indisponível”, e o classificador entendia qualquer indisponibilidade como transitória. Assim, erros permanentes de certificado ou configuração podiam consumir as tentativas disponíveis.

- O adapter agora identifica timeouts, EOF/EOF inesperado, DNS temporário e falhas de conexão transitórias (reset, recusa, aborto, host/rede indisponível e pipe quebrado).
- Outros erros de transporte continuam produzindo a mensagem segura “Bling indisponível”, mas não são repetidos. O erro de rede original não é exposto.
- HTTP 408/429/5xx continua usando a política comum. Autenticação, autorização, schema, resposta inválida, cancelamento do chamador e POST OAuth seguem sem retry.
- Foram acrescentados casos para reset de conexão (repete dentro do limite) e certificado TLS não confiável (uma tentativa, sem espera). A suíte Go completa passou em 12/10/2026.

Esta correção só altera resiliência da integração. Nenhuma tela, regra financeira, fórmula, migration ou chamada à conta real foi alterada.

## Auditoria de cobertura dos caminhos HTTP — 2026-10-12

Uma nova busca estática cobriu todos os usos de `http.Client.Do` e de `.Do(req)` na produção Go:

- Listagens de contas a receber/pagar e movimentos financeiros entram em `doFinancialResourceRequestWithLimit`, que aplica `withBlingReadRetry` antes de cada tentativa.
- Detalhes individuais de títulos, borderôs, caixas e categorias entram em `getDataRecord`, também protegido por `withBlingReadRetry`.
- Categorias usam a mesma rota de listagem. A chamada de `rate_limit.go` é o limitador que delega cada tentativa ao transporte; não é um caminho de API separado.
- A troca de código/refresh token OAuth continua como `POST` direto, sem retry automático, pois repetir pode consumir código ou refresh token de uso único.
- Não foram encontrados outros clientes HTTP de integrações externas na produção Go atual. Novos adapters devem usar `internal/integrations/httpretry` para operações repetíveis e idempotentes após confirmar replay seguro.

**Resultado:** a revisão não encontrou chamada de leitura do Bling fora da política comum nem retry automático de escrita. Nenhuma correção de código foi necessária nesta revisão. Não foram executados testes localmente; os resultados de testes e CI acima pertencem aos commits previamente validados. A conexão real do Bling segue bloqueada pela configuração/autorização ausentes, que são uma causa separada do retry.
