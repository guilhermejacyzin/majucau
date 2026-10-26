# Retomada do protótipo — 25/10/2026

## Situação

- Às 20:03 (horário de Brasília), a pedido da usuária, iniciei `scripts/start-prototype.ps1` novamente.
- O PostgreSQL isolado está ouvindo em `127.0.0.1:55438`; a interface Wails/Vite está ativa nas portas locais `34115` e `5173`.
- A execução anterior não persistiu como processo entre as sessões. O protótipo foi iniciado novamente nesta retomada.
- A pasta DPAPI ativa contém somente `.entropy` e `.scope`; não existe `bling_credentials_v1.dpapi`, que é o arquivo definido pelo código para a credencial do Bling. Isso confirma que o segredo não foi salvo nesta instância. Nenhum segredo foi lido.
- Não houve OAuth, chamada real, sincronização de dados, alteração de código, tela ou regra financeira nesta retomada.
- Fotografia inicial desta evidência: a branch estava sincronizada no commit `bf9481a`; a atualização corrente está registrada abaixo.

## Próximo passo

1. Na tela **Importações/Integrações → Bling** do protótipo aberto, informar e salvar o Client Secret do aplicativo. Essa configuração pertence a BK-041; não enviar o segredo no chat. Depois conectar e concluir a autorização.
2. Confirmar o estado conectado, testar a API, validar a paginação e realizar a sincronização inicial; corrigir qualquer falha antes de concluir os critérios de BK-041.

## Progresso e proteção

- BK-039 (execução de trabalhos em segundo plano): 50% concluída; 50% restante.
- Projeto: estimativa do backlog de 11,8% concluído; 88,2% restante (127 tarefas: 7 concluídas, 20 em andamento e 100 não iniciadas).
- Caches locais necessários ao protótipo permanecem em `artifacts/cache/`; o banco isolado e os materiais protegidos permanecem em `artifacts/retained/postgresql-dev/`. Foram preservados porque a execução está ativa; nenhum deles foi adicionado ao Git.
- O PostgreSQL preexistente na porta 5432 não foi usado nem alterado.
- Esta atualização é documental. Não foram executados testes do produto.

## Continuação da retomada — 25/10/2026

- Estado persistido consultado nesta execução: meta `active`, sem limite de orçamento e sem indicação de pausa/conclusão. A consulta de estado atual não expõe o histórico completo de transições. A ocorrência mais recente registrada no chat foi uma falha de leitura do estado da janela do Chrome; por isso, a inspeção do Drive não foi concluída. Não há evidência de que essa falha tenha alterado ou encerrado a meta.
- Repositório: branch `codex/majucau-artifact-consolidation`; o trabalho começou no commit `e55219e`, sincronizado com a origem. O protótipo continuou ativo: processos `majucau-dev` e `majucau-worker`; PostgreSQL/Vite/Wails nas portas locais 55438, 5173 e 34115.
- O arquivo protegido esperado para a credencial Bling (`artifacts/retained/postgresql-dev/credentials/worker-vault/bling_credentials_v1.dpapi`) não existe nesta instância. O nome e o conteúdo de nenhum segredo foram lidos. Portanto, não houve OAuth, chamada real à API nem sincronização.
- Corrigi em `AGENTS.md` a afirmação desatualizada de que XLS não tinha parser e que XLSX só era validado numa pasta. O código valida XLS/XLSX estruturalmente em fluxo nas três raízes autorizadas; somente CSV está ligado ao importador financeiro de recebimentos futuros. Excel continua pendente para importação financeira até contrato e mapeamento aprovados.
- Verificações: `go test -count=1 ./internal/integrations/xlsstream/... ./internal/integrations/xlsxstream/...` passou; o teste focal `TestDailyDrivePipelineValidatesXLSInStreamingAndKeepsItPending` passou; `scripts/verify-backlog.ps1` passou com 127 tarefas, 7 concluídas, 20 em andamento, 100 não iniciadas, 11,8% estimado e 88,2% restante; `git diff --check` passou. O cache temporário desses testes foi removido por `scripts/SafeCleanup.psm1`; caches do protótipo e módulos compartilhados foram preservados.
- BK-041 (adaptador e autenticação Bling) está em 80% e depende do Client Secret real, OAuth e homologação; BK-040 (validar a cobertura dos dados do Bling) ainda está em 0%; BK-042 (leitura Drive) está em 70%. BK-039 (execução de trabalhos em segundo plano) permanece em 50% por ainda faltar concluir e comprovar a integração operacional das fontes e a execução contínua. A paginação real pertence à BK-041. A média do backlog permanece em 11,8%; esses percentuais não representam aceite de homologação. Próximo passo da BK-041: a responsável salva o Client Secret na tela do Majucau, conecta e autoriza o Bling; em seguida validamos a conexão, a paginação e a sincronização inicial. Nenhuma tela ou regra financeira foi alterada.

## Checagem de gravação do cofre — 25/10/2026

- **CI GitHub:** as execuções [#40248166091](https://github.com/guilhermejacyzin/majucau/actions/runs/40248166091) (commit `1a97485`) e [#40249171034](https://github.com/guilhermejacyzin/majucau/actions/runs/40249171034) (commit documental `86dc8ec`) terminaram com resultado geral “sucesso”. Em ambas, `detect-changes` e `verify-backlog` passaram; `verify-postgres` e `verify-windows` foram pulados porque os commits só alteraram documentação. A CI completa #40245975900, registrada no resumo do projeto, passou nos quatro trabalhos. Nenhuma dessas execuções fez uma chamada real ou sincronização do Bling.
- **Estado de Goals:** a meta e o thread continuam `active`. O controle do Chrome falhou numa leitura anterior do Drive, encerrando aquela tentativa. A ferramenta de Goals só expõe o estado atual, sem o histórico de transições; a causa/autoria de uma pausa antiga permanece não comprovada.

- **Estado atualizado em 25/10/2026, 21:31 (Brasília):** o protótipo continua ativo no checkout local do Majucau, branch `codex/majucau-artifact-consolidation`, commit `3908ce1`. Uma consulta somente leitura por Named Pipe ao worker ativo retornou `BLING_STATUS=NOT_CONFIGURED`, `CLIENT_ID_SAVED=false`, `REDIRECT_URI_SAVED=false` e `SECRET_CONFIGURED=false`. O worker atual não recebeu um salvamento completo da configuração. Nenhum segredo foi lido; não houve OAuth, chamada real à API ou sincronização.
- **Verificação do caminho local:** conferi o formulário React, a chamada Wails, o IPC e o armazenamento DPAPI. Os testes passaram: `go test -count=1 ./internal/integrations/bling ./cmd/worker` e Vitest em `src/App.test.tsx`/`src/bootstrap.test.ts` (30 testes). Não encontrei erro de código reproduzível nesse caminho; os testes não validam credenciais reais nem acesso ao Bling.
- Caches do protótipo foram mantidos em `artifacts/cache/` porque a aplicação e o worker estão ativos. O probe temporário usado para ler somente os indicadores booleanos foi removido.
- Próxima ação da BK-041: na tela **Integrações → Bling**, informar o Client Secret do aplicativo e clicar **Salvar configuração**; depois conectar e autorizar. Validar a chamada real, a paginação e a sincronização inicial; se surgir erro, corrigir antes de encerrar BK-041. Não enviar o segredo no chat. BK-039 segue em 50%; BK-041 em 80%; BK-040 em 0%; BK-042 em 70%; projeto em 11,8% concluído e 88,2% restante. Nenhuma regra financeira ou tela foi alterada.
