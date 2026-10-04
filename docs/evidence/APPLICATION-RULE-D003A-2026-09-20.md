# Evidência — Valor Máximo para Aplicação (D-003-A)

## Regra aplicada

O produto usa a decisão aprovada D-003-A como regra vigente. A planilha de
referência foi recebida em 2026-10-04; sua comparação com D-003-A está em
andamento. Até a conferência e aprovação de qualquer mudança, D-003-A continua
sendo a regra ativa:

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

Resultado: aprovado para D-003-A. A equivalência com a planilha recebida em
2026-10-04 ainda não foi confirmada; ver a revisão inicial abaixo.

## Revisão inicial da planilha recebida — 2026-10-04

A planilha tem uma única página de cálculo. Ela acumula o lucro até dois meses
antes do mês selecionado e desconta o valor já investido. Na parte de caixa,
usa três entradas informadas pela pessoa: saldo da semana mais crítica,
recebimentos previstos da Nuvemshop e reserva mínima de segurança. O cálculo
soma o saldo crítico aos recebimentos previstos, desconta a reserva e limita o
resultado ao lucro disponível, sem permitir valor negativo.

O limite pelo lucro segue a forma aprovada em D-003-A. A reserva também continua
sendo uma entrada manual na referência. Ainda falta confirmar se o saldo da
semana crítica já inclui os recebimentos da Nuvemshop ou se eles devem ser
somados separadamente, como a fórmula da planilha faz. O motor atual encontra o
menor saldo diário em D0–D+60; não se declarou equivalência entre esse valor e a
entrada manual da planilha.

Nenhum valor financeiro nem dado de fornecedor da planilha foi copiado para o
repositório. Nenhuma regra, tela ou fórmula de produção foi alterada nesta
revisão.
