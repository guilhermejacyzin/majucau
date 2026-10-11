# Evidência — sugestão informativa de reserva na calculadora

## Decisão

Em 10/10/2026, Gisele aprovou exibir uma sugestão de reserva para hoje, somente para consulta:

```text
reserva_sugerida_hoje = R$ 50.000 + MAX(0,
  caixa_projetado_D0 - menor_saldo_diario_projetado_D0_D60)
```

O cálculo usa a previsão consolidada do caixa. Recebimentos da Nuvemshop não devem ser somados separadamente como uma regra adicional.

## Exemplo simples

- Caixa projetado para hoje: R$ 100.000.
- Menor saldo diário previsto nos próximos 60 dias: R$ 70.000.
- Queda prevista: R$ 100.000 − R$ 70.000 = R$ 30.000.
- Reserva sugerida para hoje: R$ 50.000 + R$ 30.000 = **R$ 80.000**.

Assim, a previsão mantém R$ 50.000 no ponto mais baixo. É uma sugestão; não é uma aplicação nem uma ordem para movimentar dinheiro.

## O que fica igual

- A reserva mínima aprovada continua em R$ 50.000.
- O cálculo do valor máximo para aplicar continua seguindo D-003-A.
- A sugestão não altera o máximo, a DRE, o fluxo de caixa ou qualquer transferência.
- Se o saldo projetado D0 ou o menor saldo D0–D+60 não estiver comprovado, a sugestão não pode ser inventada; o campo deve seguir os estados de indisponibilidade já aprovados.

## Próximo trabalho técnico

### O que o código já possui

- A tabela `daily_balances` prevê saldo de abertura, entradas, saídas, ajustes, saldo de fechamento, data, cenário e execução do cálculo.
- Tabelas de vínculo relacionam cada saldo diário a contas a receber, recebimentos, contas a pagar, pagamentos e ajustes financeiros.
- O código Go tem funções que montam os 61 saldos de D0 a D+60, encontram o menor saldo e calculam o valor máximo de aplicação.
- A busca no código não encontrou chamadas de produção dessas funções. Portanto, a estrutura e as fórmulas existem, mas não há evidência de que a tela ou a calculadora já recebam valores reais dessa fonte.

### Próximo trabalho técnico

Confirmar e ligar o saldo D0 e a série diária D0–D+60 à execução consolidada aprovada, mantendo cenário, data-base, execução, fonte e estado de qualidade. Depois documentar exemplos reproduzíveis e validar a exibição sem alterar o desenho aprovado. A fórmula da sugestão está aprovada e sua função isolada foi implementada; a origem real e a chamada de produção continuam pendentes.

## Implementação isolada da sugestão — 2026-10-11

- `internal/financial.SuggestedReserve` calcula `reserva mínima + MAX(0, fechamento D0 − menor fechamento D0–D+60)` após validar a série completa de 61 dias por `FindMinimumProjectedBalance`.
- O cálculo usa `domain.Money` com precisão exata, recusa reserva negativa e propaga erros de projeção e estouro. Ele é independente de `MaximumApplication`; não muda o máximo aprovado.
- Os testes locais confirmam o exemplo aprovado (D0 de R$ 100.000, mínimo de R$ 70.000, sugestão de R$ 80.000), que o máximo continua em R$ 20.000 nessa mesma série, e a rejeição de horizonte incompleto/reserva negativa/estouro.
- `go test ./internal/financial -count=1`, `go vet ./internal/financial` e `git diff --check` passaram localmente. A validação remota da CI fica pendente após o envio do commit.
- A função ainda não tem chamada de produção nem entrada ligada ao saldo diário real. Sem dados comprovados, a sugestão continua indisponível. Nenhuma tela, planilha ou regra de aplicação foi alterada.
