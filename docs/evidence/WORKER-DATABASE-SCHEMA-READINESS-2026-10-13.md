# Prontidão do esquema PostgreSQL no worker — 13/10/2026

## Mudança

Antes, a verificação de saúde considerava o PostgreSQL pronto após um `Ping`. Isso podia deixar o worker marcado como saudável quando o banco estava acessível, mas ainda não tinha todas as tabelas essenciais do Majucau.

O worker agora verifica, sob o mesmo limite total de 1,5 segundo:

1. se o PostgreSQL responde;
2. se existem `integration_connections`, `raw_records`, `audit_events` e `daily_balances`.

Se faltar uma dessas tabelas, a dependência PostgreSQL aparece como degradada com o código estável `DATABASE_SCHEMA_NOT_READY`. Se a consulta ao banco falhar, o detalhe continua sendo `DATABASE_UNAVAILABLE`; erros brutos do driver não são exibidos.

## Verificação

- `go test -count=1 ./cmd/worker`: passou.
- `go vet ./cmd/worker`: passou.
- `git diff --check`: passou.
- A CI do commit ainda precisa confirmar a compilação Windows e os demais gates.

## Limites

Esta verificação comprova somente a presença das quatro tabelas consultadas pelo SQL `SchemaHealth`. Ela não verifica todas as permissões do papel `majucau_runtime`, checksum/versão de migrations ou a operação em uma instalação limpa. Esses pontos continuam pendentes em BK-027 e BK-024.

Nenhuma tela aprovada, regra financeira ou regra de negócio foi alterada. Não houve conexão real com o Bling nem leitura de dados financeiros.
