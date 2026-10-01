## MAJUCAU INTELLIGENCE — FECHAMENTO DAS QUESTÕES AINDA TRATADAS COMO ABERTAS

Esta resposta substitui a interpretação anterior de que todos os itens abaixo ainda dependeriam de decisões da Majucau.

Há pontos que **já possuem regra de negócio definida** e devem apenas ser implementados. Outros pertencem à **responsabilidade técnica do desenvolvimento** e não devem retornar como perguntas de negócio.

A implementação deve respeitar integralmente as regras já consolidadas do projeto.

---

# 1. CALCULADORA DE APLICAÇÕES

## STATUS: FECHADO

Não reabrir este item como pendência da Majucau.

A regra financeira da calculadora já foi definida anteriormente e deve ser implementada conforme a especificação consolidada.

A calculadora não deve ser confundida com cartões, Nuvem Pago ou conciliação de recebíveis.

A arquitetura lógica é:

**Recebíveis previstos**
→ alimentam o fluxo de caixa projetado
→ que alimenta a disponibilidade financeira
→ que alimenta a calculadora de aplicações.

A calculadora recebe o resultado financeiro consolidado e não precisa conhecer detalhes operacionais de cartão, como bandeira, MDR, parcelamento ou taxa de processamento.

A única fonte financeira necessária à projeção que não existe diretamente no Bling é a agenda/saldo de **recebimentos futuros da Nuvem Pago**, cuja origem já está definida em arquivo externo.

Não existe nova decisão de negócio pendente neste item.

---

# 2. DRE E BALANCETE

## STATUS: FECHADO COMO REGRA DE NEGÓCIO

O **Bling já fornece a DRE por competência** e as classificações financeiras existentes no Bling estão corretas.

Portanto:

### DRE

A DRE do Majucau Intelligence deverá utilizar como fonte primária:

**Bling → DRE por competência**

Não deverá ser criada uma segunda estrutura paralela de classificação quando a classificação já estiver corretamente cadastrada no Bling.

Não:

* reinterpretar categoria;
* reclassificar conta por inferência;
* trocar competência;
* transformar ausência em zero;
* criar receitas ou despesas que não existam na origem.

A competência apresentada no Majucau Intelligence deve respeitar a competência fornecida pelo Bling.

### Balancete

O Bling também possui base suficiente para o balancete utilizado pelo projeto.

Portanto o desenvolvimento deve utilizar:

**Bling → dados contábeis/financeiros disponíveis para composição do balancete**

Não há necessidade de solicitar à Majucau uma nova estrutura paralela de balancete.

A obrigação do desenvolvimento é:

1. mapear os recursos disponíveis na API;
2. identificar campos, categorias, contas e saldos correspondentes;
3. reproduzir corretamente a informação existente no Bling;
4. preservar competência e classificação;
5. validar os totais contra o próprio Bling.

### Regra de segurança de dados

Quando algum elemento específico não puder ser obtido pela API do Bling, a aplicação deverá apresentar:

**Dado indisponível no Bling**

É expressamente proibido:

* estimar;
* inferir;
* preencher automaticamente;
* transformar ausência em R$ 0,00;
* completar uma demonstração com valores fictícios.

### Aplicações e resgates

Aplicação do principal e resgate do principal são movimentações financeiras/patrimoniais e não devem ser tratados automaticamente como receita ou despesa na DRE.

Rendimentos devem respeitar a classificação e competência existentes no Bling.

---

# 3. COMPRAS E PREÇOS

## STATUS: REGRA FUNCIONAL DEFINIDA

É necessário separar tecnicamente três assuntos que estavam sendo agrupados em uma única pendência:

1. necessidade/recomendação de compra;
2. formação de custo/preço;
3. inteligência específica do cacau.

São motores relacionados, mas não são a mesma regra.

---

## 3.1 COMPRA RECOMENDADA

O objetivo da tela/motor de compras é responder:

**o que precisa ser comprado, quanto precisa ser comprado e qual é a prioridade.**

A recomendação deverá partir dos dados operacionais reais do Bling.

A lógica deve considerar, no mínimo:

* estoque disponível;
* estoque comprometido quando aplicável;
* demanda projetada;
* necessidade de produção;
* consumo previsto pela ficha técnica/BOM;
* pedidos de compra já abertos;
* quantidade ainda pendente desses pedidos;
* previsão/data de chegada;
* lead time do item.

### Necessidade líquida

Conceitualmente:

**Necessidade bruta**
− **estoque utilizável**
− **quantidade de pedidos abertos que chegará em tempo**
= **necessidade líquida de compra**

Somente a quantidade efetivamente coberta por um pedido aberto poderá reduzir a recomendação.

### Pedido parcialmente suficiente

Exemplo:

Necessidade projetada: 100 kg
Estoque disponível: 20 kg
Pedido aberto com chegada válida: 30 kg

A necessidade remanescente será:

100 − 20 − 30 = **50 kg**

Portanto:

**pedido aberto parcial não elimina a necessidade de compra.**

O sistema deve descontar apenas a parte efetivamente coberta.

### Pedido que chega tarde demais

Pedido aberto cuja previsão de chegada seja posterior ao momento em que o insumo será necessário **não poderá ser considerado como cobertura suficiente daquela necessidade**.

Ele poderá permanecer visível como pedido em trânsito, mas não poderá mascarar risco de ruptura.

---

## 3.2 PRIORIDADE DE COMPRA

A prioridade deve ser derivada da necessidade operacional.

Itens com risco imediato de ruptura ou impacto direto na produção devem aparecer antes de itens sem risco operacional imediato.

O desenvolvimento não deve criar prioridade manual aleatória.

A classificação deve considerar elementos como:

* cobertura de estoque;
* necessidade projetada;
* lead time;
* pedido em aberto;
* data prevista de chegada;
* dependência daquele insumo para produtos com demanda prevista.

A tela deverá permitir distinguir claramente:

**necessidade imediata**,
**necessidade futura**,
**coberta por pedido**,
**parcialmente coberta**,
**não coberta**.

---

# 3.3 PREÇO E TARIFAS

A distribuição de tarifa fixa entre produtos pertence ao cálculo de custo/precificação, não à fórmula de quantidade recomendada.

Portanto:

**não misturar rateio de tarifa com necessidade de compra.**

Sempre que houver custo fixo associado a uma compra, importação, frete, operação ou pedido contendo diversos itens, o cálculo deverá manter separado:

* valor bruto do produto;
* custo variável;
* tarifa/custo fixo;
* critério de rateio;
* custo final apropriado por produto.

O sistema deve preservar rastreabilidade do rateio.

Nenhuma tarifa fixa poderá simplesmente ser multiplicada integralmente por cada produto, pois isso duplicaria o custo.

Caso o próprio Bling forneça a apropriação correspondente, utilizar o dado do Bling.

Não criar valores de rateio fictícios quando a fonte não disponibilizar informação suficiente.

---

# 3.4 CACAU

O módulo de cacau é uma camada adicional de **inteligência de compras**.

Ele não substitui a necessidade operacional calculada pelo estoque e pela produção.

A lógica é:

**necessidade operacional**
→ indica que será necessário comprar;

**inteligência do cacau**
→ ajuda a avaliar momento, risco e contexto da compra.

Devem permanecer separados:

* necessidade de compra;
* histórico de preços;
* preço atual disponível;
* alertas de mercado;
* condições climáticas nas regiões relevantes;
* demais sinais externos utilizados futuramente.

Os fornecedores atuais são pequenos produtores e as regiões relevantes deverão ser relacionadas às localidades efetivamente utilizadas pela Majucau.

Não criar regiões fictícias.

As fontes externas de preço/clima poderão ser refinadas tecnicamente sem impedir o funcionamento do motor principal de compras.

---

# 4. PRODUÇÃO E PROJEÇÕES

## STATUS: REGRA FUNCIONAL DEFINIDA

Também não devem ser tratados como uma única coisa:

* meta;
* realizado;
* forecast;
* necessidade de produção;
* projeção de estoque.

---

# 4.1 FONTE DA PRODUÇÃO

A produção realizada deverá vir do **Bling**.

Cada produto final cadastrado e produzido no Bling deverá alimentar os indicadores correspondentes.

Não criar produção fictícia.

Não preencher produção ausente como zero sem que o Bling efetivamente informe zero.

---

# 4.2 VISÃO EXECUTIVA X TELA DE PRODUÇÃO

Não existem metas independentes entre duas telas.

Existe **uma única informação**, apresentada em níveis diferentes.

### Visão Executiva

Apresenta o resumo:

**Produzido Geral**

### Tela específica de Produção

Apresenta o detalhamento por produto:

* produto;
* quantidade/meta;
* realizado;
* percentual;
* barra de progresso;
* status;
* projeções relacionadas.

Portanto:

**a tela executiva e a tela detalhada devem consumir a mesma fonte e o mesmo cálculo.**

É proibido implementar dois cálculos distintos para o mesmo indicador.

---

# 4.3 CONSISTÊNCIA DOS INDICADORES

Qualquer indicador apresentado em:

* card;
* tabela;
* gráfico;
* exportação;
* resumo executivo;

deve utilizar **a mesma função de cálculo e o mesmo conjunto de dados**.

Exemplo:

Se existe o indicador:

**Produtos com meta atingida**

ele deve ser calculado uma única vez.

O card e a tabela devem refletir exatamente o mesmo universo.

Os números existentes nos mockups são ilustrativos e não são fonte de regra de negócio.

Na aplicação real:

**mesma métrica = mesmo cálculo = mesmo resultado.**

---

# 4.4 STATUS DE PRODUÇÃO

Não criar percentuais ou tolerâncias arbitrárias.

Os estados deverão ser determinados pela relação entre:

* produção realizada;
* meta necessária;
* posição temporal do período/plano.

A lógica conceitual deve distinguir:

### Meta atingida

Produção realizada atingiu ou superou a meta correspondente.

### Em andamento

A produção ainda não atingiu 100%, mas continua compatível com o avanço esperado do período/plano.

### Abaixo da meta

A produção acumulada está inferior ao necessário para o ponto do período/plano em que se encontra.

Não utilizar simplesmente:

`<100% = abaixo da meta`

porque isso marcaria praticamente toda produção em andamento como atraso.

Também não criar uma tolerância percentual sem regra previamente cadastrada.

Quando não existir granularidade temporal suficiente para determinar atraso real, o sistema deve evitar classificar arbitrariamente como “Abaixo da meta”.

---

# 4.5 FORECAST DE VENDAS → PRODUÇÃO

A produção necessária não será obtida dividindo a meta total de faturamento de forma arbitrária entre produtos.

O fluxo correto é:

**Histórico e comportamento de vendas**
→ **forecast por produto/SKU**
→ **demanda prevista**
→ **estoque disponível/projetado**
→ **necessidade de reposição**
→ **necessidade de produção**
→ **necessidade de insumos/compras**

Conceitualmente:

**Produção necessária =
Demanda prevista

* estoque-alvo necessário
  − estoque disponível/projetado
  − produção já disponível ou válida no horizonte**

Sempre respeitando os dados efetivamente fornecidos pelo Bling.

---

# 4.6 META DE FATURAMENTO

Meta de faturamento é planejamento comercial.

Ela pode ser utilizada como referência executiva e cenário.

Não transformar automaticamente uma meta financeira total em unidades por SKU utilizando rateio inventado.

Caso exista posteriormente uma meta definida diretamente por produto, essa informação passa a ser uma entrada de planejamento específica.

Até lá:

**forecast por produto deve ser fundamentado nos dados reais/históricos disponíveis.**

---

# 4.7 COBERTURA HISTÓRICA

“Possuir dados desde 2025” não significa possuir 24 meses completos.

O sistema deverá calcular a cobertura histórica real.

Exemplo:

**Histórico disponível: 19 meses**

ou

**Histórico disponível: 21 meses**

conforme os dados existentes.

Não completar meses inexistentes.

Não tratar ausência de histórico como venda zero.

A qualidade da projeção deve ser testada considerando o histórico efetivamente disponível.

---

# 5. TELAS E NAVEGAÇÃO

## STATUS: REGRAS OBRIGATÓRIAS

Esta frente não deve ser tratada como redesenho de UX.

As telas já foram definidas e revisadas.

Aplicam-se as seguintes regras:

### REGRA 1 — Mockup é a referência visual oficial

O mockup aprovado define:

* estrutura;
* posição dos blocos;
* menu;
* cards;
* cabeçalho;
* gráficos;
* tipografia;
* cores;
* espaçamentos;
* hierarquia visual;
* botões;
* padrões de navegação.

O desenvolvimento não deverá:

* redesenhar;
* modernizar;
* reinterpretar;
* mover bloco;
* trocar card;
* reorganizar menu;
* alterar prioridade visual;

sem nova solicitação expressa.

---

### REGRA 2 — Documentação funcional define comportamento

O mockup define **como aparece**.

A documentação funcional define **como funciona**.

Logo:

**Visual/layout → mockup aprovado**

**Cálculo/regra/fonte/comportamento → documentação funcional consolidada**

---

### REGRA 3 — Em caso de versões diferentes

Utilizar sempre:

**última versão aprovada/congelada.**

Não utilizar versão anterior para sobrescrever uma decisão posterior.

---

### REGRA 4 — Não reabrir telas congeladas

Telas já aprovadas não devem retornar como questão aberta apenas porque a implementação divergiu delas.

Nesse caso:

**corrigir a implementação.**

Não redefinir o requisito.

---

### REGRA 5 — Indicador único

O mesmo indicador não poderá possuir lógicas diferentes em:

* card;
* gráfico;
* tabela;
* detalhe;
* exportação.

Todos deverão consumir a mesma camada de cálculo.

---

### REGRA 6 — Valores dos mockups

Os valores numéricos utilizados nos mockups são demonstrações visuais.

Eles não devem ser utilizados como dados reais ou fonte da regra.

No sistema, todos os valores deverão vir das fontes reais definidas.

---

### REGRA 7 — Exportação

Relatórios/telas que possuem funcionalidade de exportação deverão preservar a opção aprovada de exportação, inclusive para Excel quando especificado.

O arquivo exportado deve refletir os mesmos dados e filtros apresentados na tela.

---

### REGRA 8 — Resumo x detalhamento

Uma informação resumida no Painel Executivo pode possuir tela própria de detalhamento.

Isso não significa criar outro indicador.

Exemplo:

**Produzido Geral**
→ resumo executivo.

**Produção**
→ detalhamento por produto.

A fonte e a regra permanecem únicas.

---

# 6. PERMISSÕES, PARÂMETROS, CORREÇÕES E AUDITORIA

## STATUS: DEFINIR TECNICAMENTE POR CAPACIDADE, NÃO POR NOME DE PESSOA

O sistema não deverá hardcodar permissões diretamente no nome de um diretor.

A autorização deverá ser controlada por:

**usuário + perfil + permissões/capacidades.**

---

# 6.1 PRINCÍPIO DE MENOR PRIVILÉGIO

Cada usuário deve possuir somente as permissões necessárias para sua função.

Separar, no mínimo, capacidades como:

* visualizar;
* exportar;
* alterar metas;
* alterar parâmetros;
* resolver pendências;
* administrar usuários;
* consultar histórico/auditoria.

Permissão de leitura não implica permissão de alteração.

---

# 6.2 USUÁRIOS

Criação, alteração de perfil, ativação/inativação e administração de usuários devem ficar restritas à permissão administrativa correspondente.

Não permitir que qualquer usuário comum altere sua própria autoridade.

---

# 6.3 PARÂMETROS

Parâmetros financeiros e operacionais devem ser alteráveis apenas por usuários que possuam a permissão correspondente.

Toda alteração deve gerar registro contendo pelo menos:

* usuário;
* parâmetro;
* valor anterior;
* valor novo;
* data/hora;
* vigência quando aplicável.

---

# 6.4 METAS

Metas são **Planejamento**.

Não são fatos provenientes do Bling.

Devem ficar identificadas no sistema como entradas de planejamento.

Alterações de metas devem possuir:

* usuário responsável;
* data/hora;
* período;
* valor anterior;
* valor novo.

Não sobrescrever silenciosamente o histórico anterior.

---

# 6.5 CORREÇÕES DE PENDÊNCIAS

Pendências devem possuir ciclo de vida rastreável.

Exemplo conceitual:

**Pendente**
→ **Em análise**
→ **Resolvido**

ou outro fluxo equivalente previsto pela implementação.

A resolução deve guardar:

* quem resolveu;
* quando resolveu;
* qual informação foi alterada;
* origem utilizada;
* justificativa/registro técnico quando necessário.

---

# 6.6 ALTERAÇÕES RETROATIVAS

Alterar um parâmetro hoje não deverá recalcular silenciosamente meses anteriores.

Regra geral:

**alteração passa a valer conforme sua data de vigência.**

Período já fechado não deve mudar automaticamente.

Caso seja necessário reprocessar período anterior, isso deverá ser uma ação:

* explícita;
* autorizada;
* registrada em auditoria;
* rastreável.

Não apagar o valor anterior.

---

# 6.7 PERÍODOS JÁ FECHADOS

O projeto já possui a regra de “passar a régua” nos períodos encerrados definidos.

Esses períodos devem permanecer historicamente preservados.

A aplicação não poderá modificar resultados históricos encerrados apenas porque um parâmetro atual foi alterado.

---

# 6.8 LOG DE AUDITORIA

Para qualquer ação sensível, manter log contendo:

* usuário;
* data/hora;
* ação;
* entidade afetada;
* valor anterior;
* valor posterior;
* origem;
* identificador da operação.

O log não poderá ser editado pelo usuário comum.

---

# 7. OPERAÇÃO, HOSPEDAGEM, BACKUP E SUPORTE

## STATUS: RESPONSABILIDADE TÉCNICA DO DESENVOLVIMENTO

Este item não deve retornar como uma pergunta de regra financeira para a Majucau.

O desenvolvimento deve propor e documentar a arquitetura que atenda aos requisitos do sistema.

A Majucau valida o funcionamento desejado.

O desenvolvimento define tecnicamente como isso será entregue.

---

# 7.1 HOSPEDAGEM

O sistema não deve depender da sessão de um funcionário estar aberta para funcionar.

Se houver hospedagem local, ela deverá operar como serviço da aplicação e não como programa dependente de um usuário manualmente logado.

O desenvolvimento deverá documentar:

* máquina/host;
* sistema operacional;
* serviços executados;
* banco de dados;
* processo de inicialização;
* portas;
* dependências;
* política de atualização.

---

# 7.2 INICIALIZAÇÃO

Após reinicialização do ambiente, os serviços necessários devem conseguir ser iniciados de forma controlada e documentada.

Evitar processos manuais obscuros.

---

# 7.3 ACESSO

O sistema deve utilizar autenticação.

Acesso externo, quando habilitado, não deve expor diretamente:

* banco de dados;
* credenciais;
* tokens;
* APIs internas;
* arquivos de configuração.

Utilizar canal seguro.

---

# 7.4 CREDENCIAIS E SEGREDOS

Nunca armazenar no front-end:

* client secret;
* token Bling;
* credencial de banco;
* senha administrativa;
* credencial de integração.

Credenciais devem permanecer exclusivamente no backend/ambiente seguro.

---

# 7.5 BACKUP

O desenvolvimento deve implementar rotina automática de backup de, no mínimo:

* banco da aplicação;
* configurações necessárias;
* parâmetros;
* usuários/permissões;
* registros de processamento;
* logs necessários à recuperação;
* dados locais que não possam ser reconstruídos pelas fontes externas.

Os arquivos originais existentes no Google Drive não devem ser alterados pela aplicação.

---

# 7.6 RESTAURAÇÃO

Backup sem restauração testada não é suficiente.

Deve existir procedimento documentado contendo:

1. como restaurar banco;
2. como restaurar configuração;
3. como subir novamente a aplicação;
4. como validar integridade após recuperação.

---

# 7.7 LOGS

A aplicação deve registrar erros e processamento das integrações.

Registrar, no mínimo:

* fonte;
* execução;
* horário;
* sucesso/falha;
* arquivo/lote quando aplicável;
* identificadores processados;
* mensagem de erro.

Não expor credenciais no log.

---

# 7.8 SUPORTE

O desenvolvimento deverá entregar documentação mínima suficiente para que seja possível identificar:

* aplicação parada;
* integração falhando;
* arquivo rejeitado;
* API indisponível;
* token vencido;
* banco indisponível;
* processamento duplicado ou bloqueado.

Esse item é de engenharia e operação, não uma nova regra financeira.

---

# 8. DADOS REAIS E INTEGRAÇÕES

## STATUS: FONTES DEFINIDAS

A premissa central é:

# BLING É A FONTE PRINCIPAL DE DADOS OPERACIONAIS E FINANCEIROS.

A única informação financeira necessária ao projeto que não está no Bling é:

**saldo/agenda de recebimentos futuros da Nuvem Pago.**

Essa informação já possui fonte definida.

---

# 8.1 BLING

Devem ser extraídos do Bling, conforme disponibilidade dos recursos reais da API:

* saldos;
* caixa;
* bancos;
* Conta Caixa;
* contas a pagar;
* contas a receber existentes;
* movimentações;
* transferências;
* rendimentos;
* aplicações;
* resgates;
* depósitos;
* vendas;
* pedidos;
* produtos;
* estoque;
* produção;
* compras;
* custos;
* ficha técnica/BOM;
* classificações;
* DRE;
* balancete;
* demais elementos necessários disponíveis.

Não criar fontes paralelas para informação que já existe corretamente no Bling.

---

# 8.2 RECEBIMENTOS FUTUROS DA NUVEM PAGO

Fonte externa já definida:

`02_nuvem_pago/recebimentos_futuros`

O sistema deverá ler os arquivos colocados nessa pasta.

Permissão:

**SOMENTE LEITURA.**

É proibido:

* criar;
* mover;
* renomear;
* excluir;
* sobrescrever;

arquivos da pasta compartilhada.

---

# 8.3 OUTRAS PASTAS EXTERNAS

A estrutura compartilhada já definida possui:

`01_perdas_operacionais`

`02_nuvem_pago/recebimentos_futuros`

`03_nuvem_envio`

Não criar uma pasta adicional de investimentos.

Dados de aplicações/investimentos pertencentes à operação financeira devem ser buscados no Bling.

---

# 8.4 CONTROLE INTERNO DOS ARQUIVOS

Embora a pasta seja somente leitura, o sistema deve controlar internamente:

* arquivo processado;
* hash;
* lote;
* data de processamento;
* status;
* registros importados;
* erros;
* rejeições;
* identificador de origem.

Isso evita que o mesmo arquivo seja processado repetidamente.

---

# 8.5 RECEBÍVEL PREVISTO X RECEBIMENTO REALIZADO

A planilha da Nuvem Pago representa **recebimento futuro/projetado**.

O Bling representa o fato financeiro efetivamente registrado.

Portanto:

**Nuvem Pago futuro**
→ entra como previsão de caixa.

Quando o evento correspondente aparecer como realizado no Bling:

→ o sistema deve reconhecer que o valor deixou de ser apenas previsão;

→ evitar manter simultaneamente o mesmo recebimento como previsto e realizado.

Não duplicar caixa.

---

# 8.6 CHAVES DE CONCILIAÇÃO

Quando existir identificador do Bling disponível na informação da Nuvem, ele deverá ser utilizado prioritariamente.

Na ausência de identificador inequívoco, utilizar os critérios de conferência já definidos:

**identificação/nome + valor**

Não realizar conciliação automática quando houver ambiguidade.

Se houver mais de um candidato possível, marcar para conferência.

É preferível deixar:

**Pendente de conciliação**

do que associar registros incorretamente.

---

# 8.7 NULL NÃO É ZERO

A regra é obrigatória em todas as integrações:

**ausência de informação ≠ zero.**

Não converter automaticamente:

* null;
* vazio;
* inexistente;
* não retornado;

em:

**0**

quando não houver evidência de que o valor seja efetivamente zero.

---

# 8.8 PENDING_RECONCILIATION

Registro com status de conciliação pendente não deve ser tratado como valor conhecido/confirmado.

Deve permanecer fora dos indicadores que exigem valor reconciliado até que a conciliação seja concluída.

---

# 8.9 DEPÓSITO OPERACIONAL

O depósito operacional deve ser identificado diretamente no cadastro/dados do Bling via API.

Não hardcodar:

* código;
* nome;
* depósito “Bloqueado”;
* qualquer depósito presumido.

Se não houver identificação inequívoca:

**Dado indisponível no Bling.**

---

# 8.10 CONTA CAIXA

O caminho funcional já informado é:

**Bling → Caixa e Bancos → Conta Caixa**

O desenvolvimento deverá localizar tecnicamente na API os recursos correspondentes.

Não devolver à Majucau como nova pergunta de negócio algo que já está definido funcionalmente no sistema de origem.

---

# 8.11 ROTINA DE CONCILIAÇÃO

CORRIGIR a documentação que menciona **D+1**.

A regra da Majucau é:

# D-1

Na análise/conciliação normal, considerar movimentações somente até o dia anterior.

O movimento do próprio dia não entra automaticamente.

Somente incluir movimento do dia corrente quando houver solicitação expressa.

A situação específica de 04/09/2026 ter sido conciliada posteriormente ocorreu devido ao feriado da segunda-feira seguinte e não deve ser transformada em regra fixa para todas as sextas-feiras.

---

# 8.12 IDEMPOTÊNCIA

Toda integração deverá ser idempotente.

Processar novamente o mesmo arquivo/evento não poderá duplicar:

* recebimento;
* perda;
* frete;
* lançamento;
* registro;
* KPI.

Utilizar identificadores de origem, hash, chave de negócio e controle de processamento conforme aplicável.

---

# 8.13 RASTREABILIDADE

Todo dado externo incorporado ao sistema deve permitir identificar sua origem.

Quando aplicável, manter:

* sistema de origem;
* ID de origem;
* arquivo;
* lote;
* data de importação;
* data de referência/competência;
* status de conciliação.

---

# 8.14 REGRA DE PRECEDÊNCIA DAS FONTES

Para fatos operacionais/financeiros:

**Bling prevalece como fonte principal.**

Exceções já expressamente definidas:

* perdas operacionais → arquivo externo correspondente;
* recebimentos futuros Nuvem Pago → arquivo externo correspondente;
* dados específicos Nuvem Envio → arquivo externo correspondente.

Essas fontes externas não devem substituir fatos existentes no Bling fora do escopo para o qual foram criadas.

---

# CONCLUSÃO PARA O DESENVOLVIMENTO

A lista anterior superestimou a quantidade de decisões ainda abertas.

A situação correta é:

**1. Calculadora:** regra de negócio fechada.

**2. DRE e Balancete:** utilizar Bling, por competência e classificações existentes.

**3. Compras:** implementar necessidade líquida, cobertura parcial de pedidos e priorização operacional; manter custo/rateio separado do motor de quantidade.

**4. Produção:** Bling como realizado; forecast de vendas → estoque → produção → compras; uma única métrica compartilhada entre resumo e detalhe.

**5. Telas:** seguir os mockups aprovados como regra visual, sem redesenhar ou reabrir telas congeladas.

**6. Permissões:** implementar RBAC/capacidades, vigência, trilha de auditoria e proibir alteração retroativa silenciosa.

**7. Operação:** desenvolvimento deve definir e documentar arquitetura, segurança, backup, restauração, logs e suporte. Não é pendência financeira da Majucau.

**8. Integrações:** Bling é a fonte principal. A única informação financeira necessária fora do Bling é a agenda de recebimentos futuros da Nuvem Pago, já fornecida pela pasta compartilhada. Perdas e Nuvem Envio utilizam suas respectivas pastas conforme escopo específico.

Portanto, não devolver essas definições novamente como perguntas abertas à Majucau. A próxima etapa é **implementação, mapeamento técnico e validação contra os dados reais das fontes já definidas**.

