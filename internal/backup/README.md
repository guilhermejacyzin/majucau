# Pacote de backup local

O pacote `*.mjbk` exporta o PostgreSQL local com `pg_dump` e `pg_dumpall
--globals-only --no-role-passwords`. Senhas do banco são passadas em um
`PGPASSFILE` temporário; tokens de integração não são exportados.

## Formatos

- **V1 (leitura/restauração):** formato AES-GCM legado. A aplicação preserva a
  abertura dos ZIPs antigos e descriptografa o payload em blocos com o CNG do
  Windows. O limite máximo inerente ao AES-GCM legado continua valendo para
  arquivos V1.
- **V2 (gravação, verificação e restauração):** Tink Streaming AEAD com
  AES-256-GCM-HKDF e segmentos autenticados de 1 MiB. O limite por tamanho do
  aplicativo foi removido; espaço em disco e limites técnicos do formato ainda
  se aplicam.

O manifesto V2 registra versão, schema, instalação, origem, parâmetros
Argon2id, algoritmo e SHA-256 dos payloads cifrados. A chave Tink aleatória do
pacote fica em um keyset cifrado com AES-GCM, derivado da frase-senha, e dados
associados fixos ao formato. O arquivo e o manifesto são lidos em fluxo; só o
manifesto (até 1 MiB) e o keyset cifrado (até 64 KiB) são carregados como
blocos pequenos em memória. Entradas inesperadas, duplicadas, compactação não
permitida, hashes divergentes e tags inválidas são rejeitados.

## Proteção temporária e restauração

No Windows, dumps temporários, senha PostgreSQL e área de restauração usam uma
pasta criada já com DACL protegida para o proprietário, LocalSystem e
administradores. O restore primeiro copia e valida a mesma cópia criptografada,
depois grava os dumps descriptografados nessa área privada e só os entrega a
`pg_restore`/`psql` após autenticar todos os payloads. As saídas temporárias são
removidas ao encerrar a operação.

`Restore` valida o pacote antes de alterar o banco, cria um backup pré-restore,
exige hooks explícitos para parar/iniciar o worker, valida o banco restaurado
e retorna `RESTORED_NEEDS_RECONNECT`. Falhas após o início da restauração
retornam `RECOVERY_REQUIRED`; o pacote nunca apaga ou recria um cluster por
conta própria.

## Estado de validação

A mudança V1/V2 está em implementação e aguarda a execução da CI Windows e
PostgreSQL. Ainda são necessárias evidências de restauração V1 em Windows 10/11
limpo, restore real de PostgreSQL e integração com worker/UI/instalador antes
de fechar os gates operacionais.
