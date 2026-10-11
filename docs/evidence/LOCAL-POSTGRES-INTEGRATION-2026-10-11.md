# Execução local dos testes PostgreSQL — 2026-10-11

## Ambiente temporário

- PostgreSQL 18.3 foi inicializado em um cluster descartável dentro de
  `artifacts/cache`, vinculado somente a `127.0.0.1:55439` e sem criar serviço
  do Windows.
- O serviço PostgreSQL que já existia no computador não foi usado nem alterado.
- `scripts/verify-migrations.ps1` passou. As três migrations versionadas foram
  aplicadas ao banco `majucau_test` vazio com `ON_ERROR_STOP=1`.
- O acesso temporário usou somente fixtures sanitizadas do repositório; não
  foram usadas credenciais ou dados reais de Bling/Nuvem Pago.

## Verificações executadas

```powershell
$env:MAJUCAU_TEST_DATABASE_URL = 'postgres://postgres@127.0.0.1:55439/majucau_test?sslmode=disable'
$env:MAJUCAU_TEST_RECEIPTS_FOLDER = 'internal/integrations/bling/testdata/e2e'
$env:MAJUCAU_TEST_FUTURE_FOLDER = 'internal/integrations/nuvempago/testdata/e2e'
go test -count=1 ./internal/integrations/bling ./internal/integrations/nuvempago
```

Os dois pacotes passaram. Com as variáveis de ambiente configuradas, a execução
incluiu os testes PostgreSQL:

- `TestReceiptImportServicePostgresE2E`;
- `TestAPIResourceSyncKeepsFailureAuditAndCommitsCursorWithRAW`;
- `TestFutureImportServicePostgresE2E`.

Eles verificam persistência de importação, falha/rollback e preservação do
cursor, além de repetição idempotente da carga futura. Os testes não homologam
campos, paginação ou credenciais da API real do Bling e não substituem o ensaio
do instalador em uma máquina limpa.

## Limpeza e limite da evidência

- O cluster foi parado com sucesso, a porta 55439 ficou livre e
  `artifacts/cache` foi removida por `scripts/SafeCleanup.psm1`.
- Nenhuma tela, regra financeira, migration, serviço permanente ou dado externo
  foi alterado.
- Esta evidência comprova que os testes de integração PostgreSQL podem rodar
  localmente com uma instância isolada. Não comprova que o instalador atual
  inclua ou configure PostgreSQL dedicado.
