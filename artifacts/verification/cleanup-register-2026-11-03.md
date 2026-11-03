# Registro de limpeza — 03/11/2026

- Os arquivos temporários gerados para atualizar a planilha do backlog foram
  removidos da pasta `artifacts/verification/backlog-align-20261103`.
- A remoção da pasta inteira foi bloqueada com segurança porque ela contém
  `node_modules`, marcado pelo Windows como link para outro local
  (`ReparsePoint`). O conteúdo de destino não foi inspecionado nem apagado.
- O link permanece em `artifacts/verification/backlog-align-20261103/node_modules`
  para evitar remover dependências compartilhadas fora do destino autorizado.
