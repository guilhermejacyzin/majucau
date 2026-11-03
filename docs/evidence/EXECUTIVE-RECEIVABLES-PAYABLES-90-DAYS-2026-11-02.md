# T01: cards de vencimento em 90 dias — 02/11/2026

## Decisão aprovada

- Os cards existentes **A RECEBER** e **A PAGAR** mostram os títulos com vencimento nos próximos 90 dias.
- A janela inclui hoje e termina antes de hoje + 90 dias.
- Contas a Pagar continua representada como saída negativa.
- **MENOR SALDO 30 / 60 DIAS** não muda.
- Não foram criados cards, faixas, rótulos ou posições novas.

## Implementação

O contrato do snapshot inclui `receivables_due_next_90_days` e `payables_due_next_90_days`. A consulta calcula soma e contagem dos títulos com vencimento dentro da janela aprovada; Go encaminha as métricas ao frontend, e a T01 liga os dois cards existentes a elas. As telas T03 e T04 mantêm seus totais gerais.

## Verificação e limites

- `go test ./internal/dashboard`: passou.
- `npm run build` na cópia isolada `artifacts/verification/frontend-live`: passou.
- A captura anterior da prévia 4179 mostrava o total geral de Contas a Pagar, sem limite de período; não é evidência do total dos próximos 90 dias.
- Contas a Receber continua sem valor confirmado enquanto os registros reais do Bling não forem normalizados na tabela consultada pelo painel.
- O build local não prova sincronização real ou reconciliação com o Bling.
- O percentual do projeto permanece em 12,1% concluído / 87,9% restante.

## Checagem do worker ativo — 02/11/2026

Após reiniciar o worker atualizado, o snapshot consultado pela named pipe local confirmou:

- Contas a Pagar nos próximos 90 dias: 418 títulos, estado `PARTIAL`.
- Contas a Receber nos próximos 90 dias: zero linhas normalizadas consultáveis, estado `UNAVAILABLE`.
- O diagnóstico imprimiu apenas estados e contagens; nenhum valor financeiro ou dado de cliente foi gravado neste registro.

O estado de Contas a Receber significa “sem resultado normalizado disponível”, não total financeiro igual a zero. A verificação não substitui a comparação com a tela do Bling nem a homologação. O app Wails ativo está em `http://localhost:34115/`; `127.0.0.1:4179` é prévia estática.
