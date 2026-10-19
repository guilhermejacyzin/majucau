# Limpeza do workspace — 18/10/2026

## Removido nesta execução

- O cluster temporário PostgreSQL 18 criado em `artifacts/tmp/bk042-ci-postgres/`, depois de parar o servidor e concluir a suíte completa.
- Os caches temporários de compilação e trabalho `artifacts/cache/go-build/` e `artifacts/cache/go-tmp/`.

## Mantido para o desenvolvimento

- `artifacts/cache/go-mod/`: dependências Go centralizadas para evitar novo download na próxima execução.
- `artifacts/windows/`, `artifacts/planejamento/outputs/`, `frontend/node_modules/`, `frontend/dist/`, `artifacts/tools/` e `artifacts/retained/postgresql-dev/`: entregáveis ou recursos locais que sustentam o protótipo e não são saídas descartáveis desta tarefa.

## Registrado para limpeza manual

Os caches globais abaixo são compartilhados pelo perfil Windows e podem ser usados por outros projetos. Foram preservados para evitar removê-los sem saber quem depende deles:

- `C:\Users\Gisele\go\pkg\mod` — dependências e toolchain Go compartilhados.
- `C:\Users\Gisele\AppData\Local\go-build` — cache de compilação Go compartilhado.

A lista também foi guardada em `artifacts/retained/manual-cleanup/go-global-caches-2026-10-18.txt` e no registro local de limpeza. Os arquivos oficiais do projeto e a pasta original do Drive não foram alterados.

## Limpeza nesta continuidade — 19/10/2026

- Removidos os caches regeneráveis artifacts/cache/go-build, artifacts/cache/go-tmp, artifacts/cache/npm e artifacts/cache/vite, além dos outputs frontend-dist gerados em artifacts/verification.
- Removida a cópia duplicada temporária do sqlc em artifacts/cache/bin; a ferramenta v1.31.1 permanece no local oficial artifacts/tools/go/bin para permitir verificações futuras.
- Mantido artifacts/cache/go-mod, centralizado no projeto, para evitar baixar novamente dependências e manter o desenvolvimento offline. Também foram preservados frontend/node_modules, ferramentas e artefatos oficiais.
- Nenhum cache global do perfil Windows foi excluído; os caminhos compartilhados permanecem registrados para limpeza manual em artifacts/retained/manual-cleanup/go-global-caches-2026-10-18.txt.

## Limpeza após verificação do provisionamento Google Drive — 2026-10-19

- Removidos com `scripts/SafeCleanup.psm1`: `artifacts/cache/go-build/`, `artifacts/cache/go-tmp/`, `artifacts/cache/npm/`, `artifacts/cache/vite/` e `artifacts/verification/frontend-dist-dbbcf2bcebfa4da2803f2cdf483487ae/`.
- Mantido `artifacts/cache/go-mod/`: dependências Go necessárias para próximas verificações sem downloads; está no local central aprovado.
- Nenhum outro cache, dado do projeto, saída oficial, cache global ou arquivo do Drive foi removido. Não há novo item que dependa de limpeza manual.

## Limpeza após montagem do pacote Windows — 2026-10-19

- Removidos com `scripts/SafeCleanup.psm1` os caches regeneráveis `artifacts/cache/go-build/`, `artifacts/cache/go-tmp/`, `artifacts/cache/npm/` e `artifacts/cache/vite/`, além da pasta temporária `artifacts/verification/frontend-dist-a0d8c96b3539469fa73262d6d51596ad/` criada durante a compilação.
- Mantidos `artifacts/cache/go-mod/` e o pacote local `artifacts/windows/`, que contém o protótipo portátil para conferência. A pasta `artifacts/verification/` ficou vazia.
- Nenhum cache global do Windows, dado real do Drive ou arquivo que precise de exclusão manual foi encontrado nesta execução.

## Limpeza após correção Windows e recompilação do pacote — 2026-10-19

- Removidos com `scripts/SafeCleanup.psm1` os caches regeneráveis `artifacts/cache/go-build/`, `artifacts/cache/go-tmp/`, `artifacts/cache/npm/` e `artifacts/cache/vite/`, além do preview temporário `artifacts/verification/frontend-dist-ec10e5410c0c4490a770086c0c066d42/`.
- O script de consolidação já removeu `build/bin/`, `frontend/dist/` e `tmp/` após copiar os arquivos necessários. O pacote atualizado e seus executáveis permanecem centralizados em `artifacts/windows/`.
- Mantido `artifacts/cache/go-mod/` para preservar as dependências Go. Nenhum cache global, dado real do Drive ou item que precise de exclusão manual foi encontrado.
