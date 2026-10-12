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

## Conferência da cópia disponível — 2026-10-11

Leitura somente para consulta; nenhum valor de entrada foi copiado para esta
evidência e a planilha não foi salva ou alterada. O arquivo tem uma aba visível,
uma lista de meses em `E10`, não contém projeto VBA nem links externos e tem
fórmulas apenas nos marcadores mensais e nos três resultados.

### O que a planilha atual faz

- A instrução de DRE pede valores da DRE do Bling. A decisão aprovada do produto
  determina que o Majucau calcule sua própria DRE; o relatório pronto do Bling
  serve para comparação, não como resultado final.
- Os resultados mensais de janeiro a dezembro são campos de entrada. A fórmula
  `E12` soma os meses de janeiro até dois meses antes do mês selecionado; em
  janeiro e fevereiro, retorna zero. As marcações visuais `B8:M8` repetem a
  mesma indicação e não mostram corretamente o corte M−2, mas isso não altera
  a soma calculada em `E12`.
- “Valor já investido” (`E11`) é digitado manualmente e é subtraído em `E12`.
  Isso não comprova que o valor corresponde ao principal líquido real. A regra
  aprovada D-003-A exige calcular o principal já investido por aportes menos
  resgates, sem misturar rendimentos.
- O fluxo de caixa recebe manualmente o saldo da semana crítica, os recebimentos
  previstos da Nuvemshop e a reserva mínima. A planilha soma o saldo e o valor da
  Nuvemshop (`E20 = E15 + E16`), limita a aplicação pela reserva e pelo lucro
  elegível calculado em `E12` e mostra o saldo que sobra. Não há uma série dos
  61 saldos diários D0–D+60 nessa planilha.

### O que continua valendo no Majucau

D-003-A permanece a regra ativa. O cálculo usa o lucro líquido da DRE própria do
Majucau até M−2, subtrai o principal líquido já investido e compara esse valor
com o menor saldo diário de fechamento do fluxo consolidado em D0–D+60, menos a
reserva mínima. Não soma recebimentos da Nuvemshop separadamente. A sugestão de
reserva para hoje continua separada e somente para consulta, conforme D-003-C.

Assim, a planilha atual ajuda a entender a estrutura, mas seus campos manuais e
sua soma de Nuvemshop não substituem as regras aprovadas. A tela aprovada mostra
“Caixa consolidado”; não alterei essa tela, a planilha ou qualquer regra
financeira. O vínculo da calculadora com as fontes reais continua pendente.

A busca estática atual confirmou que as funções de domínio para projetar os
61 dias e calcular a aplicação existem, mas não encontrou chamadas de produção
nem consulta aos saldos diários. A entrada em fluxo mantém poucos saldos em
memória; falta conectá-la ao fluxo consolidado aprovado e demonstrar as entradas
reais. Esta conferência foi documental e não executou testes.

**Situação:** Estrutura documentada; comparação das fórmulas concluída; vínculo
com dados reais e exemplos reproduzíveis pendentes. Esta seção descreve a cópia
disponível em 11/10; a versão binária da conferência de 04/10 não está disponível
neste checkout para comparação de arquivo.

## Testes do processador diário em fluxo — 2026-10-11

Foram acrescentados testes unitários para `DailyProjectionFrom`. O leitor de
teste gera 1.000 movimentos sob demanda, sem montar uma lista de entrada; a
projeção resultante mantém os 61 dias esperados e os totais de D0 e D+60. Outros
casos confirmam que erro do leitor e estouro de moeda interrompem o cálculo sem
retornar uma projeção parcial, e que um leitor ausente é recusado.

No Windows local passaram `go test ./internal/financial -count=1`,
`go vet ./internal/financial` e `git diff --check`. Isso valida o helper de
projeção e os limites do seu contrato; não valida integração com fontes reais,
consulta de produção a `daily_balances` ou uso da calculadora. Nenhuma fórmula,
tela ou regra financeira foi alterada. O vínculo com o fluxo consolidado e as
entradas reais continua pendente.

A CI completa `38123193176`, no commit `6983a76`, passou nos jobs Windows e
PostgreSQL/Linux. Passaram testes e `go vet`, verificações de vulnerabilidades,
migrations/sqlc, segredos, lint/tipos/testes/build do frontend, builds desktop e
worker, instalador e versão portátil sem assinatura, smoke de instalação e
desinstalação, consolidação e upload dos outputs. Isso confirma o código e o
pipeline desta alteração, mas não prova a conexão da calculadora com dados
financeiros reais.

## Mapeamento estático das fontes — 2026-10-11

Esta revisão comparou D-003-A/D-003-C com `DATA-DICTIONARY-v1.md`, a migration
`database/migrations/000001_init.up.sql` e os chamadores no código. Foi uma
leitura estática: não acessou dados financeiros reais, não alterou schema e não
executou testes.

| Entrada da calculadora | Estrutura existente | O que falta comprovar/ligar |
|---|---|---|
| Lucro elegível de janeiro até M−2 | `statement_snapshots` (`statement_type='DRE'`) e `statement_lines`, com período, execução e versão da regra | Não há consulta de produção. Ainda é preciso usar a linha aprovada “Lucro/Prejuízo Líquido” do Majucau e validar a cobertura dos meses; não copiar a DRE pronta do Bling. |
| Já investido | `investment_calculations.already_invested` guarda o valor calculado | O schema não tem tabela de movimentos de aplicação para somar aportes e subtrair resgates de principal. Rendimentos devem ficar separados. A planilha tem campo manual, mas ele não resolve o histórico. |
| Menor saldo no fluxo D0–D+60 | `daily_balances` guarda abertura, entradas, saídas, ajustes, fechamento, tipo, cenário e execução. Existem tabelas de contribuição tipadas para recebíveis, recebimentos, pagáveis, pagamentos e ajustes. | `DailyProjectionFrom` é apenas um helper de domínio. Não foi encontrada rotina que grave/procure a série usada pela calculadora; faltam seleção da execução válida, leitor de produção e conferência da fonte/data. |
| Reserva mínima | `investment_parameters.minimum_reserve` | O valor vigente aprovado é R$ 50.000. A calculadora ainda não seleciona a versão aprovada em vigor. |
| Saldo transferível do Itaú | `financial_accounts` e `financial_balance_snapshots.reconciled_amount` podem guardar uma conta e seu saldo conciliado | Não há sincronizador de contas/saldos ligado à calculadora nem regra técnica comprovada que converta saldo conciliado em saldo transferível. Não presumir essa equivalência. |

`investment_calculations` é uma estrutura de resultado, não uma fonte dos
movimentos. O schema também conserva `include_expected_receipts` e o nome
histórico `worst_week_cash`; nenhum deles autoriza somar recebimentos da
Nuvemshop separadamente ou agrupar o mínimo por semana. O Data Dictionary foi
alinhado à regra aprovada sem alteração da migration.

**Conclusão:** a especificação ganhou um mapa verificável das estruturas e
lacunas; isso não conecta a calculadora nem habilita valores. A próxima ligação
depende, entre outros pontos, de identificar onde o histórico de aportes e
resgates de principal é oficialmente registrado.

### Retificação da conferência da planilha — 11/10/2026

Uma comparação célula a célula das duas cópias recebidas confirmou que ambas
têm as mesmas 16 fórmulas. A fórmula `E12` soma os resultados mensais de
janeiro até M−2 e desconta o valor manual de `E11`; `E20`, `E21` e `E22`
calculam o caixa com recebimentos separados da Nuvemshop, o máximo limitado
pela reserva/lucro e o saldo restante. A anotação anterior que dizia que `E11`
não era usado e que o arquivo não somava os meses até M−2 estava incorreta e
foi corrigida acima e no backlog. O contrato aprovado do Majucau não mudou:
aportes menos resgates devem vir de registros comprovados, e recebimentos da
Nuvemshop não são somados separadamente ao fluxo consolidado.

Esta conferência foi somente leitura; nenhum macro foi executado. Nenhuma
fórmula, tela ou regra financeira foi alterada.

## Confirmação do arquivo reenviado — 2026-10-12

O arquivo `Calculadora_Aplicacao_Lucro_Liquido_Profissional (1).xlsm` reenviado
em 2026-10-12 tem 11.093 bytes e SHA-256
`33d3aa26aad47563e315461a32b1fcf199bc28999090e1188bf71ec66e3d14a2`, igual ao
arquivo recebido anteriormente. A pasta de trabalho contém uma única aba
(`Calculadora`) e 16 fórmulas. A leitura foi somente consulta; nenhum macro foi
executado e nenhum valor financeiro preenchido na cópia foi reproduzido nesta
evidência.

A revisão dos rótulos e fórmulas confirmou os seguintes pontos da cópia antiga:

- A instrução em `B5` pede copiar “Resultado Líquido do Exercício” da DRE do
  Bling. A regra aprovada é calcular a DRE no Majucau; o relatório do Bling serve
  para comparação.
- `E11` é entrada manual de “Já Investido” e participa da fórmula de lucro
  acumulado em `E12`; isso não comprova aportes e resgates de principal. O sistema
  precisa usar movimentos rastreáveis, mantendo rendimentos separados.
- `E15` é um saldo semanal digitado e `E16` é uma previsão separada da
  Nuvemshop; `E20` soma esses campos. A regra aprovada usa o fluxo consolidado e
  o menor fechamento diário de D0 a D+60. Portanto, essa soma antiga não deve ser
  copiada para o cálculo do Majucau.
- `E21` aplica a reserva e o lucro acumulado ao máximo calculado. A comparação
  estrutural confirma como a planilha organiza os campos, mas não comprova que os
  dados manuais representem as fontes e o período aprovados.

D-003-A e D-003-C permanecem inalteradas. Continuam pendentes a consulta de
produção à DRE própria, o histórico comprovado de aplicações/resgates de
principal, a série consolidada de 61 saldos diários e a seleção da reserva
vigente. Nenhum cálculo, tela, regra financeira ou valor de entrada foi alterado.
