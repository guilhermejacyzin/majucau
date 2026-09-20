# Majucau — consolidação, análise e decisões, versão 1

Esta revisão incorpora a consolidação enviada pelo usuário em 13/09/2026 ao planejamento existente. O documento foi lido integralmente. Não houve acesso à pasta compartilhada, consulta à conta real do Bling, importação de dados ou desenvolvimento do produto.

Foram ajustadas 52 tarefas, mantendo os 127 IDs e o progresso: 3 levantamentos concluídos, 2,4% por quantidade e 0% de desenvolvimento. As definições da nova fonte prevalecem nos pontos expressamente respondidos; outros conflitos continuam tratados caso a caso.

## O que fica decidido nesta rodada

- **R-01 — Fontes e estruturas externas:** Bling/API + Google Drive — pasta compartilhada Majucau, preservando 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Sem pasta de Itaú, Mercado Pago, investimentos ou ajustes.
- **R-02 — Permissão e rastreio no Drive:** Somente leitura, sem qualquer mutação na origem. Metadados, hashes, lotes, rejeições e vínculos de reprocessamento ficam dentro do Majucau. Correção não sobrescreve fatos silenciosamente.
- **R-03 — Identificação do depósito:** Identificar no cadastro operacional da API, sem hardcode ou escolha automática de Bloqueado. Sem evidência inequívoca: Dado indisponível no Bling.
- **R-04 — Fonte de investimentos:** Conta Caixa no Bling; caixa.php é referência funcional. Usar API oficial ou exportação estruturada oficial. Scraping proibido; não haverá extrato externo na pasta compartilhada.
- **R-05 — Categorias de investimentos:** Transferência destinada à aplicação identifica aporte; Rendimentos identifica rendimento realizado. Resgates ainda precisam ser mapeados tecnicamente.
- **R-06 — Já Investido:** Aportes de principal − resgates de principal. Rendimentos ficam separados e não são novos aportes. Falta provar os componentes no Bling.
- **R-07 — Posição comprovada:** Exibir posição e data somente se disponíveis na fonte; caso contrário mostrar apenas componentes comprováveis. Não presumir rendimento reinvestido.
- **R-08 — Itaú e valor executável:** Itaú é a conta principal e fornece o saldo transferível, via Bling. Valor Executável Hoje = MIN(Máximo Aplicável Financeiramente; Saldo Transferível do Itaú).
- **R-09 — Mercado Pago:** Movimentos via Bling; pode compor caixa consolidado, mas não aumenta automaticamente o executável para aplicações que saem do Itaú.
- **R-10 — Nuvem Pago e liquidação:** Fonte 02_nuvem_pago/recebimentos_futuros, rotina D+1. Separar venda, direito futuro e movimento financeiro que comprova liquidação.
- **R-11 — Perdas operacionais:** Fonte 01_perdas_operacionais representa eventos de perda, nunca percentual industrial, scrap, yield ou fator automático de matéria-prima.
- **R-12 — Nuvem Envio:** Fonte 03_nuvem_envio; preservar custo efetivo, frete cobrado, diferença/subsídio, pedido, transportadora, etiqueta e data quando disponíveis, sem compensação silenciosa.
- **R-13 — Base do forecast:** Executar backtest com histórico confiável de 2025 em diante. Não presumir cobertura de 24 meses nem métricas calculadas.
- **R-14 — Refinamento do cacau:** Construir posteriormente fontes externas e regras de comparação aprovadas; isso não amplia as três estruturas do Drive nem altera automaticamente o escopo da primeira produção.

## Frentes que ainda exigem evidência ou refinamento

A lista abaixo registra as dez frentes expressamente mantidas em aberto na consolidação. É trabalho futuro da equipe, com validação da Majucau quando envolver regra financeira ou operacional. Não é um novo questionário para o usuário.

| ID | Frente | O que precisa ser comprovado | Tarefas |
|---|---|---|---|
| HT-01 | Resgates de principal | Demonstrar categoria, sinal, origem/destino, aplicação e eventos de resgate parcial e total no Bling/Conta Caixa. Tabela de campos reais e exemplos conciliados; falta de identificação não equivale a zero. | BK-040, BK-071, BK-072 |
| HT-02 | Posição atual da aplicação | Demonstrar se o Bling fornece posição real, data, instituição e aplicação ou somente movimentos. Exibir somente componentes comprováveis; não inventar posição nem reinvestimento. | BK-040, BK-071, BK-074 |
| HT-03 | Campos do depósito | Homologar identificação operacional, ID, descrição, saldo físico e estados reservado/disponível se existirem. Identificação inequívoca documentada ou Dado indisponível no Bling; sem hardcode. | BK-040, BK-048, BK-077 |
| HT-04 | Suficiência do balancete | Verificar saldos iniciais, débitos, créditos e contrapartidas reais e validar suficiência com o financeiro. Matriz de cobertura contábil; não apresentar fluxo financeiro como balancete completo. | BK-007, BK-040, BK-068 |
| HT-05 | Custo histórico | Homologar endpoint/campo e provar diferença entre custo da venda antiga e custo atual. Exemplos temporais verificáveis; se só existir custo atual, declarar ausência de custo histórico. | BK-040, BK-063 |
| HT-06 | Mapa da DRE | Finalizar origem → conta Bling → classificação gerencial → linha da DRE, incluindo competência e ausências. Mapa versionado com evidências técnicas e validação financeira; desconhecido permanece pendente. | BK-007, BK-040, BK-064, BK-066 |
| HT-07 | BOM/ficha técnica | Homologar componente, quantidade e unidade reais por produto no Bling, além dos fatores industriais que existirem. Composição demonstrada; arquivo de perdas não fornece scrap/yield. | BK-040, BK-048, BK-094 |
| HT-08 | Chaves da conciliação | Homologar chaves dos arquivos Nuvem e dos recebimentos do Bling, preservando D+1 e comprovação de liquidação. Amostras de vínculo e ambiguidade; nenhuma tripla contagem de venda/recebível/dinheiro. | BK-010, BK-040, BK-046, BK-053 |
| HT-09 | Backtest de forecast | Executar validação temporal com histórico confiável de 2025 em diante. Janelas, cobertura, fórmulas, métricas e limitações reais registradas; ainda não executado. | BK-009, BK-087, BK-090 |
| HT-10 | Fontes e comparação de cacau | Refinar posteriormente as fontes e bases comparáveis antes da construção dos sinais. Fontes e regras aprovadas; nenhuma pasta ou feed ativo presumido. | BK-011, BK-085 |

## Implicações técnicas para o backlog

- O adaptador do Drive faz leitura; área de processamento, evidências, logs, lotes, rejeições e versões ficam no Majucau. Backup e exportações precisam de destino próprio, fora das três estruturas de fontes.
- A rota do Google Drive está definida, mas IDs da pasta, autenticação e contratos dos arquivos continuam homologações técnicas. Não é necessário instalar ou conectar um plugin para registrar essas tarefas.
- Investimentos usam Conta Caixa via API oficial ou exportação estruturada oficial a homologar. A alternativa de arquivo oficial não autoriza criar pasta de investimento no Drive.
- Transferência só é aporte quando destinada à aplicação. Rendimentos realizados não são principal. Resgate não identificado não pode ser tratado como zero.
- Posição, principal líquido e rendimento são indicadores distintos. Se faltar apenas a posição total, mostrar componentes independentes comprovados; bloquear somente resultados que dependam de informação ausente.
- Caixa consolidado pode incluir Mercado Pago. O limite operacional do executável vem do Itaú, sem transformar recebíveis futuros em saldo bancário.
- D+1 foi confirmado para Nuvem/Nuvem Pago. Horário e calendário ainda precisam ser detalhados; a periodicidade não se aplica automaticamente ao Bling e demais arquivos.
- Perdas operacionais não alimentam rendimento industrial. BOM e fatores industriais requerem campos e regras próprios.
- A referência visual de 24 meses de histórico não comprova essa disponibilidade. Usar dados confiáveis de 2025 em diante e declarar insuficiência quando necessário.
- Uma insuficiência contábil do Bling não autoriza fabricar abertura/contrapartidas nem substituir balancete por fluxo de caixa. Registrar o impedimento e submetê-lo à validação pertinente.
- A construção posterior das fontes de cacau permanece no planejamento; a frase não foi interpretada como autorização para excluir o módulo da primeira produção.

## Pontos anteriores que esta rodada não respondeu

Continuam no registro, entre outros: saldo específico dentro da semana crítica, corte temporal do lucro, reserva, classificação/competência financeira, pedido de compra com cobertura parcial, regras de metas, conteúdo dos módulos sem referência completa, permissões da diretoria, acesso remoto, capacidade/rotina do Windows, recuperação, responsáveis e prazo. Esta consolidação de fontes não fecha automaticamente esses assuntos.

## Tarefas alteradas

| ID | Tarefa atual | Campos revisados |
|---|---|---|
| BK-000 | Confirmar a base de escopo e as decisões da Majucau | steps |
| BK-001 | Inventariar os materiais e suas versões | steps |
| BK-006 | Fechar os critérios da calculadora de aplicação | action, steps, accept |
| BK-007 | Definir competência, fontes e classificações financeiras | steps, accept |
| BK-008 | Definir regras pendentes de estoque, produção e compras | steps |
| BK-009 | Homologar regras de estudo, metas e qualidade de projeção | steps, accept |
| BK-010 | Homologar contratos do Bling e das três estruturas do Google Drive | title, action, steps, accept, owner |
| BK-011 | Definir fontes e critérios dos sinais do cacau | steps, accept |
| BK-016 | Modelar os domínios e o vocabulário do negócio | steps |
| BK-024 | Planejar segurança e proteção dos dados | steps, accept |
| BK-037 | Implementar T16 Parâmetros e fluxos autorizados de edição | steps, accept |
| BK-038 | Implementar proveniência e snapshots de fatos | steps, accept |
| BK-040 | Validar acesso e cobertura dos dados do Bling | steps, accept |
| BK-041 | Implementar o adaptador de leitura do Bling | steps, accept |
| BK-042 | Implementar leitura das três estruturas autorizadas do Google Drive | title, action, goal, steps, accept |
| BK-043 | Implementar validação e interpretação dos arquivos | steps, accept |
| BK-044 | Deduplicar arquivos e fatos de negócio | steps, accept |
| BK-046 | Ingerir recebíveis e eventos do Nuvem Pago | action, steps, accept |
| BK-047 | Ingerir obrigações e movimentos financeiros | action, steps |
| BK-048 | Ingerir estoque, produção, compras e notas de entrada | steps, accept |
| BK-049 | Ingerir etiquetas, faturas e frete da nota fiscal | steps, accept |
| BK-050 | Ingerir eventos de perdas operacionais | title, action, goal, steps, accept |
| BK-051 | Controlar atualização e falha por fonte | steps, accept |
| BK-052 | Implementar T15 Importações e Integrações | steps, accept |
| BK-053 | Implementar conciliação por evidência e unicidade | steps, accept |
| BK-056 | Implementar regras e leitura de Contas a Receber | steps |
| BK-060 | Implementar fluxo de caixa realizado, previsto e projetado | steps, accept, deps |
| BK-062 | Implementar T02 Fluxo de Caixa | steps |
| BK-063 | Preservar o custo Bling válido para cada venda | steps, accept |
| BK-064 | Implementar classificação e competência financeira | steps, accept |
| BK-065 | Implementar perdas e recuperações sem dupla contagem | steps, accept |
| BK-068 | Implementar o motor do balancete | action, steps, accept |
| BK-071 | Implementar movimentos de investimentos e posição comprovada no Bling | title, action, goal, steps, accept, deps |
| BK-072 | Implementar o cálculo de capacidade de aplicação | steps, accept |
| BK-073 | Implementar T08 Calculadora de Aplicação | steps, accept |
| BK-074 | Implementar a consulta de Aplicações | steps, accept |
| BK-076 | Implementar indicadores de logística no local aprovado | steps |
| BK-077 | Implementar classificação e posição de estoque | steps, accept |
| BK-085 | Implementar sinais externos do cacau | steps, accept |
| BK-087 | Preparar histórico de vendas por SKU para projeções | action, steps, accept |
| BK-090 | Validar WAPE, Bias e MASE sem vazamento temporal | steps, accept |
| BK-094 | Implementar necessidade de insumos e OP Projetada | steps, accept |
| BK-098 | Implementar registro único de pendências | steps |
| BK-105 | Rastrear cada requisito até tarefa e teste | steps |
| BK-106 | Validar regras de domínio e casos de fronteira | steps, accept |
| BK-107 | Validar integrações, concorrência e reinício | steps, accept |
| BK-111 | Validar controles de segurança antes da produção | steps, accept |
| BK-114 | Definir a topologia final da instalação local | steps, accept |
| BK-118 | Configurar backup de banco, arquivos e configuração | steps, accept |
| BK-119 | Testar restauração e recuperação do ambiente | steps |
| BK-121 | Preparar e reconciliar a carga inicial | steps |
| BK-123 | Documentar e treinar a rotina da Majucau | steps |

## Registro completo de decisões e verificações

88 registros, dos quais 60 permanecem pendentes ou parcialmente esclarecidos. Registros L/Q e HT podem apontar para a mesma frente; a contagem não representa perguntas distintas nem quantidade de falhas do produto. A natureza de cada registro separa definição recebida, decisão de negócio e verificação técnica.

### D-01 — Empresa atendida

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Somente Majucau. A resposta anterior sobre várias empresas foi corrigida pelo usuário.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-02 — Usuários

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Três pessoas da diretoria, cada uma em seu computador.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-03 — Hospedagem

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Local, em computador/servidor Windows existente. Demais características ainda desconhecidas.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-04 — Primeira produção

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Todos os módulos confirmados do material na primeira entrada em produção; etapas internas são permitidas para organizar o trabalho.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-05 — Conflitos entre fontes

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Analisar caso a caso com o usuário. Não existe precedência global da planilha sobre imagens.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-06 — Tecnologias

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Monólito com DDD e arquitetura hexagonal; backend Go; frontend JavaScript com React.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-07 — Balanço Patrimonial

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Fora da primeira versão; retirar o item dos menus por decisão explícita do usuário.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### D-08 — Natureza desta entrega

**Situação:** Respondida. **Natureza:** Definição recebida.

Decisão recebida

**Registro atual:** Somente análise e backlog. Não desenvolver o produto nesta etapa.

**Impacto / evidência necessária:** Orienta o planejamento.

**Fonte:** Conversa atual

### C-01 — Pagamento diferente do título

**Situação:** Respondida. **Natureza:** Definição recebida.

Qual o tratamento para R$1.000 com movimento de R$600?

**Registro atual:** Marcar divergência e manter o título sem baixa até a análise. Casos complementares serão detalhados em BK-005.

**Impacto / evidência necessária:** Define baixa, saldo e conciliação.

**Fonte:** C_Conflitos!A3:L3

**Tarefas:** BK-005

### C-02 — Semana crítica

**Situação:** Respondida. **Natureza:** Definição recebida.

Maior despesa ou menor saldo?

**Registro atual:** A semana com o maior total de despesas. A posição de caixa dentro dela ainda precisa ser definida.

**Impacto / evidência necessária:** Define seleção da semana no fluxo e na calculadora.

**Fonte:** C_Conflitos!A4:L4

### C-03 — Lucro elegível

**Situação:** Respondida. **Natureza:** Definição recebida.

Qual DRE fornece o lucro?

**Registro atual:** Lucro da DRE Majucau. Já Investido foi definido como aportes de principal menos resgates de principal; fonte dos resgates ainda requer homologação. Corte temporal do lucro continua sem nova resposta.

**Impacto / evidência necessária:** Define fonte do limite da calculadora.

**Fonte:** C_Conflitos!A5:L5; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

### L-01 — Gate formal

**Situação:** Planejamento autorizado. **Natureza:** Refinamento de negócio/entrega.

Aprovação formal Majucau após resolver conflitos.

**Registro atual:** Usuário pediu somente análise e backlog. Implementação será trabalho futuro; o bloqueio textual do documento não impede planejar agora.

**Impacto / evidência necessária:** Bloqueia qualquer implementação definitiva.

**Fonte:** D_Lacunas!A3:G3

### L-02 — Sistema atual

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Acesso read-only ao projeto e dados técnicos.

**Registro atual:** GitHub inspecionado e vazio. Falta confirmar se há código/dados técnicos fora do repositório.

**Impacto / evidência necessária:** Impossível classificar JÁ CONFORME/PRECISA ALTERAÇÃO por código.

**Fonte:** D_Lacunas!A4:G4

### L-03 — CMV histórico

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Mapear campo/snapshot ou decidir tratamento.

**Registro atual:** Homologar custo histórico versus atual no retorno real do Bling. Ver HT-05; o texto não comprova disponibilidade.

**Impacto / evidência necessária:** Margem/DRE histórica não podem ser exatas.

**Fonte:** D_Lacunas!A5:G5; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-007

### L-04 — Perdas — competência

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Definir occurrence/confirmation/baixa/outra data.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Resultado por período pode mudar.

**Fonte:** D_Lacunas!A6:G6

**Tarefas:** BK-007

### L-05 — Juros Nuvem

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Aprovar conta/classificação e competência.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** DRE por linha diverge.

**Fonte:** D_Lacunas!A7:G7

**Tarefas:** BK-007

### L-06 — Tarifa fixa Pricing

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Definir rateio ou excluir componente por SKU com regra explícita.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Preço sugerido por SKU incompleto.

**Fonte:** D_Lacunas!A8:G8

**Tarefas:** BK-075

### L-07 — Caixa da semana crítica

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Definir campo real do fluxo.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Máximo Aplicável indefinido mesmo após escolher semana.

**Fonte:** D_Lacunas!A9:G9

**Tarefas:** BK-006

### L-08 — Balancete — saldos iniciais

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Aprovar fonte/campos ou manter UNAVAILABLE.

**Registro atual:** Verificar no Bling abertura, débitos, créditos e contrapartidas. Sem base suficiente, não produzir balancete fictício. Ver HT-04.

**Impacto / evidência necessária:** Balancete não fecha de forma auditável.

**Fonte:** D_Lacunas!A10:G10; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-007

### L-09 — DRE — fontes complementares

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Aprovar fontes complementares e mappings.

**Registro atual:** Fontes ativas são Bling e três estruturas específicas do Drive. Finalizar cobertura de cada linha e mapa da DRE em HT-06; não adicionar planilha complementar por inferência.

**Impacto / evidência necessária:** DRE incompleta.

**Fonte:** D_Lacunas!A11:G11; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-007

### L-10 — UI Pricing

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Aprovar local em tela existente; sem nova rota automática.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Risco de violar Design Lock.

**Fonte:** D_Lacunas!A12:G12

**Tarefas:** BK-075

### L-11 — Arquivos Nuvem

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Fornecer arquivos-modelo e política operacional.

**Registro atual:** Fonte Nuvem definida em 02_nuvem_pago/recebimentos_futuros, somente leitura, rotina D+1. Faltam amostras, chaves e contrato real. Ver HT-08.

**Impacto / evidência necessária:** Bloqueia importador confiável e status CONFIRMED.

**Fonte:** D_Lacunas!A13:G13; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-010

### L-12 — Ciclo de vida de arquivo

**Situação:** Respondida. **Natureza:** Definição recebida.

Definir política: origem imutável + cópia interna ou movimento autorizado.

**Registro atual:** Origem somente de leitura. Proibido criar, mover, renomear, excluir, sobrescrever ou alterar arquivos. Controle, hash, lotes, rejeições e versões ficam internamente; arquivo corrigido não sobrescreve fatos silenciosamente.

**Impacto / evidência necessária:** Risco de permissão/destruição de evidência.

**Fonte:** D_Lacunas!A14:G14; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-010

### L-13 — Fretes/taxas na DRE

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Aprovar mapping/competência por tipo.

**Registro atual:** Nuvem Envio usa 03_nuvem_envio. Custo efetivo, frete cobrado e subsídio ficam separados. Classificação e competência da DRE ainda exigem HT-06.

**Impacto / evidência necessária:** DRE final não pode consumir todos os fatos.

**Fonte:** D_Lacunas!A15:G15; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-007

### L-14 — Recuperações

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Aprovar reconhecimento por tipo de evento.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Caixa/DRE podem antecipar ou duplicar recuperação.

**Fonte:** D_Lacunas!A16:G16

**Tarefas:** BK-007

### L-15 — Custo de carregar estoque

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Aprovar taxa/parametrização e armazenagem incremental válida.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Antecipação de compra fica sem custo econômico completo.

**Fonte:** D_Lacunas!A17:G17

**Tarefas:** BK-084

### L-16 — Mockup Logística

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Aprovar bloco/tela dentro do Design Lock.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** UI não pode ser implementada sem risco de redesign indevido.

**Fonte:** D_Lacunas!A18:G18

**Tarefas:** BK-076

### L-17 — Compras — quantidade recomendada

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Homologar fórmula, janela histórica, mínimos/múltiplos e critérios por insumo.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Sem fórmula aprovada, quantidade sugerida pode superdimensionar ou fragmentar compras.

**Fonte:** D_Lacunas!A19:G19

**Tarefas:** BK-008

### L-18 — Compras — prioridade e prazo histórico

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Homologar janela histórica, estatística e thresholds de prioridade.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Prioridade pode ficar arbitrária ou invertida.

**Fonte:** D_Lacunas!A20:G20

**Tarefas:** BK-008

### L-19 — Cacau — sinais externos

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Homologar fontes, thresholds, freshness, origem geográfica e comparabilidade de preço.

**Registro atual:** Construir fontes e comparações posteriormente, após refinamento. Não criar pasta adicional ou ativar fonte presumida. Não houve decisão explícita para retirar Cacau do escopo de produção.

**Impacto / evidência necessária:** Risco de cacau pode sinalizar falso positivo/negativo ou comparar bases incompatíveis.

**Fonte:** D_Lacunas!A21:G21; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-011

### L-20 — Forecast — thresholds de qualidade

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Homologar limites por métrica e, quando aplicável, por sazonalidade/SKU.

**Registro atual:** Backtest deve usar histórico confiável de 2025 em diante. Métricas, limites e suficiência dependem de execução e validação; não há resultado de backtest nesta etapa.

**Impacto / evidência necessária:** Sem threshold não se deve aprovar automaticamente uma versão de forecast como 'boa'.

**Fonte:** D_Lacunas!A22:G22; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-009

### L-21 — Meta de faturamento — distribuição

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Homologar preço realizado de referência, regra de distribuição e tratamento de overrides para cenário mista.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Distribuição mal definida pode transformar meta financeira em quantidades erradas.

**Fonte:** D_Lacunas!A23:G23

**Tarefas:** BK-009

### L-22 — Telas 13–15 — exportação

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Definir por tela se exporta, formato, colunas, filtros/contexto e permissões; até lá não implementar.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Implementar por inferência criaria artefato não aprovado e possível vazamento de dados técnicos.

**Fonte:** D_Lacunas!A24:G24

**Tarefas:** BK-012

### L-23 — Tela 13 — resolução por tipo

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Mapear tipos de pendência para fluxo autorizado, permissões, evidências e transição de status.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Fechar genericamente pode mascarar bloqueio ou divergência sem corrigir causa.

**Fonte:** D_Lacunas!A25:G25

**Tarefas:** BK-012

### L-24 — Tela 15 — freshness por fonte

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Homologar/configurar freshness, retry e timeout por Bling e por cada fonte externa ativa.

**Registro atual:** D+1 é rotina da Nuvem/Nuvem Pago. Falta documentar calendário, horário e política das demais fontes; não estender esse prazo ao conjunto todo.

**Impacto / evidência necessária:** Sem política por fonte, status STALE/OK pode ser arbitrário.

**Fonte:** D_Lacunas!A26:G26; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-010

### L-25 — Tela 16 — governança de edição de parâmetros

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Definir, somente quando necessário, quais parâmetros são editáveis, por quem, fluxo de aprovação, versionamento/vigência e rollback.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Transformar a central em editor genérico pode alterar cálculo crítico sem controle.

**Fonte:** D_Lacunas!A27:G27

**Tarefas:** BK-013

### L-26 — Tela 17 — identidade, ciclo de vida e matriz de permissões

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Homologar matriz de permissões e arquitetura de identidade/convites/sessão antes de implementar mutações.

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Implementação por inferência pode criar escalada de privilégio ou autenticação insegura.

**Fonte:** D_Lacunas!A28:G28

**Tarefas:** BK-013

### Q-01 — Referência visual

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Qual logo e qual organização de menu devemos adotar quando as imagens diferem? Posso corrigir textos, números e seleção de menu mantendo a estrutura aprovada?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Fixa o padrão visual sem resolver diferenças silenciosamente.

**Fonte:** T12/T13 versus demais imagens

**Tarefas:** BK-004, BK-021

### Q-02 — Módulos sem especificação completa

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Quais referências completam T01 Visão Executiva, T11 Compras, Aplicações, Pricing, Logística, Produtos, Simuladores e Relatórios? Os itens citados nos menus são módulos próprios ou atalhos?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Esses itens permanecem no inventário até decisão; nenhuma rota será criada com comportamento inventado.

**Fonte:** Menus; A_Matriz_Mestra

**Tarefas:** BK-004, BK-021, BK-074, BK-102, BK-103

### Q-03 — Máquina Windows

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

O Windows que hospedará o sistema é este computador ou outro? Qual é a edição e quem pode verificar CPU, RAM, disco e suspensão automática?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Permite dimensionar instalação e processamento.

**Fonte:** Resposta do usuário: Windows; demais dados desconhecidos

**Tarefas:** BK-014

### Q-04 — Locais de acesso

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Os três diretores acessarão apenas dentro da empresa ou também de casa e de outras redes?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Define rede, endereço e eventual acesso remoto.

**Fonte:** Uso em três computadores confirmado

**Tarefas:** BK-014

### Q-05 — Disponibilidade local

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

O servidor ficará ligado continuamente? Como o sistema deve funcionar durante falta de energia ou de internet?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Fontes externas exigem internet; posição anterior deve ter data e aviso.

**Fonte:** Hospedagem local

**Tarefas:** BK-014

### Q-06 — Contas e fontes reais

**Situação:** Respondida. **Natureza:** Definição recebida.

Quais contas bancárias, depósitos, canais e fontes já são usados: Bling, Itaú, Mercado Pago, Nuvem Pago, Nuvem Envio, CDI Plus, perdas, ajustes e cacau?

**Registro atual:** Bling/API concentra operacional/financeiro, Itaú principal, Mercado Pago e Conta Caixa/investimentos; Drive contém apenas perdas, Nuvem Pago e Nuvem Envio. Cacau continua refinamento posterior. Identificadores e capacidade real do Bling são verificações técnicas.

**Impacto / evidência necessária:** O contrato menciona fontes além das quatro ilustradas.

**Fonte:** H_Contrato_Dados; M-008; T15; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-010

### Q-07 — Estoque e pedido aplicável

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Qual deve ser o tratamento de pedido aberto insuficiente ou parcialmente recebido para recomendar nova compra? Identificação técnica do depósito está em HT-03.

**Registro atual:** Depósito deve vir inequivocamente do cadastro operacional da API, sem nome/ID presumido; ausência = Dado indisponível no Bling. Pedido insuficiente/parcial continua decisão operacional pendente.

**Impacto / evidência necessária:** Define a regra compartilhada de Estoque e Compras.

**Fonte:** T09; Q_CEI_Tela_11; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-008

### Q-08 — Situação das metas

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Quando um SKU com produção abaixo de 100% deve aparecer Em andamento ou Abaixo da meta? Qual data e situação da OP contam como concluídas?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** O mockup não contém regra suficiente para classificar as linhas.

**Fonte:** T10

**Tarefas:** BK-008, BK-079

### Q-09 — Uso dos demonstrativos

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

A DRE e o balancete serão usados para gestão interna ou precisam seguir também algum processo formal da contabilidade? Quem valida contas, competência, fechamento e correções?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Define fontes, rigor do aceite e participação da contabilidade.

**Fonte:** T06/T07

**Tarefas:** BK-007

### Q-10 — Período e limite de aplicação

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Confirmar corte temporal do lucro elegível e reserva mínima. A composição financeira de Já Investido já foi definida.

**Registro atual:** Já Investido = aportes de principal − resgates de principal, sem rendimentos. Itaú limita o executável. Permanecem corte temporal do lucro e reserva; resgates reais e posição são HT-01/02.

**Impacto / evidência necessária:** Fonte da DRE já decidida; estes componentes ainda mudam o resultado.

**Fonte:** M-030..M-034; T08; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-006, BK-071

### Q-11 — Vendas para produção projetada

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Como a previsão de venda vira quantidade a produzir, sem descontar o estoque pronto: existe lote mínimo, rendimento, perda ou antecipação de fabricação?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** A imagem mostra vendas e produção diferentes sem explicar a transformação.

**Fonte:** M-047; T12

**Tarefas:** BK-009, BK-092

### Q-12 — Relação entre metas

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

As metas de produção da T10 são independentes das metas do cenário da T12? Quem pode editar e como ficam mudanças em meses anteriores?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Evita duas metas conflitantes ou sobrescrita de histórico.

**Fonte:** T10/T12/T16

**Tarefas:** BK-009, BK-079

### Q-13 — Ações visíveis e exportação

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

O que deve acontecer em Ajuda, Atualizar, Ver detalhes e Exportar em cada tela? A atualização apenas consulta a posição ou também solicita sincronização?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Os mockups mostram controles sem todos os fluxos detalhados.

**Fonte:** T02–T07; T13–T15; L-22

**Tarefas:** BK-012

### Q-14 — Permissões da diretoria

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Os três diretores terão as mesmas permissões? Quem administra usuários, aprova metas, edita parâmetros e reprocessa arquivos?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Quatro perfis conceituais não exigem criar contas fictícias.

**Fonte:** T17; L-26

**Tarefas:** BK-013

### Q-15 — Recuperação

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Quanto tempo a Majucau pode ficar sem o sistema e quanto dado pode perder após uma falha?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Escolha de backup e recuperação depende desses limites.

**Fonte:** Requisito de produção local

**Tarefas:** BK-014

### Q-16 — Backup e responsável

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Há outro disco ou equipamento para backup? Quem cuidará de atualizações, cópias e restauração?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** A única cópia não deve depender do mesmo disco do servidor.

**Fonte:** Requisito de produção local

**Tarefas:** BK-014

### Q-17 — Prazo e equipe

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Existe data desejada para produção e quem fará desenvolvimento, validação financeira e suporte?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Ainda não há base para estimar datas e duração.

**Fonte:** Pergunta inicial respondida apenas quanto à hospedagem

**Tarefas:** BK-015

### Q-18 — Dispositivos e linguagem

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Qual o menor tamanho de tela usado? O sistema precisa atender celular? Podemos exibir status em português simples mantendo códigos técnicos internamente?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Define comportamento responsivo e ajustes de acessibilidade.

**Fonte:** Gaps UX em todas as imagens

**Tarefas:** BK-021, BK-022

### Q-19 — Conciliação detalhada

**Situação:** Parcialmente esclarecida. **Natureza:** Refinamento de negócio/entrega.

Sem ID Bling, quais critérios autorizam o vínculo por nome e valor: datas, tolerância de centavos, parcelas e mais de um movimento? Quem confirma ou desfaz manualmente?

**Registro atual:** Primeiro homologar chaves dos arquivos Nuvem e Bling em HT-08. A rotina não confunde direito futuro com liquidação. Casos ambíguos e decisões manuais continuam sem nova definição.

**Impacto / evidência necessária:** A regra visual Nome+Valor não define todas as condições de unicidade.

**Fonte:** T05; M-008; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-053

### Q-20 — Recebimentos e status

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Quais estados, datas e bases definem total a receber, recebido no mês, atraso e percentual de conciliação? Há meios além dos três mostrados?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Cartões de meios e percentuais podem representar universos diferentes.

**Fonte:** T03/T05

**Tarefas:** BK-056

### Q-21 — Calendário e caixa da semana

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Qual saldo da semana de maior despesa será usado: em que dia/horário e com quais contas? Como contar semanas, horizontes de 30/60 dias e desempates?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** C-02 foi respondida, mas L-07 e os limites de data permanecem.

**Fonte:** T02/T08; L-07

**Tarefas:** BK-006, BK-060

### Q-22 — Gráfico da DRE

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

O gráfico deve explicar participação de despesas, formação do resultado ou outra comparação? Qual base deve usar para os percentuais?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Receita, despesas e resultado são valores sobrepostos na rosca atual.

**Fonte:** T06

**Tarefas:** BK-067

### Q-23 — Saldos do balancete

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Como mostrar saldo devedor/credor e qual universo usar na composição por grupo? O zero do cartão significa diferença entre movimentos ou saldo final conhecido?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Impede interpretar falta de abertura como saldo zero.

**Fonte:** T07; L-08

**Tarefas:** BK-069

### Q-24 — Ficha técnica de insumos

**Situação:** Parcialmente esclarecida. **Natureza:** Verificação técnica e validação operacional.

Homologar a BOM real do Bling e identificar eventuais fatores industriais próprios, sem usar a planilha de perdas operacionais.

**Registro atual:** BOM/ficha técnica deve ser homologada no Bling: componente, quantidade e unidade. Perdas operacionais não fornecem scrap/yield. Fatores industriais e vigência só podem ser usados se comprovados e aprovados.

**Impacto / evidência necessária:** Necessária para a aba de insumos e OP Projetada.

**Fonte:** H_Contrato_Dados!A12:H12; T12; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026

**Tarefas:** BK-094

### Q-25 — Resumo financeiro projetado

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Quais linhas deverão aparecer na aba Resumo financeiro da T12 e como devem se relacionar com custos, preço e cenários?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** A aba é citada, mas seu conteúdo completo não foi enviado.

**Fonte:** T12

**Tarefas:** BK-095

### Q-26 — Alertas e monitoramento

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Os avisos ficam somente dentro do sistema ou devem chegar por algum canal? Para estoque, vale a prioridade Atenção do contrato ou Crítico do exemplo visual?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** Define comunicação e resolve discrepância concreta de severidade.

**Fonte:** T14; T_CEI_Tela_14

**Tarefas:** BK-120

### Q-27 — Indicadores de prazo

**Situação:** Em aberto. **Natureza:** Refinamento de negócio/entrega.

Quais fórmulas, fontes e períodos definem prazo médio de recebimento, pagamento, estoque e ciclo financeiro? NCG será um indicador necessário nesta versão?

**Registro atual:** Sem resposta nova nesta rodada; manter o refinamento vinculado.

**Impacto / evidência necessária:** O mockup combina três prazos em 28 dias sem detalhar as bases e exibe ciclo de 41 dias.

**Fonte:** T08; B_Intersecoes!A12:I12

**Tarefas:** BK-070

### R-01 — Fontes e estruturas externas

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Bling/API + Google Drive — pasta compartilhada Majucau, preservando 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Sem pasta de Itaú, Mercado Pago, investimentos ou ajustes.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §1

**Tarefas:** BK-010, BK-042, BK-052

### R-02 — Permissão e rastreio no Drive

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Somente leitura, sem qualquer mutação na origem. Metadados, hashes, lotes, rejeições e vínculos de reprocessamento ficam dentro do Majucau. Correção não sobrescreve fatos silenciosamente.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §2

**Tarefas:** BK-010, BK-024, BK-038, BK-042, BK-044, BK-118

### R-03 — Identificação do depósito

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Identificar no cadastro operacional da API, sem hardcode ou escolha automática de Bloqueado. Sem evidência inequívoca: Dado indisponível no Bling.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §3

**Tarefas:** BK-040, BK-048, BK-077

### R-04 — Fonte de investimentos

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Conta Caixa no Bling; caixa.php é referência funcional. Usar API oficial ou exportação estruturada oficial. Scraping proibido; não haverá extrato externo na pasta compartilhada.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §4

**Tarefas:** BK-040, BK-041, BK-071, BK-074

### R-05 — Categorias de investimentos

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Transferência destinada à aplicação identifica aporte; Rendimentos identifica rendimento realizado. Resgates ainda precisam ser mapeados tecnicamente.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §5

**Tarefas:** BK-064, BK-071

### R-06 — Já Investido

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Aportes de principal − resgates de principal. Rendimentos ficam separados e não são novos aportes. Falta provar os componentes no Bling.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §6

**Tarefas:** BK-006, BK-071, BK-072

### R-07 — Posição comprovada

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Exibir posição e data somente se disponíveis na fonte; caso contrário mostrar apenas componentes comprováveis. Não presumir rendimento reinvestido.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §7

**Tarefas:** BK-071, BK-073, BK-074

### R-08 — Itaú e valor executável

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Itaú é a conta principal e fornece o saldo transferível, via Bling. Valor Executável Hoje = MIN(Máximo Aplicável Financeiramente; Saldo Transferível do Itaú).

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §8

**Tarefas:** BK-006, BK-047, BK-060, BK-072

### R-09 — Mercado Pago

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Movimentos via Bling; pode compor caixa consolidado, mas não aumenta automaticamente o executável para aplicações que saem do Itaú.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §9

**Tarefas:** BK-047, BK-060, BK-072

### R-10 — Nuvem Pago e liquidação

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Fonte 02_nuvem_pago/recebimentos_futuros, rotina D+1. Separar venda, direito futuro e movimento financeiro que comprova liquidação.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §10

**Tarefas:** BK-010, BK-046, BK-051, BK-053

### R-11 — Perdas operacionais

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Fonte 01_perdas_operacionais representa eventos de perda, nunca percentual industrial, scrap, yield ou fator automático de matéria-prima.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §11

**Tarefas:** BK-050, BK-065, BK-094

### R-12 — Nuvem Envio

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Fonte 03_nuvem_envio; preservar custo efetivo, frete cobrado, diferença/subsídio, pedido, transportadora, etiqueta e data quando disponíveis, sem compensação silenciosa.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §12

**Tarefas:** BK-049, BK-064, BK-076

### R-13 — Base do forecast

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Executar backtest com histórico confiável de 2025 em diante. Não presumir cobertura de 24 meses nem métricas calculadas.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 9

**Tarefas:** BK-009, BK-087, BK-090

### R-14 — Refinamento do cacau

**Situação:** Respondida. **Natureza:** Definição recebida.

Definição recebida nesta consolidação.

**Registro atual:** Construir posteriormente fontes externas e regras de comparação aprovadas; isso não amplia as três estruturas do Drive nem altera automaticamente o escopo da primeira produção.

**Impacto / evidência necessária:** Aplicar nas tarefas vinculadas; homologação técnica continua separada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 10

**Tarefas:** BK-011, BK-085

### HT-01 — Resgates de principal

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Demonstrar categoria, sinal, origem/destino, aplicação e eventos de resgate parcial e total no Bling/Conta Caixa.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Tabela de campos reais e exemplos conciliados; falta de identificação não equivale a zero.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 1

**Tarefas:** BK-040, BK-071, BK-072

### HT-02 — Posição atual da aplicação

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Demonstrar se o Bling fornece posição real, data, instituição e aplicação ou somente movimentos.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Exibir somente componentes comprováveis; não inventar posição nem reinvestimento.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 2

**Tarefas:** BK-040, BK-071, BK-074

### HT-03 — Campos do depósito

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Homologar identificação operacional, ID, descrição, saldo físico e estados reservado/disponível se existirem.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Identificação inequívoca documentada ou Dado indisponível no Bling; sem hardcode.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 3

**Tarefas:** BK-040, BK-048, BK-077

### HT-04 — Suficiência do balancete

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica e validação financeira.

Verificar saldos iniciais, débitos, créditos e contrapartidas reais e validar suficiência com o financeiro.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Matriz de cobertura contábil; não apresentar fluxo financeiro como balancete completo.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 4

**Tarefas:** BK-007, BK-040, BK-068

### HT-05 — Custo histórico

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Homologar endpoint/campo e provar diferença entre custo da venda antiga e custo atual.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Exemplos temporais verificáveis; se só existir custo atual, declarar ausência de custo histórico.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 5

**Tarefas:** BK-040, BK-063

### HT-06 — Mapa da DRE

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica e validação financeira.

Finalizar origem → conta Bling → classificação gerencial → linha da DRE, incluindo competência e ausências.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Mapa versionado com evidências técnicas e validação financeira; desconhecido permanece pendente.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 6

**Tarefas:** BK-007, BK-040, BK-064, BK-066

### HT-07 — BOM/ficha técnica

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica e validação operacional.

Homologar componente, quantidade e unidade reais por produto no Bling, além dos fatores industriais que existirem.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Composição demonstrada; arquivo de perdas não fornece scrap/yield.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 7

**Tarefas:** BK-040, BK-048, BK-094

### HT-08 — Chaves da conciliação

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Homologar chaves dos arquivos Nuvem e dos recebimentos do Bling, preservando D+1 e comprovação de liquidação.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Amostras de vínculo e ambiguidade; nenhuma tripla contagem de venda/recebível/dinheiro.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 8

**Tarefas:** BK-010, BK-040, BK-046, BK-053

### HT-09 — Backtest de forecast

**Situação:** A verificar tecnicamente. **Natureza:** Verificação técnica.

Executar validação temporal com histórico confiável de 2025 em diante.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Janelas, cobertura, fórmulas, métricas e limitações reais registradas; ainda não executado.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 9

**Tarefas:** BK-009, BK-087, BK-090

### HT-10 — Fontes e comparação de cacau

**Situação:** Em aberto. **Natureza:** Definição de fontes e regras.

Refinar posteriormente as fontes e bases comparáveis antes da construção dos sinais.

**Registro atual:** Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.

**Impacto / evidência necessária:** Fontes e regras aprovadas; nenhuma pasta ou feed ativo presumido.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 10

**Tarefas:** BK-011, BK-085

## Fontes e histórico preservado

A análise detalhada dos 16 PNGs (15 imagens únicas), das 20 abas e das inconsistências visuais permanece no documento da versão 0, como registro histórico de leitura. As regras atuais desta versão substituem suas suposições de fontes nos pontos descritos acima. Valores ilustrativos não se tornam dados reais.

- [Texto da consolidação recebido](../FONTES.md#consolidacao-das-respostas).
- [Análise original das imagens e planilha](<MAJUCAU_ANALISE_E_DECISOES_v0.md>).
- [Backlog atualizado](<MAJUCAU_BACKLOG_DETALHADO_v1.md>).

SHA-256 da nova fonte: `ed8f04f469577e8772a91412edb6f49df65e5788beafd7f87920e226aa6aae42`.

Verificação desta entrega: consistência de IDs, vínculos e dependências; preservação das versões e originais; fórmulas de evolução; inspeção visual da planilha. Isso verifica os documentos do planejamento, não a integração ou o software Majucau.