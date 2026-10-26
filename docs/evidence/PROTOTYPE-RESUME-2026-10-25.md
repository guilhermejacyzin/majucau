# Retomada do protótipo — 25/10/2026

## Situação

- Às 20:03 (horário de Brasília), a pedido da usuária, iniciei `scripts/start-prototype.ps1` novamente.
- O PostgreSQL isolado está ouvindo em `127.0.0.1:55438`; a interface Wails/Vite está ativa nas portas locais `34115` e `5173`.
- A execução anterior não persistiu como processo entre as sessões. O protótipo foi iniciado novamente nesta retomada.
- A pasta DPAPI ativa contém somente `.entropy` e `.scope`; não existe `bling_credentials_v1.dpapi`, que é o arquivo definido pelo código para a credencial do Bling. Isso confirma que o segredo não foi salvo nesta instância. Nenhum segredo foi lido.
- Não houve OAuth, chamada real, sincronização de dados, alteração de código, tela ou regra financeira nesta retomada.
- Fotografia inicial desta evidência: a branch estava sincronizada no commit `bf9481a`; a atualização corrente está registrada abaixo.

## Próximo passo

1. Na tela **Importações/Integrações → Bling** do protótipo aberto, informar o Client Secret do aplicativo diretamente, salvar, conectar e concluir a autorização. Não enviar o segredo no chat.
3. Confirmar o estado conectado, testar a API, validar a paginação e realizar a sincronização inicial. Corrigir qualquer falha antes de encerrar BK-039.

## Progresso e proteção

- BK-039: 50% concluída; 50% restante.
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
- BK-039 permanece em 50% até OAuth, paginação e carga real serem homologados. A média do backlog permanece em 11,8%; esses percentuais não representam aceite de homologação. Próximo passo: a responsável salva o Client Secret na própria tela do Majucau, conecta e autoriza o Bling; em seguida, validar a conexão, o fim da paginação e a sincronização inicial. Nenhuma tela ou regra financeira foi alterada.

## Checagem de gravação do cofre — 25/10/2026

- O protótipo continua ativo no checkout `C:\Users\Gisele\Documents\Codex\2026-10-02\api-2\majucau` (branch `codex/majucau-artifact-consolidation`, commit `cbb2abc`).
- A pasta esperada do cofre contém `.entropy` e `.scope`, mas não contém `bling_credentials_v1.dpapi`. A ACL concede `FullControl` ao usuário ativo, a SYSTEM e a Administradores, sem regra explícita de negação. Isso não indica bloqueio de escrita pela ACL.
- Nenhum segredo foi lido ou registrado. A ausência do arquivo não prova falha do produto: falta confirmar o resultado da gravação na tela. Não houve OAuth, chamada real nem sincronização.
- Próxima ação de BK-039: informar o Client Secret diretamente na tela **Integrações → Bling**, clicar **Salvar configuração** e então conectar/autorizar. Depois verificar a sincronização inicial e a paginação. BK-039 segue 50%; backlog 11,8% concluído/88,2% restante. Nenhuma regra financeira ou tela foi alterada.
