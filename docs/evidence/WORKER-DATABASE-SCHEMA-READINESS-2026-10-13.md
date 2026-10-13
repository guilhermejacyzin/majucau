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
- A CI completa [38371847314](https://github.com/guilhermejacyzin/majucau/actions/runs/38371847314) passou nos jobs de backlog, Windows e PostgreSQL. O job Windows compilou o worker e o instalador de homologação.
- A verificação PostgreSQL anterior não executava o novo SQL de saúde contra o serviço descartável. Foi acrescentado `TestPostgresWorkerHealthProbeReportsMigratedSchema`, acionado depois das migrations na CI. A execução local passou, mas pulou a parte PostgreSQL porque este terminal não tem uma base de teste configurada. A CI completa [38373291264](https://github.com/guilhermejacyzin/majucau/actions/runs/38373291264) passou nos jobs de PostgreSQL e Windows, incluindo a nova chamada contra o banco descartável.

## Limites

Esta verificação comprova somente a presença das quatro tabelas consultadas pelo SQL `SchemaHealth`. O novo teste direciona o método para o PostgreSQL descartável após as migrations e passou na CI [38373291264](https://github.com/guilhermejacyzin/majucau/actions/runs/38373291264). A verificação não comprova a conexão do protótipo, todas as permissões do papel `majucau_runtime` ou operação em uma instalação limpa. Esses pontos continuam pendentes em BK-027 e BK-024.

Nenhuma tela aprovada, regra financeira ou regra de negócio foi alterada. Não houve conexão real com o Bling nem leitura de dados financeiros.
