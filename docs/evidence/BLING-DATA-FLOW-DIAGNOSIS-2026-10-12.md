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

Na janela de logs examinada durante o diagnóstico em 12/10, apareceram consultas de saúde do worker, estado das integrações e painel. Não apareceram eventos de início/autorização OAuth, teste da conexão ou `bling.sync`. Isso significa que não há evidência registrada de que a conta tenha sido autorizada ou de que a sincronização tenha sido iniciada naquela janela. Não prova, sozinho, que o adapter falhou.

O log técnico também informou `DIAGNOSTICS_LOG_UNAVAILABLE`, então a observabilidade daquela execução foi incompleta. O estado efetivo da conta precisa ser confirmado pelo indicador da integração no próprio Majucau. Não registrar Client ID, Client Secret, tokens, payloads ou valores reais como evidência.

## Próxima verificação

Consultar na área Integrações o status do Bling e, se disponível, “Última tentativa” e “Último sucesso”. Se estiver conectado sem sucesso anterior, executar primeiro “Testar conexão”; depois, com um filtro autorizado, usar “Sincronizar agora” e registrar somente código de erro, mensagem sanitizada e contadores. Não presumir uma janela financeira nem ativar sincronização automática antes de concluir BK-039 e as decisões de agendamento pendentes.

## Proteções

Esta revisão não alterou telas aprovadas, regras financeiras, configuração do Bling, banco de dados ou credenciais. Nenhuma chamada foi feita à conta real nesta execução; os resultados descritos acima são inspeção do código e da janela de logs já coletada.
