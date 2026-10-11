# Evidência — cancelamento na projeção diária em fluxo

**Data:** 11/10/2026
**Escopo:** helper financeiro `DailyProjectionFrom`; nenhuma consulta de produção foi conectada.

## Mudança

- `MovementReader` recebe `context.Context`, permitindo ao leitor de banco interromper sua leitura quando a chamada for cancelada.
- `DailyProjectionFrom` verifica o cancelamento antes de cada registro. Falha ou cancelamento retorna erro sem expor uma projeção parcial.
- `DailyProjection` continua como wrapper de compatibilidade para listas já disponíveis.
- A soma de entradas, saídas e ajustes, a janela D0–D+60, os 61 saldos, os tipos D0/D+ e o uso de moeda decimal exata foram preservados.

## Limites

O helper ainda não tem consumidor de produção. Não liga a projeção a dados do Bling, à tabela `daily_balances`, ao fluxo de caixa ou à calculadora. A fórmula e as regras de negócio não foram alteradas; a estrutura enviada da calculadora permanece a referência. BK-006 permanece em 30% e BK-060 continua não iniciada até haver fonte e integração comprovadas.

## Verificação

`go test -count=1 ./internal/financial` e `go vet ./internal/financial` passaram em Windows. O teste novo confirma que o contexto chega ao leitor e que o cancelamento após o primeiro registro retorna erro, sem série parcial. Os caches foram direcionados a `artifacts/cache` e removidos após a verificação. Não foi feito teste de carga nem medição em produção.
