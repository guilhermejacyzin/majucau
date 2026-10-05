# Registro de decisão G0 — Majucau Financial Intelligence

- **Status:** aprovado pela responsável pelo produto
- **Data da proposta:** 2026-08-15
- **Efeito da aprovação:** autoriza a fundação técnica e somente os módulos não bloqueados; não equivale a declarar o aplicativo pronto para produção
- **Regra:** ausência de fonte oficial nunca será substituída por estimativa silenciosa

## Como aprovar

A aprovação deve registrar uma opção para D-001 a D-005. A resposta curta recomendada é:

> Aprovo o G0 com as recomendações D-001-A, D-002-A, D-003-A, D-004-A e D-005-A.

## D-001 — Classificação de recebíveis B2B no Bling

### Problema

A API pública do Bling não fornece um campo `business_type=B2B`. Sem regra determinística, o total futuro B2B não é reproduzível.

### D-001-A — Regra conservadora versionada (recomendada)

Aplicar em ordem:

1. override manual auditado por contato/título;
2. título reconciliado com pedido Nuvemshop/Nuvem Pago: `B2C_EXCLUDED_FROM_BLING_FUTURE`;
3. contato com CNPJ válido: `B2B`;
4. canal, categoria ou contato incluído em lista B2B aprovada: `B2B`;
5. demais: `UNCLASSIFIED`, fora do total confirmado e presentes em fila de revisão.

A reconciliação do item 2 só é automática quando houver identificador direto de origem. Sem ID direto, exigir igualdade exata de loja/conexão, número externo do pedido, hash do documento do cliente, moeda, valor bruto e data de emissão. Ausência, divergência ou mais de um candidato resulta em `UNCLASSIFIED`; não haverá matching probabilístico. Override manual exige motivo e auditoria.

**Consequência:** minimiza dupla contagem e falso B2B, mas exige revisão dos casos sem classificação.

### D-001-B — Classificar todo Bling como B2B

**Não recomendada.** É simples, porém pode incluir B2C do Bling e violar a regra de composição do handoff.

### Evidência para fechar

- aprovação da opção;
- fixtures cobrindo CNPJ, CPF, origem Nuvem, documento ausente, override, parcial e cancelamento;
- reconciliação com tolerância de R$ 0,01.

## D-002 — OAuth Nuvemshop no desktop Windows

### Problema

O callback loopback de aplicativo desktop não está comprovado na documentação pública consultada. Embutir client secret ou copiar tokens manualmente é inseguro.

### D-002-A — Relay HTTPS mínimo de autorização (recomendada)

- desktop cria `state`, sessão e segredo de pareamento aleatórios;
- navegador usa o redirect HTTPS registrado;
- relay recebe somente `code/state`;
- desktop lê o código uma única vez dentro de cinco minutos;
- worker local troca o código por token usando o secret protegido por DPAPI;
- relay não recebe client secret, access token nem dados financeiros;
- replay, expiração, cancelamento e exclusão de sessão são testados;
- PKCE será usado se o endpoint oficial aceitar. Se não aceitar, D-002-A permanece bloqueada até existir threat model específico e uma decisão separada `D-002-RISK` aceitando que o fluxo não é equivalente a um cliente nativo protegido por PKCE.

**Consequência:** conexão simples para o usuário, mas cria um componente HTTPS mínimo que precisa ser publicado, monitorado e registrado no provedor.

### D-002-B — Adiar Nuvemshop até callback loopback oficial

**Segura, porém bloqueia B2C e a experiência de conexão.**

### Evidência para fechar

- aplicativo real cadastrado;
- redirect exato aceito;
- confirmação/homologação do provedor;
- teste ponta a ponta e teste adversarial de replay.
- confirmação de PKCE; se ausente, threat model e aceite explícito `D-002-RISK` com mitigação, prazo e responsável.

## D-003 — Fórmula do Valor Máximo para Aplicação

### Contexto original e atualização

Na aprovação do G0, em 2026-08-16, a planilha referenciada pelo handoff não havia sido fornecida. Por isso, D-003-A foi aprovada como fórmula própria do Majucau, sem alegar equivalência com a referência.

Em 2026-10-04, Gisele forneceu `Calculadora_Aplicacao_Lucro_Liquido_Profissional.xlsm` e pediu que a parte de segurança considerasse o fluxo de caixa. A planilha foi recebida e analisada; o mapeamento inicial e a dúvida ainda aberta estão em `docs/evidence/APPLICATION-RULE-D003A-2026-09-20.md`. D-003-A continua vigente até a comparação ser fechada e qualquer alteração de regra ser aprovada.

### D-003-A — Aprovar fórmula Majucau v1 como novo oráculo (recomendada)

```text
eligible_accumulated_profit = lucro líquido acumulado elegível até M-2
available_profit = max(0, eligible_accumulated_profit - already_invested)
worst_week_cash = menor closing_balance diário encontrado em D0..D+60
financial_limit = max(0, worst_week_cash - minimum_reserve)
maximum_investment = max(0, min(available_profit, financial_limit))
```

Somente recebimentos com fonte oficial e data válida participam. O cálculo usa BRL, decimal exato e `ROUND_HALF_UP` em duas casas apenas na fronteira de apresentação/reconciliação.

**Consequência:** criou uma regra auditável, sem alegar equivalência com uma planilha ainda não conferida.

### D-003-B — Aguardar a planilha original

**Consequência:** era a opção antes da entrega da planilha. A referência foi recebida em 2026-10-04; a comparação está pendente e não suspende D-003-A.

### Evidência para fechar

- comparação com a planilha recebida e confirmação do tratamento dos recebimentos previstos da Nuvemshop e da reserva mínima;
- casos dourados incluindo limite negativo, lucro insuficiente, pior saldo, reserva e recebimento sem fonte oficial;
- diferença máxima de R$ 0,01.

## D-004 — Fonte financeira do Nuvem Pago

### Problema

A documentação pública localizada não comprova API de ledger do lojista com taxa efetiva, líquido, agenda e data prevista de repasse. Pedidos Nuvemshop não provam esses campos.

### D-004-A — Diferir B2C confirmado e manter adapter oficial (recomendada)

- construir contrato do adapter e painel de integração;
- importar pedidos Nuvemshop apenas como operacionais/provisórios;
- mostrar o módulo financeiro Nuvem Pago como `UNAVAILABLE` ou `PARTIALLY_AVAILABLE`;
- não calcular silenciosamente taxa, líquido ou data prevista;
- aceitar futuramente API privada oficial ou arquivo exportado oficialmente, após contract tests;
- manter G2/G3/G7 bloqueados para qualquer release que prometa B2C confirmado.

**Consequência:** permite avançar fundação, Bling, UX, segurança e operação sem fabricar verdade financeira. O objetivo completo continua aberto.

### D-004-B — Retirar B2C confirmado do release inicial

**Consequência:** permite um release de escopo reduzido, mas altera explicitamente o handoff e não conclui o objetivo integral.

### Evidência para fechar o objetivo integral

- documentação/contrato oficial ou exportação oficial real;
- payload/arquivo sanitizado;
- bruto, taxa, líquido, parcela e data esperada;
- reconciliação com tolerância de R$ 0,01.

## D-005 — Defaults financeiros e operacionais

### D-005-A — Aprovar o pacote de defaults (recomendada)

- moeda única do MVP: BRL;
- timezone de negócio: `America/Sao_Paulo`; instantes técnicos em UTC;
- dinheiro: `numeric(19,4)`/decimal exato, nunca `float`;
- arredondamento: `ROUND_HALF_UP`, duas casas somente na fronteira da regra;
- relatórios de data de negócio: início e fim inclusivos;
- aging mutuamente exclusivo: vencido, hoje, 1–7, 8–15, 16–30, 31–45 e 46–60 dias; usar somente saldo aberto, excluindo cancelados/estornados;
- intervalos técnicos: `[início, fim)`;
- sincronização padrão: 15 minutos;
- `STALE`: duas execuções perdidas ou 30 minutos sem sucesso;
- pagamento parcial preserva saldo/alocação;
- estorno, chargeback e cancelamento geram evento compensatório, sem apagar o fato anterior;
- folha: CSV UTF-8 delimitado por `;`, conforme o dicionário de dados;
- RAW append-only para operação, com exceção exclusiva de redação LGPD autorizada e auditável.

### D-005-B — Fornecer defaults alternativos

Qualquer mudança deve informar moeda, timezone, escala, arredondamento, frequência, limiar de stale, política de eventos compensatórios, folha e tratamento LGPD.

## Matriz de autorização após a decisão

| Decisão | Fundação | Bling | Nuvemshop operacional | B2C financeiro | Aplicação | Produção integral |
|---|---:|---:|---:|---:|---:|---:|
| D-001-A | libera | libera após fixtures | não afeta | reduz dupla contagem | não afeta | necessária |
| D-002-A | libera contrato | não afeta | condicionada ao provedor | condicionada | não afeta | necessária |
| D-003-A | libera contrato | não afeta | não afeta | não afeta | libera após casos dourados | necessária |
| D-004-A | libera | libera | libera provisório | mantém bloqueado | exclui B2C sem fonte | integral continua bloqueada |
| D-005-A | libera | libera implementação | libera implementação | não supre fonte | libera regra técnica | necessária |

## Registro de aprovação

| Campo | Valor |
|---|---|
| Responsável | Gisele |
| Data/hora | 2026-08-16, America/Sao_Paulo |
| D-001 | D-001-A — regra conservadora versionada |
| D-002 | D-002-A — relay HTTPS mínimo, condicionado à homologação e ao tratamento de PKCE |
| D-003 | D-003-A — fórmula Majucau v1 como novo oráculo |
| D-004 | D-004-A — B2C confirmado diferido; adapter oficial mantido |
| D-005 | D-005-A — defaults financeiros e operacionais aprovados |
| Observações | Aprovação explícita recebida no task Codex. Libera G1 e módulos não bloqueados; não libera B2C confirmado nem produção integral sem as evidências externas. |

## Decisão posterior aprovada — importação parcial atômica

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-02 (America/Sao_Paulo).
- Linhas que falham na validação são contadas em `records_failed`; as linhas válidas do mesmo arquivo são persistidas e o lote termina como `PARTIAL`.
- A persistência das linhas válidas e a finalização da auditoria do lote pertencem à mesma transação PostgreSQL. Falha técnica ou cancelamento durante a gravação reverte todas as escritas dessa transação.
- Esta decisão trata importação de arquivos com erros de validação por linha. Ela não define a contagem de páginas RAW da API Bling nem o filtro/cursor incremental dos endpoints, que continuam pendentes de confirmação.

## Decisão técnica posterior aprovada — compatibilidade e formato dos backups

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-03 (America/Sao_Paulo).
- Preservar leitura/restauração dos pacotes V1 existentes e gravar novos pacotes como V2 em blocos autenticados.
- A implementação mantém V1 no AES-GCM legado e usa Streaming AEAD AES-256-GCM-HKDF do Tink no V2, com segmentos de 1 MiB. Temporários de plaintext exigem DACL privada e nenhuma restauração pode usar plaintext antes da autenticação completa.
- Esta decisão aprova formato/compatibilidade criptográfica; não altera telas aprovadas, regras financeiras ou regras de negócio.
- A prova dedicada de compatibilidade V1 passou na CI Windows `37103530380`: ela restaura fixture no formato legado em blocos maiores que 64 KiB e rejeita tag adulterada antes de parar o worker. A validação de restauração em VM Windows 10/11 limpa permanece gate operacional.

## Decisão operacional validada — integridade do Evergreen WebView2

- A CI 37100892358 recebeu pela URL oficial um bootstrapper com SHA-256 diferente do valor conhecido fixado no repositório. O build foi interrompido e rejeitou o arquivo; o hash novo, isoladamente, não foi tratado como prova de autenticidade.
- Como o Bootstrapper Evergreen é distribuído por link programático e pode ser atualizado pela Microsoft ([documentação oficial](https://learn.microsoft.com/microsoft-edge/webview2/concepts/distribution?form=MT00J1)), a preparação mantém o hash conhecido e exige, para qualquer download, Authenticode `Valid`, nome simples do signatário exatamente `Microsoft Corporation` e EKU de assinatura de código `1.3.6.1.5.5.7.3.3`.
- Na CI 37101826025, o novo hash `AA38A8CFCE6179B87181609B1C730A29EAF26138FC833AF5759E67576770F3A3` foi aceito após a assinatura passar por esses controles; thumbprint do certificado: `4028CAD637509D4744B17EC5B42AED8D7A31E6AF`. O instalador NSIS e o smoke de instalar/desinstalar passaram.
- Divergência de hash só prossegue após essa validação de assinatura; assinaturas ausentes, inválidas ou de outro signatário removem o temporário e bloqueiam o build.
- Este controle afeta apenas a cadeia de build/empacotamento; não altera telas aprovadas nem regras de negócio.

## Decisão posterior aprovada — cards EBIT e EBITDA na página T6

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo).
- Na faixa inferior da página T6, substituir visualmente o card `UNMAPPED` por dois cards: EBIT/LAJIR e EBITDA/LAJIDA. O alerta de lançamentos `UNMAPPED` continua visível na área de alertas e as pendências continuam existindo.
- Preservar todas as linhas, classificações, critérios de reconhecimento e totais já aprovados para a DRE; os cards não criam lançamentos nem alteram o resultado líquido.
- Cálculo dos indicadores com componentes contabilizados e conciliados da própria DRE:
  - EBIT/LAJIR = lucro líquido + tributos sobre o lucro + despesas financeiras − receitas financeiras.
  - EBITDA/LAJIDA = EBIT + depreciação + amortização + exaustão.
- Referência metodológica: [Resolução CVM 156](https://conteudo.cvm.gov.br/legislacao/resolucoes/resol156.html). Quando faltar qualquer componente ou reconciliação, mostrar o traço (—) e “Sem dados confirmados”; não usar zero nem valores ilustrativos.
- A decisão aprova a apresentação em cards e a metodologia derivada; não aprova EBITDA ajustado nem mudanças nas regras contábeis da DRE.

### Complemento aprovado — regra de cálculo dos cards EBIT/EBITDA

- **Aprovado por:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo).
- Seguir o exemplo fornecido pela responsável: **EBITDA é o resultado antes de descontar a depreciação; EBIT é o EBITDA depois de descontar a depreciação do período**.
- A regra de cálculo e a explicação simples devem refletir essa relação. Os cards são apenas para consulta; a tabela e as linhas aprovadas da DRE permanecem inalteradas.
- Esta confirmação aprova a relação entre os indicadores, mas não define o mapeamento de todas as contas ou a origem/competência da depreciação. Os valores reais dependem desse mapeamento e da conciliação; até lá, os cards mostram `—` e “Sem dados confirmados”. Não presumir outros componentes ou ajustes.
- Esta orientação posterior substitui as fórmulas detalhadas registradas acima para os cards. Amortização, exaustão e ajustes não serão somados aos indicadores sem validação e aprovação específicas.

## Decisão posterior aprovada — consulta dos relatórios

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo).
- Na versão inicial, os relatórios serão consultados somente nas telas do Majucau.
- PDF e Excel ficam fora do escopo até nova aprovação.
- O conteúdo detalhado de cada consulta continua pendente de validação.

## Diretriz aprovada — ordem das consultas pelo painel executivo

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo).
- Organizar o detalhamento das consultas na mesma ordem visual da tela aprovada “Visão Executiva”: cards do topo, área central, cards inferiores e alertas.
- Essa ordem serve para planejar e detalhar as consultas existentes. Não autoriza criar telas novas nem mudar fórmulas, cálculos ou regras de negócio.
- A DRE continua em sua tela própria como resultado consolidado conforme as regras aprovadas; os cards EBIT/EBITDA permanecem no lugar de UNMAPPED, com o alerta de pendências visível.

## Decisão posterior aprovada — DRE própria do Majucau

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo).
- O Majucau calculará sua própria DRE usando regras contábeis aprovadas; não copiará a DRE pronta do Bling como resultado final.
- Dados de origem do Bling poderão alimentar o cálculo e servir para comparação, preservando origem, período e diferenças explicáveis.
- A tabela e as linhas da DRE já aprovadas permanecem como estão, inclusive os cards EBIT/EBITDA no lugar do card inferior UNMAPPED. Os cards CAPEX aprovados abaixo são informativos e ficam fora da tabela/lucro.
- **Pendente:** concluir a validação contábil da política por conta/operação, escolher critérios operacionais (por exemplo, valoração do estoque e rateio de custos), mapear cada lançamento às linhas existentes e confirmar os dados de origem. Os princípios recebidos em 2026-10-04 estão em `docs/DRE-ACCOUNTING-POLICY-v1.md`. Até as dependências serem resolvidas, os valores afetados ficam indisponíveis; nenhuma classificação ou fórmula será inventada.

## Complemento de regra contábil — estoque, produção e CPV/CMV

- **Recebido de:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo). Referência detalhada: `docs/DRE-ACCOUNTING-POLICY-v1.md`.
- Princípio de produto aceito: compra de matéria-prima e embalagem não vira automaticamente despesa nem CPV/CMV. O custo permanece no estoque enquanto estiver em matéria-prima, produto em processo ou produto acabado; a parcela correspondente ao produto efetivamente vendido é reconhecida na DRE.
- Custos de transformação podem incluir matéria-prima, mão de obra direta e custos indiretos de produção apropriáveis, com alocação sistemática e capacidade normal. Desperdício anormal, ociosidade não alocável e gastos que não contribuam para trazer o estoque à condição/localização atuais não podem ser capitalizados para aumentar artificialmente estoque e lucro.
- Aquisição de máquina/equipamento deve passar pela avaliação de reconhecimento como imobilizado. Depreciação de equipamento de produção pode integrar o custo do estoque, conforme o uso e a alocação aplicável; equipamento administrativo segue sua natureza administrativa. Compra de ativo não é despesa automática do mês.
- Tributos recuperáveis só ficam fora do custo quando a recuperabilidade estiver comprovada para a empresa, operação, tributo, localidade e período. O sistema não presume crédito tributário.
- Permanecem obrigatórios: competência, distinção entre resultado e caixa, documento/fato econômico de origem e rastreabilidade até a fonte. Recebimentos, pagamentos e saldo bancário, isoladamente, não determinam receita, despesa ou lucro.
- Este complemento preserva as linhas e o desenho já aprovados da DRE e os cards EBIT/EBITDA aprovados. Ele não aprova uma nova tela, uma nova linha ou cálculo de CPV sem os dados necessários.
- **Ainda pendente de validação contábil específica:** método de valoração do estoque e consistência por natureza/uso; capacidade normal e critério de rateio dos custos indiretos; regras para perdas e armazenamento; regime tributário e créditos por operação; critérios e cadastros do imobilizado; plano de contas e mapeamento de contas/eventos às linhas da DRE.
- **Dependência de dados:** posição e movimentação de matérias-primas, produtos em processo e acabados; consumo e produção por produto; vendas/devoluções; mão de obra e custos fabris; notas de entrada e tributos; cadastro/depreciação do imobilizado; plano de contas e balancete. Se uma fonte não fornecer o dado necessário, o componente afetado permanece indisponível, sem estimativa ou substituição por caixa.
- Base normativa consultada: [CPC 16 (R1) — Estoques](https://www.cpc.org.br/Arquivos/Documentos/243_CPC_16_R1_rev%2012.pdf), [CPC 27 — Ativo Imobilizado](https://www.cpc.org.br/Arquivos/Documentos/316_CPC_27_rev%2014.pdf) e [CFC — Tributos Recuperáveis](https://cfc.org.br/tecnica/perguntas-frequentes/4066-2/). A aplicação tributária depende dos fatos concretos da empresa e deve ser validada com a contabilidade.

## Decisão posterior aprovada — dois cards informativos de CAPEX na T6

- **Aprovada por:** Gisele, nesta conversa Codex, em 2026-10-04 (America/Sao_Paulo).
- Incluir na página T6 dois cards separados da tabela e do lucro: `CAPEX PAGO NO MÊS` e `CAPEX ACUMULADO`.
- `CAPEX PAGO NO MÊS` mostra pagamentos no mês/período exibido que estejam vinculados a uma aquisição de ativo classificada como CAPEX. É uma saída de caixa informativa; pagamento genérico a fornecedor não basta para classificá-lo.
- `CAPEX ACUMULADO` representa investimentos em ativos reconhecidos como CAPEX desde o início do histórico confiável disponível. A usuária confirmou o acumulado desde o início, não somente o período selecionado.
- Nenhum dos dois cards entra no lucro da DRE. A depreciação segue sua regra contábil própria e não é substituída pelo CAPEX.
- Reduzir as dimensões da faixa inferior para acomodar os cards, mantendo linhas da DRE e alerta `UNMAPPED` como aprovados.
- A usuária confirmou que os relatórios existentes informam CAPEX. O painel ainda não recebe esses valores: o snapshot atual consulta recebíveis, recebimentos, contas a pagar e pagamentos, sem campos de CAPEX. A pendência é integrar ao snapshot os dados já informados nos relatórios e conferir o período; até a integração, mostrar `—`. Não reconstruir o CAPEX com pagamentos genéricos nem estimar por extratos bancários.
