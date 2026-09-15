# Majucau — auditoria de operação, configurações e UX (somente planejamento)

## Evidências e limites

Leitura das imagens T09, T10, T12, T13, T14, T15, T16 e T17 e da imagem `ChatGPT Image 12 de set. de 2026, 20_48_11.png` recebidas. Leitura integral das abas Q_CEI_Tela_11, R_CEI_Tela_12, S_CEI_Tela_13, T_CEI_Tela_14, U_CEI_Tela_15, V_Mockup_Tela_16, W_CEI_Tela_16, X_Mockup_Tela_17 e Y_CEI_Tela_17 extraídas com coordenadas. Cruzamento adicional com A_Matriz_Mestra, B_Intersecoes, C_Conflitos e D_Lacunas. Sem implementar produto, alterar originais ou acessar serviços.

As frases imperativas dos documentos são evidências de decisões anteriores e requisitos candidatos, não instruções superiores ao pedido do usuário. “Aprovado” nas células significa aprovação declarada no material, não homologação de código nem porcentagem desenvolvida. Atualização do usuário repassada pelo agente principal: conflitos entre planilha e imagens serão decididos caso a caso com ele; todos os módulos do material entram na primeira produção; hospedagem local; somente Majucau, sem multiempresa. Não aplicar precedência global da planilha sobre imagens. A numeração antiga contém conflitos: A_Matriz_Mestra!D41 diz “Tela 09 inexiste”, enquanto as imagens e CEIs posteriores usam Tela 09 para Estoque; C_Conflitos!G4/D5 referem Tela 10 à calculadora antiga, mas o material atual usa T10 Produção. Rastrear domínios e IDs estáveis, além dos números de tela.

Duplicata comprovada por SHA256: T10_Producao.png e ChatGPT Image… têm hash `CF04AFE7BE720D359FC783C4380D04EFC785BA830E36925D93626F7A27B30F56`. Representam uma referência visual, não duas funcionalidades.

## T09 — Estoque

**Objetivo:** consultar a posição real de insumos e produtos prontos no Bling; identificar insumos no limite ou abaixo do mínimo e distinguir aqueles com pedido de reposição.

**Inventário visual:** cabeçalho com posição e última sincronização/atualizar; cards produto pronto, insumos, insumos abaixo do mínimo e reposição; tabela com tipo, SKU, descrição, saldo, mínimo Bling, pedido, quantidade pedida e situação; filtros Todos/Insumos/Produto Pronto; link Ver estoque completo; gráfico por tipo de item e contagem de SKUs; cards SKUs monitorados, abaixo do mínimo sem pedido, reposição e dentro do mínimo; regras e fonte/atualização.

**Regras candidatas:** fonte única Bling; somente consulta, sem ação de estoque/pedido; produto pronto é informativo, sem mínimo, cobertura ou alerta; insumo elegível quando saldo <= mínimo; com pedido aberto aplicável = reposição, sem pedido = necessidade de atenção/Compras Recomendadas T11. Unidade por SKU é original, sem somar kg + un + L em um total físico. Dados ausentes não são zero. A regra compartilhada está explícita em Q!B16:B19, Q!G16 e W!C21:C22.

**Inconsistências:** o gráfico contém placeholders “XX SKUs”, “X SKUs”, “Y SKUs”; 52%/48% exige denominador por SKUs, não por unidades físicas incompatíveis. Card 128,4 t não pode ser total de todos os insumos se incluem embalagens em unidades; só seria válido como agrupamento explicitamente compatível. “Dentro do mínimo: 50 insumos” parece calculado por 56 SKUs monitorados menos seis abaixo, mas o universo de 56 inclui produtos prontos; confirmar denominador. “Abaixo do mínimo” também inclui igualdade pelo texto da regra; ajustar rótulo explicativo conforme conteúdo aprovado.

**Aceite:** saldo igual ao mínimo entra na mesma regra de falta; produto pronto jamais ganha alerta/minímo por inferência; contagem de sem pedido coincide com T11 na mesma posição/filtros; um pedido cancelado não pode ser tratado como aberto aplicável; saldo/mínimo ausente mostra indisponibilidade e não “OK”; soma física somente dentro de unidade compatível; gráfico declara seu denominador; atualização falha conserva snapshot válido com indicação de desatualizado.

**Decisões necessárias:** o que torna um pedido “aberto aplicável” por SKU, empresa, depósito, situação, saldo a receber e entrega parcial; qual saldo Bling (físico, disponível, reservado) é o oficial; tratamento de SKU sem tipo/mínimo/unidade; quais depósitos entram. Não mudar a regra aprovada de exclusão com pedido aberto sem decisão explícita.

## T10 — Produção

**Objetivo:** confrontar produção real consultada no Bling com metas internas Majucau por SKU/período.

**Inventário visual:** seletor mensal, atualizar, último sync, usuário/perfil; cards ordens em produção e concluídas, produtos com meta e metas atingidas; busca SKU/produto, filtros status e unidade; tabela SKU/produto/unidade/meta/planejado Bling/em andamento/concluído/% da meta/situação; distribuição de metas; indicadores SKUs produzidos/com meta/meta atingida/abaixo da meta; botão Definir metas de produção; notas de regra.

**Regras candidatas:** produção vem do Bling, meta interna por SKU e período; percentual usa somente concluído/meta; não cria, altera ou cancela OP; não exporta; T14 não pode criar alertas próprios de produção, mesmo quando abaixo de meta (T!A22:H22 e A47:H47). Metas precisam fluxo específico: W!B36:D36 referencia destino, W!C14 não transforma T16 em editor genérico.

**Inconsistência comprovada:** a tabela completa declara 11 de 11 produtos. Contagem visual de seus status: 3 atingidas (BOMBOM-MIX, DRAGEADO-01, LEITE-01), 5 em andamento (TRUFA-45, BARRA-70, TABLETE-50, TABLETE-100, MANTEIGA-01), 3 abaixo (BARRA-BRANCA, OVO-250, NIBS-01). Gráfico/cards dizem 6 atingidas, 3 em andamento e 2 abaixo. Percentuais corretos para as linhas seriam 27,3% / 45,5% / 27,3%, se esse for o universo. Estes números são ilustração e não regra.

**Semântica pendente:** TABLETE-100 e MANTEIGA-01 aparecem “Em andamento” com zero produção em andamento; OVO-250 e NIBS-01 têm produção em andamento e estão “Abaixo da meta”. Logo, status de OP e situação da meta são dimensões distintas; falta a fórmula/classificação por prazo. Não deduzir limiar arbitrário de 80%. Planejado versus concluído+andamento não reconcilia em vários SKUs (ex.: TRUFA 7.500 planejado e 6.800+1.200=8.000), podendo refletir conceitos/períodos diferentes; definir sem forçar igualdade.

**Aceite:** somente concluído aumenta realização; 0 de meta versus meta não cadastrada são estados distintos; ausência de meta não gera divisão por zero; contagem de OP distinta de contagem de SKU; somas e donut vêm do mesmo conjunto filtrado; percentual pode superar 100% quando produção supera meta; filtros não misturam unidades; metas editadas ficam versionadas e auditadas, sem escrita no Bling.

**Decisões:** categorias de produto elegíveis à produção (ex.: leite em pó e manteiga podem ser insumos, mas aparecem na tabela); mapeamento de status/datas Bling; OP parcial/conclusão/reabertura/cancelamento e mês de reconhecimento; regra de situação da meta; permissão e fluxo para definir meta, eventual aprovação, vigência retroativa e relação T10↔T12. Não assumir que meta comercial de T12 substitui automaticamente meta de produção T10.

## T11 — Compras (contrato disponível, imagem ausente no lote)

**Inventário contratual:** Compras Recomendadas com insumo, estoque atual/mínimo, última compra atendida, último preço faturado, prazo histórico individual, quantidade recomendada, valor estimado e prioridade; detalhe interno por linha; filtro/prioridade; exportação única Excel no bloco; bloco informativo de cacau/mercado; atualizar. Sem Exportar/Ajuda no cabeçalho, gerar pedido, necessidade líquida futura, lead time geral, tabela própria de fornecedores e antigos blocos inferiores (Q!A5:H12, A60:H67).

**Regra fechada documentalmente:** somente insumo com saldo<=mínimo Bling e sem pedido aberto aplicável; cacau usa a mesma regra; risco de mercado não cria recomendação/pedido. Último preço = unitário efetivamente faturado da NF-e de entrada vinculada à última compra atendida. Prazo real = data/hora de alteração manual para ATENDIDO menos data/hora do pedido (Q!A16:H19, A26:H31).

**Pendências:** quantidade recomendada depende de fórmula, janela, mínimos e múltiplos (L-17; Q!G29); prazo histórico agregado e faixas de prioridade (L-18); fontes, localização dos fornecedores, comparabilidade e limiares de risco cacau (L-19). Não publicar sugestões numéricas sem regra: PENDING_RULE, UNAVAILABLE quando faltar dado. Maior prazo não recebe prioridade menor se demais condições iguais (Q!A76:H77).

**Cacau:** clima, safra, estoque global, preço internacional, câmbio; sinal adverso relevante pode elevar risco, enquanto preço pago menor que referência comparável sinaliza oportunidade. Necessita compatibilidade de qualidade/produto, unidade, moeda, base e data; referências devem ser auditáveis (Q!A41:H47). Não é cotação/compra automática.

**Exportação/aceite:** uma aba com lista completa de recomendações e metadados mínimos de posição/filtro, apenas nove campos aprovados; sem exportar bloco de cacau; nenhuma mutação Bling; detalhe até PO/NF-e real; ausência de fonte não vira compra fictícia; mesma posição da T09 gera mesma quantidade de recomendações (Q!A35:H37, A71:H82). Esclarecer “lista completa” sob filtros e paginação, mantendo o contrato de contexto.

## T12 — Projeções

**Objetivo:** estudar vendas futuras e comparar com metas planejadas, quantificando produção e estoque por SKU em 12 meses.

**Inventário visual:** ano, atualizar, Exportar, Ajuda; cards faturamento Estudo/Meta, produção Meta, estoque final Meta, WAPE; gráfico mensal Realizado/Estudo/Meta com eventos; qualidade WAPE/Bias/MASE e status; escolha meta faturamento/SKU/mista e Aplicar meta e recalcular; abas Projeção por SKU/Sazonalidade/Insumos (1 ano)/Resumo financeiro; busca, período, SKUs, ajuste de metas em lote, edição por linha; tabela venda/produção/estoque Estudo/Meta/delta, preço médio e faturamento Meta; explicação de planejamento sem criar OP/compra.

**Contrato:** R!A5:H10 define 24 meses de histórico, 12 meses futuros, somente SKU e dois cenários na mesma base/período/versão. R!A14:H18 exige fontes Bling auditadas; cancelados não viram venda realizada. Calendário sazonal versionado por ano (Páscoa móvel), avaliação de períodos normais e críticos (R!A22:H24). Não há família de produtos como entidade. Alterar meta muda cenário interno, não fato externo; “read-only” aqui significa não alterar fatos do Bling, não proibir persistência interna de metas/versões.

**Cálculo:** venda projetada é início; produção deriva da demanda e não desconta estoque pronto atual; estoque_t=estoque_(t−1)+produção_t−venda_t. Primeiro saldo vem do Bling; se ausente, estoque projetado UNAVAILABLE (R!A35:H38, A48:H55). Regra exata que transforma demanda em produção merece esclarecimento: se produção for igual à venda em cada mês, estoque permanece constante; as quantidades diferentes da imagem não explicam arredondamento, sazonalidade de fabricação, perdas ou defasagem. Não inventar essas políticas.

**Métricas:** WAPE=Σ|real-prev|/Σreal ×100; Bias assinado conforme versão; MASE compara erro com baseline sazonal; backtest temporal só usa passado disponível em cada corte (R!A28:H31). Limiares de aceitação ainda L-20, preço/distribuição e tratamento de overrides L-21. Especificar caso de denominador zero, SKU novo, poucos meses, vendas negativas/devoluções, ruptura e preço ausente. Métrica existente não equivale a modelo aprovado; gráfico de projeção deve indicar PROJECTED mesmo que a métrica do backtest seja confirmada.

**Metas:** faturamento é distribuído com mix/sazonalidade/preço do estudo; SKU substitui quantidade do estudo no cenário Meta; misto combina global com overrides explícitos. Recalcular gera nova versão interna, preserva história e auditoria (R!A42:H44, A61:H64). Decidir se override altera o total global ou redistribui residual; metas que não cabem precisam validação, sem ajuste silencioso. Arredondamentos respeitam unidades e total reconciliável.

**Exportação fechada:** XLSX com Resumo_Cenario, Projecao_Vendas_SKU, Projecao_Producao_SKU, Estoque_Projetado_SKU, Metas, Qualidade_Estudo, OP_Projetada e Metadados (R!A69:H76). Dados estruturados do snapshot/cenário; PROJECTED visível; OP marcada “NÃO É ORDEM DE PRODUÇÃO REAL”; gráficos, se houver, nativos do Excel (R!A80:H82).

**Inconsistências visuais:** tabela visível totaliza R$3.163.900 Meta, produção Meta 490.000 e estoque Meta 47.000; cards mostram R$8.400.000/1.245.000/590.000 sem explicitar universo diferente (filtro “Todos os SKUs”). Pode ser amostra, mas deve ser identificado. Soma das seis linhas e seus totais internos reconcilia; por isso não tratar diferença como erro aritmético global provado sem definir universo. Delta de vendas é 18,6%, de produção é 21,9% nas linhas; card de produção mostra 18,6%. Gráfico histórico compara meses de anos distintos sem explicitar ano do realizado. Status CONFIRMED do Estudo na imagem não pode apagar a natureza projetada.

**Escopo em aberto:** abas Insumos (1 ano), Sazonalidade e Resumo financeiro não têm conteúdo completo nas imagens/CEI; consumo futuro de insumos exige composição/ficha técnica, rendimento/conversão e versões, se aprovado. Nenhuma fórmula de insumo deve ser criada por inferência. Confirmar se exportação inclui estes futuros detalhes além das oito abas contratuais.

**Aceite:** reproduzir cenário por versão; não treinar com futuro; nenhuma métrica ilustrativa; calendário móvel correto; cenários/filtros preservados no download; totais do contexto batem com detalhes; estoque inicial ausente não vira zero; nenhuma projeção cria operação real.

## T13 — Lançamentos & Pendências

**Inventário:** KPIs abertos/divergências/bloqueados/não mapeados; tabela módulo/origem/descrição/status/responsável/ação; gráfico por status; cards semana crítica, saldos iniciais, perdas, taxas Nuvem e matches; atualizar, exportar, ajuda, Ver e lista completa.

**Contrato:** central de exceções, não um CRUD genérico de lançamentos; OPEN decisão não fechada; BLOCKED falta condição obrigatória; PENDING análise/processamento; UNMAPPED dado sem classificação; STALE última posição desatualizada; DIVERGENT evidências não conciliadas (S!A11:H17). Não usar padrão fictício, conta coringa, zero ou resolver automaticamente. “Resolver” abre fluxo de origem autorizado, com dependências/evidências; cada tipo precisa transição própria L-23 (S!A30:H35).

**Modelo/aceite:** pending_id estável, módulo, origem, descrição, responsável/contexto, status, timestamps, regra e dependência; KPIs/gráfico da mesma lista/filtros; abertos em S!B21 são OPEN/PENDING, não somente OPEN; relação com alert_id opcional e batch_id/source_id em falha técnica. Alerta não é automaticamente pendência (S!A39:H40). Clique Ver não fecha nem altera Bling. Reincidência, cancelamento, responsável e SLA precisam política por tipo. Exportação permanece decisão L-22; não inventar formato/colunas (S!A34:H34, A50:H50).

**Inconsistências:** donut total 53, mas categorias 28+11+6+5+8=58; percentuais 52,8+20,8+11,3+9,4+15,1=109,4%. Há sobreposição ou erro, ambos incompatíveis com fatias exclusivas não explicadas. Cabeçalho “OPEN / BLOCKED” confunde duas condições e mostra 6; CEI separa bloqueados. Menu destaca Visão Executiva estando em T13. Exemplo “OP sem consumo vinculado” não autoriza criar alerta/validação de produção fora do contrato; T10 não possui alertas próprios.

## T14 — Alertas

**Inventário:** cards Críticos/Atenção/Informativos/Resolvidos 7 dias; tabela módulo/alerta/impacto/prioridade/atualizado/status; donut por módulo; cards caixa/recebíveis/estoque/compras; recorte alertas importantes; navegar origem/atualizar/exportar/ajuda.

**Contrato:** agregação de regra de origem aprovada, sem thresholds paralelos (T!A11:H15). Estoque <=mínimo sem pedido é ATENÇÃO; com reposição é INFORMATIVO; falha da fonte segue SLA e prioridade da fonte; produção não gera alertas; cacau depende L-19 (T!A19:H24). Portanto “5 SKUs abaixo do mínimo: Crítico” na imagem não deve prevalecer sobre o CEI sem reconciliação de versão. Resolvido exige condição encerrada/fluxo autorizado e mantém histórico; sem botão genérico de resolução.

**Aceite:** alert_id/source_ref estáveis; mesmo alerta da origem aparece uma única vez nas contagens e recortes; prioridade e situação são dimensões distintas; resolvidos em janela histórica não entram no denominador de ativos; navegação preserva contexto; T10 abaixo de meta não gera alerta novo. Exportação L-22 pendente (T!A41:H41, A51:H51).

**Inconsistência:** ativos do topo 8+17+26=51; distribuição por módulo soma57. Se períodos/universos diferentes, explicitar; se mesmo contexto, recalcular. Os14 resolvidos pertencem a outra janela e não corrigem a diferença. Financeiro=16 no gráfico e Caixa=12 no card podem ser subconjunto, mas precisam definição.

## T15 — Importações / Integrações

**Inventário:** posição/última sync, exportar/ajuda; lotes processados, registros aceitos/rejeitados, freshness%; tabela origem/lote-arquivo/data referência/importação/status/aceitos/rejeitados/hash; distribuição de lotes por estado; idempotência%; cards fontes e logs auditáveis; alertas de falha/rejeição/arquivo desatualizado; detalhes/lista completa/reprocessamento contratual.

**Contrato:** mostrar somente fontes aprovadas/configuradas; fonte principal Bling e arquivos controlados. Deduplicação por hash e chave de negócio, não apenas nome de arquivo (U!A11:H15). OK execução válida no prazo; WARNING parcial/rejeições; ERROR sem resultado válido correspondente; STALE fonte fora do SLA próprio; PENDING_REPROCESS tentativa controlada (U!A19:H23). ERROR não apaga último snapshot válido; STALE não zera.

**Dados:** batch_id estável, source/provider, nome/referência sem caminho sensível, hash, lidas/aceitas/rejeitadas, motivo de rejeição por linha, sync_ts/last_success_at, business_key/idempotency_key (U!A27:H34). Distinguir lote lógico e tentativa de processamento para histórico; política de aceitação parcial, duplicadas/ignoradas e contagem deve ser formalizada. Mesmo hash não pode impedir recuperação de falha sem publicação: deduplicar fatos e registrar nova tentativa, conforme U!A40:H40 e A52:H55.

**Interações/aceite:** consulta de lote/erros; reprocessar exige autorização, auditabilidade, idempotência e origem preservada; atualizar saúde não é implicitamente executar importação; evidências T15 são vinculadas a T13/T14 quando suas regras demandam; mesmo evento não vira contagens duplicadas (U!A38:H48). Arquivo repetido e mesmo evento em arquivos diferentes não duplicam fatos. Segredos/caminhos não aparecem nos logs/telas. Exportação L-22 ainda aberta; freshness/retry/timeout por fonte L-24. Ciclo origem imutável versus mover para processados é L-12.

**Lacunas:** fórmula/denominador freshness 96,8%; medição de idempotência 100% (requisito de integridade não deve virar KPI constante); catálogo real das fontes; formatos e volumes; agendamento/tempo de atualização; correção de registro já aceito em novo arquivo; controle concorrente de reprocessamento; critérios WARNING vs ERROR; arquivo vazio; datas/timezones; conteúdo de Ajuda. Lotes do donut 64+14+8+6+4=96 estão coerentes (percentuais têm arredondamento).

## T16 — Parâmetros

**Inventário:** cards histórico24m/horizonte12m/investimentos ativos/fontes; tabela grupo/parâmetro/valor/origem/status; mapa por grupo; atalho fonte oficial, SKU, metas produção, meta faturamento/SKU e extratos; regras. Sem Exportar/Ajuda/Salvar.

**Contrato:** central de referência/governança de regras já aprovadas, não editor genérico; vigência exige fonte/decisão; números ilustrativos não hardcoded (W!A11:H14). Mínimo pertence ao Bling, não a cadastro duplicado Majucau; categorias Transferência/Rendimentos; recebimentos futuros/perdas e extratos vêm de fontes externas controladas, sem expor segredo/caminho (W!A18:H26). Cards dinâmicos reconciliam com domínios, investimento identificado por cadastro próprio; metas produção e cenário Meta têm destinos específicos (W!A30:H38).

**Edição pendente L-25:** por parâmetro determinar ação/perfil/aprovação/motivo/vigência/rollback; anterior e novo, ator, data/hora, versão e histórico preservados (W!A48:H49). A central não oferece mudança antes de contrato específico. Não confundir dependência “fluxo não desenhado” com necessidade de desenvolver tal fluxo agora sem requisitos.

**Inconsistência:** nove linhas da tabela: Projeções3 (33,3%), Investimentos2 (22,2%), Estoque1 (11,1%), Compras1 (11,1%), Integrações2 (22,2%). Donut marca33/22/17/17/11. Ajustar dados/denominador, não presumir novo layout. Falta reserva mínima na lista visível pode ser amostra; “Ver completos” deve tornar universo claro. Referências V!A70:B73 apontam arquivo antigo `/mnt/data/dashboard_de_parâmetros_majucau.png`; registrar correspondência com T16 fornecida, não assumir arquivo local existente.

## T17 — Usuários

**Inventário:** ativos/perfis/bloqueados/convites pendentes; tabela nome/perfil/login/último acesso/status/ação; donut por perfil; cards Admin/Financeiro/Operação/Leitura/Sem acesso; Ver permissões/convite/lista/perfis; regras. Sem Exportar/Ajuda.

**Contrato:** quatro perfis conceituais, permissões efetivas no backend e negação por padrão; toda ação deve obedecer ao contrato do domínio, inclusive perfis Operação não ganham escrita no Bling porque têm “operação” no nome (Y!A11:H14, A25:H28). Nome/email não substituem ID estável. ATIVO/BLOQUEADO/PENDENTE; último acesso ausente não é data fictícia; consultar permissões/convite não muda nem reenvia/cancela (Y!A32:H37, A48:H52). Alterar perfil/status exige fluxo e permissão administrativos específicos e auditoria antes/depois/ator/timestamp (Y!A41:H44).

**Lacuna L-26:** matriz módulo×ação, quem pode ver valores financeiros, editar metas, reprocessar, exportar, gerir parâmetros e usuários; provedor de identidade, login, ativação, expiração de convite, bloqueio/desbloqueio, MFA, recuperação, sessão/revogação e proteção do último administrador (Y!A56:H59). Definir acesso individual/múltiplos perfis e eventual escopo por depósito; usuário já definiu somente Majucau, sem multiempresa. Contas bloqueadas com sessões já emitidas devem perder acesso conforme contrato; testar chamada direta de backend, não só ocultar botão.

**Contagens:**12 ativos e distribuição de perfis soma12; bloquear1 e pendentes2 podem significar15 contas totais, mas tabela inclui bloqueado com perfil e não explicita universo do donut. Fixar “perfis entre usuários ativos” ou total de contas; não presumir erro aritmético. Dados pessoais da imagem são exemplos. Referência X!A70:B73 contém nome antigo `/mnt/data/painel_de_usuários_majucau.png`; reconhecer associação com T17 recebida.

## Questões transversais de UX e escopo

1. Sidebar varia entre telas: T12/T13 incluem Inteligência, Simuladores, Relatórios, Produtos(SKU), Projeções; outras omitem esses itens. Balanço Patrimonial aparece em imagens, mas S/T/U/W/Y!B7 o excluem como legado. Esta discrepância deve ser decidida caso a caso pelo usuário, sem exclusão silenciosa. Aplicações e Calculadora aparecem com ordens/nomes diferentes. Consolidar mapa canônico de rotas por perfil, com IDs de domínio independentes da antiga numeração. Todos os módulos do material foram incluídos na primeira produção pelo usuário; componentes sem conteúdo detalhado precisam especificação, não descarte automático.
2. Design Lock do material preserva layout, cores, tipografia, logos, proporções; conteúdo legado e números ilustrativos podem ser corrigidos. Confirmar com usuário vigência e qual referência visual canônica quando logos/sidebars diferem. Responsividade, zoom, foco/teclado, contraste e estados técnicos não estão demonstrados; registrar propostas concretas antes de redesenho. A análise atual não redesenha.
3. Traduzir estados técnicos para linguagem simples mantendo códigos internamente. Separar situação de negócio, qualidade/freshness, estado de processamento e confirmação da regra/modelo; CONFIRMED não é sinônimo universal de pago, fresco, projetado aprovado ou produção concluída.
4. Todas consultas precisam carregando, vazia, sem permissão, indisponível, desatualizada, erro parcial, ausência de configuração, busca sem resultado e contexto invalidado. Não transformar ausência de dado em zero/verde/100%.
5. Cada CTA exige contrato: origem/destino, contexto/filtros preservados, permissão, leitura versus mutação interna/externa, erro, auditoria e saída. Detalhes, drawers, listas completas e formulários não aparecem nas imagens; entram no backlog de definição/UX antes de implementação.
6. Sincronização precisa position_ts e last_success por fonte; “última atualização” global pode encobrir fontes defasadas. Mudança de mês/filtro não deve produzir cards de uma versão e tabela de outra. Relatórios/exports devem usar snapshot consistente.
7. Paginação, busca, ordenação, total completo versus amostra, quantidade/unidade, casas decimais, moeda e fuso precisam contrato comum. Na mesma posição/filtros, cards/tabelas/gráficos/Excel devem reconciliar; percentuais com denominador declarado e arredondamento controlado.
8. Exportação aprovada em T11/T12 tem conteúdo próprio; T13/14/15 pendentes; T09/T10/T16/T17 sem exportação contratada. Não criar exportação universal só porque há um botão em um mockup.

## Tarefas de definição que alimentam o backlog ponta a ponta

As tarefas futuras devem iniciar em 0 e separar conclusão do planejamento (este levantamento) da conclusão do desenvolvimento (nenhum código implementado nesta análise). Percentual deve vir de entregas aceitas/pesos declarados, não da quantidade de telas lidas.

- Homologar mapa de fontes, versões e escopo canônico; remover duplicata e registrar legado; fechar quais imagens/CEIs complementares faltam.
- Fechar catálogo de métricas/estados/unidades/fontes e exemplos de cálculo por módulo; cada ambiguidade recebe decisão, responsável, impacto e tarefa dependente.
- Elaborar contratos de interação e estados UX para cada botão, detalhe, formulário, tabela e exportação.
- Formalizar T09/T11 regra compartilhada e snapshot; T10 metas e estados; T12 estudo/meta/produção/estoque e métricas; T13 ciclo por tipo; T14 origem/identidade dos alertas; T15 fontes, ingestão e reprocessamento; T16 vigência e fluxos específicos; T17 identidade e matriz de acesso.
- Preparar amostras anonimizadas e casos calculáveis com resultado esperado, incluindo ausências, duplicatas, erro parcial, unidades incompatíveis, igualdade do mínimo, versões concorrentes e acesso negado.
- Reservar implementações de backend, domínio, adaptadores, frontend, QA, operação e produção em tarefas distintas/dependentes após definições; planejamento atual não autoriza escrever código do produto.

## Perguntas prioritárias para o usuário (linguagem simples)

1. Já definido: analisar conflitos caso a caso e incluir todos os módulos do material na primeira produção. Caso concreto a decidir: Balanço Patrimonial aparece nas imagens e está removido nos CEIs; qual destino deseja para esse módulo? Faltam os detalhes/fluxos de Visão Executiva, Compras visual, Aplicações, Logística/Pricing, Simuladores, Relatórios e Produtos(SKU), conforme escopo a mapear.
2. Em Estoque, qual saldo e quais depósitos do Bling devem aparecer? Um pedido parcialmente recebido ou insuficiente ainda conta como “reposição em andamento” e exclui a recomendação?
3. Em Produção, quando uma meta abaixo de100% é “em andamento” ou “abaixo da meta”? Quem define a meta por SKU/mês, pode editar meses anteriores, e qual a relação com as metas de T12?
4. Para recomendar compras: como vocês compram hoje, qual histórico representa a rotina, quais mínimos/múltiplos por insumo e como priorizam prazos? Podemos usar amostras reais para propor uma fórmula e vocês homologarem? Quais fontes/geografias do cacau são confiáveis para vocês?
5. Nas projeções, como vendas previstas viram quantidade a produzir se não descontamos estoque atual? Há regra de arredondamento, lote, perdas ou fabricação antecipada? Qual conteúdo esperado nas abas Insumos e Resumo financeiro?
6. Ao ajustar um SKU na meta mista, mantemos o faturamento global e redistribuímos os demais, ou aceitamos novo total? Qual preço realizado deve ser a referência e quem aprova a qualidade do estudo?
7. Quais pendências podem ser resolvidas, por quem e com que evidência? Quais alertas precisam notificação além da central, se algum? T13/T14/T15 devem exportar o quê e para quais perfis?
8. Quais arquivos reais e fontes externas serão recebidos, com que frequência, volume e pasta/serviço? A origem deve ficar imutável? Quem pode reprocessar e como uma correção de arquivo já importado deve alterar o histórico?
9. Quem pode ver cada módulo, valores financeiros, dados pessoais e logs? Como deseja login/convites/MFA/recuperação? Já definido: somente Majucau, sem multiempresa, hospedagem local.
10. Quais dispositivos precisam funcionar na primeira entrega e quais alterações de conteúdo/acessibilidade/responsividade cabem no desenho aprovado? Quem homologa UX, números financeiros e entrada em produção?
