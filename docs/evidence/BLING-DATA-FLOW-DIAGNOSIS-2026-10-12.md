# Diagnóstico do fluxo de dados do Bling — 2026-10-12

## Conclusão desta revisão

A integração Bling está implementada no código do Majucau. O código também inclui o cabeçalho `enable-jwt: 1` na troca/renovação de tokens e nas consultas autenticadas, conforme o guia de migração JWT do Bling.

Isso não confirma que a conta real esteja conectada ou que uma sincronização tenha sido executada. Na tela atual, salvar credenciais, autorizar a conta, testar a conexão e sincronizar são ações separadas. Abrir o painel não inicia uma sincronização. A tela de sincronização exige uma data ou situação para limitar a consulta.

## Evidências no código

- `internal/integrations/bling/oauth.go`: envia `enable-jwt: 1` ao trocar o código OAuth e ao renovar o token; chamadas de OAuth não repetem automaticamente operações de uso único.
- `internal/integrations/bling/api_client.go`: consulta somente leitura usa `https://api.bling.com.br/Api/v3`, `Authorization: Bearer` e `enable-jwt: 1`.
- `frontend/src/App.tsx`: a área Integrações apresenta as ações “Salvar configuração”, “Conectar”, “Testar conexão” e “Sincronizar agora” separadamente. O painel informa que os dados não são sincronizados automaticamente e exige filtro explícito.
- `cmd/worker/handler.go`: a operação de sincronização retorna `BLING_SYNC_FILTER_REQUIRED` quando não recebe janela de datas ou situação.
- `docs/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md`, BK-039: agendamento e recuperação de trabalho seguem não iniciados. Gisele aprovou atualização diária às 09:30 em `America/Sao_Paulo`, releitura de todas as contas do Bling, processamento independente das fontes disponíveis e pendência dos arquivos ausentes até às 09:30 do dia seguinte. A coleta não pode ser marcada completa até o fim da paginação ser comprovado. Também falta configurar acesso persistente do serviço ao Drive.

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

Após reiniciar o protótipo, confirmar primeiro se o cadastro do aplicativo Bling aceita o endereço local exigido pelo worker e se a porta/caminho coincidem exatamente. Só então preencher no Majucau o Client ID, Client Secret e Redirect URI, salvar, conectar e autorizar a conta. Em seguida, testar a conexão e iniciar a sincronização com um filtro aprovado. Não enviar segredos pelo chat. A rotina diária às 09:30 em `America/Sao_Paulo`, a releitura de todas as contas do Bling e a regra para arquivos ausentes já foram aprovadas; a implementação segue em BK-039. A documentação consultada não mostra como confirmar a última página, portanto essa prova ainda é necessária antes de considerar a coleta completa. Também falta definir acesso persistente ao Drive pelo serviço.

Esta correção foi revisada estaticamente e formatada, mas os testes e o protótipo não foram executados nesta etapa. A validação em execução continua pendente após a reinicialização e a conexão real do Bling.

## Revalidação do JWT e do retorno OAuth — 12/10/2026

- A documentação oficial atual exige `enable-jwt: 1` na obtenção do token, na renovação e nas chamadas autenticadas, além do formato `Authorization: Bearer`. O código do Majucau já segue esses requisitos em `internal/integrations/bling/oauth.go` e `api_client.go`. A base configurada também é a URL nova `https://api.bling.com.br/Api/v3`; a URL antiga foi descontinuada em 30/04/2026. Referências: [migração JWT](https://developer.bling.com.br/migracao-jwt), [aplicativos e OAuth](https://developer.bling.com.br/aplicativos) e [migração da URL base](https://developer.bling.com.br/migracao-url).
- O print de Integrações mostra Client ID e Redirect URI vazios e status “Não conectado”. As bolinhas do campo de segredo não comprovam que haja segredo salvo. Portanto, o fluxo OAuth não começou; não é evidência de que o cabeçalho JWT esteja falhando.
- Há uma incompatibilidade de configuração que precisa ser esclarecida: a interface sugere uma URL `https://`, mas o serviço local só aceita e abre um callback de loopback `http://127.0.0.1:<porta>/<caminho>`. A documentação do Bling determina que o retorno use o “Link de redirecionamento” cadastrado e que a URL seja a mesma no fluxo de troca de token, mas não confirma se esse tipo de endereço local é aceito pelo cadastro.
- Não presumir um endereço, não inserir credenciais pelo chat e não mudar a tela ou o OAuth até verificar a URL que está cadastrada e se o Bling aceita o callback local. Caso o cadastro exija HTTPS, será necessária uma solução de retorno aprovada; simplesmente salvar HTTPS não funciona com o listener loopback atual.
- Para solicitar dados, o protótipo exige configuração e autorização OAuth, depois “Testar conexão” e “Sincronizar agora” com um filtro de data ou situação. Abrir a Visão Executiva não inicia leitura da API.

Esta verificação usou documentação pública oficial e inspeção estática do código. Nenhuma chamada foi feita à conta real, nenhum teste foi executado e nenhuma tela, regra financeira ou credencial foi alterada.

## Atualização do escopo diário aprovada — 12/10/2026

- Gisele decidiu reler todas as contas do Bling diariamente, em vez de limitar a coleta automática a alterações desde o último checkpoint ou aos próximos 60 dias. A tela manual mantém os filtros já aprovados: vencimento consulta ambas as listas, recebimento somente Contas a Receber e pagamento somente Contas a Pagar.
- Se um arquivo externo não estiver disponível às 09:30, ele continua pendente até a atualização das 09:30 do dia seguinte. As demais fontes disponíveis devem poder concluir de forma independente.
- A releitura completa não pode ser ativada enquanto o serviço não souber confirmar, sem suposição, que percorreu todas as páginas. O acesso do worker ao Google Drive e a conexão OAuth do Bling também continuam pendentes.
- Esta decisão muda o escopo da coleta automática; não altera telas aprovadas, cálculos, filtros manuais ou regras financeiras.

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

## Fechamento do diagnóstico da ausência de dados — 12/10/2026

- A conferência do código e do print fecha o diagnóstico desta etapa: o Majucau já usa `Authorization: Bearer`, `enable-jwt: 1` e o endereço atual da API do Bling. A evidência disponível não aponta falha na regra JWT.
- O print mostra o Bling “Não conectado”, Client ID e Redirect URI vazios e permissões “aguardando autorização”. Os pontos no campo de senha são o exemplo visual da tela, não prova de Client Secret salvo. A senha guardada no Chrome é a senha de entrada na conta Bling; ela não configura as credenciais do aplicativo que o Majucau exige.
- A responsável autorizou o acesso e mostrou a permissão como “Sempre permitir”. Mesmo assim, a ferramenta de navegador bloqueou a tentativa na aba existente antes de abrir o site. Não foi usado outro navegador, chamada indireta ou caminho alternativo para contornar a proteção.
- Como o fluxo OAuth não foi iniciado e os logs disponíveis não mostram comando de conexão, teste ou sincronização, não há evidência de que o Majucau tenha solicitado dados ao Bling nesta sessão. Abrir o painel não inicia a coleta.
- Continua pendente confirmar o endereço de retorno registrado no aplicativo Bling: a tela sugere HTTPS, enquanto o serviço local só recebe callback em `http://127.0.0.1:<porta>/<caminho>`. Não alterei a tela nem o OAuth, pois o endereço aceito pelo aplicativo não foi confirmado.
- Próxima etapa para homologar os dados: configurar localmente Client ID, Client Secret e o endereço de retorno aceito; autorizar a conta; testar a conexão; então iniciar uma sincronização com filtro. Nunca enviar Client Secret ou token pelo chat.
- O diagnóstico da ausência de chamadas está concluído. BK-041 continua em 69% (31% restante), pois ainda faltam conexão real, contrato de paginação/cursor e homologação. Nenhuma tela aprovada, regra financeira, regra de negócio ou credencial foi alterada. Nenhum teste foi executado nesta etapa.

## Correção dos filtros de data do Bling — 12/10/2026

- A responsável aprovou que “Vencimento” consulte as duas listas; “Recebimento”, somente Contas a Receber; e “Pagamento”, somente Contas a Pagar.
- O cliente agora envia para cada rota os nomes documentados de parâmetros. Ele recusa filtros de data incompatíveis e uma chamada que tente combinar dois tipos de data.
- A sincronização não faz mais um teste extra em Contas a Receber antes de coletar Contas a Pagar. Ela atualiza/renova o token protegido sem ler uma lista financeira e em seguida chama somente o recurso permitido pelo filtro.
- A atomicidade por recurso continua: cada lista é persistida com lote/checkpoint dentro da própria transação; quando vencimento consulta ambas e a segunda falha, o resultado geral permanece `PARTIAL`. A idempotência RAW existente não mudou.
- A documentação pública ainda não confirma o filtro “alterado desde o último checkpoint” nem o marcador de última página de Contas a Receber. Portanto esta mudança corrige consultas manuais, mas não ativa a atualização diária automática.
- A resposta do Bling sem metadados de fim continua levando a falha segura com rollback; a conexão e homologação com a conta Majucau seguem pendentes. A tela aprovada não foi editada. Os testes de regressão foram atualizados, mas não executados localmente nesta etapa; a CI será a validação automatizada após o push.

## Endereço de retorno OAuth confirmado — 12/10/2026

- A responsável informou que o campo “Link de redirecionamento” do aplicativo Bling aponta para um callback HTTPS externo em `/oauth/bling/callback`. O host não foi copiado para o repositório público; o valor completo continua disponível na conversa e no cadastro do aplicativo.
- A documentação oficial do Bling identifica esse campo como o destino do retorno após a autorização. O authorization server redireciona o navegador para o endereço cadastrado, com `code` e `state`; consulte [Aplicativos e OAuth do Bling](https://developer.bling.com.br/aplicativos).
- O código atual do Majucau aceita HTTPS ao salvar a configuração, mas o início do OAuth exige `http://127.0.0.1:<porta>/<caminho>` e abre um listener local nessa porta. Portanto, o HTTPS informado não é compatível com o callback implementado e a conexão não concluirá com essa configuração.
- Não encontrei neste repositório um serviço que receba o callback HTTPS e o encaminhe com segurança ao worker local. A existência e o funcionamento desse encaminhamento ainda precisam ser confirmados. Não alterei OAuth, credenciais ou tela; nenhuma regra financeira foi afetada.
- Próximo passo: confirmar se o endereço HTTPS é uma ponte funcional para o Majucau instalado. Se não for, escolher entre cadastrar um callback local aceito pelo Bling ou aprovar e desenhar uma ponte HTTPS segura antes de alterar o fluxo.

## Verificação do Site de retorno — 12/10/2026

- A responsável confirmou o endereço HTTPS cadastrado no campo “Link de redirecionamento”. O host não é repetido nesta evidência pública.
- A consulta somente de leitura ao inventário de Sites encontrou um Site ativo e público destinado à ponte Bling. A versão mais recente publicada é a 9; não há repositório de código conectado para inspecionar sua implementação.
- A consulta aos registros recentes do Site não encontrou chamadas no período consultado. Isso não prova que a ponte esteja quebrada, mas também não comprova que ela encaminhe `code` e `state` ao Majucau local.
- O worker do Majucau continua aceitando somente um callback local `http://127.0.0.1:<porta>/<caminho>`. Ainda falta confirmar o comportamento da ponte ou obter seu código para revisão antes de alterar o OAuth.
- Nenhuma chamada de autorização foi iniciada, nenhum segredo foi consultado e nenhuma configuração ou tela foi modificada. Nenhuma regra financeira foi afetada.
