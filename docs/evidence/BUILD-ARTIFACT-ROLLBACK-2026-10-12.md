# Preservação das saídas de build em caso de falha — 12/10/2026

## Mudança

O build local mantém uma cópia temporária dos artefatos Windows publicados antes de iniciar uma nova compilação. A cópia fica em `artifacts/cache` com identificador exclusivo da execução e contém apenas `artifacts/windows` e `artifacts/MANIFEST.txt`.

Se a compilação ou o empacotamento falhar depois do início da publicação, o script guarda as saídas parciais em `artifacts/retained/manual-cleanup/failed-build-<id>/`, acrescenta os caminhos ao registro local de limpeza manual e tenta restaurar a versão anterior. A recuperação nunca sobrescreve um caminho que não pôde ser movido com segurança. Se não houver uma versão anterior, as saídas parciais ainda ficam preservadas e registradas.

O mesmo registro é aplicado se a verificação inicial interromper a execução depois de gerar `frontend/dist` ou outras saídas conhecidas.

Quando a compilação e o empacotamento terminam com sucesso, a cópia temporária da versão anterior é removida por `SafeCleanup`. Se a recuperação ou limpeza não terminar, o caminho restante é anotado para revisão manual; caches, banco de desenvolvimento e outros dados ativos não são removidos.

## Segurança e escopo

- Caminhos são limitados à raiz do projeto e validados contra reparse points.
- A operação move apenas pastas e arquivos gerados conhecidos; não percorre nem remove caches ou dados retidos.
- Esta mudança afeta o fluxo de build e a governança de artefatos. Não altera telas, regras financeiras, integração Bling nem dados do banco.

## Verificação

`BuildFailureRetention.psm1`, `build-windows.ps1` e `SafeCleanup.psm1` passaram pela análise sintática do PowerShell. `git diff --check` não apontou problemas. Não foi executado build nem teste local; a validação de comportamento em uma falha real permanece pendente.
