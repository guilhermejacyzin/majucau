# Formatação exata de valores em reais no protótipo — 01/11/2026

## Alteração

Os valores textuais com precisão de até `numeric(19,4)` agora são arredondados
com `BigInt` para duas casas apenas ao exibi-los como reais. Isso segue a regra
`ROUND_HALF_UP` do dicionário financeiro e evita a aproximação binária de
`Number` no navegador. O mesmo formatador é usado nos cards, linhas de Contas a
Receber e Contas a Pagar e indicadores CAPEX.

O desenho das telas e os cálculos que produzem os valores não foram alterados.
Valores ausentes ou fora do contrato continuam apresentados como `—`.

## Verificação local

- `npm run test -- --reporter=dot`: 45 testes passaram em 5 arquivos, incluindo
  metades positivas/negativas, valores grandes, zero arredondado e entradas
  inválidas.
- `npm run lint`: passou sem avisos.
- `npm run typecheck`: passou.

Essa evidência comprova somente a formatação no frontend. Não comprova uma
fonte real, conciliação dos valores nem a homologação financeira ponta a ponta.

## Limpeza desta execução

As cinco prévias PNG e o arquivo de inspeção temporários da planilha foram
removidos de `artifacts/verification/backlog-v2` depois da conferência. Também
foram removidas duas cópias antigas de `frontend-dist-*` e a cópia isolada da
cache em `artifacts/verification/artifacts/cache/vite`. O relatório histórico
`artifacts/verification/CLEANUP-2026-10-26.md` foi preservado. A cache ativa do
Vite em `artifacts/cache/vite` e `frontend/dist` foram mantidos porque o
protótipo de desenvolvimento continua aberto; removê-los durante a sessão
poderia interromper a recarga ou a retomada do app.
