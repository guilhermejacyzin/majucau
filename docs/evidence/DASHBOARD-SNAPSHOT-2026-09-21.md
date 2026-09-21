# Snapshot real do dashboard — 2026-09-21

## Incremento entregue

O worker agora expõe `dashboard.snapshot` por IPC/Wails. A consulta lê somente
as tabelas normalizadas `receivables`, `receipts`, `payables` e `payments` e
retorna valores monetários como texto decimal, preservando a precisão do
PostgreSQL até a formatação visual.

O frontend consulta o snapshot de forma independente do bootstrap e preenche
somente os cards para os quais existe métrica normalizada:

- `A RECEBER`, `TOTAL A RECEBER` e recebíveis em atraso;
- `RECEBIDO NO MÊS`;
- `A PAGAR`, vencimentos de hoje e vencidos, quando o ledger Bling estiver
  normalizado;
- `PAGO NO MÊS`, quando pagamentos Bling estiverem normalizados.

Ausência de dados permanece `UNAVAILABLE` e aparece como `—`; não há valores
simulados ou zero fabricado para esconder uma fonte ausente. Os recebíveis
projetados B2C usam exclusivamente `NUVEM_PAGO`/`PROJECTED`, enquanto os
recebimentos realizados consultam exclusivamente `BLING`/`CONFIRMED`.

## Verificação

- `go test . ./cmd/... ./database/... ./internal/...`: PASS;
- `go vet . ./cmd/... ./database/... ./internal/...`: PASS;
- `frontend npm run typecheck`: PASS;
- `frontend npm run lint`: PASS;
- `frontend npm run test:run`: PASS — 3 arquivos, 21 testes;
- `frontend npm run build`: PASS;
- `git diff --check`: PASS.

## Limite conhecido

O snapshot é a primeira fatia de dados reais e não declara os módulos T1–T15
concluídos. Saldo bancário, fluxo diário, DRE, conciliação, estoque,
produção, compras e forecast continuam indisponíveis até que suas fontes,
regras e persistências sejam homologadas. O mapeamento bruto da API Bling
continua sem normalização especulativa.

