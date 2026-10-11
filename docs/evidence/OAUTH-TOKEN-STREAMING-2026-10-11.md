# Resposta OAuth em fluxo — 2026-10-11

## Mudança

`BlingOAuthClient` deixou de copiar o corpo HTTP da resposta OAuth inteira para
um `[]byte` de até 4 MiB antes da decodificação. Agora `json.Decoder` lê o corpo
diretamente por um `io.LimitedReader` e converte apenas os campos tipados que o
serviço precisa usar. O teto de 4 MiB foi preservado.

O decoder aceita exatamente um objeto JSON. Conteúdo extra, corpo acima do
limite e erro de leitura continuam falhando de modo fechado, com os mesmos
códigos públicos: formato inválido é `ErrBlingOAuthSchema`; excesso ou falha de
leitura é `ErrBlingAPIInvalidResponse`. Nenhum corpo, token ou detalhe sensível
é incluído na mensagem de erro.

## Compatibilidade

- Telas, contratos Wails e regras financeiras não foram alterados.
- OAuth continua sem retry automático, por ser uma troca de credencial de uso
  único sem garantia de repetição segura.
- Os campos de token permanecem em memória porque são necessários ao fluxo; o
  corpo JSON bruto não é mantido em uma segunda cópia.

## Verificação

- `go test ./internal/integrations/bling` — passou localmente em 2026-10-11.
- `go vet ./internal/integrations/bling` — passou localmente em 2026-10-11.
- Não houve chamada à conta real do Bling; o navegador continua bloqueando a
  página por uma preferência salva.
- Isso não substitui medição de pico de memória nem homologação real.
