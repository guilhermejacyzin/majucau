# Status da T02 — 09/11/2026, 05:40 BRT

## Correção aplicada no protótipo — 09/11/2026, 05:40 BRT

O protótipo Wails real está aberto na janela `Majucau Financial Intelligence`.
A tela Visão Executiva foi atualizada ao vivo e agora mostra a linha laranja de
saldo semanal, calculada com o último saldo Itaú conciliado e os mesmos valores
semanais de Contas a Receber (líquido) e Contas a Pagar. A origem do erro era
específica: a T01 usava os períodos crus do servidor, que não traziam abertura
e fechamento, enquanto a T02 já montava os períodos calculados. A T01 agora usa
a mesma série calculada que a T02.

Às 05:40 BRT, o processo do protótipo real foi encontrado em execução. A
ferramenta de captura da janela falhou duas vezes por tempo esgotado, então não
foi possível fazer uma nova conferência visual nesta atualização. O conteúdo
anteriormente observado pela árvore de acessibilidade confirmou os valores
abaixo; isso não substitui a conferência visual da responsável.

Valores visíveis depois da correção:

- Saldo base: R$ 62.548,33.
- Semana 09–14/11: entradas R$ 34.355,71; saídas R$ 9.486,24; fechamento
  calculado R$ 87.417,80.
- Semana 22–28/11: entradas líquidas R$ 43.790,66; saídas R$ 25.275,71;
  fechamento calculado R$ 1.287,90.

A consulta direta do snapshot confirmou 9 períodos e 61 dias, sem dias com
entrada ou saída ausente. O rótulo “Parcial” foi retirado da apresentação da
T02, conforme orientação da responsável. O estado de completude continua
preservado internamente para diagnóstico. A tela informa que as entradas e
saídas vêm das telas de Contas a Receber e Contas a Pagar. O saldo final é
calculado por saldo inicial + receitas − despesas.

**Correção de diagnóstico:** o registro anterior dizia que a carga em andamento
era a causa da tela vazia. Isso estava errado. Já existia um retrato normalizado
com valores; a falha era a T01 não reutilizar o fechamento semanal calculado.
Uma carga completa ainda é necessária para atualizar e homologar todas as
semanas. O snapshot atualmente lido tem Contas a Receber `PARTIAL` (14.526
registros) e Contas a Pagar `PARTIAL` (1.273 registros); portanto os valores da
série foram exibidos, mas a série toda ainda não está homologada.

### O que falta para a aprovação da T02

1. Concluir uma nova leitura completa de Contas a Receber e Contas a Pagar e
   confirmar a publicação dos dois relatórios.
2. Conferir todas as semanas que a responsável selecionar contra os respectivos
   relatórios, usando valor líquido em Contas a Receber e valor de Contas a
   Pagar.
3. Confirmar que o saldo inicial continua sendo o último saldo conciliado do
   Itaú e que cada fechamento segue a fórmula exibida.
4. A responsável revisar a tela aberta e aprovar os números e a apresentação.
5. T03 começa somente depois da aprovação da T02.

A tela foi corrigida; a T02 inteira ainda não foi homologada.
Evidência técnica desta correção: [`T02-EXECUTIVE-WEEKLY-BALANCE-WIRING-2026-11-09.md`](evidence/T02-EXECUTIVE-WEEKLY-BALANCE-WIRING-2026-11-09.md).

## Carga da T02 interrompida por HTTP 401 em leitura de detalhes — 09/11/2026, 03:44 BRT

**Correção de código preparada:** a carga longa de Contas a Pagar expirou depois
de preparar 67 das 104 páginas. O lote terminou `FAILED`/`RAW_ONLY`, com
`BLING_API_UNAUTHORIZED`, e não publicou os novos dados. A sessão OAuth expirou
às 03:44 enquanto eram lidos detalhes de títulos. A auditoria do adapter mostrou
que as listagens já renovavam o token após HTTP 401, mas as leituras individuais
de detalhe não acionavam essa renovação. Esse era um caminho usado pela
preparação de Contas a Pagar.

Corrigi as leituras GET de detalhe para renovar o token pelo callback protegido
do worker e repetir uma vez a leitura segura. Se a renovação falhar ou não
retornar token novo, a leitura continua falhando sem repetir com credencial
inválida. Dois testes do adapter cobrem renovação com sucesso e falha fechada;
`go test ./internal/integrations/bling` passou. Após reconstruir o worker,
executei pela interface a leitura filtrada por vencimento de 09/11/2026 a
08/01/2027. Ela foi cancelada em 10 segundos, com 500 recebíveis e zero contas
a pagar lidas. A auditoria encontrou um segundo limite: o aplicativo permitia
dois minutos, mas o serviço IPC do worker mantinha 10 segundos para o método
filtrado `bling.sync`. Corrigi a configuração para ambos os lados aceitarem
até duas horas, o suficiente para leitura paginada com os detalhes da API.

Ainda falta reconstruir/reabrir o protótipo, completar uma nova leitura real
dos dois relatórios, confirmar a publicação e comparar os valores semanais. A
correção OAuth e a correção de prazo estão locais e ainda não comprovam dados
atualizados na tela. T02 continua sem homologação visual/numérica.

**Atualização 04:09 BRT:** `scripts/verify.ps1` passou depois das duas correções:
testes Go, vet, scanner de vulnerabilidades, migrations, sqlc, scanner local de
segredos, validação do instalador, lint, tipagem e 75 testes/build do frontend.
O build terminou sem vulnerabilidades no `npm audit`. A nova leitura real ainda
precisa ser feita com o worker recém-compilado; a tentativa anterior foi
cancelada antes da correção do prazo IPC.

Evidência: [`T02-PAYABLE-DETAIL-401-REFRESH-2026-11-09.md`](evidence/T02-PAYABLE-DETAIL-401-REFRESH-2026-11-09.md).

## Correção aplicada: receita líquida da T02

O protótipo mostrava entradas R$ 0,00 apesar de Contas a Receber ter dados. A
causa confirmada foi a execução completa anterior ter expirado durante Contas a
Pagar antes de atualizar as formas de pagamento. A consulta exigia taxas da
mesma execução e convertia o valor líquido ausente em zero.

Para 22–28/11, os 146 títulos do snapshot real somam R$ 45.482,22 brutos. A
última configuração oficial válida de taxas cobre os 146 e reproduz R$ 43.790,66
líquidos, igual ao total do relatório do Bling enviado pela responsável. O
código agora usa a tabela de taxas oficial mais recente bem-sucedida da mesma
conta e prioriza a nova consulta de taxas antes da longa preparação de Contas a
Pagar.

Testes focados de dashboard/PostgreSQL, Bling e worker passaram. Às 02:36
reconstruí e abri a T02 no protótipo real: a semana 22–28/11 mostra R$ 43.790,66
de receitas e -R$ 25.275,71 de despesas, os mesmos totais dos relatórios
enviados. O saldo acumulado mostrado para a semana é R$ 898,42, calculado pela
abertura semanal exibida mais as entradas líquidas e as saídas negativas.

A sincronização nova de Contas a Pagar terminou `FAILED`/`RAW_ONLY` às 02:14,
então a tela preserva o último snapshot normalizado dessa fonte. A semana
conferida bate, mas outras semanas e uma nova carga completa ainda precisam ser
verificadas antes de homologar a T02. T03 não começou e aguarda aprovação da T02.

## Em que ponto está

O protótipo Wails real está aberto na T02 — Fluxo de Caixa. O cabeçalho mostra
última sincronização em 09/11 às 02:31 e saldo bancário R$ 62.548,33. A semana
22–28/11 já mostra receitas líquidas R$ 43.790,66 e despesas -R$ 25.275,71,
iguais aos dois relatórios enviados. Isso valida a correção dessa semana, mas
não homologa toda a série: a nova carga de Contas a Pagar não terminou e as
demais semanas ainda precisam ser comparadas.

## Resultado observado na sincronização

- Contas a Receber: `SUCCESS`, 15.155 registros lidos.
- Movimentos conciliados: `SUCCESS`, 3.162 registros lidos.
- Contas a Pagar: 104 páginas/10.301 registros lidos; detalhes preparados até
  19/104. A execução terminou às 02:14:12, sem publicar a carga completa.
- Movimentos gerais do caixa: `PARTIAL`, código
  `BLING_CASH_MOVEMENT_STATUS_MISSING`.

O lote de Contas a Pagar não ficou `RUNNING`: expirou após 30 minutos. A carga
completa não foi publicada, por isso a tela continua com o retrato anterior.

## Correção técnica da carga anterior

O commit `9384758` no GitHub aumenta o prazo compartilhado de atualização para
duas horas. O classificador de cancelamentos e prazos excedidos também foi
corrigido e publicado no commit `a787a29`. A nova build está aberta e mostra a
correção do líquido. A carga que terminou às 02:14 ocorreu antes da build nova e
permanece sem publicação completa.

## Saldo e regras aprovadas

Último valor-base aprovado: R$ 62.548,33 em 06/11/2026, informado a partir do
saldo conciliado do Itaú. O saldo atual usa apenas Itaú. A projeção do fluxo usa
Contas a Receber pelo valor líquido, Contas a Pagar pelas saídas e a base
conciliada do Itaú para abrir o saldo. Receitas − despesas não substituem o
saldo bancário atual.

## O que falta antes de aprovar a T02

1. Publicar a correção atual de seleção/ordem das taxas e reconstruir o worker.
2. Executar uma carga completa atualizada e confirmar publicação de Contas a
   Pagar, Contas a Receber, taxas e movimentos.
3. Conferir as semanas selecionadas contra os relatórios correspondentes de
   Contas a Receber (líquido) e Contas a Pagar.
4. Confirmar o saldo da primeira semana e os fechamentos semanais com a base
   conciliada do Itaú.
5. Aguardar a aprovação visual e numérica de Gisele para encerrar a T02.

Somente depois da aprovação da T02 começa a T03. Percentuais antigos de BK-062
e do projeto não foram recalculados e devem ser tratados como histórico até a
revisão do backlog inteiro.
