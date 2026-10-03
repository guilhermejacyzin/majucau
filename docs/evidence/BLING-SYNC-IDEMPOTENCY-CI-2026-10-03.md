# Replay idempotente da sincronização Bling — CI 2026-10-03

## Evidência PostgreSQL

O teste `TestAPIResourceSyncKeepsFailureAuditAndCommitsCursorWithRAW` agora
comprova três resultados no PostgreSQL descartável:

1. Quando a segunda página falha, o lote `FAILED` permanece auditável, a versão
   RAW válida anterior continua corrente, o cursor não avança e nenhuma página
   parcial sobrevive.
2. Quando duas páginas são concluídas, os RAWs, o cursor `page:2` e o lote
   `SUCCESS` são confirmados juntos; a versão RAW substituída permanece no
   histórico.
3. Ao repetir as mesmas duas páginas, a sincronização termina `SUCCESS`, não
   conta página como criada/atualizada e mantém inalterados os totais de RAW
   corrente e histórico.

A [CI 37111813159](https://github.com/guilhermejacyzin/majucau/actions/runs/37111813159)
passou nos jobs `verify-postgres` e `verify-windows`.

## Limites

As páginas são determinísticas e injetadas no adapter de teste. A evidência
comprova idempotência do snapshot RAW por hash no PostgreSQL; não valida payloads
de conta real nem homologa cursor incremental oficial do Bling. DATA-02 continua
`PARTIAL` até cobrir os demais adapters, relatórios de lote e replays contra as
fontes contratadas.
