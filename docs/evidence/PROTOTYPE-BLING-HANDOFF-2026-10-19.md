# Ponto de parada do protótipo e Bling — 19/10/2026

## Estado conferido

- Às 09:00 (horário de Brasília), iniciei novamente o protótipo pelo roteiro oficial `scripts/start-prototype.ps1`.
- O PostgreSQL 18 separado do Majucau está limitado a `127.0.0.1:55438`; o worker e a interface Wails/Vite iniciaram. A execução permanece aberta para a configuração pela usuária.
- A consulta sanitizada do estado da integração retornou `NOT_CONFIGURED` e `secret_configured:false`. Nenhum Client Secret, token, dado financeiro ou conteúdo do Drive foi lido ou registrado neste arquivo.
- A revisão estática do fluxo confirmou que a tela encaminha o segredo para o salvamento protegido e limpa o campo após a operação (`frontend/src/App.tsx`, `internal/integrations/bling/credentials.go` e `cmd/worker/handler.go`). Não foi encontrado defeito de código que explique a credencial ainda ausente.
- Não houve autorização OAuth, chamada real à API ou sincronização do Bling nesta execução. Testes de produto não foram executados durante este registro documental.
- Nenhuma tela aprovada nem regra financeira foi alterada.

## Retomada, sem trocar de tarefa antes de concluir

1. Na tela **Importações/Integrações → Bling** do Majucau aberto, a usuária informa o Client Secret do aplicativo diretamente no campo e clica **Salvar configuração**. O segredo não deve ser enviado pelo chat.
2. Na mesma tela, clica **Conectar** e conclui a autorização na página oficial do Bling.
3. Confirmar no Majucau o estado conectado, executar **Testar conexão** e conferir a paginação real. Se a autorização, chamada ou paginação falhar, corrigir esta mesma frente antes de iniciar outra.
4. Executar a sincronização inicial e conferir se os dados são recebidos e publicados com segurança; só declarar a frente homologada após evidência real, sem usar dados fictícios como comprovação.
5. Depois da conclusão da frente Bling, provisionar separadamente a identidade técnica de leitura do Drive, restrita às três pastas aprovadas, e homologar a atualização diária das 09:30.
6. Conferir os resultados do protótipo com dados de homologação e então seguir as demais tarefas de finalização do backlog.

## Governança e limpeza

- Backlog: 127 tarefas; 7 concluídas, 20 em andamento, 100 não iniciadas. Estimativa média do projeto: 11,8% concluído e 88,2% restante.
- BK-039: 50% concluído e 50% restante. Esses percentuais não significam que o protótipo esteja homologado.
- Os caches e arquivos de apoio da execução permanecem dentro de `artifacts/cache/`; o cluster e os materiais locais protegidos permanecem em `artifacts/retained/postgresql-dev/`. Estão ignorados pelo Git e são necessários enquanto o protótipo estiver aberto. O registro local de limpeza `artifacts/retained/cleanup-register.md` anota a preservação e a remoção segura depois que a execução for encerrada.
- O PostgreSQL preexistente na porta 5432 não foi usado nem alterado.
