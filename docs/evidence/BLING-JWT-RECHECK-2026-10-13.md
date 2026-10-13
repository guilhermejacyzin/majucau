# Reconfirmação de compatibilidade JWT do Bling — 13/10/2026

## O que foi conferido

- O [guia oficial de migração JWT do Bling](https://developer.bling.com.br/migracao-jwt) pede o cabeçalho `enable-jwt: 1` na troca do código de autorização por tokens, na renovação e em todas as chamadas autenticadas.
- O guia consultado informa que a data geral de bloqueio dos tokens antigos ainda está em definição. O aviso apresentado na conta Majucau informa 15/10; essa data específica da conta deve ser tratada como prazo operacional.
- No código Majucau, `internal/integrations/bling/oauth.go` envia `enable-jwt: 1` na troca e na renovação dos tokens. `api_client.go` e `api_resources.go` enviam o mesmo cabeçalho junto de `Authorization: Bearer ...` nas consultas autenticadas.

## Verificação executada

- `go test -count=1 ./internal/integrations/bling` passou.
- `go vet ./internal/integrations/bling` passou.
- Os dois comandos usaram pastas temporárias exclusivas dentro de `artifacts/cache/`; essas pastas foram removidas após a verificação.

## Estado da integração real

- O Client ID e o endereço de retorno foram informados. O endereço corresponde ao callback HTTPS aprovado para encaminhar a autorização ao Majucau local.
- Ao salvar as configurações locais, o Majucau pediu o Client Secret (`BLING_SECRET_REQUIRED`). O segredo não foi consultado nem copiado para o chat. Sem ele, não houve autorização real, teste de conexão ou chamada autenticada ao Bling.
- Conclusão: o código está alinhado com o requisito JWT publicado, mas a conta ainda precisa ser conectada e os dados reais precisam ser lidos e conferidos para provar a integração ponta a ponta.

Nenhuma tela aprovada ou regra financeira foi alterada.
