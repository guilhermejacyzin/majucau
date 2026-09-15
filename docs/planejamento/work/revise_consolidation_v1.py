import copy, hashlib, json, re
from pathlib import Path

ROOT = Path.cwd()
WORK = ROOT / 'work'
OUT = ROOT / 'outputs'
SOURCE = Path(r'sources/private/consolidacao-respostas.txt')
SOURCE_LABEL = 'Consolidação das respostas do desenvolvimento — enviada pelo usuário em 13/09/2026'
tasks = json.loads((WORK / 'backlog.json').read_text(encoding='utf-8'))
decisions = json.loads((WORK / 'decisions.json').read_text(encoding='utf-8'))
before = copy.deepcopy(tasks)
by = {t['key']: t for t in tasks}
db = {d['id']: d for d in decisions}

def edit(key, sections, **changes):
    t = by[key]
    t.update(changes)
    t['refs'] += '; Consolidação 13/09/2026, ' + sections

def add_step(key, sections, step, acceptance=''):
    t = by[key]
    edit(key, sections, steps=t['steps'] + [step], accept=t['accept'] + (' ' + acceptance if acceptance else ''))

def relate(key, *ids):
    t = by[key]
    values = [x for x in t['questions'].split(',') if x]
    t['questions'] = ','.join(dict.fromkeys(values + list(ids)))

add_step('baseline', 'consolidação integral', 'Incorporar as definições expressamente respondidas na consolidação de 13/09/2026 e manter pendências anteriores não tratadas. Não confundir definição recebida com integração homologada.')
add_step('inventory', 'consolidação integral', 'Catalogar a consolidação enviada em 13/09/2026, com hash, mudanças e vínculo às tarefas, preservando os arquivos e a análise anteriores.')
edit('investmentpolicy', '§§6–9',
     action='Registrar os componentes já decididos e fechar apenas os critérios financeiros restantes.',
     steps=[
         'Preservar semana com maior total de despesas, lucro da DRE Majucau e Já Investido = aportes de principal − resgates de principal; rendimentos ficam separados.',
         'Usar Valor Executável Hoje = MIN(Máximo Aplicável Financeiramente; Saldo Transferível do Itaú). Mercado Pago não aumenta automaticamente o executável.',
         'Fechar a posição de caixa dentro da semana crítica, corte temporal do lucro, reserva e calendário; esses detalhes não foram respondidos nesta consolidação.',
         'Definir exemplos financeiros esperados para resgates, aplicação e saldo transferível. A comprovação de campos reais pertence a BK-040/HT-01 e não é pré-requisito de conclusão desta definição; ausência de resgate identificado não equivale a zero.'],
     accept='Decisões financeiras e exemplos esperados aprovados cobrem limites, lucro negativo, período e dados incompletos. BK-006 não exige um adaptador pronto: a prova de obtenção dos componentes fica em BK-040 e condiciona BK-071/BK-072. A falta apenas da posição total não bloqueia componente independente comprovável.')
add_step('accountingpolicy', 'pendências 4–6', 'Usar o mapa origem → conta Bling → classificação gerencial → linha da DRE. Separar homologação técnica de saldos/custos da aprovação financeira de classificação e competência.', 'Nenhuma contrapartida, saldo inicial ou custo histórico é presumido a partir de movimentos insuficientes.')
edit('opsrules', '§3', steps=[
    'Consumir a identificação inequívoca do depósito pelo cadastro operacional retornado na API, conforme HT-03. Se não existir evidência suficiente, apresentar Dado indisponível no Bling.',
    'Fechar qual pedido aberto é aplicável, inclusive cobertura insuficiente ou recebimento parcial; a consolidação não responde essa regra.',
    'Decidir Em andamento e Abaixo da meta na T10 e a fórmula de quantidade, prazo histórico e prioridades da T11.'])
add_step('forecastpolicy', 'pendência 9', 'Usar histórico confiável de 2025 em diante e propor qualidade/limites com evidência de backtest; 24 meses permanece referência documental, nunca uma cobertura artificial.', 'Histórico efetivamente disponível, corte e limitações são declarados; meses ausentes não são inventados.')
edit('filecontracts', '§§1–2,10–12',
     title='Homologar contratos do Bling e das três estruturas do Google Drive',
     action='Mapear Bling/API e arquivos de perdas, Nuvem Pago e Nuvem Envio na pasta compartilhada Majucau.',
     steps=[
         'Preservar 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio; acesso somente de leitura, sem pasta de investimentos, Itaú, Mercado Pago ou novos ajustes.',
         'Obter amostras reais e mapear colunas, IDs, sinais, datas, períodos e versões; a rota da fonte está decidida, mas seu conteúdo ainda requer homologação.',
         'Especificar identificação interna por arquivo + hash + conteúdo/período + lote, sem mover ou sobrescrever o original; definir armazenamento interno, retenção e tratamento de correção rastreável.',
         'Preservar D+1 para Nuvem/Nuvem Pago; documentar calendário, horário e tolerância operacional antes de implementar o agendamento, sem estender D+1 às demais fontes.'],
     accept='Contratos identificam somente as três estruturas externas aprovadas, seus arquivos reais e o Bling. Amostras e acesso ainda precisam ser obtidos; não há nova pergunta sobre a escolha da pasta ou permissão de escrita.',
     owner='Engenharia de Integrações + responsável pelos arquivos')
add_step('cocoapolicy', 'pendência 10', 'Tratar fontes e regras do cacau como refinamento posterior; não ativar fonte nem criar pasta adicional por inferência. Registrar dependência antes da implementação.', 'Posterior não significa exclusão automática da primeira produção; alteração de escopo depende de decisão explícita.')
add_step('domainmap', '§§6,10–12', 'Separar eventos de perdas operacionais de rendimento industrial; separar venda, direito a receber, liquidação, aporte, resgate, rendimento realizado e posição comprovada.')
add_step('threatmodel', '§2', 'Projetar identidade técnica do Drive com permissão somente de leitura, restrição às estruturas autorizadas e credenciais apenas no servidor; escrita interna não concede escrita na origem.', 'O desenho não usa a pasta de fontes como destino de exportação, logs ou backup.')
add_step('parametersscreen', '§§1,4–5', 'Atualizar a consulta de fontes oficiais: investimento vem do Bling/Conta Caixa e não de extrato na pasta externa. Exibir somente as três estruturas do Drive homologadas.', 'Textos antigos do cartão Extratos — Pasta externa não permanecem como regra ativa.')
add_step('provenance', '§2', 'Guardar internamente nome do arquivo, identificador disponível, origem, hash, lote, detecção, processamento, período, status, total/aceitos/rejeitados, motivos e vínculo entre reprocessamentos.', 'Correções mantêm versão anterior e explicam seu efeito sobre os fatos; metadados não são gravados no Drive.')
edit('blingcontract', '§§3–9; pendências 1–8',
     steps=[
         'Homologar autenticação, permissões e cobertura em respostas reais, registrando endpoint, campo, exemplo protegido, data, significado, limite e ausência.',
         'Verificar produtos, vendas, títulos, contas Itaú/Mercado Pago, Conta Caixa, depósitos, estoque, OPs, compras, notas, custo histórico, BOM e elementos do balancete.',
         'Para investimentos, provar aporte, resgate parcial/total e posição separadamente. caixa.php é referência funcional; usar somente API oficial ou exportação estruturada oficial, sem scraping.',
         'Entregar evidência para HT-01 a HT-08; disponibilidade técnica não aprova por si só classificação financeira. Quando insuficiente, registrar a limitação e os módulos afetados.'],
     accept='Matriz requisito → campo real, com amostras e limitações verificadas. Nenhum campo ou endpoint é tratado como existente apenas porque foi citado na consolidação; verificações ainda não executadas neste planejamento.')
add_step('blingadapter', '§§4,8–9', 'Manter Conta Caixa, Itaú e Mercado Pago dentro da leitura oficial do Bling. Se a homologação apontar necessidade de exportação estruturada oficial de investimentos, implementar recepção local interna desse arquivo com o parser de BK-043; não usar o Drive.', 'Não contém scraping nem integração direta presumida com banco ou Mercado Pago. A alternativa oficial necessária tem recepção, autenticação e proveniência implementadas; repetição e versão corrigida são tratadas por BK-044.')
edit('fileingestion', '§§1–2',
     title='Implementar leitura das três estruturas autorizadas do Google Drive',
     goal='Importar arquivos preservando integralmente a pasta compartilhada.',
     action='Construir adaptador de leitura do Drive e área de processamento dentro do Majucau.',
     steps=[
         'Configurar a identidade e os identificadores reais da pasta/estruturas; limitar leitura a 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio.',
         'Detectar arquivos, obter conteúdo consistente e calcular hash local; se houver alteração durante a leitura, não publicar lote parcial.',
         'Guardar controle e cópia de evidência necessária no armazenamento interno, conforme retenção homologada; nunca criar, mover, renomear, excluir, sobrescrever ou editar arquivos na origem.',
         'Tratar permissão revogada, indisponibilidade de rede e versão corrigida com diagnóstico e retomada interna.'],
     accept='Leitura funciona com permissão somente de leitura; processar e reprocessar não altera nenhum arquivo do Drive. Arquivo parcial ou fora do contrato não produz fatos publicados.',
     impact='Protege os originais e mantém a trilha de importação dentro do Majucau.')
add_step('idempotency', '§2', 'Distinguir identidade do arquivo, hash do conteúdo, período e lote; detectar duplicatas de eventos mesmo em arquivos diferentes. Correção cria nova versão interna com diferença rastreável, sem substituição silenciosa.', 'O efeito da correção sobre relatórios é demonstrável e nenhuma versão externa é alterada.')
add_step('parsers', '§§1–2,4', 'Vincular parsers do Drive exclusivamente às três estruturas aprovadas. Se necessária após BK-040, implementar também interpretação da exportação estruturada oficial de investimentos do Bling recebida internamente por BK-041, fora do Drive; não habilitar fontes ilustrativas antigas.', 'A alternativa oficial, quando necessária, tem parser e recepção atribuídos, com proveniência e deduplicação por BK-044; não cria uma quarta estrutura no Drive.')
edit('receivableingest', '§10', action='Ler a agenda Nuvem/Nuvem Pago na estrutura 02_nuvem_pago/recebimentos_futuros e relacioná-la aos fatos do Bling.',
     steps=by['receivableingest']['steps'] + ['Preservar rotina D+1 e separar venda, direito futuro e liquidação comprovada; presença no arquivo de recebimentos futuros não efetua baixa.'],
     accept='Direitos e eventos reconciliam com sua origem; venda, recebível e dinheiro recebido não triplicam caixa/receita. Liquidação requer comprovação financeira e chave homologada em HT-08.')
edit('payableingest', '§§8–10', action='Ler títulos e movimentos financeiros disponíveis no Bling, incluindo contas Itaú e Mercado Pago.',
     steps=by['payableingest']['steps'] + ['Mapear identidade real de cada conta, sinal, transferência interna e evidência de liquidação; não prever arquivo bancário separado na pasta compartilhada.'])
add_step('opsingest', '§3; pendência 7', 'Homologar no cadastro operacional ID/descrição do depósito, saldo físico e, se existirem, reservado/disponível/outros estados. Ingerir também BOM por produto, componente, quantidade e unidade.', 'Sem identificação inequívoca, mostrar Dado indisponível no Bling; não hardcodar nome/ID nem escolher automaticamente Bloqueado.')
add_step('logisticsingest', '§12', 'Usar 03_nuvem_envio e preservar separadamente custo efetivo, frete cobrado do cliente, diferença/subsídio, pedido, transportadora, etiqueta e data quando disponíveis.', 'Frete cobrado do cliente não é compensado silenciosamente com custo; ausência de coluna fica explícita.')
edit('supplementaryingest', '§11',
     title='Ingerir eventos de perdas operacionais',
     goal='Receber eventos de perda com origem e vínculo financeiro verificáveis.',
     action='Interpretar arquivos da estrutura 01_perdas_operacionais.',
     steps=[
         'Mapear evento, identificador, datas, valor, item e vínculo com Bling conforme amostra homologada.',
         'Preservar ocorrência, confirmação, baixa e recuperação como fatos distintos; impedir duplicidade entre arquivo e Bling.',
         'Manter a fonte no domínio Perdas Operacionais; não extrair percentual de scrap, rendimento de receita, yield ou fator automático de aumento de matéria-prima.'],
     accept='Cada evento tem proveniência e efeito financeiro conforme regra aprovada. Arquivo não cria venda, dinheiro ou movimento de estoque; não existe ingestão de extrato externo de investimento ou planilha genérica de ajustes nesta tarefa.',
     impact='Evita misturar perdas financeiras com parâmetros industriais.')
add_step('freshness', '§10', 'Aplicar D+1 exclusivamente à rotina Nuvem/Nuvem Pago e publicar data de referência separada da data de processamento. Completar horários/calendário e limites das outras fontes no contrato.', 'Dados antigos não ficam atuais apenas porque a consulta ao Drive ou ao Bling foi bem-sucedida.')
add_step('integrationsscreen', '§§1–2', 'Substituir fontes ilustrativas antigas pelas reais: Bling e as três estruturas aprovadas; exibir metadados, rejeições e versões internas sem controles de escrita no Drive.', 'Não exibir extrato de investimentos, arquivo Itaú/Mercado Pago ou Planilha Ajustes/Cacau como fonte ativa. Quantidade de fontes declara o critério e não é número fixo do mockup.')
add_step('matchengine', '§10; pendência 8', 'Homologar chaves reais de pedido, transação, parcela e movimento nos arquivos e Bling antes de confirmar vínculos. Separar confirmação do direito e comprovação da liquidação.', 'Recebimento futuro não é evidência suficiente de dinheiro recebido; ambiguidade mantém análise manual.')
add_step('receivables', '§10', 'Distinguir vínculo com venda, direito a receber e liquidação bancária comprovada; a agenda futura em D+1 não constitui dinheiro recebido.')
edit('cashengine', '§§8–10', deps=[x for x in by['cashengine']['deps'] if x != 'supplementaryingest'] + ['payableingest'],
     steps=by['cashengine']['steps'] + ['Distinguir caixa consolidado das contas comprovadas e saldo transferível do Itaú. Mercado Pago pode compor o consolidado; recebível futuro não vira saldo bancário realizado.'],
     accept=by['cashengine']['accept'] + ' Valores de contas são obtidos do Bling; planilha de perdas não é fonte direta de saldo bancário. Transferências entre contas não criam entrada líquida consolidada.')
add_step('historicalcost', 'pendência 5', 'Comparar respostas reais de venda antiga e custo atual para provar semântica temporal do campo; guardar endpoint, campo e evidência do teste em HT-05.', 'Se só houver custo atual, histórico permanece indisponível, sem retroagir valores novos.')
add_step('cashscreen', '§§8–10', 'Identificar a origem e data do caixa consolidado e distinguir o valor transferível do Itaú para aplicação; recebíveis futuros permanecem previstos.')
add_step('financialmapping', 'pendência 6; §12', 'Implementar mapa origem → conta Bling → classificação gerencial → linha da DRE, com versão, competência e aprovação financeira.', 'Frete cobrado, custo logístico, taxas e rendimentos são identificados separadamente antes da composição aprovada.')
add_step('losses', '§11', 'Impedir dependência da planilha de perdas como fonte de scrap, yield ou ajuste automático de insumos na projeção.', 'Reconhecimento financeiro e transformação industrial continuam em regras distintas.')
edit('trialbalanceengine', 'pendência 4', action='Produzir partidas e saldos somente se a fonte homologada fornecer elementos suficientes.',
     steps=[
         'Comprovar no Bling a disponibilidade e significado dos saldos iniciais, débitos, créditos e contrapartidas; validar com o financeiro em HT-04.',
         'Vincular partidas reais ao documento e competência, preservando sua origem e natureza de saldo.',
         'Se a base for insuficiente, mostrar explicitamente a limitação; não fabricar partidas ou apresentar fluxo de caixa como balancete.',
         'Registrar diferenças sem criar contrapartida ou ajuste automático para fechar totais.'],
     accept='Balancete completo exige evidência dos elementos contábeis; ausência de abertura não equivale a zero. Base insuficiente gera estado indisponível/incompleto, sem remover o módulo do escopo automaticamente.')
edit('investmentledger', '§§4–7',
     title='Implementar movimentos de investimentos e posição comprovada no Bling',
     goal='Separar principal, rendimentos e posição com evidência de origem.',
     action='Usar Conta Caixa via API oficial ou exportação estruturada oficial homologada.',
     steps=[
         'Identificar Transferência destinada à aplicação como aporte de principal e Rendimentos como rendimento realizado; não classificar toda transferência como investimento.',
         'Homologar categoria, sinal, origem/destino, aplicação e resgates parciais/totais. Já Investido = aportes de principal − resgates de principal; período/cobertura precisam ser comprovados.',
         'Obter posição, data e instituição somente se a fonte demonstrar esses campos; principal líquido não é automaticamente posição atual.',
         'Exibir componentes independentes comprováveis e declarar os ausentes. Não presumir reinvestimento de rendimento nem resgate zero por falha de identificação.'],
     accept='Aportes e resgates reconciliam com fatos reais. Posição ausente não é inventada. Indicador dependente de resgate desconhecido permanece incompleto; indicador independente comprovado pode ser exibido.',
     deps=['blingadapter','idempotency','investmentpolicy','financialmapping','payableingest'])
add_step('investmentcalculator', '§§6–9', 'Aplicar Valor Executável Hoje = MIN(Máximo Aplicável Financeiramente; Saldo Transferível do Itaú), separando conta operacional de caixa consolidado. Usar aportes líquidos, sem somar rendimentos ao principal.', 'Itaú indisponível ou saldo transferível não comprovado impede confirmar o executável; saldo de Mercado Pago não supre essa ausência. A falta apenas da posição total não bloqueia automaticamente todos os cálculos.')
add_step('calculatorscreen', '§§6–9', 'Identificar Itaú na memória do executável, apresentar aportes/resgates/rendimentos separadamente e explicar exatamente qual entrada falta.', 'Não copiar posição ou capacidade fictícia do mockup; componentes verificados conservam sua visibilidade.')
edit('applicationsscreen', '§§4–7', steps=[
    'Homologar visual, campos e filtros usando os dados que Conta Caixa realmente disponibilizar, sem exigir extrato externo no Drive.',
    'Mostrar aportes, resgates, rendimentos realizados e suas datas; mostrar posição atual somente quando comprovada, com instituição/aplicação.',
    'Reconciliar a consulta com Bling e calculadora, mantendo explicação de ausência por componente.'],
    accept='Layout validado e totais reconciliados com a fonte oficial homologada; nenhum rendimento é presumido reinvestido e não existe scraping ou pasta de investimento.')
add_step('inventoryrules', '§3', 'Consumir somente depósito operacional inequivocamente identificado por HT-03; preservar diferença entre saldo físico, reservado e disponível retornados.', 'Identificação ambígua ou dado obrigatório ausente produz Dado indisponível no Bling e impede apenas resultados dependentes. A ausência de reservado/disponível opcionais não invalida saldo físico comprovado se a regra homologada usar esse saldo; estados opcionais não são inferidos.')
add_step('logisticsanalysis', '§12', 'Exibir custo efetivo do frete, valor cobrado do cliente e diferença/subsídio em componentes separados, com rastreio para pedido/etiqueta quando disponíveis.')
add_step('cocoasignals', 'pendência 10', 'Implementar somente depois do refinamento de fontes e comparação; não criar Planilha Cacau nem ampliar as três estruturas do Drive.', 'Fontes não homologadas não aparecem como ativas ou confiáveis.')
edit('forecasthistory', 'pendência 9', action='Preparar histórico confiável de 2025 em diante, declarando meses disponíveis por SKU.',
     steps=by['forecasthistory']['steps'] + ['Não buscar dados anteriores a 2025 para completar artificialmente a referência de 24 meses; registrar insuficiência e corte temporal antes do backtest.'],
     accept='Cobertura real começa em 2025 ou depois conforme a série; mês ausente não é venda zero. Treino e avaliação declaram quantidade de observações, limitações e origem.')
add_step('forecastvalidation', 'pendência 9', 'Executar backtest apenas com histórico confiável de 2025 em diante, registrando janelas de treino/teste e restrições de comparação sazonal.', 'Nenhum backtest foi executado na elaboração deste backlog; a tarefa requer os dados reais.')
add_step('inputprojection', '§11; pendência 7', 'Usar BOM/ficha técnica homologada do Bling, com componente, quantidade e unidade. Yield ou scrap exigem dado e regra industrial próprios; jamais derivar da planilha de perdas operacionais.', 'BOM incompleta impede resultado confirmado para o SKU afetado; zero ou fator presumido não preenche a lacuna.')
add_step('integrationverification', '§§1–2,6–10', 'Verificar em ambiente controlado leitura com permissões restritas, versão corrigida, mesmo conteúdo com outro nome, períodos sobrepostos, reprocessamento e interrupção; conferir fonte intocada antes/depois.', 'Cobrir também resgate parcial/total, posição ausente e venda/recebível/liquidação sem duplicidade; registrar evidências na implementação futura.')
add_step('pendingengine', '§§2–7', 'Distinguir arquivo rejeitado, origem atrasada, campo indisponível no Bling, resgate não identificado e posição não comprovada; registrar os indicadores afetados sem transformar ausência em zero.')
add_step('tracecoverage', 'consolidação integral', 'Vincular R-01 a R-14 e HT-01 a HT-10 aos contratos, tarefas e cenários novos; identificar cenários da versão anterior superados pela mudança de fontes.')
add_step('domainverification', '§§6–12', 'Cobrir principal aportado menos resgates, rendimento separado, limite executável pelo Itaú, frete sem compensação silenciosa e perdas operacionais sem efeito automático na BOM.', 'Tratar ausência, ambiguidade e cobertura temporal insuficiente com resultados explícitos.')
add_step('securityverification', '§2', 'Verificar em ambiente de teste que a identidade e o adaptador do Drive não conseguem mutar a origem; inspecionar escopos, caminhos e operações, sem tentar escrita destrutiva na pasta real.', 'Nem usuário administrador do Majucau ganha escrita no Drive por reprocessar um lote.')
add_step('localdesign', '§§1–2', 'Prever acesso de saída do servidor Windows ao Bling e ao Google Drive, armazenamento interno de evidências e comportamento com internet indisponível.', 'Aplicação hospedada localmente não implica independência de internet para atualizar as fontes.')
add_step('backupdesign', '§2', 'Copiar banco, metadados e evidências internas necessárias para destino de backup próprio; as três estruturas do Drive são somente leitura e não recebem cópias de segurança do sistema.', 'Restauração do Majucau não precisa escrever, mover ou renomear arquivos na origem.')
add_step('restoreexercise', '§2', 'Restaurar histórico de hashes, lotes e versões antes da retomada; verificar que reler o Drive não duplica fatos nem altera arquivos externos.')
add_step('initialload', '§§1–2; pendência 9', 'Carregar somente fontes aprovadas; projeções usam histórico confiável de 2025 em diante. Registrar ausência de posição, custos históricos, BOM ou aberturas sem preenchê-los por hipótese.')
add_step('training', '§§1–2,10', 'Explicar que Majucau consulta o Drive e mantém controles internos; correção externa é feita pelo responsável na origem. Demonstrar D+1, nova versão, rejeição e reprocessamento sem alterar arquivo original.')

def update_decision(i, status, answer, question=None, nature=None):
    d = db[i]
    d['status'], d['answer'] = status, answer
    if question: d['question'] = question
    if nature: d['nature'] = nature
    d['source'] += '; ' + SOURCE_LABEL

update_decision('C-03', 'Respondida', 'Lucro da DRE Majucau. Já Investido foi definido como aportes de principal menos resgates de principal; fonte dos resgates ainda requer homologação. Corte temporal do lucro continua sem nova resposta.')
update_decision('L-03', 'A verificar tecnicamente', 'Homologar custo histórico versus atual no retorno real do Bling. Ver HT-05; o texto não comprova disponibilidade.', nature='Verificação técnica')
update_decision('L-08', 'A verificar tecnicamente', 'Verificar no Bling abertura, débitos, créditos e contrapartidas. Sem base suficiente, não produzir balancete fictício. Ver HT-04.', nature='Verificação técnica')
update_decision('L-09', 'Parcialmente esclarecida', 'Fontes ativas são Bling e três estruturas específicas do Drive. Finalizar cobertura de cada linha e mapa da DRE em HT-06; não adicionar planilha complementar por inferência.')
update_decision('L-11', 'A verificar tecnicamente', 'Fonte Nuvem definida em 02_nuvem_pago/recebimentos_futuros, somente leitura, rotina D+1. Faltam amostras, chaves e contrato real. Ver HT-08.', nature='Verificação técnica')
update_decision('L-12', 'Respondida', 'Origem somente de leitura. Proibido criar, mover, renomear, excluir, sobrescrever ou alterar arquivos. Controle, hash, lotes, rejeições e versões ficam internamente; arquivo corrigido não sobrescreve fatos silenciosamente.')
update_decision('L-13', 'Parcialmente esclarecida', 'Nuvem Envio usa 03_nuvem_envio. Custo efetivo, frete cobrado e subsídio ficam separados. Classificação e competência da DRE ainda exigem HT-06.')
update_decision('L-19', 'Em aberto', 'Construir fontes e comparações posteriormente, após refinamento. Não criar pasta adicional ou ativar fonte presumida. Não houve decisão explícita para retirar Cacau do escopo de produção.')
update_decision('L-20', 'Parcialmente esclarecida', 'Backtest deve usar histórico confiável de 2025 em diante. Métricas, limites e suficiência dependem de execução e validação; não há resultado de backtest nesta etapa.')
update_decision('L-24', 'Parcialmente esclarecida', 'D+1 é rotina da Nuvem/Nuvem Pago. Falta documentar calendário, horário e política das demais fontes; não estender esse prazo ao conjunto todo.')
update_decision('Q-06', 'Respondida', 'Bling/API concentra operacional/financeiro, Itaú principal, Mercado Pago e Conta Caixa/investimentos; Drive contém apenas perdas, Nuvem Pago e Nuvem Envio. Cacau continua refinamento posterior. Identificadores e capacidade real do Bling são verificações técnicas.')
update_decision('Q-07', 'Parcialmente esclarecida', 'Depósito deve vir inequivocamente do cadastro operacional da API, sem nome/ID presumido; ausência = Dado indisponível no Bling. Pedido insuficiente/parcial continua decisão operacional pendente.', 'Qual deve ser o tratamento de pedido aberto insuficiente ou parcialmente recebido para recomendar nova compra? Identificação técnica do depósito está em HT-03.')
update_decision('Q-10', 'Parcialmente esclarecida', 'Já Investido = aportes de principal − resgates de principal, sem rendimentos. Itaú limita o executável. Permanecem corte temporal do lucro e reserva; resgates reais e posição são HT-01/02.', 'Confirmar corte temporal do lucro elegível e reserva mínima. A composição financeira de Já Investido já foi definida.')
update_decision('Q-19', 'Parcialmente esclarecida', 'Primeiro homologar chaves dos arquivos Nuvem e Bling em HT-08. A rotina não confunde direito futuro com liquidação. Casos ambíguos e decisões manuais continuam sem nova definição.')
update_decision('Q-24', 'Parcialmente esclarecida', 'BOM/ficha técnica deve ser homologada no Bling: componente, quantidade e unidade. Perdas operacionais não fornecem scrap/yield. Fatores industriais e vigência só podem ser usados se comprovados e aprovados.', 'Homologar a BOM real do Bling e identificar eventuais fatores industriais próprios, sem usar a planilha de perdas operacionais.', 'Verificação técnica e validação operacional')

new = [
('R-01','Fontes e estruturas externas','Bling/API + Google Drive — pasta compartilhada Majucau, preservando 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Sem pasta de Itaú, Mercado Pago, investimentos ou ajustes.','§1',['filecontracts','fileingestion','integrationsscreen']),
('R-02','Permissão e rastreio no Drive','Somente leitura, sem qualquer mutação na origem. Metadados, hashes, lotes, rejeições e vínculos de reprocessamento ficam dentro do Majucau. Correção não sobrescreve fatos silenciosamente.','§2',['filecontracts','fileingestion','provenance','idempotency','threatmodel','backupdesign']),
('R-03','Identificação do depósito','Identificar no cadastro operacional da API, sem hardcode ou escolha automática de Bloqueado. Sem evidência inequívoca: Dado indisponível no Bling.','§3',['blingcontract','opsingest','inventoryrules']),
('R-04','Fonte de investimentos','Conta Caixa no Bling; caixa.php é referência funcional. Usar API oficial ou exportação estruturada oficial. Scraping proibido; não haverá extrato externo na pasta compartilhada.','§4',['blingcontract','blingadapter','investmentledger','applicationsscreen']),
('R-05','Categorias de investimentos','Transferência destinada à aplicação identifica aporte; Rendimentos identifica rendimento realizado. Resgates ainda precisam ser mapeados tecnicamente.','§5',['investmentledger','financialmapping']),
('R-06','Já Investido','Aportes de principal − resgates de principal. Rendimentos ficam separados e não são novos aportes. Falta provar os componentes no Bling.','§6',['investmentpolicy','investmentledger','investmentcalculator']),
('R-07','Posição comprovada','Exibir posição e data somente se disponíveis na fonte; caso contrário mostrar apenas componentes comprováveis. Não presumir rendimento reinvestido.','§7',['investmentledger','calculatorscreen','applicationsscreen']),
('R-08','Itaú e valor executável','Itaú é a conta principal e fornece o saldo transferível, via Bling. Valor Executável Hoje = MIN(Máximo Aplicável Financeiramente; Saldo Transferível do Itaú).','§8',['payableingest','cashengine','investmentpolicy','investmentcalculator']),
('R-09','Mercado Pago','Movimentos via Bling; pode compor caixa consolidado, mas não aumenta automaticamente o executável para aplicações que saem do Itaú.','§9',['payableingest','cashengine','investmentcalculator']),
('R-10','Nuvem Pago e liquidação','Fonte 02_nuvem_pago/recebimentos_futuros, rotina D+1. Separar venda, direito futuro e movimento financeiro que comprova liquidação.','§10',['filecontracts','receivableingest','freshness','matchengine']),
('R-11','Perdas operacionais','Fonte 01_perdas_operacionais representa eventos de perda, nunca percentual industrial, scrap, yield ou fator automático de matéria-prima.','§11',['supplementaryingest','losses','inputprojection']),
('R-12','Nuvem Envio','Fonte 03_nuvem_envio; preservar custo efetivo, frete cobrado, diferença/subsídio, pedido, transportadora, etiqueta e data quando disponíveis, sem compensação silenciosa.','§12',['logisticsingest','financialmapping','logisticsanalysis']),
('R-13','Base do forecast','Executar backtest com histórico confiável de 2025 em diante. Não presumir cobertura de 24 meses nem métricas calculadas.','pendência 9',['forecastpolicy','forecasthistory','forecastvalidation']),
('R-14','Refinamento do cacau','Construir posteriormente fontes externas e regras de comparação aprovadas; isso não amplia as três estruturas do Drive nem altera automaticamente o escopo da primeira produção.','pendência 10',['cocoapolicy','cocoasignals']),
]
for i, theme, answer, section, keys in new:
    decisions.append(dict(id=i,theme=theme,question='Definição recebida nesta consolidação.',impact='Aplicar nas tarefas vinculadas; homologação técnica continua separada.',source=SOURCE_LABEL+'; '+section,status='Respondida',answer=answer,nature='Definição recebida',tasks=''))
    for key in keys: relate(key,i)

technical = [
('HT-01','Resgates de principal','Demonstrar categoria, sinal, origem/destino, aplicação e eventos de resgate parcial e total no Bling/Conta Caixa.','Tabela de campos reais e exemplos conciliados; falta de identificação não equivale a zero.','Verificação técnica',['blingcontract','investmentledger','investmentcalculator']),
('HT-02','Posição atual da aplicação','Demonstrar se o Bling fornece posição real, data, instituição e aplicação ou somente movimentos.','Exibir somente componentes comprováveis; não inventar posição nem reinvestimento.','Verificação técnica',['blingcontract','investmentledger','applicationsscreen']),
('HT-03','Campos do depósito','Homologar identificação operacional, ID, descrição, saldo físico e estados reservado/disponível se existirem.','Identificação inequívoca documentada ou Dado indisponível no Bling; sem hardcode.','Verificação técnica',['blingcontract','opsingest','inventoryrules']),
('HT-04','Suficiência do balancete','Verificar saldos iniciais, débitos, créditos e contrapartidas reais e validar suficiência com o financeiro.','Matriz de cobertura contábil; não apresentar fluxo financeiro como balancete completo.','Verificação técnica e validação financeira',['blingcontract','accountingpolicy','trialbalanceengine']),
('HT-05','Custo histórico','Homologar endpoint/campo e provar diferença entre custo da venda antiga e custo atual.','Exemplos temporais verificáveis; se só existir custo atual, declarar ausência de custo histórico.','Verificação técnica',['blingcontract','historicalcost']),
('HT-06','Mapa da DRE','Finalizar origem → conta Bling → classificação gerencial → linha da DRE, incluindo competência e ausências.','Mapa versionado com evidências técnicas e validação financeira; desconhecido permanece pendente.','Verificação técnica e validação financeira',['blingcontract','accountingpolicy','financialmapping','dreengine']),
('HT-07','BOM/ficha técnica','Homologar componente, quantidade e unidade reais por produto no Bling, além dos fatores industriais que existirem.','Composição demonstrada; arquivo de perdas não fornece scrap/yield.','Verificação técnica e validação operacional',['blingcontract','opsingest','inputprojection']),
('HT-08','Chaves da conciliação','Homologar chaves dos arquivos Nuvem e dos recebimentos do Bling, preservando D+1 e comprovação de liquidação.','Amostras de vínculo e ambiguidade; nenhuma tripla contagem de venda/recebível/dinheiro.','Verificação técnica',['filecontracts','blingcontract','receivableingest','matchengine']),
('HT-09','Backtest de forecast','Executar validação temporal com histórico confiável de 2025 em diante.','Janelas, cobertura, fórmulas, métricas e limitações reais registradas; ainda não executado.','Verificação técnica',['forecastpolicy','forecasthistory','forecastvalidation']),
('HT-10','Fontes e comparação de cacau','Refinar posteriormente as fontes e bases comparáveis antes da construção dos sinais.','Fontes e regras aprovadas; nenhuma pasta ou feed ativo presumido.','Definição de fontes e regras',['cocoapolicy','cocoasignals']),
]
for i,theme,question,accept,nature,keys in technical:
    decisions.append(dict(id=i,theme=theme,question=question,impact=accept,source=SOURCE_LABEL+'; pendência '+str(int(i[-2:])),status='Em aberto' if i=='HT-10' else 'A verificar tecnicamente',answer='Trabalho futuro registrado no backlog. Não houve consulta à conta real, amostras ou execução desta verificação nesta revisão.',nature=nature,tasks=''))
    for key in keys: relate(key,i)

for t in tasks:
    t['dependencies'] = [by[x]['id'] for x in t['deps']]
    if t['evidence']:
        t['evidence'] += ' Evidência original preservada na versão 0; esta revisão não altera a conclusão desse levantamento.'
for d in decisions:
    d.setdefault('nature','Definição recebida' if d['status']=='Respondida' else 'Refinamento de negócio/entrega')
    d['tasks'] = ', '.join(t['id'] for t in tasks if d['id'] in re.split(r'[,\s]+',t['questions']))

changes=[]
for original,t in zip(before,tasks):
    substantive=[field for field in ['title','action','goal','steps','accept','deps','owner'] if original[field]!=t[field]]
    if substantive: changes.append(dict(id=t['id'],title=t['title'],fields=substantive,previous_title=original['title']))

(WORK/'backlog_v1.json').write_text(json.dumps(tasks,ensure_ascii=False,indent=2),encoding='utf-8')
(WORK/'decisions_v1.json').write_text(json.dumps(decisions,ensure_ascii=False,indent=2),encoding='utf-8')
(WORK/'revision_v1.json').write_text(json.dumps(dict(source=str(SOURCE),sha256=hashlib.sha256(SOURCE.read_bytes()).hexdigest(),changes=changes),ensure_ascii=False,indent=2),encoding='utf-8')

done=sum(t['progress']==1 for t in tasks)
pending=[d for d in decisions if d['status'] not in ['Respondida','Planejamento autorizado']]
head=['# Majucau — backlog de ponta a ponta, versão 1','',
    'Atualização de 13/09/2026 com a consolidação enviada pelo usuário. Os IDs BK-000 a BK-126 foram preservados. A versão 0 permanece disponível como histórico. Nenhuma funcionalidade do produto foi desenvolvida nesta etapa.','',
    'Premissas: somente Majucau; três diretores; hospedagem local em Windows; Go + React JavaScript; monólito com DDD e arquitetura hexagonal; todos os módulos confirmados na primeira produção. Balanço Patrimonial fora do escopo.','',
    'Fontes atuais: Bling/API e Google Drive somente de leitura nas estruturas 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Itaú, Mercado Pago e investimentos vêm do fluxo oficial do Bling; não possuem arquivos na pasta compartilhada. Para investimentos, exportação estruturada oficial é alternativa a homologar quando necessária; scraping é proibido.','',
    'Regras respondidas não são integrações executadas. R-01 a R-14 registram definições desta rodada; HT-01 a HT-10 identificam as verificações/refinamentos correspondentes. Os registros L/Q antigos permanecem rastreáveis apenas para assuntos ainda pertinentes.','',
    '## Evolução e conclusão','',
    f'{len(tasks)} tarefas; {done} levantamentos concluídos com evidência; conclusão por quantidade **{done/len(tasks):.1%}**; desenvolvimento do produto **0%**. Esta revisão de regras não gera crédito fictício de desenvolvimento.','',
    'Cada tarefa recebe 100% apenas quando está Concluída e tem evidência de aceite. Peso inicial 1 mede quantidade, sem estimativa de esforço. Evolução geral = soma(peso × conclusão) / soma(pesos). Revisões de escopo/pesos devem ser registradas. A planilha separa planejamento, desenvolvimento, qualidade e produção; não consulta automaticamente o GitHub.','',
    'Para software, a conclusão exige aceite específico, regras/permissões no Go, UI coerente, rastreabilidade, tratamento de dados ausentes e verificações proporcionais executadas. Os 167 cenários da planilha original são previstos, não testes executados.','',
    '## Ordem de execução','',
    'Os IDs são estáveis e não obrigam execução estritamente numérica. Respeitar as dependências. As tarefas iniciais definem regras de negócio, contratos esperados e exemplos; a prova dos campos reais fica em BK-040 e nas homologações vinculadas, antes das funcionalidades dependentes. Não exigir o adaptador pronto para aprovar uma regra financeira. Descoberta e homologação podem exigir mais de uma rodada. Responsáveis são papéis sugeridos; esforço, prazo e nomes continuam a definir.','',
    '1. Consolidar escopo e homologar dados, regras e infraestrutura local.','2. Modelar domínios, contratos, navegação, armazenamento e segurança.','3. Preparar fundações, acesso e integrações.','4. Construir motores e interfaces dos módulos.','5. Verificar jornadas completas e validar com a Majucau.','6. Ensaiar instalação, backup, restauração e atualização no Windows.','7. Aprovar a versão completa, entrar em produção e acompanhar os primeiros ciclos.','']
last=''
for t in tasks:
    if t['phase']!=last: head+=['## '+t['phase'],''];last=t['phase']
    head += [f'### {t["id"]} — {t["title"]}','',f'**Situação:** {t["status"]}. **Evolução:** {t["progress"]:.0%}. **Tipo:** {t["kind"]}. **Responsável sugerido:** {t["owner"]}.','', '**Objetivo:** '+t['goal'],'','**O que fazer:** '+t['action'],'','**Como fazer:**','']
    head += [f'{i+1}. {step}' for i,step in enumerate(t['steps'])]
    head += ['','**Critério de aceite:** '+t['accept'],'','**Impacto:** '+t['impact'],'','**Dependências:** '+(', '.join(t['dependencies']) or 'Sem tarefa prévia; observar decisões e evidências aplicáveis.'),'','**Fonte:** '+(t['refs'] or 'Necessidade técnica de entrega e operação.')]
    if t['questions']: head += ['','**Decisões e verificações vinculadas:** '+t['questions']]
    if t['evidence']: head += ['','**Evidência:** '+t['evidence']]
    head += ['']
(OUT/'MAJUCAU_BACKLOG_DETALHADO_v1.md').write_text('\n'.join(head),encoding='utf-8')

report=['# Majucau — consolidação, análise e decisões, versão 1','',
    'Esta revisão incorpora a consolidação enviada pelo usuário em 13/09/2026 ao planejamento existente. O documento foi lido integralmente. Não houve acesso à pasta compartilhada, consulta à conta real do Bling, importação de dados ou desenvolvimento do produto.','',
    f'Foram ajustadas {len(changes)} tarefas, mantendo os 127 IDs e o progresso: 3 levantamentos concluídos, 2,4% por quantidade e 0% de desenvolvimento. As definições da nova fonte prevalecem nos pontos expressamente respondidos; outros conflitos continuam tratados caso a caso.','',
    '## O que fica decidido nesta rodada','']
for d in decisions:
    if d['id'].startswith('R-'): report += [f'- **{d["id"]} — {d["theme"]}:** {d["answer"]}']
report += ['','## Frentes que ainda exigem evidência ou refinamento','',
    'A lista abaixo registra as dez frentes expressamente mantidas em aberto na consolidação. É trabalho futuro da equipe, com validação da Majucau quando envolver regra financeira ou operacional. Não é um novo questionário para o usuário.','',
    '| ID | Frente | O que precisa ser comprovado | Tarefas |','|---|---|---|---|']
for d in decisions:
    if d['id'].startswith('HT-'): report += [f'| {d["id"]} | {d["theme"]} | {d["question"]} {d["impact"]} | {d["tasks"]} |']
report += ['','## Implicações técnicas para o backlog','',
    '- O adaptador do Drive faz leitura; área de processamento, evidências, logs, lotes, rejeições e versões ficam no Majucau. Backup e exportações precisam de destino próprio, fora das três estruturas de fontes.','- A rota do Google Drive está definida, mas IDs da pasta, autenticação e contratos dos arquivos continuam homologações técnicas. Não é necessário instalar ou conectar um plugin para registrar essas tarefas.','- Investimentos usam Conta Caixa via API oficial ou exportação estruturada oficial a homologar. A alternativa de arquivo oficial não autoriza criar pasta de investimento no Drive.','- Transferência só é aporte quando destinada à aplicação. Rendimentos realizados não são principal. Resgate não identificado não pode ser tratado como zero.','- Posição, principal líquido e rendimento são indicadores distintos. Se faltar apenas a posição total, mostrar componentes independentes comprovados; bloquear somente resultados que dependam de informação ausente.','- Caixa consolidado pode incluir Mercado Pago. O limite operacional do executável vem do Itaú, sem transformar recebíveis futuros em saldo bancário.','- D+1 foi confirmado para Nuvem/Nuvem Pago. Horário e calendário ainda precisam ser detalhados; a periodicidade não se aplica automaticamente ao Bling e demais arquivos.','- Perdas operacionais não alimentam rendimento industrial. BOM e fatores industriais requerem campos e regras próprios.','- A referência visual de 24 meses de histórico não comprova essa disponibilidade. Usar dados confiáveis de 2025 em diante e declarar insuficiência quando necessário.','- Uma insuficiência contábil do Bling não autoriza fabricar abertura/contrapartidas nem substituir balancete por fluxo de caixa. Registrar o impedimento e submetê-lo à validação pertinente.','- A construção posterior das fontes de cacau permanece no planejamento; a frase não foi interpretada como autorização para excluir o módulo da primeira produção.','',
    '## Pontos anteriores que esta rodada não respondeu','',
    'Continuam no registro, entre outros: saldo específico dentro da semana crítica, corte temporal do lucro, reserva, classificação/competência financeira, pedido de compra com cobertura parcial, regras de metas, conteúdo dos módulos sem referência completa, permissões da diretoria, acesso remoto, capacidade/rotina do Windows, recuperação, responsáveis e prazo. Esta consolidação de fontes não fecha automaticamente esses assuntos.','',
    '## Tarefas alteradas','', '| ID | Tarefa atual | Campos revisados |','|---|---|---|']
for c in changes: report += [f'| {c["id"]} | {c["title"]} | {", ".join(c["fields"])} |']
report += ['','## Registro completo de decisões e verificações','',
    f'{len(decisions)} registros, dos quais {len(pending)} permanecem pendentes ou parcialmente esclarecidos. Registros L/Q e HT podem apontar para a mesma frente; a contagem não representa perguntas distintas nem quantidade de falhas do produto. A natureza de cada registro separa definição recebida, decisão de negócio e verificação técnica.','']
for d in decisions:
    report += [f'### {d["id"]} — {d["theme"]}','',f'**Situação:** {d["status"]}. **Natureza:** {d["nature"]}.','',d['question'],'','**Registro atual:** '+(d['answer'] or 'Sem resposta nova nesta rodada; manter o refinamento vinculado.'),'','**Impacto / evidência necessária:** '+d['impact'],'','**Fonte:** '+d['source']]
    if d['tasks']: report += ['','**Tarefas:** '+d['tasks']]
    report += ['']
report += ['## Fontes e histórico preservado','',
    'A análise detalhada dos 16 PNGs (15 imagens únicas), das 20 abas e das inconsistências visuais permanece no documento da versão 0, como registro histórico de leitura. As regras atuais desta versão substituem suas suposições de fontes nos pontos descritos acima. Valores ilustrativos não se tornam dados reais.','',
    f'- [Texto da consolidação recebido](../FONTES.md#consolidacao-das-respostas).',
    f'- [Análise original das imagens e planilha](<MAJUCAU_ANALISE_E_DECISOES_v0.md>).',
    f'- [Backlog atualizado](<MAJUCAU_BACKLOG_DETALHADO_v1.md>).','',
    'SHA-256 da nova fonte: `'+hashlib.sha256(SOURCE.read_bytes()).hexdigest()+'`.','',
    'Verificação desta entrega: consistência de IDs, vínculos e dependências; preservação das versões e originais; fórmulas de evolução; inspeção visual da planilha. Isso verifica os documentos do planejamento, não a integração ou o software Majucau.']
(OUT/'MAJUCAU_ANALISE_E_DECISOES_v1.md').write_text('\n'.join(report),encoding='utf-8')
print(json.dumps(dict(tasks=len(tasks),changed_tasks=len(changes),decisions=len(decisions),pending_records=len(pending),completed_planning=done,development_progress=0),ensure_ascii=False))
