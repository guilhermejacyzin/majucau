# Status da T02 — 09/11/2026, 04:59 BRT

## Atualização verificada — 09/11/2026, 04:59 BRT

O protótipo Wails real está aberto na janela `Majucau Financial Intelligence`.
Na conferência da Visão Executiva, os indicadores e o gráfico aparecem vazios.
Não considero a T02 concluída nem os dados atuais homologados.

Uma nova leitura manual da API está ativa com a correlação
`e08c578a-8ff0-436f-b7a7-40ac8b49bf90`. O estado no banco isolado do projeto,
consultado às 04:58 BRT, é:

- Contas a Receber: `SUCCESS`, 15.157 registros.
- Movimentos conciliados: `SUCCESS`, 3.162 registros.
- Formas de pagamento: `SUCCESS`, 53 registros.
- Movimentos gerais: `PARTIAL`, 31.927 registros; falta um campo de situação
  que a API não retornou (`BLING_CASH_MOVEMENT_STATUS_MISSING`).
- Contas a Pagar: `RUNNING`, 30 de 104 páginas de detalhe preparadas (3.000
  registros preparados); ainda não publicou os registros normalizados.
- Contas contábeis: `FAILED`, sem registros (`BLING_API_READ_FAILED`).

O motivo observado para o painel vazio é a carga de Contas a Pagar ainda não
ter sido publicada. O resumo financeiro exige a mesma leitura diária completa
de Contas a Pagar, Contas a Receber e Caixa; não mistura retratos de execuções
diferentes. Por isso a interface não tem uma série atual confirmada para
mostrar enquanto essa execução permanece em andamento.

A tentativa `contas.pagar` anterior terminou `FAILED` após HTTP 401; a correção
para renovar a sessão em leituras GET de detalhe está no commit `c916117`. A
execução em andamento é a primeira verificação real dessa correção. Ela ainda
não provou que a carga termina nem que os valores semanais batem. Não reiniciar
nem duplicar essa leitura enquanto a execução estiver ativa.

### Próximas ações da T02

1. Acompanhar a execução atual até um estado terminal.
2. Se ela falhar, usar o código de erro sanitizado e os checkpoints para corrigir
   a causa e repetir a leitura com segurança.
3. Se publicar, conferir no protótipo as semanas selecionadas: entradas líquidas
   de Contas a Receber, saídas de Contas a Pagar e saldo acumulado a partir do
   último saldo conciliado do Itaú.
4. Só declarar T02 pronta depois de conferir valores e apresentação da série no
   protótipo e receber a aprovação visual/numérica da responsável.
5. T03 continua aguardando a aprovação da T02.

As decisões financeiras e de tela aprovadas não foram alteradas nesta
verificação. Esta atualização descreve o estado real no momento da consulta;
ela não é uma declaração de que a T02 foi corrigida ou homologada.

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
