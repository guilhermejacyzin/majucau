# Ligação entre a Visão Executiva e as telas de detalhe — 25/10/2026

## Objetivo e conclusão

Analisar a decisão de que a primeira tela (Visão Executiva) deve resumir os
resultados das demais telas, usando as mesmas fontes, cálculos e filtros.

A direção é adequada. O desenho técnico recomendado é manter uma única base de
dados confiáveis e fazer cada tela consultar essa base. A Visão Executiva
apresenta os totais; as telas de detalhe mostram os registros que explicam os
mesmos totais. Uma tela não deve depender do texto ou da aparência de outra
tela, e navegar entre telas não deve iniciar uma nova chamada ao Bling.

Quando um dado não existe ou ainda não foi validado, a tela deve continuar
mostrando-o como pendente/indisponível. A busca de origem acontece pelo fluxo
normal do worker, como a atualização aprovada das 09:30 ou uma sincronização
manual autorizada. Uma tela não deve fazer uma busca escondida como alternativa.

## Caminho que o código usa hoje

```text
Bling / arquivos externos
        ↓
Worker Go (sincronização e validação)
        ↓
PostgreSQL local
        ↓
Snapshot de leitura, em uma única consulta consistente
        ↓
Telas de detalhe e Visão Executiva reutilizam os campos disponíveis
```

- A interface não consulta o Bling diretamente. Ela chama o aplicativo Wails,
  que pede o snapshot ao worker por IPC. Isso mantém credenciais e chamadas
  externas fora do navegador interno da aplicação.
- O worker lê as tabelas financeiras organizadas no PostgreSQL em uma
  transação somente de leitura. O resumo e as linhas limitadas das contas a
  receber e a pagar vêm da mesma visão consistente do banco.
- O frontend guarda esse snapshot em memória e o entrega à tela aberta. O
  snapshot é lido quando o aplicativo inicia, quando a pessoa volta de
  Integrações e, enquanto o aplicativo fica aberto, a cada 60 segundos. Essa
  leitura consulta o worker/PostgreSQL local; não chama serviços externos.
- A atualização diária atual do Bling lê Contas a Receber, Contas a Pagar e
  `/caixas`. As páginas são preservadas como conteúdo original (RAW), com hash,
  checkpoints e proteção contra duplicidade. `/caixas` entrou nessa carga em
  25/10/2026.

## O que já compartilha o mesmo resultado

| Resultado | Visão Executiva | Tela de detalhe | Situação no código |
|---|---|---|---|
| A Receber | Cartão `A RECEBER` | T3 `TOTAL A RECEBER`, indicadores de recebido/atrasado e linhas de recebíveis | O cartão e o total usam `snapshot.receivables`; a tabela recebe linhas do mesmo snapshot. O total elegível segue a separação aprovada entre B2C Nuvem Pago e B2B Bling. |
| A Pagar | Cartão `A PAGAR` | T4 `TOTAL A PAGAR`, vencimentos e linhas de obrigações | O cartão e o total usam `snapshot.payables`; a tabela recebe linhas do mesmo snapshot. |
| Recebido no mês | Indicador resumido no painel | T3 e T2 Fluxo de Caixa | Reutiliza `snapshot.receipts_month`. |
| Pago no mês | Indicador resumido no painel | T2 Fluxo de Caixa | Reutiliza `snapshot.payments_month`. |

Esse compartilhamento é um bom começo: o painel não refaz esses quatro
resultados com fórmulas diferentes. As consultas do resumo e das linhas usam
filtros correspondentes no banco.

## O que ainda não está ligado ao painel ou às telas de detalhe

Na captura aprovada da Visão Executiva, os resultados abaixo ainda não são
preenchidos pelo snapshot atual:

| Parte do painel | Tela de detalhe relacionada | Falta atual |
|---|---|---|
| Saldo atual | T2 Fluxo de Caixa | `/caixas` só é guardado em RAW. Falta validar a conta real, organizar os lançamentos e aplicar a regra aprovada de saldo conciliado. |
| Menor saldo em 30/60 dias e gráfico de fluxo | T2 Fluxo de Caixa | O cálculo precisa receber saldo inicial, recebimentos, pagamentos e ajustes autorizados já organizados. Os gráficos ainda são modelos vazios. |
| Distribuição do saldo por conta/meio | T2 Fluxo de Caixa / Conciliação | Falta uma visão normalizada das contas e da situação de conciliação. |
| Vendas do mês | DRE e operação | Não há métrica desse resultado no snapshot atual; falta ligar as fontes aprovadas de venda. |
| Margem bruta | T6 DRE | A base aritmética da DRE existe, mas lançamentos reais ainda não foram classificados e conciliados. O Bling não fornece a DRE Majucau pronta. |
| Estoque/cobertura | T9 Estoque | Arquivos do Drive têm leitura estrutural, mas seus valores ainda não viram registros de estoque aprovados. |
| Produção prevista | T10 Produção | Falta contrato aprovado e transformação dos dados operacionais em registros consultáveis. |
| Compras previstas | T11 Compras | Falta ligar os dados necessários de compra, produção e estoque ao resultado exibido. |
| DRE, EBIT/EBITDA e CAPEX | T6 DRE | O snapshot de leitura atual não calcula nem devolve esses resultados. CAPEX continua fora do lucro e só deve aparecer nos indicadores informativos já aprovados. |

## Lacuna que explica por que a API pode não preencher as telas

A sincronização diária já tem código para chamar as rotas do Bling e guardar
suas respostas originais. O snapshot do painel, porém, consulta as tabelas
financeiras organizadas (`receivables`, `receipts`, `payables` e `payments`),
não as páginas RAW da API.

Na busca do código de produção desta revisão, a única inserção encontrada em
`receivables` é a importação de recebimentos futuros do Nuvem Pago, e as
inserções em `receipts` vêm do relatório/importador de recebimentos. Não foi
encontrado um conversor que transforme as páginas RAW de Contas a Receber e
Contas a Pagar da API Bling em registros normalizados para o snapshot. Também
não foi encontrado um conversor de `/caixas` para movimentações organizadas,
saldo conciliado ou CAPEX.

Consequência prática: “a chamada da API passou” e “o dado apareceu nos cartões”
são etapas diferentes. Mesmo após autenticar e baixar páginas, os números só
podem alimentar as telas depois de validar os campos reais, normalizar e
reconciliar os registros conforme as regras aprovadas. Até essa etapa, os
valores devem permanecer pendentes/indisponíveis.

## Modelo recomendado para as telas

1. **Uma carga por fonte, no worker.** Bling e pastas autorizadas do Drive são
   lidos pelos fluxos já aprovados; abrir uma página não gera chamadas extras.
2. **RAW preservado.** Cada resposta original mantém origem, versão, hash e
   lote, permitindo reprocessar sem perder o material recebido.
3. **Conversão controlada.** Um conversor aprovado transforma cada registro em
   dados financeiros/operacionais tipados. Campos incertos ficam pendentes;
   categoria de origem não vira automaticamente categoria contábil.
4. **Cálculo no Majucau.** Uma única regra do domínio calcula cada resultado.
   A tela de detalhe mostra seus componentes; a Visão Executiva mostra o total
   correspondente, com o mesmo período e critérios.
5. **Snapshot compartilhado e identificável.** As telas consultam a mesma
   versão dos resultados, com data de referência e estado por fonte/métrica.
   Assim, um total no painel pode ser conferido contra as linhas de detalhe.
6. **Atualização segura.** A atualização das 09:30 publica uma nova versão
   somente quando cada fonte necessária conclui; uma falha mantém o último
   dado válido marcado com seu estado real. A interface precisa reler o
   snapshot após a atualização do worker para mostrar a versão nova.
7. **Ausência sem estimativa.** Se faltar um arquivo, campo, classificação ou
   autorização, a métrica dependente fica pendente/indisponível; não se cria
   zero nem se busca um valor por uma rota paralela não homologada.

“Buscar só quando faltar” deve significar: o worker busca a fonte no ciclo
normal, e as telas reutilizam o resultado armazenado. Se faltar somente um
campo no modelo local, primeiro se verifica se o dado já está no RAW. Uma nova
chamada ao Bling só é necessária quando o contrato oficial pedir outro recurso
ou detalhe; ela ocorre pelo adapter do worker, com retry/paginação e auditoria,
nunca como chamada direta da tela.

## Próximas etapas técnicas, em ordem

1. Confirmar autenticação e sincronização reais da conta Bling e validar o
   término da paginação de cada recurso usado.
2. Comparar respostas reais sanitizadas com o contrato de dados aprovado.
3. Implementar conversão RAW → registros tipados para Contas a Receber,
   Contas a Pagar e movimentos de caixa, preservando atomicidade e idempotência.
4. Homologar os totais das telas T3/T4 contra as linhas de detalhe e os
   relatórios originais.
5. Ligar movimentos de caixa somente após conciliação às métricas de saldo e
   fluxo. Mapear CAPEX apenas com a árvore/categoria e períodos comprovados.
6. Ligar cada indicador da Visão Executiva à respectiva tela de detalhe e às
   fontes aprovadas, uma família por vez. A DRE mantém regras próprias do
   Majucau; os dados do Bling são entradas, não a DRE pronta.
7. Atualizar o snapshot na tela após uma carga concluída e mostrar a data/estado
   corretos de cada fonte, sem mudar o desenho aprovado.
8. Demonstrar lineage de cada cartão até o cálculo, linhas que o compõem,
   registros normalizados e resposta/arquivo de origem.

## Pontos de atenção encontrados

- O snapshot é relido a cada 60 segundos enquanto o aplicativo estiver aberto.
  Uma carga concluída pode levar até esse intervalo para aparecer numa tela que
  permaneça aberta; a confirmação depende de dados já normalizados no banco.
- O cabeçalho comum de telas usa `last_bling_sync_at`, que é a tentativa mais
  recente concluída com status `SUCCESS` ou `PARTIAL`; esse horário não
  significa, por si só, que todas as fontes do painel concluíram.
- O snapshot já distingue `UNAVAILABLE` de zero confirmado quando a carga
  correspondente concluiu sem registros. Essa proteção deve continuar por
  métrica, sem esconder carga parcial de outra fonte.
- A conclusão da CI dos testes do código não comprova conexão real. A última
  evidência anterior a esta correção registrou Bling sem Client Secret aceito;
  não houve chamada autenticada à conta real durante esta alteração.

## Proteção e escopo

Esta revisão é de arquitetura e leitura de código. Não altera telas, fórmulas,
regras financeiras, categorias, filtros, permissões ou dados. A análise não
homologa o Bling real, os relatórios do Drive nem o protótipo com dados reais.

Fontes revisadas: `AGENTS.md`, `API-INTEGRATION-MAP.md`,
`docs/evidence/DASHBOARD-SNAPSHOT-2026-09-21.md`,
`docs/evidence/DASHBOARD-REFRESH-AFTER-INTEGRATION-2026-10-12.md`,
`docs/evidence/BLING-DAILY-CASH-MOVEMENTS-2026-10-25.md`,
`frontend/src/App.tsx`, `frontend/src/mockupDashboard.tsx`,
`frontend/src/bootstrap.ts`, `app.go`, `cmd/worker/handler.go`,
`internal/dashboard/stream.go`, `internal/integrations/bling/api_sync.go`,
`internal/integrations/bling/daily_sync.go`,
`internal/integrations/bling/oauth_service.go` e buscas nas migrations e
persistências financeiras de produção.

## Atualização automática em tela aberta — 26/10/2026

- O frontend agora atualiza o estado do worker e o snapshot financeiro local a cada 60 segundos, sem navegar para outra tela. Leituras em andamento não se sobrepõem; ao fechar a aplicação, o timer é cancelado.
- A atualização somente consulta o worker/PostgreSQL. Ela não aciona OAuth, não sincroniza o Bling e não muda filtros, layout ou cálculos.
- Teste focal: `frontend/src/App.test.tsx` comprova que o valor muda após uma leitura em segundo plano mantendo a mesma tela aberta. TypeScript, ESLint e os 27 testes deste arquivo passaram.
- Limite: isso propaga registros que já foram normalizados. As páginas do Bling continuam RAW-only; logo, a atualização automática não faz os valores do Bling aparecerem antes da conversão financeira aprovada.

## Correção do estado dos dados — 2026-10-25

A migration `000013` acrescenta `data_stage` ao lote de integração: `RAW_ONLY`
quando a fonte apenas foi lida/arquivada, `NORMALIZED` quando os fatos foram
gravados nas tabelas financeiras e `UNVERIFIED` para lotes antigos ou de
auditoria sem fatos. O valor também aparece na view `integration_sync_logs`.

O adapter da API Bling marca os lotes como `RAW_ONLY`; os importadores que
persistem fatos financeiros marcam seus lotes como `NORMALIZED`. A Visão
Executiva só confirma A Receber e A Pagar quando os respectivos lotes diários
foram normalizados. Recebido no Mês e Pago no Mês dependem de `/caixas`
normalizado, em vez de presumirem que a leitura de contas a receber/pagar
confirma esses movimentos. O horário de sincronização continua informando a
hora real da leitura RAW. As telas e fórmulas permanecem iguais.

`go test -count=1 ./...` e `go vet ./...` passaram. Também passaram a geração
determinística do sqlc, a conferência das migrations, dos limites de arquitetura
Go e do backlog. Foi acrescentado um teste PostgreSQL que confirma que lotes
RAW-only não viram zeros e que cada métrica aguarda seu próprio recurso
normalizado. Ele não executou localmente, pois `MAJUCAU_TEST_DATABASE_URL` não
estava configurada; o banco ativo do protótipo não foi usado. Na CI
A primeira tentativa da CI
[40256582913](https://github.com/guilhermejacyzin/majucau/actions/runs/40256582913)
passou no job PostgreSQL, nos testes, scans, frontend e builds do job Windows.
O smoke final do ciclo de vida do serviço Windows falhou esperando o estado
`Running` (`WORKER_SERVICE_WAIT_FOR_RUNNING`), embora a coleta de diagnóstico
imediatamente posterior tenha observado o serviço em `Running`. Reexecutamos
somente o job Windows; o smoke passou e os quatro jobs do run ficaram verdes.
A falha inicial não se repetiu, mas sua causa exata continua sem confirmação.
A conversão da resposta API Bling RAW para registros financeiros, a validação
com a conta real e as outras famílias de tela seguem pendentes; esta correção
evita apresentar a leitura RAW como zero financeiro.

## Navegação da T01 para os detalhes — 01/11/2026

Os botões “Ver detalhes” e links já desenhados na T01 agora navegam para a tela
existente que explica cada indicador. Caixa abre Fluxo de Caixa; receber e pagar
abrem suas telas; vendas e margem abrem DRE; estoque, produção e compras abrem
seus respectivos módulos. Gráfico, distribuição do saldo e avisos também levam
à origem correspondente.

A navegação só troca de tela. Ela não busca dados, não cria fórmula paralela e
não muda os valores ou o desenho aprovado. Quando houver dados, o módulo de
origem e a T01 continuam usando a base/snapshot do Majucau; ausências continuam
pendentes ou indisponíveis.

## Rechecagem da projeção e da regra da T01 — 02/11/2026

Gisele confirmou novamente que a T01 é resultado das telas de origem: quando
um valor válido aparecer na tela correspondente, a T01 deve mostrar o mesmo
resultado a partir do snapshot compartilhado. A T01 não cria uma fórmula nem
faz uma busca paralela ao Bling para substituir o resultado ausente de outra
tela. A regra já está registrada em `AGENTS.md` e permanece protegida.

**A projeção T02 ainda não tem valores**, por estes motivos técnicos
confirmados:

- A regra financeira exige que o saldo inicial de hoje seja o saldo final
  conciliado de ontem. O núcleo de cálculo D0–D+60 existe, mas não tem uma
  leitura de produção desse saldo inicial nem de movimentos normalizados.
- No contrato OpenAPI oficial do Bling, `GET /contas-contabeis` retorna dados
  cadastrais da conta (`id`, descrição, tipo e alias), não o saldo atual. A
  resposta de `/caixas` lista lançamentos e não traz um campo de saldo final.
  A referência pública também não oferece, na resposta de `/caixas`, um total
  ou marcador de última página que permita provar uma releitura histórica
  completa. Fonte: [Referência oficial da API Bling](https://developer.bling.com.br/referencia).
- O sincronizador ainda guarda `/caixas` como dado bruto; não há movimentos
  normalizados alimentando o motor de projeção. Contas a Receber também está
  `RAW_ONLY`. Contas a Pagar tem normalização, mas a última carga real foi
  revertida por `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`; o valor negativo não foi
  convertido em zero nem publicado.
- A política de atualização da T01 mantém saldo/projeção indisponíveis enquanto
  a T02 não tiver esses resultados. Nenhum valor estimado, conta não conciliada
  ou fechamento antigo será tratado como atual.

Logo, a API respondeu e o Majucau recebeu registros, mas isso não significa
que existam resultados financeiros prontos para as telas. A próxima
dependência para projetar saldo real é uma fonte comprovada do saldo conciliado
de D-1 por conta; a interpretação do saldo negativo em Contas a Pagar também
continua pendente de regra aprovada. Não foi feita inspeção visual do Bling
nesta atualização porque a política salva do navegador bloqueou a observação
da página.

Validação local: 57 testes frontend em 6 arquivos, lint e typecheck passaram.
A CI [#41274539700](https://github.com/guilhermejacyzin/majucau/actions/runs/41274539700)
passou nos quatro jobs, incluindo Windows e PostgreSQL. A conferência manual do
clique não foi concluída porque a captura da janela do protótipo expirou; não
afirmo validação visual.

## Cobertura real dos indicadores da T01 — auditoria de 01/11/2026

Conferi o contrato TypeScript do snapshot, os pontos em que os cards leem esse
snapshot e o leitor Go que consulta o PostgreSQL. O clique de um card apenas
abre a tela de origem; não inicia uma chamada ao Bling.

| Indicador da T01 | Origem/detalhe | Situação encontrada no código |
|---|---|---|
| A Receber | Contas a Receber | Usa `snapshot.receivables`, o mesmo total da tela de detalhe. A consulta inclui B2C projetado do Nuvem Pago e B2B do Bling, conforme o filtro existente. |
| A Pagar | Contas a Pagar | Usa `snapshot.payables`, o mesmo total da tela de detalhe, ligado ao último lote diário normalizado. |
| Saldo atual | Fluxo de Caixa | Ainda não existe saldo de conta no contrato do snapshot. Não há saldo por banco para mostrar. |
| Menor saldo 30/60 dias e gráfico projetado | Fluxo de Caixa | Não há série de saldo projetado no snapshot; o gráfico continua sem dados. |
| Distribuição do saldo | Fluxo de Caixa | O snapshot não traz saldos por conta/meio nem estado por banco. |
| Vendas do mês e margem bruta | DRE | A navegação abre DRE, mas o snapshot executivo não fornece esses indicadores. |
| Estoque/cobertura | Estoque | Não há métrica correspondente no snapshot executivo. |
| Produção prevista | Produção | Não há métrica correspondente no snapshot executivo. |
| Compras previstas | Compras | Não há métrica correspondente no snapshot executivo. |

Os itens sem métrica permanecem como “—”/“Sem dados”; não são preenchidos por
estimativa paralela. O contrato também prevê campos CAPEX opcionais, mas o leitor
atual não os preenche. Os avisos mostrados na T01 ainda são textos configurados
na interface, não alertas gerados pelas métricas do snapshot.

Gisele aprovou mostrar o valor de `A RECEBER` ou `A PAGAR` quando o snapshot vier
com estado `PARTIAL`, acompanhado do aviso “Parcial”. Implementei o aviso no
mesmo texto de apoio já existente (por exemplo, “Parcial · B2C + B2B”); o valor
original não é alterado nem recalculado, e nenhum elemento ou espaço novo foi
adicionado à tela. Valores confirmados mantêm o texto atual. Um teste de
regressão verifica os dois casos.

**Resposta sobre Itaú e Mercado Pago:** hoje o detalhe não lista os saldos desses
bancos. O sistema ainda não tem saldo por conta normalizado para confirmá-los;
não vou afirmar que os dois aparecem. Essa lacuna foi registrada para a próxima
etapa de integração do Fluxo de Caixa. Nenhuma consulta direta extra ao Bling foi
feita.

### Separação das origens de saldo — orientação confirmada em 02/11/2026

O resultado de caixa deve manter cada origem separada:

| Resultado | Fonte prevista | Regra de apresentação quando não há dado válido |
|---|---|---|
| Itaú | Conta/movimentações correspondentes no Bling | Não criar saldo. A falta deste dado não elimina outros saldos disponíveis. |
| Mercado Pago | Conta/movimentações correspondentes no Bling | Não criar saldo. A falta deste dado não elimina outros saldos disponíveis. |
| Nuvem | Arquivo externo autorizado no Drive | Gisele informou que ainda não colocou arquivo nessa pasta. Portanto, não há resultado Nuvem a exibir; não mostrar zero, estimativa nem substituir por outro saldo. |

Gisele esclareceu que uma origem externa sem arquivo não deve produzir um
resultado nem bloquear os dados bancários do Bling. Não é necessário definir o
conteúdo da fonte Nuvem até que um arquivo seja colocado nessa pasta. Esta
orientação não autoriza inferir saldo a partir de recebimentos futuros, misturar
Nuvem com Itaú/Mercado Pago ou alterar telas. Quando houver dados válidos, cada
origem deve conservar seu próprio nome e lineage. A regra financeira de
continuidade permanece: saldo inicial D = saldo final conciliado D-1, e a
projeção só pode ser publicada depois que essa base e as movimentações
necessárias forem comprovadas.

**Situação técnica:** `/caixas` ainda está somente em RAW, sem fatos normalizados,
e o contrato consultado não fornece saldo final pronto nem prova suficiente de
histórico completo. Portanto Itaú e Mercado Pago ainda não estão prontos para
publicação na T02. Nuvem está sem arquivo externo e deve simplesmente ficar fora
do resultado; essa ausência não bloqueia os saldos Bling.

**Implicação para homologação:** os dois totais de contas compartilham dados
comprovados. Os demais indicadores da T01 ainda dependem de suas telas/fontes de
origem serem alimentadas e expostas no snapshot. O código não apresenta ainda um
aviso dinâmico dizendo qual origem está faltando; antes de acrescentar tal aviso
na tela aprovada, será necessário definir com Gisele onde ele cabe sem redesenho.

## Normalização diária de Contas a Pagar — 26/10/2026

A carga diária completa de `contas.pagar` agora converte cada título para
`payables`, preserva respostas da obrigação, contato e categoria em RAW, cria
lineage e registra os títulos associados ao lote normalizado. O resumo e as
linhas da tela T4 filtram contas de API pela associação do último snapshot
completo; a Visão Executiva continua usando o mesmo snapshot. Uma falha em
qualquer detalhe reverte a carga e mantém o retrato anterior. Contas a Receber,
pagamentos, `/caixas`, conciliação, CAPEX e as demais telas ainda não têm esse
conversor. Os testes Go locais e as verificações estáticas passaram; o teste
PostgreSQL está aguardando execução da CI, pois não há URL de banco de testes
configurada nesta máquina. Nenhuma tela, cálculo ou regra financeira foi
alterada. Evidência: `docs/evidence/BLING-PAYABLE-API-NORMALIZATION-2026-10-26.md`.
