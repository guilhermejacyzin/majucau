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
- nenhum arquivo real ou dado pessoal foi versionado.

O comando amplo `go test ./...` não é usado como gate neste checkout: os caches consolidados em `artifacts/cache/` e dependências locais contêm fontes Go auxiliares, que não pertencem ao módulo do produto. O CI e `scripts/verify.ps1` usam o conjunto explícito de pacotes acima.

## Estado dos gates

| Gate | Estado | Evidência atual | Falta para aceite |
|---|---|---|---|
| G0 — Arquitetura | Aprovado | decisões e artefatos versionados | nenhuma no escopo aprovado |
| G1 — Fundação | Parcial | testes Go, frontend, build e E2E opt-in versionado | executar E2E PostgreSQL/worker/desktop e CI em instalação limpa |
| G2 — Integrações | Parcial/bloqueado | adapters, DPAPI, contratos Bling, parser/persistência, IPC/UI e E2E opt-in do futuro Nuvem Pago | executar E2E em banco descartável, credencial/payload oficial, OAuth Nuvemshop e ledger Nuvem Pago |
| G3 — Tesouraria | Parcial | contrato D0–D+60 e menor saldo | saldo D-1 do Bling, persistência, cenários, lineage e reconciliação |
| G4 — Executivo | Parcial | shell T1–T15 e tooltips estruturais | dados reais, drill-down e comparação visual/funcional por tela |
| G5 — Resultado | Bloqueado | contratos e decisões documentados | DRE/P&L/EBITDA, folha e contabilidade próprios |
| G6 — Operação | Parcial | preflight, journal, diagnóstico e bundle local | backup/restore, update/rollback, assinatura e matriz Windows completa |
| G7 — Produção | Não iniciado | nenhum aceite em VM limpa | instalador assinado, Windows 10/11, primeiro uso, reboot, upgrade, uninstall e aceite final |

## Critério para mudar a conclusão

Só será possível declarar “pronto para produção” depois de G2–G7 produzirem as evidências correspondentes, incluindo fonte oficial ou decisão formal sobre os pontos bloqueados. Telas que abrem, compilação verde ou dados provisórios não substituem esses gates.

Próximo incremento técnico: executar o E2E PostgreSQL e o smoke visual do
fluxo no instalador, sem alterar a origem exclusiva Bling para recebimentos
realizados.

