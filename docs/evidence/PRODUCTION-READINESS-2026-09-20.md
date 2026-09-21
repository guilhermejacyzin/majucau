# Diagnóstico de prontidão para produção — 2026-09-20

## Resultado

O produto **ainda não está pronto para produção**. O número operacional usado no plano permanece em aproximadamente **40% de trabalho restante**, mas isso não equivale a 60% de aceite de produção: a matriz de rastreabilidade contabiliza somente 2 requisitos como `VERIFIED` de 47.

Não existe uma data responsável de lançamento enquanto os gates externos e os módulos ainda não implementados permanecerem sem evidência. O próximo marco verificável é fechar a fundação de dados e a integração Bling com payload/fixture sanitizado e persistência real.

## Evidência executada

- `go test . ./cmd/... ./database/... ./internal/...`: PASS.
- `go vet . ./cmd/... ./database/... ./internal/...`: PASS.
- `frontend/npm run typecheck`: PASS.
- `frontend/npm run lint`: PASS.
- `frontend/npm run test -- --run`: PASS — 3 arquivos, 19 testes.
- `frontend/npm run build`: PASS — bundle Vite produzido.
- `git diff --check`: PASS.
- Matriz atual: `VERIFIED 2`, `PARTIAL 16`, `BLOCKED 10`, `NOT_STARTED 14`, `DOCUMENTED 5`.

Incremento adicional validado nesta revisão:

- primeiro snapshot real do dashboard ligado por IPC/Wails ao PostgreSQL;
  cards de recebíveis, recebido no mês e pagamentos/obrigações só exibem
  métricas normalizadas, com estado `UNAVAILABLE` explícito quando a fonte
  ainda não existe; detalhes em
  `docs/evidence/DASHBOARD-SNAPSHOT-2026-09-21.md`;

- T3/T4 agora expõem linhas sanitizadas das tabelas normalizadas de recebíveis
  e obrigações no snapshot do dashboard, mantendo os filtros de origem
  (Nuvem Pago B2C projetado e Bling B2B classificado; obrigações Bling com
  saldo aberto) e sem transportar RAW para a UI;

- parser e importação determinística da pasta controlada `02_nuvem_pago/recebimentos_futuros`;
- contrato `nuvem_pago_recebimentos_futuros_csv_v1`, origem `NUVEM_PAGO` e status `PROJECTED`;
- preservação de datas, parcelas, bruto, taxas, juros, custos totais e líquido;
- rejeição de linhas fora de `Entrada`/`Venda`, datas inválidas, datas invertidas e duplicidades;
- `go test` e `go vet` do pacote `internal/integrations/nuvempago`: PASS;
- serviço de persistência transacional implementado para lote, RAW versionado
  e upsert idempotente em `receivables` (`B2C`/`PROJECTED`);
- preview/importação de futuros ligado ao worker IPC, Wails e painel de
  Integrações, com resposta sanitizada e fail-closed sem configuração;
- teste E2E opt-in criado para PostgreSQL: duas linhas projetadas, uma rejeição,
  reexecução idempotente, RAW corrente e zero escrita em `receipts`;
- execução real PASS do E2E Nuvem Pago em PostgreSQL 18.3 descartável e
  execução real PASS do E2E Bling no mesmo cluster isolado;
- smoke automatizado do bundle portátil Windows x64 PASS: hashes, preflight,
  WebView2, porta, health do worker em console e diagnóstico sanitizado; o
  preflight corretamente reportou `BLOCKED` por elevação/reboot da máquina de
  validação;
- fonte do instalador NSIS agora executa o `installer-helper`/diagnostics antes
  de escrever no computador, aborta em preflight bloqueado e instala o helper;
  `scripts/test-installer-nsis.ps1`: PASS;
- GitHub Actions run 60 PASS: Go/frontend/Wails/worker/helper, `makensis`,
  guard NSIS e upload do artefato `majucau-g1-unsigned`.
- GitHub Actions runs 76 e 77: toda a cadeia até o NSIS passou; o smoke de
  instalação/desinstalação encerrou com código `2` (`BLOCKED`) no preflight.
  O run 77 registrou explicitamente `PREFLIGHT_DIAGNOSTIC_NOT_FOUND`; isso
  mantém o gate fechado até uma VM Windows limpa com elevação real e captura
  do diagnóstico sanitizado. Detalhes em
  `docs/evidence/NSIS-INSTALL-SMOKE-2026-09-21.md`.
- GitHub Actions run 80: a cadeia de build continuou passando e o diagnóstico
  determinístico identificou `OS_UNSUPPORTED`. O runner hospedado é Windows
  Server, mas o produto aceita somente Windows 10/11 workstation x64; o CI
  passou a registrar esse ambiente como inelegível para smoke desktop sem
  enfraquecer a política. O gate continua exigindo VM Windows workstation
  limpa.
- GitHub Actions runs 81–83: `success` após a classificação explícita do
  runner Server. Isso confirma build, testes, contrato NSIS e artefato, mas
  não substitui a evidência de instalação/desinstalação em Windows 10/11
  workstation.
- GitHub Actions run 86: `success` após o incremento das tabelas T3/T4,
  cobrindo novamente Go/frontend, vulnerabilidades, migrations, sqlc, build
  Wails/worker/helper e contrato NSIS; o smoke desktop continua exigindo
  workstation Windows 10/11 limpa.
- GitHub Actions run 92: `success` após a fundação do manifesto de pacote e a
  correção específica do runner Windows. O CI executou os testes do validador
  com checksum, schema, duplicidade, traversal e caminhos protegidos, além de
  repetir a cadeia completa de build e contrato NSIS. Isso comprova a fatia
  determinística do manifesto, mas não backup/restore ou update/rollback reais.
- nenhum arquivo real ou dado pessoal foi versionado.

### Incremento de 2026-09-21

- fluxo Bling de **Reconectar** e **Desconectar** habilitado na tela de
  Integrações;
- desconexão remove somente tokens/escopos/expirações protegidos, preserva
  Client ID/Redirect/segredo protegido e mantém todos os dados financeiros
  importados;
- estado persistido como `NOT_CONFIGURED`, sessões OAuth pendentes encerradas
  e resposta sanitizada no worker/Wails;
- testes Go, vet, TypeScript, lint, Vitest (23 testes) e build Vite passaram;
- evidência detalhada em `docs/evidence/BLING-DISCONNECT-2026-09-21.md`.

- validador de manifesto de release/backup implementado sem efeitos colaterais;
  checksums de migrations e artefatos, janela de schema e rejeição de entradas
  inseguras documentados em `docs/evidence/RELEASE-MANIFEST-2026-09-21.md`;
  evidência publicada no run 92.

- pacote inicial de backup PostgreSQL em `*.mjbk`: `pg_dump`/`pg_dumpall` sem
  senha em argumentos, payloads AES-GCM com chave derivada por Argon2id,
  manifesto com hashes e política explícita de tokens excluídos, e verificador
  contra schema incompatível, duplicidades e entradas inesperadas;
  detalhes em `docs/evidence/BACKUP-PACKAGE-2026-09-21.md`;
- o pacote ainda não restaura nem altera um cluster: integração com worker/UI,
  restore controlado, backup do estado anterior e rollback permanecem gates
  abertos;
- o run 98 falhou corretamente ao detectar que o binário oficial do WebView2
  mudou; o pin foi atualizado para o SHA-256 observado no download oficial e o
  run 99 está em execução para confirmar o instalador.

Esses incrementos melhoram a cobertura de INT-05 e de OPS-02/OPS-03, mas não
alteram a conclusão de produção: ainda faltam fonte/payload oficial, regras
financeiras, resultados, backup/restore, assinatura e a matriz real de
instalação Windows.

O comando amplo `go test ./...` não é usado como gate neste checkout: os caches consolidados em `artifacts/cache/` e dependências locais contêm fontes Go auxiliares, que não pertencem ao módulo do produto. O CI e `scripts/verify.ps1` usam o conjunto explícito de pacotes acima.

## Estado dos gates

| Gate | Estado | Evidência atual | Falta para aceite |
|---|---|---|---|
| G0 — Arquitetura | Aprovado | decisões e artefatos versionados | nenhuma no escopo aprovado |
| G1 — Fundação | Parcial | testes Go, frontend, build, E2E PostgreSQL real e CI Windows completo até o NSIS; smoke real bloqueado pelo preflight | E2E desktop em instalação limpa e validação operacional pós-instalação |
| G2 — Integrações | Parcial/bloqueado | adapters, DPAPI, contratos Bling, parser/persistência, IPC/UI e E2E real Bling/Nuvem Pago | credencial/payload oficial, OAuth Nuvemshop e ledger Nuvem Pago |
| G3 — Tesouraria | Parcial | contrato D0–D+60 e menor saldo | saldo D-1 do Bling, persistência, cenários, lineage e reconciliação |
| G4 — Executivo | Parcial | shell T1–T15 e tooltips estruturais | dados reais, drill-down e comparação visual/funcional por tela |
| G5 — Resultado | Bloqueado | contratos e decisões documentados | DRE/P&L/EBITDA, folha e contabilidade próprios |
| G6 — Operação | Parcial | preflight, journal, diagnóstico, bundle local, smoke automatizado, guard NSIS e CI com classificação explícita de runner Server | smoke em VM Windows workstation, backup/restore, update/rollback, assinatura e matriz Windows completa |
| G7 — Produção | Parcial inicial | CI produz instalador NSIS x64 unsigned e bloqueia corretamente preflight não aceito | assinatura, Windows 10/11 limpo, primeiro uso, reboot, upgrade, rollback, uninstall e aceite final |

## Critério para mudar a conclusão

Só será possível declarar “pronto para produção” depois de G2–G7 produzirem as evidências correspondentes, incluindo fonte oficial ou decisão formal sobre os pontos bloqueados. Telas que abrem, compilação verde ou dados provisórios não substituem esses gates.

Reavaliação de planejamento em 2026-09-21: aproximadamente **36–40% do
trabalho ainda falta**. O intervalo é deliberado: 36% é a estimativa de
execução do backlog técnico, enquanto 40% é a margem conservadora até os gates
de aceite externo. Isso não significa 60% de aceite de produção; a matriz
continua com somente 2 requisitos `VERIFIED` de 47.

Próximo incremento técnico: integrar o pacote de backup ao worker sem alterar a
origem exclusiva Bling para recebimentos realizados, implementar restore
controlado em staging e obter a evidência de instalação em VM Windows 10/11
limpa. Assinatura, update/rollback e módulos financeiros G5 continuam gates
posteriores.
