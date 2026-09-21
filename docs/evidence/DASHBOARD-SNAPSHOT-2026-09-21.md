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

As tabelas centrais dos mockups T3 e T4 também recebem linhas sanitizadas do
ledger normalizado, limitadas às 50 mais próximas e sem transportar o RAW:

- T3 mostra cliente, origem, vencimento, bruto, líquido e status de cada
  recebível elegível (Nuvem Pago B2C projetado ou Bling B2B classificado);
- T4 mostra fornecedor, documento, vencimento, valor em aberto, categoria e
  status de cada obrigação Bling com saldo aberto.

O filtro da consulta é o mesmo da métrica agregada. Portanto, uma linha
`UNCLASSIFIED` do Bling não entra silenciosamente no total confirmado.

Ausência de dados permanece `UNAVAILABLE` e aparece como `—`; não há valores
simulados ou zero fabricado para esconder uma fonte ausente. Os recebíveis
projetados B2C usam exclusivamente `NUVEM_PAGO`/`PROJECTED`, enquanto os
recebimentos realizados consultam exclusivamente `BLING`/`CONFIRMED`.

## Verificação

- `go test . ./cmd/... ./database/... ./internal/...`: PASS;
- `go vet . ./cmd/... ./database/... ./internal/...`: PASS;
- `frontend npm run typecheck`: PASS;
- `frontend npm run lint`: PASS;
- `frontend npm run test:run`: PASS — 3 arquivos, 22 testes;
- `frontend npm run build`: PASS;
- `git diff --check`: PASS.
- GitHub Actions run 86: PASS em Windows, incluindo testes, vulnerabilidade,
  geração SQL, build Wails, worker, preflight e contrato NSIS; o smoke segue
  corretamente condicionado a workstation Windows 10/11, não a Windows Server.

## Limite conhecido

O snapshot é a primeira fatia de dados reais e não declara os módulos T1–T15
concluídos. Saldo bancário, fluxo diário, DRE, conciliação, estoque,
produção, compras e forecast continuam indisponíveis até que suas fontes,
regras e persistências sejam homologadas. O mapeamento bruto da API Bling
continua sem normalização especulativa.
