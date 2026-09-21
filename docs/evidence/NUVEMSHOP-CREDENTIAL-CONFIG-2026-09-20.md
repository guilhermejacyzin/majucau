# Evidência — configuração segura de credenciais Nuvemshop — 2026-09-20

## Escopo

Este marco fecha o formulário editável de credenciais da Nuvemshop sem afirmar que o OAuth ou a sincronização financeira estão homologados. O fluxo entregue é:

`React → binding Wails → Named Pipe autenticado → worker Go → DPAPI → PostgreSQL`

## Comportamento entregue

- A tela **Configurações > Integrações > Nuvemshop** aceita `App ID`, `Redirect URI HTTPS` e `Client Secret`.
- O segredo atravessa somente a requisição transitória até o worker; não é mantido no estado React, não é devolvido na resposta IPC, não é gravado em log e não é salvo no PostgreSQL.
- O `Client Secret` é protegido pelo Windows DPAPI no mesmo cofre local usado pelas integrações.
- O PostgreSQL guarda somente `App ID`, `Redirect URI`, a referência estável do segredo (`nuvemshop_credentials_v1`) e o estado de configuração.
- O Redirect URI é aceito apenas quando é HTTPS e não é loopback. Isso mantém o formulário pronto para o relay homologado D-002-A sem inventar um callback local inseguro.
- O botão de salvar não inicia OAuth nem sincronização. Essas ações permanecem bloqueadas até que o relay HTTPS, PKCE, escopos, callback e payloads reais sejam homologados.

## Evidência automatizada

- `go test ./internal/integrations/nuvemshop ./cmd/worker ./internal/ipc ./internal/application`: aprovado.
- `go vet ./internal/integrations/nuvemshop ./cmd/worker ./internal/ipc ./internal/application`: aprovado.
- `npm run lint`: aprovado.
- `npm run typecheck`: aprovado.
- `npm run test:run`: 19 testes aprovados.
- `npm run build`: aprovado.
- Testes de contrato confirmam que o Client Secret chega ao worker e não aparece na resposta pública.

## Limite conhecido

O gate INT-02/G2 continua bloqueado: faltam relay HTTPS publicado, credenciais reais fora do repositório, homologação de OAuth/PKCE, Store ID, escopos e fixtures reais. Nenhum pedido, parcela, cancelamento, reembolso, taxa ou recebimento é derivado do simples salvamento desta configuração.
