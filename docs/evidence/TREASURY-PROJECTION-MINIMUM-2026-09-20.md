# Evidência — projeção D0–D+60 e menor saldo

Data: 2026-09-20  
Regra: `FIN-TES-002`, `FIN-TES-003`, seção 9 do handoff  
Escopo: núcleo determinístico, sem alegar sincronização ou persistência do Bling

## O que foi implementado

- `DailyProjection` produz 61 datas consecutivas de calendário, de D0 a D+60.
- D0 permanece `PROVISIONAL`; D+1 até D+60 permanecem `PROJECTED`.
- Cada saldo de abertura subsequente é o fechamento do dia anterior.
- `FindMinimumProjectedBalance` exige exatamente a série completa, valida continuidade, status e a equação `abertura + entradas - saídas + ajustes = fechamento`.
- O resultado conserva data, saldo de fechamento e composição de entradas, saídas e ajustes.
- Em empate, o primeiro menor saldo da série é escolhido de forma determinística.

## Evidência executada

```text
go test ./internal/financial
```

Casos cobertos:

1. menor saldo com data e composição preservadas;
2. série truncada rejeitada;
3. série não contínua rejeitada;
4. continuidade D0–D+60 já existente;
5. classificação temporal de D0 e D+1 já existente.

## Limitação explícita

Esta evidência não prova consulta ao Bling, reconciliação financeira, gravação em `daily_balances` ou drill-down de registros RAW. Esses gates continuam pendentes até existir payload/fixture sanitizada e fluxo de persistência homologado.

