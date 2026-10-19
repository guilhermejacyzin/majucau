# Validação em fluxo de TSV do Drive — 2026-10-19

## Entrega

- O importador da estrutura `02_nuvem_pago/recebimentos_futuros` agora reconhece arquivos `.tsv`, baixa cada arquivo para staging temporário privado e confere os metadados, tamanho, versão e hash com o fluxo existente do Google Drive.
- O TSV é validado linha a linha, sem acumular o arquivo nem guardar linhas. O limite de campos agora acompanha o separador tab; também se aplicam os limites existentes de 1 MiB por campo, 4 MiB por registro e 256 campos.
- Cancelamento é verificado entre registros. TSV válido, vazio, malformado ou acima dos limites permanece pendente; somente CSV com contrato aprovado pode gerar fatos financeiros. Arquivos temporários são removidos antes de encerrar a operação.
- A alteração não modifica telas, regras financeiras, contratos ou arquivos na origem.

## Verificação executada

- `scripts/verify.ps1`: aprovado em 2026-10-19.
- CI #682: aprovada nos quatro jobs de detecção de mudanças, backlog, PostgreSQL e Windows: [execução no GitHub Actions](https://github.com/guilhermejacyzin/majucau/actions/runs/39199970231).
- Testes Go e `go vet`: aprovados.
- `TestDailyDrivePipelineMarksEmptyReceiptsFolderPending` confirma que a pasta de recebimentos futuros sem arquivos registra fonte ausente e lote parcial, sem baixar nem importar arquivos.
- `govulncheck`: nenhuma chamada vulnerável; permanece um aviso de vulnerabilidade em módulo declarado, sem chamada identificada no código, já registrado em `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- Checksums de migrations, geração SQLC sem diferenças, scan de secrets e smoke/checks do instalador: aprovados.
- ESLint, TypeScript, 33 testes frontend, build frontend e `npm audit` com zero vulnerabilidades: aprovados.
- Testes novos cobrem o limite de colunas TSV, delimitadores dentro de valores entre aspas, limite de registro e staging removido após processar arquivos TSV válidos e inválidos sem importar seus valores.

## Estado que permanece

- BK-042 continua em **70%**. O avanço desta validação não comprova conexão real nem sincronização do Drive.
- A conta técnica de leitura ainda não foi criada/configurada; nenhum arquivo real foi lido ou importado. Os contratos das demais pastas e de JSON/XLS/XLSX continuam pendentes; XLS ainda não tem parser.
- Está aguardando resposta da responsável se uma pasta vazia às 09:30 significa arquivo diário ausente ou se pode ser normal. O comportamento não foi presumido nem alterado.
- Projeto: **11,7% concluído**, estimativa ponderada pelo backlog (7 de 127 tarefas estão marcadas concluídas). Nenhuma tarefa foi marcada concluída por esta alteração.

## Próximo passo

Continuar BK-042 após definir a regra de pasta vazia; provisionar a identidade técnica de leitura somente após confirmação na ação externa, então executar e conferir a primeira sincronização real. Não publicar TSV como dado financeiro até homologar seu contrato.

## Limpeza desta execução

- Com `scripts/SafeCleanup.psm1`, removidos os caches recriáveis `artifacts/cache/go-build/`, `artifacts/cache/go-tmp/`, `artifacts/cache/npm/`, `artifacts/cache/vite/` e a prévia `artifacts/verification/frontend-dist-882f77c361984a4aa8b9810347e409f7/` gerada pela verificação.
- Após o teste regressivo, também removidos com `SafeCleanup` `artifacts/cache/go-build/`, `artifacts/cache/go-tmp/`, `artifacts/cache/npm/`, `artifacts/cache/vite/` e `artifacts/verification/frontend-dist-ccab21ea3fea4bc6aadba4369318c382/`.
- Preservados `artifacts/cache/go-mod/` para evitar downloads e o pacote oficial em `artifacts/windows/`. Não foi encontrado item novo que precise de exclusão manual.

## Retificação posterior — escopo e horário dos arquivos atrasados

O registro original descreve o primeiro caminho TSV validado em recebimentos futuros. Em seguida, `DRIVE-ALL-SOURCE-STRUCTURAL-STREAMING-2026-10-19.md` ampliou a verificação estrutural em fluxo para CSV, TSV, JSON, XLS e XLSX nas três pastas aprovadas. A interpretação financeira continua pendente de contrato.

Também foi definida a regra da pasta vazia: a origem ausente fica pendente/parcial, sem reutilizar dados anteriores como atuais. O teste `TestDailyDrivePipelineMarksEmptyReceiptsFolderPending` cobre esse comportamento; a pergunta anotada mais acima já foi respondida depois.

Em 19/10/2026, a rotina do Drive também foi corrigida para não iniciar uma coleta imediata quando o serviço começa depois das 09:30. Arquivos ausentes ou recebidos após esse horário só serão processados na execução programada das 09:30 do dia seguinte, conforme a decisão registrada no BK-039.
