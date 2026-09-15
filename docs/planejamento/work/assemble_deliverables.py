import json,re
from pathlib import Path

out=Path('outputs'); data=Path('work/source_extract')
tasks=json.loads(Path('work/backlog.json').read_text(encoding='utf-8'))
book=json.loads((data/'workbook.json').read_text(encoding='utf-8'))
manifest=json.loads((data/'manifest.json').read_text(encoding='utf-8'))
decisions=[]
def decision(i,theme,question,impact,source,status='Em aberto',answer=''):
    decisions.append(dict(id=i,theme=theme,question=question,impact=impact,source=source,status=status,answer=answer))

for i,theme,answer in [
('D-01','Empresa atendida','Somente Majucau. A resposta anterior sobre várias empresas foi corrigida pelo usuário.'),
('D-02','Usuários','Três pessoas da diretoria, cada uma em seu computador.'),
('D-03','Hospedagem','Local, em computador/servidor Windows existente. Demais características ainda desconhecidas.'),
('D-04','Primeira produção','Todos os módulos confirmados do material na primeira entrada em produção; etapas internas são permitidas para organizar o trabalho.'),
('D-05','Conflitos entre fontes','Analisar caso a caso com o usuário. Não existe precedência global da planilha sobre imagens.'),
('D-06','Tecnologias','Monólito com DDD e arquitetura hexagonal; backend Go; frontend JavaScript com React.'),
('D-07','Balanço Patrimonial','Fora da primeira versão; retirar o item dos menus por decisão explícita do usuário.'),
('D-08','Natureza desta entrega','Somente análise e backlog. Não desenvolver o produto nesta etapa.')]:
    decision(i,theme,'Decisão recebida','Orienta o planejamento.','Conversa atual','Respondida',answer)
decision('C-01','Pagamento diferente do título','Qual o tratamento para R$1.000 com movimento de R$600?','Define baixa, saldo e conciliação.','C_Conflitos!A3:L3','Respondida','Marcar divergência e manter o título sem baixa até a análise. Casos complementares serão detalhados em BK-005.')
decision('C-02','Semana crítica','Maior despesa ou menor saldo?','Define seleção da semana no fluxo e na calculadora.','C_Conflitos!A4:L4','Respondida','A semana com o maior total de despesas. A posição de caixa dentro dela ainda precisa ser definida.')
decision('C-03','Lucro elegível','Qual DRE fornece o lucro?','Define fonte do limite da calculadora.','C_Conflitos!A5:L5','Respondida','A DRE calculada pelo Majucau. Período elegível e composição do valor já investido ainda precisam ser detalhados.')

for row in book['D_Lacunas'][2:]:
    n=next(iter(row)); rn=re.sub(r'\D','',n)
    v=lambda col:row.get(col+rn,'')
    ident=v('A');status='Em aberto';answer=''
    if ident=='L-01':status='Planejamento autorizado';answer='Usuário pediu somente análise e backlog. Implementação será trabalho futuro; o bloqueio textual do documento não impede planejar agora.'
    if ident=='L-02':status='Parcialmente esclarecida';answer='GitHub inspecionado e vazio. Falta confirmar se há código/dados técnicos fora do repositório.'
    decision(ident,v('B'),v('F'),v('E'),'D_Lacunas!A'+rn+':G'+rn,status,answer)

questions=[
('Q-01','Referência visual','Qual logo e qual organização de menu devemos adotar quando as imagens diferem? Posso corrigir textos, números e seleção de menu mantendo a estrutura aprovada?','Fixa o padrão visual sem resolver diferenças silenciosamente.','T12/T13 versus demais imagens'),
('Q-02','Módulos sem especificação completa','Quais referências completam T01 Visão Executiva, T11 Compras, Aplicações, Pricing, Logística, Produtos, Simuladores e Relatórios? Os itens citados nos menus são módulos próprios ou atalhos?','Esses itens permanecem no inventário até decisão; nenhuma rota será criada com comportamento inventado.','Menus; A_Matriz_Mestra'),
('Q-03','Máquina Windows','O Windows que hospedará o sistema é este computador ou outro? Qual é a edição e quem pode verificar CPU, RAM, disco e suspensão automática?','Permite dimensionar instalação e processamento.','Resposta do usuário: Windows; demais dados desconhecidos'),
('Q-04','Locais de acesso','Os três diretores acessarão apenas dentro da empresa ou também de casa e de outras redes?','Define rede, endereço e eventual acesso remoto.','Uso em três computadores confirmado'),
('Q-05','Disponibilidade local','O servidor ficará ligado continuamente? Como o sistema deve funcionar durante falta de energia ou de internet?','Fontes externas exigem internet; posição anterior deve ter data e aviso.','Hospedagem local'),
('Q-06','Contas e fontes reais','Quais contas bancárias, depósitos, canais e fontes já são usados: Bling, Itaú, Mercado Pago, Nuvem Pago, Nuvem Envio, CDI Plus, perdas, ajustes e cacau?','O contrato menciona fontes além das quatro ilustradas.','H_Contrato_Dados; M-008; T15'),
('Q-07','Estoque e pedido aplicável','Qual saldo e quais depósitos do Bling entram no estoque? Um pedido insuficiente ou parcialmente recebido ainda exclui o insumo da recomendação de compra?','Define a regra compartilhada de Estoque e Compras.','T09; Q_CEI_Tela_11'),
('Q-08','Situação das metas','Quando um SKU com produção abaixo de 100% deve aparecer Em andamento ou Abaixo da meta? Qual data e situação da OP contam como concluídas?','O mockup não contém regra suficiente para classificar as linhas.','T10'),
('Q-09','Uso dos demonstrativos','A DRE e o balancete serão usados para gestão interna ou precisam seguir também algum processo formal da contabilidade? Quem valida contas, competência, fechamento e correções?','Define fontes, rigor do aceite e participação da contabilidade.','T06/T07'),
('Q-10','Período e limite de aplicação','Mantemos o lucro acumulado de janeiro até dois meses antes e a reserva mínima documentada? O que conta como já investido: aportes, aportes líquidos de resgates ou posição do CDI Plus?','Fonte da DRE já decidida; estes componentes ainda mudam o resultado.','M-030..M-034; T08'),
('Q-11','Vendas para produção projetada','Como a previsão de venda vira quantidade a produzir, sem descontar o estoque pronto: existe lote mínimo, rendimento, perda ou antecipação de fabricação?','A imagem mostra vendas e produção diferentes sem explicar a transformação.','M-047; T12'),
('Q-12','Relação entre metas','As metas de produção da T10 são independentes das metas do cenário da T12? Quem pode editar e como ficam mudanças em meses anteriores?','Evita duas metas conflitantes ou sobrescrita de histórico.','T10/T12/T16'),
('Q-13','Ações visíveis e exportação','O que deve acontecer em Ajuda, Atualizar, Ver detalhes e Exportar em cada tela? A atualização apenas consulta a posição ou também solicita sincronização?','Os mockups mostram controles sem todos os fluxos detalhados.','T02–T07; T13–T15; L-22'),
('Q-14','Permissões da diretoria','Os três diretores terão as mesmas permissões? Quem administra usuários, aprova metas, edita parâmetros e reprocessa arquivos?','Quatro perfis conceituais não exigem criar contas fictícias.','T17; L-26'),
('Q-15','Recuperação','Quanto tempo a Majucau pode ficar sem o sistema e quanto dado pode perder após uma falha?','Escolha de backup e recuperação depende desses limites.','Requisito de produção local'),
('Q-16','Backup e responsável','Há outro disco ou equipamento para backup? Quem cuidará de atualizações, cópias e restauração?','A única cópia não deve depender do mesmo disco do servidor.','Requisito de produção local'),
('Q-17','Prazo e equipe','Existe data desejada para produção e quem fará desenvolvimento, validação financeira e suporte?','Ainda não há base para estimar datas e duração.','Pergunta inicial respondida apenas quanto à hospedagem'),
('Q-18','Dispositivos e linguagem','Qual o menor tamanho de tela usado? O sistema precisa atender celular? Podemos exibir status em português simples mantendo códigos técnicos internamente?','Define comportamento responsivo e ajustes de acessibilidade.','Gaps UX em todas as imagens'),
('Q-19','Conciliação detalhada','Sem ID Bling, quais critérios autorizam o vínculo por nome e valor: datas, tolerância de centavos, parcelas e mais de um movimento? Quem confirma ou desfaz manualmente?','A regra visual Nome+Valor não define todas as condições de unicidade.','T05; M-008'),
('Q-20','Recebimentos e status','Quais estados, datas e bases definem total a receber, recebido no mês, atraso e percentual de conciliação? Há meios além dos três mostrados?','Cartões de meios e percentuais podem representar universos diferentes.','T03/T05'),
('Q-21','Calendário e caixa da semana','Qual saldo da semana de maior despesa será usado: em que dia/horário e com quais contas? Como contar semanas, horizontes de 30/60 dias e desempates?','C-02 foi respondida, mas L-07 e os limites de data permanecem.','T02/T08; L-07'),
('Q-22','Gráfico da DRE','O gráfico deve explicar participação de despesas, formação do resultado ou outra comparação? Qual base deve usar para os percentuais?','Receita, despesas e resultado são valores sobrepostos na rosca atual.','T06'),
('Q-23','Saldos do balancete','Como mostrar saldo devedor/credor e qual universo usar na composição por grupo? O zero do cartão significa diferença entre movimentos ou saldo final conhecido?','Impede interpretar falta de abertura como saldo zero.','T07; L-08'),
('Q-24','Ficha técnica de insumos','Existe composição por produto no Bling ou em arquivo? Qual fonte define quantidades, unidades, rendimento, perdas e vigência?','Necessária para a aba de insumos e OP Projetada.','H_Contrato_Dados!A12:H12; T12'),
('Q-25','Resumo financeiro projetado','Quais linhas deverão aparecer na aba Resumo financeiro da T12 e como devem se relacionar com custos, preço e cenários?','A aba é citada, mas seu conteúdo completo não foi enviado.','T12'),
('Q-26','Alertas e monitoramento','Os avisos ficam somente dentro do sistema ou devem chegar por algum canal? Para estoque, vale a prioridade Atenção do contrato ou Crítico do exemplo visual?','Define comunicação e resolve discrepância concreta de severidade.','T14; T_CEI_Tela_14'),
('Q-27','Indicadores de prazo','Quais fórmulas, fontes e períodos definem prazo médio de recebimento, pagamento, estoque e ciclo financeiro? NCG será um indicador necessário nesta versão?','O mockup combina três prazos em 28 dias sem detalhar as bases e exibe ciclo de 41 dias.','T08; B_Intersecoes!A12:I12')
]
for q in questions:decision(*q)
for d in decisions:
    d['tasks']=', '.join(t['id'] for t in tasks if d['id'] in re.split(r'[,\s]+',t['questions']))
Path('work/decisions.json').write_text(json.dumps(decisions,ensure_ascii=False,indent=2),encoding='utf-8')

scope=[
('T01','Visão Executiva','Sem imagem','Menus','executivescreen'),
('T02','Fluxo de Caixa','Imagem recebida','T02_Fluxo_de_Caixa.png','cashscreen'),
('T03','Contas a Receber','Imagem recebida','T03_Contas_a_Receber.png','receivablesscreen'),
('T04','Contas a Pagar','Imagem recebida','T04_Contas_a_Pagar.png','payablesscreen'),
('T05','Conciliação','Imagem recebida','T05_Conciliacao.png','reconciliationscreen'),
('T06','DRE','Imagem recebida','T06_DRE.png','drescreen'),
('T07','Balancete','Imagem recebida','T07_Balancete.png','trialbalancescreen'),
('T08','Calculadora de Aplicação','Imagem recebida','T08_Calculadora_de_Aplicacao.png','calculatorscreen'),
('T09','Estoque','Imagem recebida','T09_Estoque.png','inventoryscreen'),
('T10','Produção','2 arquivos idênticos','T10_Producao.png + ChatGPT Image 12 de set. de 2026, 20_48_11.png','productionscreen'),
('T11','Compras','Contrato disponível; imagem ausente','Q_CEI_Tela_11','purchasesscreen'),
('T12','Projeções','Imagem e contrato','T12_Projecoes.png + R_CEI_Tela_12','projectionsscreen'),
('T13','Lançamentos e Pendências','Imagem e contrato','T13_Lancamentos_e_Pendencias.png + S_CEI_Tela_13','pendingscreen'),
('T14','Alertas','Imagem e contrato','T14_Alertas.png + T_CEI_Tela_14','alertsscreen'),
('T15','Importações e Integrações','Imagem e contrato','T15_Importacoes_e_Integracoes.png + U_CEI_Tela_15','integrationsscreen'),
('T16','Parâmetros','Imagem e contrato','T16_Parametros.png + V_Mockup_Tela_16 + W_CEI_Tela_16','parametersscreen'),
('T17','Usuários','Imagem e contrato','T17_Usuarios.png + X_Mockup_Tela_17 + Y_CEI_Tela_17','usersscreen'),
('Sem número confirmado','Aplicações','Sem imagem própria','Menu + T08','applicationsscreen'),
('Bloco a localizar','Pricing / preço e margem','Matriz; UI pendente','M-021..M-023','pricing'),
('Bloco a localizar','Logística','Matriz; UI pendente','M-013..M-016','logisticsanalysis'),
('Escopo a detalhar','Produtos, Simuladores e Relatórios','Citações em menus','T12/T13','extrascreens')]
bykey={t['key']:t for t in tasks}
scope=[dict(code=a,module=b,coverage=c,source=d,task=bykey[e]['id']) for a,b,c,d,e in scope]
Path('work/scope.json').write_text(json.dumps(scope,ensure_ascii=False,indent=2),encoding='utf-8')

lines=['# Majucau — backlog de ponta a ponta, versão 0','',
'Esta versão organiza o trabalho até a produção local. É uma base para discussão: as decisões ainda abertas estão identificadas e impedem fechar os detalhes das tarefas dependentes. O produto não foi desenvolvido nesta etapa.','',
'Premissas atuais: uma empresa, Majucau; três diretores; computadores separados; hospedagem local em Windows; todos os módulos confirmados na primeira produção; Go + React JavaScript; monólito com DDD e arquitetura hexagonal. Balanço Patrimonial excluído por decisão do usuário.','',
'## Como acompanhar a evolução','',
f'O catálogo contém {len(tasks)} tarefas, de BK-000 a {tasks[-1]["id"]}. Três levantamentos estão concluídos com evidência. As demais tarefas estão abertas; desenvolvimento do produto: **0%**. A conclusão por quantidade do catálogo provisório é **{100*3/len(tasks):.1f}%**. Esse valor inclui os levantamentos e não representa porcentagem de software pronto.','',
'Na planilha, cada tarefa recebe 100% somente quando a situação é Concluída e existe evidência de aceite. Tarefas em andamento continuam sem crédito de conclusão até atingir o aceite. O peso inicial é 1 para todas: portanto a medida inicial é por quantidade, não por esforço. Quando a equipe estimar, os pesos poderão representar esforço relativo. Subdividir tarefas extensas em entregas verificáveis antes de iniciar.','',
'Evolução geral = soma(peso × conclusão) / soma(pesos). A planilha também separa Planejamento, Desenvolvimento, Qualidade e Produção. Alterar o escopo ou os pesos altera o denominador; registrar a mudança e sua data. O percentual não aumenta apenas por passar tempo ou por o documento afirmar que uma regra foi aprovada. Não há atualização automática de atividade externa nesta entrega.','',
'## Ordem de execução','',
'A numeração identifica tarefas; a execução deve respeitar Dependências. Há trabalho de UX, infraestrutura e testes que pode ocorrer em paralelo. As fases são marcos internos de construção, sem reduzir o escopo da primeira produção pedido pela Majucau. Cada tarefa de desenvolvimento só começa quando seus contratos e decisões materiais estão definidos.','',
'1. Fechar escopo, dados, fórmulas e operação local.','2. Modelar domínios, contratos, navegação e segurança.','3. Preparar projeto, acesso, banco e integrações.','4. Construir motores e interfaces dos módulos.','5. Conferir jornadas completas, dados e cálculos com a Majucau.','6. Ensaiar instalação, backup, restauração e atualização no Windows.','7. Aprovar a versão completa, colocar em produção e acompanhar os primeiros ciclos.','',
'## Critério comum de conclusão','',
'Para tarefas de software, além do aceite específico: regras e permissões validadas no Go; interface coerente com o contrato; estados de erro/ausência/atraso cobertos; rastreabilidade preservada; verificações proporcionais executadas; documentação necessária atualizada; evidência registrada. Os 167 cenários da aba I_Testes são um catálogo previsto e não evidência de testes executados.','',
'Os responsáveis são papéis sugeridos, ainda sem nomes ou alocação. Prazo, esforço e hardware aguardam levantamento. Itens sem especificação completa são tarefas de refinamento, não promessas de implementação imediata.','']
last=''
for t in tasks:
    if t['phase']!=last:
        lines+=['## '+t['phase'],''];last=t['phase']
    lines+=['### '+t['id']+' — '+t['title'],'',
        f'**Situação:** {t["status"]}. **Evolução:** {int(t["progress"]*100)}%. **Tipo:** {t["kind"]}. **Responsável sugerido:** {t["owner"]}.','',
        '**Objetivo:** '+t['goal'],'', '**O que fazer:** '+t['action'],'','**Como fazer:**','']
    lines += [f'{i+1}. {s}' for i,s in enumerate(t['steps'])]
    lines+=['','**Critério de aceite:** '+t['accept'],'','**Impacto:** '+t['impact'],'',
        '**Dependências:** '+(', '.join(t['dependencies']) or 'Nenhuma tarefa prévia; observar decisões aplicáveis.'),'',
        '**Fonte:** '+(t['refs'] or 'Necessidade técnica para entregar e operar o produto descrito pelo usuário.')]
    if t['questions']:lines+=['','**Decisões a acompanhar:** '+t['questions']]
    if t['evidence']:lines+=['','**Evidência:** '+t['evidence']]
    lines+=['']
(out/'MAJUCAU_BACKLOG_DETALHADO_v0.md').write_text('\n'.join(lines),encoding='utf-8')

report=['# Majucau — análise das referências e decisões, versão 0','',
'O material foi lido para planejar o desenvolvimento. Nenhum arquivo original foi alterado e nenhuma funcionalidade, issue, commit ou instalação do produto foi criada. O GitHub foi consultado em modo de leitura.','',
'## Decisões recebidas nesta conversa','']
for d in decisions[:11]:report+=['- **'+d['id']+' — '+d['theme']+':** '+d['answer']]
report+=['','Os três conflitos C-01, C-02 e C-03 foram respondidos. Isso não resolve automaticamente a posição de caixa dentro da semana, o corte temporal do lucro ou a composição do já investido. Planilha e imagens continuam sendo referências analisadas caso a caso. Frases como "aprovado" e "bloqueado" nos documentos descrevem seu estado documental; não comprovam desenvolvimento, teste ou uma nova autorização nesta conversa.','',
'## Material e cobertura','',
'Foram examinados 16 arquivos PNG, correspondentes a 15 imagens únicas, e as 20 abas da planilha. A extração encontrou 845 linhas não vazias, incluindo títulos e cabeçalhos. A matriz contém 58 regras; o catálogo I_Testes contém 167 cenários previstos. O arquivo de imagem com nome ChatGPT é idêntico ao T10, confirmado por SHA-256.','',
'| Tela | Módulo | Cobertura | Tarefa da interface |','|---|---|---|---|']
for s in scope:report += [f'| {s["code"]} | {s["module"]} | {s["coverage"]} | {s["task"]} |']
report+=['','**Documentos apenas citados:** Markdown.md, RC-3 Financeiro DEV_LOCK, RC-3 BUSINESS_RULE_LOCK_WIREFRAMES e RC-2.2 Comercial/Logística. Eles não vieram no pacote. Seus conteúdos completos não foram lidos; as afirmações sobre eles provêm da consolidação fornecida. Também faltam o visual de T01, o visual de T11 e a especificação completa dos módulos/blocos adicionais da tabela.','',
'**GitHub:** o [repositório Majucau](https://github.com/guilhermejacyzin/majucau) está vazio. A [consulta de arquivos](https://api.github.com/repos/guilhermejacyzin/majucau/contents) retornou explicitamente "This repository is empty". O nome padrão configurado é main, ainda sem commit materializado. Não há implementação versionada para auditar. Isso não prova inexistência de trabalho fora desse repositório.','',
'## Arquitetura proposta para o cenário informado','',
'Um dos computadores Windows, ou uma máquina dedicada, hospeda a aplicação. Os três diretores acessam pelo navegador em um endereço da rede. O servidor concentra dados, login, regras e sincronizações. Ainda é preciso confirmar qual computador será o servidor e se haverá acesso fora da empresa.','',
'```mermaid','flowchart LR','  U[Três diretores no navegador] --> R[Interface React JS]','  R --> G[Monólito Go: API e casos de uso]','  G --> D[Domínios: financeiro, operação e planejamento]','  D --> P[PostgreSQL local — proposta]','  G --> I[Adaptadores de integração]','  I --> B[Bling pela internet]','  I --> F[Arquivos controlados]','  G -. avaliação opcional .-> PY[Ferramenta Python local]','  P --> BK[Backup separado do servidor]','```','',
'Arquitetura hexagonal significa manter as regras de negócio separadas de banco, HTTP e formatos de arquivo. DDD organiza o sistema pela linguagem do negócio. Monólito significa uma aplicação principal dividida internamente em módulos, com uma instalação coordenada. A tela não contém a regra financeira oficial: ela consulta o Go.','',
'**Banco proposto:** PostgreSQL local, sujeito à avaliação do Windows e da operação. Valores monetários e quantidades precisam de representação exata e arredondamento aprovado. PostgreSQL documenta `numeric` como tipo exato; no Go e no contrato com JavaScript também será necessário evitar conversões aproximadas para dinheiro. [Tipos numéricos do PostgreSQL](https://www.postgresql.org/docs/current/datatype-numeric.html).','',
'**Integrações:** Bling por adaptador de leitura, tokens no servidor, acesso mínimo, paginação e retomada. O fluxo OAuth exige validar o endereço de retorno para a instalação local. A documentação de limites orienta fracionar consultas históricas; limites serão reconfirmados ao implementar. Nuvem por arquivo é requisito candidato da consolidação, sem criar API adicional por suposição. [Aplicativos Bling](https://developer.bling.com.br/aplicativos), [limites Bling](https://developer.bling.com.br/limites).','',
'**Python:** avaliar para séries temporais e preparação de planilhas após medir necessidade. pandas ou Polars devem ser escolhidos pelos formatos, memória e tempo em amostras reais. NumPy pode apoiar operações numéricas quando o algoritmo precisar. Uma ferramenta Python local controlada pelo Go preserva contratos e versões; não deve criar uma segunda regra financeira ou gravar diretamente nas tabelas dos outros domínios. Streamlit pode servir a experimentos internos, mas a interface oficial continua React. Nenhuma dessas ferramentas adicionais foi escolhida como requisito obrigatório.','',
'**Operação local:** Windows, PostgreSQL e aplicação precisam iniciar corretamente após reinício, sobreviver a falhas de integração e permitir restauração. Hospedagem local não elimina internet para atualizar Bling e sinais externos. Backup inclui banco e arquivos, em destino separado, com ensaio de recuperação. [Backup PostgreSQL](https://www.postgresql.org/docs/current/backup.html).','',
'**Entrega:** como o GitHub é público, a proposta é validar e empacotar sem executar código de contribuições não confiáveis na máquina de produção ou na rede local. A instalação da versão será controlada. O GitHub documenta riscos de executores próprios que alcançam segredos e serviços de rede. [Segurança dos executores GitHub Actions](https://docs.github.com/en/actions/reference/security/secure-use#hardening-for-self-hosted-runners).','',
'## Leitura detalhada de T02 a T08','',
'As observações abaixo descrevem o material recebido. Quando citarem conflito C-01/C-02/C-03 ou Balanço Patrimonial, considerar a decisão atual registrada no início deste documento. Os valores dos exemplos não foram tratados como dados reais da Majucau.','']
fin=Path('work/financeiro.md').read_text(encoding='utf-8')
fin=fin.split('## Inventário funcional por tela',1)[1].split('## Perguntas que precisam de resposta',1)[0]
report+=[fin,'','## Leitura detalhada de T09 a T17 e Compras por contrato','']
ops=Path('work/operacao_ux.md').read_text(encoding='utf-8')
ops=ops.split('## T09 — Estoque',1)[1].split('## Questões transversais',1)[0]
report+=['### T09 — Estoque',ops,'',
'## Pontos transversais que viram tarefas','',
'- Uma origem pode aparecer em várias telas, mas a regra deve ter um único responsável. T13 trata pendências, T14 chama atenção e T15 explica a saúde da fonte.','- Status financeiro, estado da conciliação, qualidade do dado e estado de processamento não são a mesma coisa. Os códigos precisam de rótulos claros e contratos separados.','- Totais precisam declarar população, filtros, período, unidade, fonte e posição. Um recorte não pode parecer total completo.','- Detalhes, formulários, confirmações, estados vazios, falhas, carregamento, sessão expirada e comportamento em telas menores não estão completamente especificados pelos PNGs.','- Referências antigas dizem que Tela 09 não existe e chamam a calculadora de Tela 10. O mapa atual identifica Estoque como T09, Produção como T10 e Calculadora como T08. Essa equivalência será confirmada nos documentos para evitar associação incorreta.','- A origem real de histórico de custos, saldos iniciais e fichas técnicas precisa ser demonstrada com dados. Não substituir por custo atual, zero ou fórmula presumida.','- A conclusão do produto exige todos os módulos confirmados, dados reconciliados, segurança, instalação Windows, backups, restauração, treinamento e operação assistida.','',
'## Perguntas e decisões a acompanhar','',
'As questões abaixo compõem o registro completo para refinamento. Elas não precisam ser respondidas todas de uma vez. Se uma definição não existir, registrar "a definir" e preparar uma proposta com exemplos antes da decisão. Senhas, tokens e chaves não devem ser enviados por este questionário.','']
for d in decisions[11:]:
    report += [f'### {d["id"]} — {d["theme"]}', '', '**Situação:** '+d['status'], '', d['question'], '', '**Por que importa:** '+d['impact'], '', '**Referência:** '+d['source']]
    if d['answer']:report+=['','**Já esclarecido:** '+d['answer']]
    if d['tasks']:report+=['','**Tarefas relacionadas:** '+d['tasks']]
    report+=['']
report+=['## Registro de leitura das abas','', '| Aba | Linhas não vazias | Uso na análise |','|---|---:|---|']
for name,rows in book.items():report += [f'| {name} | {len(rows)} | Requisitos, rastreabilidade e lacunas; conteúdo documental analisado. |']
report+=['','## Identificação dos arquivos recebidos','', '| Arquivo | SHA-256 |','|---|---|']
for m in manifest:report += ['| '+m['name']+' | '+m['sha256']+' |']
(out/'MAJUCAU_ANALISE_E_DECISOES_v0.md').write_text('\n'.join(report),encoding='utf-8')
print('Documentos criados; decisões:',len(decisions),'itens no mapa:',len(scope))
