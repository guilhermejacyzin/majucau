# Remoção dos indicadores de validade da tela Estoque — 02/11/2026

## Solicitação

Remover da tela **09 — Estoque**, conforme as capturas enviadas:

- card `ENVELHECENDO`;
- card `VALIDADE`;
- alerta `Lotes com validade próxima`.

## Resultado

Os três itens foram retirados da configuração da tela Estoque. O subtítulo da
tabela foi ajustado de “Saldo, cobertura e validade” para “Saldo e cobertura”,
para não deixar referência ao indicador removido. Os outros cards, alertas,
tabelas e regras foram mantidos. Não houve alteração de dados, integração ou
cálculo de estoque.

## Verificação

Foi incluído um teste de interface que exige a presença de `ESTOQUE FÍSICO` e
`GIRO`, verifica a ausência dos três itens removidos e confirma o novo
subtítulo. O teste e o build frontend passaram numa cópia isolada do projeto.
Captura atual: `docs/evidence/current-screens-2026-11-02/T09-estoque.png`.
