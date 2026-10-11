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
