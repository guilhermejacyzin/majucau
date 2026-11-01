# Correção do verificador do backlog no Windows PowerShell 5.1 — 01/11/2026

## Problema observado

Ao executar `scripts/verify-backlog.ps1` no Windows PowerShell 5.1.19041.6456,
o verificador lia as 127 tarefas do documento detalhado, mas contava apenas um
registro no JSON estruturado. O teste existente também falhava porque o campo
`id` não era encontrado no primeiro item.

## Causa e correção

O resultado de `ConvertFrom-Json` era colocado dentro de `@(...)`. No Windows
PowerShell 5.1, essa forma preservava o array JSON completo como um único objeto
na expressão do verificador. Atribuí diretamente o resultado do decoder à
variável da lista. A validação de cada tarefa e todos os critérios existentes
permaneceram iguais.

## Validação local

- Windows PowerShell 5.1.19041.6456: teste do verificador passou; backlog
  conferido com 126 tarefas ativas, 7 concluídas, 20 em andamento e 99 não
  iniciadas; 11,9% concluído e 88,1% restante.
- PowerShell 7.6.5: o mesmo teste e a mesma conferência passaram.
- `git diff --check` passou.
- A CI da mudança ainda depende do envio ao GitHub.

Esta correção afeta somente a ferramenta de governança do backlog. Não altera
telas, regras financeiras, dados do produto nem o percentual de tarefas
concluídas. Próximo passo: publicar e confirmar os quatro trabalhos da CI.
