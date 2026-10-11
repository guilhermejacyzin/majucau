# Evidência — cancelamento entre linhas na importação CSV

**Data:** 11/10/2026
**Escopo:** parsers de recebimentos Bling e agenda futura Nuvem Pago.

## Mudança

- `StreamReceiptsCSV` e `StreamFutureCSV` recebem o mesmo contexto da importação.
- O contexto é verificado antes e depois de cada leitura limitada de registro. Uma operação cancelada para mesmo quando linhas inválidas não têm um callback de aviso.
- Os chamadores de prévia e importação passam o contexto existente para os parsers. O erro é propagado ao serviço; o caminho de importação retorna antes do commit e mantém o rollback já existente.
- Limites de tamanho/quantidade de campos, separadores, validação de linha, contagens, gravação parcial de linhas válidas, transação, deduplicação e telas não foram alterados.

## Verificação

Testes focais nos pacotes Bling e Nuvem Pago confirmam que cancelar após uma linha rejeitada retorna `context.Canceled` e impede o aceite das linhas seguintes. `go vet` foi executado para os dois pacotes. A execução não homologou importação contra uma conta real ou banco de produção.

O cache de compilação, os módulos e os temporários do Go ficaram em `artifacts/cache` e foram removidos depois das verificações.
