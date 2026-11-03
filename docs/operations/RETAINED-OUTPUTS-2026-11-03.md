# Saídas temporárias mantidas — 03/11/2026

## Registro

| Caminho | Motivo para manter | Próxima ação segura |
|---|---|---|
| `artifacts/cache/dev` | O processo `majucau-worker-folder02.exe` (PID 7760) está executando a partir desta pasta durante o uso do protótipo. Apagá-la interromperia o worker e a consulta ativa. | Preservar enquanto o protótipo/worker estiver aberto. Depois que o worker for encerrado normalmente, remover a pasta pela rotina `SafeCleanup`. |
| `artifacts/cache/sites` | A rotina segura de limpeza recebeu “acesso negado” em um arquivo dentro de `node_modules`. | Manter no lugar consolidado de caches; investigar permissões depois que o processo que usa esse diretório for fechado. |
| `artifacts/cache/spreadsheet` e `artifacts/cache/spreadsheet-work` | A rotina segura encontrou links (`reparse points`) dentro de `node_modules` e bloqueou a exclusão para não apagar conteúdo fora da pasta do projeto. | Manter até identificar o destino dos links; remover apenas após confirmar que os destinos também são descartáveis e ficam dentro do workspace autorizado. |

As outras caches que puderam ser removidas com segurança foram excluídas. A
rotina também encontrou um caminho já ausente dentro do cache npm e continuou a
limpeza. Nenhum dado financeiro ou credencial foi copiado para este registro.
