# Conexão do Bling e preparo para JWT — 13/10/2026

## Resultado da verificação

O teste foi feito no protótipo local, sem iniciar uma sincronização nem importar registros financeiros.

- A tela mostra o Bling como **Não conectado**, sem data de último sucesso.
- O teste de conexão exibiu `BLING_OAUTH_CONFIG_INVALID` e uma instrução genérica sobre Redirect URI loopback.
- A revisão dos diagnósticos sanitizados do worker encontrou tentativas de salvar a configuração encerradas com `BLING_SECRET_REQUIRED`. O Client Secret do aplicativo não foi aceito/salvo. A conta e a senha usadas para entrar no site do Bling são diferentes desse segredo.
- O fluxo de teste usa o registro salvo no banco local. Como a configuração não foi salva, o erro posterior não comprova que o endereço de retorno hospedado esteja incorreto.
- O endereço exato de retorno `https://majucau-bling-bridge.giselegalha.chatgpt.site/oauth/bling/callback` está autorizado no código atual. A aplicação o encaminha para um endereço temporário local, verifica o `state` e troca o código usando o mesmo endereço registrado. Os testes do código cobrem esse fluxo. Não houve autorização OAuth real nesta verificação.

## JWT e endereço da API

A documentação atual do Bling pede o cabeçalho `enable-jwt: 1` ao trocar o código OAuth por tokens, ao renovar tokens e nas chamadas autenticadas à API.

O código atual atende a esses três pontos:

- `internal/integrations/bling/oauth.go` inclui `enable-jwt: 1` na troca e na renovação dos tokens;
- `internal/integrations/bling/api_client.go` e `api_resources.go` incluem o cabeçalho nas consultas autenticadas;
- `internal/integrations/bling/api_client.go` usa a base atual `https://api.bling.com.br/Api/v3`.

Portanto, a implementação no repositório está preparada para a migração JWT. A conexão real ainda não foi comprovada porque o Client Secret não foi salvo e não houve autorização da conta.

Fontes oficiais consultadas em 13/10/2026:

- [Bling — migração para autenticação JWT](https://developer.bling.com.br/migracao-jwt)
- [Bling — descontinuação da URL antiga da API](https://developer.bling.com.br/migracao-url)

## Próximo passo seguro

Na tela **Integrações → Bling**, inserir o **Client Secret do aplicativo** no campo correspondente e clicar em **Salvar configuração**. O segredo deve ser digitado diretamente no protótipo e não enviado pelo chat. Depois de salvar com sucesso, usar **Conectar**, autorizar o Majucau no Bling e então testar a leitura mínima. A sincronização continua sendo uma ação separada.

## Limites

- Nenhum Client Secret, senha, token, código OAuth, dado de cliente ou valor financeiro foi incluído neste registro.
- O teste executado não gravou nem sincronizou dados financeiros.
- Nenhuma tela aprovada, regra financeira ou regra de negócio foi alterada.
- A conexão real, a sincronização de ponta a ponta e a homologação dos dados continuam pendentes.
