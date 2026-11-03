# Status atual do Majucau — 03/11/2026

## Atualização mais recente — T02 exibe números reais na tela — 03/11/2026, 05:49 BRT

O serviço local entrega aos cards aprovados da T02: **Entradas 30 dias:
R$ 69.994,12 (Parcial)** e **Saídas 30 dias: -R$ 263.930,47 (Parcial)**.
O fluxo diário contém 61 datas, hoje até D+60. A saída segue negativa no valor
interno. A T01 reutiliza o mesmo snapshot.

As causas corrigidas foram: a tela escolhia a execução diária anterior; o
catálogo separado de contas responde 403, embora o endpoint de movimentos
forneça IDs e nomes de conta; as situações das contas a pagar são numéricas,
mas a projeção esperava palavras; e os valores de taxas vinham com oito casas
decimais. O fallback identifica somente uma conta chamada Itaú e uma chamada
Mercado Pago; quando não há match único, mantém indisponível. Os movimentos
listados com situação vazia entram em valor Parcial; situações não reconhecidas
continuam fora da soma. Nada foi alterado no desenho aprovado nem nas regras
financeiras.

O worker foi compilado e consultado por IPC; a consulta confirmou os valores e
as 61 datas. Também corrigi a camada Wails, que recebia o snapshot do worker,
mas descartava `cash_flow` antes de entregá-lo à tela. O novo executável Wails
foi aberto e conferido visualmente: os dois cards aparecem com aviso Parcial e
o gráfico exibe a série. Isso confirma a ligação técnica, mas não é aceite
financeiro/homologação. Ainda não está comprovado que a composição por APIs
equivale ao relatório pronto: ainda falta comparar a mesma janela/opções, obter
saldo inicial/float e concluir indicadores derivados. O serviço retorna
`PARTIAL` deliberadamente.

BK-062 está em 25% (75% restante); projeto em 12,2% (87,8% restante).
Detalhes e próximos passos: [evidência T02](./evidence/T02-FLOW-SNAPSHOT-2026-11-03.md).

> **Nota de leitura:** o conteúdo abaixo desta atualização preserva o histórico
> das demais frentes e diagnósticos anteriores. Quando ele disser que T02 não
> tem valores, que o catálogo de contas precisa ser liberado ou que não há série,
> essas observações são anteriores à correção acima; o serviço atual entrega a
> composição parcial documentada em `T02-FLOW-SNAPSHOT-2026-11-03.md`.

## Retomada da etapa 02 — conferência atual

Voltei à pasta `02_nuvem_pago/recebimentos_futuros`; a pasta 03 continua sem ser processada.

- No Drive, confirmei que está presente `extrato_lancamentos_futuros_02_11_2026.csv`. A prévia mostra o extrato de recebimentos futuros. O navegador informa que não consegue baixar esse arquivo nesta sessão.
- Na cópia local de mesmo nome, conferi o formato CSV separado por ponto e vírgula: 635 linhas de dados, 14 campos, todas classificadas como `Entrada`/`Venda`, com datas e valores reconhecíveis e sem ID de transação repetido. Não expus nomes nem valores neste registro.
- A importação anterior de 635 linhas foi feita usando a cópia local. Como não baixei o objeto do Drive, não há hash comparável e não posso afirmar que as duas cópias são idênticas.
- Na instância aberta do protótipo, a T01 agora mostra o card **A RECEBER** preenchido com a indicação **B2C + B2B**. Isso substitui a anotação histórica abaixo que dizia que esse card continuava vazio. A T01 não mostra a decomposição por fonte; nesta conferência não consegui abrir a T03 para validar linhas e total nela.
- A regra aprovada no código para cruzar recebíveis do Bling e Nuvem Pago é nome normalizado do cliente + valor bruto, sem usar data; ocorrências iguais são pareadas uma a uma. A leitura atual da T01, sozinha, não comprova quantos pares foram encontrados.
- O status versionado anterior ainda indica que a leitura diária direta do Drive não tem a credencial técnica configurada. Assim, a atualização automática das 09:30 continua pendente, mesmo com o arquivo visível na conta pelo navegador.

**Para fechar a etapa 02, nesta ordem:** (1) obter uma cópia legível do arquivo exato do Drive e conferir a identidade da cópia importada; (2) garantir uma carga atual normalizada de Contas a Receber do Bling; (3) conferir o cruzamento por nome + valor e o total na T03; (4) confirmar que a T01 reflete o mesmo total; (5) concluir e validar a rotina diária do Drive. Só então a pasta 02 estará encerrada e poderemos começar a 03.

## Atualização confirmada — pasta 02 do Drive

Retomei a pasta `MAJUCAU_INTELLIGENCE_INPUT/02_nuvem_pago/recebimentos_futuros` antes de qualquer outra pasta. O arquivo `extrato_lancamentos_futuros_02_11_2026.csv` foi identificado no Drive; o fluxo CSV validou e importou 635 linhas no PostgreSQL isolado do protótipo, sem erros, e a interface agora mostra os recebimentos dessa fonte como **Parcial** tanto na T03 quanto na T01. O total exibido foi conferido nas duas telas; não registrei o valor financeiro neste arquivo versionado.

**Importante:** o navegador do Drive não permitiu baixar o arquivo para conferir seu hash. A importação foi feita pela cópia local CSV com o mesmo nome; portanto, a igualdade entre a cópia no Drive e a cópia local não foi comprovada por hash. A importação do arquivo local está comprovada no banco; a leitura automática direta do Drive ainda não está concluída. A rotina diária do Drive segue sem credencial técnica configurada. Não processei a pasta 03.

A migration 000018 foi aplicada somente ao banco local isolado `majucau_dev` (127.0.0.1:55438). O script de permissões e sua verificação passaram. Ajustei a allowlist para não dar leitura ampla à tabela de vínculos: o worker só pode ler os dois campos necessários à deduplicação. O valor fica parcial até o Bling ter uma carga de Contas a Receber normalizada e disponível para comparação. Nenhuma tela aprovada ou regra financeira foi alterada.

**Próxima etapa da própria pasta 02:** fechar a leitura diária segura do Drive, atualizar a cópia pelo caminho automatizado e confirmar a conciliação sem duplicar por nome + valor bruto. A pasta 03 permanece parada até a etapa 02 ser encerrada.

**Verificações concluídas:** leitura do painel com usuário normal do aplicativo (`snapshot_query_ok`); T01 e T03 conferidas visualmente; testes `go test ./internal/dashboard -count=1`; regressão PostgreSQL `TestReadStreamSummaryPostgresRequiresNormalizedDailyStages`; checagem de checksum das migrations; `verify_runtime.sql`; `git diff --check`.

> As anotações abaixo registram o histórico anterior a essa importação; trechos que dizem que o CSV não foi importado ou que T03 está indisponível foram superados por esta atualização.

## Em que fase estamos

Estamos na etapa de **fazer a integração do Bling transformar os dados recebidos em informações que as telas conseguem mostrar**. O protótipo abre e o Bling está conectado, mas ainda não está pronto para homologação.

## O que já foi confirmado

- O login da API do Bling está conectado e o teste de conexão foi aprovado.
- O Bling respondeu com **14.901 registros de Contas a Receber** na última carga. Eles chegaram ao sistema, mas foram guardados como dados brutos; ainda não alimentam a tabela consultada pelo painel.
- O cálculo do fluxo já trata Contas a Pagar como saída negativa; testes automatizados cobrem essa regra. A tela própria de Contas a Pagar mantém o formato aprovado.
- O Bling trouxe **31.032 movimentos de caixa** entre 01/01/2026 e 30/10/2026, em nove contas. O código agora inclui entradas com situação `R` ou vazia nos indicadores mensais e marca o resultado como **Parcial** quando a situação falta. A tela ainda precisa receber essa nova versão e ser conferida.
- Conferi a T01 no protótipo agora: os cards e o gráfico continuam mostrando `—`/“Sem dados confirmados”. Isso é compatível com os estados atuais das fontes, não prova que os dados estejam aparecendo nas demais telas.
- A verificação de permissões do banco passou.

## Por que ainda há campos vazios

Conectar o Bling é como abrir a torneira: primeiro os dados chegam. Depois o Majucau precisa conferir, organizar e registrar cada tipo na tabela usada pela tela. Receber dados pela API não significa que eles já estão prontos para aparecer.

Neste momento:

- Contas a Receber chegou, mas continua em dados brutos.
- A consulta segura ao banco do protótipo confirmou 15 lotes `SUCCESS` de Contas a Receber em `RAW_ONLY`, zero recebíveis normalizados e ausência da tabela de publicação da migration 000018. O worker pulava a carga diária porque o verificador de conclusão não exigia `NORMALIZED` para essa fonte. Corrigi o critério e acrescentei uma regressão; a migration precisa ser aplicada antes de publicar os fatos.
- O erro antigo de Contas a Pagar (`BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`) era da validação do saldo em aberto do título; não significava que o saldo de Caixa e Bancos estava negativo. Conta a pagar é saída e deve ser negativa no fluxo. Às 20:39, a tentativa iniciada às 19:57 seguia esperando dentro do banco havia aproximadamente 42 minutos e ainda não tinha publicado um novo lote.
- Preparei a correção para buscar páginas e detalhes do Bling sem manter aberta a transação que publica os dados. As páginas são guardadas em checkpoints limitados; no final, RAW, valores normalizados, snapshot, cursor e auditoria são publicados juntos em uma transação. A integração PostgreSQL descartável confirmou saída negativa, falha sem publicação parcial, preservação do retrato anterior e ausência de transação aberta durante as chamadas ao Bling. Esse código ainda não foi enviado ao GitHub nem aplicado ao worker que está aberto.
- Todas as 31.032 linhas de caixa vieram com `situacao` vazia. A orientação da usuária foi usar o fluxo de caixa do Bling como fonte dos movimentos registrados; passei a mostrar esses valores com aviso **Parcial**, sem mudar o sinal: entrada positiva e saída negativa.
- A execução das 19:57 terminou Contas a Receber em `SUCCESS`/`RAW_ONLY`; os dados foram recebidos, mas ainda não passaram para a tabela usada pela tela. O worker iniciou a etapa seguinte, mas Contas a Pagar segue sem lote normalizado confirmado.
- O sistema não tem saldo atual confirmado, separado por banco. Movimento de caixa não pode ser confundido com saldo da conta.
- A leitura do Google Drive está pendente porque o worker informa que as pastas autorizadas ainda não foram configuradas nele.

## O que falta, na ordem

1. Aplicar a nova versão ao protótipo, repetir a sincronização de `/caixas` e conferir os indicadores com aviso Parcial.
2. Concluir e conferir Contas a Receber e Contas a Pagar na sincronização diária.
3. Transformar Contas a Receber em dados consultáveis, mantendo itens sem classificação como pendentes.
4. Fazer a atualização completa caber no tempo e nos limites da API sem deixar títulos de fora.
5. Ligar e conferir os dados nas telas de origem; depois conferir T01, que deve reunir esses mesmos resultados.
6. Obter uma fonte oficial de saldo atual para cada conta bancária; não inferir o saldo somando movimentos.
7. A leitura do arquivo Nuvem no Drive está bloqueada pela falta da credencial protegida da conta técnica. O worker tentou recuperar a execução diária de 09:30, mas falhou antes de ler pastas ou arquivos (`GOOGLE_DRIVE_FOLDERS_NOT_CONFIGURED`, 0 pastas/arquivos lidos). Os IDs estão em um manifesto local, porém, por segurança, o worker só aceita os IDs junto da chave técnica em um pacote DPAPI. Próximo passo: provisionar essa conta de leitura e repetir a sincronização. Ver [evidência da tentativa](./evidence/BLING-RECEIVABLE-NORMALIZATION-2026-11-02.md).
8. Conferir DRE e CAPEX conforme as regras contábeis já aprovadas. Fluxo de caixa não é igual à DRE.

## Meta e estimativa

- **Meta:** protótipo funcional e pronto para homologação até 30/11/2026.
- **Backlog:** 12,1% concluído; **87,9% falta**. BK-058, BK-059 e BK-062 continuam em 5% cada até os resultados reais aparecerem e serem conferidos.
- **Tempo:** ainda não há uma estimativa confiável. A carga de Contas a Pagar continua em andamento e o resultado não foi publicado. Vou medir o tempo quando essa etapa terminar, sem apresentar estimativa inventada.

Atualização técnica e evidências de 02/11: [staging paginado e publicação atômica de Contas a Pagar](./evidence/BLING-API-BOUNDED-READ-ATOMIC-PUBLISH-2026-11-02.md).

Nenhuma tela ou regra financeira aprovada foi alterada nesta atualização.

## Correção após a conferência da tela aprovada

A responsável confirmou que a T02 deve refletir o resultado do relatório
**Fluxo de Caixa** do Bling; **somente os saldos por instituição** vêm de
**Caixas e Bancos**. Removi da T01 os dois avisos “Obrigações vencidas
aguardando fonte Bling” e “Recebimentos em atraso aguardando conciliação”, que
foram apontados como desnecessários. Também retirei a ligação incorreta dos
cards T02 “Entradas 30 dias” e “Saídas 30 dias” com os totais das telas de
Contas a Receber e Contas a Pagar.

Da captura do relatório foi possível registrar o período 02/11/2026–09/12/2026,
duas contas selecionadas, float e prazo de recebimento ativados, saldo inicial
R$ 74.463,36, receitas R$ 76.600,30, despesas -R$ 281.052,64, ajuste R$ 0,00 e
saldo final projetado -R$ 129.988,98. A tabela tem agrupamento semanal; não
permite reconstruir com precisão todos os dias da janela móvel de 30 dias. Os
valores foram registrados como evidência, não como dados vivos do protótipo.

**Ainda falta:** ligar o resultado real do relatório à T02 e a T01. A referência
pública atual da [API do Bling](https://developer.bling.com.br/referencia)
documenta os recursos de Caixas e Bancos, Contas a Pagar e Contas a Receber; não
lista um endpoint de relatório de Fluxo de Caixa. Portanto, não marquei a
integração do relatório como concluída nem usei os números da imagem como se
fossem atualização automática. Evidência: [leitura visual do relatório](./evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md).

### Histórico — projeções provisórias de 30 dias — 03/11/2026 (corrigidas abaixo)

Naquela implementação provisória, os cards receberam vencimentos de Contas a
Receber/Pagar. A revisão posterior confirmou que isso não respeita a fonte
aprovada da T02 e removeu a ligação. Este trecho é histórico; a correção vigente
está registrada na seção “Correção da fonte dos valores T02” abaixo.

O gráfico, a semana crítica, o menor saldo projetado e o saldo atual por banco
continuam sem cálculo completo. Não inferi saldo bancário a partir dos
movimentos. A T02 ainda não está pronta nem homologada. Ver [evidência dos
componentes de 30 dias e verificações](./evidence/T02-30-DAY-SOURCE-CARDS-2026-11-03.md).

Relatório técnico completo, com evidências e limitações: [auditoria do caminho dos dados do Bling](./evidence/BLING-DATA-PROJECTION-AUDIT-2026-11-02.md).

## Ajustes adicionais de conteúdo solicitados

Na T01, foram retirados os avisos genéricos de fonte/conciliação; na T02,
“Saídas previstas sem categoria”, “Recebimentos em atraso” e “Conciliação
pendente” foram removidos, mantendo o aviso de saldo projetado abaixo da
reserva. Na T10, foram retirados “Lotes com perda acima do esperado” e
“Separar plano realizado e obrigatório”. Em T11, os quatro avisos de cacau e
fornecedores permanecem porque a responsável pediu explicação antes de decidir
se os mantém. São textos-modelo sem confirmação de dados, não falhas ativas; a
grade foi empilhada para leitura.

Teste de interface e build frontend foram executados numa cópia isolada; a
atualização ainda não foi iniciada no executável Wails que já estava aberto.
Detalhes e explicações dos quatro avisos de Compras: [registro de conteúdo das
telas](./evidence/SCREEN-CONTENT-UPDATES-2026-11-02.md).

## Ajuste pedido na tela Estoque

Na tela **09 — Estoque**, foram removidos os cards `ENVELHECENDO` e
`VALIDADE` e o alerta `Lotes com validade próxima`, conforme as duas capturas
enviadas. Os demais itens ficaram inalterados. Foi acrescentado um teste para
proteger essa composição. Evidência: [remoção dos indicadores de validade](./evidence/INVENTORY-EXPIRATION-INDICATORS-REMOVAL-2026-11-02.md).


## Período dos cards da Visão Executiva — 02/11/2026

O período dos cards existentes **A RECEBER** e **A PAGAR** foi confirmado: vencimentos nos próximos 90 dias. A T01 agora liga cada card a uma métrica específica de 90 dias, em vez do total geral. Contas a Pagar mantém sinal negativo de saída. O card **MENOR SALDO 30 / 60 DIAS** e o desenho aprovado da tela continuam inalterados.

A captura anterior da prévia `127.0.0.1:4179` mostrava um valor geral sem período e não confirma o resultado de 90 dias. Essa prévia não acessa o worker. O app conectado está em `http://localhost:34115/`. Para conferir o novo valor de Contas a Pagar nessa aplicação é necessário carregar a versão atual do worker; Contas a Receber continuará sem número até a normalização dos dados Bling RAW-only.

**Verificações:** `go test ./internal/dashboard` passou; `npm run build` passou em `artifacts/verification/frontend-live`. A prévia estática recebeu o frontend atualizado. O progresso geral permanece **12,1% concluído / 87,9% restante**.

**Próximo passo:** reiniciar o protótipo conectado com o worker atualizado, conferir Contas a Pagar em 90 dias e implementar/validar a normalização de Contas a Receber.

### Histórico — ponte dos componentes de 30 dias da T02 — 03/11/2026 (corrigida abaixo)

A tela 02 deve reproduzir o relatório **Fluxo de Caixa** do Bling. O relatório
mostrado pela responsável é a referência do resultado e da apresentação que
precisam ser conferidos. O saldo atual continua vindo de **Caixas e Bancos**;
contas a pagar/receber não substituem o resultado desse relatório.

Naquele momento, a ponte Wails passou a transportar as métricas de vencimento
de 30 dias até a interface. Depois, foi identificada e removida a ligação que
usava esses valores nos cards do Fluxo de Caixa, pois eles não substituem o
relatório aprovado. A correção vigente está na seção abaixo.

**Verificação local:** `go test -count=1 . ./internal/dashboard
./internal/application` passou. A validação PostgreSQL e a atualização/abertura
do executável Wails ainda precisam ocorrer para confirmar os valores na
instância conectada. Evidência: [T02 componentes e ponte Wails](./evidence/T02-30-DAY-SOURCE-CARDS-2026-11-03.md).

### Correção da fonte dos valores T02 — 03/11/2026

A T02 deve refletir o relatório **Fluxo de Caixa** do Bling. Os cards “Entradas
30 dias” e “Saídas 30 dias” estavam ligados aos totais de vencimentos nas telas
de Contas a Receber e Contas a Pagar; removi essa ligação. Enquanto a fonte
oficial do relatório não estiver conectada, os cards mostram “—”, preservando
seu desenho e os dados próprios nas telas de origem. **Caixas e Bancos continua
sendo a fonte somente dos saldos por instituição.**

A referência pública da API consultada não documenta um endpoint do relatório
Fluxo de Caixa, então a integração automática ainda não foi resolvida. A T02
continua em 5%; o projeto permanece em 12,1% concluído e 87,9% restante. Não
houve mudança de regra financeira aprovada, e a tela não está homologada.

### Continuação — preparação da integração de Contas a Receber

O código local agora prepara a normalização diária das contas recebidas do Bling. Cada conta mantém seu vínculo com o dado original, não é classificada automaticamente como B2B/B2C e só passa a valer no painel depois da publicação de um snapshot completo. A T01 e a tela de origem usam o mesmo conjunto publicado. A composição Bling + Nuvem segue a regra aprovada de comparar nome do cliente e valor bruto, ignorando a data; se o arquivo diário da Nuvem não estiver atualizado, o total combinado fica marcado como **Parcial**.

**Ainda não está rodando no protótipo:** falta aplicar a migration 000018 e as permissões, validar a regressão PostgreSQL na CI, iniciar o worker corrigido e conferir os números reais nas telas. Gisele avisou que o arquivo externo de recebimentos futuros foi atualizado. A tentativa diária do worker falhou antes de ler o Drive porque a credencial técnica ainda não foi provisionada; os logs registram 0 pastas e 0 arquivos lidos. O arquivo, portanto, não foi importado. Depois de aplicar a migration e executar a carga Bling, T03 poderá mostrar resultado **Parcial**; o total com Nuvem e a conferência consolidada em T01 dependem da credencial segura do Drive e da importação do arquivo.

O projeto continua em **12,1% concluído / 87,9% restante** e não está homologado. Evidência da preparação e dos passos pendentes: [normalização de Contas a Receber](./evidence/BLING-RECEIVABLE-NORMALIZATION-2026-11-02.md).

### Verificação após atualização do worker — 02/11/2026

Reiniciei o app Wails e o worker com o snapshot de 90 dias. A consulta sanitizada do worker ativo confirmou 418 títulos de Contas a Pagar na janela, com estado **Parcial**. Contas a Receber está **Indisponível** porque ainda não há linhas normalizadas; não exibi zero como valor confirmado. O total geral que aparecia na imagem da prévia não corresponde à métrica de 90 dias.

A aba estática `127.0.0.1:4179` continua sem conexão com o banco. O app local conectado permanece em `http://localhost:34115/`. Falta conferir visualmente no app e normalizar Contas a Receber. O projeto não está pronto para homologação. Meta geral: 12,1% concluído / 87,9% restante.

### Rechecagem do painel conectado — 03/11/2026

Li a árvore de acessibilidade da janela Majucau em execução, sem registrar
valores financeiros individuais. A Visão Executiva mostrou dados atuais de
Contas a Receber e Contas a Pagar, com indicação **Parcial** em Contas a Pagar.
Isso confirma que o Bling já está alimentando parte do protótipo. O saldo
bancário, o gráfico de projeção e os menores saldos continuam indisponíveis.

A causa da projeção vazia está no código: o contrato `DashboardSnapshot` ainda
não tem campos para entradas/saídas projetadas, série diária, menor saldo ou
semana crítica; `EmptyChart` continua sendo um placeholder. A T02 não pode
mostrar como resultado do relatório os totais simples das telas a pagar/receber.
A API pública consultada não documenta uma rota específica do relatório; a
composição por recursos da API precisa ser comprovada contra o Fluxo de Caixa do
Bling antes de alimentar T02 e T01. A tela e as regras aprovadas não foram
alteradas. BK-062 permanece em 5%; o projeto permanece em 12,1% concluído / 87,9%
restante.

### Preparação técnica da fonte de formas de pagamento — 03/11/2026

Foi acrescentada ao código uma leitura diária e independente dos detalhes das
formas de pagamento do Bling, preservados como RAW. Eles podem explicar destino,
taxas e prazo usados pelo relatório Fluxo de Caixa, mas ainda não são usados
para calcular valores. A etapa não bloqueia a leitura de Contas a Pagar, Contas
a Receber ou Caixas e Bancos. `go test ./internal/integrations/bling` passou
localmente. A T02 continua sem projeção publicada e BK-062 fica em 5% até haver
cálculo conectado, conferência com o relatório real e resultados visíveis.
Evidência: [fonte de configuração das formas de pagamento](./evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md).

### Continuação — ligação das contas financeiras — 03/11/2026

Para reproduzir o mesmo relatório por API oficial, a T02 precisa reconhecer quais contas foram selecionadas no Bling. Acrescentei a normalização do ID e do nome da conta, com vínculo à resposta de origem e atualização idempotente. A carga diária agora exige essa etapa normalizada para ser considerada concluída. Isso prepara a correspondência de Itaú/Mercado Pago; não cria saldo nem preenche a projeção.

`go test ./internal/integrations/bling ./internal/financial ./database/gen`, `go vet ./internal/integrations/bling`, `scripts/verify-go-boundaries.ps1`, `scripts/verify-migrations.ps1` e `git diff --check` passaram. A integração PostgreSQL foi pulada localmente por falta de `MAJUCAU_TEST_DATABASE_URL`. Ainda falta aplicar a migration ao protótipo, compor e conferir a série via API, ligá-la ao snapshot/T02 e validar no app. A tela aprovada e as regras financeiras permanecem intactas. BK-062 segue em 15%; o projeto segue em 12,2% concluído / 87,8% restante. Evidência: [T02 e relatório Fluxo de Caixa](./evidence/T02-PAYMENT-METHOD-SOURCE-2026-11-03.md).
