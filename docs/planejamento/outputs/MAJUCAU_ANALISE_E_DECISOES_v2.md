# Majucau — fechamento vigente e rastreabilidade, versão 2

Esta revisão corrige a interpretação que tratava regras já definidas e responsabilidades técnicas como perguntas de negócio. O fechamento foi lido integralmente e aplicado ao backlog. O usuário confirmou expressamente: somente atualizar o backlog nesta rodada.

70 tarefas revisadas; IDs BK-000 a BK-126 preservados. Três levantamentos continuam concluídos, 2,4% por quantidade e 0% de desenvolvimento. Não houve acesso à conta real Bling/Drive nem implementação ou teste do produto.

## Substituições expressas

| Interpretação anterior | Regra vigente | Referência |
|---|---|---|
| DRE calculada/classificada independentemente no Majucau | Reproduzir DRE Bling por competência, preservando contas e classificações. | §2 |
| Rotina Nuvem D+1 | Análise/conciliação normal dos movimentos até D-1; dia corrente só com solicitação expressa. Agenda futura não é eliminada. | §8.11 |
| Existência de pedido elimina recomendação | Só a quantidade pendente que chega em tempo reduz a necessidade; parcial/tardio não mascara falta. | §3.1 |
| Produção projetada não desconta estoque pronto | Considerar demanda, estoque-alvo, estoque utilizável e produção válida, sem dupla contagem. | §4.5 |
| Ratear automaticamente meta financeira por mix/preço | Meta financeira é cenário comercial; forecast por SKU depende de dados reais. Meta direta só quando cadastrada. | §4.6 |
| Reabrir diferenças de números e layout como novas decisões | Números são ilustrativos; visual vem do mockup congelado e função do contrato vigente. | §5 |
| Infraestrutura/permissões como questionário financeiro | Engenharia propõe capacidades, serviço local, segurança, backup, recuperação e suporte; valida funcionamento desejado. | §§6–7 |

## Regras do fechamento

- **F-01 — Calculadora: contrato fechado:** Regra fechada: disponibilidade consolidada alimenta a calculadora. Contrato preserva jan..M−2, principal líquido e fórmulas consolidadas; DRE apresentada no Majucau agora reproduz o Bling. Cartão/MDR/parcelas não pertencem à calculadora.
- **F-02 — DRE por competência do Bling:** Fonte primária: DRE por competência do Bling. Preservar classificação, contas e competência; não criar estrutura financeira/gerencial paralela ou fatos fictícios. Validar contra o próprio Bling.
- **F-03 — Balancete: reproduzir a origem:** O usuário informa base suficiente no Bling. Mapear recursos reais, reproduzir contas/saldos/competência e comparar totais; elemento específico não acessível = Dado indisponível no Bling.
- **F-04 — Compra líquida e cobertura parcial:** Necessidade bruta − estoque utilizável − cobertura pendente que chega em tempo. Exemplo: 100−20−30=50 kg. Pedido parcial não exclui insumo; pedido tardio não cobre necessidade anterior à chegada.
- **F-05 — Prioridade de compra:** Priorizar risco imediato de ruptura e impacto na produção usando cobertura, demanda, lead time, pedidos e datas reais. Distinguir temporalidade e cobertura; não inventar prioridade aleatória.
- **F-06 — Custos e tarifas separados:** Bruto, variável, fixo, critério de rateio e custo apropriado separados. Usar apropriação Bling existente; não multiplicar tarifa integral por produto nem inventar rateio. Não misturar custo com quantidade a comprar.
- **F-07 — Cacau independente do motor principal:** Inteligência adicional usa localidades reais dos pequenos produtores e sinais comparáveis. Refinar fontes de preço/clima sem impedir compras operacionais; ausência de sinal não cria região ou preço.
- **F-08 — Produção e métrica única:** Realizado vem do Bling. T01 Produzido Geral e T10 detalhe usam mesma função, fonte, snapshot e universo; mesma métrica em qualquer componente tem mesmo resultado.
- **F-09 — Status temporal de produção:** Realizado >= meta = atingida. Em andamento/abaixo dependem do plano no ponto temporal; sem granularidade, não inventar atraso, tolerância ou andamento linear.
- **F-10 — Produção considera estoque:** Forecast por SKU → demanda → estoque/reposição → produção → BOM/insumos/compras. Considerar estoque utilizável e produção válida sem dupla contagem; regra antiga sem desconto de estoque foi superada.
- **F-11 — Meta financeira não fabrica SKUs:** Meta de faturamento é referência comercial/cenário. Forecast se fundamenta em histórico; meta direta por produto só é planejamento quando existir explicitamente. Não ratear total financeiro automaticamente em unidades.
- **F-12 — Cobertura histórica real:** Usar histórico confiável de 2025 em diante e medir cobertura real por SKU; não completar meses ou assumir venda zero. Qualidade requer backtest real.
- **F-13 — Precedência e design congelado:** Visual/layout: mockup aprovado. Fonte/cálculo/comportamento: documentação consolidada. Usar última versão aprovada/congelada, sem redesenho. Artefato ausente precisa ser localizado, sem reabrir a aprovação.
- **F-14 — Exportação aprovada preservada:** Manter exportações aprovadas, inclusive Excel onde especificado. Arquivo usa os mesmos dados, filtros, período, snapshot e métricas da tela.
- **F-15 — Capacidades e menor privilégio:** Usuário + perfil + capacidade. Separar leitura/exportação/metas/parâmetros/pendências/usuários/auditoria; não hardcodar nomes nem permitir autoelevação.
- **F-16 — Vigência, fechamento e auditoria:** Metas são planejamento. Alterações registram usuário, antes/depois, data, período e vigência. Período fechado não muda silenciosamente; reprocessamento é explícito, autorizado, auditado e preserva versão anterior.
- **F-17 — Operação é engenharia:** Windows local por serviços independentes de login humano, inicialização controlada, autenticação, segredos no backend, backup automático, restauração testada, logs e suporte documentados. Acesso externo só quando habilitado.
- **F-18 — Bling e exceções específicas:** Bling é principal para fatos operacionais/financeiros. Drive read-only: perdas operacionais, recebíveis futuros Nuvem e dados específicos Nuvem Envio. Exceções não substituem fatos Bling fora de seu escopo.
- **F-19 — Conciliação normal em D-1:** Análise/conciliação normal de movimentos vai até o dia anterior. Dia corrente somente mediante solicitação expressa. Não generalizar a exceção de 04/09/2026 em regra fixa de sexta-feira/feriado. D-1 não corta a agenda futura nem define frequência da coleta.
- **F-20 — Conciliação, ausência e idempotência:** ID Bling prioritário; identificação/nome+valor no fallback; ambiguidade não concilia automaticamente. Ausência != zero; PENDING_RECONCILIATION fora de indicadores que exigem reconciliação. Repetição não duplica fatos/KPIs; origem é rastreável.

## Execução técnica ainda não realizada

As dez frentes HT abaixo são trabalho futuro, não novas perguntas de negócio. Os registros L/Q legados foram reclassificados para preservar rastreabilidade. Uma regra definida não comprova endpoint, campo, amostra, instalação ou resultado de teste.

| ID | Trabalho técnico | Tarefas |
|---|---|---|
| HT-01 | Mapear categoria, sinal, conta e aplicação, incluindo resgate parcial/total no Bling. Conceito aportes menos resgates está definido. | BK-040, BK-071, BK-072 |
| HT-02 | Obter posição, data, instituição e aplicação quando acessíveis; exibir componentes conhecidos sem presumir reinvestimento. | BK-040, BK-071, BK-074 |
| HT-03 | Homologar identificação operacional inequívoca, saldos obrigatórios e estados opcionais. Não hardcodar depósito. | BK-040, BK-048, BK-077 |
| HT-04 | Mapear recursos existentes e reproduzir contas/saldos/competência; comparar totais contra Bling, sem nova estrutura contábil. | BK-007, BK-040, BK-068 |
| HT-05 | Demonstrar custo histórico versus atual na origem, sem reconstruir a DRE por método paralelo. | BK-040, BK-063 |
| HT-06 | Mapear recursos/campos e correspondência técnica de linhas da DRE por competência; preservar classificação e validar totais contra Bling. | BK-007, BK-040, BK-064, BK-066 |
| HT-07 | Homologar componentes, quantidades e unidades; converter produção necessária em consumo por data sem fatores de perdas operacionais. | BK-040, BK-048, BK-094 |
| HT-08 | Implementar ID Bling/fallback nome+valor sem ambiguidade, D-1 normal e exceção corrente expressa; preservar agenda futura e impedir dupla contagem. | BK-010, BK-040, BK-046, BK-053 |
| HT-09 | Executar backtest com histórico confiável de 2025 em diante, cobertura real, métricas e limitações comprovadas. | BK-009, BK-087, BK-090 |
| HT-10 | Refinar fontes, localidades reais e comparações, sem bloquear recomendação operacional. Verificar somente fontes habilitadas. | BK-011, BK-085 |

## Cuidados incorporados à especificação

- DRE/balancete têm cadeia própria de obtenção no Bling. Arquivos Nuvem, perdas e fretes não reconstroem uma segunda demonstração.
- Agenda futura → caixa projetado → disponibilidade → calculadora. Corte de movimentos realizados, data da coleta e competência permanecem conceitos separados.
- Entrada ausente ou pendente de conciliação não vira zero. Pendente fica fora apenas dos indicadores que exigem reconciliação, com sua condição visível.
- O campo/contrato Caixa Semana Crítica será localizado tecnicamente na referência consolidada. O fechamento da regra não autoriza escolher um saldo inicial/final/mínimo por conveniência.
- A fórmula de §4.5 veio com um marcador de lista antes de estoque-alvo. Foi registrada a normalização técnica aditiva: demanda + estoque-alvo − estoque utilizável − produção válida ainda não incorporada ao saldo. Não é multiplicação literal. Entradas obrigatórias e vigências devem ser comprovadas; necessidade não positiva não gera produção negativa.
- T01/T10 compartilham a mesma informação de produção. Isso não iguala meta operacional, forecast e meta financeira da T12.
- T09/T11 só conciliam contagens quando a métrica e universo forem equivalentes. Necessidade futura pode existir acima do mínimo atual.
- O cálculo cronológico usa posição inicial e saldo anterior; BK-092 não depende do resultado futuro de BK-093. Produção já no estoque não é descontada novamente.
- Pedido pendente é descontado uma única vez, limitado à cobertura em tempo; pedido tardio continua visível.
- Fontes de cacau podem ser refinadas sem bloquear compras principais. Os testes de sinais habilitados continuam necessários.
- Localizar arquivo aprovado ausente não é reabrir uma tela. Não há autorização para redesenhar, reorganizar menu ou inventar conteúdo.
- Períodos fechados preservam snapshots e versões; parâmetro atual não retroage. Reprocessamento exige capacidade, solicitação explícita e auditoria.
- Host local funciona por serviços sem sessão humana. Acesso externo é condicional, sem exposição ativada nesta rodada.

## Tarefas revisadas

| ID | Tarefa | Campos alterados |
|---|---|---|
| BK-000 | Confirmar a base de escopo e as decisões da Majucau | action, steps, accept, status |
| BK-001 | Inventariar os materiais e suas versões | steps |
| BK-004 | Localizar e versionar os mockups e contratos aprovados | title, goal, action, steps, accept, status |
| BK-005 | Detalhar os casos de pagamento divergente após decidir C-01 | steps, accept, status |
| BK-006 | Especificar tecnicamente a regra fechada da calculadora | title, action, steps, accept, status |
| BK-007 | Especificar a reprodução da DRE e do balancete do Bling | title, goal, action, steps, accept, status, owner |
| BK-008 | Especificar necessidade líquida e acompanhamento da produção | title, action, steps, accept, status |
| BK-009 | Especificar forecast, metas e planejamento de reposição | title, action, steps, accept, status, owner |
| BK-010 | Homologar contratos do Bling e das três estruturas do Google Drive | steps, status |
| BK-011 | Definir fontes e critérios dos sinais do cacau | steps, accept, status, owner |
| BK-012 | Especificar ações e exportações já previstas | title, action, steps, accept, status |
| BK-013 | Especificar perfis, capacidades, vigência e auditoria | title, action, steps, accept, status, owner |
| BK-014 | Inventariar o Windows e propor a operação local | title, action, steps, accept, status, owner |
| BK-015 | Organizar responsáveis, estimativas e marcos de entrega | steps, status |
| BK-016 | Modelar os domínios e o vocabulário do negócio | steps |
| BK-019 | Descrever contratos de API e estados de leitura | steps |
| BK-020 | Modelar estado, proveniência e consistência entre telas | steps, accept |
| BK-021 | Transcrever o design congelado em componentes | title, action, steps, accept |
| BK-022 | Especificar acessibilidade, telas pequenas e estados de exceção | steps |
| BK-033 | Aplicar permissões no backend e na interface | steps |
| BK-034 | Registrar auditoria de acesso e mudanças | steps |
| BK-036 | Implementar registro e vigência dos parâmetros | steps |
| BK-037 | Implementar T16 Parâmetros e fluxos autorizados de edição | steps, accept |
| BK-038 | Implementar proveniência e snapshots de fatos | steps |
| BK-039 | Implementar execução controlada de trabalhos em segundo plano | steps |
| BK-040 | Validar acesso e cobertura dos dados do Bling | steps, accept |
| BK-041 | Implementar o adaptador de leitura do Bling | steps |
| BK-046 | Ingerir recebíveis e eventos do Nuvem Pago | steps, accept |
| BK-051 | Controlar atualização e falha por fonte | steps, accept |
| BK-053 | Implementar conciliação por evidência e unicidade | steps, accept |
| BK-054 | Implementar revisão, confirmação e desfazimento de conciliação | steps |
| BK-056 | Implementar regras e leitura de Contas a Receber | steps |
| BK-060 | Implementar fluxo de caixa realizado, previsto e projetado | steps, accept |
| BK-061 | Implementar seleção e memória da semana crítica | steps |
| BK-064 | Preservar contas, classificações e competência do Bling | title, goal, action, steps, accept, deps |
| BK-065 | Implementar perdas e recuperações sem dupla contagem | steps, accept |
| BK-066 | Obter e reproduzir a DRE por competência do Bling | title, goal, action, steps, accept, deps |
| BK-067 | Implementar T06 DRE | steps |
| BK-068 | Reproduzir o balancete a partir da base do Bling | title, action, steps, accept, deps |
| BK-069 | Implementar T07 Balancete | steps |
| BK-071 | Implementar movimentos de investimentos e posição comprovada no Bling | steps |
| BK-072 | Implementar o cálculo de capacidade de aplicação | steps, accept |
| BK-075 | Implementar análise de preço e margem no local aprovado | steps, accept |
| BK-077 | Implementar classificação e posição de estoque | steps |
| BK-079 | Implementar metas de produção por SKU e período | steps |
| BK-080 | Implementar indicadores de produção realizada | steps, accept |
| BK-081 | Implementar T10 Produção | steps |
| BK-082 | Avaliar necessidade e cobertura temporal das compras | title, action, steps, accept, deps |
| BK-084 | Calcular necessidade líquida e prioridade operacional de compra | title, action, steps, accept |
| BK-085 | Implementar sinais externos do cacau | steps, accept, condition |
| BK-086 | Implementar T11 Compras e exportação de recomendações | action, steps, deps |
| BK-091 | Implementar forecast e metas explícitas como planejamento distinto | title, action, steps, accept |
| BK-092 | Implementar produção projetada a partir das vendas | action, steps, accept, deps |
| BK-093 | Implementar evolução do estoque projetado | steps, accept |
| BK-094 | Implementar necessidade de insumos e OP Projetada | steps |
| BK-095 | Implementar resumo financeiro dos cenários | steps, accept |
| BK-096 | Implementar T12 Projeções e ajuste de metas | steps |
| BK-102 | Implementar T01 Visão Executiva | action, steps |
| BK-103 | Implementar os módulos adicionais confirmados no escopo | action, steps |
| BK-104 | Implementar exportações financeiras e de controle aprovadas | steps, accept |
| BK-105 | Rastrear cada requisito até tarefa e teste | steps |
| BK-106 | Validar regras de domínio e casos de fronteira | steps, deps |
| BK-107 | Validar integrações, concorrência e reinício | steps |
| BK-108 | Validar fidelidade visual e acessibilidade | steps |
| BK-110 | Executar jornadas completas de ponta a ponta | steps |
| BK-114 | Definir a topologia final da instalação local | steps, accept |
| BK-118 | Configurar backup de banco, arquivos e configuração | steps |
| BK-119 | Testar restauração e recuperação do ambiente | steps |
| BK-120 | Configurar monitoramento operacional local | steps |
| BK-123 | Documentar e treinar a rotina da Majucau | steps |

## Registro vigente de regras e verificações

Este registro não é um questionário. Definida indica regra recebida; Execução técnica pendente indica trabalho por executar; Referência a localizar indica artefato aprovado não identificado no pacote disponível. Atribuição efetiva de acesso/host e validação de funcionamento pertencem à implantação.

### D-01 — Empresa atendida

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Somente Majucau. A resposta anterior sobre várias empresas foi corrigida pelo usuário.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-02 — Usuários

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Três pessoas da diretoria, cada uma em seu computador.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-03 — Hospedagem

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Local, em computador/servidor Windows existente. Demais características ainda desconhecidas.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-04 — Primeira produção

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Todos os módulos confirmados do material na primeira entrada em produção; etapas internas são permitidas para organizar o trabalho.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-05 — Conflitos entre fontes

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Precedência atual: mockup aprovado define visual; documentação consolidada define cálculo, fonte e comportamento; última versão aprovada/congelada prevalece. A orientação anterior de decidir cada conflito caso a caso foi substituída nos pontos fechados.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-06 — Tecnologias

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Monólito com DDD e arquitetura hexagonal; backend Go; frontend JavaScript com React.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-07 — Balanço Patrimonial

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Fora da primeira versão; retirar o item dos menus por decisão explícita do usuário.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### D-08 — Natureza desta entrega

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Somente atualização do backlog nesta rodada, confirmado expressamente pelo usuário após o envio do fechamento. Implementação e validação real são tarefas futuras.

**Fonte:** Conversa atual; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### C-01 — Pagamento diferente do título

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Marcar divergência e manter o título sem baixa até a análise. Casos complementares serão detalhados em BK-005.

**Fonte:** C_Conflitos!A3:L3; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-005

### C-02 — Semana crítica

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** A semana com o maior total de despesas. A posição de caixa dentro dela ainda precisa ser definida.

**Fonte:** C_Conflitos!A4:L4; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### C-03 — Lucro elegível

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Origem atual do resultado é a DRE por competência do Bling, apresentada no Majucau. A interpretação anterior de uma DRE calculada/classificada independentemente foi superada.

**Fonte:** C_Conflitos!A5:L5; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### L-01 — Gate formal

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Usuário confirmou somente atualização do backlog nesta rodada; o fechamento funcional não representa implementação executada.

**Fonte:** D_Lacunas!A3:G3; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### L-02 — Sistema atual

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Verificar tecnicamente a origem e evidências do item; não é nova pergunta de negócio.

**Fonte:** D_Lacunas!A4:G4; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### L-03 — CMV histórico

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Verificar tecnicamente a origem e evidências do item; não é nova pergunta de negócio.

**Fonte:** D_Lacunas!A5:G5; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-04 — Perdas — competência

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Perdas operacionais mantêm fonte própria; efeito financeiro e competência seguem os fatos registrados/classificados no Bling. Não criar ou deslocar linhas da DRE por inferência.

**Fonte:** D_Lacunas!A6:G6; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-05 — Juros Nuvem

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Juros, taxas e rendimentos respeitam conta/classificação e competência existentes no Bling. Mapear tecnicamente, sem nova classificação de negócio.

**Fonte:** D_Lacunas!A7:G7; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-06 — Tarifa fixa Pricing

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Apropriação Bling tem preferência; custos bruto/variável/fixo/critério/final separados. Não duplicar tarifa nem fabricar rateio sem dados.

**Fonte:** D_Lacunas!A8:G8; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-075

### L-07 — Caixa da semana crítica

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** A calculadora está fechada. Localizar tecnicamente o contrato/campo Caixa Semana Crítica na referência vigente; não escolher início/fim/mínimo sem evidência. Não retornar como pergunta financeira.

**Fonte:** D_Lacunas!A9:G9; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-006

### L-08 — Balancete — saldos iniciais

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Reproduzir balancete existente no Bling com campos reais, competência e classificação. Dado específico não acessível pela API permanece indisponível.

**Fonte:** D_Lacunas!A10:G10; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-09 — DRE — fontes complementares

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** DRE usa origem Bling; sem nova estrutura ou fonte contábil paralela. Exceções externas têm função própria e não completam linhas ausentes com hipóteses.

**Fonte:** D_Lacunas!A11:G11; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-10 — UI Pricing

**Situação:** Referência a localizar. **Natureza:** Rastreabilidade documental.

**Regra / trabalho vigente:** Localizar última versão aprovada/congelada do bloco/tela; não reabrir UX nem criar rota por inferência.

**Fonte:** D_Lacunas!A12:G12; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-075

### L-11 — Arquivos Nuvem

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Ler a agenda futura na estrutura definida; mapear arquivos/chaves tecnicamente. Conciliação normal de movimentos usa D-1, sem cortar vencimentos futuros da agenda.

**Fonte:** D_Lacunas!A13:G13; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010

### L-12 — Ciclo de vida de arquivo

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Origem somente de leitura. Proibido criar, mover, renomear, excluir, sobrescrever ou alterar arquivos. Controle, hash, lotes, rejeições e versões ficam internamente; arquivo corrigido não sobrescreve fatos silenciosamente.

**Fonte:** D_Lacunas!A14:G14; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010

### L-13 — Fretes/taxas na DRE

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Frete cobrado/custo/subsídio separados nos dados específicos; efeito na DRE respeita classificação/competência Bling, sem nova reclassificação.

**Fonte:** D_Lacunas!A15:G15; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-14 — Recuperações

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Recuperação é evento rastreável; caixa depende de fato financeiro registrado e DRE respeita a origem. Aprovação de ressarcimento não cria dinheiro.

**Fonte:** D_Lacunas!A16:G16; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### L-15 — Custo de carregar estoque

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Custo de carregamento pertence à análise econômica quando houver dado/regra comprovável. Sua ausência não bloqueia quantidade/prioridade do motor de compras; não inventar taxa.

**Fonte:** D_Lacunas!A17:G17; Fechamento das questões ainda tratadas como abertas — 13/09/2026

### L-16 — Mockup Logística

**Situação:** Referência a localizar. **Natureza:** Rastreabilidade documental.

**Regra / trabalho vigente:** Localizar última versão aprovada/congelada do bloco/tela; não reabrir UX nem criar rota por inferência.

**Fonte:** D_Lacunas!A18:G18; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-076

### L-17 — Compras — quantidade recomendada

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Necessidade líquida usa demanda bruta, estoque utilizável e cobertura em tempo; pedido parcial reduz somente parte efetivamente coberta.

**Fonte:** D_Lacunas!A19:G19; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-008

### L-18 — Compras — prioridade e prazo histórico

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Prioridade operacional usa risco, cobertura, lead time, pedidos e datas reais, com ordenação explicável. Estatística/janela é especificação técnica rastreável.

**Fonte:** D_Lacunas!A20:G20; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-008

### L-19 — Cacau — sinais externos

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Refinamento técnico de fontes comparáveis e localidades reais; sinais não bloqueiam motor principal de compras.

**Fonte:** D_Lacunas!A21:G21; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-011

### L-20 — Forecast — thresholds de qualidade

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Avaliação temporal e critérios de qualidade são execução técnica com cobertura histórica real, sem valores do mockup como prova.

**Fonte:** D_Lacunas!A22:G22; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-009

### L-21 — Meta de faturamento — distribuição

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Não distribuir meta financeira automaticamente por SKU. Forecast vem do histórico; metas diretas por produto só quando explicitamente cadastradas.

**Fonte:** D_Lacunas!A23:G23; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-009

### L-22 — Telas 13–15 — exportação

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Exportações previstas/aprovadas permanecem; implementar contrato técnico com mesmos dados/filtros e Excel quando especificado, sem perguntar novamente se a função entra.

**Fonte:** D_Lacunas!A24:G24; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-012

### L-23 — Tela 13 — resolução por tipo

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Ciclo de pendência rastreável, com ator/data/alteração/origem e justificativa técnica; resolução trata a causa dentro da capacidade correspondente.

**Fonte:** D_Lacunas!A25:G25; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-012

### L-24 — Tela 15 — freshness por fonte

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Frequência/timeout/validade por fonte são configurações técnicas; análise/conciliação de movimentos usa D-1. Não manter D+1 como regra corrente.

**Fonte:** D_Lacunas!A26:G26; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010

### L-25 — Tela 16 — governança de edição de parâmetros

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Parâmetros editáveis por capacidade, com antes/depois, usuário, data e vigência. Período fechado não se altera silenciosamente; reprocessamento explícito auditado.

**Fonte:** D_Lacunas!A27:G27; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-013

### L-26 — Tela 17 — identidade, ciclo de vida e matriz de permissões

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Arquitetura de identidade e matriz de capacidades são responsabilidade técnica; autorização não depende de nomes de diretores e segue menor privilégio.

**Fonte:** D_Lacunas!A28:G28; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-013

### Q-01 — Referência visual

**Situação:** Referência a localizar. **Natureza:** Rastreabilidade documental.

**Regra / trabalho vigente:** Reproduzir a última referência visual aprovada/congelada; localizar o arquivo oficial se necessário. Não redesenhar menu/logo ou reabrir tela congelada.

**Fonte:** T12/T13 versus demais imagens; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-004, BK-021

### Q-02 — Módulos sem especificação completa

**Situação:** Referência a localizar. **Natureza:** Rastreabilidade documental.

**Regra / trabalho vigente:** Localizar os artefatos aprovados dos módulos. Ausência física no pacote é rastreio documental, não nova aprovação de tela nem licença para inventar conteúdo.

**Fonte:** Menus; A_Matriz_Mestra; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-004, BK-021, BK-074, BK-102, BK-103

### Q-03 — Máquina Windows

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Inventariar host Windows e recursos por inspeção técnica quando disponível; não inventar CPU/RAM/disco nem devolver dimensionamento como pergunta financeira.

**Fonte:** Resposta do usuário: Windows; demais dados desconhecidos; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-014

### Q-04 — Locais de acesso

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Acesso autenticado. Acesso externo é condicional a habilitação explícita e usa canal seguro; esta rodada não publica/expoe serviços.

**Fonte:** Uso em três computadores confirmado; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-014

### Q-05 — Disponibilidade local

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Propor serviços sem sessão humana, inicialização controlada e comportamento documentado com falhas de energia/internet; validar funcionamento desejado na implantação.

**Fonte:** Hospedagem local; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-014

### Q-06 — Contas e fontes reais

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Bling/API concentra operacional/financeiro, Itaú principal, Mercado Pago e Conta Caixa/investimentos; Drive contém apenas perdas, Nuvem Pago e Nuvem Envio. Cacau continua refinamento posterior. Identificadores e capacidade real do Bling são verificações técnicas.

**Fonte:** H_Contrato_Dados; M-008; T15; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010

### Q-07 — Estoque e pedido aplicável

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Pedido parcial não elimina compra: descontar só quantidade pendente em tempo. Depósito vem inequivocamente do Bling; ausência não é hipótese.

**Fonte:** T09; Q_CEI_Tela_11; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-008

### Q-08 — Situação das metas

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Status depende de realizado, meta e posição temporal do plano. Sem granularidade suficiente, não classificar atraso arbitrário ou assumir progresso linear.

**Fonte:** T10; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-008, BK-079

### Q-09 — Uso dos demonstrativos

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** DRE por competência e balancete do projeto utilizam Bling e suas classificações. Validar totais tecnicamente; não solicitar estrutura contábil paralela.

**Fonte:** T06/T07; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007

### Q-10 — Período e limite de aplicação

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Calculadora fechada. Consolidar fórmulas M-030/M-034 preservadas pelo fechamento, usando fonte DRE Bling e aportes líquidos de resgates; limite executável Itaú.

**Fonte:** M-030..M-034; T08; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-006, BK-071

### Q-11 — Vendas para produção projetada

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Produção necessária considera demanda, estoque-alvo e estoque/produção válidos no horizonte; contrato anterior sem desconto de estoque foi superado. Documentar normalização técnica da fórmula.

**Fonte:** M-047; T12; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-009, BK-092

### Q-12 — Relação entre metas

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** T01/T10 compartilham a mesma informação de produção. Metas são planejamento versionado, distinto de realizado e forecast T12; não criar metas independentes para o mesmo indicador nem sincronizar conceitos diferentes automaticamente.

**Fonte:** T10/T12/T16; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-009, BK-079

### Q-13 — Ações visíveis e exportação

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Botões e navegação seguem referência aprovada; contrato de execução e exportação é detalhamento técnico, sem redesenhar ou retirar função existente.

**Fonte:** T02–T07; T13–T15; L-22; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-012

### Q-14 — Permissões da diretoria

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Usuário, perfil e capacidade com menor privilégio; atribuição concreta é configuração administrativa explícita na implantação, sem hardcode por pessoa.

**Fonte:** T17; L-26; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-013

### Q-15 — Recuperação

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Desenvolvimento propõe objetivos de recuperação, backup e procedimento; medir restauração e validar funcionamento desejado. Não registrar prazo/perda tolerável como já aprovados sem evidência.

**Fonte:** Requisito de produção local; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-014

### Q-16 — Backup e responsável

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Projetar backup automático de dados locais, configuração e histórico; destino seguro fora do disco único e das três estruturas read-only; documentação de operador/suporte.

**Fonte:** Requisito de produção local; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-014

### Q-17 — Prazo e equipe

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Organizar capacidade, estimativas e responsáveis por papel tecnicamente. Prazo real depende de alocação e evidências, sem prometê-lo ou converter isso em regra financeira.

**Fonte:** Pergunta inicial respondida apenas quanto à hospedagem; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-015

### Q-18 — Dispositivos e linguagem

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Preservar visual congelado; acessibilidade, foco e estados técnicos dentro do padrão existente. Não usar responsividade como autorização de redesenho.

**Fonte:** Gaps UX em todas as imagens; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-021, BK-022

### Q-19 — Conciliação detalhada

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** ID Bling prioritário; identificação/nome + valor no fallback; ambiguidade fica para conferência, sem tolerância de valor inventada.

**Fonte:** T05; M-008; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-053

### Q-20 — Recebimentos e status

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Agenda futura é previsão; Bling fornece fato financeiro. Aplicar D-1 ao realizado normal; pendente de conciliação fica fora de indicador que exige valor reconciliado.

**Fonte:** T03/T05; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-056

### Q-21 — Calendário e caixa da semana

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Calculadora fechada; calendário/corte e contrato de caixa são especificação técnica. Corte normal D-1; não escolher posição crítica por conveniência nem criar regra fixa de sexta-feira.

**Fonte:** T02/T08; L-07; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-006, BK-060

### Q-22 — Gráfico da DRE

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Preservar gráfico aprovado e usar métrica/base de cálculo da documentação funcional; valores demonstrativos não impõem resultado. Não somar subtotais sobrepostos como população única.

**Fonte:** T06; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-067

### Q-23 — Saldos do balancete

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Usar contas, natureza e saldos do Bling, preservando competência; ausência de saldo não equivale a zero. Conferir totais na mesma posição.

**Fonte:** T07; L-08; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-069

### Q-24 — Ficha técnica de insumos

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Homologar campos reais de BOM, unidades e fatores industriais existentes; perdas operacionais não são yield/scrap. Não pedir nova regra genérica já fechada.

**Fonte:** H_Contrato_Dados!A12:H12; T12; Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-094

### Q-25 — Resumo financeiro projetado

**Situação:** Referência a localizar. **Natureza:** Rastreabilidade documental.

**Regra / trabalho vigente:** Implementar aba prevista na referência aprovada; cenário é planejamento, sem forçar meta de faturamento a coincidir com SKUs por rateio fictício.

**Fonte:** T12; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-095

### Q-26 — Alertas e monitoramento

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Comportamento/prioridade vêm da documentação consolidada e visual da última referência aprovada. Modelar monitoramento técnico e canais habilitados sem notificações externas presumidas.

**Fonte:** T14; T_CEI_Tela_14; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-120

### Q-27 — Indicadores de prazo

**Situação:** Execução técnica pendente. **Natureza:** Execução técnica.

**Regra / trabalho vigente:** Especificar tecnicamente fórmulas, unidades, períodos e fontes dos indicadores já aprovados; reproduzir componentes comprováveis e indisponibilidade, sem inventar números do mockup.

**Fonte:** T08; B_Intersecoes!A12:I12; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-070

### R-01 — Fontes e estruturas externas

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Bling/API + Google Drive — pasta compartilhada Majucau, preservando 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Sem pasta de Itaú, Mercado Pago, investimentos ou ajustes.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §1; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010, BK-042, BK-052

### R-02 — Permissão e rastreio no Drive

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Somente leitura, sem qualquer mutação na origem. Metadados, hashes, lotes, rejeições e vínculos de reprocessamento ficam dentro do Majucau. Correção não sobrescreve fatos silenciosamente.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §2; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010, BK-024, BK-038, BK-042, BK-044, BK-118

### R-03 — Identificação do depósito

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Identificar no cadastro operacional da API, sem hardcode ou escolha automática de Bloqueado. Sem evidência inequívoca: Dado indisponível no Bling.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §3; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-048, BK-077

### R-04 — Fonte de investimentos

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Conta Caixa no Bling; caixa.php é referência funcional. Usar API oficial ou exportação estruturada oficial. Scraping proibido; não haverá extrato externo na pasta compartilhada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §4; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-041, BK-071, BK-074

### R-05 — Categorias de investimentos

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Transferência destinada à aplicação identifica aporte; Rendimentos identifica rendimento realizado. Resgates ainda precisam ser mapeados tecnicamente.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §5; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-064, BK-071

### R-06 — Já Investido

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Aportes de principal − resgates de principal. Rendimentos ficam separados e não são novos aportes. Falta provar os componentes no Bling.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §6; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-006, BK-071, BK-072

### R-07 — Posição comprovada

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Exibir posição e data somente se disponíveis na fonte; caso contrário mostrar apenas componentes comprováveis. Não presumir rendimento reinvestido.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §7; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-071, BK-073, BK-074

### R-08 — Itaú e valor executável

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Itaú é a conta principal e fornece o saldo transferível, via Bling. Valor Executável Hoje = MIN(Máximo Aplicável Financeiramente; Saldo Transferível do Itaú).

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §8; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-006, BK-047, BK-060, BK-072

### R-09 — Mercado Pago

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Movimentos via Bling; pode compor caixa consolidado, mas não aumenta automaticamente o executável para aplicações que saem do Itaú.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §9; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-047, BK-060, BK-072

### R-10 — Nuvem Pago e liquidação

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Fonte Nuvem permanece em 02_nuvem_pago/recebimentos_futuros. A rotina anterior D+1 foi explicitamente substituída por D-1 para análise/conciliação de movimentos; agenda futura continua projetada.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §10; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010, BK-046, BK-051, BK-053

### R-11 — Perdas operacionais

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Fonte 01_perdas_operacionais representa eventos de perda, nunca percentual industrial, scrap, yield ou fator automático de matéria-prima.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §11; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-050, BK-065, BK-094

### R-12 — Nuvem Envio

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Fonte 03_nuvem_envio; preservar custo efetivo, frete cobrado, diferença/subsídio, pedido, transportadora, etiqueta e data quando disponíveis, sem compensação silenciosa.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; §12; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-049, BK-064, BK-076

### R-13 — Base do forecast

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Executar backtest com histórico confiável de 2025 em diante. Não presumir cobertura de 24 meses nem métricas calculadas.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 9; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-009, BK-087, BK-090

### R-14 — Refinamento do cacau

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Cacau é inteligência adicional com fontes técnicas a refinar; não bloqueia o motor principal de compras e não cria novas estruturas Drive.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 10; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-011, BK-085

### HT-01 — Resgates de principal

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Mapear categoria, sinal, conta e aplicação, incluindo resgate parcial/total no Bling. Conceito aportes menos resgates está definido.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 1; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-071, BK-072

### HT-02 — Posição comprovada de aplicações

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Obter posição, data, instituição e aplicação quando acessíveis; exibir componentes conhecidos sem presumir reinvestimento.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 2; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-071, BK-074

### HT-03 — Campos reais de estoque/depósito

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Homologar identificação operacional inequívoca, saldos obrigatórios e estados opcionais. Não hardcodar depósito.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 3; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-048, BK-077

### HT-04 — Extração e conferência do balancete Bling

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Mapear recursos existentes e reproduzir contas/saldos/competência; comparar totais contra Bling, sem nova estrutura contábil.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 4; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007, BK-040, BK-068

### HT-05 — Custo histórico para análise por SKU

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Demonstrar custo histórico versus atual na origem, sem reconstruir a DRE por método paralelo.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 5; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-063

### HT-06 — Extração e conferência da DRE Bling

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Mapear recursos/campos e correspondência técnica de linhas da DRE por competência; preservar classificação e validar totais contra Bling.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 6; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-007, BK-040, BK-064, BK-066

### HT-07 — BOM e demanda temporal de insumos

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Homologar componentes, quantidades e unidades; converter produção necessária em consumo por data sem fatores de perdas operacionais.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 7; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-040, BK-048, BK-094

### HT-08 — Chaves reais e corte D-1

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Implementar ID Bling/fallback nome+valor sem ambiguidade, D-1 normal e exceção corrente expressa; preservar agenda futura e impedir dupla contagem.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 8; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-010, BK-040, BK-046, BK-053

### HT-09 — Validação temporal do forecast

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Executar backtest com histórico confiável de 2025 em diante, cobertura real, métricas e limitações comprovadas.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 9; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-009, BK-087, BK-090

### HT-10 — Sinais adicionais do cacau

**Situação:** Execução técnica pendente. **Natureza:** Homologação técnica.

**Regra / trabalho vigente:** Refinar fontes, localidades reais e comparações, sem bloquear recomendação operacional. Verificar somente fontes habilitadas.

**Fonte:** Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026; pendência 10; Fechamento das questões ainda tratadas como abertas — 13/09/2026

**Tarefas:** BK-011, BK-085

### F-01 — Calculadora: contrato fechado

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Regra fechada: disponibilidade consolidada alimenta a calculadora. Contrato preserva jan..M−2, principal líquido e fórmulas consolidadas; DRE apresentada no Majucau agora reproduz o Bling. Cartão/MDR/parcelas não pertencem à calculadora.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §1

**Tarefas:** BK-006, BK-019, BK-061, BK-072

### F-02 — DRE por competência do Bling

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Fonte primária: DRE por competência do Bling. Preservar classificação, contas e competência; não criar estrutura financeira/gerencial paralela ou fatos fictícios. Validar contra o próprio Bling.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §2

**Tarefas:** BK-007, BK-040, BK-064, BK-066, BK-067

### F-03 — Balancete: reproduzir a origem

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** O usuário informa base suficiente no Bling. Mapear recursos reais, reproduzir contas/saldos/competência e comparar totais; elemento específico não acessível = Dado indisponível no Bling.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §2

**Tarefas:** BK-007, BK-040, BK-068, BK-069

### F-04 — Compra líquida e cobertura parcial

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Necessidade bruta − estoque utilizável − cobertura pendente que chega em tempo. Exemplo: 100−20−30=50 kg. Pedido parcial não exclui insumo; pedido tardio não cobre necessidade anterior à chegada.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §3.1

**Tarefas:** BK-008, BK-082, BK-084, BK-094

### F-05 — Prioridade de compra

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Priorizar risco imediato de ruptura e impacto na produção usando cobertura, demanda, lead time, pedidos e datas reais. Distinguir temporalidade e cobertura; não inventar prioridade aleatória.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §3.2

**Tarefas:** BK-008, BK-083, BK-084, BK-086

### F-06 — Custos e tarifas separados

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Bruto, variável, fixo, critério de rateio e custo apropriado separados. Usar apropriação Bling existente; não multiplicar tarifa integral por produto nem inventar rateio. Não misturar custo com quantidade a comprar.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §3.3

**Tarefas:** BK-075, BK-084

### F-07 — Cacau independente do motor principal

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Inteligência adicional usa localidades reais dos pequenos produtores e sinais comparáveis. Refinar fontes de preço/clima sem impedir compras operacionais; ausência de sinal não cria região ou preço.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §3.4

**Tarefas:** BK-011, BK-085, BK-086

### F-08 — Produção e métrica única

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Realizado vem do Bling. T01 Produzido Geral e T10 detalhe usam mesma função, fonte, snapshot e universo; mesma métrica em qualquer componente tem mesmo resultado.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §4.1–4.3

**Tarefas:** BK-020, BK-080, BK-081, BK-102

### F-09 — Status temporal de produção

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Realizado >= meta = atingida. Em andamento/abaixo dependem do plano no ponto temporal; sem granularidade, não inventar atraso, tolerância ou andamento linear.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §4.4

**Tarefas:** BK-008, BK-080, BK-081

### F-10 — Produção considera estoque

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Forecast por SKU → demanda → estoque/reposição → produção → BOM/insumos/compras. Considerar estoque utilizável e produção válida sem dupla contagem; regra antiga sem desconto de estoque foi superada.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §4.5

**Tarefas:** BK-009, BK-092, BK-093, BK-094

### F-11 — Meta financeira não fabrica SKUs

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Meta de faturamento é referência comercial/cenário. Forecast se fundamenta em histórico; meta direta por produto só é planejamento quando existir explicitamente. Não ratear total financeiro automaticamente em unidades.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §4.6

**Tarefas:** BK-009, BK-079, BK-091, BK-095, BK-096

### F-12 — Cobertura histórica real

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Usar histórico confiável de 2025 em diante e medir cobertura real por SKU; não completar meses ou assumir venda zero. Qualidade requer backtest real.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §4.7

**Tarefas:** BK-087, BK-089, BK-090

### F-13 — Precedência e design congelado

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Visual/layout: mockup aprovado. Fonte/cálculo/comportamento: documentação consolidada. Usar última versão aprovada/congelada, sem redesenho. Artefato ausente precisa ser localizado, sem reabrir a aprovação.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §5

**Tarefas:** BK-000, BK-004, BK-021, BK-022, BK-108

### F-14 — Exportação aprovada preservada

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Manter exportações aprovadas, inclusive Excel onde especificado. Arquivo usa os mesmos dados, filtros, período, snapshot e métricas da tela.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §5.7

**Tarefas:** BK-012, BK-097, BK-104, BK-109

### F-15 — Capacidades e menor privilégio

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Usuário + perfil + capacidade. Separar leitura/exportação/metas/parâmetros/pendências/usuários/auditoria; não hardcodar nomes nem permitir autoelevação.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §6.1–6.2

**Tarefas:** BK-013, BK-032, BK-033, BK-035

### F-16 — Vigência, fechamento e auditoria

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Metas são planejamento. Alterações registram usuário, antes/depois, data, período e vigência. Período fechado não muda silenciosamente; reprocessamento é explícito, autorizado, auditado e preserva versão anterior.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §6.3–6.8

**Tarefas:** BK-034, BK-036, BK-037, BK-038, BK-079, BK-098

### F-17 — Operação é engenharia

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Windows local por serviços independentes de login humano, inicialização controlada, autenticação, segredos no backend, backup automático, restauração testada, logs e suporte documentados. Acesso externo só quando habilitado.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §7

**Tarefas:** BK-014, BK-114, BK-115, BK-118, BK-119, BK-120, BK-123

### F-18 — Bling e exceções específicas

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Bling é principal para fatos operacionais/financeiros. Drive read-only: perdas operacionais, recebíveis futuros Nuvem e dados específicos Nuvem Envio. Exceções não substituem fatos Bling fora de seu escopo.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §8.1–8.5,8.14

**Tarefas:** BK-010, BK-046, BK-049, BK-052, BK-060, BK-065

### F-19 — Conciliação normal em D-1

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** Análise/conciliação normal de movimentos vai até o dia anterior. Dia corrente somente mediante solicitação expressa. Não generalizar a exceção de 04/09/2026 em regra fixa de sexta-feira/feriado. D-1 não corta a agenda futura nem define frequência da coleta.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §8.11

**Tarefas:** BK-010, BK-039, BK-051, BK-053, BK-054, BK-056, BK-060, BK-123

### F-20 — Conciliação, ausência e idempotência

**Situação:** Definida. **Natureza:** Regra/atribuição definida.

**Regra / trabalho vigente:** ID Bling prioritário; identificação/nome+valor no fallback; ambiguidade não concilia automaticamente. Ausência != zero; PENDING_RECONCILIATION fora de indicadores que exigem reconciliação. Repetição não duplica fatos/KPIs; origem é rastreável.

**Fonte:** Fechamento das questões ainda tratadas como abertas — 13/09/2026; §8.6–8.13

**Tarefas:** BK-020, BK-038, BK-044, BK-053, BK-054, BK-056, BK-107

## Fontes e preservação

- [Fechamento enviado pelo usuário](<../FONTES.md#fechamento-das-questoes>).
- Confirmação posterior do usuário: somente atualizar o backlog.
- [Backlog vigente](<MAJUCAU_BACKLOG_DETALHADO_v2.md>).
- [Análise original dos mockups e planilha](<MAJUCAU_ANALISE_E_DECISOES_v0.md>), preservada como histórico de leitura; regras superadas não orientam a implementação.
- [Consolidação anterior](<MAJUCAU_ANALISE_E_DECISOES_v1.md>), preservada como histórico.

SHA-256 da nova fonte: `0f413addc1e0ab7e3b1d294b86c99c458a3439fa6661e73d6cef01d70ba10341`.

As verificações desta entrega cobrem documentos, IDs, dependências, rastreabilidade, fontes preservadas e fórmulas de evolução. Nenhuma delas representa homologação dos dados Bling, execução de backtest ou teste do produto.