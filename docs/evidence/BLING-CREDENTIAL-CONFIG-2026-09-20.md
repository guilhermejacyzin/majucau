# Evidência — configuração segura de credenciais Bling — 2026-09-20

## Escopo

Este marco fecha a fronteira de configuração editável solicitada para a API do Bling. Ele não reabre as regras de origem dos dados nem altera o cliente de leitura já testado. O fluxo entregue é:

`React → binding Wails → Named Pipe autenticado → worker Go → DPAPI → PostgreSQL`

## Comportamento entregue

- A tela **Configurações > Integrações > Bling** aceita `Client ID`, `Redirect URI` e `Client Secret`.
- O segredo atravessa somente a requisição transitória até o worker; não é mantido no estado React, não é devolvido na resposta IPC, não é gravado em log e não é salvo no PostgreSQL.
- O `Client Secret` é protegido pelo Windows DPAPI em `%ProgramData%\Majucau\Secrets` (ou pelo diretório definido em `MAJUCAU_SECRET_DIR`).
- O PostgreSQL guarda apenas os metadados públicos e a referência estável do segredo (`bling_credentials_v1`), com status `NOT_CONFIGURED`/`AUTH_ERROR` conforme o caso.
- Falha na persistência pública tenta restaurar o segredo anterior, evitando estado parcialmente aplicado.
- A UI informa sucesso/erro de forma sanitizada. Conectar, testar, reconectar, desconectar e sincronizar continuam ações distintas e não são liberadas antes da homologação completa do OAuth.

## Evidência automatizada

- `go test ./...`: aprovado.
- `go vet ./...`: aprovado.
- `npm run lint`: aprovado.
- `npm run typecheck`: aprovado.
- `npm run test:run`: 17 testes aprovados.
- `npm run build`: aprovado.
- `wails build -clean -trimpath`: aprovado; bindings regenerados e `build/bin/majucau.exe` produzido.
- `git diff --check`: aprovado.

## Limite conhecido

Este marco ainda não afirma que a conta Bling foi autorizada. O próximo marco implementa o início do fluxo OAuth no navegador externo, callback local/retorno seguro, teste de credencial e estado operacional no worker. Nenhum dado financeiro é sincronizado pelo salvamento de credenciais.
