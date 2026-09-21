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

## Limite atual do gate

Esta entrega fecha apenas a geração e a verificação criptográfica. Ainda falta
integrar a operação ao worker/UI, implementar restore controlado, backup do
estado anterior, validação em VM Windows limpa e evidência de rollback/update.
Portanto o gate de instalação/continuidade operacional permanece `PARTIAL` e o
produto não deve ser anunciado como pronto para produção.
