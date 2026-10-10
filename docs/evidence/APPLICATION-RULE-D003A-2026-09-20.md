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

## Leitura estrutural detalhada do arquivo — 2026-10-04

A leitura foi somente leitura: o arquivo não foi alterado e nenhum macro foi
executado. Embora tenha extensão `.xlsm`, o pacote não contém projeto VBA nem
links externos. Há uma aba visível, `Calculadora`, com uma única área de
trabalho (`A1:N27`). Valores de entrada foram inspecionados apenas para
identificar suas células e formatos; nenhum valor financeiro foi copiado para
este documento.

Fórmulas encontradas:

- `E12`: acumula o lucro líquido da linha mensal de janeiro até dois meses
  antes do mês de referência; janeiro/fevereiro resultam em zero. Depois
  subtrai o valor informado como “já investido”. Isso corresponde ao corte
  temporal escrito na própria planilha e à parte de lucro aprovada em D-003-A.
- `E20 = E15 + E16`: soma “Saldo na semana mais crítica” e “Recebimentos
  previstos da Nuvemshop”. Continua pendente confirmar se `E15` já inclui esses
  recebimentos; somá-los novamente poderia duplicá-los.
- `E21 = MAX(0, MIN(E20 - E17, E12))`: limita a aplicação ao menor entre o
  caixa projetado depois da reserva mínima (`E17`) e o lucro disponível (`E12`),
  sem resultado negativo.
- `E22 = E20 - E21`: mostra o caixa projetado restante após a aplicação.

Não existe campo com o nome “seguro”. O campo correspondente mais próximo é
“Reserva mínima de segurança” (`E17`). Foi solicitada confirmação simples para
saber se “seguro” queria dizer essa reserva de caixa ou uma proteção/seguro da
aplicação. Até a resposta, não alterar nem renomear essa regra.

Nota de qualidade da referência: as fórmulas das células `B8:M8` são iguais.
Quando o mês selecionado é março ou posterior, todas elas mostram “somado”,
embora `E12` some somente janeiro até o mês de referência menos dois. Isso
parece uma marcação visual imprecisa e não muda `E12`/`E21`. Não reproduzir essa
marcação na tela do Majucau sem confirmação.

## Resposta sobre recebimentos previstos — 2026-10-10

- Foi perguntado se o saldo da semana mais crítica (`E15`) já inclui os
  recebimentos previstos da Nuvemshop (`E16`). A responsável respondeu:
  “Não sei; mantenha a regra atual por enquanto”.
- D-003-A continua sendo a fórmula ativa no produto. Não substituir o cálculo
  pelo fluxo `E15 + E16` da planilha enquanto essa relação não for conhecida e
  aprovada; somar um recebimento já incluído no saldo crítico poderia contá-lo
  duas vezes.
- A célula `E17` da referência está rotulada “Reserva mínima de segurança”. A
  planilha de origem permanece sem alteração; nenhum cálculo, tela ou regra do
  produto foi mudado por esta resposta.
