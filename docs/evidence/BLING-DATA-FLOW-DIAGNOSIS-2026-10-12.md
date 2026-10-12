# Diagnóstico do fluxo de dados do Bling — 2026-10-12

## Conclusão desta revisão

A integração Bling está implementada no código do Majucau. O código também inclui o cabeçalho `enable-jwt: 1` na troca/renovação de tokens e nas consultas autenticadas, conforme o guia de migração JWT do Bling.

Isso não confirma que a conta real esteja conectada ou que uma sincronização tenha sido executada. Na tela atual, salvar credenciais, autorizar a conta, testar a conexão e sincronizar são ações separadas. Abrir o painel não inicia uma sincronização. A tela de sincronização exige uma data ou situação para limitar a consulta.

## Evidências no código

- `internal/integrations/bling/oauth.go`: envia `enable-jwt: 1` ao trocar o código OAuth e ao renovar o token; chamadas de OAuth não repetem automaticamente operações de uso único.
- `internal/integrations/bling/api_client.go`: consulta somente leitura usa `https://api.bling.com.br/Api/v3`, `Authorization: Bearer` e `enable-jwt: 1`.
- `frontend/src/App.tsx`: a área Integrações apresenta as ações “Salvar configuração”, “Conectar”, “Testar conexão” e “Sincronizar agora” separadamente. O painel informa que os dados não são sincronizados automaticamente e exige filtro explícito.
- `cmd/worker/handler.go`: a operação de sincronização retorna `BLING_SYNC_FILTER_REQUIRED` quando não recebe janela de datas ou situação.
- `docs/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md`, BK-039: agendamento, recuperação de trabalho e atualização diária às 09:30 seguem não iniciados. Gisele confirmou `America/Sao_Paulo` em 12/10; o comportamento para arquivos externos ausentes continua por definir.

## Evidência de execução e limite

Na janela atual do worker, iniciada em 12/10 às 17:23, aparecem somente consultas de saúde do worker, estado das integrações e painel. Não aparecem chamadas para iniciar autorização OAuth, testar a conexão ou sincronizar o Bling. Isso confirma que, nessa sessão, o Majucau ainda não recebeu o comando para conectar/testar/sincronizar; portanto, não há evidência de que o adapter tenha falhado ao consultar a API.

O worker informou `DIAGNOSTICS_LOG_UNAVAILABLE` porque a pasta de logs de instalação em `%ProgramData%\Majucau` não existe neste ambiente de desenvolvimento. O worker continuou funcionando e enviou os registros sanitizados para `artifacts/retained/postgresql-dev/worker.stderr.log`, conforme o modo de inicialização local. Esse aviso não bloqueou chamadas ao Bling. O print da tela Integrações recebido em 12/10/2026 mostra o Bling como “Não conectado”, Client ID e Redirect URI vazios, permissões “aguardando autorização” e “Último sucesso: —”. As bolinhas no campo Client Secret são o texto de exemplo definido pela interface e não provam que um segredo foi salvo. Não registrar Client ID, Client Secret, tokens, payloads ou valores reais como evidência.

## Conferência do print e do painel — 12/10/2026

- O código já envia `Authorization: Bearer` e `enable-jwt: 1` nas consultas autenticadas. Isso confirma o cabeçalho JWT no código, mas não uma autorização da conta real.
- A configuração, a autorização, o teste e a sincronização são etapas separadas. O painel não inicia a sincronização; a primeira tentativa depende de preencher a configuração no Majucau, autorizar no Bling e iniciar um filtro de sincronização.
- No build mostrado no print da Visão Executiva, “Última sincronização” usava `snapshot.as_of`, que é a hora da consulta ao painel. A correção agora usa o fim do lote Bling mais recente com estado `SUCCESS` ou `PARTIAL`; sem lote, o valor fica “—”. OAuth e teste de conexão não contam como sincronização.
- A métrica “A receber” era marcada como `CONFIRMED` mesmo com zero registros, permitindo mostrar R$ 0,00 sem dados. Agora, zero registros resulta em estado `UNAVAILABLE`, que a tela existente já exibe como “—”. O layout e os textos não mudaram; nenhuma regra financeira foi alterada. Uma sincronização com cobertura suficiente para confirmar saldo zero ainda exigiria metadados de cobertura, que o lote atual não guarda.
- A tentativa de leitura direta do site do Bling foi recusada pela política de permissão salva do navegador, apesar dos prints anteriores das configurações. Não foi usado outro caminho para contornar o bloqueio.
- O log mais recente continua contendo apenas saúde do worker, status de integração e consulta do painel; não há operação OAuth, teste ou sincronização Bling registrada.

## Conferência do endereço de retorno OAuth — 12/10/2026

Foi identificada uma incompatibilidade que pode impedir a conexão:

- `frontend/src/App.tsx` mostra `https://...` como exemplo no campo “Redirect URI”.
- `internal/integrations/bling/credentials.go` aceita HTTPS ao salvar, mas `internal/integrations/bling/oauth_service.go` só inicia a autorização com `http://127.0.0.1:<porta>/<caminho>` e abre essa porta local para receber o retorno. Assim, uma URL HTTPS pode ser salva e depois falhar ao clicar em “Conectar”.
- A [documentação oficial do Bling sobre aplicativos](https://developer.bling.com.br/aplicativos) diz que o endereço cadastrado é usado no retorno da autorização e que o Client ID e o Client Secret aparecem nas informações do aplicativo. Ela não confirma se o cadastro do Bling aceita o endereço local exigido pelo Majucau.

Antes de orientar o preenchimento, precisamos confirmar que o aplicativo cadastrado no Bling aceita o endereço local e a porta que o Majucau usa. Não inventar um endereço nem pedir credenciais pelo chat. A diferença entre a dica visual e o contrato do worker fica registrada para uma correção técnica, sem alteração da tela aprovada.

## Próxima verificação

Após reiniciar o protótipo, confirmar primeiro se o cadastro do aplicativo Bling aceita o endereço local exigido pelo worker e se a porta/caminho coincidem exatamente. Só então preencher no Majucau o Client ID, Client Secret e Redirect URI, salvar, conectar e autorizar a conta. Em seguida, testar a conexão e iniciar a sincronização com um filtro aprovado. Não enviar segredos pelo chat. A rotina diária das 09:30 em `America/Sao_Paulo` continua pendente de implementação; também falta definir o comportamento quando arquivos externos ainda não chegaram.

Esta correção foi revisada estaticamente e formatada, mas os testes e o protótipo não foram executados nesta etapa. A validação em execução continua pendente após a reinicialização e a conexão real do Bling.

## Revalidação do JWT e do retorno OAuth — 12/10/2026

- A documentação oficial atual exige `enable-jwt: 1` na obtenção do token, na renovação e nas chamadas autenticadas, além do formato `Authorization: Bearer`. O código do Majucau já segue esses requisitos em `internal/integrations/bling/oauth.go` e `api_client.go`. A base configurada também é a URL nova `https://api.bling.com.br/Api/v3`; a URL antiga foi descontinuada em 30/04/2026. Referências: [migração JWT](https://developer.bling.com.br/migracao-jwt), [aplicativos e OAuth](https://developer.bling.com.br/aplicativos) e [migração da URL base](https://developer.bling.com.br/migracao-url).
- O print de Integrações mostra Client ID e Redirect URI vazios e status “Não conectado”. As bolinhas do campo de segredo não comprovam que haja segredo salvo. Portanto, o fluxo OAuth não começou; não é evidência de que o cabeçalho JWT esteja falhando.
- Há uma incompatibilidade de configuração que precisa ser esclarecida: a interface sugere uma URL `https://`, mas o serviço local só aceita e abre um callback de loopback `http://127.0.0.1:<porta>/<caminho>`. A documentação do Bling determina que o retorno use o “Link de redirecionamento” cadastrado e que a URL seja a mesma no fluxo de troca de token, mas não confirma se esse tipo de endereço local é aceito pelo cadastro.
- Não presumir um endereço, não inserir credenciais pelo chat e não mudar a tela ou o OAuth até verificar a URL que está cadastrada e se o Bling aceita o callback local. Caso o cadastro exija HTTPS, será necessária uma solução de retorno aprovada; simplesmente salvar HTTPS não funciona com o listener loopback atual.
- Para solicitar dados, o protótipo exige configuração e autorização OAuth, depois “Testar conexão” e “Sincronizar agora” com um filtro de data ou situação. Abrir a Visão Executiva não inicia leitura da API.

Esta verificação usou documentação pública oficial e inspeção estática do código. Nenhuma chamada foi feita à conta real, nenhum teste foi executado e nenhuma tela, regra financeira ou credencial foi alterada.

## Proteções

Esta revisão não alterou telas aprovadas, regras financeiras, configuração do Bling, banco de dados ou credenciais. Nenhuma chamada foi feita à conta real nesta execução; os resultados descritos acima são inspeção do código e da janela de logs já coletada.

## Rechecagem do protótipo e do acesso — 12/10/2026

- A permissão do navegador mostrada pela responsável lista `https://www.bling.com.br` como “Sempre permitir”, mas a ferramenta de navegação do assistente ainda recusou a leitura do site por uma preferência salva. Não houve tentativa por outro navegador ou caminho.
- A interface `majucau-dev.exe` em execução iniciou às 18:07:01; o commit `3bf03a9`, que corrige os dois indicadores, foi criado às 18:07:49. Portanto, a instância aberta começou antes da correção e precisa ser reaberta para exibi-la.
- O log sanitizado do worker terminou às 18:14:44 e contém chamadas `health.get`, `integration.status` e `dashboard.snapshot`; não registra início OAuth, teste de conexão nem sincronização. Isso reforça que o adapter ainda não recebeu uma operação para consultar o Bling.
- A instância não foi encerrada durante esta conferência para preservar o protótipo aberto. A validação visual da correção permanece pendente de reabertura; a conexão real permanece pendente de configuração/autorização e de uma sincronização filtrada.

Não foram alterados código, telas, regras financeiras, credenciais ou dados. Nenhum teste foi executado.

## Nova tentativa de acesso ao Bling — 12/10/2026

- Depois da responsável confirmar a permissão e autorizar o acesso, tentei abrir a aba de login já existente do Bling no Chrome. A ferramenta recusou a navegação por uma preferência de segurança salva para `www.bling.com.br`, apesar do print das configurações mostrar “Sempre permitir”.
- A recusa ocorreu antes de abrir a página; não foi possível clicar em “Entrar”, consultar a área de desenvolvedores nem fazer qualquer chamada à API. Não tentei outro navegador ou outro caminho.
- Isso não aponta falha da senha nem da API. O print anterior da tela Integrações segue mostrando Client ID e Redirect URI vazios e o Bling como “Não conectado”; a configuração OAuth da integração continua pendente.
- O próximo diagnóstico de conexão deve confirmar o endereço de retorno OAuth aceito/cadastrado no app Bling e compatibilizá-lo com o callback loopback exigido pelo worker, sem alterar a tela aprovada. Não preencher um endereço presumido nem compartilhar Client ID/segredo no chat.
- Nova tentativa após a responsável confirmar “Sempre permitir” para `https://www.bling.com.br` foi bloqueada pela mesma preferência salva na ferramenta. Nenhum acesso alternativo foi tentado; a causa visível continua sendo a integração não configurada/autorizada, conforme o print da tela Integrações.
