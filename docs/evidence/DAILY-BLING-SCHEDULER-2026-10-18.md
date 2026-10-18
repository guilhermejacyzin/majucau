# Agendamento local diário do Bling — 18/10/2026

Esta evidência inicial foi complementada por `DAILY-BLING-RECOVERY-2026-10-18.md`, que registra persistência da identidade diária, proteção contra concorrência e retomada após reinício. Use o documento complementar para o estado mais recente.

## Implementado inicialmente

- `internal/scheduler/daily.go` agenda às 09:30 em `America/Sao_Paulo` e permite cancelar a espera pelo contexto do worker.
- O worker Windows relê Contas a Receber e Contas a Pagar do Bling sem filtros de data.
- A execução usa o retry GET limitado existente. Cada recurso conserva sua transação; a falha de um recurso não confirma dados parciais daquele recurso.
- Quando a resposta não comprova o fim da paginação, a sincronização falha com `BLING_API_PAGINATION_UNKNOWN` e não confirma o cursor.

## Verificações iniciais

- `go test ./internal/scheduler ./cmd/worker -count=1`, `go test ./... -count=1`, `go vet ./...` e `git diff --check` passaram em Windows durante a implementação inicial.
- Os testes exercitam horário, fuso, tratamento de falha, sincronização sem janela de datas e redação do código de erro.

## Limitações que continuam válidas

- Nenhuma chamada real ao Bling foi executada. A paginação permanece bloqueada até comprovar como a resposta real informa o fim da lista.
- Ainda não há checkpoint por página; após uma interrupção, a rotina relê a carga integral.
- O worker precisa estar em execução no horário ou reiniciar depois dele para recuperar a carga do dia. O job ainda não coleta arquivos do Drive.
- A coleta Drive requer identidade técnica somente leitura, IDs das pastas, credencial protegida e integração com importação. Quando ligada, deve preservar a regra de deixar arquivo recebido depois das 09:30 para a atualização do dia seguinte.
- Não foram alteradas telas, regras financeiras ou dados reais. BK-039 continua parcial.
