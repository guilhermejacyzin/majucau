# Auditoria de T02–T08 — material para planejamento, sem implementação

## Cobertura e limites

- Lidas visualmente as sete imagens T02_Fluxo_de_Caixa, T03_Contas_a_Receber, T04_Contas_a_Pagar, T05_Conciliacao, T06_DRE, T07_Balancete e T08_Calculadora_de_Aplicacao, incluindo cabeçalhos, indicadores, tabelas, gráficos, ações e avisos.
- Lidas integralmente C_Conflitos (3 conflitos), D_Lacunas (26 lacunas), I_Testes (linhas 1–169; catálogo de 167 cenários, não execução) e J_Autoauditoria. Consultadas entradas financeiras da A_Matriz_Mestra e H_Contrato_Dados por busca dirigida; isso não equivale a releitura integral dessas duas abas.
- Dados numéricos e regras impressas nos mockups são evidências do material. Não constituem, por si, decisões atuais do usuário, dados reais, testes aprovados ou implementação existente.
- Arquitetura hexagonal DDD monolítica, backend Go e frontend React JS são decisões expressas do usuário. Nenhuma fórmula financeira foi escolhida nesta análise.
- Referências RC-2.2, RC-3, LOCK e OPEN são reproduzidas pela consolidação. Os documentos originais referenciados não foram examinados nesta subtarefa.

## Inventário funcional por tela

### T02 — Fluxo de Caixa

**O que aparece:** posição 16/08/2026 14:32; sincronização 14:31; saldo atual R$258.420 com indicação de banco conciliado e atualização própria 14:25; entradas 30 dias R$542.680; saídas 30 dias R$589.230; menores saldos 30/60 dias R$112.540/R$91.870, datas e folgas; gráfico de entradas, saídas, realizado e projetado; semana crítica 14/09–20/09, saídas R$186.400, entradas R$121.300 e reserva R$50.000; posição da semana pendente de decisão; recebido no mês R$301.240, pago R$278.960, reserva R$50.000, saldo transferível R$96.800 e nível de caixa OK. Alertas: abaixo do mínimo, saídas sem categoria, atrasos e conciliação pendente.

**Ações visíveis:** atualizar, exportar (menu sem formatos definidos), ajuda, expandir indicadores, ver fluxo completo, detalhes da semana, abrir lançamentos/recebíveis/conciliação a partir dos alertas. Não há filtro de período explícito equivalente ao que aparece em outras telas.

**Dados e dependências:** contas e saldos bancários, movimentos realizados, títulos a pagar/receber, recebimentos previstos, critérios de deduplicação entre Bling e fontes externas, calendário e cortes de data, reserva versionada, origem e qualidade de cada parcela. Conciliação, importações, parâmetros e regra da semana crítica precedem o cálculo confiável.

**Trabalho de backlog:** especificar dicionário de KPIs e séries; separar saldo, fluxo líquido e movimentos; implementar posteriormente consultas por posição/período com mesma base entre cartões, gráfico e detalhe; estabelecer navegação contextual; criar cenários de saldo negativo, dado ausente, fonte antiga e valores previstos substituídos por realizados.

**Aceite sugerido:** saldo projetado reproduz a equação aprovada usando exatamente os movimentos selecionados; cada entrada aparece uma vez; mudar período atualiza todas as regiões; um clique no KPI abre o conjunto de registros que explica seu valor; fontes indisponíveis não viram zero; semana crítica respeita critério, janela, desempate e posição de caixa aprovados.

### T03 — Contas a Receber

**O que aparece:** total R$367.890, 82 clientes/28 boletos; recebido no mês R$214.560; atraso R$58.240/5 recebimentos; conciliação 91,3%. Tabela de cliente, origem Bling, vencimento, valor bruto, líquido e status. Distribuição: CONFIRMED R$223.480, PENDING_RECONCILIATION R$74.360, PROVISIONAL R$41.700, DIVERGENT R$28.350. Cartões de Pix R$86.200, boletos R$74.800 e cartão R$149.560; 3 ambíguos, 2 chargebacks. Alertas sobre atraso, ambiguidade, 7 transações sem ID Bling e regras de correspondência.

**Ações visíveis:** atualizar, exportar, ajuda, detalhes/recebimentos/atrasos/conciliação, lista completa, distribuição por status e regras.

**Dados e dependências:** título, pedido, transação, evento de liquidação, cliente, parcelas, moeda, datas previstas/efetivas, valores brutos/líquidos/taxas, estornos/chargebacks, identificadores externos e conciliação. O catálogo I_Testes exige preservar valores exportados de taxas e líquido; tabela contratual não pode cobrar novamente evento já líquido.

**Trabalho de backlog:** definir universo do total e do recebimento; distinguir venda, direito a receber, previsão e dinheiro liquidado; modelar eventos de estorno/contestação sem apagar a origem; separar situação do título, correspondência e qualidade da fonte; permitir análise por vencimento e trilha de origem.

**Aceite sugerido:** pedido Bling e evento Nuvem não somam duas vendas; previsão substituída por liquidação não duplica caixa; evento especial preserva sinal e tipo; título futuro pode estar conciliado documentalmente sem ser apresentado como recebido; totais por status e meio de pagamento reconciliam com o universo explicitado.

### T04 — Contas a Pagar

**O que aparece:** total R$412.670; vence hoje R$38.420/12 contas; atraso R$67.950/9 obrigações; divergências R$14.880. Tabela de fornecedor, documento, vencimento, valor, categoria e status. Distribuição CONFIRMED R$249.820, PROVISIONAL R$80.200, OVERDUE R$67.950, DIVERGENT R$14.700. Cartões de fornecedores ativos, fretes, impostos por competência, regra sem pagamento parcial, próximos sete dias; alertas por vencimento, atraso, diferença de valor e proibição de parciais.

**Ações visíveis:** atualizar, exportar, ajuda, detalhes, contas que vencem hoje, atrasos, conciliação, lista completa, distribuição e regra.

**Dados e dependências:** obrigação constituída, fornecedor/documento/parcela, vencimento, movimento e evidência de quitação, classificação, eventual despesa de frete/tributo. A_Matriz_Mestra M-006 exclui compras apenas sugeridas/forecast da constituição automática de A Pagar.

**Trabalho de backlog:** resolver C-01 antes de implementar baixa/saldo; descrever máquina de estados de obrigação e conciliação; calcular aging e vencimentos sem misturar qualidade do dado com atraso; ligar divergência à evidência original sem alterar valores para fechar.

**Aceite sugerido:** cenário título R$1.000/movimento R$600 tem um único resultado aprovado; testar também R$999,99, dois movimentos de R$500 e pagamento integral após vencimento; mesma obrigação não é repetida por importação; sugestão de compra não cria conta a pagar; vencido usa corte de data aprovado.

### T05 — Conciliação

**O que aparece:** 1.284 conciliados, 54 pendentes, 3 ambíguos, 17 divergentes; total 1.358. Tabela liga venda Bling, transação Nuvem, cliente, valor bruto, data de recebimento e status MATCHED/PENDING/AMBIGUOUS/DIVERGENT. Prioridade ID Bling, fallback nome+valor; reuso bloqueado; chargebacks vinculados à origem; 7 transações sem ID para análise manual.

**Ações visíveis:** atualização, exportação, ajuda, detalhes por situação, conciliação completa, regras, transações e divergências. A imagem não define a operação de confirmar, desfazer ou resolver manualmente um vínculo.

**Dados e dependências:** chaves compostas de transação e evento, identificação de título/pedido/movimento, regras de normalização e unicidade, motivos, candidatos, regras versionadas, ator/data/evidência da decisão. A_Matriz_Mestra menciona também Itaú e Mercado Pago conforme o caso, enquanto a tela destaca apenas Bling × Nuvem: a cobertura precisa ser confirmada.

**Trabalho de backlog:** contrato de matching determinístico conservador; sem escolher candidato aproximado; seleção de candidatos e explicação do resultado; exclusividade de uso do movimento com controle concorrente; revisão manual e eventual desfazimento com permissão e histórico; reprocessamento idempotente.

**Aceite sugerido:** um único par com evidência suficiente pode conciliar; dois candidatos permanecem pendentes/ambíguos; atraso isolado não invalida pagamento; eventos legítimos distintos do mesmo pedido não são deduplicados por order_id; concorrência não permite consumir o mesmo movimento duas vezes; alteração manual mantém antes/depois, motivo, ator e fonte.

### T06 — DRE

**O que aparece:** receita bruta R$512.800; líquida R$468.920; lucro bruto R$196.840 e margem 41,8%; resultado líquido R$74.380. Tabela de receita bruta, deduções, líquida, CMV/CPV, lucro bruto, despesas operacionais, perdas, resultado operacional, resultado financeiro, tributos e resultado líquido. Gráfico de composição; cartões de custos/despesas/perdas/tributos e 12 UNMAPPED. Alertas: classificação ausente, competência de perdas pendente, resultado financeiro em revisão e ausência de conta automática de fechamento.

**Ações visíveis:** atualizar, exportar, ajuda, abrir detalhes dos resultados, DRE completa, composição e pendências por tipo.

**Dados e dependências:** fato original, documento, competência, classificação aprovada e versão, CMV de SKU com evidência temporal, fontes complementares para folha/depreciação/tributos, taxas/fretes/juros, perdas e recuperações vinculadas. L-03, L-04, L-05, L-09, L-13 e L-14 continuam materialmente relevantes.

**Trabalho de backlog:** fechar plano de composição e fontes por linha; aprovar competência/mappings; preservar UNMAPPED sem conta coringa; reconstruir DRE por versão e período; indicar completude e confiabilidade do resultado; fornecer memória por linha até o documento; evitar dupla redução por perda já registrada no Bling.

**Aceite sugerido:** tabela reconcilia de receita bruta a resultado líquido; fato só entra na linha aprovada e uma única vez; linha obrigatória sem fonte permanece visível e indisponível; custo histórico não é substituído silenciosamente pelo custo atual; perdas/recuperações não antecipam dinheiro ou resultado sem evento reconhecido; resultado incompleto não é exibido como confirmado.

### T07 — Balancete

**O que aparece:** débitos R$684.220; créditos R$684.220; saldo R$0; 9 divergências. Tabela com conta, descrição, débito, crédito, saldo e status OK/DIVERGENT/OPEN; grupos contábeis no gráfico. Indicadores de 186 lançamentos, contrapartidas 100% rastreáveis, saldos iniciais OPEN, ausência de ajuste automático e origem Bling+Majucau. Alertas dizem que ausência de saldo inicial não vira zero e diferença não recebe conta automática.

**Ações visíveis:** atualizar, exportar, ajuda, detalhes, divergências, balancete completo, grupos, regras e saldos.

**Dados e dependências:** contas e natureza do saldo, débitos/créditos, contrapartidas, saldo inicial oficial, documentos, competência, origem, versão da regra e status. O contrato H17 explicita a estrutura mínima; L-08 permanece sem fonte de saldos iniciais/contrapartidas completas.

**Trabalho de backlog:** mapear contas/contrapartidas e fonte dos saldos iniciais; separar diferença entre movimentos do saldo final da conta; estabelecer convenção de sinal e agrupamento; produzir rastreabilidade por lançamento e período; definir comportamento de fechamento, ajustes autorizados e reabertura, se fizerem parte do escopo.

**Aceite sugerido:** cada entrada possui documento, conta, contrapartida, competência e origem; um balancete equilibrado pode continuar com pendências reais; ausência de saldo inicial não fabrica saldo final zero; diferença não produz ajuste automático; agrupamentos e detalhes reproduzem o mesmo conjunto de entradas.

### T08 — Calculadora de Aplicação

**O que aparece:** lucro líquido elegível R$74.380 (DRE Majucau ago/2026); caixa R$258.420; reserva R$50.000; capacidade R$96.800. Memória contém lucro, menos R$18.400 transferidos, caixa e menos reserva. Condições: fonte DRE e reserva OK; semana crítica OPEN com selo OK; PMR/PMP/PME 28 dias; ciclo financeiro 41 dias; decisão OPEN. Investimentos: transferido R$18.400, rendimento R$2.480 (+27,8% sobre R$1.940), posição R$385.400, um investimento CDI Plus e extrato 15/08. Aviso explícito de que a tela não executa transferências/aplicações.

**Ações visíveis:** ver fórmula, cálculo/memória, fontes, regra, parâmetros/decisão, histórico de transferências/rendimentos, extratos e aplicações. Não há exportar; T16 reforça ausência de exportação na T08.

**Dados e dependências:** DRE e janela elegível, fonte/semântica de valor já investido, fluxo e semana/posição crítica, reserva vigente, posição de investimento, saldo transferível e qualidade temporal. Depende de C-02/C-03/L-07, além das dependências de DRE e CDI Plus.

**Trabalho de backlog:** aprovar inputs e regra, distinguir máximo analítico e executável hoje, criar memória reprodutível com versão e fontes; bloquear resultado definitivo com input crítico OPEN/UNAVAILABLE/DIVERGENT/STALE segundo política aprovada; manter diagnósticos de prazos separados da fórmula; definir navegação de Aplicações cuja imagem não foi fornecida nesta subtarefa.

**Aceite sugerido:** mesma versão de dados/regras reproduz o cálculo; lucro já líquido de investimento não sofre segunda subtração; nenhum clique transfere dinheiro ou grava aplicação; input crítico ausente bloqueia, não vira zero; T08 não exporta; indicadores diagnósticos não alteram fórmula sem regra; limite final segue somente fórmula aprovada, nunca valor ilustrativo R$96.800.

## Divergências verificáveis e ambiguidades importantes

| ID | Evidência | Consequência para o backlog |
|---|---|---|
| FIN-01 | T02: R$112.540−R$50.000=R$62.540, mas folga exibida R$42.540. R$91.870−R$50.000=R$41.870, mas folga R$21.870. Ambas as folgas correspondem a reserva de R$70.000. | Definir fórmula/base de folga e única reserva vigente; corrigir dados de exemplo após decisão. |
| FIN-02 | T02, posição 16/08: mínimo de 30 dias em 18/09 e de 60 dias em 18/10 excedem respectivas janelas corridas. Gráfico rotulado próximos 30 dias termina 13/09. | Definir data de referência, contagem inclusiva, calendário e horizonte comuns; não tratar datas ilustrativas como lógica. |
| FIN-03 | T02 alerta saldo abaixo do mínimo em 20/09, mas menor saldo de 60 dias é R$91.870 e reserva exibida R$50.000. | Se mesmo universo/data, não conciliam; explicitar filtros/versões ou ajustar exemplo. |
| FIN-04 | T02 recebido no mês R$301.240; T03 R$214.560, ambos até 16/08. | Não é erro provado sem universo: confirmar se T02 inclui outras entradas/contas e nomear escopos. |
| FIN-05 | T03: meios de pagamento mostrados somam R$310.560, contra total R$367.890; diferença R$57.330. | Confirmar outros meios/subconjunto ou exigir reconciliação integral; gráfico de status, por sua vez, soma corretamente R$367.890. |
| FIN-06 | T04: divergências no cartão R$14.880 e na rosca R$14.700. A rosca soma exatamente R$412.670. | Padronizar fonte/base e fixture única para cartão e gráfico. |
| FIN-07 | T04: Cacau Premium vence 18/08 e aparece OVERDUE na posição 16/08. | Explicar corte de vencimento/qualidade ou corrigir status ilustrativo. |
| FIN-08 | T03 #5492 CONFIRMED e T05 AMBIGUOUS; #5498 DIVERGENT e T05 MATCHED; #5504 CONFIRMED e T05 DIVERGENT. | Se enums representam dimensões diferentes, separar situação do título e estado da conciliação na interface/contrato; se mesmos estados, corrigir inconsistência. |
| FIN-09 | T03 conciliação 91,3%; T05 1.284/1.358≈94,55%. | Definir denominador (valor, transações, títulos), filtros e arredondamento; não sincronizar por número fixo. |
| FIN-10 | T06: todas as subtrações da tabela chegam corretamente a R$74.380; lucro bruto/receita líquida=196.840/468.920≈41,98%, não 41,8%. | Aprovar denominador e arredondamento da margem; dado gráfico não valida fórmula. |
| FIN-11 | T06 rosca: valores R$468.920, R$272.080, R$64.400, R$31.520 e R$74.380 somam R$911.300; percentuais mostrados 46,8/29,0/12,6/6,7/4,9 não são suas proporções. Receita, despesas e resultado também são grandezas sobrepostas. | Definir pergunta analítica e denominador; qualquer ajuste do gráfico que altere design depende da decisão sobre a fidelidade exigida. |
| FIN-12 | T07 seis linhas já somam R$684.220 em cada lado, consistente com cabeçalhos. Fornecedores: 72.350−81.990=−9.640, mas saldo mostra +9.640; Despesas mostra −16.760. | Pode ser convenção por natureza contábil, não erro conclusivo. Documentar sinais devedores/credores; não somar saldos de naturezas distintas sem semântica. |
| FIN-13 | T07 saldo agregado zero aparece com saldos iniciais OPEN. | Zero é válido como diferença de movimentos, mas não prova saldo final contábil conhecido. Nomear cada indicador e propagar indisponibilidade. |
| FIN-14 | T08 R$96.800 é maior que lucro elegível exibido R$74.380; matriz M-034 limita máximo pelo lucro elegível. Se R$18.400 ainda deve ser deduzido, restariam R$55.980. | Incompatibilidade condicional com a fórmula documentada; confirmar regra e significado dos inputs, sem escolher automaticamente outra fórmula. |
| FIN-15 | T08 memória não mostra como obter R$96.800; exibe semana/decisão OPEN e selo OK, enquanto I94/I96/I97 preveem bloqueio com pendências críticas. | Definir saída bloqueada/indicativa e sua comunicação; um número chamativo não deve fingir cálculo aprovado. |
| FIN-16 | T08 usa DRE ago/2026 para posição ago/2026. M-030 descreve acumulado jan..M−2 e já investido CDI Plus, além do conflito sobre fonte Bling/Majucau. | Aprovar fonte, janela anual, comportamento jan/fev e definição de principal investido; corrigir referência temporal do exemplo. |
| FIN-17 | T08 coloca PMR/PMP/PME em um único valor de 28 dias e ciclo financeiro 41 dias. | São três indicadores sem memória/denominadores explicitados; pedir definição individual e função apenas diagnóstica, conforme I99. |
| FIN-18 | Menu de T02–T08 inclui Balanço Patrimonial; matriz M-057 e teste I169 dizem que a rota foi removida. | Tratar navegação dos mocks como legado a confirmar no inventário geral, sem criar escopo automaticamente. |
| FIN-19 | C-01 confirma conflito parcial permitido × proibido; T04 mostra proibição. C-02 confirma maior despesa × menor saldo; T02 mostra maior despesa. C-03 confirma DRE Bling × Majucau; T08 mostra Majucau. | A imagem escolhe visualmente opções, mas não resolve formalmente os conflitos já registrados na consolidação. |
| FIN-20 | C-02/C-03 e J04/J06 referem a calculadora como Tela 10; imagens atuais usam T10 para Produção e T08 para Calculadora. | Criar identificadores funcionais estáveis e tabela de equivalência de versões; evitar ligar tarefa/aceite à tela errada. |

## Perguntas que precisam de resposta do responsável

1. **Precedência e validade:** a consolidação é a base vigente? As três decisões C-01/C-02/C-03 já foram tomadas depois dessa versão? As menções antigas a Tela 10 devem apontar para a calculadora T08?
2. **Parciais:** com obrigação R$1.000 e movimento R$600, qual saldo/status deve aparecer? Há tolerância de centavos, agrupamento de vários pagamentos, juros/multas, descontos ou estorno de quitação?
3. **Semana crítica:** maior despesa ou menor saldo; qual horizonte e início/fim de semana; qual desempate; qual posição dentro da semana alimenta a calculadora?
4. **Lucro elegível:** DRE Majucau ou Bling; acumulado até M−2 ou outro período; como tratar início do ano, prejuízo e valor já aplicado; R$18.400 é transferência no período, principal acumulado ou posição? Não confundir posição R$385.400 e transferência R$18.400.
5. **Caixa:** quais empresas/contas/bancos/moedas entram; saldo bancário inclui aplicações; que data determina realizado, previsto e projetado; quais recebimentos externos já estão no Bling; qual janela compõe entradas/saídas/mínimos; o que representa a curva do gráfico?
6. **Recebíveis:** fonte real de cada modalidade, pagamentos parcelados e liquidação; origem e periodicidade dos arquivos Nuvem; cálculo do percentual conciliado; como tratar retenções, antecipações, chargebacks, estornos e taxas?
7. **Conciliação:** quais pares são cobertos (Bling/Nuvem/Itaú/Mercado Pago); campos reais únicos; normalização de nome; tolerâncias; correspondência 1:1, 1:N ou N:1; quem confirma/desfaz e qual evidência/motivo deve fornecer?
8. **DRE:** responsável pela aprovação das classificações/competência; fontes de folha/depreciação/tributos/frete/juros/comissões; regra de perdas e recuperações; campo histórico de custo SKU; comportamento quando há linhas/meses incompletos?
9. **Balancete:** fonte e data do saldo inicial; plano de contas e contrapartidas; convenção de sinal; período de competência; escopo gerencial pretendido; regras de fechamento/reabertura e quem homologa resultados?
10. **Investimentos:** layout/arquivo de CDI Plus, atualização da posição, reserva vigente e governança; significado exato de máximo, capacidade e executável hoje; a tela continua estritamente informativa sem exportar/transferir?
11. **UX das sete telas:** formatos e colunas de exportação por tela, detalhe esperado em cada “Ver”, filtros e período, comparação de períodos, dispositivos prioritários, comportamento de indisponível/antigo/incompleto, tradução dos estados técnicos e como compatibilizar clareza/acessibilidade com o design declarado imutável no material?

## Pacotes financeiros de backlog e ordem sugerida

Esta é ordem de dependências, sem compromisso de prazo ou aprovação implícita.

1. **Decisões e contratos financeiros:** resolver conflitos, dicionário de indicadores, cardinalidades de vínculos, fontes reais e critérios de completude. Impacto: impede construir cálculos contraditórios. Evidência de conclusão: decisão escrita + exemplos de entrada/saída + responsáveis.
2. **Fatos financeiros e proveniência:** ingestão segura/idempotente, chaves, dinheiro sem ponto flutuante, datas/competência, versões e estados de qualidade. Impacto: evita duplicidade e permite explicar cada valor. Depende da fundação compartilhada.
3. **Títulos e conciliação:** leitura de receber/pagar, estados, matching e análise contextual. Impacto: define bases confiáveis para caixa e resultados. Aceite cruza T03/T04/T05 e importações.
4. **Caixa e reserva:** séries e indicadores rastreáveis, semana crítica somente com regra aprovada. Impacto: visualiza compromissos e horizonte de liquidez. Aceite cruza títulos, recebimentos externos e saldos.
5. **Classificação, perdas, DRE e balancete:** mappings e fontes aprovados, memória por documento, linha ausente visível e estados de completude. Impacto: resultado e saldo explicáveis. Não inferir custo histórico ou saldo inicial.
6. **Investimentos e T08:** positionamento dos extratos, lucro elegível, máximo/executável, memória e bloqueios. Impacto: torna o cálculo reprodutível. Depende de caixa, DRE, parâmetros e resolução dos conflitos.
7. **Homologação financeira integrada:** cenário conhecido do começo ao fim, reconciliação de todos os cards/listas/gráficos/exports, reversões, importações repetidas, fontes antigas e contingência. Impacto: impede que telas individualmente coerentes contem histórias diferentes.

Nenhum pacote foi desenvolvido ou testado. A evolução de implementação dessas tarefas deve iniciar em 0%/não iniciada até haver evidência de código existente ou execução futura; aprovação de documento e execução de software são marcos distintos.
