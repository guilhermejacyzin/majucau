# Atomicidade e checkpoint da sincronização Bling — CI 2026-10-03

## Mudança

`apiResourceSyncService.Sync` cria um savepoint depois de inserir o lote e antes
de ler páginas. Em falha de leitura, persistência, cancelamento, limite de
paginação, cursor ou fechamento do lote, a transação volta ao savepoint: as
alterações de RAW e cursor da tentativa são revertidas, enquanto o lote é
finalizado como `FAILED` e gravado com código sanitizado. A gravação da auditoria
usa contexto independente com timeout de cinco segundos para continuar possível
quando o contexto da operação foi cancelado.

Em sucesso, RAW, cursor e status `SUCCESS` permanecem na mesma transação e são
confirmados juntos. O encerramento defensivo da transação também usa um contexto
limitado que não herda cancelamento.

## Evidência PostgreSQL

`TestAPIResourceSyncKeepsFailureAuditAndCommitsCursorWithRAW` usa páginas
determinísticas e um PostgreSQL descartável:

1. Prepara um lote anterior válido, um RAW corrente da página 1 e um cursor
   anterior.
2. A sincronização recebe uma nova página 1 e falha ao buscar a página 2.
3. Confirma que o lote `FAILED` e o código da falha foram persistidos, o RAW
   corrente anterior e o cursor não mudaram e nenhuma página parcial ficou
   gravada.
4. Repete a sincronização com duas páginas bem-sucedidas e confirma o cursor
   final, dois RAWs correntes e a preservação da versão RAW histórica.

A [CI 37110518416](https://github.com/guilhermejacyzin/majucau/actions/runs/37110518416)
passou nos jobs `verify-postgres` e `verify-windows`. O job PostgreSQL executou
essa integração; o job Windows passou nos testes Go, `go vet`, verificações de
segurança, build, smoke do instalador e consolidação dos outputs.

## Limites desta evidência

O teste controla as páginas retornadas pelo adapter e comprova atomicidade do
estado persistido. Ele não homologa um cursor incremental de retomada com o
Bling real, nem define a política de estado `stale`; esses pontos continuam
pendentes e mantêm INT-04 como `PARTIAL`. Não houve alteração de telas nem de
regras financeiras.
