# Evidência — Valor Máximo para Aplicação (D-003-A)

## Regra aplicada

O produto usa a decisão aprovada D-003-A como oráculo vigente enquanto a
planilha histórica não for fornecida:

```text
lucro_disponivel = max(0, lucro_elegivel - valor_ja_investido)
limite_financeiro = max(0, menor_saldo_diario_D0_D60 - reserva_minima)
valor_maximo = max(0, min(lucro_disponivel, limite_financeiro))
```

O motor recebe exatamente 61 saldos consecutivos, D0 até D+60, usa dinheiro
decimal exato e não transforma ausência de recebimento oficial em estimativa.

## Casos dourados executados

- limite financeiro menor que o lucro disponível: o resultado fica limitado
  pelo pior saldo menos a reserva;
- lucro já investido maior que o lucro elegível: lucro disponível é zero;
- reserva mínima acima do pior saldo: limite financeiro e valor máximo são
  zero;
- horizonte inválido ou datas não consecutivas: a entrada é rejeitada;
- parâmetros negativos e regra sem versão: a entrada é rejeitada.

## Verificação

Comando executado:

```text
go test ./internal/financial
go test . ./cmd/... ./database/... ./internal/...
```

Resultado: aprovado. A equivalência com a planilha original permanece fora de
escopo até que o arquivo seja entregue; a decisão D-003-A é o oráculo vigente
registrado no G0.
