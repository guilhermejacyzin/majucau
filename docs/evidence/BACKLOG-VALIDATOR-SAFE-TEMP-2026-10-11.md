# Limpeza segura do teste do verificador de backlog — 11/10/2026

## Mudança

`scripts/test-verify-backlog.ps1` valida a pasta `artifacts/tmp` e seus diretórios pais com `Assert-MajucauProjectPath` antes de criar o fixture. A limpeza usa `Remove-MajucauProjectFile` com o caminho relativo esperado e remove somente o arquivo aleatório criado pelo próprio teste. A pasta compartilhada `artifacts/tmp` não é apagada, evitando interferência em outra execução.

## Verificações

- `pwsh -NoProfile -File scripts/test-verify-backlog.ps1` — aprovado.
- `pwsh -NoProfile -File scripts/test-safe-cleanup.ps1` — aprovado.
- `pwsh -NoProfile -File scripts/verify-backlog.ps1` — 127 tarefas; 7 concluídas, 11 em andamento, 109 não iniciadas; 8,5% concluído e 91,5% restante.
- `git diff --check` — aprovado.

## Limites

Esta mudança protege o fixture local do teste contra caminhos fora do projeto e reparse points. Ela não valida ACLs do Windows nem altera código do aplicativo, telas ou regras financeiras. A CI deve executar as mesmas verificações no commit publicado.
