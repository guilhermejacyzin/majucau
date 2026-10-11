# Verificação do backlog em atualizações documentais — 11/10/2026

## Problema observado

O verificador do backlog e seu teste de rejeição estavam dentro do job `verify-windows`. Esse job só roda quando `detect-changes` encontra alterações de produto. No run `38130699588`, do commit `9c4c727`, a mudança era documental: `detect-changes` passou, mas `verify-windows` e `verify-postgres` foram ignorados. Assim, a CI terminou com sucesso sem validar os números do backlog.

## Correção

- Criado o job independente `verify-backlog`, executado em todos os pushes e pull requests.
- O job confere contagens, estados e percentuais com `scripts/verify-backlog.ps1` e executa o teste existente de rejeição de resumo incoerente em `scripts/test-verify-backlog.ps1`.
- Os jobs de Windows e PostgreSQL continuam condicionados a mudanças que precisem de validação do produto; a verificação documental do backlog não depende deles.
- O novo status `verify-backlog` pode ser exigido junto dos status de produto quando a proteção de `main` for configurada por uma pessoa administradora.

Esta alteração corrige o pipeline; o resultado da CI deste commit será registrado após a execução remota. Nenhuma tela ou regra financeira foi alterada.
