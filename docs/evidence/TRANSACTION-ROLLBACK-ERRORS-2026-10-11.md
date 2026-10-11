# Evidência — propagação de falhas ao desfazer transações

**Data:** 11/10/2026
**Escopo:** importações CSV do Bling e da Nuvem Pago, sincronização de contas a pagar/receber pela API do Bling e leitura do snapshot do painel.

## Mudança

- Foi criado `internal/dbtx.Rollback`, auxiliar comum que tenta desfazer a transação com um prazo máximo de cinco segundos.
- O contexto usado na operação pode estar cancelado. O rollback usa um contexto separado para tentar limpar a transação mesmo assim, preservando valores de rastreamento, mas sem herdar o cancelamento.
- `pgx.ErrTxClosed` é considerado um encerramento normal, como após `Commit`. Outras falhas de rollback são devolvidas com contexto.
- Os quatro caminhos agora juntam uma falha de limpeza ao erro original. Quando o rollback falha, importações e sincronização não retornam contadores como se o lote tivesse terminado normalmente.
- A transação, o momento do `Commit`, os dados e as regras de contagem existentes foram preservados. Não houve alteração de tela ou regra financeira.

## Verificação

`go test -count=1 ./internal/dbtx ./internal/dashboard ./internal/integrations/bling ./internal/integrations/nuvempago` e `go vet` nos mesmos pacotes passaram localmente.

Os testes unitários cobrem contexto cancelado, prazo máximo, preservação de valores de rastreamento, transação já encerrada e falha de rollback. Não foi executado teste contra PostgreSQL real nesta verificação. A CI do commit publicado ainda precisa confirmar a mudança no fluxo completo.

O cache de compilação, os módulos e os temporários do Go foram mantidos em `artifacts/cache` durante a verificação e removidos ao final.
