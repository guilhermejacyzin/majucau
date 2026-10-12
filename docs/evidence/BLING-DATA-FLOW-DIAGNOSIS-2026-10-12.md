# Diagnóstico do fluxo de dados do Bling — 2026-10-12

## Conclusão desta revisão

A integração Bling está implementada no código do Majucau. O código também inclui o cabeçalho `enable-jwt: 1` na troca/renovação de tokens e nas consultas autenticadas, conforme o guia de migração JWT do Bling.

Isso não confirma que a conta real esteja conectada ou que uma sincronização tenha sido executada. Na tela atual, salvar credenciais, autorizar a conta, testar a conexão e sincronizar são ações separadas. Abrir o painel não inicia uma sincronização. A tela de sincronização exige uma data ou situação para limitar a consulta.

## Evidências no código

- `internal/integrations/bling/oauth.go`: envia `enable-jwt: 1` ao trocar o código OAuth e ao renovar o token; chamadas de OAuth não repetem automaticamente operações de uso único.
- `internal/integrations/bling/api_client.go`: consulta somente leitura usa `https://api.bling.com.br/Api/v3`, `Authorization: Bearer` e `enable-jwt: 1`.
- `frontend/src/App.tsx`: a área Integrações apresenta as ações “Salvar configuração”, “Conectar”, “Testar conexão” e “Sincronizar agora” separadamente. O painel informa que os dados não são sincronizados automaticamente e exige filtro explícito.
- `cmd/worker/handler.go`: a operação de sincronização retorna `BLING_SYNC_FILTER_REQUIRED` quando não recebe janela de datas ou situação.
- `docs/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md`, BK-039: agendamento, recuperação de trabalho e atualização diária às 09:30 seguem não iniciados; o comportamento para arquivos externos ausentes e o fuso horário continuam por definir.

## Evidência de execução e limite

Na janela atual do worker, iniciada em 12/10 às 17:23, aparecem somente consultas de saúde do worker, estado das integrações e painel. Não aparecem chamadas para iniciar autorização OAuth, testar a conexão ou sincronizar o Bling. Isso confirma que, nessa sessão, o Majucau ainda não recebeu o comando para conectar/testar/sincronizar; portanto, não há evidência de que o adapter tenha falhado ao consultar a API.

O worker informou `DIAGNOSTICS_LOG_UNAVAILABLE` porque a pasta de logs de instalação em `%ProgramData%\Majucau` não existe neste ambiente de desenvolvimento. O worker continuou funcionando e enviou os registros sanitizados para `artifacts/retained/postgresql-dev/worker.stderr.log`, conforme o modo de inicialização local. Esse aviso não bloqueou as chamadas ao Bling. O estado efetivo da conta precisa ser confirmado pelo indicador da integração no próprio Majucau. Não registrar Client ID, Client Secret, tokens, payloads ou valores reais como evidência.

## Próxima verificação

Consultar na área Integrações o status do Bling e, se disponível, “Última tentativa” e “Último sucesso”. Se estiver conectado sem sucesso anterior, executar primeiro “Testar conexão”; depois, com um filtro autorizado, usar “Sincronizar agora” e registrar somente código de erro, mensagem sanitizada e contadores. Não presumir uma janela financeira nem ativar sincronização automática antes de concluir BK-039 e as decisões de agendamento pendentes.

## Proteções

Esta revisão não alterou telas aprovadas, regras financeiras, configuração do Bling, banco de dados ou credenciais. Nenhuma chamada foi feita à conta real nesta execução; os resultados descritos acima são inspeção do código e da janela de logs já coletada.
