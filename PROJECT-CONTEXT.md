# Contexto profundo — Majucau Financial Intelligence

Este arquivo é o ponto de retomada técnico-funcional do projeto. Ele descreve objetivo, decisões, arquitetura, dados, segurança, regras financeiras, gates e pendências. Leia-o junto com `AGENTS.md`; em caso de divergência, a decisão aprovada mais recente e registrada prevalece, sem enfraquecer segurança ou invariantes por inferência.

## 1. Objetivo, usuária e estado

O objetivo é entregar um aplicativo desktop single-user para Windows 10/11 x64 que forneça inteligência financeira executiva, sincronize fontes oficiais, calcule regras próprias deterministicamente e permita rastrear cada número até sua origem.

- uma pessoa utiliza a instalação;
- o fluxo final não exige terminal, Docker, Node.js ou configuração manual do PostgreSQL;
- o aplicativo abre pelo atalho após a instalação;
- integrações são configuradas por tela segura no first-run ou depois;
- PostgreSQL, worker, jobs, backup e migrations são internos.

Estado verificável em 2026-08-16:

- G0 aprovado com D-001-A, D-002-A, D-003-A, D-004-A e D-005-A;
- G1 em implementação e ainda sem aceite;
- sincronização autorizada para o repositório GitHub privado indicado pela usuária;
- sem credenciais, aplicativos cadastrados, payloads reais ou teste ponta a ponta;
- produção bloqueada pelos gates G1–G7 e evidências externas.

## 2. Escopo

Dentro do produto: visão executiva; tesouraria e projeção D0–D+60; recebíveis B2B e, quando houver fonte oficial, B2C; obrigações e realizados; DRE, P&L, EBITDA, orçamento, forecast, capital de giro, aplicação, alertas, conciliação, auditoria, lineage, Bling, Nuvemshop, adapter Nuvem Pago, folha CSV, backup, restore e update.

Fora do MVP: previsão por SKU/demanda, planejamento de produção, otimização de estoque, recomendação de compras, pricing comercial avançado, CRM, marketing e integrações diretas com Appmax, Mercado Pago ou Itaú. Inclusão exige nova decisão de escopo.

## 3. Decisões aprovadas

### D-001-A — B2B conservador

Ordem: override manual auditado; correlação direta/exata com Nuvem exclui B2C do Bling; CNPJ válido classifica B2B; canal/categoria/contato aprovado classifica B2B; demais ficam `UNCLASSIFIED` e fora do total confirmado.

Sem identificador direto, o match exige igualdade exata de conexão/loja, número externo, hash do documento, moeda, bruto e data de emissão. Zero ou vários candidatos permanecem `UNCLASSIFIED`. Matching probabilístico é proibido.

### D-002-A — OAuth Nuvemshop

Desktop cria `state`, sessão e segredo de pareamento; navegador usa redirect HTTPS registrado; relay recebe apenas `code/state`; leitura é de uso único e expira em cinco minutos; worker local troca o código e protege o token.

PKCE é obrigatório quando suportado. Sem PKCE, permanece bloqueado até threat model e aceite `D-002-RISK`. Clipboard, token manual, login em WebView e client secret embutido são proibidos.

### D-003-A — aplicação

```text
eligible_accumulated_profit = lucro líquido elegível acumulado até M-2
available_profit = max(0, eligible_accumulated_profit - already_invested)
worst_week_cash = menor closing_balance diário em D0..D+60
financial_limit = max(0, worst_week_cash - minimum_reserve)
maximum_investment = max(0, min(available_profit, financial_limit))
```

Somente recebimentos com fonte oficial e data válida participam.

### D-004-A — Nuvem Pago

Construir contrato e painel. A tabela aprovada em 2026-09-20 define cartão 1x/2x/3x sempre em D+30, com tarifas de 2,59% + R$ 0,35, 4,49% + R$ 0,35 e 5,44% + R$ 0,35; boleto em D+2 por R$ 2,39; e PIX na hora por 0,99%, todos com TPV exibido como grátis. Ela não substitui um ledger. Manter financeiro confirmado `UNAVAILABLE` ou `PARTIALLY_AVAILABLE` até API/exportação oficial fornecer bruto, taxa efetiva, líquido, parcela, status e data esperada/efetiva. Pedido Nuvemshop ou tarifa de painel não comprovam liquidação.

### D-005-A — defaults

- BRL; timezone de negócio `America/Sao_Paulo`; instantes técnicos UTC;
- dinheiro `numeric(19,4)`, sem `float`; `ROUND_HALF_UP` na fronteira;
- intervalos de negócio inclusivos; técnicos `[início,fim)`;
- sync 15 minutos; `STALE` após duas perdas ou 30 minutos;
- aging exclusivo: vencido, hoje, 1–7, 8–15, 16–30, 31–45, 46–60;
- parcial usa saldo aberto; cancelamento/estorno/chargeback geram compensação;
- folha CSV UTF-8 separada por `;`;
- RAW append-only, exceto redação LGPD autorizada e auditada.

## 4. Arquitetura e trust boundaries

```text
Usuária
  │
React/TypeScript em WebView2
  │ bindings Wails
majucau.exe
  │ Named Pipe + SID/DACL + protocolo versionado
majucau-worker.exe — NT SERVICE\MajucauWorker
  ├── DPAPI/ACL ─► vault
  ├── SCRAM/loopback ─► PostgreSQL dedicado
  └── TLS/OAuth/read-only ─► provedores
```

`majucau.exe` apresenta a UI, não persiste secrets e não acessa banco/APIs. O worker autentica e autoriza IPC, controla OAuth/tokens/sync, executa cálculos, migrations, jobs, backup/restore e auditoria. PostgreSQL aceita apenas loopback, usa SCRAM e migrations forward-only checksumadas sob advisory lock.

O relay OAuth é externo e mínimo: pareia `code/state` curto, nunca recebe token, client secret ou dado financeiro; exige hospedagem, observabilidade, limpeza, proteção contra replay e homologação Nuvemshop.

Regras de fronteira:

- todo input externo é não confiável;
- ação material é autorizada no worker;
- IPC possui versão, request ID, allowlist de método, limite e timeout;
- logs têm códigos/correlação, sem payload completo ou secret;
- falha preserva último dado válido;
- retry somente transitório, com backoff/jitter;
- cursor avança apenas após commit integral.

## 5. Dados, RAW, idempotência e lineage

Identidade externa normalizada:

```text
(connection_id, source_system, source_entity, source_id)
```

Cada versão acrescenta `payload_hash`; sincronização não sobrescreve payload anterior. Redação LGPD exige tombstone, hash anterior, motivo, ator, instante e auditoria.

Fluxo de ingestão:

1. criar lote;
2. buscar página com timeout/cancelamento;
3. normalizar identidade e hash canônico;
4. acrescentar RAW quando novo;
5. transformar idempotentemente;
6. gravar contribuição tipada;
7. atualizar contadores e cursor na mesma transação;
8. confirmar lote e publicar status sanitizado.

Lineage mínimo:

```text
card/linha → snapshot/cálculo + regra → contribuição tipada
→ entidade financeira → RAW → conexão/provedor/ID original
```

FK polimórfica não é integridade. Tabelas de contribuição possuem FKs reais; views de união são leitura.

## 6. Regras financeiras protegidas

```text
Total a Receber = B2C confirmado Nuvem + B2B confirmado Bling
opening(D) = closing(D-1)
closing(D) = opening(D) + inflows(D) - outflows(D) ± authorized_adjustments(D)
```

B2C espelhado no Bling é excluído. `UNCLASSIFIED` fica em revisão. D0 é provisório; D+1–D+60 são projetados; a série contém 61 dias contínuos. Quebra de continuidade é erro. O menor saldo inclui data, cenário e composição.

Aging usa `open_balance`, exclui cancelados/estornados e aloca exatamente uma faixa.

DRE, P&L, EBITDA, forecast, capital de giro e aplicação são calculados pelo motor Majucau. Demonstrativos prontos do Bling não são verdade contábil. Débitos = créditos; ativo = passivo + patrimônio líquido. Diferença não zero gera `DIVERGENT`.

## 7. Integrações e tela de credenciais

Bling: contas a receber/pagar, realizados, contas financeiras, caixa/banco, categorias, contatos, pedidos, compras e NF-e; OAuth externo; read-only; incremental; paginado; idempotente; resiliente a 429/timeout.

Nuvemshop: pedidos, clientes, status, parcelas/gateway quando disponíveis, cancelamento, reembolso e void. Financeiro de pedido é provisório até fonte oficial.

Nuvem Pago: adapter e UI permanecem visíveis mesmo indisponíveis. Nunca inventar fee, net ou settlement date.

Cada card da tela de integrações possui campos mascarados, labels, Conectar/Testar/Reconectar/Desconectar, conta/loja, scopes, estado, última tentativa/sucesso e erro sanitizado. Salvar não inicia sync sem confirmação.

Credencial trafega somente no Named Pipe autenticado e é cifrada pelo worker com DPAPI sob a identidade do serviço. O banco guarda somente `secret_ref`.

## 8. UX e acessibilidade

Todo componente funcional declara `helpText` central ou justificativa. Tooltip é texto curto, abre por hover/foco, fecha por Escape/blur, usa `aria-describedby`/`role=tooltip`, não contém ações e não esconde informação crítica.

Estados obrigatórios: loading, empty, error, stale, partial e confirmed. Validar teclado, foco, leitor de tela, contraste e zoom 200%. Preservar nomes oficiais; não criar card genérico “Caixa”. Menu: “Tesouraria e Projeção Financeira”.

## 9. Segurança e LGPD

- menor privilégio e autorização no worker;
- DPAPI + entropia persistida + ACL serviço/admin;
- Named Pipe com DACL mínima e SID;
- PostgreSQL loopback/SCRAM;
- secrets fora de argumentos, logs, banco, frontend, dump, Git e backup comum;
- fixtures reais minimizadas/sanitizadas;
- vulnerabilidade, secrets, SBOM e licenças no pipeline;
- Authenticode no pacote final;
- retenção e redação de RAW/backups auditáveis.

Single-user não elimina RBAC: ações administrativas/materialmente destrutivas continuam exigindo autorização.

## 10. Instalação, upgrade e recuperação

Instalação alvo: validar Windows/UAC/espaço; instalar WebView2 se ausente; instalar UI/worker/PostgreSQL; criar `%ProgramData%` e ACLs; gerar segredos e configurar loopback/SCRAM; iniciar banco/migrations; registrar serviços e dependências; criar atalhos; abrir first-run.

Máquinas de terceiros são ambiente não confiável e potencialmente divergente. O instalador deve fazer preflight antes da primeira mutação, manter journal de fases, usar escrita/rename atômicos, ser idempotente, retomar ou reparar execução interrompida e nunca interferir em PostgreSQL/serviços de terceiros. Cada falha conhecida possui detecção, código estável, rollback/retomada e instrução simples. Falha desconhecida bloqueia com segurança e gera support bundle sanitizado; “continuar mesmo assim” não pode contornar risco de perda, exposição ou schema parcial.

Upgrade: verificar assinatura/hash, backup pré-update, parar serviços, trocar binários, migrar e executar smoke. Rollback binário apenas com schema compatível; senão restaurar banco/roles/binários atomicamente.

Backup: `pg_dump -Fc`, manifesto, SHA-256, globals/roles, criptografia autenticada e chave DPAPI. Só é validado após restore real e checagem funcional.

Uninstall não apaga dados silenciosamente. Remoção material exige escolha explícita e recuperável.

A promessa verificável não é antecipar literalmente toda falha possível, e sim cobrir sistematicamente as classes conhecidas, testar os modos de maior impacto e garantir contenção, diagnóstico e recuperação para o desconhecido. A matriz executável fica em `docs/WINDOWS-INSTALLATION-RESILIENCE.md`.

## 11. Testes e gates

Automação: Go unidade/limites/overflow/IPC/security; SQL `sqlc` determinístico/parser/migration real; frontend typecheck/lint/Vitest/axe/build; contratos com OpenAPI e fixtures sanitizadas; `govulncheck`, npm audit, secret scan, SBOM/licenças.

Casos obrigatórios: duplicidade, idempotência, limites de datas, timezone, negativo, parcial, cancelamento, estorno, chargeback, vencidos, ausência de sync, mudança tardia de status, projeção, cenário, aplicação e arredondamento.

- G0 arquitetura: aprovado;
- G1 fundação: build, migration real, worker/pipe, desktop e CI;
- G2 integrações: OAuth, vault, RAW/idempotência e contratos reais;
- G3 tesouraria: reconciliação e casos aprovados;
- G4 executivo: dashboard, drill-down, estados e acessibilidade;
- G5 resultado: regras/oráculos contábeis e financeiros;
- G6 operação: backup/restore/update/logs/security;
- G7 produção: instalador assinado, VMs limpas e aceite.

Nenhum gate é aprovado porque uma tela existe. Vincular evidência a `REQUIREMENTS-TRACEABILITY.md`.

## 12. Pendências que não podem ser inventadas

1. apps/credenciais Bling e Nuvemshop, fora de chat/Git;
2. redirect Nuvemshop homologado e relay publicado;
3. confirmação PKCE ou aceite D-002-RISK;
4. fonte oficial Nuvem Pago completa;
5. payloads/fixtures e contratos reais sanitizados;
6. plano de contas e revisão DRE/P&L/EBITDA/Capital de Giro;
7. arquivo real de folha e total por competência;
8. certificado Authenticode;
9. VMs limpas Windows 10/11;
10. wireframe final aprovado.

## 13. Protocolo de continuidade

Antes de alterar: ler `AGENTS.md`, este contexto e a decisão; localizar requisito/evidência; preservar invariantes; registrar decisão material; implementar na menor superfície; testar positivos/negativos/regressão; atualizar rastreabilidade/operação; fazer revisão adversarial.

No handoff, informar objetivo, resultado, arquivos, comandos realmente executados, evidência, risco residual, rollback e próxima dependência. Nunca alegar prontidão sem prova reproduzível.
