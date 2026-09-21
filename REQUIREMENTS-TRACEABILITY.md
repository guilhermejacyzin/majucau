# Matriz de rastreabilidade — Majucau Financial Intelligence

- **Baseline:** handoff técnico-funcional de 30 páginas, revisado em 2026-08-15
- **Estado avaliado:** G0 aprovado em 2026-08-16; fundação G1 em implementação e ainda sem aceite do gate
- **Objetivo:** impedir que requisito, regra protegida ou evidência de aceite desapareça entre arquitetura, implementação, teste e produção
- **Decomposição funcional:** `docs/FUNCTIONAL-DECOMPOSITION-v1.md` detalha cada módulo, subfunção, contrato, regra e critério de aceite relacionado às seções do handoff

## Legenda

- `DOCUMENTED`: contrato e gate definidos; implementação ainda não provada.
- `PARTIAL`: há cobertura, mas falta regra, detalhe ou evidência material.
- `BLOCKED`: depende de aprovação, fonte oficial, credencial, oráculo ou revisão externa.
- `NOT_STARTED`: previsto, sem implementação/evidência.
- `VERIFIED`: evidência correspondente foi executada e registrada; não implica aprovação automática dos demais gates.

## Governança e escopo

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| GOV-01 | Consolidar posição, projeção, recebíveis, realizados, obrigações, resultado, planejamento, gestão, contabilidade, indicadores, alertas, conciliações e rastreabilidade | pp. 1–3, §§1–4 | ADR §§1–3; ERD §§1–7; Plano fases 7–11 | G0–G7 | Matriz funcional, suíte e aceite por módulo | PARTIAL |
| GOV-02 | Não desenvolver previsão por SKU/demanda, produção, otimização de estoque, compras recomendadas, pricing avançado, CRM, marketing, Appmax, Mercado Pago ou Itaú | pp. 3–4, §4 | Plano §1, “Fora do escopo”; AGENTS “Produto e plataforma” | G0 | Revisão de dependências, rotas e backlog sem itens fora do escopo | DOCUMENTED |
| GOV-03 | Entregar ADR, ERD, mapa de APIs, dicionário e plano antes do código; manter `AGENTS.md` | pp. 24, 29–30, §§45–53 | Cinco artefatos G0; AGENTS; Registro G0 | G0/G1 | Aprovação explícita e `AGENTS.md` na raiz do repositório | VERIFIED |
| GOV-04 | Regra material não muda sem aprovação expressa e nova evidência | p. 28, §50 | Data §§14–15; AGENTS “Invariantes”; Registro G0 | G0/G3/G5 | Versão, aprovação, diff e regressão por regra | BLOCKED |

## Arquitetura, plataforma e operação

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| ARC-01 | Windows 10/11 x64, React/TypeScript, Go, PostgreSQL local, Wails e instalador nativo sem Docker | pp. 11–12, 29, §§16–17 | ADR §§2–4, 9; Plano fases 1–2 | G1/G7 | Builds reproduzíveis e smoke em VMs limpas | NOT_STARTED |
| ARC-02 | Worker continua com UI fechada; Named Pipe autenticado, ACL, SID e protocolo versionado | pp. 11, 22, §§16, 39–42 | ADR §§3.2–3.3, 7; ERD §3; AGENTS | G1/G6 | Testes janela fechada, reboot, SID inválido, DACL e concorrência | NOT_STARTED |
| ARC-03 | Dashboard nunca consulta APIs diretamente; usar sincronização, banco local e última atualização por fonte | pp. 22–23, §§42–43 | ADR §§3, 8; API §2; Plano fases 5–8 | G2/G4 | Teste offline/cache, network inspection e UI por fonte | NOT_STARTED |
| OPS-01 | NSIS x64 assinado, WebView2, PostgreSQL dedicado, porta sob lock, serviços, migrations, atalho, upgrade/uninstall | pp. 24–25, §§46–47 | ADR §9; Plano fases 2 e 12 | G1/G6/G7 | Matriz Windows online/offline, UAC, porta ocupada, upgrade e uninstall | NOT_STARTED |
| OPS-02 | Backup/restore com dump, globals, manifesto, hash, criptografia, retenção e validação pós-restore | pp. 24–25, §§46–47 | ADR §10; Plano §15.1 | G6/G7 | Backup e restore reais, incluindo RAW, snapshots, roles e continuidade | NOT_STARTED |
| OPS-03 | Atualização aceita somente pacote assinado e mantém compatibilidade binário/schema | criticidade e operação do handoff | ADR §9.2; Plano §15.2 | G6/G7 | Assinatura/hash, migration, smoke, falha e restore conjunto | NOT_STARTED |
| OPS-04 | Instalação em máquina de terceiro detecta, recupera ou bloqueia com segurança falhas conhecidas e produz diagnóstico sanitizado para falhas desconhecidas | solicitação explícita da usuária em 2026-08-16 | AGENTS “Instalação em máquinas de terceiros”; matriz de resiliência Windows; helper de preflight | G1/G6/G7 | Casos automatizados + matriz executada em VMs para install/repair/upgrade/rollback/uninstall e falhas parciais | PARTIAL |

## Dados, sincronização e integrações

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| DATA-01 | RAW versionado sem sobrescrita operacional; lineage obrigatório até fonte original | pp. 1, 12, 21, 30 | ERD §§1, 5–7; Data §§3, 5; AGENTS | G2/G4/G5 | Reprocessamento e percurso card → regra → entidade → RAW | PARTIAL |
| DATA-02 | Idempotência, chave de origem, hash, lote e logs com contagens/erros/retries | p. 13, §§19–20 | ERD §3; Data §4.2; view `integration_sync_logs` | G2 | Replay, falha antes do commit, cursor e relatório de lote | PARTIAL |
| DATA-03 | Separar caixa/competência, realizado/projetado, datas e decimal exato | pp. 1, 6–7, 16 | ADR §4; Data §§2, 9, 15; Registro D-005 | G0/G3/G5 | D-005 aprovada; faltam casos executados de timezone, limites, precisão e arredondamento | PARTIAL |
| INT-01 | Bling: contas a pagar, realizados, B2B, contatos/fornecedores, notas, categorias, compras, contas financeiras e caixa/banco | pp. 4–5, §5.1 | API §3; Data §§6–8; Plano fase 5 | G2/G3 | Credencial real, OpenAPI/fixtures, reconciliação e RAW | PARTIAL |
| INT-02 | Nuvemshop: OAuth, loja, escopos, pedidos, clientes, status, parcelas, cancelamento/reembolso | pp. 4–6, §§5.2, 6 | ADR §6.3; API §4; Registro D-002 | G2 | Redirect aceito, relay publicado, E2E, replay/TTL/PKCE | BLOCKED |
| INT-03 | Nuvem Pago: taxa, líquido, repasse e data somente por fonte oficial | pp. 4–6, §§5.2, 6; quatro capturas de tarifas aprovadas em 2026-09-20 | ADR §6.4; API §5.1–5.2; `docs/source/Nuvem-Pago-Taxas-2026-09-20.md`; Registro D-004 | G2/G3/G7 | Tarifário aprovado e versionado + API/arquivo oficial de ledger, contrato e reconciliação R$ 0,01 | PARTIAL |
| INT-04 | Polling incremental, paginação, limite, retry/backoff, schema mismatch e último dado válido | pp. 13, 22–23, §§20–21, 42–43 | ADR §8; API §§2–6; AGENTS | G2/G6 | Testes 429, timeout, cursor, schema incompatível e stale | PARTIAL |
| INT-05 | Área frontend editável para configurar dados públicos e secrets de Bling/Nuvemshop, iniciar OAuth, testar, reconectar, desconectar e sincronizar; Nuvem Pago segue contrato oficial | solicitação da usuária + pp. 22–23 | ADR §6; API §§7.1–7.3; Functional Decomposition M13-F01–F12; Plano fase 4; evidência `docs/evidence/NUVEMSHOP-CREDENTIAL-CONFIG-2026-09-20.md` | G2/G4/G7 | Formulários Bling/Nuvemshop e IPC/DPAPI/redaction cobertos; OAuth Nuvemshop, E2E real, ACL e reconciliação ainda pendentes | PARTIAL |

## Tesouraria, recebíveis e obrigações

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| FIN-00 | Saldo Inicial D0 nasce do saldo financeiro conciliado D-1 do Bling no mesmo conjunto de contas e filtros | pp. 4, 6, §§5.1, 7 | Data §6.3; ERD §4; Plano fase 5 | G2/G3 | Fixture por conta, composição, RAW e reconciliação com Bling em R$ 0,01 | NOT_STARTED |
| FIN-01 | B2C futuro somente Nuvem; B2B futuro somente Bling; excluir B2C Bling; total = B2C + B2B | pp. 5–6, §6 | API §3.5; Data §7; Registro D-001/D-004 | G2/G3 | D-001/D-004 aprovadas; faltam fixtures, fonte oficial B2C e reconciliação | PARTIAL |
| FIN-02 | Realizados e contas a pagar vêm do Bling e reconciliam | pp. 2–6, §§3, 6 | API §3.2; Data §§7.2, 8; Plano fase 5 | G2/G3 | Recebimentos/pagamentos/obrigações R$ 0,01 | NOT_STARTED |
| FIN-03 | Aging vencido, hoje, 7, 15, 30, 45 e 60 dias sem dupla contagem | pp. 2–3, §3 | Data §7; Plano fase 5; Registro D-005 | G3 | D-005 aprovada; faltam casos executados de limites, parciais, cancelados, timezone e totais | PARTIAL |
| FIN-04 | Saldo inicial D = final D-1; saldo final = inicial + entradas - saídas ± ajustes; D0 provisório; D+1–D+60 projetado | pp. 6–7, §§7–9, 25–26 | Data §9; ERD §5; Plano fase 7 | G3 | D0–D60, viradas, continuidade R$ 0,00 e cenários | NOT_STARTED |
| FIN-05 | Menor saldo diário D0–D60 com data, cenário e composição | p. 7, §9 | ERD §5; Plano fase 7 | G3 | Caso dourado e drill-down de entradas/saídas | NOT_STARTED |
| FIN-06 | Valor Máximo para Aplicação reproduz planilha ou regra formalmente aprovada | pp. 19–20, §§33–34 | Data §12; Registro D-003 | G5/G7 | Fórmula D-003-A aprovada; faltam casos dourados e reconciliação de R$ 0,01 | PARTIAL |

## Experiência executiva e acessibilidade

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| UX-01 | Primeira tela usa nomes oficiais e todos os cards obrigatórios | pp. 7–9, §§10–11 | Plano fase 8; AGENTS “Frontend” | G4 | Teste por label/card e comparação com referência | DOCUMENTED |
| UX-02 | Preservar sidebar, cards clicáveis, estados, gráficos, alertas, drill-down, filtros e atualização do wireframe | p. 23, §44 | Plano fase 8 | G4 | Arquivo do wireframe aprovado, comparação visual lado a lado e checklist | BLOCKED |
| UX-03 | Cada card possui nome, valor, status, fonte, atualização, comparação, ação e drill-down | pp. 9–10, §§12–15 | Data §§3, 11; Plano fase 8 | G4/G5 | Teste de contrato e percurso completo por card | NOT_STARTED |
| UX-04 | Tooltips simples em todos os componentes funcionais, acessíveis e sem esconder regra crítica | solicitação da usuária | ADR §5.2; Plano fase 4; AGENTS | G4/G7 | Hover, foco, Escape, axe, leitor de tela, teclado e zoom 200% | NOT_STARTED |

## Resultado, planejamento, contabilidade e folha

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| RESULT-01 | DRE própria parametrizável; não usar DRE pronta do Bling | p. 17, §27 | Data §11; ERD §6; Plano fase 9 | G5 | Plano de contas, classificação, reconciliação e lineage | BLOCKED |
| RESULT-02 | P&L reconciliável; ajustes têm motivo, valor, usuário, data e aprovação | p. 17, §28 | ERD §§6–7; Plano fase 9 | G5 | Diferenças, ajuste auditado e novo snapshot | BLOCKED |
| RESULT-03 | EBITDA e margem com realizado, orçado, forecast e variação | p. 18, §29 | Views/snapshots do ERD; Plano fase 9 | G5 | Casos aprovados e alerta de queda relevante | BLOCKED |
| ACC-01 | Balancete e balanço próprios; débitos = créditos; Ativo = Passivo + PL; divergência explícita | pp. 18–19, §§30–31 | ERD §6; Plano fase 11 | G5/G7 | Casos balanceados/desbalanceados e `DIVERGENT` | BLOCKED |
| PAY-01 | Folha com Produção/Comercial/Administrativo, CLT, sócios/pró-labore e desligados | pp. 4, 19, §§5.3, 32 | Data §16; Plano fase 9 | G5 | Arquivo real, parser, totais e PII | BLOCKED |
| PAY-02 | Adiantamento não é nova despesa; PLR não aplicável | p. 19, §32 | Data §16; Plano fase 9; AGENTS | G5 | Casos de compensação e rejeição de PLR sem aprovação | DOCUMENTED |
| PLAN-01 | Orçado x Realizado com favorabilidade por natureza | p. 20, §35 | Data §10.1; ERD §6; Plano fase 10 | G5 | Receita/despesa favorável/desfavorável e percentuais | NOT_STARTED |
| PLAN-02 | Forecast versionado, premissas, cenário, usuário e histórico não sobrescrito | p. 21, §36 | Data §10.2; ERD §6; Plano fase 10 | G5 | Versões imutáveis e auditoria | NOT_STARTED |
| MGMT-01 | Capital de Giro calculado pelo motor Majucau | pp. 1–3, §3 | Data §17; view `working_capital_snapshots` | G5 | Plano de contas circulante aprovado e casos dourados | BLOCKED |
| ALERT-01 | Oito tipos iniciais de alerta do handoff | p. 21, §37 | Data §17; ERD §7; Plano fase 8 | G3/G4/G5 | Aprovação de thresholds/severidade, disparo, resolução e referência por tipo | PARTIAL |

## Segurança, privacidade e autorização

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| SEC-01 | Credenciais nunca no código, frontend, Git, logs ou backup comum; tokens protegidos no backend | pp. 1, 22, §39 | ADR §7; AGENTS “Segurança” | G2/G6/G7 | Scan, DPAPI, logs, dump, reboot, upgrade e reconexão | PARTIAL |
| SEC-02 | RBAC ADMIN/DIRECTOR/FINANCE/ACCOUNTING/VIEWER e autorização no worker | p. 22, §40 | ERD §§3, 7; Plano §15.3 | G6/G7 | Catálogo, SID, matriz e testes negativos por ação | PARTIAL |
| SEC-03 | LGPD: mínimo necessário, acesso, logs, criptografia e exclusão/anonimização | p. 22, §41 | ADR §§7, 10; Data §5; Registro D-005 | G6/G7 | Procedimento ponta a ponta em RAW, índices e backups | BLOCKED |

## Qualidade e critérios de aceite

| ID | Requisito | Fonte | Artefato/contrato | Gate | Evidência final | Estado atual |
|---|---|---|---|---|---|---|
| QA-01 | A pagar, B2B, B2C, recebimentos e pagamentos fecham em R$ 0,01 | pp. 26–28, §48 | API §9; Plano fases 5–7 | G2/G3/G7 | Relatório por fonte, filtros, oráculo e diferença | NOT_STARTED |
| QA-02 | Saldo diário e continuidade fecham em R$ 0,00; aplicação em R$ 0,01 | pp. 27–28, §48 | Plano fases 7 e 10 | G3/G5/G7 | Execução D0–D60 e casos dourados | NOT_STARTED |
| QA-03 | Testar duplicidade, idempotência, datas, negativos, cancelamentos, estornos, vencidos, ausência de sync, mudança de status, projeção, cenário, aplicação e arredondamento | p. 28, §49 | Plano §16; AGENTS DoD | G2/G3/G5/G7 | Suíte automatizada e relatório por caso | DOCUMENTED |
| QA-04 | “Pronto” exige código, testes, lineage, fonte, regra, erros, logs e aceite | p. 30, §§52–53 | Plano §16; AGENTS DoD | G4–G7 | Checklist preenchido e evidência vinculada | DOCUMENTED |

## Decisão resolvida e evidências ainda inexistentes

- **Resolvido:** G0 e D-001-A a D-005-A aprovados explicitamente em 2026-08-16 e registrados em `G0-DECISION-REGISTER.md`.

Permanecem pendentes:

1. Credenciais/aplicativos reais Bling e Nuvemshop fornecidos fora de chat/código.
2. Redirect Nuvemshop homologado e relay HTTPS publicado.
3. API ou arquivo oficial do Nuvem Pago com taxa, líquido, agenda e data prevista.
4. Payloads sanitizados, OpenAPI versionado e fixtures reais.
5. Plano de contas e revisão contábil de DRE/P&L/EBITDA/Capital de Giro.
6. Arquivo real de folha e validação de seu total por competência.
7. Certificado/serviço Authenticode.
8. VMs limpas Windows 10/11 e evidências de instalação, atualização, backup e restore.
9. Arquivo do wireframe aprovado da Visão Executiva para comparação visual.

Nenhum desses itens deve ser reinterpretado como concluído por existir uma intenção no plano. O estado muda somente quando a evidência correspondente for produzida e revisada.
