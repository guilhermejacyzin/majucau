# Evidência — normalização de nomes de campos sensíveis

**Data:** 11/10/2026
**Escopo:** helper `internal/security.RedactJSON`.

## Mudança

- Antes, a comparação de chaves reconhecia somente grafias com o mesmo separador, como `access_token`. Variações comuns de API, como `accessToken`, `clientSecret` e `API-Key`, podiam passar sem mascaramento.
- Agora, a comparação ignora maiúsculas/minúsculas e caracteres separadores. A lista cobre campos comuns de token, segredo, senha, credencial, autorização, cookie e chaves privadas/de API.
- O limite de entrada de 1 MiB e a forma de mascaramento das strings foram preservados. Campos não sensíveis continuam visíveis.
- O teste inclui nomes em camelCase, hífen, diferenças de caixa e campos aninhados.

## Verificação e limite

`go test -count=1 ./internal/security` e `go vet ./internal/security` passaram localmente.

`RedactJSON` não tem chamador de produção no momento. O logger do worker evita esse risco usando campos permitidos individualmente e não aceita payload livre. Portanto, esta mudança melhora o helper e sua cobertura de teste, mas não prova que os logs de produção aplicam mascaramento genérico. BK-031 permanece em 80%.

Nenhuma tela, regra financeira ou dado persistido foi alterado. O cache Go foi colocado em `artifacts/cache` e removido após a verificação.
