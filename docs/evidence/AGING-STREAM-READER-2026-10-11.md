# Leitura incremental do aging — 11/10/2026

## Ponto identificado

A auditoria do requisito de streaming encontrou `CalculateAging` em `internal/financial/aging.go` recebendo uma lista completa `[]OpenReceivable`. Antes desta mudança, o código de produção não tinha consumidor de aging; as únicas chamadas eram os testes. Mesmo assim, o contrato disponível obrigava um futuro leitor de banco a montar a lista inteira antes de calcular.

## Mudança

- Foi acrescentado `OpenReceivableReader`, que entrega um título por chamada, e `CalculateAgingFrom`, que soma os resultados nas faixas fixas do aging sem guardar os títulos processados.
- O novo cálculo recebe `context.Context`, verifica cancelamento durante a leitura, propaga erros da fonte e devolve resultado vazio quando a leitura ou validação falha.
- `CalculateAging` permanece disponível como wrapper para a lista já existente. A integração futura deve usar `CalculateAgingFrom` com leitura de banco em fluxo.
- Foram acrescentados cenários para mil títulos fornecidos um de cada vez, falha da fonte e cancelamento após o primeiro título.

## Regras e limites

`AgingBucketFor`, o fuso e a data de negócio, as faixas D-005-A, a base no saldo aberto e a exclusão de títulos cancelados/estornados não foram alterados. Esta mudança não conecta aging a uma tela, consulta de produção ou fonte nova. BK-056 permanece parcial e depende de BK-046, BK-053 e BK-020.

## Verificação

`go test -count=1 ./internal/financial` passou em Windows. `GOCACHE`, `GOMODCACHE` e `GOTMPDIR` foram isolados em `artifacts/cache`, e o diretório foi removido ao final da execução. Não foi feito teste de carga nem medição de memória em produção.
