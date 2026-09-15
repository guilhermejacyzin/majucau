import copy, hashlib, json, re
from pathlib import Path

W=Path('work'); O=Path('outputs')
SRC=Path(r'sources/private/fechamento-questoes.txt')
LABEL='Fechamento das questões ainda tratadas como abertas — 13/09/2026'
tasks=json.loads((W/'backlog_v1.json').read_text(encoding='utf-8'))
decisions=json.loads((W/'decisions_v1.json').read_text(encoding='utf-8'))
original=copy.deepcopy(tasks); olddec=copy.deepcopy(decisions)
by={t['key']:t for t in tasks}; db={d['id']:d for d in decisions}
for t in tasks:
    if t['status']=='Aguarda decisão': t['status']='Não iniciada'

def edit(key,section,**fields):
    t=by[key]; t.update(fields)
    t['refs'] += '; Fechamento vigente 13/09/2026, '+section

def extra(key,section,step,accept=''):
    t=by[key]; edit(key,section,steps=t['steps']+[step],accept=t['accept']+(' '+accept if accept else ''))

def link(key,*ids):
    t=by[key]; t['questions']=','.join(dict.fromkeys([i for i in t['questions'].split(',') if i]+list(ids)))

edit('baseline','integral',action='Registrar o fechamento vigente e sua precedência sem reabrir regras de negócio.',
     steps=['Preservar Majucau única empresa, três diretores, Windows local, Go, React JS e DDD hexagonal monolítico.',
            'Aplicar o fechamento mais recente aos cálculos, fontes, compras, produção, permissões e operação; mockup aprovado governa o visual.',
            'Registrar que o usuário confirmou somente atualização do backlog nesta rodada. Planejar implementação, mapeamento e validação como tarefas futuras.'],
     accept='Referência vigente, mudanças e atividades futuras estão rastreadas; regra definida não é apresentada como nova pergunta à Majucau.')
extra('inventory','integral','Catalogar este segundo fechamento, seu hash e os trechos superados nas versões 0/1; preservar todos os originais.')
edit('missing','§5',title='Localizar e versionar os mockups e contratos aprovados',
     goal='Disponibilizar à implementação as referências visuais oficiais de cada tela.',
     action='Localizar a última versão aprovada/congelada, preservando escopo e navegação.',
     steps=['Identificar no material entregue a versão oficial de cada tela e vincular seu arquivo; não usar ordem de nome/data como prova de aprovação.',
            'Registrar como Referência a localizar o arquivo aprovado que não esteja no pacote, incluindo T01, T11 e Aplicações. Isso não reabre a aprovação da tela.',
            'Manter a exclusão explícita de Balanço Patrimonial e os módulos confirmados; resolver equivalência de números antigos pelo módulo.'],
     accept='Cada tela tem referência oficial identificada ou localização documental rastreada; não há redesenho, reorganização de menu ou falsa alegação de ter recebido arquivo ausente.')
edit('partial','§8.6–8.8; C-01 mantida',steps=[
    'Aplicar valor diferente = divergência e título sem baixa até análise, inclusive diferença de centavos sem tolerância inventada.',
    'Derivar movimento financeiro efetivamente registrado do Bling, preservando estorno, desconto e tarifa conforme a origem; não dar baixa no ERP pelo Majucau.',
    'Criar exemplos de aceite e trilha de conferência; candidato ambíguo não é conciliado automaticamente.'],
    accept='Título e movimento mantêm estados distintos e rastreáveis; a divergência segue C-01, sem pedir novamente a regra financeira.')
edit('investmentpolicy','§§1–2',title='Especificar tecnicamente a regra fechada da calculadora',
     action='Transcrever o contrato consolidado da calculadora e os estados das entradas.',
     steps=['Registrar lucro elegível pela DRE exibida no Majucau, agora reproduzida do Bling por competência; preservar janela consolidada jan..M−2 e dedução única de Já Investido.',
            'Registrar Já Investido = aportes de principal − resgates de principal; rendimentos separados. Reserva = MAX(50.000; maior reserva aprovada vigente), conforme M-034 mantida pelo fechamento.',
            'Registrar Máximo = MAX(0; MIN(Caixa Semana Crítica − Reserva; Lucro Elegível)); Executável = MIN(Máximo; saldo transferível do Itaú). Semana crítica continua a de maior total de despesas.',
            'A porta da calculadora recebe os resultados financeiros consolidados, com período, fonte e qualidade. Ela não recebe bandeira, MDR, parcelamento ou tarifas de cartão.',
            'Localizar o contrato/campo consolidado de Caixa Semana Crítica e detalhes de fronteira temporal na documentação vigente; não escolher saldo inicial/final/mínimo por conveniência. É rastreabilidade técnica, sem nova pergunta de negócio.'],
     accept='Contrato financeiro fechado transcrito com referências e exemplos técnicos; nenhuma entrada ausente é inventada. A execução do cálculo depende das entradas comprovadas, sem exigir integração pronta para concluir a especificação.')
edit('accountingpolicy','§2',title='Especificar a reprodução da DRE e do balancete do Bling',
     goal='Preservar a competência, classificação e os valores corretos da origem.',
     action='Registrar contratos de leitura dos demonstrativos existentes no Bling.',
     steps=['Adotar DRE por competência do Bling como fonte primária e suas classificações existentes como referência; não criar classificação gerencial paralela.',
            'Usar a base contábil/financeira do Bling para o balancete do projeto; mapear contas, saldos, débitos, créditos e contrapartidas quando retornados.',
            'Planejar comparação de totais com o Bling por período e conta. Campo específico não acessível pela API deve aparecer como Dado indisponível no Bling.',
            'Preservar competência dos rendimentos; aplicação e resgate de principal não são automaticamente receita/despesa. Não alterar resultado por fontes externas fora de seu escopo.'],
     accept='Especificação traduz a origem para o contrato de leitura, sem nova estrutura contábil, inferência de competência, preenchimento de ausência ou reabertura financeira.',owner='Arquitetura + Integrações')
edit('opsrules','§§3–4',title='Especificar necessidade líquida e acompanhamento da produção',
     action='Traduzir as regras funcionais fechadas em contratos e exemplos verificáveis.',
     steps=['Usar estoque utilizável, comprometimento aplicável, demanda, BOM, produção válida, saldo pendente dos pedidos e chegada em tempo.',
            'Necessidade líquida desconta somente cobertura efetiva: 100 kg − 20 kg − 30 kg = 50 kg. Pedido parcial não exclui o item e pedido tardio não cobre a necessidade daquela data.',
            'Separar estado temporal (imediata/futura) de cobertura (coberta/parcial/não coberta); priorizar risco operacional e impacto de produção, com critérios explicáveis.',
            'Meta atingida exige realizado >= meta; em andamento e abaixo da meta dependem do avanço esperado no período/plano. Sem granularidade temporal, não inventar atraso ou tolerância.'],
     accept='Exemplos cobrem pedido parcial, tardio, cancelado, atendido e ausência de dado; custo/tarifa/cacau não determinam quantidade necessária.')
edit('forecastpolicy','§4',title='Especificar forecast, metas e planejamento de reposição',
     action='Separar previsão baseada em vendas, meta comercial, produção e trajetória de estoque.',
     steps=['Usar histórico confiável de 2025 em diante com cobertura real por SKU; não completar 24 meses nem converter ausência em venda zero.',
            'Manter meta de faturamento como cenário comercial; não distribuí-la automaticamente em unidades por SKU. Meta direta por produto só é entrada quando cadastrada explicitamente.',
            'Normalizar tecnicamente a expressão de reposição de §4.5 como demanda + estoque-alvo − estoque utilizável − produção válida ainda não incorporada ao estoque. O marcador de lista antes de estoque-alvo no texto não é interpretado como multiplicação; registrar esta normalização.',
            'Calcular cronologicamente, sem descontar duas vezes produção que já integra o saldo. Estoque-alvo, produção válida e demais entradas exigem dados efetivos; ausência não assume zero.',
            'Propor e testar modelos, métricas e critérios técnicos com histórico real; registrar limitações. T01 e T10 compartilham a métrica; cenário T12 não sobrescreve automaticamente meta operacional.'],
     accept='Contratos separam fatos, previsão e planejamento; fórmula técnica tem unidade e período coerentes; receita alvo não cria demanda física fictícia.',owner='Arquitetura + Dados')
edit('filecontracts','§8',steps=[
    'Manter as três estruturas do Drive somente de leitura e contratos específicos: perdas, agenda Nuvem e Nuvem Envio.',
    'Mapear amostras reais, colunas, IDs, sinais, período e versões; registrar processamento e rejeições internamente, sem mutar o Drive.',
    'Definir a detecção técnica de novas versões e idempotência por arquivo, hash, chave de negócio e lote.',
    'Aplicar D-1 à análise/conciliação normal de movimentos realizados. Frequência de sincronização é configuração técnica distinta; agenda futura preserva seus vencimentos.'])
edit('cocoapolicy','§3.4',steps=[
    'Relacionar fornecedores pequenos produtores às localidades efetivamente usadas pela Majucau; não criar regiões fictícias.',
    'Refinar tecnicamente fontes de preços/clima, unidades, datas, cobertura e comparabilidade, mantendo contexto separado da necessidade operacional.',
    'Especificar ausência/atraso dos sinais sem impedir o funcionamento do motor principal de compras; não ampliar as estruturas do Drive por inferência.'],
    accept='Contrato separa necessidade, histórico/preço atual e sinais externos; refinamento dos sinais não é bloqueio da recomendação operacional.',owner='Engenharia de Dados')
edit('actionpolicy','§§5–6',title='Especificar ações e exportações já previstas',action='Mapear os controles aprovados a casos de uso, dados e capacidades.',
     steps=['Preservar botões, navegação e exportações dos mockups/contratos congelados; não pedir nova aprovação da função existente.',
            'Definir tecnicamente contratos de detalhes, atualização e downloads, preservando filtros, período, snapshot, dados e Excel onde especificado.',
            'Resolução de pendência registra responsável, data, alteração, origem e justificativa; resolver exige tratar a causa e não apenas esconder o alerta.'],
     accept='Cada controle aprovado tem caso de uso e capacidade associados; exportação reproduz a tela. Ausência de arquivo de referência é localização documental, sem função inventada.')
edit('accesspolicy','§6',title='Especificar perfis, capacidades, vigência e auditoria',
     action='Definir tecnicamente autorização por usuário, perfil e capacidade, com menor privilégio.',
     steps=['Separar visualizar, exportar, alterar metas, alterar parâmetros, resolver pendências, reprocessar, administrar usuários e consultar auditoria; não hardcodar nomes de diretores.',
            'Restringir criação, perfil, ativação e inativação à capacidade administrativa; usuário comum não promove a própria autoridade.',
            'Especificar login, sessão, revogação e provisionamento com atribuição explícita de perfis na implantação, sem tornar nomes pessoais requisitos de arquitetura.',
            'Alterações usam vigência e histórico antes/depois. Período fechado só muda por reprocessamento explícito, autorizado e auditado.'],
     accept='Matriz por capacidade e protocolo de atribuição definidos tecnicamente; consulta não implica alteração e ações sensíveis conservam evidência.',owner='Arquitetura + Segurança')
edit('localrequirements','§7',title='Inventariar o Windows e propor a operação local',
     action='Levantar o host existente e documentar solução operável por serviços.',
     steps=['Verificar tecnicamente host, Windows, CPU, RAM, disco, energia e rede; não inventar características não acessíveis.',
            'Propor serviços independentes de login humano, inicialização controlada, banco, portas, dependências, atualização e tratamento de falta de internet.',
            'Propor backup, retenção, recuperação, monitoramento, suporte e acesso autenticado. Acesso externo permanece condicionado a habilitação explícita e canal seguro.',
            'Registrar requisitos de funcionamento para validação da Majucau na implantação; obter permissões/acessos materiais quando necessários, sem devolver decisões de infraestrutura como perguntas financeiras.'],
     accept='Inventário real e proposta técnica documentados; limitações de acesso/hardware são evidenciadas, sem depender de sessão de funcionário.',owner='Infraestrutura + Segurança')
edit('deliveryplan','§7',steps=['Propor esforço, sequência e responsáveis por papel conforme capacidade técnica conhecida; não prometer prazo inventado.',
     'Preservar IDs, subdividir entregas quando necessário e registrar pesos de esforço apenas com base explícita.',
     'Planejar marcos internos até a primeira produção e critérios de aceite; refinamento dos sinais de cacau não impede testar/usar o motor principal de compras.'])
extra('domainmap','§§1–4','Separar compras por quantidade, apropriação de custo e inteligência de cacau. Resultados reproduz Bling; calculadora recebe apenas contratos financeiros consolidados. Uma métrica tem um dono e um cálculo para todos os consumidores.')
extra('dataquality','§§4.3,8.7–8.8','Aplicar uma função e um conjunto de dados/snapshot por métrica em card, tabela, gráfico, detalhe e exportação. PENDING_RECONCILIATION fica fora de indicadores que exigem reconciliação.', 'Ausência não vira zero; ausência de campo opcional não invalida componentes independentes comprovados.')
edit('designsystem','§5',title='Transcrever o design congelado em componentes',
     action='Implementar especificação técnica fiel ao mockup aprovado, sem redesenho.',
     steps=['Capturar estrutura, blocos, cabeçalho, menu, cards, gráficos, tipografia, cores, espaçamentos, botões e hierarquia da última versão comprovadamente aprovada.',
            'Separar contrato visual do funcional: números demonstrativos não são regra; fonte e cálculo seguem o fechamento vigente.',
            'Preparar componentes reutilizáveis mantendo posição e hierarquia; erro futuro de implementação se corrige no código, sem reabrir a tela congelada.'],
     accept='Especificação preserva o visual oficial; diferenças entre arquivos exigem localizar a versão aprovada, sem escolher versão antiga ou redesenhar por conveniência.')
edit('uxstates','§5',steps=['Especificar carregamento, vazio, falha, indisponibilidade e acesso negado dentro dos blocos e padrões aprovados.',
     'Implementar teclado, foco, rótulos, zoom e leitura de tabelas preservando hierarquia e disposição.',
     'Conferir referência congelada; não autorizar mudança estrutural ou responsiva que redesenhe a tela por iniciativa técnica.'])

extra('apicontract','§1','Definir portas distintas para demonstrativos Bling, caixa projetado e disponibilidade financeira. A porta da calculadora recebe lucro elegível, caixa crítico, reserva, aportes líquidos e saldo transferível, com versão/qualidade; detalhes de cartão ficam nos módulos de origem.')
extra('authorization','§6','Controlar capacidades por usuário/perfil, inclusive reprocessar período fechado e solicitar inclusão do dia corrente na conciliação; nunca vincular autoridade ao nome de uma pessoa.')
extra('accessaudit','§6.8','Persistir usuário, data/hora, ação, entidade, antes/depois, origem e identificador da operação sensível; usuário comum não altera logs.')
extra('parameterversions','§6.3–6.7','Versionar parâmetros e metas por vigência, preservando os períodos encerrados. Alteração atual não recalcula mês fechado; reprocessamento anterior exige ação explícita, autorizada e auditada, com versão anterior consultável.')
edit('parametersscreen','§6',steps=['Preservar o layout T16 e mostrar valores, origem, estado e vigência reais.',
    'Permitir alteração somente pela capacidade correspondente, validando tipo/unidade e registrando usuário, valor anterior/novo e data.',
    'Aplicar a vigência registrada, mantendo períodos encerrados; reprocessamento retroativo passa por ação explícita e auditada.',
    'Apresentar fontes atuais Bling e três estruturas Drive; eliminar a informação funcional superada de investimento em pasta externa.'],
    accept='Mudanças autorizadas preservam histórico e vigência; usuário de leitura não altera regras; a tela conserva o design congelado.')
extra('provenance','§6.6–6.8','Preservar snapshots e registros de fechamento de períodos. Reprocessamento explícito cria nova versão vinculada à anterior com ator, motivo, origem e operação; não apagar o histórico encerrado.')
extra('jobengine','§8.11','Separar agendamento da coleta, data da competência e corte de análise D-1. Movimentos de hoje podem ser ingeridos e armazenados sem compor a conciliação normal; inclusão requer solicitação expressa, capacidade e trilha.')
edit('blingcontract','§§2,8',steps=[
    'Mapear autenticação, recursos e campos reais da API para DRE por competência, balancete, saldos/contas, movimentos, custos, produção, BOM, compras, resgates e posição.',
    'Registrar endpoint, campo, amostra protegida, competência, significado, data, limites e correspondência com a informação funcional do Bling; não pedir nova estrutura de contas à Majucau.',
    'Para DRE/balancete, validar totais contra o Bling com mesma competência e posição, preservando classificações. Existência funcional informada não comprova endpoint específico disponível.',
    'Para Conta Caixa, usar API oficial ou alternativa de exportação estruturada oficial de investimentos já permitida, sem scraping nem pasta adicional no Drive.',
    'Campo específico não acessível = Dado indisponível no Bling. Produzir evidência técnica por componente, sem inferir, estimar, preencher ausência ou fabricar zero.'],
    accept='Matriz de campos e resultados de validação registrados; nenhuma prova foi executada nesta atualização de backlog. Negócio definido, integração a executar.')
extra('blingadapter','§2','Implementar extração da DRE por competência e base de balancete por recursos reais homologados; expor limitações pontuais com Dado indisponível no Bling, mantendo classificação e competência da origem.')
edit('receivableingest','§8.2,8.5,8.11',steps=[
    'Ler a agenda em 02_nuvem_pago/recebimentos_futuros preservando IDs, nomes, valores, datas e parcelas que existirem; ela representa previsão.',
    'Obter fatos realizados no Bling e manter datas de ocorrência, competência, coleta e corte distintas; análise/conciliação normal usa D-1.',
    'Relacionar previsão e liquidação correspondente para evitar duplicidade. Uma previsão futura não é excluída somente por ter vencimento depois de D-1.',
    'Preservar taxas, juros, chargebacks e estornos comprovados no módulo pertinente, sem enviar lógica de cartão à calculadora.'],
    accept='Venda, direito futuro e dinheiro recebido não duplicam caixa; o corte D-1 não elimina a agenda futura necessária à projeção.')
edit('freshness','§8.11',steps=['Registrar separadamente tentativa/sucesso da coleta, data de referência/competência e corte D-1 da análise.',
    'Definir frequência, timeout, repetição e validade por fonte como configuração técnica; D-1 não é periodicidade de sincronização.',
    'Falha preserva o último snapshot válido, com aviso e data. Não generalizar o caso de 04/09/2026 em regra de sexta-feira ou feriado.'],
    accept='Coleta recente não mascara referência antiga; sem primeira carga, componente é indisponível. Agenda futura permanece projetada e não é cortada como movimento realizado.')
edit('matchengine','§8.6–8.8,8.11',steps=['Priorizar identificador inequívoco do Bling; na ausência, identificação/nome + valor, com normalização conservadora documentada e chaves reais homologadas.',
    'Não conciliar automaticamente quando houver mais de um candidato; bloquear reutilização e preservar eventos distintos.',
    'Aplicar D-1 aos movimentos realizados da análise normal; dia corrente exige solicitação expressa e auditada.',
    'Manter PENDING_RECONCILIATION fora dos indicadores que exigem reconciliação, preservando dados e origem; pendência não significa valor zero.'],
    accept='Identificador prioritário e fallback seguem regra definida; ambiguidade vira conferência. Valor diferente em A Pagar segue C-01, sem tolerância inventada. Nenhuma regra especial fixa de sexta-feira é criada.')
extra('manualmatch','§8.11','Implementar solicitação expressa para incluir movimentos do dia corrente, com capacidade, escopo, solicitante, data e auditoria; a exceção não altera silenciosamente a rotina padrão ou períodos fechados.')
edit('receivables','§8.5–8.11',steps=['Separar valor bruto/líquido, vencimento, liquidação, canal e estado de reconciliação conforme origem.',
    'Calcular previsto, aberto, realizado e atraso no contexto temporal correspondente. Movimentos realizados seguem D-1, salvo solicitação expressa do dia corrente.',
    'Quando a liquidação correspondente integrar o universo realizado, retirar sua participação duplicada na previsão desse mesmo cálculo; preservar vínculo/histórico.',
    'Indicadores confirmados exigem dados reconciliados; manter pendências visíveis separadamente, sem convertê-las em zero.'])
edit('cashengine','§§1,8.5–8.11',steps=[
    'Usar saldos e movimentos do Bling com o mesmo corte/posição; análise normal dos movimentos usa D-1. Não combinar saldo de hoje com movimento de ontem como se fossem uma posição conciliada.',
    'Incluir agenda futura Nuvem como previsão, preservando vencimentos; realizado correspondente substitui sua participação futura no universo aplicável, sem dupla contagem.',
    'Separar consolidado comprovado de saldo transferível Itaú. Transferências internas não produzem nova receita ou aumento líquido consolidado.',
    'Calcular série temporal, horizonte, reserva e disponibilidade com o contrato consolidado; propagar ausência e pendência de reconciliação somente aos indicadores dependentes.'],
    accept='Trajetória de caixa é rastreável e consistente por corte; a calculadora consome disponibilidade consolidada sem conhecer regras de cartão. Nenhum saldo desconhecido é zero.')
extra('criticalweek','§1','Selecionar a semana de maior total de despesas e vincular Caixa Semana Crítica ao contrato vigente do fluxo. Localizar a posição oficial nos artefatos/campos; não eleger inicial/final/mínimo por conveniência técnica.')
edit('financialmapping','§2',title='Preservar contas, classificações e competência do Bling',
    goal='Traduzir recursos da API sem criar uma contabilidade paralela.',
    action='Implementar correspondência técnica entre campos/contas da origem e contratos de leitura.',
    steps=['Mapear IDs, contas, categorias, natureza, hierarquia, competência e linhas existentes no Bling.',
           'Preservar classificação e competência da origem; correspondência técnica não reclassifica receita, custo, perda, juros, frete ou tributo.',
           'Campo sem correspondência ou não retornado gera diagnóstico e Dado indisponível no Bling, nunca conta coringa.',
           'Versionar a correspondência técnica e preservar período fechado; mudança de integração não altera histórico silenciosamente.'],
    accept='Cada valor apresentado chega ao registro e classificação do Bling; não existe mapa gerencial concorrente nem troca de competência por inferência.',
    deps=['blingadapter','accountingpolicy','provenance'])
edit('losses','§§2,8.14',steps=['Registrar eventos de perdas operacionais e recuperações de sua fonte específica, separando ocorrência, confirmação e liquidação.',
    'Vincular fatos correspondentes do Bling; rendimentos, perdas ou recuperações financeiras respeitam classificação e competência da origem.',
    'Não acrescentar nova receita/despesa à DRE reproduzida nem baixa de caixa por existir um evento no arquivo.',
    'Manter perdas operacionais separadas de yield, scrap e necessidade industrial de insumos.'],
    accept='Mesmo evento não reduz resultado duas vezes; ressarcimento sem recebimento comprovado não vira caixa. Diagnóstico externo não reescreve a DRE oficial.')
edit('dreengine','§2',title='Obter e reproduzir a DRE por competência do Bling',
    goal='Apresentar no Majucau a demonstração correta da origem.',
    action='Consumir os recursos oficiais de DRE e preservar seu significado.',
    steps=['Obter linhas, subtotais e resultado líquido por competência, com os recursos reais homologados do Bling.',
           'Preservar contas, categorias, sinais, competência e valores; cálculo de apresentação não reconstrói classificação paralela a partir de vendas/arquivos.',
           'Conferir linhas e totais contra o próprio Bling, guardando consulta, período, posição e evidência.',
           'Campo específico não obtido aparece como Dado indisponível no Bling. Preservar snapshot fechado e bloquear alteração retroativa silenciosa.'],
    accept='DRE no Majucau reproduz a DRE Bling na mesma competência e posição; não estima nem completa linhas com fatos fictícios. Custo SKU e arquivos externos podem apoiar diagnósticos, sem constituir outro motor de resultado.',
    deps=['blingadapter','financialmapping','provenance','dataquality'])
edit('drescreen','§§2,5',steps=['Preservar layout T06 e apresentar competência, contas, valores e resultado vindos do contrato Bling.',
    'Usar a mesma camada de leitura/cálculo para card, tabela, gráfico, detalhe e exportação, com base de percentual explícita.',
    'Apresentar ausência de campo como Dado indisponível no Bling no espaço previsto; não copiar números ilustrativos ou criar segunda classificação.'])
edit('trialbalanceengine','§2',title='Reproduzir o balancete a partir da base do Bling',
    action='Extrair e compor os dados contábeis/financeiros existentes no Bling.',
    steps=['Mapear contas, abertura, débitos, créditos, contrapartidas e natureza dos saldos conforme recursos reais disponíveis.',
           'Preservar competência/classificação e vínculos da origem; não solicitar nova estrutura contábil à Majucau.',
           'Conferir totais e saldos contra o próprio Bling para o mesmo período e posição.',
           'Elemento não obtido pela API = Dado indisponível no Bling; não estimar partida, criar conta de fechamento ou transformar ausência em zero.'],
    accept='Balancete reproduz a informação Bling; ausência pontual fica explícita, sem demonstração fictícia ou classificação paralela.',
    deps=['blingadapter','financialmapping','provenance','dataquality'])
extra('trialbalancescreen','§2','Mostrar natureza, saldos e contas provenientes do Bling, preservando competência e campo indisponível; cartão e tabela não tratam diferença entre movimentos como prova de saldo final conhecido.')
extra('investmentledger','§2','Principal aplicado/resgatado não vira automaticamente receita/despesa. Rendimentos mantêm a classificação e competência da origem Bling; não há segunda classificação gerencial.')
edit('investmentcalculator','§§1–2',steps=[
    'Consumir contratos financeiros consolidados de resultado elegível, disponibilidade, caixa crítico, reserva e principal líquido; DRE no Majucau reproduz Bling por competência.',
    'Aplicar a janela e fórmulas transcritas em BK-006, deduzindo principal líquido uma única vez e limitando o executável ao saldo transferível Itaú.',
    'Propagar corte, versão e qualidade das entradas; detalhes de bandeira, MDR, taxas e parcelas não integram este domínio.',
    'Expor componentes comprováveis e indicar ausência dos demais sem zero fictício; nenhuma transferência ou aplicação é executada.'],
    accept='Memória de cálculo reproduz o contrato consolidado com entradas verificáveis; falta apenas da posição total não bloqueia componentes independentes. Não reabre critérios de negócio.')
edit('pricing','§3.3',steps=['Manter valor bruto, custo variável, tarifa/custo fixo, critério de rateio e custo final apropriado por produto em componentes rastreáveis.',
    'Usar a apropriação fornecida pelo Bling quando existente; ausência de informação suficiente não autoriza distribuir por um critério inventado.',
    'Garantir que a tarifa fixa de uma operação não seja multiplicada integralmente por cada item; reconciliar soma apropriada com custo original.',
    'Preservar análise de preço/margem e layout aprovados, separada do motor de quantidade de compras; não gravar preço no Bling.'],
    accept='Apropriação não duplica custo fixo; parcela indisponível é declarada. Necessidade operacional de compra funciona independentemente de ausência de rateio/custo econômico.')
extra('inventoryrules','§3','Distinguir saldo atual, comprometimento, mínimo cadastrado, necessidade futura e cobertura por pedido. Existência de pedido não comprova cobertura completa; universos distintos de T09/T11 não têm contagens forçadas a coincidir.')
edit('productiontargets','§§4,6.4',steps=['Armazenar meta explícita por produto/período como planejamento, separada de fato Bling e de forecast.',
    'Registrar usuário, data/hora, período, valor anterior/novo e versão, com capacidade correspondente; não apagar histórico.',
    'T01 e T10 usam o mesmo registro e métrica para resumo/detalhe. T12 preserva cenário/forecast; dependência técnica não sincroniza conceitos diferentes automaticamente.'])
edit('productionrules','§4.1–4.4',steps=['Usar produção efetivamente informada pelo Bling por produto, unidade, período e estado; ausência não vira zero.',
    'Calcular uma vez realizado, meta, percentual e status para card, tabela, gráfico e Produzido Geral na executiva.',
    'Meta atingida: realizado >= meta correspondente. Em andamento/abaixo: comparar com avanço esperado comprovado do plano no ponto temporal.',
    'Sem granularidade temporal suficiente, informar avaliação de atraso indisponível, mantendo componentes conhecidos; não assumir progresso linear, tolerância ou atraso por <100%.'],
    accept='Mesma métrica e mesmo universo produzem os mesmos resultados em T01/T10; quantidades incompatíveis não são somadas e produção ausente permanece desconhecida.')
extra('productionscreen','§4.2–4.4','Compartilhar a métrica canônica com Produzido Geral da T01; apresentar quantidade/meta/realizado/progresso/status no layout aprovado, sem copiar números do mockup.')
edit('purchaseeligibility','§3.1–3.2',title='Avaliar necessidade e cobertura temporal das compras',
    action='Identificar necessidade por insumo/data e cobertura efetiva de estoque e pedidos.',
    steps=['Consumir demanda bruta de insumos da produção/BOM, estoque utilizável, comprometimento aplicável e pedidos pendentes.',
           'Considerar somente quantidade ainda pendente do pedido cuja chegada é válida antes da necessidade; recebida/cancelada não é novamente contada como pendente.',
           'Manter pedido parcial e tardio visíveis, sem excluir o insumo da avaliação. Classificar cobertura integral, parcial ou inexistente e horizonte imediato/futuro.'],
    accept='Pedido aberto não elimina item automaticamente. T09 e T11 coincidem somente em métricas realmente equivalentes, com mesmo universo e posição; recomendação não cria pedido.',
    deps=['inventoryrules','opsrules','inputprojection'])
edit('purchasesizing','§3.1–3.3',title='Calcular necessidade líquida e prioridade operacional de compra',
    action='Calcular quanto comprar e por que aquele insumo tem prioridade.',
    steps=['Necessidade líquida = necessidade bruta − estoque utilizável − cobertura pendente que chega em tempo; necessidade não positiva gera nenhuma compra adicional e excedente permanece visível.',
           'Reproduzir 100 kg − 20 kg − 30 kg = 50 kg; pedido tardio não reduz cobertura na data de necessidade e não pode esconder ruptura.',
           'Priorizar risco imediato de ruptura e impacto na produção usando cobertura, lead time, demanda e datas disponíveis; documentar ordenação e motivo, sem pontuação/tolerância arbitrária.',
           'Manter custo, tarifa, rateio e inteligência cacau separados da quantidade; ausência desses custos não bloqueia recomendação operacional comprovável.'],
    accept='Quantidade e prioridade são explicáveis por insumo/data; pedidos parciais/tardios não mascaram necessidade, cobertura não é contada duas vezes e nenhuma ordem é criada no Bling.',questions='F-04,F-05,F-06')
edit('cocoasignals','§3.4',condition='Fontes externas específicas habilitadas após refinamento técnico; não bloqueia compras operacionais.',
    steps=['Relacionar sinais somente a localidades reais dos fornecedores da Majucau e registrar origem, unidade e data.',
           'Separar necessidade operacional, preço histórico/atual e contexto de preço/clima; não criar região ou comparação fictícia.',
           'Refinar e validar fontes externas independentemente do motor principal; manter indisponibilidade explícita quando um sinal não estiver habilitado.'],
    accept='Sinais habilitados são rastreáveis e comparáveis; sinal ausente não impede cálculo principal de compra e nenhum sinal executa pedido.')
edit('purchasesscreen','§§3.4,5',action='Construir a tela conforme referência aprovada, com recomendação operacional e contexto separado.',
    steps=['Apresentar necessidade, quantidade, cobertura, prazo e prioridade pelo contrato funcional vigente nos blocos aprovados.',
           'Preservar bloco de inteligência do cacau com componentes disponíveis; ausência/refinamento de fonte externa não bloqueia motor principal.',
           'Manter exportação aprovada da recomendação, com mesmos filtros e snapshot da tela; não criar pedidos no Bling.'],
    deps=[k for k in by['purchasesscreen']['deps'] if k!='cocoasignals'])
edit('targetscenario','§4.6',title='Implementar forecast e metas explícitas como planejamento distinto',
    action='Versionar estudo por SKU, meta financeira e metas diretas existentes sem distribuição inventada.',
    steps=['Preservar forecast fundamentado no histórico e identificar o modelo/versão utilizados.',
           'Registrar meta de faturamento como referência comercial/cenário; não convertê-la automaticamente em unidades por SKU por mix, preço ou rateio.',
           'Quando existir meta direta de produto, registrá-la explicitamente por período e autor, mantendo histórico.',
           'Comparar planejamento e forecast sem forçar soma dos SKUs a atingir uma meta financeira por redistribuição automática.'],
    accept='Meta financeira não fabrica demanda ou produção por SKU; estudo, metas explícitas e realizado têm identidades e versões distintas.')
edit('productionprojection','§4.5',action='Calcular reposição/produção líquida por SKU e horizonte considerando o estoque.',
    steps=['Usar demanda prevista, estoque-alvo necessário, estoque utilizável e produção válida ainda não incorporada ao saldo; aplicar normalização técnica registrada em BK-009.',
           'Respeitar datas de disponibilidade e unidades. Produção concluída já incluída no estoque não é descontada novamente.',
           'Executar cálculo cronológico por período usando posição inicial/saldo anterior e entradas válidas; BK-093 apresenta a trajetória desse mesmo plano, sem dependência circular.',
           'Necessidade não positiva não cria produção negativa; excesso permanece na projeção de estoque. Dado obrigatório ausente não assume zero.'],
    accept='Estoque utilizável reduz a necessidade conforme a regra vigente, substituindo o contrato anterior. Resultado é planejamento, sem OP real, e não desconta duas vezes estoque/produção.',
    deps=['targetscenario','forecastpolicy','opsingest','inventoryrules'])
edit('inventoryprojection','§4.5',steps=['Usar posição inicial comprovada e o mesmo plano cronológico da necessidade de produção.',
    'Evoluir saldo por período com entradas válidas de produção e demanda, garantindo que cada evento conte uma vez.',
    'Manter saldo do período anterior como abertura seguinte, com unidade coerente e ausência explicitada; não alimentar BK-092 com seu próprio resultado futuro.'],
    accept='Trajetória reconcilia saldo anterior + entradas de produção únicas − demanda; produção já incorporada não reaparece como entrada futura.')
extra('inputprojection','§3.1,4.5','Entregar demanda bruta de cada insumo por data a Compras, usando necessidade de produção e BOM comprovada; informar quantidades, unidades e dependência do produto que motivou a demanda.')
edit('projectionfinancial','§4.6',steps=['Calcular faturamento previsto somente com quantidades e preços de referência comprováveis, separados da meta financeira comercial.',
    'Exibir diferença entre forecast e meta sem rateio automático para forçar igualdade.',
    'Identificar custos/parcelas disponíveis e ausentes; cenário não modifica competência, resultado realizado, caixa ou obrigação no Bling.'],
    accept='Resumo reconcilia com os SKUs que realmente pertencem ao forecast; meta financeira permanece planejamento específico, sem criar fato realizado.')
edit('projectionsscreen','§§4–5',steps=['Preservar layout T12, abas, cenário, filtros, gráficos e controles aprovados, usando cobertura histórica real e métricas calculadas.',
    'Permitir edição de metas explícitas com capacidade e histórico; a meta financeira total não é distribuída automaticamente em unidades.',
    'Mostrar forecast, metas diretas, produção necessária e estoque projetado conforme seus contratos e dados, com ausência explícita.'])
edit('executivescreen','§4.2,5',action='Construir a tela a partir da referência aprovada localizada, reutilizando métricas existentes.',
    steps=['Preservar blocos, menu e indicadores da última versão aprovada/congelada.',
           'Produzido Geral e detalhamento T10 compartilham função, período, snapshot e universo; quantidades com unidades diferentes não formam soma sem significado.',
           'Reutilizar resultados de caixa, DRE Bling, estoque e planejamento; origem atrasada, parcial ou indisponível permanece identificada.'])
edit('extrascreens','§5',action='Localizar contratos aprovados e decompor a implementação dos itens confirmados.',
    steps=['Vincular comportamento de Produtos, Simuladores e Relatórios às últimas referências aprovadas; localizar artefato ausente sem redesenhar.',
           'Decompor fluxos de implementação e critérios de aceite conforme os contratos existentes.',
           'Preservar navegação, permissões e fontes; não inventar rota ou comportamento por texto ilustrativo.'])
edit('approvedexports','§5.7',steps=['Implementar os formatos/colunas definidos nas referências aprovadas para T02–T07 e T13–T15, incluindo Excel quando especificado.',
    'Consumir mesmos dados, filtros, snapshot, unidade e período da tela, com capacidade de exportação.',
    'Conferir totais, tipos, tratamento seguro de conteúdo e falhas de download; ausência de dado não é zero.'],
    accept='Botões de exportação aprovados permanecem funcionais e geram conteúdo coerente com a tela; não exigir nova decisão sobre uma função já congelada.')
extra('tracecoverage','integral','Vincular F-01 a F-20 ao código futuro e critérios de aceite; regras antigas superadas permanecem apenas em histórico, sem bloquear novamente o negócio.')
extra('domainverification','§§2–4,6,8','Cobrir D-1/exceção corrente, DRE/balancete Bling, compra parcial/tardia, produção líquida, plano temporal ausente, meta financeira sem rateio e histórico fechado preservado.')
by['domainverification']['deps'].append('losses')
extra('integrationverification','§§7–8','Validar serviços e retomada após reinício sem sessão humana logada; repetir eventos/arquivos sem duplicar recebimento, perda, frete, lançamento ou KPI.')
edit('uiverification','§5',steps=['Comparar cada tela com o mockup aprovado/congelado, preservando blocos, posições, cores, tipografia e navegação.',
    'Conferir card/tabela/gráfico/detalhe/exportação com a mesma métrica e conjunto de dados; números do mockup são ilustrativos.',
    'Verificar teclado, foco, mensagens e estados de ausência sem redesenho; corrigir implementação divergente.'])
edit('e2everification','§§1–8',steps=['Validar fluxo Bling + agenda futura → conciliação → caixa projetado/disponibilidade → calculadora, com D-1 e exceção explícita.',
    'Validar em caminho independente DRE/balancete extraídos do Bling e reconciliados na competência/posição, sem reconstrução a partir do caixa.',
    'Validar forecast por SKU → estoque/produção necessária → BOM/insumos → compras parciais e temporais → exportação.',
    'Validar permissões, metas, vigência, período fechado, reprocessamento auditado e recuperação sem alterar arquivos de origem.'])
extra('e2everification','§8.14','Validar perda operacional → vínculo ao fato Bling ou pendência → correção/reprocessamento idempotente; manter a trilha externa sem recalcular uma DRE paralela.')
edit('localdesign','§7',steps=['Projetar aplicação, banco e jobs como serviços independentes de login de funcionário no Windows local.',
    'Documentar host, sistema, serviços, inicialização, portas, dependências, dados persistentes e atualização.',
    'Usar autenticação e credenciais somente no backend; acesso externo, se habilitado expressamente, usa canal seguro sem exposição direta de banco, segredos ou interfaces internas.',
    'Documentar saída para Bling/Drive, armazenamento interno, ausência de internet e recuperação.'],
    accept='Topologia opera sem sessão humana e tem inicialização controlada; este backlog não habilita acesso externo nem instala serviço.')
extra('backupdesign','§7.5','Incluir banco, configurações, parâmetros, metas, usuários/capacidades, registros de processamento, períodos fechados, logs necessários e dados locais não reconstruíveis pela origem.')
extra('restoreexercise','§7.6','Validar integridade de banco, configuração, permissões, fechamentos e histórico após restauração; documentar como subir os serviços sem sessão humana.')
edit('productionmonitoring','§7.7–7.8',steps=['Monitorar serviço, banco, disco, processamento e backup; registrar fonte, execução, horário, sucesso/falha, arquivo/lote e identificadores.',
    'Diagnosticar aplicação parada, API indisponível, token vencido, rejeição, duplicidade e processamento bloqueado sem revelar credenciais.',
    'Documentar procedimentos de resposta e suporte; aplicar limites técnicos justificados e alertas pelo canal habilitado.'])
edit('training','§§7–8',steps=['Documentar uso, capacidades, metas, vigência, período fechado, reprocessamento e auditoria.',
    'Explicar Drive somente de leitura, arquivo corrigido, rejeição e repetição segura; movimentos normais usam D-1 e dia corrente exige solicitação expressa.',
    'Entregar diagnóstico de aplicação, banco, API, token, arquivos, processamento e recuperação dos serviços; realizar exercício com responsáveis na implantação.'])

rules=[
('F-01','Calculadora: contrato fechado','Regra fechada: disponibilidade consolidada alimenta a calculadora. Contrato preserva jan..M−2, principal líquido e fórmulas consolidadas; DRE apresentada no Majucau agora reproduz o Bling. Cartão/MDR/parcelas não pertencem à calculadora.','§1',['investmentpolicy','investmentcalculator','apicontract','criticalweek']),
('F-02','DRE por competência do Bling','Fonte primária: DRE por competência do Bling. Preservar classificação, contas e competência; não criar estrutura financeira/gerencial paralela ou fatos fictícios. Validar contra o próprio Bling.','§2',['accountingpolicy','blingcontract','financialmapping','dreengine','drescreen']),
('F-03','Balancete: reproduzir a origem','O usuário informa base suficiente no Bling. Mapear recursos reais, reproduzir contas/saldos/competência e comparar totais; elemento específico não acessível = Dado indisponível no Bling.','§2',['accountingpolicy','blingcontract','trialbalanceengine','trialbalancescreen']),
('F-04','Compra líquida e cobertura parcial','Necessidade bruta − estoque utilizável − cobertura pendente que chega em tempo. Exemplo: 100−20−30=50 kg. Pedido parcial não exclui insumo; pedido tardio não cobre necessidade anterior à chegada.','§3.1',['opsrules','purchaseeligibility','purchasesizing','inputprojection']),
('F-05','Prioridade de compra','Priorizar risco imediato de ruptura e impacto na produção usando cobertura, demanda, lead time, pedidos e datas reais. Distinguir temporalidade e cobertura; não inventar prioridade aleatória.','§3.2',['opsrules','purchasehistory','purchasesizing','purchasesscreen']),
('F-06','Custos e tarifas separados','Bruto, variável, fixo, critério de rateio e custo apropriado separados. Usar apropriação Bling existente; não multiplicar tarifa integral por produto nem inventar rateio. Não misturar custo com quantidade a comprar.','§3.3',['pricing','purchasesizing']),
('F-07','Cacau independente do motor principal','Inteligência adicional usa localidades reais dos pequenos produtores e sinais comparáveis. Refinar fontes de preço/clima sem impedir compras operacionais; ausência de sinal não cria região ou preço.','§3.4',['cocoapolicy','cocoasignals','purchasesscreen']),
('F-08','Produção e métrica única','Realizado vem do Bling. T01 Produzido Geral e T10 detalhe usam mesma função, fonte, snapshot e universo; mesma métrica em qualquer componente tem mesmo resultado.','§4.1–4.3',['productionrules','productionscreen','executivescreen','dataquality']),
('F-09','Status temporal de produção','Realizado >= meta = atingida. Em andamento/abaixo dependem do plano no ponto temporal; sem granularidade, não inventar atraso, tolerância ou andamento linear.','§4.4',['opsrules','productionrules','productionscreen']),
('F-10','Produção considera estoque','Forecast por SKU → demanda → estoque/reposição → produção → BOM/insumos/compras. Considerar estoque utilizável e produção válida sem dupla contagem; regra antiga sem desconto de estoque foi superada.','§4.5',['forecastpolicy','productionprojection','inventoryprojection','inputprojection']),
('F-11','Meta financeira não fabrica SKUs','Meta de faturamento é referência comercial/cenário. Forecast se fundamenta em histórico; meta direta por produto só é planejamento quando existir explicitamente. Não ratear total financeiro automaticamente em unidades.','§4.6',['forecastpolicy','targetscenario','productiontargets','projectionfinancial','projectionsscreen']),
('F-12','Cobertura histórica real','Usar histórico confiável de 2025 em diante e medir cobertura real por SKU; não completar meses ou assumir venda zero. Qualidade requer backtest real.','§4.7',['forecasthistory','forecastmodel','forecastvalidation']),
('F-13','Precedência e design congelado','Visual/layout: mockup aprovado. Fonte/cálculo/comportamento: documentação consolidada. Usar última versão aprovada/congelada, sem redesenho. Artefato ausente precisa ser localizado, sem reabrir a aprovação.','§5',['missing','designsystem','uxstates','uiverification','baseline']),
('F-14','Exportação aprovada preservada','Manter exportações aprovadas, inclusive Excel onde especificado. Arquivo usa os mesmos dados, filtros, período, snapshot e métricas da tela.','§5.7',['actionpolicy','approvedexports','projectionexport','exportverification']),
('F-15','Capacidades e menor privilégio','Usuário + perfil + capacidade. Separar leitura/exportação/metas/parâmetros/pendências/usuários/auditoria; não hardcodar nomes nem permitir autoelevação.','§6.1–6.2',['accesspolicy','identity','authorization','usersscreen']),
('F-16','Vigência, fechamento e auditoria','Metas são planejamento. Alterações registram usuário, antes/depois, data, período e vigência. Período fechado não muda silenciosamente; reprocessamento é explícito, autorizado, auditado e preserva versão anterior.','§6.3–6.8',['parameterversions','provenance','parametersscreen','productiontargets','accessaudit','pendingengine']),
('F-17','Operação é engenharia','Windows local por serviços independentes de login humano, inicialização controlada, autenticação, segredos no backend, backup automático, restauração testada, logs e suporte documentados. Acesso externo só quando habilitado.','§7',['localrequirements','localdesign','localhostprep','backupdesign','restoreexercise','productionmonitoring','training']),
('F-18','Bling e exceções específicas','Bling é principal para fatos operacionais/financeiros. Drive read-only: perdas operacionais, recebíveis futuros Nuvem e dados específicos Nuvem Envio. Exceções não substituem fatos Bling fora de seu escopo.','§8.1–8.5,8.14',['filecontracts','receivableingest','cashengine','losses','logisticsingest','integrationsscreen']),
('F-19','Conciliação normal em D-1','Análise/conciliação normal de movimentos vai até o dia anterior. Dia corrente somente mediante solicitação expressa. Não generalizar a exceção de 04/09/2026 em regra fixa de sexta-feira/feriado. D-1 não corta a agenda futura nem define frequência da coleta.','§8.11',['filecontracts','jobengine','freshness','matchengine','manualmatch','cashengine','receivables','training']),
('F-20','Conciliação, ausência e idempotência','ID Bling prioritário; identificação/nome+valor no fallback; ambiguidade não concilia automaticamente. Ausência != zero; PENDING_RECONCILIATION fora de indicadores que exigem reconciliação. Repetição não duplica fatos/KPIs; origem é rastreável.','§8.6–8.13',['matchengine','manualmatch','dataquality','receivables','idempotency','provenance','integrationverification']),
]

# Conversão do registro herdado: nenhuma pergunta de negócio superada continua ativa.
answers={
'D-05':'Precedência atual: mockup aprovado define visual; documentação consolidada define cálculo, fonte e comportamento; última versão aprovada/congelada prevalece. A orientação anterior de decidir cada conflito caso a caso foi substituída nos pontos fechados.',
'D-08':'Somente atualização do backlog nesta rodada, confirmado expressamente pelo usuário após o envio do fechamento. Implementação e validação real são tarefas futuras.',
'C-03':'Origem atual do resultado é a DRE por competência do Bling, apresentada no Majucau. A interpretação anterior de uma DRE calculada/classificada independentemente foi superada.',
'L-04':'Perdas operacionais mantêm fonte própria; efeito financeiro e competência seguem os fatos registrados/classificados no Bling. Não criar ou deslocar linhas da DRE por inferência.',
'L-05':'Juros, taxas e rendimentos respeitam conta/classificação e competência existentes no Bling. Mapear tecnicamente, sem nova classificação de negócio.',
'L-06':'Apropriação Bling tem preferência; custos bruto/variável/fixo/critério/final separados. Não duplicar tarifa nem fabricar rateio sem dados.',
'L-07':'A calculadora está fechada. Localizar tecnicamente o contrato/campo Caixa Semana Crítica na referência vigente; não escolher início/fim/mínimo sem evidência. Não retornar como pergunta financeira.',
'L-08':'Reproduzir balancete existente no Bling com campos reais, competência e classificação. Dado específico não acessível pela API permanece indisponível.',
'L-09':'DRE usa origem Bling; sem nova estrutura ou fonte contábil paralela. Exceções externas têm função própria e não completam linhas ausentes com hipóteses.',
'L-11':'Ler a agenda futura na estrutura definida; mapear arquivos/chaves tecnicamente. Conciliação normal de movimentos usa D-1, sem cortar vencimentos futuros da agenda.',
'L-13':'Frete cobrado/custo/subsídio separados nos dados específicos; efeito na DRE respeita classificação/competência Bling, sem nova reclassificação.',
'L-14':'Recuperação é evento rastreável; caixa depende de fato financeiro registrado e DRE respeita a origem. Aprovação de ressarcimento não cria dinheiro.',
'L-15':'Custo de carregamento pertence à análise econômica quando houver dado/regra comprovável. Sua ausência não bloqueia quantidade/prioridade do motor de compras; não inventar taxa.',
'L-17':'Necessidade líquida usa demanda bruta, estoque utilizável e cobertura em tempo; pedido parcial reduz somente parte efetivamente coberta.',
'L-18':'Prioridade operacional usa risco, cobertura, lead time, pedidos e datas reais, com ordenação explicável. Estatística/janela é especificação técnica rastreável.',
'L-19':'Refinamento técnico de fontes comparáveis e localidades reais; sinais não bloqueiam motor principal de compras.',
'L-20':'Avaliação temporal e critérios de qualidade são execução técnica com cobertura histórica real, sem valores do mockup como prova.',
'L-21':'Não distribuir meta financeira automaticamente por SKU. Forecast vem do histórico; metas diretas por produto só quando explicitamente cadastradas.',
'L-22':'Exportações previstas/aprovadas permanecem; implementar contrato técnico com mesmos dados/filtros e Excel quando especificado, sem perguntar novamente se a função entra.',
'L-23':'Ciclo de pendência rastreável, com ator/data/alteração/origem e justificativa técnica; resolução trata a causa dentro da capacidade correspondente.',
'L-24':'Frequência/timeout/validade por fonte são configurações técnicas; análise/conciliação de movimentos usa D-1. Não manter D+1 como regra corrente.',
'L-25':'Parâmetros editáveis por capacidade, com antes/depois, usuário, data e vigência. Período fechado não se altera silenciosamente; reprocessamento explícito auditado.',
'L-26':'Arquitetura de identidade e matriz de capacidades são responsabilidade técnica; autorização não depende de nomes de diretores e segue menor privilégio.',
'Q-01':'Reproduzir a última referência visual aprovada/congelada; localizar o arquivo oficial se necessário. Não redesenhar menu/logo ou reabrir tela congelada.',
'Q-02':'Localizar os artefatos aprovados dos módulos. Ausência física no pacote é rastreio documental, não nova aprovação de tela nem licença para inventar conteúdo.',
'Q-03':'Inventariar host Windows e recursos por inspeção técnica quando disponível; não inventar CPU/RAM/disco nem devolver dimensionamento como pergunta financeira.',
'Q-04':'Acesso autenticado. Acesso externo é condicional a habilitação explícita e usa canal seguro; esta rodada não publica/expoe serviços.',
'Q-05':'Propor serviços sem sessão humana, inicialização controlada e comportamento documentado com falhas de energia/internet; validar funcionamento desejado na implantação.',
'Q-07':'Pedido parcial não elimina compra: descontar só quantidade pendente em tempo. Depósito vem inequivocamente do Bling; ausência não é hipótese.',
'Q-08':'Status depende de realizado, meta e posição temporal do plano. Sem granularidade suficiente, não classificar atraso arbitrário ou assumir progresso linear.',
'Q-09':'DRE por competência e balancete do projeto utilizam Bling e suas classificações. Validar totais tecnicamente; não solicitar estrutura contábil paralela.',
'Q-10':'Calculadora fechada. Consolidar fórmulas M-030/M-034 preservadas pelo fechamento, usando fonte DRE Bling e aportes líquidos de resgates; limite executável Itaú.',
'Q-11':'Produção necessária considera demanda, estoque-alvo e estoque/produção válidos no horizonte; contrato anterior sem desconto de estoque foi superado. Documentar normalização técnica da fórmula.',
'Q-12':'T01/T10 compartilham a mesma informação de produção. Metas são planejamento versionado, distinto de realizado e forecast T12; não criar metas independentes para o mesmo indicador nem sincronizar conceitos diferentes automaticamente.',
'Q-13':'Botões e navegação seguem referência aprovada; contrato de execução e exportação é detalhamento técnico, sem redesenhar ou retirar função existente.',
'Q-14':'Usuário, perfil e capacidade com menor privilégio; atribuição concreta é configuração administrativa explícita na implantação, sem hardcode por pessoa.',
'Q-15':'Desenvolvimento propõe objetivos de recuperação, backup e procedimento; medir restauração e validar funcionamento desejado. Não registrar prazo/perda tolerável como já aprovados sem evidência.',
'Q-16':'Projetar backup automático de dados locais, configuração e histórico; destino seguro fora do disco único e das três estruturas read-only; documentação de operador/suporte.',
'Q-17':'Organizar capacidade, estimativas e responsáveis por papel tecnicamente. Prazo real depende de alocação e evidências, sem prometê-lo ou converter isso em regra financeira.',
'Q-18':'Preservar visual congelado; acessibilidade, foco e estados técnicos dentro do padrão existente. Não usar responsividade como autorização de redesenho.',
'Q-19':'ID Bling prioritário; identificação/nome + valor no fallback; ambiguidade fica para conferência, sem tolerância de valor inventada.',
'Q-20':'Agenda futura é previsão; Bling fornece fato financeiro. Aplicar D-1 ao realizado normal; pendente de conciliação fica fora de indicador que exige valor reconciliado.',
'Q-21':'Calculadora fechada; calendário/corte e contrato de caixa são especificação técnica. Corte normal D-1; não escolher posição crítica por conveniência nem criar regra fixa de sexta-feira.',
'Q-22':'Preservar gráfico aprovado e usar métrica/base de cálculo da documentação funcional; valores demonstrativos não impõem resultado. Não somar subtotais sobrepostos como população única.',
'Q-23':'Usar contas, natureza e saldos do Bling, preservando competência; ausência de saldo não equivale a zero. Conferir totais na mesma posição.',
'Q-24':'Homologar campos reais de BOM, unidades e fatores industriais existentes; perdas operacionais não são yield/scrap. Não pedir nova regra genérica já fechada.',
'Q-25':'Implementar aba prevista na referência aprovada; cenário é planejamento, sem forçar meta de faturamento a coincidir com SKUs por rateio fictício.',
'Q-26':'Comportamento/prioridade vêm da documentação consolidada e visual da última referência aprovada. Modelar monitoramento técnico e canais habilitados sem notificações externas presumidas.',
'Q-27':'Especificar tecnicamente fórmulas, unidades, períodos e fontes dos indicadores já aprovados; reproduzir componentes comprováveis e indisponibilidade, sem inventar números do mockup.',
'R-10':'Fonte Nuvem permanece em 02_nuvem_pago/recebimentos_futuros. A rotina anterior D+1 foi explicitamente substituída por D-1 para análise/conciliação de movimentos; agenda futura continua projetada.',
'R-14':'Cacau é inteligência adicional com fontes técnicas a refinar; não bloqueia o motor principal de compras e não cria novas estruturas Drive.',
}
technical_ids={'L-02','L-03','Q-03','Q-15','Q-16','Q-17','Q-27'}
reference_ids={'L-10','L-16','Q-01','Q-02','Q-25'}
for d in decisions:
    d['previous_status']=d['status']; d['previous_question']=d['question']; d['previous_answer']=d['answer']
    d['source']+='; '+LABEL
    d['question']='Definição vigente e trabalho técnico vinculado.'
    d['status']='Definida'; d['nature']='Regra/atribuição definida'
    if d['id'] in answers: d['answer']=answers[d['id']]
    if d['id'] in reference_ids:
        d['status']='Referência a localizar';d['nature']='Rastreabilidade documental'
        d['answer']=answers.get(d['id'],'Localizar última versão aprovada/congelada do bloco/tela; não reabrir UX nem criar rota por inferência.')
    elif d['id'] in technical_ids:
        d['status']='Execução técnica pendente';d['nature']='Execução técnica'
        d['answer']=answers.get(d['id'],'Verificar tecnicamente a origem e evidências do item; não é nova pergunta de negócio.')
    elif d['id']=='L-01': d['answer']='Usuário confirmou somente atualização do backlog nesta rodada; o fechamento funcional não representa implementação executada.'
    if not d['answer']: d['answer']='Aplicar o fechamento vigente e executar as tarefas técnicas vinculadas; sem nova pergunta de negócio.'
    d['impact']='Aplicar a definição vigente; execução, evidência e aceite são registrados nas tarefas relacionadas.'

ht=[
('Resgates de principal','Mapear categoria, sinal, conta e aplicação, incluindo resgate parcial/total no Bling. Conceito aportes menos resgates está definido.'),
('Posição comprovada de aplicações','Obter posição, data, instituição e aplicação quando acessíveis; exibir componentes conhecidos sem presumir reinvestimento.'),
('Campos reais de estoque/depósito','Homologar identificação operacional inequívoca, saldos obrigatórios e estados opcionais. Não hardcodar depósito.'),
('Extração e conferência do balancete Bling','Mapear recursos existentes e reproduzir contas/saldos/competência; comparar totais contra Bling, sem nova estrutura contábil.'),
('Custo histórico para análise por SKU','Demonstrar custo histórico versus atual na origem, sem reconstruir a DRE por método paralelo.'),
('Extração e conferência da DRE Bling','Mapear recursos/campos e correspondência técnica de linhas da DRE por competência; preservar classificação e validar totais contra Bling.'),
('BOM e demanda temporal de insumos','Homologar componentes, quantidades e unidades; converter produção necessária em consumo por data sem fatores de perdas operacionais.'),
('Chaves reais e corte D-1','Implementar ID Bling/fallback nome+valor sem ambiguidade, D-1 normal e exceção corrente expressa; preservar agenda futura e impedir dupla contagem.'),
('Validação temporal do forecast','Executar backtest com histórico confiável de 2025 em diante, cobertura real, métricas e limitações comprovadas.'),
('Sinais adicionais do cacau','Refinar fontes, localidades reais e comparações, sem bloquear recomendação operacional. Verificar somente fontes habilitadas.'),
]
for n,(theme,requirement) in enumerate(ht,1):
    d=db[f'HT-{n:02}'];d.update(theme=theme,question=requirement,status='Execução técnica pendente',nature='Homologação técnica',answer='Regra funcional definida. Mapeamento, implementação e validação ainda não executados nesta atualização do backlog.',impact='Evidência técnica deve mostrar origem, resultado, limitações e casos de aceite; ausência não pode virar zero.')
for i,theme,answer,section,keys in rules:
    decisions.append(dict(id=i,theme=theme,question='Definição recebida no fechamento vigente.',impact='Aplicar nas tarefas vinculadas e verificar com evidência na implementação futura.',source=LABEL+'; '+section,status='Definida',answer=answer,nature='Regra/atribuição definida',tasks=''))
    for k in keys:link(k,i)

for t in tasks:
    t['dependencies']=[by[k]['id'] for k in t['deps']]
    # Não alterar crédito de conclusão apenas porque o requisito foi fechado.
    assert t['progress']==next(o['progress'] for o in original if o['id']==t['id'])
for d in decisions:
    d['tasks']=', '.join(t['id'] for t in tasks if d['id'] in re.split(r'[,\s]+',t['questions']))

changes=[]
for old,t in zip(original,tasks):
    changed=[k for k in ['title','goal','action','steps','accept','deps','status','owner','condition'] if old.get(k)!=t.get(k)]
    if changed: changes.append(dict(id=t['id'],title=t['title'],fields=changed))
scope=json.loads((W/'scope.json').read_text(encoding='utf-8'))
for s in scope:
    if s['code']=='T01' or s['module']=='Aplicações':s['coverage']='Referência aprovada a localizar no pacote'
    elif s['code']=='T11':s['coverage']='Contrato recebido; visual aprovado a localizar'
    elif s['module'] in ['Pricing / preço e margem','Logística','Produtos, Simuladores e Relatórios']:s['coverage']='Localizar referência funcional/visual aprovada'
    s['source']+='; fechamento vigente (função/fonte)'

(W/'backlog_v2.json').write_text(json.dumps(tasks,ensure_ascii=False,indent=2),encoding='utf-8')
(W/'decisions_v2.json').write_text(json.dumps(decisions,ensure_ascii=False,indent=2),encoding='utf-8')
(W/'scope_v2.json').write_text(json.dumps(scope,ensure_ascii=False,indent=2),encoding='utf-8')
meta=dict(source=str(SRC),sha256=hashlib.sha256(SRC.read_bytes()).hexdigest(),changes=changes,scope='Somente atualização do backlog, confirmado pelo usuário nesta rodada.',tasks=len(tasks),completed_planning=3,development_progress=0,technical_fronts=10,business_questions_reopened=0)
(W/'revision_v2.json').write_text(json.dumps(meta,ensure_ascii=False,indent=2),encoding='utf-8')

doc=['# Majucau — backlog de ponta a ponta, versão 2','',
     'Revisão de 13/09/2026: fechamento funcional enviado pelo usuário incorporado. Escopo desta rodada confirmado: somente atualizar o backlog. Nenhum código do produto, integração, serviço ou teste real foi executado.','',
     f'Foram revisadas {len(changes)} tarefas mantendo os IDs BK-000 a BK-126. As versões anteriores estão preservadas como histórico. Suas regras superadas não devem orientar a implementação.','',
     '## Base vigente','',
     'Somente Majucau, três diretores, hospedagem local em Windows, backend Go, frontend React JavaScript, monólito DDD hexagonal. Balanço Patrimonial permanece excluído. Mockup aprovado define visual; documentação consolidada define fonte, cálculo e comportamento; prevalece a última versão aprovada/congelada.','',
     'Bling é a fonte principal. DRE por competência e balancete reproduzem a informação existente na origem, sem classificação paralela. Drive somente de leitura conserva 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Não existe pasta adicional de investimentos.','',
     'Análise/conciliação normal usa movimentos até D-1; dia corrente requer solicitação expressa. A agenda de recebimentos futuros continua alimentando o caixa projetado. D-1 é corte da análise, não frequência de sincronização.','',
     'As regras funcionais foram fechadas. A engenharia ainda executará especificação técnica, mapeamento das fontes, implementação e validação. Artefato aprovado ausente no pacote será localizado, sem reabrir sua aprovação ou inventar seu conteúdo.','',
     '## Evolução','',
     '127 tarefas, 3 levantamentos concluídos com evidência, 2,4% por quantidade. Desenvolvimento do produto: **0%**. Fechamento de regra não equivale a código entregue ou homologação executada.','',
     'Cada tarefa recebe 100% quando está Concluída e possui evidência de aceite. Peso inicial 1 mede quantidade. Evolução geral = soma(peso × conclusão) / soma(pesos). O denominador inclui tarefas condicionais ainda previstas; quando o escopo for efetivamente alterado, registrar a versão e revisar pesos/denominador. Não existe atualização automática pelo GitHub.','',
     '## Dependências e ordem prática','',
     'IDs e grupos temáticos foram preservados; sua ordem numérica não é cronograma. As dependências de cada tarefa governam a execução. Em particular, a projeção de insumos BK-094 precede a recomendação BK-082/BK-084.','',
     '1. Transcrever contratos fechados e localizar referências; desenhar domínios, acesso, persistência e operação local.','2. Preparar fundações e adaptadores; comprovar campos reais do Bling/arquivos.','3. Reproduzir DRE/balancete do Bling em caminho próprio; construir conciliação, caixa e disponibilidade que alimentam a calculadora.','4. Construir estoque/produção realizados e forecast; calcular produção líquida, trajetória de estoque e insumos por data.','5. Implementar compras líquidas e prioridade com cobertura parcial/temporal; sinais adicionais de cacau são independentes do motor principal.','6. Integrar telas fiéis aos mockups, controles, exportações, capacidades, vigência e histórico.','7. Executar verificações, validar com a Majucau, ensaiar instalação/backup/restauração, entrar em produção e acompanhar a operação.','',
     'Tarefas condicionais de fontes externas do cacau permanecem previstas, sem bloquear o funcionamento do motor principal. Nenhuma ausência de dados permite valor fictício: componentes obrigatórios ausentes produzem indisponibilidade, e componentes independentes comprovados permanecem utilizáveis.','',
     '## Critério comum de conclusão','',
     'Além do aceite específico: regras vigentes e capacidades aplicadas no backend, visual congelado preservado, métricas compartilhadas, ausência distinta de zero, origem e competência rastreáveis, períodos fechados preservados, verificações proporcionais executadas e evidência registrada. Os 167 cenários da planilha original são previstos e precisam ser atualizados nas regras superadas; não foram executados nesta etapa.','']
last=''
for t in tasks:
    if t['phase']!=last:doc+=['## '+t['phase'],''];last=t['phase']
    doc += [f'### {t["id"]} — {t["title"]}','',f'**Situação:** {t["status"]}. **Evolução:** {t["progress"]:.0%}. **Tipo:** {t["kind"]}. **Responsável sugerido:** {t["owner"]}.','',
            '**Objetivo:** '+t['goal'],'','**O que fazer:** '+t['action'],'','**Como fazer:**','']
    doc += [f'{i+1}. {step}' for i,step in enumerate(t['steps'])]
    doc += ['','**Critério de aceite:** '+t['accept'],'','**Impacto:** '+t['impact'],'','**Dependências:** '+(', '.join(t['dependencies']) or 'Sem tarefa prévia; observar contrato e evidências aplicáveis.'),'','**Fonte:** '+(t['refs'] or 'Requisito técnico para entregar e operar o produto.')]
    if t['questions']:doc+=['','**Regras e verificações vinculadas:** '+t['questions']]
    if t.get('condition'):doc+=['','**Condição:** '+t['condition']]
    if t['evidence']:doc+=['','**Evidência:** '+t['evidence']]
    doc+=['']
(O/'MAJUCAU_BACKLOG_DETALHADO_v2.md').write_text('\n'.join(doc),encoding='utf-8')

report=['# Majucau — fechamento vigente e rastreabilidade, versão 2','',
    'Esta revisão corrige a interpretação que tratava regras já definidas e responsabilidades técnicas como perguntas de negócio. O fechamento foi lido integralmente e aplicado ao backlog. O usuário confirmou expressamente: somente atualizar o backlog nesta rodada.','',
    f'{len(changes)} tarefas revisadas; IDs BK-000 a BK-126 preservados. Três levantamentos continuam concluídos, 2,4% por quantidade e 0% de desenvolvimento. Não houve acesso à conta real Bling/Drive nem implementação ou teste do produto.','',
    '## Substituições expressas','',
    '| Interpretação anterior | Regra vigente | Referência |','|---|---|---|',
    '| DRE calculada/classificada independentemente no Majucau | Reproduzir DRE Bling por competência, preservando contas e classificações. | §2 |',
    '| Rotina Nuvem D+1 | Análise/conciliação normal dos movimentos até D-1; dia corrente só com solicitação expressa. Agenda futura não é eliminada. | §8.11 |',
    '| Existência de pedido elimina recomendação | Só a quantidade pendente que chega em tempo reduz a necessidade; parcial/tardio não mascara falta. | §3.1 |',
    '| Produção projetada não desconta estoque pronto | Considerar demanda, estoque-alvo, estoque utilizável e produção válida, sem dupla contagem. | §4.5 |',
    '| Ratear automaticamente meta financeira por mix/preço | Meta financeira é cenário comercial; forecast por SKU depende de dados reais. Meta direta só quando cadastrada. | §4.6 |',
    '| Reabrir diferenças de números e layout como novas decisões | Números são ilustrativos; visual vem do mockup congelado e função do contrato vigente. | §5 |',
    '| Infraestrutura/permissões como questionário financeiro | Engenharia propõe capacidades, serviço local, segurança, backup, recuperação e suporte; valida funcionamento desejado. | §§6–7 |','',
    '## Regras do fechamento','']
for d in decisions:
    if d['id'].startswith('F-'):report += [f'- **{d["id"]} — {d["theme"]}:** {d["answer"]}']
report += ['','## Execução técnica ainda não realizada','',
    'As dez frentes HT abaixo são trabalho futuro, não novas perguntas de negócio. Os registros L/Q legados foram reclassificados para preservar rastreabilidade. Uma regra definida não comprova endpoint, campo, amostra, instalação ou resultado de teste.','',
    '| ID | Trabalho técnico | Tarefas |','|---|---|---|']
for d in decisions:
    if d['id'].startswith('HT-'):report += [f'| {d["id"]} | {d["question"]} | {d["tasks"]} |']
report += ['','## Cuidados incorporados à especificação','',
    '- DRE/balancete têm cadeia própria de obtenção no Bling. Arquivos Nuvem, perdas e fretes não reconstroem uma segunda demonstração.','- Agenda futura → caixa projetado → disponibilidade → calculadora. Corte de movimentos realizados, data da coleta e competência permanecem conceitos separados.','- Entrada ausente ou pendente de conciliação não vira zero. Pendente fica fora apenas dos indicadores que exigem reconciliação, com sua condição visível.','- O campo/contrato Caixa Semana Crítica será localizado tecnicamente na referência consolidada. O fechamento da regra não autoriza escolher um saldo inicial/final/mínimo por conveniência.','- A fórmula de §4.5 veio com um marcador de lista antes de estoque-alvo. Foi registrada a normalização técnica aditiva: demanda + estoque-alvo − estoque utilizável − produção válida ainda não incorporada ao saldo. Não é multiplicação literal. Entradas obrigatórias e vigências devem ser comprovadas; necessidade não positiva não gera produção negativa.','- T01/T10 compartilham a mesma informação de produção. Isso não iguala meta operacional, forecast e meta financeira da T12.','- T09/T11 só conciliam contagens quando a métrica e universo forem equivalentes. Necessidade futura pode existir acima do mínimo atual.','- O cálculo cronológico usa posição inicial e saldo anterior; BK-092 não depende do resultado futuro de BK-093. Produção já no estoque não é descontada novamente.','- Pedido pendente é descontado uma única vez, limitado à cobertura em tempo; pedido tardio continua visível.','- Fontes de cacau podem ser refinadas sem bloquear compras principais. Os testes de sinais habilitados continuam necessários.','- Localizar arquivo aprovado ausente não é reabrir uma tela. Não há autorização para redesenhar, reorganizar menu ou inventar conteúdo.','- Períodos fechados preservam snapshots e versões; parâmetro atual não retroage. Reprocessamento exige capacidade, solicitação explícita e auditoria.','- Host local funciona por serviços sem sessão humana. Acesso externo é condicional, sem exposição ativada nesta rodada.','',
    '## Tarefas revisadas','', '| ID | Tarefa | Campos alterados |','|---|---|---|']
for c in changes:report += [f'| {c["id"]} | {c["title"]} | {", ".join(c["fields"])} |']
report += ['','## Registro vigente de regras e verificações','',
    'Este registro não é um questionário. Definida indica regra recebida; Execução técnica pendente indica trabalho por executar; Referência a localizar indica artefato aprovado não identificado no pacote disponível. Atribuição efetiva de acesso/host e validação de funcionamento pertencem à implantação.','']
for d in decisions:
    report += [f'### {d["id"]} — {d["theme"]}','',f'**Situação:** {d["status"]}. **Natureza:** {d["nature"]}.','',
               '**Regra / trabalho vigente:** '+(d['question'] if d['id'].startswith('HT-') else d['answer']),'','**Fonte:** '+d['source']]
    if d['tasks']:report += ['','**Tarefas:** '+d['tasks']]
    report += ['']
report += ['## Fontes e preservação','',
    f'- [Fechamento enviado pelo usuário](<../FONTES.md#fechamento-das-questoes>).',
    '- Confirmação posterior do usuário: somente atualizar o backlog.',
    f'- [Backlog vigente](<MAJUCAU_BACKLOG_DETALHADO_v2.md>).',
    f'- [Análise original dos mockups e planilha](<MAJUCAU_ANALISE_E_DECISOES_v0.md>), preservada como histórico de leitura; regras superadas não orientam a implementação.',
    f'- [Consolidação anterior](<MAJUCAU_ANALISE_E_DECISOES_v1.md>), preservada como histórico.','',
    'SHA-256 da nova fonte: `'+meta['sha256']+'`.','',
    'As verificações desta entrega cobrem documentos, IDs, dependências, rastreabilidade, fontes preservadas e fórmulas de evolução. Nenhuma delas representa homologação dos dados Bling, execução de backtest ou teste do produto.']
(O/'MAJUCAU_ANALISE_E_DECISOES_v2.md').write_text('\n'.join(report),encoding='utf-8')
print(json.dumps({k:v for k,v in meta.items() if k not in ['changes','source','sha256']},ensure_ascii=False))
print('Tarefas revisadas:',len(changes))
