# Evidência — pacote de backup PostgreSQL — 2026-09-21

## Entregue

- `internal/backup.Create` cria um pacote atômico em `%ProgramData%` (o
  chamador fornece o diretório) a partir de `pg_dump` e `pg_dumpall`.
- Credenciais são passadas por `PGPASSFILE` temporário com permissão restrita;
  senha, DSN e token não aparecem nos argumentos nem no manifesto.
- O pacote registra `app_version`, `schema_version`, `install_id`, origem,
  timestamp, parâmetros Argon2id e SHA-256 dos dois payloads.
- Payloads usam AES-GCM com associated data do caminho e o arquivo final é
  gravado em temporário e renomeado atomicamente.
- `internal/backup.Verify` valida formato, schema, política de tokens, lista
  exata de entradas, duplicidades, hashes e passphrase.
- `internal/backup.Restore` valida antes de mutar, cria backup pré-restore,
  exige hooks de ciclo de vida do worker, executa `pg_restore --clean
  --if-exists --exit-on-error` e `psql` com `PGPASSFILE`, chama validação do
  banco e só inicia o worker com status `RESTORED_NEEDS_RECONNECT`.

## Limite atual do gate

Esta entrega fecha a biblioteca de geração, verificação e restore controlado,
mas ainda falta integrar a operação ao worker/UI, exercitar um PostgreSQL real
em staging, validar constraints/RAW/snapshots/auditoria, e obter evidência em
VM Windows limpa de backup, update e rollback. Portanto o gate de
instalação/continuidade operacional permanece `PARTIAL` e o produto não deve
ser anunciado como pronto para produção.
