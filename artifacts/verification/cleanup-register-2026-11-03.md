# Registro de limpeza — 03/11/2026

- Os arquivos temporários gerados para atualizar a planilha do backlog foram
  removidos da pasta `artifacts/verification/backlog-align-20261103`.
- A remoção da pasta inteira foi bloqueada com segurança porque ela contém
  `node_modules`, marcado pelo Windows como link para outro local
  (`ReparsePoint`). O conteúdo de destino não foi inspecionado nem apagado.
- O link permanece em `artifacts/verification/backlog-align-20261103/node_modules`
  para evitar remover dependências compartilhadas fora do destino autorizado.

- O worker atualizado do protótipo está ativo como `artifacts/cache/dev/majucau-worker-status5.exe`.
  Seus redirecionamentos de saída ficam em `artifacts/cache/dev/majucau-worker-status5.stdout.log`
  e `artifacts/cache/dev/majucau-worker-status5.stderr.log`. Não removi esses três arquivos
  porque o worker ainda está em execução e a T02 depende dele; podem ser removidos manualmente
  depois que o protótipo for fechado e esse worker for encerrado.
