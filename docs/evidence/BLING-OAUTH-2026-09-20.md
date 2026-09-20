# Evidência — OAuth Bling no desktop — 2026-09-20

## Escopo fechado neste marco

O fluxo de Authorization Code do Bling passou a existir ponta a ponta no recorte seguro do desktop:

`UI → Wails → Named Pipe → worker → navegador externo → callback 127.0.0.1 → troca de código → teste GET mínimo → DPAPI`

O fluxo não inicia sincronização nem grava fatos financeiros. O teste consulta uma página com `pagina=1` e `limite=1` e grava apenas o estado operacional sanitizado.

## Regras de segurança aplicadas

- O navegador é externo; o login não é capturado no WebView.
- Cada tentativa cria `state` aleatório de 256 bits e uma sessão com TTL de cinco minutos.
- O listener aceita somente `http://127.0.0.1:porta/caminho` previamente configurado e registrado no aplicativo Bling.
- `localhost`, host público, HTTPS não-loopback, porta ausente, query e fragmento são rejeitados no início.
- Callback com `state` incorreto recebe erro e não encerra a sessão válida; replay não troca o código.
- Client Secret, authorization code, access token e refresh token não atravessam a resposta IPC, não aparecem na URL e não entram em logs.
- Tokens e Client Secret ficam no mesmo bundle DPAPI do worker. O PostgreSQL recebe somente metadados, escopos e expirações.
- Falha de teste, persistência ou troca de token muda o estado operacional para erro sem apagar o último dado financeiro válido.

## Evidência automatizada

- Teste de URL de autorização sem segredo.
- Teste de rejeição de Redirect URI inseguro.
- Teste local de callback, `state`, troca de código, teste GET e armazenamento de token em cofre fake.
- Teste de contratos IPC para iniciar, consultar estado e testar conexão.
- Teste frontend de iniciar OAuth e acompanhar somente a mensagem sanitizada.
- `go test ./...`: aprovado.
- `go vet ./...`: aprovado após o frontend embutido ser gerado.
- `npm run lint`, `npm run typecheck`, `npm run test:run` (18 testes) e `npm run build`: aprovados.
- `wails build -clean -trimpath`: aprovado; bindings regenerados e executável Windows produzido.

## Limites do gate

O gate G2 ainda não está concluído: falta executar com credencial real, confirmar o Redirect URI no cadastro real do aplicativo Bling, validar payloads reais sanitizados e provar paginação, rate limit, refresh após reinício e sincronização idempotente. Nuvemshop permanece condicionada ao relay HTTPS D-002-A; Nuvem Pago continua `UNAVAILABLE` para confirmação financeira conforme D-004-A.
