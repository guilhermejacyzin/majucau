# Evidência — build Windows e consolidação de artefatos — 2026-09-20

## Resultado

O pipeline `scripts/build-windows.ps1` foi executado com Go 1.26.8, Wails 2.13.0 e frontend Vite. A execução passou por testes, vet, govulncheck, migrations, geração sqlc determinística, scan de secrets, lint, typecheck, 19 testes frontend, build frontend e `npm audit` sem vulnerabilidades.

O build Wails `windows/amd64` foi concluído e também foram compilados o worker e o helper do instalador. Os executáveis finais foram copiados para `artifacts/windows/` e o inventário foi gerado em `artifacts/MANIFEST.txt`. O pipeline também produz `artifacts/windows/majucau-windows-x64-portable.zip`, um bundle portátil para validação local com SHA-256 dos executáveis.

## Isolamento dos caches

Go build cache, Go module cache e npm cache ficam em `artifacts/cache/`. O cache de módulos Go recebe um `go.mod` sentinela durante o pipeline para que comandos Wails que enumeram `./...` não interpretem dependências cacheadas como código do projeto. Esse arquivo é ignorado junto com o restante de `artifacts/`.

As saídas intermediárias `build/bin`, `frontend/dist` e `tmp` são removidas pela consolidação padrão, preservando somente `artifacts/` para inspeção e distribuição.

## Limite

O bundle portátil é somente um artefato de validação/diagnóstico; ele não instala PostgreSQL, serviço, ACLs ou WebView2. O artefato atual é um build Windows x64 não assinado. O instalador NSIS e a assinatura Authenticode permanecem marcos posteriores do G7.
