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

## Retificação de fonte e aprovação da sugestão de reserva — 2026-10-10

- Gisele esclareceu que recebimentos da Nuvemshop não são uma regra aprovada
  para a calculadora. A fórmula ativa deve usar o fluxo de caixa consolidado;
  não somar esses recebimentos separadamente por causa do campo `E16` da
  planilha.
- Gisele aprovou exibir somente como sugestão, sem alterar o máximo da D-003-A:

  ```text
  reserva_sugerida_hoje = R$ 50.000 + MAX(0,
    caixa_projetado_D0 - menor_saldo_diario_projetado_D0_D60)
  ```

- Exemplo: caixa projetado em D0 de R$ 100.000 e menor saldo diário previsto
  de R$ 70.000 produzem uma sugestão de R$ 80.000. A queda prevista é
  R$ 30.000; ao atravessá-la, a reserva mínima de R$ 50.000 permanece no ponto
  mais baixo projetado.
- A sugestão é informativa e não altera a reserva mínima aprovada, D-003-A,
  o valor máximo para aplicação ou qualquer transferência. Os campos do fluxo
  consolidado precisam ser mapeados antes da implementação; dados ausentes não
  podem ser substituídos por zero ou por recebimentos presumidos.

## Reauditoria da cópia mais recente da calculadora — 2026-10-11

A cópia enviada em 10/10 foi lida em modo somente leitura. Nenhum macro foi
executado e o arquivo não foi alterado. Gisele confirmou que a calculadora deve
manter a estrutura já enviada. Isso preserva sua estrutura de apresentação;
não aprova a soma separada de recebimentos da Nuvemshop, que já foi excluída da
regra da calculadora em 10/10.

A cópia ainda contém estes elementos antigos:

- A instrução em `B5` pede colar o resultado líquido da DRE do Bling. A decisão
  vigente determina que a DRE seja calculada pelo Majucau; o relatório do Bling
  é referência de comparação, não a fonte final da DRE.
- `E12` acumula o lucro mensal até M−2 e desconta o valor já investido. Esse
  corte coincide com o contrato D-003-A.
- `E20 = E15 + E16` soma o saldo da semana crítica com os recebimentos da
  Nuvemshop. Essa soma não deve ser copiada para o cálculo ativo. D-003-A usa
  diretamente o menor saldo diário de fechamento do fluxo consolidado em
  D0–D+60.
- `E21` limita o valor ao menor entre caixa depois da reserva e lucro
  disponível. O cálculo ativo continua sendo o contrato documentado em
  D-003-A, com reserva mínima aprovada de R$ 50.000; a reserva sugerida fica
  separada e é somente informativa.

A leitura estática do código confirmou que `MaximumApplication`,
`DailyProjection` e `FindMinimumProjectedBalance` existem como funções de
domínio. Para evitar acumular uma lista de movimentos sem limite, foi incluída
`DailyProjectionFrom`, que lê um movimento por chamada e mantém somente 61
totais diários. A versão antiga com fatia foi preservada como wrapper de
compatibilidade para conjuntos pequenos; leitores de banco devem usar a nova
entrada em fluxo. A busca ainda não encontrou chamadas de produção para essas
funções nem consultas de produção à tabela `daily_balances`; a migration define
a estrutura, mas o fluxo de caixa consolidado ainda não está conectado à
calculadora. Portanto, a regra está especificada, mas o cálculo ainda não
opera de ponta a ponta com dados reais. Se o saldo D0, os 61 fechamentos
diários, o lucro elegível ou o principal líquido já investido não estiverem
comprovados, o resultado dependente deve continuar indisponível, nunca ser
substituído por zero.

A alteração de `DailyProjectionFrom` recebeu revisão estática e formatação
Go. Não executei testes locais. Na CI `38111929797`, os testes Go e o job PostgreSQL passaram; o Windows falhou mais tarde porque o escopo não localizava `SafeCleanup` pelo nome curto. A correção de escopo foi confirmada na CI `38112462467`, com sucesso nos jobs Windows e PostgreSQL. Essa correção de empacotamento não altera a fórmula financeira nem comprova integração ponta a ponta. A calculadora ainda não chama esse helper; sua estrutura aprovada permanece a referência.

Esta reauditoria não alterou a planilha, a tela ou qualquer regra financeira.
O helper de projeção diária foi alterado para oferecer leitura em fluxo; a
calculadora ainda não chama esse helper. Não foram executados testes.

## Conferência da planilha enviada — 2026-10-11

Leitura somente para consulta do arquivo recebido, sem salvar alterações:
encontrei uma única aba (`Calculadora`) com as seções de DRE, fluxo de caixa e
resultado. A seção de fluxo tem campos separados para “Saldo na semana mais
crítica” e “Recebimentos previstos da Nuvemshop”; a fórmula atual soma esses
dois valores antes de descontar a reserva. A planilha não contém uma lista de
saldos por dia até 60 dias: o saldo da pior semana é uma entrada única.

A responsável confirmou que a estrutura da calculadora enviada deve
permanecer. Essa confirmação preserva a apresentação aprovada; ela não aprova
a soma separada de recebimentos da Nuvemshop. Prevalece a regra financeira já
registrada em 10/10: usar o menor saldo diário do fluxo consolidado em D0–D+60,
sem somar Nuvemshop novamente. A fonte e o cálculo pertencem ao fluxo
consolidado aprovado; a conexão dos dados reais à calculadora ainda está
pendente de implementação e comprovação. Nenhuma tela, planilha ou regra
financeira foi alterada nesta conferência.

**Situação:** Estrutura confirmada; integração técnica pendente.
