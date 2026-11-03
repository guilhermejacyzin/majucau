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
- CI do commit documental `4a4eabf` — o verificador do backlog passou; a etapa
  de checagem da mensagem falhou porque os campos obrigatórios precisam estar
  no corpo, não no título. Por essa falha, os jobs Windows e PostgreSQL foram
  pulados. A próxima mensagem foi ajustada para deixar os campos no corpo; a
  CI desse novo commit ainda precisa terminar.

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
faixas e saldos com o relatório Fluxo de Caixa do Bling. A permissão não foi
alterada nesta execução: o controle do navegador não conseguiu confirmar a
página de cadastro ativa. O quadro não confirma saldo inicial, float ou
reserva.
