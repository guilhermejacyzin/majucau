# T02 — quadro Semana Crítica ligado ao Fluxo de Caixa — 03/11/2026

## Mudança

O quadro “Semana Crítica” da T02 estava sempre mostrando traços, mesmo quando
o snapshot já continha as faixas semanais obtidas pela composição da API oficial
do Bling. Agora o backend escolhe, entre as faixas disponíveis, aquela com o
maior total de saídas e envia o período, as saídas e as entradas pelo contrato
da T02. A T01 continua usando a mesma composição.

A tela aprovada não foi redesenhada. O quadro passa a preencher os campos
existentes de período, saídas e entradas e sinaliza os valores como “Parcial”
quando a fonte estiver incompleta. Reserva mínima e posição de caixa continuam
sem valor, pois não foram confirmadas por uma fonte financeira; nenhum saldo
foi estimado.

## Validação

- `go test ./internal/dashboard -count=1` — passou, incluindo seleção da maior
  saída, preservação de estado parcial, desempate estável e rejeição de valor
  inválido.
- TypeScript (`tsc --noEmit`) — passou.
- ESLint (`eslint . --max-warnings=0`) — passou.
- Suíte da interface — 70 testes passaram, incluindo exibição na T02 de período,
  entradas, saídas parciais e ausência de saldo/reserva não comprovados.
- Build Vite de produção (`tsc && vite build`) — passou.

Os comandos frontend foram executados na cópia isolada
`artifacts/verification/frontend-tests-20261103`, porque o `frontend/node_modules`
local é um junction para uma pasta de verificação incompleta. Não apaguei nem
substituí esse destino compartilhado. A build desktop Wails e a sincronização
diária completa ainda não foram executadas após esta mudança.

## Limites

O dado continua Parcial até completar a permissão somente de leitura de Contas
Financeiras, executar a sincronização diária das 09:30 e comparar todas as
faixas e saldos com o relatório Fluxo de Caixa do Bling. O quadro não confirma
saldo inicial, float ou reserva.
