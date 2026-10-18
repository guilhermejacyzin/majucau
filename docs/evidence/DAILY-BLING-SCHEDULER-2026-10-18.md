# Agendamento local diário do Bling — 18/10/2026

## Implementado

- `internal/scheduler/daily.go` calcula o próximo horário diário em `America/Sao_Paulo` e permite cancelar a espera pelo contexto do worker.
- O worker Windows inicia esse agendamento junto ao serviço. Às 09:30 chama a sincronização de Contas a Receber e Contas a Pagar sem filtros de data, para reler todo o conteúdo suportado pelo contrato do Bling.
- A execução usa o retry GET limitado existente. Cada recurso conserva sua transação; a falha de um recurso não confirma dados parciais daquele recurso. O processo não informa sucesso quando o serviço retorna estado incompleto.
- Quando a resposta não comprova o fim da paginação, o worker mantém o diagnóstico `BLING_API_PAGINATION_UNKNOWN`; não muda para sucesso nem avança o cursor.
- Se o worker iniciar depois das 09:30, o próximo disparo será no dia seguinte. O job ainda não coleta arquivos do Drive; o leitor Drive não está ligado ao worker.

## Verificações

- `go test ./internal/scheduler ./cmd/worker -count=1` — passou no Windows.
- `go test ./... -count=1` — passou no Windows.
- `go vet ./...` — passou no Windows.
- `git diff --check` — passou.
- Os testes exercitam a fronteira de horário, o fuso, a execução seguinte após uma falha, a chamada sem janela de datas e a redação do código de erro.

## Limitações e bloqueios restantes

- Nenhuma chamada real ao Bling foi executada. A paginação permanece bloqueada até provar o marcador de fim com a resposta real; nesse caso, o job falha sem confirmar o lote.
- O agendamento ainda não persiste a tentativa do job, não impede duas instâncias em computadores diferentes e não retoma páginas confirmadas depois de uma reinicialização.
- O Windows precisa manter o worker em execução para disparar no horário. Se ele iniciar após 09:30, não há execução compensatória no mesmo dia.
- A coleta Drive ainda requer identidade técnica, IDs das pastas, credencial protegida e integração com importação. Quando ligada, deve comparar a data do arquivo com o instante programado para deixar arquivo que chegou depois das 09:30 para o próximo dia.
- Não foram alteradas telas, regras financeiras, schemas ou dados reais. BK-039 continua parcial.
