# Decomposição funcional máxima — Majucau Financial Intelligence v1

## 0. Propósito e autoridade

Este documento destrincha o produto em módulos, submódulos, funções, regras, dados, estados, permissões, integrações, lineage e critérios de aceite. Ele foi produzido a partir do handoff técnico-funcional Markdown de 2026-08-16 e da referência visual da Visão Executiva.

Ele não transforma a imagem em fonte de números e não aprova novas regras financeiras. A imagem define composição, hierarquia e intenção de UX. O handoff define escopo e regras. `AGENTS.md`, decisões G0 e a fonte oficial prevalecem sobre qualquer inferência visual.

### Legenda de estado

- `CONFIRMED`: fonte oficial, regra e reconciliação aprovadas;
- `PARTIALLY_CONFIRMED`: parte dos componentes confirmada;
- `PROVISIONAL`: pode mudar por falta de conciliação ou fechamento;
- `PROJECTED`: estimativa futura calculada pelo motor;
- `PENDING_RECONCILIATION`: importado, mas ainda não comparado;
- `STALE`: último dado válido excedeu a política de atualização;
- `DIVERGENT`: diferença não zero ou regra inconsistente;
- `UNAVAILABLE`: fonte, autorização ou contrato não disponível;
- `NOT_CONFIGURED`: módulo ou integração ainda sem configuração.

### Regra de completude

Uma função só será marcada como pronta quando tiver código, contrato, fonte, regra, teste positivo, teste negativo, estado de falha, autorização, log sanitizado, lineage, documentação operacional e critério de aceite executado. A presença de uma rota ou card não satisfaz essa definição.

## 1. Árvore completa do produto

| Módulo | Nome oficial | Conteúdo principal | Gate predominante |
|---|---|---|---|
| M00 | Inicialização e contexto | first-run, empresa, data-base, cenário, workspace, saúde local | G1/G4 |
| M01 | Visão Executiva | 20 indicadores, gráficos, alertas, atualizações e drill-down | G4 |
| M02 | Tesouraria e Projeção Financeira | saldo D-1/D0/D+1–D+60, entradas, saídas, reserva, cenários | G3 |
| M03 | Recebíveis | B2C Nuvem, B2B Bling, total, aging e revisão | G2/G3 |
| M04 | Recebimentos Realizados | recebimentos pagos/recebidos do Bling e reconciliação | G2/G3 |
| M05 | Obrigações a Pagar | contas abertas, vencimentos, saldos, aging e filtros | G2/G3 |
| M06 | Pagamentos Realizados | pagamentos do Bling, datas, valores e reconciliação | G2/G3 |
| M07 | Resultado Econômico | DRE, P&L, EBITDA e margens | G5 |
| M08 | Planejamento e Forecast | orçamento, realizado, variações, versões e premissas | G5 |
| M09 | Capital de Giro | ativos circulantes, passivos circulantes e necessidade | G5 |
| M10 | Tesouraria de Aplicações | reserva, lucro elegível, limite e aplicação máxima | G5/G7 |
| M11 | Contabilidade Gerencial e Patrimonial | plano de contas, partidas, balancete, balanço e divergência | G5/G7 |
| M12 | Conciliações e Divergências | comparação de fontes, diferenças, fila e resolução | G3/G5/G6 |
| M13 | Cadastros e Configurações | integrações, credenciais, contas, categorias, parâmetros e usuários | G2/G6/G7 |
| M14 | Folha e Encargos | importação, departamentos, CLT, pró-labore, encargos e provisões | G5 |
| M15 | Alertas e Centro de Atenção | liquidez, vencidos, sync, EBITDA, contabilidade e dados stale | G3/G4/G5 |
| M16 | Sincronização e Dados RAW | conectores, lotes, paginação, idempotência, normalização e checkpoints | G2 |
| M17 | Auditoria e Lineage | origem, regra, snapshot, RAW, fonte e histórico de ações | G2/G4/G5/G6 |
| M18 | Segurança, Acesso e LGPD | RBAC, SID, DPAPI, logs, redação e autorização | G2/G6/G7 |
| M19 | Operação Desktop | worker, PostgreSQL, instalação, update, backup, restore e suporte | G1/G6/G7 |

## 2. Requisitos transversais a todos os módulos

### T01 — Contexto financeiro obrigatório

Todo valor deve declarar moeda, data-base, intervalo, cenário, status, origem, horário de cálculo e versão da regra. O frontend não recebe apenas um `number`.

### T02 — Separação semântica

O domínio separa:

- realizado versus projetado;
- caixa versus competência;
- fonte operacional versus cálculo Majucau;
- dado confirmado versus dado provisório;
- tentativa de sincronização versus sincronização concluída;
- documento original versus snapshot derivado.

### T03 — Lineage

O caminho mínimo é:

```text
componente visual
  -> DTO/snapshot
  -> regra e versão
  -> contribuição tipada
  -> entidade tratada
  -> raw_record versionado
  -> source_system/source_entity/source_id
  -> conexão/provedor
```

### T04 — Falha segura

Falha de API, banco, worker, parser, regra ou permissão não pode apagar, zerar, duplicar ou substituir silenciosamente o último dado válido. O módulo deve preservar o estado anterior e exibir `STALE`, `PARTIALLY_CONFIRMED`, `DIVERGENT` ou `UNAVAILABLE`.

### T05 — Ações

Qualquer ação material passa pelo worker, verifica perfil/SID, registra auditoria e retorna código estável. O botão desabilitado no React não é controle de segurança.

### T06 — Acessibilidade

Todo filtro, card clicável, gráfico, tabela, alerta, campo, botão e status possui label visível, tooltip simples, foco visível, `aria-describedby` quando necessário e alternativa textual. Gráficos têm tabela/resumo acessível.

## 3. M00 — Inicialização e contexto de uso

### Objetivo

Garantir que a aplicação saiba para qual empresa, período, cenário e instalação está apresentando informações, sem permitir que contexto ausente pareça dado válido.

| ID | Função | Entradas | Saídas/estado | Requisitos do handoff |
|---|---|---|---|---|
| M00-F01 | First-run | instalação, consentimento, diretórios, perfil Windows | instalação concluída ou bloqueada | §§1, 39–43, 46 |
| M00-F02 | Identificar empresa/workspace | cadastro local e conta externa | empresa ativa, ID interno, nome legível | §§11, 15, 22 |
| M00-F03 | Selecionar data-base | data de negócio | snapshot com `as_of` e timezone | §§7–9, 43 |
| M00-F04 | Selecionar cenário | BASE/CONSERVATIVE/STRESS/OPTIMISTIC | cenário aplicado sem alterar realizado | §§26, 43 |
| M00-F05 | Saúde do worker | IPC health | OK/DEGRADED/UNAVAILABLE e dependências | §§16, 21, 39, 46 |
| M00-F06 | Estado de atualização | sync logs por provedor | última tentativa, último sucesso, stale | §§20–21, 43 |
| M00-F07 | Filtros globais | período, empresa, cenário | `GlobalFilterState` propagado aos módulos | §§11, 35, 43 |
| M00-F08 | Orientação contextual | foco/hover/teclado | tooltips em linguagem simples; sem botão global `Ajuda` | regra visual vigente, confirmada em 20/09/2026 |

### Critérios de aceite

- iniciar sem integração não mostra números de produção;
- uma data-base muda todas as consultas compatíveis, não apenas o título;
- cenário não altera registros realizados;
- último sucesso e última tentativa são distintos;
- worker indisponível não gera zeros nem falsa conexão;
- restart preserva empresa, filtros seguros e estado de integração sem expor secrets.

## 4. M01 — Visão Executiva

### Objetivo

Permitir uma leitura rápida da posição financeira, tendência, risco, origem e próxima ação. É um agregador; não é fonte contábil nem lugar para esconder divergências.

### Composição visual realinhada

1. cabeçalho: Visão Executiva, empresa, data-base, última atualização, ajuda;
2. filtros globais: período, empresa, cenário;
3. faixa de tesouraria: sete cards;
4. faixa de recebíveis/obrigações: cinco cards;
5. faixa de recebimentos e pagamentos realizados;
6. gráficos de saldo e fluxo;
7. DRE/P&L e EBITDA;
8. alertas e últimas atualizações.

| ID | Função | Dados | Estado mínimo | Drill-down |
|---|---|---|---|---|
| M01-F01 | Renderizar 20 indicadores | snapshots executivos | vazio/loading/error/confirmed | módulo correspondente |
| M01-F02 | Mostrar status e origem | `MetricValue` | status textual + cor + ícone | regra/fonte |
| M01-F03 | Mostrar comparação | realizado/orçado/forecast | favorável/desfavorável por natureza | planejamento |
| M01-F04 | Evolução de saldo | 61 `daily_balances` | série ou sem série confirmada | tesouraria |
| M01-F05 | Fluxo projetado | entradas, saídas, saldo | série com legenda/tabela | projeção |
| M01-F06 | Resumo DRE/P&L | snapshots versionados | período e reconciliação | resultado |
| M01-F07 | Resumo EBITDA | realizado/orçado/forecast | margem e variação | resultado |
| M01-F08 | Lista de alertas | alertas abertos | severidade, causa e valor | módulo/registro |
| M01-F09 | Últimas atualizações | sync logs e cálculos | lote, fonte, horário, status | log sanitizado |
| M01-F10 | Drill-down universal | `LineagePath` | caminho até RAW ou causa da indisponibilidade | origem |

### Cards oficiais

| Grupo | Cards | Regra principal |
|---|---|---|
| Tesouraria | Saldo Inicial do Dia; Saldo Final Projetado Hoje; Saldo Projetado em 30 Dias; Saldo Projetado em 60 Dias; Menor Saldo Projetado — 60 Dias; Reserva Mínima; Valor Máximo para Aplicação | D-1, D0–D60, reserva e D-003-A |
| Recebíveis | A Receber B2C — Nuvem; A Receber B2B — Bling; Total a Receber; Recebido no Mês | não duplicidade e fonte oficial |
| Obrigações | A Pagar; Obrigações Vencidas; Pago no Mês | saldo aberto, aging e realizado |
| Resultado | DRE / P&L; EBITDA | motor Majucau e snapshots |
| Planejamento | Orçado x Realizado; Forecast | versão e favorabilidade |
| Gestão | Capital de Giro; Alertas | plano de contas, thresholds e riscos |

### Aceite

Nenhum valor da imagem de referência pode aparecer sem fixture explicitamente sintética ou fonte real. O número do card, a data e o status devem vir do mesmo snapshot para evitar uma tela internamente inconsistente.

## 5. M02 — Tesouraria e Projeção Financeira

### Funções

| ID | Função | Regra/entrada | Resultado |
|---|---|---|---|
| M02-F01 | Importar saldo conciliado | Bling, mesmo conjunto de contas/filtros | saldo final D-1 |
| M02-F02 | Calcular saldo inicial | `opening(D)=closing(D-1)` | Saldo Inicial do Dia |
| M02-F03 | Classificar movimentos | entradas, saídas, ajustes autorizados | composição diária |
| M02-F04 | Calcular D0 | inicial + movimentos em andamento | `PROVISIONAL` |
| M02-F05 | Calcular D+1–D+60 | fluxo diário, cenário e premissas | `PROJECTED` |
| M02-F06 | Calcular saldo final | `initial + inflows - outflows ± adjustments` | série contínua |
| M02-F07 | Encontrar menor saldo | mínimo de 61 pontos | data, saldo, cenário, composição |
| M02-F08 | Aplicar reserva mínima | política aprovada | risco abaixo da reserva |
| M02-F09 | Comparar cenários | mesmos fatos, premissas distintas | séries comparáveis |
| M02-F10 | Exportar/abrir detalhe | filtro, fonte e regras | relatório auditável |

### Invariantes

- horizonte contém exatamente D0 até D+60;
- nenhuma quebra de data é silenciosa;
- D+1 recebe o saldo final de D0;
- saldo não é acumulado apenas por totais finais;
- ajuste manual exige autorização e auditoria;
- ausência de uma fonte não vira saída zero.

## 6. M03 — Recebíveis

### Submódulos

| ID | Submódulo | Fonte | Funções |
|---|---|---|---|
| M03-A | B2C futuro | Nuvem/Nuvem Pago oficial | importar, normalizar, parcelas, taxa, líquido, data prevista, status |
| M03-B | B2B futuro | Bling | importar, classificar, revisar `UNCLASSIFIED`, excluir espelho B2C |
| M03-C | Total | calculado | somar B2C confirmado + B2B confirmado |
| M03-D | Aging | calculado | vencido, hoje, 1–7, 8–15, 16–30, 31–45, 46–60 |
| M03-E | Detalhe | RAW + tratado | cliente, documento, parcela, saldo aberto, fonte |

### Regras críticas

1. B2C futuro é exclusivamente Nuvem/Nuvem Pago oficial.
2. B2B futuro é somente Bling após classificação aprovada.
3. B2C do Bling é excluído do total futuro.
4. `UNCLASSIFIED` fica fora do confirmado e entra em fila de revisão.
5. Aging usa saldo aberto, exclui cancelado/estornado e aloca cada item em exatamente uma faixa.
6. Taxa, líquido e data de repasse não são inferidos de pedido Nuvemshop.

### Aceite

Duplicidade, parcela, parcial, cancelamento, estorno, chargeback, vencimento no limite, B2C espelhado e classificação ambígua devem ter fixtures e reconciliação de R$ 0,01.

## 7. M04 — Recebimentos Realizados

- fonte: Bling, situação paga/recebida;
- não usar Nuvem para alimentar este indicador;
- funções: listar, filtrar por período, agrupar por conta/categoria, conciliar, abrir documento, mostrar data de recebimento e valor;
- saída: `REALIZED` ou `PENDING_RECONCILIATION`, nunca `PROJECTED`;
- deve preservar recebimento parcial e compensações;
- card executivo mostra período exato e última atualização;
- aceite: fechar com Bling em R$ 0,01, repetir importação sem duplicar e manter último dado em falha.

## 8. M05 — Obrigações a Pagar

### Funções

- listar contas abertas do Bling;
- distinguir original, juros, multa, desconto e saldo aberto;
- filtrar vencidas, hoje e janelas futuras;
- agrupar por fornecedor, categoria, centro de custo e conta;
- mostrar vencimento, competência e origem;
- permitir drill-down para documento e RAW;
- sinalizar parcial, cancelado, estornado e divergente;
- alimentar projeção somente com itens elegíveis e data válida;
- nunca somar novamente um pagamento realizado ao saldo aberto.

### Critérios

`A Pagar` e `Obrigações Vencidas` são conjuntos relacionados, não duas importações independentes. Cada item deve aparecer uma vez na composição e ser reconciliável em R$ 0,01.

## 9. M06 — Pagamentos Realizados

- fonte: Bling, situação paga e data de pagamento;
- funções: período, conta, fornecedor, categoria, centro de custo, detalhe, exportação e reconciliação;
- status: `REALIZED`, `PENDING_RECONCILIATION` ou `DIVERGENT`;
- não usar vencimento como pagamento;
- não usar obrigação aberta como pagamento;
- alimentar fluxo histórico, não obrigação futura novamente;
- aceitar estorno/compensação como evento posterior auditado;
- aceite com tolerância de R$ 0,01 e replay idempotente.

## 10. M07 — Resultado Econômico

### M07-A — DRE

Receita Bruta → Deduções/tributos → Receita Líquida → CPV/CMV → Lucro Bruto → despesas comerciais → administrativas → outras operacionais → Resultado Operacional → Resultado Financeiro → Resultado Antes dos Tributos → Tributos sobre Lucro → Resultado Líquido.

Cada linha precisa de plano de contas, regra de classificação, período, competência, contribuições e snapshot versionado. DRE pronta do Bling nunca é copiada como verdade.

### M07-B — P&L

P&L é um modelo gerencial reconciliável com DRE. Qualquer diferença usa `adjustment_id`, descrição, motivo, valor, usuário, data e aprovação. O sistema não permite exclusão silenciosa de linhas.

### M07-C — EBITDA

Mostrar EBITDA, margem, realizado, orçado, forecast e variação. EBITDA ajustado exige ajustes tipados, versionados e rastreáveis.

### Aceite

Plano de contas aprovado, casos balanceados, diferenças explícitas e reconciliação DRE/P&L precedem o estado confirmado. Sem isso, o módulo permanece `BLOCKED`/`PROVISIONAL`.

## 11. M08 — Planejamento e Forecast

| Função | Comportamento obrigatório |
|---|---|
| Orçamento | account/category, período e valor planejado |
| Realizado comparável | mesma dimensão, período e fonte |
| Variação | valor e percentual, com favorabilidade dependente da natureza |
| Forecast | versão, cenário, premissas, autor e timestamp |
| Histórico | nova versão não sobrescreve anterior |
| Cenários | BASE, CONSERVATIVE, STRESS, OPTIMISTIC |
| Drill-down | linha → período → contribuição → fonte/ajuste |

Receita maior pode ser favorável; despesa maior pode ser desfavorável. A UI não pode aplicar uma regra universal de sinal.

## 12. M09 — Capital de Giro

### Funções

- classificar ativos circulantes e passivos circulantes pelo plano aprovado;
- calcular `current_assets - current_liabilities`;
- mostrar composição, período e versão do plano;
- separar caixa, recebíveis, estoque se aprovado, obrigações e ajustes;
- mostrar tendência e variação;
- impedir `CONFIRMED` sem aprovação contábil das classificações;
- alertar divergência e falta de dados.

## 13. M10 — Tesouraria de Aplicações

### Entradas

- lucro líquido acumulado elegível até M-2;
- valor já investido;
- menor saldo projetado/pior semana;
- reserva mínima;
- recebimentos oficiais elegíveis;
- cenário e versão da regra.

### Regra D-003-A

```text
available_profit = max(0, eligible_accumulated_profit - already_invested)
financial_limit = max(0, worst_week_cash - minimum_reserve)
maximum_investment = max(0, min(available_profit, financial_limit))
```

### Proteções

- não usar lucro sem fechamento aprovado;
- não usar pedido sem fonte oficial de liquidação como recebimento confirmado;
- não permitir aplicação acima do limite financeiro;
- mostrar cálculo, data, versão e cada componente;
- bloquear o estado confirmado até a planilha de referência ou casos dourados serem fornecidos;
- ação de investir, se entrar no escopo, exige autorização separada e auditoria; o MVP é analítico/read-only.

## 14. M11 — Contabilidade Gerencial e Patrimonial

### Submódulos

1. Plano de contas: contas, grupos, natureza, competência e vigência.
2. Períodos: abertura, fechamento, reabertura autorizada e corte.
3. Lançamentos: partidas, débitos, créditos, histórico, origem e usuário.
4. Ajustes: motivo, aprovação, valor e reversão.
5. Balancete: conta, saldo inicial, débitos, créditos e saldo final.
6. Balanço: Ativo, Passivo e Patrimônio Líquido.
7. Validação: `Ativo = Passivo + PL`.
8. Divergência: status `DIVERGENT`, diferença, causa e fila.

Não usar balancete financeiro do Bling como balancete contábil próprio. O módulo pode estar visível no shell antes de estar liberado, mas deve declarar `NOT_IMPLEMENTED`/`BLOCKED` de forma honesta.

## 15. M12 — Conciliações e Divergências

### Funções

- comparar Bling versus registros internos;
- comparar Nuvem versus recebíveis provisórios/oficiais;
- comparar DRE versus P&L;
- validar saldo diário e continuidade;
- detectar duplicidade, ausência, diferença de valor, data ou status;
- criar caso de divergência com severidade e responsável;
- registrar evidência, ajuste aprovado e resolução;
- reprocessar snapshot sem sobrescrever RAW;
- manter histórico de aberto, investigado, resolvido e reaberto.

### Aceite

Toda diferença diferente de zero deve ser visível, quantificada, explicável ou permanecer `DIVERGENT`; nunca ser arredondada para desaparecer.

## 16. M13 — Cadastros e Configurações

### Integrações

Esta é a área editável em que a pessoa configura as APIs pela interface, como em um aplicativo desktop normal. O frontend apresenta e valida o formulário; o worker é o único componente que persiste secrets, executa OAuth, testa a conexão e inicia a sincronização.

| ID | Função | Comportamento da tela | Contrato/segurança |
|---|---|---|---|
| M13-F01 | Listar configurações | mostrar um card por provedor, estado, conta/loja, scopes, última tentativa e último sucesso | nunca retornar token ou secret; estado pode ser `NOT_CONFIGURED`, `CONNECTED`, `STALE`, `AUTH_ERROR`, `UNAVAILABLE` |
| M13-F02 | Editar identificação pública | permitir informar/alterar Client ID, App ID, Redirect URI, Store ID, conta e demais campos não secretos | validação de formato no React e validação autoritativa no worker |
| M13-F03 | Editar secret | permitir digitar ou substituir Client Secret/chave privada com campo mascarado; vazio significa manter o valor salvo, e “remover” exige ação explícita | valor existe somente na memória da interação e no IPC protegido; nunca em localStorage, bundle, log, URL, erro ou estado persistido do React |
| M13-F04 | Salvar configuração | validar campos, pedir confirmação quando houver troca de credencial e enviar um comando de configuração ao worker | worker verifica SID/perfil, grava com DPAPI/ACL, registra auditoria sanitizada e retorna apenas metadados mascarados |
| M13-F05 | Iniciar autenticação | botão “Conectar” abre o navegador externo e acompanha o estado na tela | Authorization Code, `state` aleatório, PKCE quando suportado, callback/relay temporário; login não é digitado em WebView |
| M13-F06 | Testar conexão | executar uma leitura mínima de baixo custo e explicar o resultado em linguagem simples | worker usa o token protegido; resposta diferencia rede, autorização, escopo, rate limit, schema e sucesso |
| M13-F07 | Reconectar | permitir trocar conta/loja ou refazer OAuth sem apagar os dados RAW já importados | tokens antigos são revogados/removidos somente após a nova autorização ser confirmada |
| M13-F08 | Desconectar | pedir confirmação e interromper novos jobs | remove tokens/callbacks, mantém dados históricos e registra auditoria |
| M13-F09 | Sincronizar agora | permitir iniciar polling manual depois de conexão válida | ação passa pelo worker, cria `sync_batch`, usa idempotência e mostra tentativa/sucesso separados |
| M13-F10 | Editar escopos/parâmetros | apresentar escopos de leitura, loja/conta e parâmetros do adapter; salvar uma alteração não inicia sync silencioso | mudanças materiais criam versão e auditoria |
| M13-F11 | Exibir erro e diagnóstico | mostrar mensagem simples, código técnico copiável sem secret, correlação e orientação de correção | redaction obrigatória; payload bruto só no diagnóstico protegido |
| M13-F12 | Configurar fonte futura | manter o card Nuvem Pago editável somente conforme o contrato oficial disponível | sem endpoint oficial, exibir `UNAVAILABLE` e não aceitar campos inventados como taxa, líquido ou repasse |

#### Campos editáveis por provedor

- **Bling:** Client ID, Client Secret, Redirect URI, escopos aprovados e conta/contexto selecionado.
- **Nuvemshop:** App ID/Client ID, Client Secret, Redirect URI HTTPS, Store ID após autorização e escopos aprovados.
- **Nuvem Pago:** esquema de campos só será habilitado quando existir API/arquivo oficial que comprove autenticação e ledger; até lá a tela explica a limitação e não transforma Nuvemshop em fonte financeira.

#### Modelo de dados da tela

O React recebe um `IntegrationConfigView` sem segredo:

```text
provider, display_name, status, account_or_store_id,
redirect_uri, scopes, secret_configured, secret_updated_at,
last_attempt_at, last_success_at, last_error_code, can_edit, can_connect
```

O formulário envia ao worker um comando separado, nunca um DTO executivo:

```text
SaveIntegrationConfig(provider, public_fields, secret_write_intent)
StartOAuth(provider, state, pkce_challenge)
TestIntegration(provider)
DisconnectIntegration(provider)
SyncIntegration(provider)
```

O `secret_write_intent` contém o valor somente durante a chamada IPC e não pode ser ecoado na resposta. O worker devolve um código estável, o estado sanitizado e os horários de tentativa/sucesso. Assim, a experiência é editável e normal para a pessoa usuária, mas a fronteira de segurança continua no backend local.

### Cadastros internos

- empresa/workspace;
- contas financeiras;
- categorias e centros de custo;
- fornecedores e clientes normalizados;
- plano de contas;
- parâmetros de reserva mínima;
- parâmetros de cenário;
- thresholds de alerta;
- períodos e competências;
- usuários, perfis e permissões;
- mapeamentos B2B;
- regras de classificação e suas versões.

Qualquer edição material gera auditoria e, quando alterar cálculo, nova versão da regra e snapshot.

## 17. M14 — Folha e Encargos

### Importação

- arquivo/relatório oficial de folha;
- UTF-8, delimitador `;`, competência `YYYY-MM`;
- hash de arquivo e linha;
- rejeição de linha inválida sem confirmar lote;
- validação de totais por competência;
- PII mascarada na UI;
- histórico de CLT, sócios/pró-labore e desligados;
- departamentos Produção, Comercial e Administrativo;
- provisões, FGTS, encargos e pró-labore.

### Regras protegidas

- adiantamento salarial é compensação contra folha, não nova despesa;
- PLR `NOT_APPLICABLE` até aprovação;
- folha não alimenta DRE/EBITDA confirmado antes de total aprovado;
- arquivo real ainda é dependência externa.

## 18. M15 — Alertas e Centro de Atenção

### Alertas iniciais

1. saldo projetado abaixo da reserva mínima;
2. obrigações vencidas;
3. recebíveis vencidos;
4. grande pagamento próximo;
5. queda relevante de EBITDA;
6. erro de sincronização;
7. divergência contábil;
8. dado desatualizado.

Cada alerta possui código, severidade, regra/threshold versionado, data de detecção, valor, entidade, fonte, status aberto/resolvido, usuário, motivo de resolução e link. Alertas não alteram números; apenas apontam risco.

## 19. M16 — Sincronização e Dados RAW

### Pipeline

```text
agendador
 -> cliente autenticado
 -> timeout/cancelamento/rate limit
 -> paginação/cursor
 -> sync_batch
 -> payload canônico + hash
 -> raw_records append-only
 -> normalização idempotente
 -> contribuições tipadas
 -> snapshot/cálculo
 -> commit integral
 -> cursor/último sucesso
```

### Funções técnicas

- OAuth e renovação segura;
- polling incremental;
- paginação com limites do provedor;
- retry apenas transitório, backoff e jitter;
- `429`, timeout, DNS, TLS e `5xx` classificados;
- schema mismatch bloqueia confirmação;
- checkpoint só após commit;
- replay sem duplicidade;
- último dado válido preservado;
- `integration_sync_logs` com contagens e correlation ID;
- reprocessamento por versão sem alterar RAW histórico;
- redação LGPD exclusivamente auditada.

## 20. M17 — Auditoria e Lineage

### Percursos obrigatórios

- card → snapshot → regra → contribuições → entidade → RAW → provedor;
- DRE/P&L → conta → lançamento/ajuste → fonte;
- alerta → threshold → valor/composição → registro;
- ação de configuração → usuário/SID → autorização → resultado;
- sincronização → lote → páginas → contagens → cursor.

### Eventos auditáveis

- conexão/desconexão/reconexão;
- teste de conexão;
- mudança de parâmetros;
- classificação B2B;
- ajuste manual;
- aprovação/reprovação;
- redação LGPD;
- backup/restore;
- update/repair/uninstall;
- exportação de dados/diagnóstico.

Nenhum evento registra token, senha, payload integral ou PII desnecessária.

## 21. M18 — Segurança, Acesso e LGPD

### Acesso

Perfis: `ADMIN`, `DIRECTOR`, `FINANCE`, `ACCOUNTING`, `VIEWER`. A autorização é no worker por SID/catálogo; a UI só reflete permissões.

### Proteção

- DPAPI e ACL para tokens/client secrets;
- Named Pipe com versão, limite, DACL e SID;
- PostgreSQL loopback/SCRAM, role runtime sem superuser;
- sem secrets em frontend, Git, logs, argumentos, dump ou backup comum;
- logs sanitizados com correlation ID;
- mínimos privilégios e ações destrutivas com confirmação;
- LGPD: minimização, retenção, acesso, redação/tombstone e tratamento de backups.

## 22. M19 — Operação Desktop

### Instalação

- preflight Windows/UAC/reboot/espaço/path/porta/WebView2;
- pacote assinado e hash;
- PostgreSQL dedicado, sem apropriar instalação externa;
- serviços e dependências reais;
- migrations com backup/advisory lock;
- ACL/SID/DPAPI;
- atalho, first-run e smoke test.

### Atualização

- assinatura/hash/compatibilidade;
- backup antes de migration;
- parar UI/worker ordenadamente;
- trocar binários atomicamente;
- smoke e rollback conjunto se schema incompatível.

### Backup/restore

- dump, globals/roles, manifesto, hash, criptografia e retenção;
- não copiar cluster ativo;
- validar dump antes do update;
- restaurar com worker parado;
- validar RAW, constraints, snapshots, auditoria e continuidade;
- reconectar integrações em outra máquina por DPAPI.

### Repair/uninstall

- repair seletivo de binários, serviços, ACL, WebView2, banco e migrations;
- reinstall preserva install ID, SID, dados, secrets e backups;
- uninstall preserva dados por padrão;
- exclusão completa exige escolha explícita, confirmação reforçada e backup.

## 23. Matriz de requisitos do handoff → funções

| Seções do handoff | Decomposição | Requisitos existentes |
|---|---|---|
| §§1–4 | M00, M01 e limites de escopo | GOV-01, GOV-02 |
| §§5–6 | M03, M04, M05, M06, M13, M16 | INT-01 a INT-05, FIN-01/02 |
| §§7–9 | M02 | FIN-00, FIN-04, FIN-05 |
| §10 | M01/M02 | UX-01 |
| §§11–15 | M01, M17 | UX-02, UX-03, UX-04 |
| §§16–21 | M16, M17, M18 | ARC-03, DATA-01/02, INT-04, SEC-01 |
| §§22–26 | M02–M06, M08 | DATA-03, FIN-03/04 |
| §§27–31 | M07, M11, M12 | RESULT-01/02/03, ACC-01 |
| §32 | M14 | PAY-01/02 |
| §§33–34 | M10 | FIN-06 |
| §§35–37 | M08, M15 | PLAN-01/02, ALERT-01 |
| §§38–41 | M17/M18 | DATA-01, SEC-01/02/03 |
| §§42–44 | M01, M02, M13, M16, M19 | ARC-03, UX-02 |
| §§45–47 | governança, M19 | GOV-03, OPS-01/02/03/04 |
| §§48–50 | QA transversal e invariantes | QA-01/02/03/04, GOV-04 |
| §§51–53 | gates e Definition of Done | GOV-03/04, QA-04 |

## 24. Catálogo de aceite por módulo

| Módulo | Aceite funcional mínimo | Evidência de produção |
|---|---|---|
| M00 | first-run, contexto, filtros e saúde sem falso sucesso | instalação em VM limpa e reboot |
| M01 | 20 indicadores, gráficos, alertas, updates e drill-down | comparação visual + estados + lineage |
| M02 | D0–D60, continuidade, menor saldo e cenários | casos dourados R$ 0,00 |
| M03 | B2C/B2B, não duplicidade e aging | reconciliação R$ 0,01 + RAW |
| M04 | recebimentos pagos Bling | reconciliação R$ 0,01 |
| M05 | obrigações abertas e vencidas | reconciliação R$ 0,01 |
| M06 | pagamentos Bling | reconciliação R$ 0,01 |
| M07 | DRE/P&L/EBITDA | plano de contas, oráculo e ajustes |
| M08 | orçamento, variação e forecast imutável | versões e cenários |
| M09 | capital de giro | classificação contábil aprovada |
| M10 | aplicação conforme D-003-A/planilha | casos dourados R$ 0,01 |
| M11 | balancete/balanço e divergência | débito=crédito; A=P+PL |
| M12 | divergência detectada e resolvida | fila, evidência e auditoria |
| M13 | credenciais e parâmetros seguros | OAuth/DPAPI/ACL/UX |
| M14 | folha e compensações | arquivo real e total aprovado |
| M15 | alertas acionáveis | thresholds, severidade e resolução |
| M16 | sync idempotente e resiliente | replay, 429, cursor, schema mismatch |
| M17 | lineage completo e auditoria | percurso reproduzível |
| M18 | autorização, LGPD e secrets | testes negativos, scan e redação |
| M19 | instalação/update/backup/restore | assinatura, VMs e recuperação |

## 25. Estado atual do realinhamento

### Entregue na fundação

- contratos de domínio Money/status/identidade;
- migration e constraints RAW/auditoria;
- worker/Named Pipe/DPAPI/preflight;
- shell React inicial, tooltips e estados fail-closed;
- documentação G0/G1 e fontes arquivadas.

### Ainda não implementado como funcionalidade real

- sidebar completa e roteamento dos 13 módulos;
- DTO executivo rico e filtros globais ligados ao backend;
- valores reais nos 20 cards;
- gráficos, alertas e últimas atualizações alimentados por dados;
- drill-down card → RAW;
- conectores Bling/Nuvemshop/Nuvem Pago;
- DRE/P&L/EBITDA operacional;
- orçamento/forecast, capital de giro e contabilidade;
- folha, conciliação, backup/restore e instalador final.

Esse inventário é intencional: ele evita declarar que uma referência visual foi implementada quando, tecnicamente, só existe a fundação de apresentação.
