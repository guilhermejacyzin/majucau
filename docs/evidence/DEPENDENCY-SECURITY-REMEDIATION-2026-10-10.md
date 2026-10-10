# Atualização de dependências e verificação de segurança — 2026-10-10

## Correções aplicadas

- Go foi atualizado de 1.26.6 para 1.26.9 no `go.mod`, no CI e no README. O relatório oficial do Go identifica `net/http` como afetado antes de 1.26.9 ([GO-2026-6617](https://pkg.go.dev/vuln/GO-2026-6617)); o histórico oficial registra 1.26.9 como atualização de segurança ([Go release history](https://go.dev/doc/devel/release)).
- `golang.org/x/text` foi atualizado de 0.39.0 para 0.41.0, versão corrigida para [GO-2026-6629](https://pkg.go.dev/vuln/GO-2026-6629).
- `golang.org/x/crypto` foi atualizado de 0.53.0 para 0.56.0; junto com os requisitos do módulo, `x/net` passou para 0.57.0 e `x/sys` para 0.47.0. Isso corrige as três falhas de SSH sinalizadas pelo scanner: [GO-2026-6355](https://pkg.go.dev/vuln/GO-2026-6355), [GO-2026-6354](https://pkg.go.dev/vuln/GO-2026-6354) e [GO-2026-6303](https://pkg.go.dev/vuln/GO-2026-6303).
- `source-map-js` foi atualizado de 1.2.1 para 1.2.2 no lockfile, versão corrigida para o [aviso de alta severidade GHSA-68fv-2mgg-jv7q](https://github.com/advisories/ghsa-68fv-2mgg-jv7q). O `package.json` não precisou mudar.

## Verificação executada

O script oficial `scripts/verify.ps1` terminou com sucesso após as atualizações. Ele executou:

- testes Go do aplicativo, worker, instalador, banco e pacotes internos;
- `go vet` e `govulncheck`;
- conferência de checksums das migrations, geração determinística do `sqlc` e scan de segredos;
- verificação estática do contrato do instalador NSIS;
- lint, typecheck, 24 testes da interface, build e `npm audit`.

`go mod tidy -diff` também terminou sem diferenças. O `npm audit` encontrou zero vulnerabilidades. O `govulncheck` encontrou zero vulnerabilidades alcançáveis pelo código.

## Aviso residual e limites

O scanner ainda mostra um aviso de módulo sobre `golang.org/x/crypto/openpgp` ([GO-2026-5932](https://pkg.go.dev/vuln/GO-2026-5932)): esse pacote é considerado inseguro, não mantido e sem versão corrigida. O código do Majucau importa `golang.org/x/crypto/argon2`, não importa OpenPGP, e `govulncheck` não encontrou chamadas ao pacote afetado. O aviso deve continuar visível até a ferramenta de scan deixar de relatar código não alcançável ou a dependência de Argon2 ser substituída por decisão técnica.

As verificações foram locais e sem banco de produção. Elas não substituem CI remota, restore real, VM Windows limpa, assinatura do instalador ou homologação funcional. Nenhuma tela ou regra financeira foi alterada.
