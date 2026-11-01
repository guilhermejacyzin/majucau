# Governança de revisão do repositório — 01/11/2026

## O que foi configurado

- `.github/CODEOWNERS` atribui as mudanças ao proprietário `@guilhermejacyzin` para solicitar sua revisão.
- `CONTRIBUTING.md` explica que essa solicitação, sozinha, não exige aprovação nem bloqueia merge.
- O procedimento também registra que aprovação mínima, proteção da branch e checks obrigatórios precisam ser configurados nas regras do GitHub por uma conta administradora.

## O que foi verificado

- O verificador local `scripts/verify-backlog.ps1` foi executado em Windows PowerShell e conferiu 126 tarefas ativas, 7 concluídas, 20 em andamento, 99 não iniciadas, 12% concluído e 88% restante.
- A conta GitHub conectada tem permissão de escrita, sem permissão administrativa. As verificações anteriores do repositório indicaram que `main` não tem proteção ativa.
- A presença de `CODEOWNERS` encaminha um pedido de revisão, mas não comprova que a revisão foi aprovada nem que o GitHub impedirá um merge.
- A CI desta atualização só pode ser confirmada após o push; este arquivo não a declara aprovada.

## Pendente

Uma pessoa administradora precisa exigir as verificações apropriadas, configurar proteção da branch e decidir o mínimo de aprovações. Não alteramos `main`, regras financeiras ou telas aprovadas nesta etapa.
