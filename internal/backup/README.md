# Pacote de backup local

O pacote `*.mjbk` é uma exportação criptografada do PostgreSQL local. Ele é
gerado com `pg_dump` e `pg_dumpall --globals-only --no-role-passwords`, nunca
coloca a senha nos argumentos dos processos e não exporta tokens de Bling,
Nuvem ou Nuvem Pago.

O arquivo é um ZIP controlado com:

- `backup-manifest.json`, contendo versão, schema, instalação, origem, KDF,
  hashes e tamanhos;
- `payload/database.dump.gcm`, o dump custom do banco cifrado;
- `payload/globals.sql.gcm`, os objetos globais sem senhas cifrados.

A proteção usa Argon2id para derivação de chave e AES-GCM com dados associados
ao caminho de cada payload. A verificação rejeita schema incompatível, hash
divergente, entrada duplicada, caminho inesperado e passphrase incorreta.

`Restore` é uma operação destrutiva protegida por contrato: valida o pacote
antes de mutar, cria um novo backup do estado atual, exige hooks explícitos
para parar/iniciar o worker, executa `pg_restore`/`psql` com credenciais fora
dos argumentos, valida o banco restaurado e só então retorna
`RESTORED_NEEDS_RECONNECT`. Falhas deixam o chamador em
`RECOVERY_REQUIRED`; o pacote nunca apaga ou recria um cluster por conta
própria.
