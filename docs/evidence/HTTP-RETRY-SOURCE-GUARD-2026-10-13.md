# Verificação automática das chamadas HTTP de produção — 13/10/2026

## O que foi conferido

O retry seguro já existia e era usado nas leituras do Bling. Esta mudança acrescenta uma barreira automática para que uma futura integração não faça uma chamada HTTP direta sem revisão.

`internal/integrations/httpretry/source_policy_test.go` lê os arquivos Go de produção e verifica:

- chamadas diretas `.Do` permanecem nos caminhos inventariados;
- `net/http.Get`, `Head`, `Post` e `PostForm` diretos não sejam introduzidos;
- os caminhos esperados continuem presentes: consultas do Bling protegidas pelo retry, troca OAuth de uso único sem repetição e transporte do limitador de frequência;
- a chamada `sync.Once.Do` usada apenas para fechar o Named Pipe continue identificada como uma operação local, sem relação com HTTP.

Uma nova chamada precisa entrar na revisão e atualizar explicitamente o inventário. Operações repetíveis devem continuar usando `internal/integrations/httpretry`; operações OAuth de uso único permanecem sem retry automático.

## Verificação

- `go test -count=1 ./internal/integrations/httpretry` passou em 13/10/2026.
- O Go usou os caches já centralizados em `artifacts/cache/`; nenhum segredo, dado real do Bling ou arquivo externo foi lido.
- O teste não acessa a rede nem altera dados, telas, regras financeiras, migrations ou configurações do protótipo.

## Limites

Esta barreira detecta mudanças no código Go de produção que não estejam no inventário revisado. Ela não prova autenticação real, completude de paginação, sincronização com a conta Majucau ou homologação financeira. Esses passos continuam pendentes da configuração do aplicativo Bling e da conferência dos dados reais.
