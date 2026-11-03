# Saídas temporárias mantidas — 03/11/2026

## Registro

| Caminho | Motivo para manter | Próxima ação segura |
|---|---|---|
| `artifacts/cache/dev` | O worker atualizado `majucau-worker-weekly.exe` (PID 14032) e o protótipo Wails atualizado `majucau-dev-wails-report.exe` (PID 18648) estão ativos. Os executáveis `majucau-dev-wails-flowfix.exe` (PID 12024) e `build/bin/majucau-dev.exe` (PID 21192) também continuam abertos. Excluir a pasta interromperia o protótipo/worker. | Preservar os executáveis ativos. Depois que cada janela/worker for encerrado normalmente, remover somente o binário sem processo pela rotina `SafeCleanup`. |
| `artifacts/retained/postgresql-dev` | A instância PostgreSQL 18 isolada do Majucau está ativa em `127.0.0.1:55438` (PID 9776). Os logs `worker-weekly.stdout.log` e `worker-weekly.stderr.log` pertencem ao worker ativo. | Preservar enquanto o protótipo estiver aberto; encerrar pelo `pg_ctl` e remover apenas após confirmar que nenhum processo a usa. |
| `frontend/dist` | Saída compilada necessária para `go:embed all:frontend/dist` na próxima compilação Wails. É cópia do build frontend validado e foi incorporada ao app ativo. | Manter até a próxima compilação; substituir somente depois de compilar e validar o novo frontend. |
| `artifacts/cache/go-mod` | Cache central de módulos Go do projeto, preservado para evitar novo download na próxima compilação. | Manter centralizado; limpar quando não houver build Go em execução, se a remoção não atrapalhar a continuidade. |
| `frontend/node_modules` | A instalação `npm ci` falhou com `EPERM` ao tentar substituir `@rollup/rollup-win32-x64-msvc/rollup.win32-x64-msvc.node`. A pasta não foi apagada ou substituída para preservar dependências em uso. | Manter e reparar em uma execução própria após identificar o processo que mantém o arquivo bloqueado; temporariamente, validar builds numa cópia isolada dentro de `artifacts/verification`. |
| `artifacts/cache/sites` | A rotina segura de limpeza recebeu “acesso negado” em um arquivo dentro de `node_modules`. | Manter no lugar consolidado de caches; investigar permissões depois que o processo que usa esse diretório for fechado. |
| `artifacts/cache/spreadsheet` e `artifacts/cache/spreadsheet-work` | A rotina segura encontrou links (`reparse points`) dentro de `node_modules` e bloqueou a exclusão para não apagar conteúdo fora da pasta do projeto. | Manter até identificar o destino dos links; remover apenas após confirmar que os destinos também são descartáveis e ficam dentro do workspace autorizado. |
| `%LOCALAPPDATA%\go-build` | Cache de compilação Go compartilhado no perfil do Windows; foi usado pelos testes desta execução e pode servir a outros projetos. | Preservado. Se quiser removê-lo manualmente, fechar os processos Go e executar a limpeza do cache global; para as próximas verificações do Majucau usar `artifacts/cache/go-build`. |

Na execução de 03/11 às 07:02 BRT, os diretórios temporários de verificação,
`artifacts/cache/npm`, `artifacts/cache/go-build` e `artifacts/cache/go-tmp`,
além do executável Wails de erro e do arquivo de build intermediário, foram
removidos pela rotina segura. A cópia de verificação e os caches temporários
foram criados sob `artifacts/`. Nenhum valor financeiro ou credencial foi
copiado para este registro.
