# Atualidade dos dados Bling no painel — 19/10/2026

## O que motivou a correção

O painel lia os registros financeiros já gravados e mostrava valores e detalhes mesmo quando a atualização diária das 09:30 ainda não tinha terminado. Assim, dados antigos podiam parecer atuais. O horário da última sincronização era mostrado, mas não controlava a disponibilidade de cada recurso.

## O que mudou

- O painel identifica a execução diária esperada usando o mesmo horário de Brasília e o mesmo `correlation_id` estável que o worker grava nos lotes do Bling.
- Contas a Receber e Contas a Pagar só ficam disponíveis depois que o lote daquele recurso, daquele dia, termina com `SUCCESS`.
- Enquanto faltar a execução diária válida, o valor não é enviado ao painel e as linhas antigas não são transmitidas. O horário da última sincronização concluída continua sendo apresentado com sua data real.
- Uma execução bem-sucedida sem registros é um zero confirmado; não é tratada como fonte ausente.
- A função de gerar o identificador diário foi compartilhada com o worker sem alterar o formato persistido dos identificadores anteriores.
- O teste de integração executa a consulta contra PostgreSQL quando `MAJUCAU_TEST_DATABASE_URL` está configurada; a CI já fornece essa variável. Na máquina local, sem banco de teste configurado, esse teste é ignorado.

## Verificação

- `go test -count=1 ./internal/scheduler ./internal/dashboard ./internal/integrations/bling`: passou.
- `scripts/verify.ps1`: passou, incluindo Go, `go vet`, `govulncheck`, migrations, SQLC, varredura de segredos, verificações do instalador, ESLint, TypeScript, 33 testes de interface, build e `npm audit` sem vulnerabilidades.
- `scripts/verify-backlog.ps1` e `python docs/planejamento/work/verify_package.py`: passaram; conferiram as 127 tarefas, os percentuais, os links locais e a integridade dos arquivos listados. A lista de integridade tinha hashes antigos para dois documentos de governança e foi atualizada junto com o backlog.
- `git diff --check`: passou.
- O log completo foi guardado temporariamente em `artifacts/verification/verify-2026-10-19-dashboard-daily-freshness.log` e removido após a validação; a limpeza está registrada em `docs/evidence/WORKSPACE-CLEANUP-2026-10-18.md`.
- Teste PostgreSQL real desta nova consulta: pendente da CI após o push.

## Governança e próximos passos

BK-039 continua em 50% (50% restante); a média do backlog permanece em 11,8% concluída (88,2% restante), com 127 tarefas: 7 concluídas, 20 em andamento e 100 não iniciadas. Não aumentei o percentual porque ainda faltam comprovar a última página da API na conta real, integrar e operar a leitura do Drive de forma persistente, validar serviço e PostgreSQL numa VM limpa e homologar a sincronização real.

Próximo passo: enviar esta correção e aguardar a CI Windows/PostgreSQL. Depois, continuar na própria BK-039 até resolver a paginação real do Bling e a leitura programada do Drive. Nenhuma tela aprovada, regra de negócio ou regra financeira foi alterada.
