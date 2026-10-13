# Compatibilidade do verificador do backlog com PowerShell do Windows — 13/10/2026

## Problema

O verificador do backlog e seu teste estavam salvos em UTF-8 sem BOM. O Windows PowerShell 5.1 interpretava os textos acentuados com a codificação local e falhava ao abrir o script. A primeira execução também mostrou que `$PSScriptRoot` não estava disponível ao calcular um parâmetro padrão no bloco `param`.

## Correção

- Salvei `scripts/verify-backlog.ps1` e `scripts/test-verify-backlog.ps1` com BOM UTF-8, preservando o conteúdo em português.
- Mudei o caminho padrão do arquivo do backlog para ser resolvido depois do início do script, com base no caminho de execução. Chamadas que passam `-Path` continuam aceitas.

## Verificação

- Windows PowerShell 5.1.19041.6456: o verificador contou 127 tarefas, 7 concluídas, 18 em andamento e 102 não iniciadas; 11% concluído e 89% restante.
- PowerShell 7.6.5: o mesmo resultado.
- `scripts/test-verify-backlog.ps1` passou nas duas versões. O teste confirmou que o verificador aceita um resumo correto e rejeita percentual divergente.
- Os arquivos temporários do teste ficaram em `artifacts/tmp` e foram removidos pelo bloco de limpeza do próprio teste. A Junction preexistente `artifacts/tmp/node_modules` foi preservada; já consta no registro local de limpeza e aponta para dependências compartilhadas do Codex.

Nenhuma tela aprovada, regra financeira ou dado do projeto foi alterado. A mudança melhora apenas a execução local da governança no Windows.
