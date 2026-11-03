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
- Build desktop Wails `wails build -clean -trimpath` — passou em cópia isolada
  de `artifacts/verification/majucau-wails-20261103`.
- Compilação Windows do worker (`go build ./cmd/worker`) e do instalador
  auxiliar (`go build ./cmd/installer-helper`) — passou na mesma cópia.
- CI do commit `1d01742` — o verificador do backlog passou; a etapa de checagem
  da mensagem falhou e os jobs Windows e PostgreSQL foram pulados. A inspeção
  local confirmou que os campos ficaram no título multilinha e o corpo do
  commit estava vazio. Portanto, a tentativa anterior de corrigir a mensagem
  não funcionou. O próximo commit deve usar um título curto e colocar os campos
  obrigatórios no corpo.

Os comandos frontend foram executados na cópia isolada
`artifacts/verification/frontend-tests-20261103`, porque o `frontend/node_modules`
local é um junction para uma pasta de verificação incompleta. A build Wails e
os binários auxiliares também foram compilados fora do checkout de trabalho;
nenhuma dependência compartilhada foi apagada ou substituída. A versão isolada
não foi aberta contra o banco/worker do protótipo e a sincronização diária
completa ainda não foi executada após esta mudança.

## Limites

O dado continua Parcial até completar a permissão somente de leitura de Contas
Financeiras, executar a sincronização diária das 09:30 e comparar todas as
faixas e saldos com o relatório Fluxo de Caixa do Bling. Em 03/11/2026, a
responsável aprovou habilitar somente leitura para Contas Financeiras. A
permissão ainda não foi alterada: esta sessão não tem controle operacional da
tela de cadastro do aplicativo Bling para aplicar a mudança. Depois de habilitar
a permissão, será necessário autorizar o Majucau novamente e conferir a carga.
O quadro não confirma saldo inicial, float ou reserva.
