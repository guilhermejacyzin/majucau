# Evidência — build Windows e consolidação de artefatos — 2026-09-20

## Resultado

O pipeline `scripts/build-windows.ps1` foi executado com Go 1.26.8, Wails 2.13.0 e frontend Vite. A execução passou por testes, vet, govulncheck, migrations, geração sqlc determinística, scan de secrets, lint, typecheck, 19 testes frontend, build frontend e `npm audit` sem vulnerabilidades.

O build Wails `windows/amd64` foi concluído e também foram compilados o worker e o helper do instalador. Os executáveis finais foram copiados para `artifacts/windows/` e o inventário foi gerado em `artifacts/MANIFEST.txt`. O pipeline também produz `artifacts/windows/majucau-windows-x64-portable.zip`, um bundle portátil para validação local com SHA-256 dos executáveis.

## Isolamento dos caches

Go build cache, Go module cache e npm cache ficam em `artifacts/cache/`. O cache de módulos Go recebe um `go.mod` sentinela durante o pipeline para que comandos Wails que enumeram `./...` não interpretem dependências cacheadas como código do projeto. Esse arquivo é ignorado junto com o restante de `artifacts/`.

As saídas intermediárias `build/bin`, `frontend/dist` e `tmp` são removidas pela consolidação padrão, preservando somente `artifacts/` para inspeção e distribuição.

## Limite

O bundle portátil é somente um artefato de validação/diagnóstico; ele não instala PostgreSQL, serviço, ACLs ou WebView2. O artefato atual é um build Windows x64 não assinado. O instalador NSIS e a assinatura Authenticode permanecem marcos posteriores do G7.

## Complemento — geração do inventário em fluxo — 2026-10-11

`scripts/ArtifactManifest.psm1` gera `artifacts/MANIFEST.txt` escrevendo uma linha por arquivo, sem manter todos os caminhos em memória. O inventário inclui apenas `artifacts/windows` e `artifacts/verification`; caches em `artifacts/cache` e conteúdo preservado em `artifacts/retained` são excluídos. Links/reparse points são recusados. As CIs `38110918694` (`41f47dc`), `38111428867` (`74633ff`) e `38111468872` (`fa876eb`) falharam no Windows ao encontrar funções `SafeCleanup` nos scripts chamadores. Na CI `38111929797` (`02ea6a0`), PostgreSQL, testes Go e verificações do frontend passaram; o Windows não localizou `SafeCleanup` pelo nome curto porque o módulo estava em outro escopo. O commit `7d13f87` carregou o módulo de limpeza também no escopo dos scripts chamadores. A CI [38112462467](https://github.com/guilhermejacyzin/majucau/actions/runs/38112462467) confirmou com sucesso a compilação e a consolidação no Windows e a verificação de backup/restore no PostgreSQL. O instalador sem assinatura ficou centralizado no artefato `majucau-g1-unsigned` da própria execução, com expiração em 2026-10-12 04:50 UTC; mantido temporariamente para conferência e diagnóstico.

## Registro de caches locais mantidos — 2026-10-11

A pasta central `artifacts/cache` está ignorada pelo Git. Na conferência havia
201.027 arquivos, cerca de 1,81 GiB, distribuídos em `bin`, `dev`, `go-mod` e
`nsis-3.13`; `downloads` e `go-install-stage` estavam vazias. Os itens são
ferramentas, fontes/dependências e instalador local usados para reconstruir o
projeto. Foram mantidos nessa pasta central porque removê-los exige novo
download e pode impedir uma compilação sem internet. Ficam registrados para
limpeza manual quando a responsável decidir que não precisa mais deles. Esta
execução não criou outputs de build locais.
