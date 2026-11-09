# Evidência — renovação OAuth em detalhes da T02

**Data:** 09/11/2026 (horário de Brasília)  
**Tarefa:** T02 — Fluxo de Caixa  
**Estado da implementação:** correção local implementada; reconstrução e conferência com a API real pendentes.

## Falha observada

A sincronização real de Contas a Pagar, correlação `00466a54-6143-47fa-97ae-488b464eabc6`, terminou às 03:44:28 como `FAILED` e `RAW_ONLY`, com código sanitizado `BLING_API_UNAUTHORIZED`. Foram lidos 10.301 registros e preparados 67 de 104 lotes de detalhe. Como não houve publicação atômica, os dados dessa execução não substituíram o retrato anterior.

O registro local da conexão indicava expiração do token de acesso às 03:44:28; o token de renovação ainda tinha validade registrada até 08/12/2026. A causa no código foi confirmada por inspeção do adapter: `listFinancialResource` atualizava o token ao receber HTTP 401, mas `getDataRecord` apenas retornava o erro. Os pedidos de detalhe de contas a pagar, contatos e categorias usavam `getDataRecord`.

Depois de reconstruir o worker com essa correção, a leitura manual do período exibido na T02 (09/11/2026 a 08/01/2027, por vencimento) foi cancelada após 10 segundos: 500 recebíveis lidos, 0 obrigações lidas. O aplicativo permitia 2 minutos, mas o servidor IPC aplicava o limite padrão de 10 segundos ao método `bling.sync`. A carga interrompida não publicou os dados novos.

## Correção

`getDataRecord` agora aciona o mesmo callback OAuth protegido pelo worker após HTTP 401, altera o cabeçalho Authorization e repete o GET uma única vez. Renovação sem token novo ou com erro continua falhando com segurança; não há repetição da chamada com credencial rejeitada. Nenhuma regra financeira, fonte de dados ou tela foi alterada.

O limite de leitura do método filtrado `bling.sync` agora compartilha o prazo de até duas horas do desktop e do worker, igual ao refresh completo. Isso evita que o worker cancele uma consulta paginada antes do prazo que o aplicativo já permite.

## Verificação executada

- `TestGetPayableRefreshesExpiredAccessTokenOnce`: confirma um 401, uma renovação e uma segunda leitura GET com token novo.
- `TestGetPayableDoesNotRetryWhenTokenRefreshFails`: confirma que falha de renovação não provoca uma segunda leitura.
- `go test ./internal/integrations/bling`: passou em 09/11/2026.
- `scripts/verify.ps1`: passou em 09/11/2026, incluindo todos os pacotes Go, `go vet`, govulncheck, migrations, scanner local de segredos, validação do instalador, lint, tipagem, 75 testes do frontend, build do frontend e `npm audit` sem vulnerabilidades.
- `scripts/verify.ps1`: passou após as correções OAuth e IPC, incluindo todos os testes Go, `go vet`, govulncheck, migrations, sqlc, scanner de segredos, validação do instalador, lint, tipagem, 75 testes/build do frontend e `npm audit`.

## Falta para concluir a correção

1. Reconstruir/reabrir o worker/protótipo usando este código.
2. Executar uma nova atualização oficial do Bling e confirmar que Contas a Receber e Contas a Pagar concluem sem expiração/cancelamento.
3. Confirmar publicação normalizada dos dois conjuntos sem deixar dados parciais como atuais.
4. Conferir a série semanal da T02 com Contas a Receber pelo valor líquido e Contas a Pagar, mantendo o saldo inicial aprovado do Itaú.
5. Abrir T02 atualizada para aprovação visual e numérica da responsável.

**Conclusão:** a falha de programação foi corrigida e testada no adapter. Ainda não há comprovação de recuperação da carga real nem correção dos valores na tela; T02 permanece aberta e T03 não deve começar até a aprovação da T02.
