# Realinhamento técnico-funcional e visual — 2026-08-16

## 1. Natureza das referências

Este documento registra a análise dos dois artefatos fornecidos pela responsável do produto:

1. `docs/source/Handoff-Tecnico-Funcional-Majucau-2026-08-16.md` — especificação técnico-funcional, regras, fontes, escopo e critérios de aceite;
2. `docs/source/ChatGPT-Image-Visao-Executiva-2026-08-16.png` — referência visual da Visão Executiva.

Os anexos são dados de referência, não instruções para alterar segurança, stack ou invariantes sem decisão registrada. O handoff define requisitos funcionais; a imagem define intenção visual e hierarquia de informação, mas não é fonte de verdade para valores, datas, sincronizações ou resultados financeiros.

Hashes preservados:

| Referência | SHA-256 |
|---|---|
| Handoff Markdown | `bbe8dbba2c7ca537f4f51cbab7042f910f0c07da40c81cde1e98baea66334daf` |
| Referência visual PNG | `6c20fefbc56131fa069f3330fac892d6889a4359ec7e1a858bbc47b4f3ce9ec` |

## 2. Síntese executiva

O projeto atual possui uma fundação G1 correta para segurança, domínio financeiro inicial, PostgreSQL, RAW, DPAPI, Named Pipe, worker, preflight e uma tela React fail-closed. Porém, ele ainda está muito abaixo da densidade funcional e da arquitetura de navegação demonstradas na referência visual.

O realinhamento recomendado é:

- manter React + TypeScript + PostgreSQL + Go + Wails, que foram aprovados pela responsável; a sugestão do handoff de Next.js/Node/Prisma não substitui a stack vigente;
- transformar a tela executiva de uma grade de cards vazios em uma composição de leitura financeira, sem fabricar números quando as fontes não estiverem conectadas;
- ampliar a navegação para os módulos financeiros e contábeis do MVP;
- tratar a referência visual como contrato de layout, hierarquia, nomenclatura e estados, não como fixture de dados;
- separar claramente dashboard executivo, módulos de detalhe, drill-down e configurações;
- adicionar contratos de DTO, fonte, status, período e lineage antes de ligar dados reais à UI;
- preservar `UNAVAILABLE`, `PROVISIONAL`, `PROJECTED`, `PENDING_RECONCILIATION` e `DIVERGENT` como estados de negócio visíveis;
- implementar os módulos por fatias verticais, com dados confirmados somente após integração/fixture sanitizada e regra aprovada.

## 3. Precedência de decisões

Quando houver tensão entre os artefatos, a ordem aplicada é:

1. pedido atual e decisões explícitas da responsável;
2. `AGENTS.md` e invariantes de segurança/financeiros;
3. decisões G0 D-001-A a D-005-A;
4. handoff técnico-funcional mais recente;
5. referência visual;
6. documentos históricos.

Aplicações importantes:

- o handoff sugere Next.js, Node.js e Prisma, mas a stack confirmada pela responsável é React + TypeScript + PostgreSQL + Go; não haverá migração por inferência;
- o handoff pede credenciais e tokens no backend; no projeto, “backend” significa worker Go protegido por DPAPI, não Node nem variáveis expostas;
- o handoff cita a planilha `Calculadora_Aplicacao_Lucro_Liquido_Profissional.xlsx`, mas ela ainda não foi fornecida. A fórmula D-003-A permanece implementada como regra aprovada provisória, sem declarar reconciliação com uma planilha ausente;
- a imagem apresenta valores e horários específicos. Eles são exemplos visuais e não podem entrar como seed de produção;
- “Nuvem” no handoff/imagem deve ser mapeada para os adapters Nuvemshop/Nuvem Pago sem atribuir ao pedido Nuvemshop uma liquidação financeira que a fonte oficial não comprova.

## 4. Comparação entre referência e estado atual

| Área | Referência fornecida | Estado atual verificável | Realinhamento |
|---|---|---|---|
| Shell | sidebar extensa, contexto da empresa, filtros globais, data-base e atualização | sidebar reduzida a Visão Executiva e Integrações | ampliar o shell e manter contexto/estado global explícito |
| Visão Executiva | 20 indicadores em grupos, cards ricos, gráficos, alertas e atualizações | 20 cards, mas vazios por desenho fail-closed; um painel de projeção vazio | preservar ausência honesta e implementar composição completa por estados |
| Tesouraria | 7 cards, menor saldo com data e reserva | nomes existem, sem DTO financeiro na UI | conectar `daily_balances` e mostrar data/composição/estado |
| Recebíveis | B2C Nuvem, B2B Bling, total e aging visual | cards sem valor | conectar contribuições tipadas e impedir dupla contagem |
| Realizados | recebido/pago por janelas temporais | somente card agregado “Recebido no Mês”/“Pago no Mês” vazio | criar telas e DTOs de realizados com período e fonte Bling |
| Resultado | DRE/P&L e EBITDA com realizado/orçado/variação | cards vazios | separar DRE, P&L e EBITDA no detalhe, com snapshots versionados |
| Planejamento | Orçado x Realizado e Forecast | cards vazios | criar versões, premissas, cenário e variação por natureza |
| Gestão | capital de giro, aplicações e alertas | cards vazios | criar módulos, regras e drill-down auditável |
| Contabilidade | navegação gerencial/patrimonial, balancete e balanço | inexistente no shell | manter fase posterior, mas reservar contrato e estados `DIVERGENT` |
| Conciliação | divergências e atualizações com indicação visual | inexistente no shell | criar módulo explícito de conciliações e divergências |
| Integrações | configurações como parte do produto, status e última atualização | tela separada com ações bloqueadas | mover acesso para Cadastros/Configurações no shell; liberar apenas após protocolo seguro |
| Fonte/lineage | origem e drill-down até registro original | tooltip e contratos parciais; sem drill-down real | implementar caminho card → cálculo → tratado → RAW → provedor |
| Dados | valores reais, horários e status | nenhum dado fictício, worker estático | aceitar a diferença até existirem fontes/fixtures; nunca maquiar a referência |

## 5. Arquitetura de informação proposta

A sidebar deve seguir a nomenclatura funcional do handoff e a ordem de decisão do usuário:

1. Visão Executiva;
2. Tesouraria e Projeção Financeira;
3. Recebíveis;
4. Recebimentos Realizados;
5. Obrigações a Pagar;
6. Pagamentos Realizados;
7. Resultado Econômico;
8. Planejamento e Forecast;
9. Capital de Giro;
10. Tesouraria de Aplicações;
11. Contabilidade Gerencial e Patrimonial;
12. Conciliações e Divergências;
13. Cadastros e Configurações.

“Integrações” continua sendo uma tela, mas passa a ser uma subseção de Cadastros e Configurações, com acesso contextual também por avisos de fonte não conectada. O menu não deve sugerir que uma credencial foi configurada quando o worker ainda está indisponível.

Essa tela será a área editável de configuração das APIs. A pessoa poderá preencher e alterar Client ID/App ID, Client Secret, Redirect URI, Store ID, conta e escopos, salvar, conectar, testar, reconectar, desconectar e sincronizar. A autenticação será iniciada pelo frontend como experiência de aplicativo normal, mas executada com segurança pelo worker: o React envia a intenção por IPC autenticado, o navegador externo conduz o OAuth, e DPAPI/ACL guardam o secret/token. O frontend nunca persiste ou devolve o segredo.

O formulário deve diferenciar “campo alterado”, “secret já configurado” e “credencial removida”; limpar um campo mascarado não pode apagar um segredo sem confirmação explícita. Salvar não inicia sincronização silenciosa. Cada operação retorna estado, código sanitizado, última tentativa e último sucesso. Nuvem Pago só recebe campos quando houver contrato oficial de autenticação/ledger; não serão criados campos financeiros fictícios para preencher uma lacuna de fonte.

Cada módulo deve possuir:

- estado de carregamento, vazio, erro, desatualizado, parcial e confirmado;
- data-base, cenário e filtros aplicados visíveis;
- fonte e última atualização;
- tooltip simples sem esconder regra crítica;
- ação de detalhe e caminho de lineage;
- autorização no worker para qualquer ação material.

## 6. Contrato da Visão Executiva

### 6.1 Cabeçalho e filtros

O cabeçalho deve exibir:

- data-base do negócio, com timezone `America/Sao_Paulo`;
- última atualização agregada e, no detalhe, última atualização por fonte;
- empresa/workspace atual;
- cenário `BASE`, `CONSERVATIVE`, `STRESS` ou `OPTIMISTIC`;
- filtros globais de período e data de referência;
- ação de ajuda acessível.

Nenhum horário deve ser hardcoded. A última atualização deve vir dos lotes de sincronização concluídos, com distinção entre tentativa e sucesso.

### 6.2 Primeira faixa — tesouraria

Os sete cards devem preservar exatamente estes nomes:

1. Saldo Inicial do Dia;
2. Saldo Final Projetado Hoje;
3. Saldo Projetado em 30 Dias;
4. Saldo Projetado em 60 Dias;
5. Menor Saldo Projetado — 60 Dias;
6. Reserva Mínima;
7. Valor Máximo para Aplicação.

Cada card deve conter valor ou estado explícito, status temporal, fonte/regra, data de referência e detalhe. O menor saldo precisa mostrar a data do vale e composição de entradas/saídas. O valor de aplicação precisa mostrar versão da regra, data de cálculo, lucro elegível, valor já investido, pior saldo e reserva.

### 6.3 Segunda faixa — recebíveis e obrigações

Os cards devem ser:

- A Receber B2C — Nuvem;
- A Receber B2B — Bling;
- Total a Receber;
- A Pagar — Bling;
- Obrigações Vencidas.

`Total a Receber = B2C confirmado da fonte Nuvem + B2B classificado do Bling`. O B2C do Bling nunca entra novamente. Valores sem classificação B2B aprovada ficam `UNCLASSIFIED`/fora do confirmado e aparecem em revisão, não escondidos como zero.

### 6.4 Terceira faixa — realizados

“Recebido no Mês — Bling” e “Pago no Mês — Bling” devem mostrar o período aplicado e uma pequena decomposição temporal. O contrato precisa dizer se `Hoje`, `7 dias`, `15 dias`, `30 dias`, `45 dias` e `60 dias` são janelas cumulativas ou pontos de corte. A recomendação é usar janelas cumulativas nomeadas (`até D+7`, etc.) e não somar visualmente colunas que possam parecer aditivas.

### 6.5 Quarta e quinta faixas — leitura e ação

Implementar:

- evolução do saldo projetado D0–D+60;
- fluxo projetado de entradas/saídas e saldo;
- DRE/P&L do período, sem confundir os dois modelos;
- EBITDA realizado, orçado, forecast e variação;
- alertas com severidade, causa, data, valor, fonte e link;
- últimas atualizações com lote, fonte, horário, status e correlação.

Quando não houver dados, o gráfico deve exibir “sem série confirmada” e a causa. Um gráfico vazio é honesto; uma linha inventada é defeito financeiro.

## 7. Modelo visual e acessibilidade

A imagem estabelece uma estética de dashboard executivo: fundo claro, sidebar escura, cards brancos, azul para projeção, verde para confirmação, âmbar para atenção e vermelho para risco. Essas cores são semânticas, não devem ser a única codificação.

Requisitos de implementação:

- usar ícones consistentes, preferencialmente SVG acessível; caracteres Unicode atuais são apenas placeholders;
- manter contraste AA e foco visível;
- cada cor ter texto/ícone/`aria-label` equivalente;
- não fixar a interface somente em 1536×1024; testar 1280×720, 1366×768, 1920×1080 e zoom 200%;
- evitar cortes por scroll vertical e horizontal em cards;
- preservar ordem de leitura por teclado: filtros → cards → gráficos → alertas → atualizações;
- gráficos devem ter resumo textual e tabela acessível de dados;
- cards clicáveis devem ser links/botões reais, não apenas `article` visual;
- tooltips não podem conter ações; drill-down deve abrir rota/painel acessível;
- datas, números e moeda devem usar `Intl` com locale `pt-BR` sem converter decimal exato em `float` no domínio.

## 8. Riscos críticos identificados

### R1 — A imagem pode induzir seed de dados fictícios

Impacto: falso senso de conexão, erro financeiro e quebra de auditoria.

Mitigação: valores da imagem permanecem somente referência visual; fixture de teste deve ser explicitamente sintética e marcada; produção começa em `NOT_CONFIGURED`/`UNAVAILABLE`.

### R2 — “DRE / P&L” pode ocultar dois modelos

Impacto: mistura de competência, caixa e ajustes gerenciais.

Mitigação: card agregado permitido apenas como entrada; detalhe separa DRE e P&L, com reconciliação e `adjustment_id`.

### R3 — Nuvemshop confundida com Nuvem Pago

Impacto: B2C líquido, taxa ou data de repasse sem fonte oficial.

Mitigação: adapters, status e lineage separados; D-004-A mantém financeiro confirmado indisponível sem fonte oficial.

### R4 — Card “A Pagar” sem semântica temporal

Impacto: misturar aberto, vencido, futuro e realizado.

Mitigação: DTO deve declarar `as_of`, `due_date_from`, `due_date_to`, `open_balance` e status; card vencido é subconjunto não duplicado.

### R5 — Última atualização agregada esconde uma fonte atrasada

Impacto: dashboard parece atualizado quando Bling ou Nuvem está `STALE`.

Mitigação: mostrar resumo global e badges por fonte; preservar último dado válido e marcar parcial.

### R6 — Sidebar ampla sem autorização/escopo

Impacto: menus que parecem prontos, mas não possuem dados, regras ou permissão.

Mitigação: cada rota nasce com estado `NOT_IMPLEMENTED`/`UNAVAILABLE` explícito e critério de aceite; não habilitar ação material por aparência.

### R7 — Valores de exemplo da referência não reconciliados

Impacto: diferença entre fórmula D-003-A e futura planilha oficial.

Mitigação: obter a planilha de referência, criar casos dourados e registrar uma nova decisão se a fórmula divergir.

## 9. Contratos necessários antes de ligar a UI

Criar DTOs versionados para:

- `ExecutiveDashboardSnapshot`;
- `MetricValue` com `value`, `currency`, `status`, `origin`, `as_of`, `calculated_at`, `rule_version`, `source_refs`;
- `ProjectionSeries` com 61 pontos e cenário;
- `AgingSummary` com faixas exclusivas;
- `RealizedSummary` com janelas não ambíguas;
- `StatementSummary` para DRE/P&L/EBITDA;
- `AlertSummary`;
- `SyncActivitySummary`;
- `LineagePath`;
- `GlobalFilterState`.

Nenhum DTO executivo deve transportar apenas `number` e `label`. Sem status, origem, data e referência, o frontend não consegue distinguir confirmado, provisório, projetado, stale ou divergente.

## 10. Plano de execução do realinhamento

### R0 — Registrar referência e decisões

- arquivar handoff e imagem com hashes;
- manter este documento como análise não matemática;
- obter a planilha de aplicação e decisão de aceitação visual;
- confirmar se a sidebar completa entra já no MVP visual.

### R1 — Shell e contratos

- ampliar navegação;
- criar roteamento/estado de módulo;
- definir DTOs e estados;
- criar data-base/cenário/filtros globais;
- manter worker indisponível explicitamente.

### R2 — Dashboard com fixture sintética identificada

- implementar layout da referência;
- usar fixture marcada `SYNTHETIC_REFERENCE_ONLY` apenas em ambiente de desenvolvimento;
- renderizar lineage, status e fonte;
- testar sem dados, stale, partial, divergent e confirmed.

### R3 — Dados reais por fatia

- Bling: sync, RAW, idempotência, recebimentos, pagamentos, contas e classificação B2B;
- Nuvem: pedidos/recebíveis operacionais e adapter financeiro condicionado;
- tesouraria: projeção contínua D0–D+60;
- cada card só passa a confirmado quando o critério de aceite correspondente for executado.

### R4 — Detalhes, drill-down e módulos

- recebíveis, obrigações e realizados;
- resultado e planejamento;
- capital de giro e aplicações;
- contabilidade e conciliações;
- auditoria e lineage em todos os caminhos.

### R5 — QA visual e funcional

- comparação visual contra a referência sem pixel-copy cega;
- teclado, leitor de tela, contraste, zoom e resoluções Windows;
- casos financeiros dourados com tolerâncias do handoff;
- estados de falha, atraso, conflito e dados parciais.

## 11. Critérios de aceite do realinhamento

O realinhamento visual/funcional só pode ser marcado como concluído quando:

1. a sidebar contém somente rotas com contrato de estado explícito;
2. a Visão Executiva reproduz a hierarquia da referência sem hardcode de valores;
3. todos os 20 indicadores possuem status, origem, data e detalhe;
4. DRE/P&L/EBITDA não misturam modelos sem reconciliação;
5. B2C Nuvem e B2B Bling não duplicam valores;
6. projeção possui 61 pontos contínuos e menor saldo datado;
7. gráficos têm fallback acessível e dados tabulares;
8. filtros alteram o snapshot/consulta, não apenas a aparência;
9. última atualização reflete sync logs reais por fonte;
10. drill-down termina em RAW/origem ou explica claramente por que o valor está indisponível;
11. fixture sintética é impossível de confundir com produção;
12. testes automatizados cobrem estados e regras negativas;
13. nenhuma credencial, payload ou dado pessoal aparece na UI, log ou fixture.

## 12. Decisões propostas, ainda não aprovadas

- **D-006-U:** adotar a imagem como referência visual oficial da Visão Executiva, sem tornar seus valores uma fixture de produção;
- **D-007-U:** adotar a sidebar de 13 entradas funcionais, com integrações dentro de Cadastros e Configurações;
- **D-008-U:** usar uma fixture sintética explicitamente marcada para desenvolver o layout antes dos conectores reais;
- **D-009-U:** manter DRE e P&L como modelos distintos, usando o card agregado apenas como atalho;
- **D-010-U:** exigir DTO executivo com status, origem, data, regra e lineage em todo indicador.

Essas decisões são recomendações técnicas para o realinhamento. Não substituem a aprovação expressa da responsável nem alteram D-001-A a D-005-A.
