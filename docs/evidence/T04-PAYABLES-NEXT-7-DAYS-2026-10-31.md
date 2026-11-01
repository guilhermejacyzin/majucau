# T04 — métrica de vencimentos dos próximos sete dias — 31/10/2026

## Entrega desta fatia

O card existente `VENCIMENTOS 7 DIAS` da tela Contas a Pagar agora recebe uma
métrica calculada pelo backend, sem alteração de posição, tamanho, texto ou
estilo da tela.

A Visão Executiva continua recebendo `A PAGAR` do mesmo snapshot usado por
Contas a Pagar (`TOTAL A PAGAR`). Não existe consulta paralela da T01 ao Bling.
Quando faltar um resultado na tela de origem, ele deve ficar pendente ou
indisponível e a responsável deve ser sinalizada; uma busca direta no Bling só
poderá ser feita após autorização explícita para aquele dado.

O cálculo usa a data de negócio do Majucau e soma somente o `open_balance`
positivo de títulos com vencimento entre amanhã e o sétimo dia seguinte,
inclusive. O intervalo técnico é `[data de negócio + 1, data de negócio + 8)`;
portanto não inclui títulos vencidos, que vencem hoje nem os que vencem no
oitavo dia. Os estados `CANCELLED`, `CANCELED`, `REFUNDED`, `REVERSED` e
`ESTORNADO` são excluídos, seguindo a lista já usada pela regra de aging
aprovada. A métrica vem da mesma carga diária normalizada do Bling usada para
Contas a Pagar; até essa carga estar atualizada, o backend a marca indisponível
e a tela não apresenta valor como confirmado.

Arquivos alterados nesta fatia:

- `internal/dashboard/stream.go` e `internal/application/contracts.go`: consulta,
  estado de disponibilidade e campo `payables_due_next_7_days` no snapshot.
- `frontend/src/bootstrap.ts` e `frontend/src/mockupDashboard.tsx`: validação do
  contrato e vínculo ao card já aprovado.
- `internal/dashboard/stream_test.go` e
  `internal/dashboard/stream_integration_test.go`: cobertura de contrato, limites
  de data, carga atual e estados cancelados/estornados.
- `frontend/src/App.test.tsx` e `frontend/src/bootstrap.test.ts`: cobertura do
  contrato de interface, do mesmo total em T01 e T04 e da exibição no card
  existente de vencimentos.

## Evidência local

- `go test -count=1 ./internal/dashboard`: passou. O teste dedicado de PostgreSQL
  foi ignorado porque `MAJUCAU_TEST_DATABASE_URL` não está configurada nesta
  máquina; portanto a consulta SQL ainda precisa ser confirmada em PostgreSQL.
- `npm run typecheck`: passou.
- `npm run test:run`: passou, 39 testes em 3 arquivos.
- `scripts/verify.ps1`: passou por completo após corrigir uma dependência npm
  indireta; Go, `go vet`, verificações de migrations/SQLC/segredos/instalador,
  lint, typecheck, 39 testes, build e `npm audit` finalizaram com sucesso.
- Um `npm ci` limpo foi executado numa cópia temporária centralizada em
  `artifacts/verification/`, confirmando `minimatch@11.0.1`; lint, typecheck,
  39 testes, build e `npm audit` também passaram com a árvore de dependências
  do lockfile atualizado.
- O `npm audit` encontrou zero vulnerabilidades. O `govulncheck` encontrou zero
  vulnerabilidades alcançáveis pelo código; reportou um módulo requerido que
  não é chamado pelos símbolos analisados, conforme a saída do verificador.
- A CI [41261894619](https://github.com/guilhermejacyzin/majucau/actions/runs/41261894619)
  passou nos quatro trabalhos: `detect-changes`, `verify-backlog`,
  `verify-postgres` e `verify-windows`. O teste de integração SQL foi executado
  em PostgreSQL descartável e passou.
- O teste PostgreSQL dedicado não executou localmente porque
  `MAJUCAU_TEST_DATABASE_URL` não está configurada; a execução aprovada de CI
  acima cobre essa validação. A homologação com a conta real do Bling ainda está
  pendente.

## Limite do aceite

Esta é uma fatia funcional de T04, não a conclusão de T04 nem de BK-058/BK-059.
O backlog permanece sem alteração de status/percentual até todos os critérios
dessas tarefas serem comprovados. A consulta SQL precisa passar no PostgreSQL
descartável da CI e os valores ainda precisam ser homologados contra a conta
real autorizada do Bling.
