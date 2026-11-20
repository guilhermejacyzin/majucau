# Especificação de implementação da DRE Intelligence Majucau

**Versão 1 para desenvolvimento e aprovação gradual** · competência parametrizada (AAAA-MM) · 15/11/2026

Destinatários: desenvolvimento, financeiro e contabilidade. Escopo: reconstruir as regras observadas no painel de referência dentro do projeto próprio, com origem de cada número, auditoria e visão ajustada explicitamente separada.

## 1. Contrato do trabalho

**Objetivo.** Na tela DRE do Intelligence Majucau, gerar a DRE para qualquer competência mensal selecionada a partir dos relatórios de base do Bling e das premissas auditáveis; reproduzir o resultado do painel estudado como **espelho de referência**; apresentar, em paralelo e com ponte numérica, correções aprovadas pela Gisele na **DRE gerencial ajustada**. Esta especificação registra regras implementadas no código de referência. Ela não atesta que os números daquele painel sejam lucro contábil.

**Período da tela e base de validação.** O seletor de competência da T06 controla todos os cards, linhas, subtotais, gráficos, detalhes, status, conciliações e snapshots apresentados. Ao mudar a competência, todos esses resultados devem ser consultados ou recalculados para a nova seleção; nunca manter setembro fixo, misturar competências ou reaproveitar silenciosamente valores de outro mês. A DRE pronta de **setembro de 2026** é a base de comparação para validar a reprodução das regras; setembro é um caso de validação, não um período fixo/default da tela. Para qualquer outra competência selecionada, a tela deve mostrar apenas o resultado daquele período, com a disponibilidade real das respectivas fontes. Dados históricos podem ser usados como insumo de reconstrução de custo quando a regra exigir, mas o resultado exibido continua atribuído à competência selecionada.

**Fora do escopo imediato:** copiar o aplicativo do outro desenvolvedor, consumir o resultado pronto de DRE do Bling como entrada, escolher hospedagem futura para venda a outras empresas ou alterar já o código do Intelligence sem passar pelas aprovações. A DRE gerada pertence ao projeto e ao banco da Gisele. Os diretores consultam snapshots publicados, sem acesso direto ao PostgreSQL.

**Fonte auditada.** Repositório FinanceiroMajucau/Majucau, pasta `projetos/ativos/dashboard-majucau`, commit `3b67a0418d930a69b8875e58a419ed20398e3a6c`. Arquivos principais: `09_dashboard/metrics/dre.py` (76 especificações), `09_dashboard/pages/financeiro/dre.py`, `09_dashboard/metrics/financeiro.py`, `07_5_cmv_cmp.py`, `07_3_valorizacao_insumos.py`, `08_6_tributario.py`, `data/ancoras/mapa_contabil.json`. Conferir caminhos relativos no checkout do repositório de origem; os identificadores e nomes abaixo são a evidência técnica, não arquivos para executar no Intelligence.

**Projeto de destino.** Repositório local `api-2/majucau`: Wails/React, worker Go por Named Pipe e PostgreSQL local na implementação atual; a DRE da navegação é T06 no código atual. O `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md` define como destino um serviço central autenticado com PostgreSQL privado para consulta por outras pessoas; essa migração ainda não está implementada e seu provedor ainda não foi escolhido. `internal/financial/dre.go` já calcula subtotais após receber valores classificados, mas ainda não importa, classifica nem compõe receita ou CMV. `database/migrations/000001_init.up.sql` já dispõe de `raw_records`, `integration_sync_batches`, `financial_rule_versions`, `calculation_runs`, `statement_snapshots`, `statement_lines` e `statement_line_contributions`. A implementação exige migrações aditivas e adaptador para o cálculo existente.

> **Regra de integridade:** ausência de arquivo, coluna, custo, competência ou cobertura histórica não é R$ 0,00. Registrar ausência, impedir confirmação da linha dependente, exibir motivo e não publicar lucro como fechado. Zero só quando a fonte completa comprovou movimento zero.

### Vocabulário de status

| Status | Condição e ação |
| --- | --- |
| IMPORTADO | Arquivo recebido com hash, período, esquema e totais de controle; ainda não validado. |
| PENDENTE | Fonte, mapeamento, custo ou reconciliação ausente; não preencher com zero. |
| PROVISÓRIO | Cálculo reproduzível, porém há estimativa, ressalva ou conciliação aberta. |
| DIVERGENTE | Espelho e referência ou fontes entre si não fecham; mostrar diferença e lançamentos. |
| CONFIRMADO | Fontes completas, classificação aprovada, conferências aceitas e snapshot publicado. |

## 2. Caminho do dado, do relatório à tela

1. O worker do Intelligence consulta o Bling em modo somente leitura, pelos recursos oficialmente disponíveis e autorizados, e recebe os relatórios/arquivos que não forem fornecidos pela integração. A DRE não usa como entrada o relatório pronto de DRE do Bling. Cada coleta ou arquivo recebe lote, hash SHA-256, período, data de extração, esquema, quantidade, soma de colunas monetárias e responsável. Os conectores que ainda não existem devem ser construídos e validados antes de automatizar a DRE; a importação de arquivos é a alternativa auditável por fonte.
2. Conservar a resposta original da API em `raw_records` e os arquivos recebidos no armazenamento controlado da mesma implantação, com hash e referência no banco; parsear formatos brasileiros, IDs como texto canônico e valores decimais. No protótipo esse banco é local; depois da migração, os dados oficiais ficam na implantação central privada. Reimportação idêntica é idempotente; fonte revisada cria novo lote e nova execução, sem apagar evidência anterior.
3. Normalizar categorias, recebimentos, pagamentos, movimentos de caixa, vendas e itens, NF de saída/entrada, compras, produtos, estruturas, ordens de produção, posições de estoque e complementos tributários. Registrar relação `source_batch_id + source_row_id` em cada linha.
4. Executar as 76 regras de categoria da seção 12 e os dois motores calculados (CMV e custo de envios) no período de competência. Rejeitar IDs sem correspondência e registrar linhas excluídas por motivo.
5. Compor linhas de exibição, subtotais, espelho de referência e, quando aprovadas, as correções ajustadas. Guardar cada parcela com origem e versão da regra.
6. Conciliar receitas e notas, pedido e NF, tributação PGDAS, compras e estoque, OP e fichas, fretes, títulos e caixa. Publicar snapshot do período somente com status explícito.
7. Interface T06 consulta apenas snapshot publicado. Clique numa linha mostra categoria, lançamentos, arquivo, linha de origem, fórmula, exclusões, versão e diferença entre espelho e ajuste. Os indicadores T01, se usados, leem o mesmo snapshot publicado.

### 2.1 Onde fica a base e para onde volta o resultado

| Momento | Local dos dados e do cálculo | Quem acessa |
| --- | --- | --- |
| Protótipo atual | O worker Go e o PostgreSQL do Intelligence rodam no computador da Gisele. O ambiente de desenvolvimento documenta uma instância isolada em `artifacts/retained/postgresql-dev`; isso é caminho de desenvolvimento, não hospedagem definitiva. | Uso local; o projeto atual ainda não é uma implantação de acesso independente para diretores. |
| Implantação central prevista para este projeto | Serviço do Intelligence consulta fontes, importa relatórios, executa regras, guarda RAW, tabelas normalizadas, custos, cálculos e snapshots em PostgreSQL privado. Provedor, região e endereço físico ainda serão definidos. O computador da Gisele é o ponto de administração, não precisa ficar ligado para consultas após a implantação central. | Gisele administra/valida/publica; os diretores autorizados consultam pela aplicação autenticada, sem acesso direto ao banco. Escopo atual: equipe deste projeto, sem desenho para venda a outras empresas. |

**Destino do número:** Bling → leitura de dados de origem → worker/serviço do Intelligence → banco do Intelligence → regras de DRE → snapshot da competência → tela DRE do Intelligence e consultas autorizadas. **Não há etapa de gravar a DRE calculada de volta no Bling.** O Bling continua sendo uma fonte de documentos e transações. A sincronização diária às 09:30 America/Sao_Paulo é uma decisão registrada para o serviço central, ainda sujeita à implementação dos conectores e do processamento da DRE; permitir recálculo controlado e manter o último snapshot publicado se uma fonte falhar.

> **“O meu será a base”:** a Majucau é dona da base oficial e das aprovações. Enquanto o protótipo é local, a base técnica está no computador da Gisele. Para outras pessoas consultarem independentemente, a base técnica passará a residir na implantação central privada do mesmo Intelligence; não se cria uma cópia de DRE em cada computador.

## 3. Contrato de entradas e campos mínimos

A tabela nomeia as entidades que o código da referência consome. São os **campos efetivamente usados no motor**, não uma afirmação de que qualquer CSV já enviado contém todos. O desenvolvedor deverá registrar, no importador, o relatório concreto, cabeçalhos efetivos, versão e colunas ausentes antes de habilitar o processamento da competência escolhida.

| Entidade | Campos exigidos no espelho | Uso e chave |
| --- | --- | --- |
| Dim_Categorias | categoria_id, categoria_pai_id, nome | Mapa de IDs e descendentes; mudanças históricas exigem versão. |
| Dim_Situacoes | situacao_id, modulo | Situações financeiras válidas. Se dimensão faltar, referência aceita registros e conciliação deve denunciar a limitação. |
| Fato_Recebimentos | categoria_id, competencia, valor_original, situacao, origem_tipo, origem_id | Receita e demais créditos; origem liga venda/NF. Guardar título_id, pedido_id/nf_id e condições quando disponíveis. |
| Fato_Pagamentos | categoria_id, competencia, valor_original, situacao, pagamento_id | DAS, GNRE, ICMS-ST e despesas; pagamentos e baixa não substituem competência. |
| Fato_Caixas_Bancos | categoria_id, competencia, valor, tipo D/C, tipo_lancamento | Complemento sem título; tipo_lancamento 2 é baixa e não entra novamente. |
| Fato_Vendas | pedido_id, nf_id, produto_id, quantidade, data_saida, receita_item, pedido_numero_loja, cliente_tipo | Saídas por SKU e curva de pedidos. Preservar operação fiscal, cancelamento, vínculo com pedido e NF para auditoria. |
| Fato_NF_Saidas | nf_id, data_emissao, natureza/CFOP e status quando presentes | Data substituta e verificação de vendas tributadas; NF e pedido não são receitas adicionais por si. |
| Fato_Pedidos_Compra | data_pedido, produto_id, item_quantidade, item_valor_unitario | Entrada de insumos com preço para CMP; chave item e pedido. |
| Fato_NF_Entrada | data_operacao, produto_sku, quantidade, valor_unitario | Fallback de compra sem produto coberto por pedido no espelho; vínculo por SKU. |
| Fato_Producao | data_previsao_final, produto_id, quantidade, ordem_id | Explosão da ficha e entrada de acabado; data prevista é ressalva. |
| Dim_Estruturas | produto_final_id, componente_id, quantidade, sem_estrutura | Ficha técnica e composição da caixa de embarque. |
| Dim_Produtos | produto_id, sku, estoque_atual, preco_custo | Tradução SKU/ID e fallback de estoque/custo. |
| Âncoras | posição de estoque insumos e acabados em 09/01/2026; relatório de visão financeira usado como âncora de custo | Necessárias para repetir reconstrução desde 01/01/2025; sem elas CMV histórico fica pendente. |
| PGDAS e complementos | competência, receita declarada, DAS devido, identificador documental | Conferência externa; não substitui automaticamente a categoria DAS do espelho. |

**Identidade:** IDs do Bling nunca em ponto flutuante. Canonizar como texto sem sufixo “.0”; preservar original, espaço e zeros de SKU. Datas com fuso America/Sao_Paulo; competência de título é o campo `competencia`, não vencimento, pagamento ou importação. Valor de título = `valor_original`; caixa = módulo de `valor` após selecionar direção. Internamente usar decimal de precisão definida e política de arredondamento versionada; exibir centavos, guardar custo unitário com mais casas.

## 4. Receita bruta: regra exata do espelho e controle de operação

**Espelho:** categoria Bling `12805089610` (“Venda de produtos”). Somar `Fato_Recebimentos.valor_original` com `competencia` no período selecionado e `categoria_id` correspondente. Adicionar `abs(Fato_Caixas_Bancos.valor)` da mesma categoria/competência, apenas `tipo=C` e `tipo_lancamento != 2`. Se `Dim_Situacoes` válida estiver presente, filtrar situação do módulo financeiro; o código comenta 1, 2 e 3. Não adicionar total de pedido e total de NF novamente à soma.

**Detalhe obrigatório para conferência:** para cada contribuição mostrar título/caixa, origem_tipo, origem_id, pedido_id, nf_id, cliente, data de competência, situação, valor original, valor incluído e arquivo. Separar no relatório de controle: vendas com NF, vendas sem NF, reenvios, bonificações, remessas, devoluções, canceladas e operações não classificadas. O espelho não contém filtro explícito de natureza fiscal nessa soma: a eventual diferença entre receita financeira categorizada e faturamento efetivo precisa aparecer, não ser varrida para dentro do total.

| Caso | Espelho observado | Conferência e eventual ajuste, somente com aprovação |
| --- | --- | --- |
| Pedido com NF | Entra uma vez pelo título ou crédito direto categorizado; NF valida, não soma de novo. | Vincular pedido, título e NF; impedir dupla contagem da mesma operação. |
| Pedido sem NF | Pode entrar pela mesma categoria se gerou título/crédito. | Exibir separadamente o bruto sem NF; validar operação e competência com financeiro. |
| Reenvio | A referência reconhece categoria de frete de reenvio, mas não prova filtro universal de pedidos reenviados na receita/CMV. | Identificar pedido original, motivo, NF/CFOP e itens; não criar nova venda se for mera reposição. Custo logístico separado quando aplicável. |
| Bonificação/remessa/conserto | Não há exclusão universal demonstrada no motor categorial. | Controlar por operação fiscal e vínculo; retirar do faturamento gerencial ajustado quando não houver venda, sem esconder quantidade/custo de estoque. |
| Devolução/estorno | Categoria própria `12805089622` como dedução. | Conferir pedido/NF original, devolução real e competência; prevenir lançamento também como receita negativa e dedução duplicada. |

> O código da referência reconhece duplicidade residual: a mesma NF pode ter lançamentos de caixa por pedido e nota. É proibido “acertar” receita forçando um total: gerar lista de pares candidatos, evidências e reconciliação manual/automatizada versionada.

**Conciliação da competência selecionada:** total dos títulos incluídos + total dos créditos diretos incluídos = receita espelho; contagens por documento; total com nota + sem nota = vendas classificadas; listar documentos fora da regra, duplicados e diferenças para fiscal/PGDAS. O PGDAS mede declaração tributária, não deve substituir silenciosamente receita operacional.

## 5. CMV/CMP: reconstrução passo a passo

**Origem:** `07_5_cmv_cmp.py`, resultado `df_cmv_cmp_mensal` com `mes` e `cmv_cmp`, injetado em `09_dashboard/metrics/dre.py` como categoria sintética `-1` e exibido “Mercadoria Vendida”. Não vem da DRE pronta do Bling e não é a soma dos pagamentos a fornecedores.

1. **Janela/âncora:** reconstrução inicia em `2025-01-01`. Ler posição de insumos e produtos acabados de `2026-01-09`; reconstruir estoque anterior como `max(0, posição_âncora − entradas + saídas)`. Ler custo unitário da visão financeira usada pelo projeto; se modo de âncora desligado, usar estoque_atual/preco_custo de Dim_Produtos. Guardar qual âncora e hash. No código a constante se chama ANCORA_2024, mas o caminho utilizado é `data/ancoras/relatorio_07_06_2026_vf_atual.csv`; registrar essa divergência de nomenclatura, não presumir estoque 2024.
2. **Ficha técnica:** Dim_Estruturas, produto_final_id → componente_id/quantidade, apenas composição válida e quantidade positiva. Versionar a ficha por data. O código usa dicionário; componente duplicado pode sobrescrever outro, então listar duplicações antes de calcular.
3. **Compras:** para cada item de Fato_Pedidos_Compra desde 2025 com data_pedido, quantidade positiva e preço unitário positivo, criar entrada. Em Fato_NF_Entrada, mapear produto_sku por Dim_Produtos e criar entrada somente se o produto não tem qualquer pedido de compra no conjunto filtrado. Essa deduplicação é por produto, não por documento; preservar no espelho e medir NFs legítimas potencialmente omitidas.
4. **Produção:** Fato_Producao com data_previsao_final, produto_id e quantidade positiva. Para cada OP, gerar saída de cada componente = quantidade produzida × quantidade da ficha e entrada da quantidade de acabado. O motor usa data prevista como data de evento; exibir divergência para data real quando fornecida.
5. **Vendas:** Fato_Vendas por produto_id e quantidade positiva. Data da saída para custo: primeiro competência de recebimento com origem_tipo=notafiscal/nf_id; depois origem_tipo=venda/pedido_id; depois emissão da NF; por último data_saida. A referência faz drop_duplicates por NF/pedido nas datas e pode escolher primeiro arbitrário. Auditar datas concorrentes, produtos sem vínculo e operações sem venda.
6. **Ordenação:** para o mesmo dia, entrada de compra → saída de insumo → entrada do acabado → saída de venda. Empates adicionais devem ser determinísticos por documento/item/ordem; preservar a sequência do espelho para comparação.
7. **CMP de insumo:** em compra, `novo_CMP = (quantidade_anterior × CMP_anterior + quantidade_comprada × preco_unitario)/(quantidade_anterior + quantidade_comprada)`. A saída de produção diminui saldo, mantendo CMP.
8. **Custo do acabado:** para uma unidade de PF, `custo_OP_unitário = soma(quantidade_de_cada_componente_por_unidade × CMP_do_componente_no_evento)`. Nova entrada de PF recalcula seu CMP ponderado com saldo e custo anteriores; sem ficha, mantém CMP anterior.
9. **Venda:** `CMV_do_item = quantidade_vendida × CMP_do_produto_antes_da_saída`; somar pela data de competência atribuída ao evento dentro do período selecionado. Persistir pedido, item, SKU, quantidade, CMP usado, total, ficha, OP/âncora e eventos de custo que contribuíram. O código aplica piso zero ao saldo após saída; isso pode ocultar estoque negativo e não deve ser silencioso na visão ajustada.

**Exceções hard-coded no código original:** SKU `0000000009539` tem divisor 1000; SKU `0000000009287` é ignorado na âncora; overrides de custo para IDs `16259125822 = 264,20` e `16259125821 = 240,00`, além de uma exclusão de item “SC PEROLA METALIZADO”. Reproduzir no perfil espelho somente com log da regra e motivo; não generalizar a outros meses/produtos sem validar a base. Mostrar impacto monetário de cada exceção.

**MOD:** a afirmação de que o custo do produto já traz mão de obra deve ser comprovada por ficha, componente e preço do período. O módulo `07_3_valorizacao_insumos.py` exclui pseudoinsumo “Custos MOD- minutos de trabalho” da projeção de compras, o que por si só não prova sua inclusão no CMV. Se o custo do acabado já inclui MOD, não debitar a mesma MOD novamente como “Mão de Obra Produção” na DRE ajustada; espelho conserva a linha e a ponte evidencia duplicidade. Se ficha não trouxer MOD, informar custo incompleto, sem estimativa disfarçada.

**Gate:** sem âncora, compra, estoque, ficha ou venda auditável, CMV é PENDENTE. Com CMP zero, estoque insuficiente, produção sem ficha ou SKU sem mapeamento, marcar evento, somar impacto mensurável e bloquear CONFIRMADO. Conciliar valor de estoque inicial + compras/produção − baixas/consumo − CMV com saldo final na granularidade de produto e custo, explicando diferenças e arredondamento.

## 6. Demais parcelas e apresentação

### 6.1 Regras de categoria

Para cada uma das 76 categorias na matriz final: selecionar categoria_id e competência no mês; se a origem for recebimentos usar valor_original do título e crédito sem baixa, se pagamentos usar valor_original e débito sem baixa; caixa usa abs(valor). Exceção “Alimentação” é só caixa. As categorias pai INSS, Provisões, Vale alimentação e Vale transporte são subtotais dos descendentes identificados por Dim_Categorias; não adicioná-las junto com as filhas ao agrupar. Categoria nova sem destino no mapa é erro e impede publicação. Situação válida vem de Dim_Situacoes; preservar comportamento de fallback da referência com alerta.

### 6.2 Deduções

DAS 14675501032, GNRE 14712313565, ICMS-ST 14662426036, descontos 12805089623, devoluções 12805089622, outros tributos 14708854913. O realizado mensal da referência usa Fato_Pagamentos por competência, inclusive DAS; `08_6_tributario.py` calcula projeção para visão semanal, não substitui o valor realizado mensal. Comparar o DAS ao PGDAS da mesma competência. **Regra operacional confirmada por Gisele (2026-11-15): imposto pago no mês 11 refere-se ao mês 10.** Portanto, o mês de pagamento não deve ser usado sozinho como competência. Quando a apuração/guia da competência correta ainda não estiver disponível, calcular uma estimativa para essa competência usando como base os impostos já pagos e a regra de cálculo de origem; exibir separadamente como estimativa, com memória de cálculo e vínculo aos pagamentos/documentos usados, sem apresentá-la como realizado confirmado. Não presumir que o valor pago é igual ao imposto devido, nem inventar alíquota/fórmula: reproduzir a regra tributária documentada e reconciliar com PGDAS/guia da mesma competência quando disponível. DIFAL/GNRE requer distinguir imposto de venda e de compra; não inserir tributo de compra sem validação na linha de deduções.

### 6.3 Custo de envios e fretes

O motor gerencial de “Custo de Envios” é categoria sintética `-2`. Caixa de embarque produto `16612242717`: custo unitário da ficha × CMP dos componentes; o código usa R$ 5,6184702866 como comparação de sanidade ±30%, não como custo substituto. Sem ficha, retorna zero no espelho; na publicação da DRE ajustada isso é falta de dado. Pedidos B2C: perfil de ticket dos últimos 12 meses, percentis P30/P90; fallback com tickets 100/200/500 e pesos 50/40/10, cortes 150/375; até faixa média 1 caixa, alta 2. B2B: ABC de clientes do ano anterior, limites acumulados A até 80%, B até 95%, C restante; caixas por pedido A=10, B=5, C=2. O realizado da referência usa recebimentos por competência até o corte e deduplica origem_id. Mostrar pedido, segmentação, quantidade estimada de caixas, custo unitário e total. Classificar separado o frete pago pela Majucau, frete cobrado do cliente, transportadora e reenvio; não somar frete cobrado como despesa na visão ajustada sem verificar sua natureza.

### 6.4 Ordem exata do waterfall do painel

| Bloco | Linhas na ordem | Subtotal exibido |
| --- | --- | --- |
| Receita | Receita Operacional Bruta | — |
| Deduções | DAS; GNRE; ICMS-ST; Descontos Incondicional; Devoluções de vendas; Outros Tributos | Receita Líquida |
| Custos variáveis | Mercadoria Vendida; Custo de Envios | — |
| Despesas variáveis | Frete pago Majucau; Frete pago por cliente; Comissões; Marketing; Taxas de Vendas; Outras Desp. Comerciais | Margem de Contribuição |
| Fixos | Mão de Obra Produção; Ocupação; Mão de Obra Administrativo; Benefícios Gerais; Serviços Profissionais; Consumo e Manutenção; Software; Viagens e Transporte; Promotoria de Vendas; Pró Labore; Retirada de Lucro; Outras Despesas | EBITDA Gerencial |
| Resultado financeiro | Receita Financeira; Despesa Financeira; Outras Receitas | Lucro Líquido |
| CAPEX | Equipamentos (exibir CAPEX); Edifícios e Construções; Maquinário Produtivo; Tecnologia; Móveis e Utensílios | Lucro Líquido Ajustado |

Somente Receita Operacional Bruta, Receita Financeira e Outras Receitas **somam** no waterfall. Todas as demais linhas subtraem o valor positivo da base. Fórmulas: RL = RB − soma(deduções); Margem de Contribuição = RL − CMV − Custo de Envios − despesas variáveis; EBITDA Gerencial da referência = Margem de Contribuição − fixos; Lucro Líquido referência = EBITDA Gerencial + receitas financeiras e outras receitas − despesa financeira; Lucro Líquido Ajustado referência = Lucro Líquido − CAPEX. A referência não estabelece depreciação/IR separadamente neste waterfall; não chamar seu EBITDA de EBITDA contábil.

### 6.5 Ajustes que exigem decisão explícita

| Achado no espelho | Tratamento proposto na visão ajustada | Ponte |
| --- | --- | --- |
| “Retirada de Lucro” subtrai em fixos | Distribuição de lucro fora do resultado. | Adicionar de volta o valor retirado; lançar em movimentação patrimonial. |
| “Empréstimos pagos” entra em Despesa Financeira | Separar principal amortizado (fora da DRE) e juros/encargos (despesa financeira). | Adicionar principal de volta; juros mantidos uma vez, comprovados por contrato/extrato. |
| “Parcelamento Simples Nacional” em Outras Despesas | Desmembrar principal do tributo e juros/multa; detectar DAS já reconhecido. | Mostrar cada parcela e período de origem. |
| CAPEX após lucro como “Lucro Líquido Ajustado” | CAPEX em investimento/ativo; não rotular resultado após CAPEX como lucro líquido. | Conservar total de investimento e fluxo de caixa separados. |
| Provisões de 13º/férias presentes no mapa | A exclusão de provisões foi informada para um fechamento anterior; exigir decisão expressa por competência antes de estender a regra. Gastos e competência aprovados requerem classificação própria. | Excluir linha provisionada uma vez, mantendo memória e aprovação. |
| Mão de obra produção e frete cliente/reenvio | Conferir inclusão no custo do produto e natureza de receita/despesa antes de mover. | Impacto calculado por lançamento, sem ajuste automático cego. |
| Marketing no cartão | Reconhecer mês de competência da campanha e conciliar fatura do cartão/contas a pagar. | Diferença entre competência e caixa em memória. |

Para cada ajuste: `rule_version`, origem, justificativa, documento, aprovador, data e valor do efeito. O usuário deve poder percorrer “valor do painel → ajustes individuais → valor do Intelligence”, com a soma da ponte fechando em centavos.

## 7. Persistência, versões e contrato para a tela

**Reaproveitar:** batches/import e raw_records para imutabilidade da origem; financial_rule_versions para perfil da regra; calculation_runs para input_hash, período e status; statement_snapshots e statement_lines para publicação; statement_line_contributions para rastreio. Não transformar saldo atual de contas a receber/pagar em movimento histórico. Os normalizadores atuais de receivables e bling_cash_movements não carregam todos os campos de categoria, competência e tipo_lancamento exigidos; a migration aditiva `000025_dre_calculation_lineage.up.sql` cria as tabelas analíticas e a trilha de memória descritas abaixo sem alterar T02/T03/T04. Ela ainda não foi aplicada no banco local e não há writer/reader ligado ao worker ou à T06.

| Objeto (migration 000025) | Campos/índices essenciais |
| --- | --- |
| dre_source_rows | id, batch_id, source_table, source_record_id, payload_raw_ref, competencia, categoria_id, valor_original, valor, tipo, tipo_lancamento, situacao, origem_tipo, origem_id, pedido_id, nf_id, produto_id, qty, operation_type, included, exclusion_reason; unique(batch_id, source_table, source_record_id, item_key). Campos específicos ficam em tabelas especializadas ou JSON validado. |
| dre_inventory_events | run_id, event_id, source_row_id, event_date, priority, product_id, qty_delta, unit_cost_before, unit_cost_after, value_effect, bom_version, anchor_id, warning; unique(run_id, event_id). |
| dre_sale_costs | run_id, pedido_id, item_id, product_id, event_date, quantity, cmp_before, cmv, inventory_event_id, cost_completeness; índice(run_id, event_date). |
| dre_line_sources | snapshot_id, line_code, source_row_id ou inventory_event_id, category_id, contribution_amount, inclusion_rule, exclusion_reason, adjustment_id; FK e soma controlada por linha. A contribuição atual ligada apenas a accounting_entry_line_id/daily_balance_id não cobre todo CMV e vendas. |
| dre_adjustments | run_id, target_line_code, source_line_code, amount_signed, reason_code, evidence_ref, approved_by, approved_at, rule_version; preserva espelho imutável. |
| statement_lines (extensão) | availability/status/reason; valor NULL ou campo acompanhado de disponibilidade quando falta fonte. Jamais materializar ausência como 0. unique(snapshot_id, line_code, view_type). |
| dre_calculation_steps / dre_calculation_operands | statement_line_id, snapshot_id, execução, ordem, operação/fórmula, operandos vinculados a linhas ou constantes, valor, versão da regra e disponibilidade/motivo; resultados e operandos indisponíveis não armazenam valor numérico. |

**Transação:** extrair dados do Bling via leitura e/ou receber relatórios → importar lote no banco do Intelligence → validar → calcular em execução isolada no worker/serviço do Intelligence → reconciliar → gravar contribuições, linhas e snapshot atomicamente → publicar ponteiro “atual” apenas após gates. Revisão cria nova execução/snapshot; versão antiga permanece consultável. Idempotência por hash de arquivos, intervalo, perfil e versão de regra. Controle de acesso: Gisele pode importar/validar/publicar; diretores apenas ler snapshots liberados. Registrar usuário e hora de cada ação.

**Memória de cálculo obrigatória:** ao abrir detalhes de qualquer linha, permitir refazer o caminho do total até os lançamentos/documentos originais. Mostrar competência/período, perfil do demonstrativo, versão da regra, cada parcela incluída com valor e disponibilidade, operação/fórmula aplicada, valores intermediários e resultado; listar também exclusões e seus motivos, ajustes aprovados e documentos/linhas de origem. Preservar precisão e arredondamento definidos pela regra. Para tributos estimados, mostrar a competência a que se referem, pagamentos/documentos usados como base, cada etapa da fórmula aprovada, arredondamentos e identificação explícita de estimativa ou realizado. Não exibir uma fórmula tributária até a orientação especializada e a fonte correspondente serem registradas. A memória persistida é imutável junto da execução/snapshot; recálculo gera nova versão e mantém a anterior auditável.

**Estado no código em 2026-11-20:** `internal/financial.CalculateDRE` devolve memória aritmética para subtotais e EBITDA (operandos, operação, valores e disponibilidade do resultado). A migration 000025 permite gravar `amount = NULL` quando uma linha não está disponível e acrescenta armazenamento para linhas de origem, contribuições incluídas/excluídas, passos e operandos imutáveis, ligados por FKs à mesma execução e snapshot. A migration 000026 adiciona uma função restrita para finalizar a execução. `internal/drememory` agora grava e consulta atomicamente execução, snapshot, linhas e passos/operandos aritméticos, exigindo regra aprovada e hash das entradas. Esse adaptador ainda não grava fontes/contribuições, não está ligado ao worker e a migration não foi aplicada localmente. A view de capital de giro também deixa o total indisponível quando falta um de seus saldos, sem convertê-lo em zero. A T06 agora lista as linhas e abre a regra/fonte prevista para cada uma, incluindo a regra já registrada de que pagamento feito no mês 11 se refere à competência 10. Esse detalhe é uma especificação estática: não mostra valores nem lançamentos daquela execução e não deve ser chamado de memória persistida. A memória completa continua indisponível ao usuário. A fórmula para estimativa de tributos segue pendente da orientação do especialista.

**API interna sugerida:** `POST /dre/import-batches` (recebe arquivos e retorna checklist), `POST /dre/runs` (period=AAAA-MM, profile=reference|adjusted, batches, rule_version), `GET /dre/runs/{id}/issues`, `GET /dre/snapshots?period=AAAA-MM`, `GET /dre/snapshots/{id}/lines/{code}/sources`, `POST /dre/runs/{id}/approve`. Ajustar rotas ao transporte Named Pipe/API existente; não abrir conexão direta da tela com PG.

**Compatibilidade:** o CalculateDRE de Go recebe efeitos com sinal (receita positiva, gastos negativos), calcula lucro bruto e indicadores, e propaga indisponibilidade; fazer adaptador a partir das linhas gerenciais, preservando essa propriedade. O waterfall do painel tem blocos diferentes (margem, CAPEX pós-lucro e EBITDA chamado gerencial): modelar dois perfis/resultados versionados, não forçar encaixe que misture CAPEX ou retirada no NET_INCOME. EBIT/EBITDA ajustados exigem depreciação e classificação comprovadas; se ausentes, mostrar indisponível.

## 8. Ordem de implementação e pontos de aprovação

1. **Inventário do mês:** localizar os arquivos do período selecionado, as âncoras necessárias e o PGDAS da mesma competência; gerar manifesto de cabeçalhos, períodos, contagens, hashes, totais, colunas ausentes e correspondência com as 12 entidades. Entregar à Gisele; não alterar originais.
2. **Persistência:** migrações aditivas, importadores, IDs/datas/decimais, regras de acesso e idempotência. Testar reimportação idêntica, arquivo corrigido e proveniência.
3. **Mapa:** carregar as 76 categorias e 2 sentinels como configuração versionada; comparar catálogo Bling real e o mapa. Qualquer categoria usada sem regra, homônima ou pai com lançamento próprio vira pendência.
4. **Receita e deduções:** implementar seleção única compartilhada por cálculo e detalhe; entregar o período selecionado com/sem NF, reenviados, bonificações, reembolsos, DAS/GNRE/ICMS-ST e divergências. **Aprovação 1** de Gisele para classificação de venda.
5. **CMV:** carregar histórico desde 2025, posição e custo âncora, compras, fichas, OPs, vendas; construir eventos, CMP por item e auditoria de faltas. **Aprovação 2** para âncora, MOD, kits e regras de reenvio.
6. **Outras linhas:** categorias, caixas de embarque, fretes, salários, marketing, despesas e financeiro, com documento e competência. **Aprovação 3** para cada classificação controversa.
7. **Waterfall:** montar espelho e comparativo com o painel do outro projeto, linha a linha. Desvio = nosso espelho − valor da referência; discriminar diferença de fonte, data, filtro, custo, arredondamento e regra. Nenhum alvo ajustado artificialmente.
8. **Versão ajustada:** somente regras aprovadas, com ponte assinada por linha; revisão contábil e financeira. **Aprovação 4** do resultado gerencial do período selecionado.
9. **Interface e publicação:** T06 com período, status, visão, subtotal, drill-down e histórico; diretores apenas consulta. Confirmar a competência selecionada e repetir o processamento para outras competências com os mesmos contratos e regras versionadas, sem reutilizar valores de um mês em outro.

### Critérios de aceite verificáveis

- Mesmos arquivos e versão de regra produzem mesmo snapshot e mesmo valor em cada linha; importação repetida não duplica contribuições.
- Soma das contribuições incluídas = linha analítica; soma das categorias não-subtotal = linha gerencial; subtotal = fórmula do waterfall; ponte ajustada fecha exatamente.
- Uma venda com pedido e NF entra uma vez; bonificação, reenvio e remessa aparecem identificados no controle mesmo quando não alteram receita.
- Cada custo de venda mostra SKU, qty, data, CMP e eventos/âncora; produto sem custo não vira CMV zero confirmado.
- Valores do painel de referência podem diferir por erro ou fonte; toda diferença é mensurada, explicada e fica visível no relatório de reconciliação.
- Sem fonte completa, a linha e o lucro dependente permanecem indisponíveis; não existe “0,00” de conveniência.
- Teste para uma competência escolhida e repetição em outra competência: receita, DAS, custo de envios, CMV, fixos, financeiro, CAPEX, totais e três exemplos de rastreamento por linha; comparar também diretoria read-only sem capacidade de modificar/publicar.

## 9. Riscos concretos a mostrar antes de fechar números

| Regra do painel | Risco | Bloqueio/medição |
| --- | --- | --- |
| União títulos + caixa por tipo_lancamento | NF duplicada pedido/nota; complemento não casado por identidade. | Listar pares por NF, pedido, valor e data; diferença contra faturamento. |
| Compra por pedido; NF por produto sem pedido | NFs reais daquele produto omitidas e/ou compras duplicadas. | Anti-join documento/item; quantificar custo e quantidade. |
| OP por data_previsao_final | Competência de produção diferente da real. | Comparar data final real e posição de estoque. |
| Saldo negativo truncado e CMP zero | Lucro artificialmente alto. | Registrar quantidade insuficiente e CMV faltante, status pendente. |
| MOD no CMP e folha em fixos | Dupla contabilização potencial. | Ficha e custo unitário comprovados antes da decisão. |
| Frete de reenvio crédito mapeado para linha subtrativa | Sinal econômico possivelmente invertido. | Exibir parcela isolada e resultado de reclassificação aprovado. |
| Empréstimo/retirada/parcelamento no resultado | Lucro da referência contaminado por saída de caixa/patrimônio. | Separar principal, juros, distribuição e tributo de origem. |
| Precisão Python float vs Go decimal | Diferenças pequenas mesmo com mesma regra. | Relatar centavos por etapa e política de arredondamento; não forçar igualdade de centavos até comparar eventos. |

**Limite da versão 1:** sem inspecionar os CSV/XLS/PGDAS da competência selecionada e os valores das execuções correspondentes do outro painel, esta especificação comprova a lógica do código, não os montantes daquele período. O inventário e a conciliação da etapa 1 convertem isso em números auditáveis.

## 10. Matriz de implementação: 76 categorias por ID

Transcrição estruturada de `ESPECIFICACOES` cruzada ao `mapa_contabil.json` da revisão estudada. “Pagar + caixa débito” é a regra padrão; “Receber + caixa crédito” é exceção explícita. Pai é apenas subtotal na analítica. O produto de todos os mapeamentos é a linha indicada; perfis ajustados só mudam após aprovação. Além das 76 categorias, `-1` = Mercadoria Vendida e `-2` = Custo de Envios, motores próprios sem categoria Bling.

| # | ID Bling | Nome na regra | Linha da DRE | Fonte | Pai? |
| --- | --- | --- | --- | --- | --- |
| 1 | 12805089610 | Venda de produtos | Receita Operacional Bruta | Receber + caixa crédito | Não |
| 2 | 14675501032 | DAS | DAS | Pagar + caixa débito | Não |
| 3 | 14712313565 | GNRE | GNRE | Pagar + caixa débito | Não |
| 4 | 14662426036 | ICMS - ST | ICMS-ST | Pagar + caixa débito | Não |
| 5 | 14708854913 | Despesas Tributárias | Outros Tributos | Pagar + caixa débito | Não |
| 6 | 12805089623 | Descontos Incondicional | Descontos Incondicional | Pagar + caixa débito | Não |
| 7 | 12805089649 | Água | Ocupação | Pagar + caixa débito | Não |
| 8 | 12805089652 | Aluguel | Ocupação | Pagar + caixa débito | Não |
| 9 | 12805089657 | Energia elétrica | Ocupação | Pagar + caixa débito | Não |
| 10 | 12805089663 | Material de escritório | Consumo e Manutenção | Pagar + caixa débito | Não |
| 11 | 14709805883 | Viagens | Viagens e Transporte | Pagar + caixa débito | Não |
| 12 | 14654434674 | Associação | Outras Despesas | Pagar + caixa débito | Não |
| 13 | 12805089677 | Internet | Ocupação | Pagar + caixa débito | Não |
| 14 | 12805089632 | Alimentação | Benefícios Gerais | Caixa apenas | Não |
| 15 | 14574663307 | Contabilidade | Serviços Profissionais | Pagar + caixa débito | Não |
| 16 | 12805089659 | Limpeza | Consumo e Manutenção | Pagar + caixa débito | Não |
| 17 | 14742571971 | Manutenção | Consumo e Manutenção | Pagar + caixa débito | Não |
| 18 | 12805089666 | Material de uso e consumo | Consumo e Manutenção | Pagar + caixa débito | Não |
| 19 | 14651501556 | Seguro fiança aluguel | Ocupação | Pagar + caixa débito | Não |
| 20 | 12805089669 | Serviços de terceiros | Serviços Profissionais | Pagar + caixa débito | Não |
| 21 | 12805089672 | Software | Software | Pagar + caixa débito | Não |
| 22 | 14710254275 | Transporte Adm | Viagens e Transporte | Pagar + caixa débito | Não |
| 23 | 14723831735 | Salários Administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 24 | 12805089689 | Salários Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 25 | 14710985091 | INSS | Mão de Obra Administrativo | Pagar + caixa débito | Sim; subtotal, não somar de novo |
| 26 | 14723834711 | INSS Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 27 | 14723834712 | INSS Administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 28 | 14723832882 | Provisões Administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Sim; subtotal, não somar de novo |
| 29 | 14711102848 | Provisões Produção | Mão de Obra Produção | Pagar + caixa débito | Sim; subtotal, não somar de novo |
| 30 | 14723833327 | 13º administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 31 | 14723833329 | FGTS Administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 32 | 14724551786 | Férias Administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 33 | 14723833524 | Rescisões e Distratos Administativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 34 | 14711102849 | 13º Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 35 | 14710985090 | FGTS Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 36 | 14711102850 | Férias Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 37 | 14708862679 | Rescisões e Distratos Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 38 | 14676619729 | Acrescimo | Receita Financeira | Receber + caixa crédito | Não |
| 39 | 14667746967 | Degustação | Outras Desp. Comerciais | Pagar + caixa débito | Não |
| 40 | 14622282327 | Desconto comercial | Outras Desp. Comerciais | Pagar + caixa débito | Não |
| 41 | 13420258281 | Empréstimos pagos | Despesa Financeira | Pagar + caixa débito | Não |
| 42 | 14708862678 | Exames | Benefícios Gerais | Pagar + caixa débito | Não |
| 43 | 14651924946 | Feiras | Outras Desp. Comerciais | Pagar + caixa débito | Não |
| 44 | 14676620728 | Multa | Despesa Financeira | Pagar + caixa débito | Não |
| 45 | 14712110952 | Parcelamento Simples Nacional | Outras Despesas | Pagar + caixa débito | Não |
| 46 | 12805089685 | Plano de saúde | Benefícios Gerais | Pagar + caixa débito | Não |
| 47 | 13617374919 | Promotoria de Vendas | Promotoria de Vendas | Pagar + caixa débito | Não |
| 48 | 14719210154 | Pró Labore | Pró Labore | Pagar + caixa débito | Não |
| 49 | 14686015776 | Rapel | Outras Desp. Comerciais | Pagar + caixa débito | Não |
| 50 | 14711009838 | Retirada de Lucro | Retirada de Lucro | Pagar + caixa débito | Não |
| 51 | 14708873118 | Treinamento e Capacitação | Benefícios Gerais | Pagar + caixa débito | Não |
| 52 | 12805089690 | Vale alimentação | Benefícios Gerais | Pagar + caixa débito | Sim; subtotal, não somar de novo |
| 53 | 14723833588 | Vale Alimentação Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 54 | 14723833587 | Vale Alimentação administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 55 | 12805089691 | Vale transporte | Benefícios Gerais | Pagar + caixa débito | Sim; subtotal, não somar de novo |
| 56 | 14723835713 | Vale transporte Produção | Mão de Obra Produção | Pagar + caixa débito | Não |
| 57 | 14723835712 | Vale transporte Administrativo | Mão de Obra Administrativo | Pagar + caixa débito | Não |
| 58 | 14012222439 | Rendimentos | Receita Financeira | Receber + caixa crédito | Não |
| 59 | 13410620780 | Marketing / Publicidade | Marketing | Pagar + caixa débito | Não |
| 60 | 14653668920 | Uniformes | Benefícios Gerais | Pagar + caixa débito | Não |
| 61 | 14708855109 | Outras Despesas | Outras Despesas | Pagar + caixa débito | Não |
| 62 | 14676994636 | Juros Recebidos | Receita Financeira | Receber + caixa crédito | Não |
| 63 | 12805089622 | Devoluções de vendas | Devoluções de vendas | Pagar + caixa débito | Não |
| 64 | 13410775589 | CAPEX | Equipamentos | Pagar + caixa débito | Não |
| 65 | 14697197416 | Frete pago por cliente | Frete pago por cliente | Pagar + caixa débito | Não |
| 66 | 14636464955 | Frete pago Majucau | Frete pago Majucau | Pagar + caixa débito | Não |
| 67 | 12805089640 | Comissões | Comissões | Pagar + caixa débito | Não |
| 68 | 14651548827 | Receita diversas | Outras Receitas | Receber + caixa crédito | Não |
| 69 | 12805089692 | Juros pagos | Despesa Financeira | Pagar + caixa débito | Não |
| 70 | 12805089618 | Indenização de seguro | Outras Receitas | Receber + caixa crédito | Não |
| 71 | 14742487763 | Frete pago pelo cliente - reenvio | Frete pago por cliente | Receber + caixa crédito | Não |
| 72 | 14742571775 | Edifícios e Construções | Edifícios e Construções | Pagar + caixa débito | Não |
| 73 | 14742571578 | Maquinário produtivo | Maquinário Produtivo | Pagar + caixa débito | Não |
| 74 | 14742571776 | Tecnologia | Tecnologia | Pagar + caixa débito | Não |
| 75 | 14744260089 | Móveis e Utensílios | Móveis e Utensílios | Pagar + caixa débito | Não |
| 76 | 14648938031 | Taxas de Marketplaces | Taxas de Vendas | Pagar + caixa débito | Não |

Há ainda categorias catalogadas fora das 76 (transferências, ajustes, compras de insumos, componentes de embalagem, itens sem movimento e categorias pendentes). Não transformar ausência na matriz em exclusão automática: cada ID novo ou movimentado exige motivo, versão e aprovação. O código de referência declara “Tarifa bancária” e “Combustíveis e lubrificantes” como pendências ainda sem especificação; verificar movimentos da competência selecionada e exibir eventual lacuna.

**Entrega de controle:** esta é uma instrução para implementação e revisão gradual. O desenvolvedor deve devolver em cada aprovação: arquivos usados, contagens e totais de origem, tabelas geradas, diferenças, rastreabilidade por linha, problemas impeditivos e valores resultantes; nenhuma linha pode ser aceita apenas porque “bateu” com o painel.

## Nota de implementação — 20/11/2026

Foi criado o adaptador `internal/drememory` para persistir e reler, de forma atômica, linhas e passos aritméticos de uma execução DRE, com status de disponibilidade e operandos. A migration `000026_dre_run_finalization.up.sql` restringe a finalização do cálculo a execuções DRE com snapshot; a allowlist do runtime foi ajustada para conceder apenas leitura/escrita necessárias e execução da função de finalização. Isso ainda não implementa importação/classificação das fontes, contribuições `dre_line_sources`, chamada pelo worker ou consulta pela T06; logo, não é memória de uma execução financeira real nem libera confirmação. A fórmula dos tributos segue pendente do especialista.

