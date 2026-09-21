# Majucau Financial Intelligence

Aplicativo desktop de inteligência financeira para uso individual em Windows 10/11 x64. O produto consolida tesouraria, recebíveis, obrigações, projeções, resultado e alertas em uma interface executiva, preservando a origem e a rastreabilidade de cada valor.

> **Estado atual:** G0 aprovado em 2026-08-16; fundação G1 em implementação. Este repositório ainda não representa um release aprovado para produção.

## O que o produto entrega

- visão executiva com os cards financeiros oficiais do handoff;
- projeção diária contínua de D0 a D+60;
- recebíveis B2B classificados de forma conservadora a partir do Bling;
- pedidos Nuvemshop operacionais/provisórios;
- tela de integrações editável, com configuração e OAuth/teste do Bling no worker; Nuvemshop e Nuvem Pago continuam condicionados às evidências oficiais aprovadas;
- trilha de auditoria e lineage até o registro RAW de origem;
- estados explícitos para dado confirmado, provisório, parcial, desatualizado ou indisponível;
- aplicação desktop e worker nativos, sem Docker ou Node.js na máquina do usuário;
- PostgreSQL local dedicado e instalador Windows NSIS.

O tarifário Nuvem Pago aprovado está registrado em `docs/source/Nuvem-Pago-Taxas-2026-09-20.md`: cartão 1x/2x/3x sempre D+30 (2,59% + R$ 0,35; 4,49% + R$ 0,35; 5,44% + R$ 0,35), boleto D+2 (R$ 2,39) e PIX na hora (0,99%). A tabela anterior de cartão D+2/D+14 foi substituída e não é usada. O B2C financeiro confirmado continua desabilitado enquanto não existir fonte oficial do ledger para valor bruto, taxa efetiva, líquido, status e data de liquidação. O sistema nunca substitui uma fonte ausente por uma estimativa silenciosa.

A screen de Integrações exibe esse tarifário como configuração/projeção identificada no card Nuvem Pago. O motor Go calcula a tarifa com decimal exato e `ROUND_HALF_UP`; nenhum desses valores é apresentado como recebimento confirmado.

## Stack fixada

| Camada | Tecnologia |
|---|---|
| Desktop | Wails v2.13.0 |
| Interface | React 19 + TypeScript + Vite |
| Backend e worker | Go 1.26.6 |
| Banco local | PostgreSQL x64 |
| Acesso SQL | `pgx/v5` + código gerado por `sqlc` |
| Segredos | Windows DPAPI sob a identidade do worker |
| IPC local | Windows Named Pipe autenticado e versionado |
| Instalador | NSIS x64 |
| CI | GitHub Actions em Windows |

## Arquitetura resumida

```text
React/TypeScript (WebView2)
          │ bindings Wails, sem secrets persistidos
          ▼
majucau.exe
          │ Named Pipe local, protocolo versionado e SID autorizado
          ▼
majucau-worker.exe (Windows Service)
   ├── integrações read-only
   ├── regras financeiras determinísticas
   ├── vault DPAPI
   ├── migrations/jobs/backup
   └── pgx/sqlc ──► PostgreSQL local dedicado
```

O frontend nunca acessa PostgreSQL nem APIs externas diretamente. O worker é a fronteira de autorização, persistência, credenciais, sincronização e auditoria.

## Decisões G0 aprovadas

- **D-001-A:** B2B conservador; ambiguidades ficam `UNCLASSIFIED` e fora do total confirmado.
- **D-002-A:** OAuth Nuvemshop por relay HTTPS mínimo, condicionado à homologação e ao tratamento de PKCE.
- **D-003-A:** fórmula Majucau v1 para Valor Máximo para Aplicação.
- **D-004-A:** adapter Nuvem Pago preparado, mas B2C financeiro confirmado diferido até fonte oficial.
- **D-005-A:** BRL, `America/Sao_Paulo`, decimal exato com quatro casas, `ROUND_HALF_UP`, aging exclusivo e sync padrão de 15 minutos.

Leia [PROJECT-CONTEXT.md](PROJECT-CONTEXT.md) antes de implementar ou revisar qualquer módulo. As regras de engenharia obrigatórias estão em [AGENTS.md](AGENTS.md).

## Estrutura do repositório

```text
cmd/worker/              processo/serviço Windows de background
cmd/installer-helper/    preflight read-only usado pelo instalador
internal/application/    contratos de casos de uso
internal/domain/         dinheiro, identidades e estados
internal/financial/      regras financeiras versionadas
internal/ipc/            protocolo e Named Pipe
internal/installer/      política e probes de compatibilidade da máquina
internal/security/       DPAPI, vault e sanitização
database/migrations/     migrations PostgreSQL forward-only
database/queries/        SQL fonte usado pelo sqlc
database/gen/            acesso Go gerado pelo sqlc
frontend/                interface React/TypeScript e testes
build/windows/           manifesto e instalador NSIS
scripts/                 bootstrap, validação e build
.github/workflows/       validação contínua
```

## Desenvolvimento local

Pré-requisitos: Windows 10/11 x64, Go 1.26.6, Node.js 24.x, Wails CLI 2.13.0, `sqlc` 1.31.1, WebView2 Runtime e PostgreSQL descartável para testes reais de migration.

```powershell
./scripts/bootstrap.ps1
./scripts/verify.ps1
./scripts/build-windows.ps1
./scripts/build-windows.ps1 -Installer
```

Os executáveis e instaladores finais são consolidados em `artifacts/windows/`, com o inventário em `artifacts/MANIFEST.txt`. O pipeline também gera `majucau-windows-x64-portable.zip` com hashes para validação local; esse bundle não é o instalador de produção. Os caches de Go/npm ficam em `artifacts/cache/`; após a consolidação, as saídas intermediárias (`build/bin`, `frontend/dist` e `tmp`) são removidas por padrão. Use `-KeepSourceOutputs` somente para diagnóstico local.

O script de validação não transforma automaticamente uma checagem externa em aprovada. PostgreSQL real, assinatura Authenticode, VM limpa e backup/restore exigem evidências próprias.

## Regras financeiras protegidas

- dinheiro nunca usa `float`;
- B2C futuro vem somente de fonte oficial Nuvem/Nuvem Pago;
- B2B futuro vem somente do Bling após classificação aprovada;
- B2C do Bling não pode ser somado novamente;
- total a receber = B2C Nuvem + B2B Bling;
- saldo inicial do dia = saldo final conciliado do dia anterior;
- saldo final = saldo inicial + entradas - saídas ± ajustes autorizados;
- D0 é provisório; D+1 a D+60 são projetados e contínuos;
- correções criam nova versão, sem reescrever snapshots aprovados;
- cada valor material deve chegar ao RAW e à fonte original por lineage tipado.

Qualquer mudança nessas regras exige versão, regressão financeira e aprovação funcional registrada.

## Integrações e credenciais

A tela **Configurações > Integrações** é a área editável para preencher os dados públicos e secrets das APIs de Bling, Nuvemshop e, quando houver fonte oficial, Nuvem Pago. O Bling permite salvar Client ID, Redirect URI e Client Secret pelo IPC autenticado, iniciar OAuth no navegador externo, acompanhar o callback loopback, testar uma leitura mínima e abrir uma sincronização filtrada explícita com tipo de data e período; o secret e os tokens vão para DPAPI e os metadados públicos para PostgreSQL. A Nuvemshop também permite salvar App ID, Redirect URI HTTPS e Client Secret pelo mesmo caminho protegido; OAuth e sincronização continuam bloqueados até o relay HTTPS/PKCE ser homologado. A sincronização só é executada depois que a pessoa informa um filtro e o Bling está conectado. Para a pessoa usuária, a experiência é normal de aplicativo — editar, salvar, conectar no navegador externo, acompanhar o estado e executar um período escolhido — sem o React persistir ou executar secrets.

- secrets e tokens não entram no frontend, Git, logs ou banco em texto aberto;
- a UI envia a credencial ao worker por IPC local autenticado;
- somente o worker persiste o segredo no vault DPAPI;
- OAuth abre o navegador externo;
- integrações do MVP são somente leitura;
- falha de sincronização preserva o último dado válido e nunca vira zero.

Não coloque credenciais reais em issues, commits, fixtures, `.env`, documentação ou mensagens de chat.

## Qualidade e gates

| Gate | Critério resumido | Estado |
|---|---|---|
| G0 | arquitetura e decisões | Aprovado |
| G1 | build, banco, worker, desktop e CI | Em execução |
| G2 | OAuth, credenciais, RAW e idempotência reais | Bling OAuth implementado; credencial real, RAW/idempotência e reconciliação ainda pendentes |
| G3 | tesouraria e reconciliações | Pendente |
| G4 | dashboard, tooltips e drill-down | Pendente |
| G5 | DRE/P&L/EBITDA e planejamento | Pendente |
| G6 | backup, restore, update, logs e segurança | Pendente |
| G7 | instalador assinado, VMs limpas e aceite | Pendente |

Os requisitos e as evidências esperadas estão em [REQUIREMENTS-TRACEABILITY.md](REQUIREMENTS-TRACEABILITY.md). Não use “pronto”, “produção” ou “validado” sem a execução do gate correspondente.

## Empacotamento Windows

O destino é um instalador x64 simples: a usuária instala, abre pelo atalho e começa o first-run sem terminal. O pacote final deverá incluir UI, worker, PostgreSQL dedicado, migrations e WebView2 quando necessário; criar contas/ACLs; sobreviver a reboot; preservar dados em upgrade/uninstall; e oferecer backup/restore testado.

O `installer-helper.exe` também oferece `diagnostics --output CAMINHO\diagnostico.zip`, que exporta somente o contrato de preflight e instruções sanitizadas. O comando pode retornar `BLOCKED` e ainda gerar o diagnóstico; ele não coleta segredos, tokens, DSNs, payloads ou caminhos completos.

O fluxo será defensivo para máquinas de terceiros: preflight antes de alterar o sistema, operações idempotentes, journal de fases, repair/retomada, rollback seguro, proteção contra instalações concorrentes e diagnóstico sanitizado. Cenários desconhecidos devem bloquear sem corromper o estado e fornecer um código acionável; não existe opção genérica para ignorar falhas críticas.

Builds G1 são deliberadamente não assinados. O G7 exige certificado Authenticode, verificação de hash/assinatura e smoke tests em Windows 10 e 11 limpos.

## Documentação principal

- [PROJECT-CONTEXT.md](PROJECT-CONTEXT.md) — contexto profundo e continuidade;
- [docs/REALINHAMENTO-2026-08-16.md](docs/REALINHAMENTO-2026-08-16.md) — análise crítica das novas referências e plano de realinhamento;
- [docs/FUNCTIONAL-DECOMPOSITION-v1.md](docs/FUNCTIONAL-DECOMPOSITION-v1.md) — decomposição máxima de módulos, funções, regras, dados e aceite;
- [ADR-001-ARQUITETURA-E-EMPACOTAMENTO.md](ADR-001-ARQUITETURA-E-EMPACOTAMENTO.md) — arquitetura Windows;
- [ERD-v1.md](ERD-v1.md) — entidades e relações;
- [DATA-DICTIONARY-v1.md](DATA-DICTIONARY-v1.md) — semântica dos dados;
- [API-INTEGRATION-MAP.md](API-INTEGRATION-MAP.md) — integrações e lacunas oficiais;
- [G0-DECISION-REGISTER.md](G0-DECISION-REGISTER.md) — decisões aprovadas;
- [IMPLEMENTATION-PLAN.md](IMPLEMENTATION-PLAN.md) — fases e gates;
- [UX-TOOLTIP-CATALOG.md](UX-TOOLTIP-CATALOG.md) — explicações acessíveis;
- [REQUIREMENTS-TRACEABILITY.md](REQUIREMENTS-TRACEABILITY.md) — requisito → artefato → evidência.
- [docs/evidence/G1-VALIDATION-2026-08-16.md](docs/evidence/G1-VALIDATION-2026-08-16.md) — comandos executados, resultados e riscos residuais da fundação G1.
- [docs/WINDOWS-INSTALLATION-RESILIENCE.md](docs/WINDOWS-INSTALLATION-RESILIENCE.md) — matriz de falhas, recuperação e aceite em máquinas de terceiros.
- [docs/evidence/INSTALLATION-RESILIENCE-2026-08-16.md](docs/evidence/INSTALLATION-RESILIENCE-2026-08-16.md) — execução local do preflight, build e runtime desktop/worker.
- [docs/evidence/BLING-OAUTH-2026-09-20.md](docs/evidence/BLING-OAUTH-2026-09-20.md) — sessão OAuth desktop, callback loopback, teste mínimo e evidências automatizadas.
- [docs/evidence/BLING-RAW-SOURCES-2026-09-20.md](docs/evidence/BLING-RAW-SOURCES-2026-09-20.md) — recursos `contas/receber` e `contas/pagar` preservados em RAW idempotente.
- [docs/evidence/BLING-SYNC-2026-09-20.md](docs/evidence/BLING-SYNC-2026-09-20.md) — contrato IPC de sincronização filtrada, cursor transacional e limites do gate.
- [docs/evidence/BUILD-ARTIFACTS-2026-09-20.md](docs/evidence/BUILD-ARTIFACTS-2026-09-20.md) — build Windows x64 validado e caches/outputs consolidados em `artifacts/`.

## Distribuição

O repositório é privado. Não redistribua código, documentação, dados, fixtures, credenciais ou builds sem autorização da responsável pelo produto. Licenças de terceiros e SBOM serão gerados e revisados antes do gate de produção.
