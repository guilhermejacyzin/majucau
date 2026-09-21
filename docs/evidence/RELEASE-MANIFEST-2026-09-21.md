# Evidência — validador de manifesto de pacote — 2026-09-21

## Escopo

Foi implementada a primeira fundação verificável para os gates de atualização
e recuperação: `internal/installer/manifest.go` valida um manifesto JSON sem
alterar a máquina. O contrato cobre versão do manifesto, versão do aplicativo,
janela de compatibilidade do schema, checksums de migrations e inventário de
artefatos.

## Casos cobertos

- pacote válido com migration e artefato, incluindo tamanho opcional;
- checksum divergente (`CHECKSUM_MISMATCH`);
- schema atual fora da janela declarada (`SCHEMA_INCOMPATIBLE`);
- JSON/identidade inválidos, path traversal, duplicidade case-insensitive e
  caminhos protegidos de token/secret/credencial/senha (`PACKAGE_INVALID`);
- arquivos ausentes, diretórios e links que escapem da raiz são rejeitados.

## Limite da evidência

Esta implementação é somente um validador de contrato. Ela não executa
backup, criptografia, `pg_dump`, `pg_restore`, restore de banco, update,
rollback, retenção, destino externo ou integração com NSIS/worker. Portanto
OPS-02, OPS-03 e G6/G7 permanecem abertos até a implementação e execução
dessas operações em PostgreSQL e Windows 10/11 x64 limpos.

## Validação prevista

```text
go test ./internal/installer
go vet ./internal/installer
```
