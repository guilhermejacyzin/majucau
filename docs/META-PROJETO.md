# Meta do Majucau

### T02 — permissão e CI conferidas — 03/11/2026, 07:44 BRT

A responsável autorizou habilitar **somente leitura** para Contas Financeiras.
Essa permissão ainda não foi aplicada na configuração do aplicativo Bling; por
isso, a API segue sem acesso às contas e a T02 continua parcialmente alimentada.
O guia oficial do Bling informa que uma alteração de escopos só vale depois de
salvar o aplicativo e revoga automaticamente as autorizações dos usuários; em
seguida, o Majucau precisa ser autorizado novamente. Referência:
[Aplicativos e escopos OAuth do Bling](https://developer.bling.com.br/aplicativos).

Atualizei a evidência técnica para corrigir o estado da CI: o commit `1d01742`
falhou porque os campos obrigatórios foram colocados no título multilinha, e o
corpo ficou vazio. O commit `762f809`, com título curto e campos no corpo,
passou na verificação de mensagem e no validador do backlog. Como esse commit
alterou somente documentação, os jobs Windows e PostgreSQL foram pulados; isso
não é uma validação de produto. BK-062 segue em 30% (70% restante), e a meta do
projeto em 12,3% (87,7% restante). Próximo: habilitar o escopo somente leitura,
autorizar o Majucau e conferir a próxima sincronização das 09:30. A T02 não
está homologada.

### T02 — quadro Semana Crítica ligado aos dados semanais — 03/11/2026, 07:30 BRT

Corrigi uma falha visível na T02: o gráfico semanal recebia dados da API, mas
o quadro “Semana Crítica” permanecia todo em traços. Agora ele apresenta o
período de maior saída semanal e as entradas/saídas dessa mesma semana a partir
da composição do relatório Bling. Valores incompletos seguem com aviso “Parcial”.
O quadro não calcula saldo, reserva mínima ou posição de caixa sem uma fonte
confirmada. O desenho e as regras financeiras não mudaram. A T01 segue usando a
mesma composição.

Passaram `go test ./internal/dashboard -count=1`, TypeScript, ESLint, os 70
testes frontend e o build Vite. Também passaram a build desktop Wails e a
compilação Windows do worker e do instalador auxiliar, todos em cópias isoladas
dentro de `artifacts/verification`. Ainda faltam ativar a permissão somente de
leitura de Contas Financeiras no Bling, abrir a nova versão com o serviço real,
executar a carga diária das 09:30 e comparar o resultado completo. A permissão
não foi alterada: o controle do navegador não conseguiu confirmar a página do
cadastro ativa. BK-062 permanece em 30% / 70% restante; a média do projeto
permanece em 12,3% / 87,7% restante. Evidência:
[`T02-CRITICAL-WEEK-DISPLAY-2026-11-03.md`](evidence/T02-CRITICAL-WEEK-DISPLAY-2026-11-03.md).

### T02 — teste real da conexão Bling — 03/11/2026, 07:17 BRT

Regra reafirmada: a T02 deve mostrar o resultado do relatório **Fluxo de Caixa
do Bling**, reproduzido com dados obtidos diretamente pela API oficial. Os
totais simples das telas T03/T04 não substituem esse resultado. A tela T01 deve
reutilizar o mesmo resultado da T02.

O worker executou o teste de conexão pela API oficial e recebeu uma linha de
`GET /contas/receber` (`pagina=1`, `limite=1`). Isso confirma que a autorização
salva permite uma leitura real, mas não é a sincronização diária completa nem
confirma os valores da T02. Às 07:17, o protótipo ainda mostrava a última carga
de 05:25; a atualização diária das 09:30 ainda era futura. A T02 continua
parcial até a publicação da carga completa e comparação com o relatório. Não
houve alteração de regra financeira ou tela. Evidência:
[`T02-BLING-API-CONNECTION-TEST-2026-11-03.md`](evidence/T02-BLING-API-CONNECTION-TEST-2026-11-03.md).

## Meta ativa no Codex — atualização de 03/11/2026

- **Objetivo:** entregar um protótipo funcional e homologável até **30/11/2026**.
- **Escopo:** concluir T02–T04 e T06–T17; a T01 deve reunir os resultados dessas telas já aprovadas.
- **Exceção confirmada:** T05 (Conciliação) permanece removida do projeto.
- **Regras de proteção:** preservar telas aprovadas e regras financeiras; concluir cada etapa com implementação, verificações, documentação e evidência antes de marcá-la pronta; enviar commits pequenos ao GitHub.
- **Andamento registrado:** a média estimada do backlog é 12,3% concluída / 87,7% restante (126 tarefas ativas; 7 concluídas, 23 em andamento e 96 não iniciadas). BK-058 e BK-059 — Contas a Pagar — estão cada uma em 5% pela correção parcial do sinal de saída; BK-062 — Fluxo de Caixa — está em 30%: os dados parciais da API oficial atravessam worker, Wails e chegam aos cards e ao gráfico semanal aprovados da T02. Após corrigir o formato da mensagem, a CI 41607450718 passou nas verificações de backlog, Windows e PostgreSQL. Ainda faltam comparar com o relatório Bling no mesmo período, completar saldo de abertura/float e homologar. A integração e o protótipo continuam sem homologação.
- **Onde encontrar:** esta meta aparece no painel **Metas/Goals** do chat e fica registrada neste arquivo (`docs/META-PROJETO.md`).

### T02 — agrupamento no formato semanal do relatório — 03/11/2026, 07:02 BRT

A série da API agora também chega agrupada por semana, no mesmo corte do
relatório Fluxo de Caixa do Bling. Os totais são decimais exatos, preservam
entrada positiva e saída negativa e ficam **Parciais** enquanto houver lacunas.
T01 e T02 consomem o mesmo snapshot. A nova versão Wails abriu no protótipo e
carregou o snapshot local; a última sincronização disponível é 05:25 BRT, antes
da execução diária das 09:30. O gráfico da T01 foi confirmado na árvore de
acessibilidade; a T02 usa o mesmo componente de gráfico. Não foi calculado
saldo projetado sem saldo inicial/float comprovado.

Verificações: `go test ./internal/dashboard ./ -count=1`, `npm run typecheck`,
70 testes frontend, `npm run build` e build pela CLI Wails passaram. O teste e
a build da interface foram realizados em cópia isolada, porque a pasta de
dependências já aberta pelo protótipo não pôde ser substituída.

BK-062 passa a **30%** (70% restante); a média estimada do backlog passa a
**12,3%** (87,7% restante; 126 tarefas ativas). Ainda falta concluir a
comparação do mesmo relatório: há diferença de R$ 0,01 em uma semana, a
primeira faixa começa em 02/11 e a carga disponível começa em 03/11, e o saldo
inicial/float não está comprovado. A série continua Parcial. Evidência:
[`T02-BLING-REPORT-WEEKLY-SERIES-2026-11-03.md`](evidence/T02-BLING-REPORT-WEEKLY-SERIES-2026-11-03.md).

### T02 — contrato público do saldo por conta — 03/11/2026

Conferi o esquema OpenAPI oficial atual. `GET /caixas` retorna lançamentos; o
campo `saldo` é somente a marca `S`/`N` de ajuste após o lançamento. A lista e o
detalhe de `GET /contas-contabeis` retornam identidade/configuração da conta,
sem saldo inicial nem saldo atual. A Central de Ajuda confirma que o saldo
inicial é definido no cadastro e que o controle de movimentações bancárias no
Bling é interno, sem integração bancária automática.

Isso confirma por que os cards de saldo continuam sem valor comprovado. Não
alterei o desenho nem inventei posição bancária. Próxima verificação: procurar
nos lançamentos da API um registro explícito de saldo inicial ou ajuste que
permita reconstruir as contas aprovadas, sem mudar o início CAPEX de
01/01/2026. BK-062 permanece em 25% e o projeto em 12,2% até a comparação e a
homologação. Evidência: `docs/evidence/T02-BLING-SALDO-SCHEMA-2026-11-03.md`.

### T02 — correção de título cancelado e comparação com o relatório — 03/11/2026, 06:30 BRT

Corrigi um valor excedente confirmado: a API do Bling grava o status cancelado como código `5`, enquanto os filtros de projeção só reconheciam palavras como `CANCELLED`. Um título de **R$ 1.794,15** era somado como saída. Agora esse código é excluído do fluxo e dos totais de contas a pagar por período. A regressão PostgreSQL reproduz esse título e confirma que ele não entra na série nem no total aberto de 90 dias.

Depois de recompilar e reiniciar o worker do protótipo, a T02 aberta passou a mostrar **Entradas 30 dias: R$ 69.994,12 (Parcial)** e **Saídas 30 dias: -R$ 262.136,32 (Parcial)**. A saída diminuiu exatamente **R$ 1.794,15**; a tela continua sem saldo atual e sem menor saldo/semana crítica, pois o saldo de abertura não está comprovado. A T01 continua usando o snapshot comum e não recebeu fórmula paralela.

Comparação com a captura aprovada do relatório do Bling: as cinco faixas de saídas disponíveis (08/11 a 09/12) agora conferem exatamente. Quatro das cinco faixas de entradas conferem; em 15–21/11 há diferença de **R$ 0,01**. A primeira faixa 02–07/11 ainda não pode ser comparada integralmente porque a sincronização disponível começa em 03/11. O saldo inicial do relatório, **R$ 74.463,36**, também não pode ser reconstruído com segurança pelo histórico parcial disponível. Portanto, a equivalência permanece incompleta e os números continuam marcados como **Parcial**.

**Progresso mantido:** BK-062 permanece em **25%** (75% restante); a média estimada do projeto permanece em **12,2%** (87,8% restante). Foram executados `go test ./internal/dashboard -count=1` com PostgreSQL local e `go vet ./internal/dashboard`; ambos passaram. Evidência: `docs/evidence/T02-CASH-FLOW-STATUS5-2026-11-03.md`.

### T02 — valores da API do Bling exibidos na tela — atualização 06:07 BRT

A leitura do serviço local entrega valores reais da última carga do Bling para
a T02: Entradas 30 dias **R$ 69.994,12 (Parcial)** e Saídas 30 dias
**-R$ 263.930,47 (Parcial)**; a série tem 61 datas. A T02 e o gráfico parcial
foram confirmados visualmente no protótipo atualizado. A T01 reutiliza esse
mesmo snapshot. Não alterei telas aprovadas nem regras financeiras.

Corrigi a leitura do lote mais recente completo (antes a tela podia buscar o
horário de execução anterior), a identificação de Itaú/Mercado Pago pelos nomes
que vêm nos próprios movimentos quando o catálogo financeiro não está
autorizado, os status numéricos dos títulos a pagar e taxas com casas decimais
extras sem valor significativo. Movimentos que chegam sem `situacao` são
incluídos em estado Parcial, conforme a orientação registrada; códigos
desconhecidos continuam sem interpretação.

O defeito que mantinha os cards vazios estava no DTO Wails: o worker já enviava
`cash_flow`, mas o aplicativo desktop não encaminhava esse campo ao React.
Acrescentei o campo ao contrato Wails e ao mapeamento, com teste de regressão
para os totais e uma data da série. A janela antiga foi substituída pelo
executável Wails recompilado; no protótipo aberto, T02 mostra os dois valores
Parciais e o gráfico mostra a série do Bling.

**Progresso:** BK-062 passa a 25% (75% restante); o projeto permanece em 12,2%
(87,8% restante), conforme o critério já registrado. Os valores estão visíveis
na tela e no serviço, mas a série ainda não foi comparada ao relatório Bling
com a mesma janela e opções, não reproduz o saldo inicial/float, e portanto não
comprova equivalência nem conclusão. Evidência:
`docs/evidence/T02-FLOW-SNAPSHOT-2026-11-03.md`.

### T02 — primeira ligação do fluxo de caixa ao snapshot — 03/11/2026

O snapshot agora compõe uma série diária limitada a 61 datas (hoje até D+60)
usando apenas os lotes normalizados do Bling da mesma execução: movimentos em
Caixas, títulos a receber e a pagar, contas selecionadas Itaú/Mercado Pago e
termos conhecidos das formas de pagamento. A T02 liga os totais das próximas
30 datas aos dois cards aprovados e desenha as séries diárias de entradas e
saídas no gráfico existente. A T01 consome o mesmo campo do snapshot.

Os valores permanecem **Parciais**: a composição ainda não foi comparada com um
relatório atualizado da conta real, dias úteis/feriados e float não estão
comprovados, e a API documentada não fornece saldo inicial por conta. Por isso
não preenchemos saldo atual, menor saldo, saldo projetado nem semana crítica
por inferência. `go test ./internal/...`, `go vet` dos pacotes alterados,
`git diff --check`, build TypeScript/Vite, lint e 70 testes frontend passaram.
A CI 41587012467 passou pelo teste de integração PostgreSQL, pelos testes de
código, pela build desktop e pelo smoke Windows.
Não foi iniciada sincronização nem alterado o executável Wails que está aberto.

BK-062 avança de 15% para 20%; a estimativa geral se mantém em 12,2% concluída
e 87,8% restante, por arredondamento da média de 126 tarefas. Próximo passo:
aplicar a build ao protótipo e comparar a série
com o Fluxo de Caixa do Bling usando as contas e opções aprovadas. Evidência:
`docs/evidence/T02-FLOW-SNAPSHOT-2026-11-03.md`.

### T02 — persistência das condições das formas de pagamento — 03/11/2026

As condições das formas de pagamento agora são salvas numa tabela tipada,
vinculadas ao lote diário e à evidência RAW. A sincronização só marca essa fonte
como completa após normalizar os dados; isso evita descartar taxa, prazo,
destino e configuração de dias úteis que influenciam o relatório de Fluxo de
Caixa. Os campos ausentes continuam marcados como desconhecidos. Os testes Go
direcionados passaram. BK-062 avança de 13% para 15%; a média do projeto mantém
12,2% concluídos e 87,8% restantes após arredondamento. A série de caixa ainda
não é composta nem aparece na T02; saldo inicial, dias úteis/feriados,
comparação com o relatório atualizado e ligação visual continuam pendentes.
Evidência: `docs/evidence/T02-PAYMENT-METHOD-NORMALIZATION-2026-11-03.md`.

### T02 — leitura de contas financeiras pela API oficial — 03/11/2026

Incluí na atualização diária a consulta paginada `GET /contas-contabeis`, para
obter do Bling a identidade das contas financeiras selecionadas. A resposta é
guardada somente como dado bruto de origem, com a mesma proteção e retomada por
página das demais leituras. Se essa fonte falhar, as outras fontes ainda são
tentadas, mas a execução do dia permanece incompleta para nova tentativa.

Os testes do conector, domínio financeiro e SQLC passaram localmente. Isso ainda
não põe valores na T02: os dados de conta não são saldo, e a composição da série
continua sem uma fonte comprovada para abertura/float. Também falta ligar contas,
formas de pagamento e títulos à série, comparar os períodos com o relatório real
do Bling e disponibilizar os resultados no snapshot da tela, mantendo o saldo
bancário separado. Por isso BK-062 segue em 13% (87% restante); o projeto fica
em 12,1% concluído (87,9% restante). A CI e a atualização do protótipo ainda
precisam concluir.

Evidência: `docs/evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md`.

### T02 — ligação entre ID e nome da conta — 03/11/2026

Avanço local: a resposta de contas do Bling agora é validada e guardada de
forma tipada (ID + descrição), ligada à resposta original. A atualização diária
passa a exigir que essa fonte esteja normalizada; receber só a resposta bruta
não basta para encerrar o processo. Isso prepara a seleção das contas Itaú e
Mercado Pago usada no relatório Fluxo de Caixa. Não é saldo e ainda não alimenta
a série ou a tela 02.

Verificações locais: testes Go do conector Bling e do cálculo financeiro,
`go vet`, geração SQLC, limites Go, checksum da migration e `git diff --check`
passaram. O teste que exige PostgreSQL foi incluído, mas não rodou aqui porque
`MAJUCAU_TEST_DATABASE_URL` não está configurada. Esta migration ainda não foi
aplicada ao protótipo e esta alteração ainda precisa ser publicada e validada na
CI. BK-062 segue em 15%; o projeto segue em 12,2% concluído / 87,8% restante,
pois ainda não há série composta e visível.

### T02 — agrupamento semanal do fluxo — 03/11/2026

Foi implementado um conversor da projeção diária para os períodos de domingo a
sábado do relatório do Bling, incluindo semanas parciais no início/fim do
intervalo, saídas negativas e validação da conta diária e continuidade do
saldo. Os testes unitários e `go vet` passaram. BK-062 avança de 10% para 12%;
o projeto permanece em 12,1% concluído / 87,9% restante pela média arredondada
do backlog. A T02 continua sem série real no snapshot e sem valores no gráfico:
falta alimentar a composição com dados oficiais, resolver o saldo inicial e
validar com o relatório real. Nenhuma tela ou regra financeira foi alterada.

### T02 — destino de forma de pagamento — 03/11/2026

A API oficial documenta destinos 1 (Contas a Receber/Pagar), 2 (Ficha
Financeira) e 3 (Caixa e Bancos). O parser agora mantém o destino tipado. A
projeção futura só aceita destino conhecido 1; os outros dois destinos e
configuração ausente não são presumidos. Os testes focados e `go vet` passaram;
a CI do commit anterior concluiu verde. BK-062 avança de 12% para 13%; o projeto
continua 12,1% / 87,9%. Ainda falta ligar essa regra aos títulos, compor e
conferir o resultado real do relatório, obter o saldo inicial oficial e integrar
os dados à T02 sem alterar seu desenho. Evidência:
`docs/evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md`.

### T02 — rechecagem em execução — 03/11/2026

Na janela atual do protótipo, a Visão Executiva recebeu dados de Contas a
Receber e Contas a Pagar do Bling, ambos indicados como parciais. O saldo
bancário permaneceu indisponível e o gráfico do fluxo continuou sem série
confirmada. Isso comprova que há dados Bling chegando a outras métricas, mas
não comprova a composição do relatório Fluxo de Caixa. A T02 precisa apresentar
esse mesmo resultado, reproduzido pelas APIs oficiais e conferido contra o
relatório com as opções aprovadas; não deve substituir o resultado por totais
simples de receber/pagar. Nenhuma tela ou regra financeira foi alterada.
Evidência: `docs/evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md`.

### Decisão de fonte da T02 — 03/11/2026

A responsável confirmou que a T02 deve reproduzir o mesmo resultado do relatório
“Fluxo de Caixa” do Bling usando somente APIs oficiais. A API consultada não tem
um endpoint documentado que devolva o relatório pronto. Na decisão inicial, a
T02 estava em 5%; em 03/11/2026, o cálculo base de taxa e prazo e a validação
automática da referência semanal foram implementados, e o percentual estimado
passou para 10%, ainda sem uma série composta via API para comparação nem ligação
com a tela. A API de contas financeiras lista nomes/IDs, mas não documenta
o saldo; a API de Caixas e Bancos retorna lançamentos. O saldo por banco permanece
um dado separado do saldo projetado e não será preenchido por estimativa.
Detalhes: `docs/evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md`.

### T02 — avanço no cálculo das formas de recebimento — 03/11/2026

O domínio agora calcula taxa líquida por parcela conforme o exemplo oficial do
Bling e calcula prazo corrido. Para prazo em dias úteis, exige calendário que
considere feriados, evitando declarar como equivalente uma contagem apenas de
segunda a sexta. Foi criado um normalizador técnico para taxa, valor fixo, prazo
e configuração de dias úteis a partir do detalhe RAW da forma de pagamento.
Campos ausentes continuam desconhecidos. Testes focados e `go vet` passaram.

Esse componente ainda não alimenta a T02 nem tem conexão com uma série real. A
comparação contra o relatório do Bling, o float bancário por conta, a origem do
saldo inicial projetado e a ligação de cards/gráfico seguem pendentes. Portanto
BK-062 avança de 5% para **8%**; a T02 ainda não está pronta para homologação e
a média geral arredondada permanece em 12,1% concluída / 87,9% restante.
Evidência: `docs/evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md`.

### Rechecagem do painel conectado — 03/11/2026 (estado antes do cálculo de domínio)

Li a árvore de acessibilidade da janela Majucau em execução, sem copiar valores
financeiros para a documentação. A Visão Executiva mostrou resultados atuais de
Contas a Receber e Contas a Pagar, com a indicação de parcial no segundo. Isso
confirma que o Bling já está alimentando parte do protótipo. O saldo bancário e
a projeção de caixa continuaram indisponíveis.

A causa técnica da projeção vazia está no contrato do painel: `DashboardSnapshot`
não tem métricas para entradas/saídas projetadas, série diária, menor saldo ou
semana crítica. O gráfico `EmptyChart` é ainda um placeholder fixo. Portanto,
os dados de A Receber/A Pagar que aparecem na T01 não são automaticamente o
resultado do relatório Fluxo de Caixa. A T02 precisa receber o resultado do
relatório Bling (ou uma reprodução da API validada contra ele), e a T01 deve
reutilizar o mesmo resultado. A tela aprovada não foi alterada.

O Chrome não autorizou nesta sessão a leitura do conteúdo da aba do relatório
Bling. Não houve tentativa de contornar a política. A evidência já enviada pela
responsável continua sendo a referência visual. BK-062 permanece em 5% e o
projeto em 12,1%; esta constatação esclarece a causa, mas ainda não liga nem
homologa a projeção.

### Histórico — componentes de vencimento na T02 — 03/11/2026 (corrigido abaixo)

Preparei no worker duas métricas de vencimento para a T02, separadas das janelas
de 90 dias: títulos previstos a receber (com Nuvem Pago deduplicado pela regra
aprovada) e títulos previstos a pagar (com sinal negativo). A interface passou a
mostrar esses componentes nos cards existentes com a indicação **Parcial · A
receber/A pagar**. Eles não são apresentados como se fossem o relatório
completo do Fluxo de Caixa.

Os testes Go focados, `go vet`, os 19 testes direcionados do frontend,
`tsc --noEmit` e `vite build` passaram em ambiente isolado. A consulta de
integração PostgreSQL não foi executada localmente porque
`MAJUCAU_TEST_DATABASE_URL` não está configurada. A mudança ainda não foi
aplicada ao executável Wails nem comparada com o relatório real; a T02 continua
incompleta e BK-062 permanece em 5%, assim como o progresso geral segue em
12,1% concluído / 87,9% restante. A ligação desses valores aos cards foi
retirada na revisão posterior porque vencimentos não substituem o relatório.
Faltam a fonte automática do relatório, a série do gráfico, o saldo oficial por
banco, semana crítica, menores saldos 30/60 e comparação completa com a fonte
aprovada. Evidência: `docs/evidence/T02-30-DAY-SOURCE-CARDS-2026-11-03.md`.

### Correção da fonte dos cards T02 — 03/11/2026

Na revisão do código, encontrei que a interface ainda ligava “Entradas 30 dias”
ao total de vencimentos a receber e “Saídas 30 dias” ao total de vencimentos a
pagar. Essa ligação não cumpre a regra aprovada: os dois cards devem refletir o
relatório Fluxo de Caixa do Bling, e Caixas e Bancos fornece somente os saldos
por instituição. Retirei essa associação da interface; enquanto o resultado
oficial do relatório não estiver disponível, esses cards permanecem sem valor,
em vez de exibir totais de outras telas como substitutos. Os indicadores de
Contas a Receber/Pagar continuam disponíveis em suas telas próprias.

A referência pública da API consultada não documenta um endpoint do relatório
Fluxo de Caixa. A captura não é uma fonte contínua nem informa toda a série
diária. A T02 continua em 5% e a média geral permanece em 12,1% concluída / 87,9%
restante. Evidência: `docs/evidence/T02-30-DAY-SOURCE-CARDS-2026-11-03.md`.

### Contrato público de saldo bancário — 02/11/2026

Na especificação OpenAPI atual, `GET /caixas` lista lançamentos; o campo
`saldo` é uma marcação `S`/`N` de ajuste após o lançamento, não o saldo
bancário. `GET /contas-contabeis` lista contas por ID, descrição, tipo e alias,
sem valor de saldo. Também não encontrei rota documentada para o relatório
Fluxo de Caixa. Assim, a API consultada não fornece ainda uma fonte confirmada
para o saldo bancário atual nem para o relatório. Não vou calcular o saldo
somando movimentos parciais nem preencher a T02 com vencimentos. Próxima
investigação: identificar fontes oficiais para os dois resultados e validá-las
contra o Bling. Evidência: `docs/evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md`.

### Investigação de fonte automática do relatório — 03/11/2026

Conferi também a página oficial do Bling MCP. Ela descreve consultas feitas por
assistentes e montagem de panoramas financeiros sob demanda, mas não oferece na
documentação um contrato de dados ou agendamento para o worker do Majucau. Como
também cita ferramentas de criação/atualização, não será usado como origem
financeira canônica de produção. A T02 continua seguindo a regra aprovada: o
resultado vem do relatório Fluxo de Caixa; somente os saldos por conta vêm de
Caixas e Bancos. Ainda falta demonstrar uma ingestão oficial que reproduza o
relatório e comparar o resultado com a referência antes de liberar os valores.
Detalhes: `docs/evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md`.

### Correção de fontes do fluxo e avisos da Visão Executiva — 02/11/2026

A responsável esclareceu que o resultado da T02 vem do relatório **Fluxo de
Caixa** do Bling; somente os saldos por banco vêm de **Caixas e Bancos**.
Corrigi a interface para não preencher “Entradas 30 dias”/“Saídas 30 dias” com
os totais de Contas a Receber/Pagar. Também retirei da T01 os avisos
“Obrigações vencidas aguardando fonte Bling” e “Recebimentos em atraso
aguardando conciliação”, solicitados como desnecessários. Não alterei o desenho
dos cards nem as regras financeiras.

A captura do relatório foi transcrita como evidência: 02/11–09/12/2026, duas
contas, float e prazo de recebimento ativados, receitas de R$ 76.600,30,
despesas de -R$ 281.052,64 e saldo final projetado de -R$ 129.988,98. Ela só
mostra valores semanais, não a série diária para alimentar janelas móveis. A
referência pública da API consultada não lista um recurso de relatório de Fluxo
de Caixa; portanto, a ligação automática desse resultado ainda não está
concluída. Não inseri os números da imagem como valores permanentes. Evidência:
`docs/evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md`.

O percentual geral permanece em **12,1% concluído / 87,9% restante**; essa
correção visual e de ligação não conclui BK-062 nem valida o fluxo real. Próximo
passo: implementar a ingestão autorizada do resultado do relatório, ligar a T02
e fazer a T01 consumir os mesmos dados; comparar a série com o Bling antes de
marcar a tarefa concluída.

### Atualização de avisos das telas Estoque, Produção e Compras — 02/11/2026

Atendendo às capturas mais recentes, retirei da T09 os cards “Envelhecendo” e
“Validade” e o alerta “Lotes com validade próxima”. Na T10, retirei “Lotes com
perda acima do esperado” e “Separar plano realizado e obrigatório”. A T02
também não mostra mais “Saídas previstas sem categoria”, “Recebimentos em
atraso” e “Conciliação pendente”; permanece o aviso de saldo abaixo da reserva.
Os cards, o restante dos avisos e as regras financeiras não foram alterados.

Na T11 — Compras, mantive os quatro avisos para a responsável decidir depois de
entender. Eles não são erros reais: são mensagens-modelo com “Sem dados
confirmados”. Significados simples e dados necessários estão no registro de
conteúdo das telas. Para que o texto possa ser lido, organizei os cartões em
uma coluna dentro da área de alertas. As imagens aprovadas originais em
`docs/source/` foram preservadas como referência; não foram sobrescritas por
imagens de protótipo diferentes.

Os testes frontend e o build TypeScript/Vite passaram numa cópia isolada em
`artifacts/verification`. O executável Wails aberto não foi substituído ou
reiniciado ainda. O percentual geral continua 12,1%/87,9%; essas correções
visuais não concluem as telas nem a homologação.

### Atualização técnica — transação longa na sincronização Bling (02/11/2026, 20:39)

Confirmei que a carga antiga de Contas a Pagar permanecia aguardando o Bling dentro de uma transação PostgreSQL havia aproximadamente 42 minutos. O último lote publicado continuava com o código antigo `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`; isso é diferente do saldo bancário: Contas a Pagar deve mesmo ser negativa no fluxo de caixa.

Preparei a correção para buscar as páginas e detalhes fora da transação de publicação, guardar progresso em checkpoints limitados e publicar RAW, normalizados, snapshot, cursor e auditoria juntos ao final. Um teste de integração PostgreSQL verificou a ausência de transação aberta durante as chamadas, sinal negativo de saída, falha segura sem substituir o retrato anterior e remoção dos checkpoints após sucesso. As suítes focadas Go, `go vet` e os quatro conjuntos de integração PostgreSQL passaram em Windows e em um banco descartável local.

Esse código ainda não foi enviado ao GitHub nem aplicado ao worker antigo do protótipo; a sincronização real de Contas a Pagar ainda não terminou e a tela não foi validada com um lote novo. O backlog continua em 12,1% concluído / 87,9% restante; BK-058, BK-059 e BK-062 continuam em 5% cada. Próximo passo: registrar essa mudança em commit pequeno, enviar ao GitHub e aguardar CI; depois aplicar a nova versão ao protótipo, sem encerrar a sincronização antiga até resolver seu estado com segurança. Evidência: `docs/evidence/BLING-API-BOUNDED-READ-ATOMIC-PUBLISH-2026-11-02.md`.

### Atualização operacional — dados reais do Bling (02/11/2026, 20:13)

Esta atualização substitui a fotografia operacional mais antiga registrada abaixo. No PostgreSQL isolado, o Bling aparece conectado. `/caixas` terminou `SUCCESS`/`NORMALIZED` e trouxe 31.032 movimentos; 14.385 entradas e 16.647 saídas, distribuídas em nove contas, com datas de 01/01/2026 a 30/10/2026. Todos chegaram com `situacao` vazia. A usuária esclareceu que o fluxo de caixa do Bling é a fonte registrada das despesas; ajustei os indicadores mensais existentes para considerar as linhas `R` e as que vieram sem situação, deixando a carga como `PARTIAL` para o painel mostrar o valor com aviso. Situações novas diferentes de `R` continuam fora dos totais até serem interpretadas. A interface não mudou; entradas ficam positivas e saídas negativas. Os testes locais e o teste da consulta em PostgreSQL isolado passaram, mas essa alteração ainda precisa de build, sincronização e conferência visual no protótipo.

Contas a Receber terminou `SUCCESS`/`RAW_ONLY` (14.901 registros lidos), então ainda não alimenta a tabela da tela. O último lote publicado de Contas a Pagar ainda é uma falha antiga `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`; o código foi corrigido para normalizar saldos devedores como saídas negativas e os testes focados passaram. Na rechecagem das 20:29, a tentativa mais recente estava em uma transação PostgreSQL aberta havia cerca de 30 minutos, sem publicação confirmada. A causa identificada é que a leitura dos detalhes Bling ocorre dentro da transação; o próximo ajuste precisa mover essas chamadas para staging seguro em blocos e manter a publicação final atômica. Não declarar a correção comprovada na conta até o lote terminar. A T01 continua sem a nova build e precisa ser conferida depois da atualização. O worker informou `GOOGLE_DRIVE_FOLDERS_NOT_CONFIGURED`, e não há saldo atual oficial separado por banco. Não houve alteração de tela ou das regras de sinal aprovadas. Percentuais permanecem 12,1% concluído / 87,9% restante; BK-058, BK-059 e BK-062 em 5% cada. Sem previsão confiável de término antes da carga de pagar concluir, das fontes serem normalizadas e dos dados aparecerem e serem conferidos nas telas. Consulte [status simples](./STATUS-ATUAL-2026-11-02.md) e [auditoria detalhada](./evidence/BLING-DATA-PROJECTION-AUDIT-2026-11-02.md).

### Fotografia operacional inicial — 02/11/2026, 19:29

Na primeira inspeção, a carga ainda não havia terminado: Contas a Receber estava em dados brutos, Contas a Pagar estava processando e `/caixas` ainda não tinha sido publicado. Essa fotografia foi atualizada acima após consultar os lotes reais no banco.

## O que queremos entregar

Concluir o backlog aprovado do Majucau e entregar primeiro um protótipo que possa ser testado e homologado. Depois da homologação, seguir para a finalização e entrada em produção previstas no próprio backlog.

## Regras que devem ser preservadas

- Não alterar telas já aprovadas.
- Não alterar regras financeiras ou de negócio aprovadas.
- Se houver dúvida sobre um requisito, dado ou interpretação, parar essa parte do trabalho e perguntar antes de implementar.
- Nunca alterar telas aprovadas nem regras financeiras ou de negócio sem autorização explícita.
- Registrar cada avanço, teste, pendência e próximo passo no projeto e enviar as atualizações ao GitHub.

## Andamento registrado

- **Sinal de saída em Contas a Pagar — 02/11/2026:** Gisele esclareceu que conta a pagar representa saída e deve aparecer negativa nos resultados financeiros. Removi a rejeição de `saldo` negativo na normalização; também converto para negativo os saldos que o Bling envia positivos, mantenho o sinal quando já vem negativo, permito esse valor no banco e incluí as saídas nos totais, vencimentos e linhas existentes de T04/T01. A interface/desenho não foi alterada. `scripts/verify.ps1` passou localmente; como não havia DSN de teste configurado, preparei PostgreSQL 18 descartável e executei a suíte de integração afetada. Ela encontrou uma expectativa antiga (`+150.2500`) no teste; atualizei para `-150.2500` e as quatro suítes de integração PostgreSQL passaram. A CI inicial `41533971213` detectou essa expectativa antiga; a nova CI [#41535352924](https://github.com/guilhermejacyzin/majucau/actions/runs/41535352924) passou em Windows e PostgreSQL. O protótipo ainda precisa receber a migration, sincronizar o Bling e ter os números conferidos. BK-058/BK-059 estão em 5% cada; projeto estimado em 12,1% concluído (87,9% restante). Evidências: `docs/evidence/BLING-PAYABLE-OUTFLOW-SIGN-2026-11-02.md` e `docs/evidence/PAYABLE-OUTFLOW-POSTGRES-TEST-FIX-2026-11-02.md`.
- **Fluxo de Caixa Bling — 02/11/2026:** a carga agora normaliza movimentos de `/caixas`, separa entrada de saída e faz o pagamento realizado aparecer negativo nos indicadores mensais existentes. A tentativa de Caixa/Bancos não é mais pulada por falha em Contas a Pagar. Os testes Go focados e os checksums passaram. Neste registro inicial, a integração PostgreSQL local ainda não tinha sido executada e a nova migration precisava chegar ao protótipo. BK-062 está em 5% (95% restante). Após a correção do sinal de Contas a Pagar, a média geral foi atualizada para 12,1% (87,9% restante). Próximo passo: CI no push, carregar as migrations, reiniciar e comparar os valores reais com o Bling. Evidência: `docs/evidence/BLING-CASH-MOVEMENTS-NORMALIZATION-2026-11-02.md`.

- **Correção do smoke de serviço Windows — 02/11/2026:** a CI [#41355117965](https://github.com/guilhermejacyzin/majucau/actions/runs/41355117965) confirmou que o worker já estava `Running`, mas o helper retornava erro porque a consulta ao Windows ultrapassava o prazo no instante da resposta. Ajustei a espera para aceitar o estado correto quando confirmado e acrescentei teste de regressão. `go test ./internal/installer` e `git diff --check` passaram localmente; nova CI necessária para concluir a validação no runner. Nenhuma tela ou regra financeira mudou. Progresso do projeto permanece 12,0% (88,0% restante); evidência: `docs/evidence/WINDOWS-WORKER-STATE-DEADLINE-2026-11-02.md`.
- **Diagnóstico claro da permissão Bling — 01/11/2026:** corrigi o tratamento do HTTP 403 recebido ao consultar detalhe de categoria. O Majucau agora informa o nome exato da permissão de leitura necessária (`Categorias — Receitas e Despesas`) e orienta refazer a autorização, sem mostrar a resposta interna do provedor. Os testes focados do adapter e worker e `git diff --check` passaram localmente. Essa mudança não altera o cadastro Bling, não conclui OAuth e não comprova conexão real. BK-041 permanece em 80% (20% restante); projeto em 12,0% (88,0% restante). Próximo passo: salvar o escopo de leitura no aplicativo Bling, concluir a autorização pelo Majucau e repetir o teste real. Evidência: `docs/evidence/BLING-CATEGORY-SCOPE-DIAGNOSTIC-2026-11-01.md`.
- **Governança do repositório — 01/11/2026:** adicionei `.github/CODEOWNERS` para solicitar revisão do proprietário do repositório em todas as mudanças e expliquei em `CONTRIBUTING.md` que isso não ativa proteção de branch nem exige aprovação/CI. A conta conectada tem permissão de escrita, sem administração; não foi possível ativar essas regras. BK-025 avançou de 70% para 80%. Backlog atual: 126 tarefas ativas (7 concluídas, 20 em andamento, 99 não iniciadas), 12,0% concluído e 88,0% restante. Atualizei a planilha oficial; preservei suas fórmulas e conferi visualmente o resumo e BK-025.

- **Estado atualizado da credencial Bling — 01/11/2026:** confirmei somente que o arquivo protegido das credenciais existe no cofre local; não li seu conteúdo. A última captura do Majucau mostra a autorização interrompida (`BLING_NOT_CONNECTED`), e uma chamada de API registrada separadamente retornou HTTP 403 ao consultar categorias. A situação correta é: o segredo não precisa ser salvo de novo; a autorização/permissão e a sincronização ainda precisam ser confirmadas no fluxo do próprio Majucau. Nenhuma permissão externa, tela ou regra foi alterada. Evidência: `docs/evidence/BLING-AUTHORIZATION-STATE-2026-11-01.md`. BK-041 continua em 80%; projeto em 11,9% concluído / 88,1% restante.
- **Correção dos verificadores locais do Windows — 01/11/2026:** corrigi a análise do template `go list` no Windows PowerShell 5.1 e a leitura das mensagens em português do verificador NSIS (BOM UTF-8). `scripts/verify.ps1` passou com testes Go, limites de arquitetura, vet, govulncheck, migrations, SQLC, scan de secrets, NSIS, lint, typecheck, 61 testes frontend, build e `npm audit` sem vulnerabilidades. O bundle `frontend/dist` não pôde ser movido pelo Windows e ficou registrado para limpeza manual sem interromper processos ativos. BK-025 permanece em 70%; projeto em 11,9% / 88,1%. Evidência: `docs/evidence/WINDOWS-LOCAL-VERIFICATION-2026-11-01.md`.
- **Navegação T16/T17 e Integrações — 01/11/2026:** corrigi o encaminhamento incorreto de Parâmetros e Usuários para Integrações; removi da tela de Integrações a configuração direta da Nuvemshop, conforme a decisão registrada; e mantive sem valores os dados de T16/T17 que ainda não têm fonte. `scripts/verify.ps1` passou com 60 testes de frontend, verificações Go, lint, typecheck, build e auditoria de dependências. T16/T17 não estão concluídas, o percentual permanece 11,9% / 88,1%, e a homologação real do Bling segue pendente. Evidência: `docs/evidence/T16-T17-ROUTING-AND-INTEGRATIONS-UI-2026-11-01.md`.
- **Correção do verificador local do backlog — 01/11/2026:** encontrei uma falha reproduzível no Windows PowerShell 5.1: a leitura do JSON estruturado tratava a lista inteira de tarefas como um único registro e recusava o backlog real. Corrigi a leitura no script de governança, sem alterar o produto. O teste de regressão e a conferência do backlog passaram tanto no Windows PowerShell 5.1.19041.6456 quanto no PowerShell 7.6.5: 126 tarefas ativas, 7 concluídas, 20 em andamento e 99 não iniciadas; 11,9% concluído e 88,1% restante. A CI [#41334076496](https://github.com/guilhermejacyzin/majucau/actions/runs/41334076496) passou nos quatro trabalhos, incluindo Windows e PostgreSQL. Esta manutenção não marca nenhuma tarefa do produto como concluída. Evidência: `docs/evidence/BACKLOG-PS51-ROOT-ARRAY-2026-11-01.md`.
- **Navegação da T01 para as telas de detalhe — 01/11/2026:** os botões e links existentes “Ver detalhes” agora abrem os módulos correspondentes (caixa, receber, pagar, DRE, estoque, produção e compras). Mantive posição, texto, desenho, cards e cálculos; não acrescentei busca externa. A T01 e os detalhes continuam usando os dados fornecidos pelo snapshot Majucau. 57 testes em 6 arquivos, lint e typecheck passaram localmente; a CI [#41274539700](https://github.com/guilhermejacyzin/majucau/actions/runs/41274539700) passou nos quatro jobs, inclusive Windows e PostgreSQL. A captura da janela expirou, então o clique manual no protótipo não foi confirmado. Evidência: `docs/evidence/EXECUTIVE-DETAIL-DATA-FLOW-2026-10-25.md`. BK-102 permanece em 1%; esta mudança cobre só a navegação, e a estimativa do projeto continua 11,9% concluído / 88,1% restante.
- **Indicação de dado parcial na T01 — 01/11/2026:** Gisele aprovou mostrar o valor parcial com o aviso “Parcial”. A interface agora acrescenta “Parcial” à linha de apoio que já existe nos cards ligados ao snapshot; mantém o mesmo valor e não adiciona elementos nem altera espaço/desenho. Valores confirmados permanecem iguais. A suíte local passou com 58 testes em 6 arquivos, além de lint e typecheck; a CI [#41279100812](https://github.com/guilhermejacyzin/majucau/actions/runs/41279100812) passou nos quatro trabalhos, incluindo Windows e PostgreSQL. Evidência: `docs/evidence/EXECUTIVE-DETAIL-DATA-FLOW-2026-10-25.md`.
- **Retirada T05 — confirmação de Gisele, 01/11/2026:** perguntada se ainda localizava a página ou referência de Conciliação, Gisele respondeu “Não localizei nada”. A página independente, navegação T05 e alerta residual já haviam sido removidos e publicados; o card aprovado de Contas a Receber permanece.
- **Auditoria de dados da T01 — 01/11/2026:** o painel e a tela de Contas a Receber compartilham o mesmo total; o mesmo vale para Contas a Pagar. Saldo por banco, menor saldo 30/60 dias, gráfico, distribuição, vendas, margem, estoque, produção e compras ainda não são entregues pelo snapshot. Sem dados, a tela mantém “—”; não há consulta direta ao Bling ao clicar. Não foi possível confirmar saldos de Itaú ou Mercado Pago porque ainda não existe saldo por conta no contrato. A tela ainda não gera aviso dinâmico de fonte faltante. O caso de valor parcial foi resolvido conforme aprovação de Gisele: agora recebe o rótulo “Parcial” na linha de apoio existente. Evidência: `docs/evidence/EXECUTIVE-DETAIL-DATA-FLOW-2026-10-25.md`. O progresso geral permanece 11,9% concluído / 88,1% restante.
- **Atualização do protótipo T01 — 01/11/2026:** confirmei que a página e o botão T05 já não existiam; a Visão Executiva ainda exibia um alerta legado que parecia manter a tela de Conciliação. Removi somente esse alerta. Mantive os demais cards da Visão Executiva, o card aprovado de Contas a Receber e todas as regras internas. O protótipo atualizou por hot reload; 46 testes, lint e typecheck passaram. A CI [#41270518884](https://github.com/guilhermejacyzin/majucau/actions/runs/41270518884) passou nos quatro jobs e o commit `e2bdb90` foi enviado ao GitHub. O progresso do projeto permanece em 11,9% concluído (88,1% restante), pois essa remoção de referência visual não conclui uma tarefa de dados do backlog.
- 127 tarefas registradas; uma (BK-055/T05) está fora de escopo. Das 126 ativas, 7 estão concluídas, 20 em andamento e 99 não iniciadas.
- Progresso estimado atualizado: **11,9% concluído; 88,1% restante**. Essa média serve para acompanhamento do backlog e não prova que o sistema esteja pronto para homologação ou produção.
- Resumo em linguagem simples, com o que está pronto, validado e pendente em 31/10/2026: [STATUS-ATUAL-2026-10-31.md](./STATUS-ATUAL-2026-10-31.md).
- Registro integral do processo, com o estado real do protótipo, decisões aprovadas, evidências, tarefas pendentes e ordem segura de retomada: [PROCESSO-REAL-PROJETO-MAJUCAU-2026-10-31.docx](../artifacts/planejamento/outputs/PROCESSO-REAL-PROJETO-MAJUCAU-2026-10-31.docx).
- **Pré-checagem da conexão real do Bling — 31/10/2026:** conferi o cofre usado pelo protótipo sem ler seu conteúdo; o arquivo protegido `bling_credentials_v1.dpapi` não existe. O protótipo não estava aberto, então iniciei `scripts/start-prototype.ps1`: o banco isolado PostgreSQL 18 iniciou na porta 55438, o worker local iniciou e a janela Wails abriu no painel sem dados confirmados. Não houve OAuth, teste da API nem leitura de dados financeiros reais. A política do navegador também negou a leitura da página Bling; não tentei contornar. A janela está aberta para a responsável entrar em Integrações e salvar o Client Secret diretamente no campo protegido. **BK-041 está 80% concluída (20% restante); o projeto está estimado em 11,8% concluído (88,2% restante).** Nenhuma tela aprovada ou regra financeira mudou. Evidência: `docs/evidence/BLING-CONNECTION-PRECHECK-2026-10-31.md`.
- **CAPEX acumulado — início confirmado e estado da integração — 26/10/2026:** Gisele confirmou que o card começa em **01/01/2026 (inclusive)**. Padronizei essa data na política da DRE, na decomposição funcional e no contrato de evidências; a implementação diária do Bling já usa essa data como início do filtro de `/caixas`. A rotina guarda movimentos como RAW, mas ainda não os classifica como CAPEX nem preenche os cards. Tentei conferir a sessão Bling autorizada no Chrome; a proteção do navegador recusou a leitura da página e proibiu contornar o bloqueio por outra superfície. O PDF e o print existentes não ligam categoria, data e valor em cada linha; por isso, os cards continuam sem valor confirmado. Nenhuma tela ou regra financeira foi alterada. BK-007 continua em 20% (80% restante); projeto em 11,8% (88,2% restante). Próximo passo: usar uma exportação sanitizada linha a linha com categoria, data de pagamento/liquidação e valor, e depois conciliar os totais antes de exibir valores.
- **Dados de Contas a Pagar do Bling ligados ao painel — 26/10/2026:** concluí a primeira conversão da carga diária completa da API para registros organizados. A obrigação, o fornecedor e a categoria mantêm suas respostas RAW e lineage; as telas Visão Executiva e Contas a Pagar usam o mesmo retrato completo. A publicação, o cursor e o sucesso são atômicos: uma referência inválida preserva o retrato anterior. Testes Go (`go test ./...`), `go vet ./...`, limites de arquitetura, verificação de migrations, busca de segredos e conferência do backlog passaram localmente. A CI [#40310803417](https://github.com/guilhermejacyzin/majucau/actions/runs/40310803417) passou nos quatro jobs, inclusive PostgreSQL e Windows. A CI usa dados de teste; não houve OAuth nem consulta à conta Bling real. Recebíveis, pagamentos realizados, caixa, CAPEX e demais módulos continuam pendentes. BK-041 permanece em 80% até comprovar autenticação, leitura e homologação reais; projeto em 11,8% concluído (88,2% restante). A data inicial aprovada para CAPEX acumulado continua 01/01/2026. Nenhuma tela ou regra financeira foi alterada. Evidência: `docs/evidence/BLING-PAYABLE-API-NORMALIZATION-2026-10-26.md`.
- **Proteção da janela da calculadora — 26/10/2026:** corrigi o cálculo para exigir a data D0 explicitamente, validar que há 61 dias consecutivos, que D0 está marcado como provisório e D+1–D+60 como projetados, que cada saldo fecha com entradas, saídas e ajustes, e que o fechamento de cada dia vira a abertura do seguinte. A sugestão de reserva também confere D0. A fórmula D-003-A/D-003-C, a reserva mínima, as telas e os valores não mudaram. O teste focal e `scripts/verify.ps1` passaram localmente. A CI [#40310803417](https://github.com/guilhermejacyzin/majucau/actions/runs/40310803417) passou nos quatro jobs, inclusive verificação PostgreSQL e Windows; os dados são de teste e não houve uso de dados reais. BK-006 permanece em 30%, pois ainda faltam ligar a regra às fontes aprovadas e provar entradas reproduzíveis; projeto em 11,8% (88,2% restante). Evidência: `docs/evidence/APPLICATION-D0-PROJECTION-VALIDATION-2026-10-26.md`.
- **Cálculo dos cards EBIT e EBITDA — 26/10/2026:** implementei a conta separada das linhas da DRE conforme o exemplo aprovado: EBIT usa o resultado operacional após depreciação/amortização; EBITDA soma de volta a despesa reconhecida no período. A função usa moeda exata, rejeita depreciação positiva e mantém os dois valores indisponíveis quando falta resultado ou depreciação comprovados. O exemplo de R$ 120.000 de EBIT com R$ 30.000 de depreciação produz R$ 150.000 de EBITDA. `scripts/verify.ps1` passou; nenhuma tela nem regra foi redesenhada. BK-007 permanece em 20% porque ainda faltam validação contábil da empresa, mapeamento de contas/fontes e conciliação com dados reais; os cards continuam sem valor real. Regra registrada em `docs/DRE-ACCOUNTING-POLICY-v1.md`.
- **Verificação da ponte OAuth do Bling — 26/10/2026:** o conector Sites confirma que “Majucau Bling Bridge” está publicada na versão 16 e no endereço cadastrado no aplicativo Bling. Abri a fonte exata da versão 16, commit `fdce727d8f3d868b5bea418d52208039054e7cf8`, em `artifacts/cache/sites/majucau-bling-bridge`. O código encaminha o retorno OAuth ao endereço local fixo `127.0.0.1:43821`; os 28 testes, a compilação e a validação do pacote passaram. A auditoria de dependências de produção não encontrou vulnerabilidades; a auditoria completa apontou quatro alertas moderados somente no conjunto de ferramentas de desenvolvimento (`drizzle-kit`/`esbuild`), sem atualizar ou rebaixar versões automaticamente. O código do Majucau já usa o endereço atual `https://api.bling.com.br/Api/v3`, portanto não encontrei necessidade de alterar o host. A verificação segura do cofre continua mostrando ausência do arquivo DPAPI do Client Secret; nenhum segredo foi lido. Não houve autorização OAuth, chamada real ou sincronização. O worker aberto continua sendo a versão anterior à correção de erro e precisa reiniciar pelo ciclo normal do protótipo. Próximos passos desta mesma tarefa: reiniciar o protótipo, salvar o Client Secret diretamente na tela do Majucau e concluir OAuth, teste e sincronização. BK-041 continua em 80% (20% restante); o projeto permanece em 11,8% (88,2% restante). Nenhuma tela ou regra financeira mudou.
- **Erro de sincronização Bling mais claro — 26/10/2026:** confirmei no código que a falta do Client Secret retornava `BLING_SECRET_REQUIRED`, mas a sincronização diária a convertia em `BLING_SYNC_FAILED`. Corrigi o código e a mensagem em português; o código também aparece no erro da sincronização manual. `go test -count=1 ./cmd/worker ./internal/integrations/bling` passou, e a CI [#40266621491](https://github.com/guilhermejacyzin/majucau/actions/runs/40266621491) passou nos quatro jobs. A alteração foi enviada no commit `2e974a7`. A conferência atual do processo ativo confirmou que o worker está em modo de desenvolvimento, usando `artifacts/retained/postgresql-dev/credentials/worker-vault`; essa pasta contém somente os marcadores `.entropy` e `.scope`, sem `bling_credentials_v1.dpapi`. O segredo não foi lido e nenhuma chamada real ao Bling aconteceu. O worker ativo foi iniciado antes da correção, então ainda precisa ser reiniciado pelo ciclo normal do protótipo para carregar o código novo. Nenhuma tela ou regra financeira foi alterada. BK-041 permanece em 80% (20% restante), e o backlog geral em 11,8% (88,2% restante). Próxima dependência: encerrar e reabrir o protótipo para atualizar o worker, salvar o Client Secret na tela do Majucau e confirmar `secret_configured:true`; depois concluir OAuth e testar a conexão real.
- **Correção da atualização compartilhada — 26/10/2026:** após sincronização Bling bem-sucedida ou importação total/parcial de recebimentos do Bling e do Nuvem Pago, a aplicação agora busca novamente o mesmo snapshot usado pela Visão Executiva e pelas telas de detalhe, sem exigir que a pessoa saia de Integrações. Operação falha não provoca atualização; erro na consulta nova mantém o último snapshot válido. A alteração foi somente de comportamento, sem mudar desenho aprovado nem regra financeira. Verificação: 26 testes de `App.test.tsx`, `npm run build` e `npm run lint` passaram. Limite confirmado: a API do Bling ainda guarda respostas em `RAW_ONLY`; esta correção não transforma nem apresenta esse conteúdo como dado financeiro. BK-040 (contrato/cobertura real) e OAuth real da BK-041 continuam necessários para completar o ciclo de dados. Esta correção está concluída (0% restante); o progresso geral do backlog permanece 11,8% (88,2% restante).
- **Revisão do JWT e do cofre Bling — 26/10/2026:** comparei o código com o [guia oficial de migração JWT do Bling](https://developer.bling.com.br/migracao-jwt). O conector já envia `enable-jwt: 1` na troca e renovação de token e nas requisições autenticadas; `go test ./internal/integrations/bling ./cmd/worker` passou. A verificação local do worker ativo encontrou apenas os metadados técnicos `.entropy` e `.scope` no cofre; falta o arquivo cifrado `bling_credentials_v1.dpapi`. O código `BlingCredentialService.Load` informa `secret_configured:false` quando essa credencial não existe. Portanto, esta instância ainda não consegue completar OAuth; a causa confirmada é configuração local ausente, não a migração JWT. Nenhum segredo foi lido. Não houve chamada real ao Bling. Próximo passo: salvar o Client Secret do aplicativo na tela do Majucau, confirmar o estado `secret_configured:true`, conectar e autorizar; depois validar a cobertura real antes de liberar dados financeiros. O navegador continuou negando acesso à aba do Bling pela política de origem, sem tentativa de contorno.
- **Mapa confirmado no backlog v2:** BK-039 — execução de trabalhos em segundo plano, 50%; BK-040 — validar cobertura real dos dados do Bling, 0%; BK-041 — adaptador/autenticação/leitura do Bling, 80%; BK-042 — leitura das três estruturas do Drive, 70%. A configuração de Client Secret e OAuth é BK-041, não BK-039.
- **Esclarecimento de execução — 25/10/2026:** as CI [#40248166091](https://github.com/guilhermejacyzin/majucau/actions/runs/40248166091) e [#40249171034](https://github.com/guilhermejacyzin/majucau/actions/runs/40249171034) terminaram com resultado geral “sucesso”, mas, por serem mudanças apenas documentais, só executaram `detect-changes` e `verify-backlog`; PostgreSQL e Windows foram pulados. A CI completa #40245975900, registrada abaixo, passou nos quatro trabalhos. Nenhuma dessas execuções chamou a API real do Bling. O protótipo local está aberto, mas o worker informa `NOT_CONFIGURED` e `secret_configured:false`; ainda não houve OAuth nem sincronização real. Para executar essa etapa, o Client Secret precisa ser salvo na tela do Majucau e a conexão autorizada.
- **Revalidação da primeira tarefa crítica — 25/10/2026:** o worker ativo respondeu à consulta somente leitura com `CLIENT_ID_SAVED=false`, `REDIRECT_URI_SAVED=false` e `SECRET_CONFIGURED=false`; portanto, esta instância ainda não recebeu um salvamento completo do Bling. Os testes locais direcionados passaram: Go (`internal/integrations/bling` e `cmd/worker`) e interface (30 testes de `App.test.tsx`/`bootstrap.test.ts`). Isso verifica o caminho local e os estados da tela, mas não substitui OAuth nem chamada real. Não encontrei falha reproduzível de código no caminho de salvamento; BK-041 continua aguardando a configuração no protótipo. Evidência: `docs/evidence/PROTOTYPE-RESUME-2026-10-25.md`.
- **CI de 25/10/2026:** o job PostgreSQL do commit `351d167` passou, incluindo a persistência de retry. O primeiro job Windows falhou no smoke de remoção do serviço; acrescentei códigos sanitizados. A execução completa [CI #40245975900](https://github.com/guilhermejacyzin/majucau/actions/runs/40245975900) passou nos quatro jobs. A falha anterior não se repetiu e não teve causa determinística confirmada. O backlog v2 mantém BK-039 (trabalhos em segundo plano) em 50%; a autenticação e leitura do Bling pertencem à BK-041 (80%), e a leitura do Drive à BK-042 (70%). BK-040 (validar cobertura real do Bling) continua em 0%. A homologação real das fontes ainda falta. Nenhuma tela aprovada ou regra financeira foi alterada.
- **Retomada em 25/10/2026:** iniciei novamente o protótipo pelo roteiro oficial. PostgreSQL isolado, worker e interface estão ativos. A leitura segura do estado do worker por Named Pipe retornou Bling `NOT_CONFIGURED` e `secret_configured:false`; o arquivo DPAPI do Client Secret não existe. Nenhum segredo foi lido; não houve OAuth, chamada real à API nem sincronização. A revisão do caminho de salvamento não encontrou defeito de código comprovado; a credencial ainda não foi configurada. No backlog v2, BK-041 (adapter de leitura Bling) está em 80%, BK-040 (cobertura de dados Bling) em 0%, BK-042 (leitura Drive) em 70% e BK-039 (execução em segundo plano) em 50%. Próximo passo da BK-041: a usuária salva o Client Secret dentro do Majucau, conecta e autoriza; depois validamos teste, paginação e integração dos dados. Nenhuma tela aprovada ou regra financeira mudou. Evidência: `docs/evidence/PROTOTYPE-RESUME-2026-10-25.md`.
- **Atualização técnica em 25/10/2026:** corrigi a gravação da quantidade de tentativas HTTP repetidas nos lotes de sincronização de Bling e Drive. A contagem só aumenta quando uma nova tentativa realmente é enviada; falhas e sucessos persistem o número no banco. Atualizei o Go para 1.26.10 e corrigi versões vulneráveis de Vite/Vitest no frontend. Os testes locais de retry, Bling e Drive passaram; a CI #40245161243 aprovou PostgreSQL e a CI #40245975900 passou por completo. BK-039 permanece em 50% no escopo de trabalhos em segundo plano; a autenticação Bling está em BK-041 (80%) e a leitura Drive em BK-042 (70%). Ainda faltam credenciais e homologação reais. Nenhuma tela aprovada ou regra financeira mudou. Evidências: `docs/evidence/RETRY-COUNT-AND-DEPENDENCIES-2026-10-25.md` e `docs/evidence/WINDOWS-WORKER-SERVICE-LIFECYCLE-2026-10-19.md`.
- **Ponto de parada em 19/10/2026 às 09:00 (Brasília):** reabri `scripts/start-prototype.ps1` no ambiente isolado do Majucau. PostgreSQL 18 limitado a `127.0.0.1:55438`; worker e interface iniciados. O Bling permanecia `NOT_CONFIGURED` (`secret_configured:false`), sem OAuth nem sincronização. A revisão de código não encontrou defeito no envio e salvamento protegido do segredo; o bloqueio era a configuração ainda não concluída. Esse fluxo e a paginação real pertencem a BK-041; BK-039 é a execução dos trabalhos agendados e segue parcial. O Client Secret deve ser salvo na tela do Majucau, nunca enviado no chat. Depois, verificar OAuth, teste, paginação e sincronização. A rotina diária e serviço persistente permanecem relacionados a BK-039; Drive e cobertura de dados seguem BK-042/BK-040. Evidência: `docs/evidence/PROTOTYPE-BLING-HANDOFF-2026-10-19.md`.
- Corrigi o horário dos arquivos do Drive: quando o worker inicia depois das 09:30, ele agora espera a próxima execução diária em vez de coletar imediatamente. Isso garante que um arquivo recebido após o horário combinado só seja processado às 09:30 do dia seguinte. A agenda do Bling permanece independente. Testes locais e a CI completa #39214216651 passaram; nenhuma tela ou regra financeira foi alterada. Evidência: `docs/evidence/DRIVE-NEXT-DAY-SCHEDULE-2026-10-19.md`.
- Corrigi a configuração do Bling: os campos Client ID e endereço de retorno agora aparecem preenchidos com os dados públicos do aplicativo já fornecidos pela usuária quando ainda não há valores salvos. Eles continuam editáveis e só são gravados ao clicar em “Salvar configuração”; valores já salvos continuam prevalecendo. O Client Secret não é preenchido, não é revelado e deve ser informado no Majucau, onde o worker o protege. Nenhuma tela aprovada ou regra financeira foi alterada. A compilação completa do app, worker, auxiliar, instalador e pacote portátil passou; 33 testes frontend, os testes Go e as verificações locais também passaram. A usuária precisa instalar/abrir o novo pacote, salvar o Client Secret do aplicativo no Majucau e autorizar a conexão. Não houve conexão real ao Bling. Evidência e hashes: `docs/evidence/BLING-INSTALLER-UPDATED-2026-10-18.md`.
- Próxima entrega em andamento: confirmar a CI PostgreSQL da validação do Drive e, depois, validar a integração real do Drive e do Bling. O caminho estrutural agora cobre as três fontes aprovadas, mas nenhum arquivo real foi importado.
- Nesta atualização, o Majucau passou a guardar páginas confirmadas do Bling e retomá-las depois de uma interrupção. A publicação do novo retrato acontece inteira ou não acontece. A CI do GitHub passou para o commit `951d4a6`.
- Nesta execução, corrigi uma suposição insegura: página vazia sem explicação do Bling não comprova que a lista terminou. O sistema agora falha com segurança, preserva os dados válidos anteriores e não avança o cursor. Os testes locais, `go vet`, teste PostgreSQL, verificação do backlog e CI Windows passaram no commit `aff28d3` ([CI #39173420786](https://github.com/guilhermejacyzin/majucau/actions/runs/39173420786)).
- No Drive, a listagem agora tem checkpoint persistente de cada página, versões de metadados e publicação atômica do inventário completo. Se a execução falhar, o retrato anterior permanece e a leitura pode continuar do último marcador confirmado. Depois de incluir os modelos SQLC, a CI completa passou no commit `ef83bed` nos quatro jobs Windows, PostgreSQL, backlog e detecção de mudanças. BK-042 está em 52%; não houve homologação real.
- Em 18/10, a leitura do Drive passou a percorrer subpastas dentro das mesmas três raízes aprovadas. O progresso de cada pasta e página fica salvo; se houver interrupção, o worker retoma e só publica o inventário completo. O staging privado está ligado ao importador de CSV de recebimentos futuros, que mantém arquivos atrasados como pendentes. O teste local completo passou; o teste de retomada com PostgreSQL foi acrescentado à CI. A pasta real é `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`; os IDs foram identificados, mas não foram gravados no Git nem configurados localmente. A identidade técnica/credencial ainda não foi provisionada, e o CSV disponível no Drive estava desatualizado; não houve download nem importação de valores reais. BK-042: 68% estimado; projeto: 11,7% estimado pela média das 127 tarefas. Isso não representa homologação nem prontidão de produção.
- A conferência da regra JWT do Bling confirmou que o código envia `enable-jwt: 1` ao trocar e renovar tokens e em todas as consultas autenticadas. Isso confirma compatibilidade do código, mas não a autorização da conta real.
- Em 19/10, preparei um caminho seguro para configurar a leitura do Drive sem alterar telas: a chave técnica e os três IDs são gravados juntos no cofre DPAPI; o pedido vem por Named Pipe com SID Windows autenticado; e a alteração gera auditoria sanitizada no PostgreSQL. A agenda das 09:30 continua ativa e relê a configuração protegida. As suítes focadas e `scripts/verify.ps1` passaram localmente, inclusive 33 testes de interface, build, `go vet`, `govulncheck`, scan de secrets e `npm audit`; a CI #680 passou nos quatro jobs: backlog, detecção de mudanças, PostgreSQL e Windows; a migration, as permissões e a correção dos caminhos locais foram validadas. A conta humana confirmada continua sendo `financeiro@majucau.com.br`; nenhuma credencial técnica foi criada nem configurada. Evidência: `docs/evidence/GOOGLE-DRIVE-CREDENTIAL-PROVISION-2026-10-19.md`.
- Nesta execução, o caminho de recebimentos futuros passou a validar TSV em fluxo dentro do staging privado, aplicando os mesmos limites de campo, registro e quantidade de colunas, checando cancelamento e limpando os temporários. TSV válido ou inválido continua pendente e não produz fatos financeiros, porque o contrato ainda não foi aprovado. Acrescentei um teste regressivo confirmando que uma pasta de recebimentos vazia é registrada como fonte ausente/parcial, conforme a regra aprovada para arquivos externos indisponíveis às 09:30. `scripts/verify.ps1` passou por inteiro; a CI #682 também passou nos quatro jobs: detecção de mudanças, backlog, PostgreSQL e Windows. Nenhuma tela ou regra financeira mudou. Evidência: `docs/evidence/DRIVE-TSV-STREAMING-PENDING-2026-10-19.md`. BK-042 permanece em 70%; esta validação não comprova sincronização real.
- Na etapa anterior, o worker passou a validar JSON em fluxo dentro de `02_nuvem_pago/recebimentos_futuros`; não interpreta os campos e mantém todo arquivo pendente até existir contrato aprovado. O commit `40955cb` foi enviado ao GitHub e a verificação completa local passou. O resultado remoto da CI não ficou disponível para confirmação. Evidência: `docs/evidence/DRIVE-JSON-STREAMING-PENDING-2026-10-19.md`.
- Nesta etapa, o worker também reconhece e valida `.xls` em fluxo no mesmo caminho. O leitor limita o arquivo a 64 MiB, linhas/abas/células e 256 colunas antes de alocar a linha; planilhas válidas, inválidas e acima dos limites continuam pendentes, sem criar valores financeiros. Arquivos temporários são removidos. `scripts/verify.ps1` passou integralmente e o commit `deba395` foi enviado à branch `codex/majucau-artifact-consolidation`. Consultei a conexão do GitHub, mas ela não retornou status nem execução de workflow para esse SHA; a CI remota fica **não confirmada**. Evidência: `docs/evidence/DRIVE-XLS-STREAMING-PENDING-2026-10-19.md`. BK-043 permanece em 40% (60% restante), porque validar a estrutura não substitui aprovar os layouts e interpretar os dados. O projeto segue em 11,7% concluído e 88,3% restante.
- A revisão de segurança do leitor `.xlsx` corrigiu um limite que faltava: agora textos, atributos, comentários, CDATA e declarações XML têm teto individual de 1 MiB aplicado enquanto os bytes são lidos, antes de o decoder formar o token. Cinco casos acima do limite são recusados sem publicar linhas nem deixar temporários. `scripts/verify.ps1` e `scripts/verify-backlog.ps1` passaram; o `govulncheck` não encontrou vulnerabilidades alcançáveis e registrou uma vulnerabilidade em módulo requerido sem chamada alcançável. O commit `d777be5` foi enviado e a CI [#39204918731](https://github.com/guilhermejacyzin/majucau/actions/runs/39204918731) passou nos quatro trabalhos: detecção de mudanças, backlog, Windows e PostgreSQL. Nenhuma planilha real foi lida e nenhum valor financeiro foi importado. Evidência: `docs/evidence/DRIVE-XLSX-XML-TOKEN-LIMIT-2026-10-19.md`. BK-043 continua em 40% (60% restante), e a média do projeto continua em 11,7% (88,3% restante); esta proteção não altera as regras, telas ou contratos aprovados.
- Em 19/10, implementei o cálculo aritmético da DRE própria do Majucau para valores já classificados: linhas aprovadas, período obrigatório, decimal exato e totais dependentes indisponíveis quando falta qualquer entrada. Os testes cobrem um caso financeiro de referência, falta de CMV e overflow. `scripts/verify.ps1` passou por completo: testes Go, revisão arquitetural, `go vet`, segurança, instalador e frontend (33 testes). BK-066 passou de 0% para 5%; ainda faltam mapeamento contábil, lançamentos reais, conciliação e publicação dos resultados. Não alterei telas nem regras de negócio, não conectei fontes e não marquei os valores como confirmados. A média do backlog foi recalculada para 11,8% concluído e 88,2% restante. Evidência: `docs/evidence/DRE-CALCULATION-ENGINE-2026-10-19.md`.
- A meta persistida segue **ativa**. Em 19/10, abri o Drive no mesmo Chrome e confirmei que a sessão está conectada como `financeiro@majucau.com.br`; a pasta `MAJUCAU_INTELLIGENCE_INPUT` aparece. Naquela primeira conferência, não abri nem alterei documentos. A conta humana de consulta é separada da identidade técnica somente leitura que o worker ainda precisa.
- Nesta continuação, conferi em modo somente leitura as estruturas das três fontes no Drive. O Excel de `01_perdas_operacionais` contém dados pessoais e anotações operacionais, mas não valor financeiro direto por ocorrência; passa pela validação estrutural em fluxo, sem interpretação financeira. O CSV de `03_nuvem_envio/Frete` também continua sem contrato/importador financeiro aprovado. Os arquivos visíveis tinham data de modificação de 11/09/2026, portanto permanecem pendentes como dados desatualizados. Nenhum valor, nome, telefone ou linha real foi copiado para o Git, nem dado financeiro foi importado. Evidência: `docs/evidence/GOOGLE-DRIVE-READONLY-SCHEMA-INSPECTION-2026-10-19.md`. Projeto permanece em 11,8% estimado; não houve mudança em telas ou regras financeiras.
- Corrigi uma brecha na configuração do Drive: o worker deixou de aceitar IDs de pasta vindos de variáveis do Windows. Agora a leitura só pode ser habilitada pelo pacote protegido que grava juntos a chave de serviço e os três IDs aprovados. O teste regressivo confirma que definir variáveis antigas não habilita o acesso. Isso não alterou telas nem regras financeiras; a validação real da conta de serviço continua pendente. Evidência: `docs/evidence/GOOGLE-DRIVE-PROTECTED-FOLDER-CONFIG-2026-10-19.md`.
- Corrigi a recuperação da rotina diária do Drive: quando o worker começa depois das 09:30, ele recupera a execução prevista daquele dia, na mesma identidade idempotente. Arquivos modificados depois do corte continuam pendentes até a próxima execução das 09:30. Os testes do worker e do agendador passaram. A tarefa BK-039 permanece em 50% porque o worker ainda não é instalado como serviço persistente, e a credencial real do Drive, as permissões, a leitura real e a homologação ainda faltam. Evidência: `docs/evidence/DAILY-GOOGLE-DRIVE-CATCH-UP-2026-10-19.md`.
- Em 19/10, confirmei no Chrome do perfil financeiro que a conta conectada é `financeiro@majucau.com.br` e que a pasta de entrada é `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com as três pastas esperadas. O protótipo continua usando DPAPI do usuário atual; o worker iniciado como serviço agora seleciona DPAPI no escopo da máquina e valida um marcador de escopo, mantendo ACL restrita. A verificação completa local passou: suíte Go, `go vet`, segurança, instalador e interface (33 testes). O serviço ainda não está instalado, e o provisionamento da pasta, credencial do PostgreSQL, identidade técnica do Google e sincronização real continuam pendentes. BK-039 fica em 50%; a média do projeto não muda. Evidência: `docs/evidence/DPAPI-SERVICE-SCOPE-2026-10-19.md`.
- A primeira CI da correção do Drive executou no run #39220194655: backlog passou; o verificador de mudanças rejeitou somente o formato dos percentuais no corpo do commit. Sem reescrever o histórico remoto, conforme `CONTRIBUTING.md`, enviei um commit corretivo separado. Sua CI documental passou no run #39220333458. Depois acrescentei um teste do limite exato das 09:30 e a CI completa passou no run #39220729936, com verificações Go, Windows, instalador, bundle portátil, PostgreSQL e governança. A conta real do Drive e a API real do Bling ainda não foram homologadas.
- Regra aprovada para a rotina das 09:30: processar cada fonte disponível e marcar como pendentes os dados da fonte que não chegou. Uma pasta vazia não substitui os dados anteriores nem é tratada como atualização confirmada; o sistema registra a fonte como ausente/parcial.
- Em 19/10, conectei `01_perdas_operacionais` e `03_nuvem_envio` à validação estrutural em fluxo, além do caminho já existente de recebimentos futuros. CSV, TSV, JSON, XLS e XLSX passam um arquivo por vez por staging privado e limites; os registros de status são guardados no PostgreSQL com contagens apenas, sem nomes nem conteúdo dos arquivos. As fontes de perdas e frete continuam pendentes e não alimentam a DRE, pois ainda não há contrato aprovado. `scripts/verify.ps1` passou localmente. O teste que grava e repete o estado no PostgreSQL foi adicionado, mas não pôde ser executado localmente porque esta máquina não tem `MAJUCAU_TEST_DATABASE_URL`; sua confirmação depende da CI após o push. Evidência: `docs/evidence/DRIVE-ALL-SOURCE-STRUCTURAL-STREAMING-2026-10-19.md`.
- As telas aprovadas e as regras financeiras permaneceram inalteradas.
- Nesta etapa, liguei o instalador NSIS ao ciclo do serviço Windows do worker: uma cópia versionada é registrada como `NT SERVICE\MajucauWorker`, com SID restrito, início automático atrasado, cofre/logs protegidos e remoção segura que preserva dados. Corrigi também a configuração de identidade do Named Pipe para usar uma chave de registro com permissões mínimas, em vez de uma variável de ambiente não documentada. A verificação completa local passou, incluindo os 33 testes de interface; o teste privilegiado do serviço está no CI e ainda aguarda execução. Não instalei serviço nem alterei dados deste computador. BK-039 continua em 50%; o projeto permanece em 11,8% estimado (88,2% restante). Evidência: `docs/evidence/WINDOWS-WORKER-SERVICE-LIFECYCLE-2026-10-19.md`.
- A primeira execução do smoke privilegiado no GitHub, [CI #39231983131](https://github.com/guilhermejacyzin/majucau/actions/runs/39231983131), falhou ao registrar/iniciar o serviço; backlog, mensagem do commit e PostgreSQL passaram. O instalador respondia com erro genérico, então ainda não sabemos qual etapa do Windows falhou. Acrescentei códigos de erro estáveis e diagnóstico sanitizado do Service Control Manager; os testes focados e `scripts/verify.ps1` passaram localmente. Vou usar a próxima CI para identificar e corrigir a causa antes de considerar essa etapa concluída. BK-039 continua em 50% e o projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- A segunda execução, [CI #39233455017](https://github.com/guilhermejacyzin/majucau/actions/runs/39233455017), também falhou antes de confirmar o serviço e reportou `WORKER_SERVICE_CONFIG_INVALID`; os outros trabalhos passaram e a tentativa temporária foi removida. Os eventos do Windows não identificaram a fase. Agora o código devolve um identificador estável para cada passo de preparação/registro, sem expor dados locais, para localizar e corrigir a falha na próxima CI. Após essa alteração, `scripts/verify.ps1` e `scripts/verify-backlog.ps1` passaram localmente. A etapa continua aberta: BK-039 em 50%; projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- A terceira execução, [CI #39234596899](https://github.com/guilhermejacyzin/majucau/actions/runs/39234596899), identificou `WORKER_SERVICE_PATH_VALIDATION`. O instalador passava um caminho válido com espaços em `C:\Program Files`, mas o código o analisava como uma linha de comando do serviço, exigindo aspas antes da hora. Separei as duas validações: caminhos de arquivo aceitam espaços, e o caminho do executável é citado tanto na criação quanto na atualização do serviço. A verificação completa local passou com teste regressivo para esse caso. Agora aguardo o novo smoke Windows para confirmar o ciclo inteiro. BK-039 continua em 50%; projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- A quarta execução, [CI #39235534818](https://github.com/guilhermejacyzin/majucau/actions/runs/39235534818), passou da validação do caminho, mas encontrou aspas duplicadas no executável registrado: a biblioteca Windows já cita o parâmetro de `CreateService`. Removi as aspas manuais só nesse ponto e mantive a citação explícita na atualização da configuração do serviço. O smoke ainda precisa confirmar o restante do ciclo. BK-039 continua em 50%; projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- A quinta execução, [CI #39236674784](https://github.com/guilhermejacyzin/majucau/actions/runs/39236674784), confirmou que o serviço registrou e iniciou. O smoke parou porque cobrava a existência da subpasta DPAPI `Vault`, que só é criada quando há banco configurado; a CI de serviço não tem esse banco, corretamente. Agora o instalador prepara a pasta vazia `WorkerSecrets\Vault` com ACL restrita, sem criar credenciais, e o smoke verifica sua permissão. A verificação completa local passou; aguardo a CI confirmar o restante do ciclo. BK-039 continua em 50%; projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- A sexta execução, [CI #39237629665](https://github.com/guilhermejacyzin/majucau/actions/runs/39237629665), chegou à conferência de permissões do Registro com o serviço instalado e iniciado. O teste PowerShell falhou porque o tipo .NET `RegistryRights` não está disponível nesse runner. Substituí a referência ao tipo pelos valores de máscara `KEY_READ` e `KEY_ALL_ACCESS`, iguais aos esperados pelo código Go; a etapa continua aberta até a nova CI passar no ciclo todo. BK-039 continua em 50%; projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- A sétima execução, [CI #39238763587](https://github.com/guilhermejacyzin/majucau/actions/runs/39238763587), passou nos quatro trabalhos, inclusive o smoke Windows de registrar, iniciar e remover o serviço. Também conferiu conta/SID restrito, início automático atrasado, ACLs de pastas e registro e preservação dos dados. O cofre vazio é preparado pelo instalador com permissões mínimas; nenhuma credencial foi criada. Esta etapa do serviço foi concluída com êxito e está documentada em `docs/evidence/WINDOWS-WORKER-SERVICE-LIFECYCLE-2026-10-19.md`. BK-039 permanece em 50%; ainda faltam VM limpa com reinicialização, configuração segura do PostgreSQL e homologação real do Drive e do Bling. O projeto permanece em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada.
- Em 19/10, reconstruí e abri o protótipo atual. `scripts/build-windows.ps1 -Installer` passou nas verificações Go, segurança, migrations/SQLC, lint, TypeScript, 33 testes de interface, build Wails e geração do instalador/pacote portátil. Os arquivos consolidados e hashes estão em `artifacts/windows/` e `docs/evidence/BLING-INSTALLER-UPDATED-2026-10-18.md`. A tela já tem os valores públicos do Bling como padrão quando não há configuração salva; o Client Secret não tem valor padrão. Ainda não houve OAuth, chamada real ou sincronização: é necessário salvar o segredo diretamente no Majucau e autorizar a conta. A janela de confirmação do Google Cloud também aguarda confirmação de identidade pela usuária; nenhuma conta técnica ou permissão foi criada.

- **Retomada e diagnóstico — 25/10/2026:** a consulta atual do Goals e a leitura do thread mostram a meta `active`, sem pausa ou conclusão. O controle de janela do Chrome falhou durante uma inspeção anterior do Drive, encerrando aquela tentativa de leitura; não há evidência de que tenha alterado a meta. A API disponível mostra somente o estado atual, sem histórico de transições; por isso a causa e o autor de uma pausa antiga não podem ser comprovados. Corrigi uma instrução desatualizada sobre XLS/XLSX: os leitores estruturais em fluxo existem nas três raízes, mas não convertem valores em dados financeiros. Testes dos leitores e do worker passaram; o backlog v2 confirmou 127 tarefas (7 concluídas, 20 em andamento, 100 não iniciadas; 11,8% estimado). O protótipo está ativo. BK-039 é a execução em segundo plano (50%); BK-041 é o Bling (80%); BK-040 valida cobertura Bling (0%); BK-042 é Drive (70%). Não houve OAuth nem sincronização real. Evidência: `docs/evidence/PROTOTYPE-RESUME-2026-10-25.md`.

- **Checagem do cofre Bling — 25/10/2026:** o protótipo está ativo; a consulta somente leitura ao worker retornou `NOT_CONFIGURED` e `secret_configured:false`, e o arquivo esperado não existe. A ACL concede controle total ao usuário, SYSTEM e Administradores. Nenhum segredo foi lido. A revisão do formulário, IPC e serviço DPAPI não encontrou erro de código confirmado; o Client Secret ainda não foi salvo nesta instância. Isso mantém BK-041 (Bling, 80%) pendente de configuração/OAuth real; BK-039 (rotinas em segundo plano, 50%) não é a tarefa de login. Próxima ação: salvar o segredo na tela, conectar e autorizar, depois testar a API e a paginação.

## O que ainda falta para homologar o protótipo

1. Autorizar e validar a conexão real do Bling e confirmar como a lista informa que chegou à última página.
2. Provisionar a identidade técnica de leitura com acesso somente às três raízes e executar a primeira sincronização real. CSV pode seguir o contrato já implementado; TSV, JSON, XLS e XLSX são validados de forma estrutural em fluxo no caminho de recebimentos futuros e permanecem pendentes até seus contratos serem aprovados e homologados.
3. Rodar o protótipo com dados de homologação, conferir os resultados e registrar qualquer divergência antes de produção.

Além desses pontos, há outras tarefas do backlog para concluir a operação, segurança, instalação e produção. Ainda não há prazo total confiável: a estimativa depende de validações externas e de tarefas que ainda não começaram.

## Próximo passo

Próximo passo da BK-041: salvar no Majucau o Client Secret do aplicativo Bling, conectar e autorizar a conta, testar e sincronizar para conferir a paginação real. A usuária deve digitar o segredo na própria tela, sem enviá-lo no chat. Depois, iniciar BK-040 para validar cobertura dos dados reais e continuar BK-042 para Drive; BK-039 permanece na frente de execução em segundo plano e está em 50%. Para o Drive, a identidade técnica precisa de acesso somente de leitura às três pastas autorizadas. Também faltam validar reinicialização em VM limpa e configurar PostgreSQL seguro. A CI de instalação já passou; o projeto permanece em 11,8% estimado (88,2% restante).

## Estado da meta e interrupções — 19/10/2026

- A consulta persistida do Codex (`get_goal`) confirma que a meta continua **ativa**, sem orçamento máximo configurado. Não foi chamada nenhuma ação para pausar ou concluir a meta nesta continuação.
- O histórico recente mostra encerramentos normais de turnos do chat, não encerramento da meta. Também registra esperas por CI e interrupções anteriores do Chrome quando o controle não conseguiu confirmar a URL da página. A sessão financeira do Drive foi aberta corretamente; a tentativa seguinte ao Google Cloud Console abriu uma confirmação de identidade e permanece pendente.
- O histórico antigo menciona que uma meta anterior ficou pausada e depois foi recriada, mas não contém o evento nem o agente/ação que provocou a pausa. O estado atual do Goals expõe apenas o estado vigente, não o histórico de transições. Portanto, a causa exata daquela pausa antiga **não pode ser comprovada** pelos registros disponíveis; não será inventada.
- O backlog foi contado novamente: 127 tarefas, 7 concluídas, 20 em andamento e 100 não iniciadas. BK-039 segue em 50%; o projeto permanece estimado em 11,8% concluído e 88,2% restante.
- Em 19/10, confirmei no Chrome a conta `financeiro@majucau.com.br` e as três estruturas de pastas. Os IDs estão no arquivo local ignorado `artifacts/retained/google-drive-folder-ids.json`, registrado no inventário local. A validação encontrou três escopos distintos; nenhum ID, segredo ou arquivo financeiro foi adicionado ao Git.
- Em 19/10, iniciei o protótipo atual contra o PostgreSQL 18.3 isolado do Majucau (`127.0.0.1:55438`): worker e banco responderam OK, e a rotina diária de Bling/Drive ficou armada para 09:30. A autenticação real do Bling (BK-041) continuava `NOT_CONFIGURED`; não houve OAuth nem sincronização real. O horário da verificação era 08:34, antes do primeiro disparo diário. O serviço PostgreSQL da porta 5432 não foi usado. O protótipo foi deixado aberto para a responsável salvar o Client Secret diretamente na tela e autorizar a conexão. BK-039 refere-se ao trabalho em segundo plano e continua em 50%; o projeto segue em 11,8% estimados. Evidência: `docs/evidence/PROTOTYPE-DAILY-WORKER-RUNTIME-2026-10-19.md`.
- Em 19/10 às 08:43, identifiquei duas janelas Majucau: a antiga foi fechada normalmente e o protótipo atual permaneceu aberto. O teste focal da interface passou (1 teste) e confirmou que o Client Secret ainda precisa ser informado; não houve conexão real ao Bling. A captura anterior com campo mascarado não prova que havia segredo salvo. Nenhuma tela aprovada foi alterada.
- Em 19/10 às 08:47, a suíte front-end passou (33 testes/3 arquivos) e `go test ./internal/integrations/bling ./cmd/worker` passou; isso valida a interface e o tratamento local da configuração ausente, não uma conexão real. O protótipo continua ativo, com o segredo do Bling ainda não configurado.
- Em 19/10 às 08:50, passaram `go test ./internal/integrations/googledrive ./cmd/worker`. São testes locais do leitor Drive e worker; a conta real e a identidade técnica ainda não foram usadas.
- Confirmei agora no mesmo Chrome do financeiro que o Drive está conectado como `financeiro@majucau.com.br` e que a pasta de entrada é `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com as três pastas esperadas. Em `01_perdas_operacionais`, `02_nuvem_pago/Recebimentos_futuros` e `03_nuvem_envio/Frete` havia arquivos modificados em 11/09/2026. Consultei somente nomes, pastas e datas; nenhum arquivo foi aberto, baixado, alterado ou importado. A conta humana do Drive está acessível, mas a identidade técnica do worker continua pendente (BK-042). A autenticação do Bling ainda depende do Client Secret e OAuth (BK-041). BK-039 continua em 50% no escopo de execução em segundo plano; projeto em 11,8% estimado (88,2% restante). Nenhuma tela ou regra financeira foi alterada. Evidência: `docs/evidence/GOOGLE-DRIVE-CREDENTIAL-PROVISION-2026-10-19.md`.

## Retomada de CAPEX e validação do Bling — 26/10/2026

- Revisei visualmente o relatório geral de Contas a Pagar (período 01/01/2026 a 30/09/2026): ele lista vencimento, liquidação, situação e valor pago, mas não a categoria por lançamento. O print anterior mostra CAPEX agrupado com subcategorias, sem ligar essas categorias às datas individuais. Portanto, o PDF sozinho não permite calcular com segurança o CAPEX pago no mês.
- Consultei a referência pública atual da API Bling: `GET /caixas` lista data, valor, débito/crédito e situação; o detalhe da movimentação contém a categoria e o cadastro permite percorrer a hierarquia. A referência informa que datas omitidas usam o mês atual; a sincronização existente chama essa rota sem datas. Assim, ela não cobriria o histórico do card acumulado. A tela e os valores financeiros não foram alterados.
- Corrigi o adapter para rejeitar datas de caixa inválidas e períodos invertidos antes de qualquer chamada HTTP. A usuária aprovou 01/01/2026 como início do CAPEX acumulado. A carga diária agora envia `dataInicial=2026-01-01` e `dataFinal` igual à data agendada em Brasília para `GET /caixas`; isso evita que a API use o mês atual como padrão. Testes cobrem a faixa enviada e a conversão do horário agendado. Isso prepara a consulta histórica, mas não prova que toda a paginação terminou nem publica dados financeiros.
- A checagem segura do cofre local ainda mostra apenas os arquivos de escopo/entropia, sem o arquivo DPAPI do Client Secret. Nenhum segredo foi lido. Logo, não houve login, OAuth ou sincronização real; o controle de navegador desta sessão também não permite automatizar a autenticação.
- Ainda faltam resolver a paginação segura de `GET /caixas`, obter o Client Secret no cofre e concluir OAuth, persistir movimentos e suas categorias com idempotência, e conciliar os totais reais antes de preencher os cards. A consulta histórica também precisa ser homologada na conta Bling.
- A BK-067 permanece em 25%; a correção de validação fortalece o adapter, mas não conclui a classificação, persistência, período histórico ou conciliação. O projeto permanece estimado em 11,8% concluído e 88,2% restante. Nenhuma tela aprovada ou regra financeira foi alterada.

### Atualização — documentação de paginação e acesso — 26/10/2026

- A documentação geral oficial do Bling diz que listas GET usam `pagina` e `limite` (100 por padrão). O adapter de `/caixas` envia esse limite, mas só considera página curta/vazia como terminal se a resposta confirmar o limite aplicado ou trouxer metadados suficientes; sem essa confirmação, a carga permanece incompleta.
- A documentação de limites da API informa que filtros com intervalo maior que um ano retornam erro 400. A faixa aprovada de 01/01/2026 até a data de hoje ainda está dentro desse limite; antes da virada do limite, será necessário dividir a leitura histórica em períodos sem lacunas e manter a publicação atômica.
- Tentei ler a aba Chrome já aberta em “Caixas e bancos”. O controle de navegador recusou o acesso ao domínio por uma permissão salva/política, mesmo com “Sempre permitir” aparecendo nas configurações. Não tentei contornar esse bloqueio nem cliquei em autenticação.
- A conferência de nomes de arquivos do cofre mostrou somente `.entropy` e `.scope`, sem `bling_credentials_v1.dpapi`; o segredo não foi lido. Assim, a conexão real continua sem comprovação.
- Próximo passo: restabelecer acesso autorizado do navegador ou validar via resposta/exportação sanitizada; salvar o Client Secret no Majucau sem enviá-lo pelo chat; então concluir OAuth, confirmar que `/caixas` aceita o limite e detectar o fim real da lista; depois persistir e conciliar movimentos antes de liberar os cards.

### Implementação local do limite de página de `/caixas` — 26/10/2026

- A consulta de movimentos envia `limite=100`. A resposta só é tratada como página final quando o Bling confirma o tamanho aplicado ou fornece metadados suficientes; uma página vazia sem esses dados continua ambígua e não pode substituir o retrato anterior.
- Testes locais em `internal/integrations/bling` e `go vet` passaram. Não houve chamada real porque o controle do Chrome recusou acesso à origem e o cofre segue sem o arquivo DPAPI do Client Secret.
- A etapa avançou em confiabilidade local, mas a integração real e a normalização permanecem pendentes; percentuais não mudaram.

### Atualização das telas após carga em segundo plano — 26/10/2026

- Se o worker finalizar uma carga enquanto a pessoa permanece numa tela aberta, o frontend agora relê o snapshot local a cada 60 segundos. Uma chamada em andamento impede sobreposição; fechar a aplicação limpa o timer.
- Isso atualiza somente dados já persistidos no PostgreSQL. Não dispara chamadas ao Bling ou ao Drive, nem altera o desenho ou as regras financeiras. A tela pode levar até um minuto para refletir uma carga concluída.
- `frontend/src/App.test.tsx` passou com 27 testes; `npm run typecheck` e ESLint direcionado passaram. Não houve revisão visual do protótipo ativo. As respostas Bling seguem RAW-only e ainda não preenchem as tabelas operacionais.
- BK-102 permanece em 1% e o percentual geral não muda: ainda faltam integração real, conversão RAW, reconciliação e aceites das demais telas. Evidência: `docs/evidence/EXECUTIVE-DETAIL-DATA-FLOW-2026-10-25.md`.

### Integração diária de Contas a Pagar — resultado da CI — 26/10/2026

- A API do Bling agora transforma as contas a pagar da carga diária completa em registros tipados com valores decimais exatos, datas, situação, fornecedor, categoria e vínculos RAW. A Visão Executiva e T4 leem o mesmo snapshot completo; a publicação é transacional e preserva o snapshot anterior em falha.
- A CI #40273517345 passou nos quatro jobs, incluindo migrations e permissões em PostgreSQL descartável, teste de integração da carga e preservação do snapshot, suíte Windows, frontend e builds do aplicativo, worker e instalador. O teste local PostgreSQL foi pulado por falta de uma URL de banco dedicada; não alterei o banco ativo do protótipo.
- A CI utiliza respostas de teste. Isso comprova o código e o banco, não a autenticação nem os dados reais da conta Bling; portanto, a sincronização real e a homologação continuam pendentes. BK-041 fica em 80% (20% restante) e a estimativa do projeto permanece em 11,8% concluído (88,2% restante).
- Nenhuma tela aprovada ou regra financeira mudou. A regra aprovada do CAPEX acumulado continua começando em 01/01/2026 e o CAPEX permanece fora do lucro da DRE.
- Evidência: `docs/evidence/BLING-PAYABLE-API-NORMALIZATION-2026-10-26.md`; [CI #40273517345](https://github.com/guilhermejacyzin/majucau/actions/runs/40273517345).

### Cards de vencimento em 90 dias na T01 — 02/11/2026

A responsável confirmou que os cards existentes **A RECEBER** e **A PAGAR** devem mostrar os títulos com vencimento nos próximos 90 dias. A janela começa hoje e termina antes da data de hoje + 90 dias. O card existente **MENOR SALDO 30 / 60 DIAS** permanece como está; não foram criadas faixas nem cards novos. A pagar continua como saída negativa.

O snapshot e a ligação da T01 agora usam métricas específicas de vencimento em 90 dias, em vez dos totais gerais. O SQL calcula soma e quantidade. A validação confirma a métrica de Contas a Pagar. Contas a Receber continua sem número confirmado enquanto os registros recebidos da API estiverem apenas em RAW e não forem normalizados na fonte consultada pelo painel.

`go test ./internal/dashboard` e o build TypeScript/Vite isolado passaram. O valor visto na captura da prévia antiga `127.0.0.1:4179` não deve ser interpretado como total de 90 dias. A prévia é estática; o app conectado está em `http://localhost:34115/`. A atualização do app conectado depende de reiniciar o worker com o código novo e depois completar a normalização de Contas a Receber.

O backlog permanece em **12,1% concluído / 87,9% restante**. Esse ajuste não conclui homologação real nem o projeto.

### Conferência do protótipo conectado após reinício — 02/11/2026

O worker foi recompilado e o app Wails foi reaberto com o contrato de 90 dias. A leitura sanitizada do snapshot ativo confirmou **418 contas a pagar** na janela, estado **Parcial**. Não registrei o valor financeiro em arquivo. Contas a Receber continua **Indisponível**, com zero linhas normalizadas para consulta; não mostrei zero como se fosse saldo confirmado. A prévia 4179 foi recompilada, mas continua estática e não representa os dados desse worker.

Próximo passo: concluir o normalizador de Contas a Receber e reconciliar suas linhas com o Bling; depois fazer a conferência visual no app Wails. O projeto permanece em 12,1% concluído / 87,9% restante.

### Preparação da normalização de Contas a Receber e deduplicação — 02/11/2026

Implementei localmente o caminho diário para converter Contas a Receber do Bling de RAW em fatos tipados com contato e linhagem, incluindo página de detalhe para validar saldo aberto. O worker faz as chamadas externas fora da transação de publicação; o lote final e o snapshot são publicados juntos. A migration 000018 acrescenta o snapshot de recebíveis e um escritor de contato controlado; atualizei os grants e a checagem de privilégios. A consulta do painel filtra o último snapshot válido e aplica a regra aprovada de duplicidade entre Bling e Nuvem Pago: nome normalizado + valor bruto exato, ignorando a data; repetições dentro de uma fonte são preservadas. Se não houver importação diária concluída da Nuvem, o resultado combinado aparece como Parcial.

Gisele informou que atualizou o arquivo de recebimentos futuros na pasta do Drive. O worker tentou recuperar a execução diária das 09:30 em 02/11, mas falhou antes de ler qualquer arquivo (`GOOGLE_DRIVE_FOLDERS_NOT_CONFIGURED`, 0 pastas/arquivos lidos). Os IDs estão em um manifesto local, mas o worker só os aceita no mesmo pacote protegido por DPAPI que contém a chave da conta técnica Google; essa credencial ainda não foi provisionada. O arquivo atualizado não foi importado. Próximo passo: provisionar a conta técnica com acesso somente de leitura às três pastas aprovadas, executar novamente a sincronização, validar os valores reais em T03 e conferir a consolidação idêntica em T01. Não alterei o desenho aprovado nem regras financeiras. O progresso geral permanece em **12,1% concluído / 87,9% restante** até a validação real. Consulte [evidência e próximos passos](./evidence/BLING-RECEIVABLE-NORMALIZATION-2026-11-02.md).

### Correção do pulo de Contas a Receber — 02/11/2026

Uma consulta somente de leitura ao banco isolado confirmou 15 cargas Bling de Contas a Receber em `RAW_ONLY`, nenhum fato normalizado e a tabela de snapshot da migration 000018 ainda ausente. O critério de carga diária exigia estágio `NORMALIZED` para Contas a Pagar e Caixa, mas não para Contas a Receber; por isso o worker tratava a carga incompleta como concluída e a pulava. Corrigi esse requisito para as três fontes, acrescentei regressão PostgreSQL e registrei o checksum da migration 000018. Os testes Go dos módulos Bling/dashboard/worker e `scripts/verify-migrations.ps1` passaram. A integração PostgreSQL local continua sem execução; a tentativa de criar base descartável foi bloqueada pela revisão automática e nenhum banco foi alterado. A CI deve confirmar a regressão após o push. Ainda falta aplicar a migration ao protótipo e recuperar a carga; o recebível Bling ficará parcial até a importação Nuvem, que depende da credencial técnica segura do Drive. Progresso geral permanece em **12,1% / 87,9% restante**. Ver [evidência detalhada](./evidence/BLING-RECEIVABLE-NORMALIZATION-2026-11-02.md).
### Retomada sequencial da pasta 02 — 02/11/2026

Antes de avançar para a pasta 03, foi concluída a validação da importação de recebimentos futuros da pasta 02. A cópia local do CSV foi validada e importada no banco isolado do protótipo; T03 e T01 agora refletem o mesmo resultado **Parcial**. A migration 000018 e os privilégios foram aplicados/verificados no banco local. A igualdade do arquivo do Drive e da cópia local ainda não foi demonstrada por hash, pois o Drive bloqueou o download nesta sessão; a rotina automática diária do Drive também segue pendente de credencial técnica. Não alterei o desenho aprovado nem regra financeira e não iniciei a pasta 03.

O progresso geral do backlog permanece **12,1% concluído / 87,9% restante** até que os pesos e critérios oficiais das tarefas sejam reavaliados. Evidência atualizada: [status de 02/11](./STATUS-ATUAL-2026-11-02.md) e [normalização de recebíveis](./evidence/BLING-RECEIVABLE-NORMALIZATION-2026-11-02.md).

### Atualização T02 — fontes identificadas para a projeção — 03/11/2026

Conferi o OpenAPI oficial atual do Bling e completei a captura dos vínculos de
origem de cada título: forma de pagamento, conta financeira (`portador`) e, em
receber, conta contábil. Eles agora são guardados em colunas próprias com uma
migration nova. Isso permite relacionar os títulos às configurações de taxa,
prazo e destino que o relatório Fluxo de Caixa considera. Nenhum valor foi
calculado com esses campos, e a T02 ainda não mostra a série. Os testes focais,
`go vet`, limites de arquitetura e checksums passaram localmente; a CI do commit
anterior está vermelha somente no smoke de serviço Windows, que precisa ser
reexecutado. O percentual não muda: BK-062 5% / 95% restante e projeto 12,1% /
87,9% restante. Consulte [evidência técnica T02](./evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md).
### Atualização T02 — 03/11/2026

A T02 é a tela do relatório **Fluxo de Caixa**. A captura enviada é a referência
para reproduzir e conferir o resultado; o saldo bancário oficial permanece em
**Caixas e Bancos**. Correção da nota anterior: as métricas simples de
vencimento a pagar/receber em 30 dias **não são consumidas pelos cards da T02**,
porque não substituem o resultado consolidado do relatório Fluxo de Caixa. O
desenho dos cards não mudou. Isso não conclui a T02: faltam a série completa,
semana crítica, projeções de saldo e comparação integral com o relatório. O
backlog geral permanece 12,1% concluído / 87,9% restante até nova avaliação dos
critérios oficiais. Ver
[evidência T02](./evidence/T02-30-DAY-SOURCE-CARDS-2026-11-03.md).

Na conferência da ajuda oficial em 03/11, confirmei por que os totais de T03 e
T04 não podem ser ligados diretamente à T02: o relatório considera taxa,
prazo, dias úteis e destino da forma de pagamento, além do filtro de contas e
do float selecionado. O próximo passo é implementar essa composição usando os
contratos reais da API e provar que ela bate com o relatório de referência;
somente então alimentar a série da T02 e o consolidado da T01. O Bling não
documenta uma rota pública específica para ler o relatório. Não alterei telas
nem regra financeira. Detalhes e fontes estão em
[evidência do Fluxo de Caixa](./evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md).

Em seguida, foi adicionada uma leitura diária RAW das configurações detalhadas
das formas de pagamento (destino, taxa e prazo), independente das leituras de
Contas a Pagar, Contas a Receber e Caixas e Bancos. Essa preparação não calcula
projeções nem altera regra ou tela; serve para montar e conferir a composição
com o relatório Bling. Os testes Go passaram; CI PostgreSQL e conferência real
continuam pendentes. Ver [evidência da fonte](./evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md).
