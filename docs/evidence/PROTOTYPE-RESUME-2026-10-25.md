# Retomada do protótipo — 25/10/2026

## Situação

- Às 20:03 (horário de Brasília), a pedido da usuária, iniciei `scripts/start-prototype.ps1` novamente.
- O PostgreSQL isolado está ouvindo em `127.0.0.1:55438`; a interface Wails/Vite está ativa nas portas locais `34115` e `5173`.
- A execução anterior não persistiu como processo entre as sessões. O protótipo foi iniciado novamente nesta retomada.
- A pasta DPAPI ativa contém somente `.entropy` e `.scope`; não existe `bling_credentials_v1.dpapi`, que é o arquivo definido pelo código para a credencial do Bling. Isso confirma que o segredo não foi salvo nesta instância. Nenhum segredo foi lido.
- Não houve OAuth, chamada real, sincronização de dados, alteração de código, tela ou regra financeira nesta retomada.
- A branch está sincronizada com `origin/codex/majucau-artifact-consolidation` no commit `bf9481a`.

## Próximo passo

1. Na tela **Importações/Integrações → Bling** do protótipo aberto, informar o Client Secret do aplicativo diretamente, salvar, conectar e concluir a autorização. Não enviar o segredo no chat.
3. Confirmar o estado conectado, testar a API, validar a paginação e realizar a sincronização inicial. Corrigir qualquer falha antes de encerrar BK-039.

## Progresso e proteção

- BK-039: 50% concluída; 50% restante.
- Projeto: estimativa do backlog de 11,8% concluído; 88,2% restante (127 tarefas: 7 concluídas, 20 em andamento e 100 não iniciadas).
- Caches locais necessários ao protótipo permanecem em `artifacts/cache/`; o banco isolado e os materiais protegidos permanecem em `artifacts/retained/postgresql-dev/`. Foram preservados porque a execução está ativa; nenhum deles foi adicionado ao Git.
- O PostgreSQL preexistente na porta 5432 não foi usado nem alterado.
- Esta atualização é documental. Não foram executados testes do produto.
