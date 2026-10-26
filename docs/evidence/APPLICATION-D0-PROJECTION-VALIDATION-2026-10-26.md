# Validação da janela da calculadora — 2026-10-26

## Correção

`MaximumApplication` antes exigia 61 datas em sequência, mas não confirmava que
a primeira data era o D0 solicitado, que o primeiro saldo era provisório, que
os 60 seguintes eram projetados, nem que cada saldo fechava corretamente com
entradas, saídas e ajustes do dia ou que a abertura de cada dia correspondia ao
fechamento anterior. Uma série com datas consecutivas e valores inconsistentes
poderia, portanto, produzir um limite de aplicação.

O cálculo agora recebe a data de referência (`AsOfDate`) e reutiliza a validação
completa da projeção de 61 dias. A série precisa começar no D0 solicitado e
passar pelas mesmas verificações que o mínimo projetado, incluindo a passagem
do saldo de fechamento para a abertura do dia seguinte. `SuggestedReserve`
também exige que a data da primeira linha corresponda ao D0 informado.

## Regras preservadas

- A fórmula aprovada D-003-A continua igual.
- A sugestão D-003-C continua informativa e não altera o valor máximo.
- Reserva mínima, dados de entrada, desenho das telas e resultados financeiros
  aprovados não foram modificados.
- Uma projeção ausente ou inconsistente falha sem produzir um valor de
  aplicação.

## Verificações

- `go test ./internal/financial -count=1` passou.
- `scripts/verify.ps1` passou: testes Go, fronteiras de arquitetura, `go vet`,
  verificação de dependências, migrations, geração SQLC, busca de segredos,
  smoke do instalador, lint, typecheck, testes da interface, build e `npm audit`.
- `govulncheck` não encontrou vulnerabilidade alcançável pelo código. Ele
  encontrou uma vulnerabilidade em um módulo requerido que não é chamado pelo
  código verificado.
- Os testes opt-in que exigem um PostgreSQL externo não foram executados nesta
  máquina. Essa alteração cobre o cálculo puro e não altera persistência.

## Limites restantes

A calculadora ainda não está ligada à DRE própria nem ao fluxo de caixa com
fontes reais. A data inicial do CAPEX acumulado, aprovada como 01/01/2026, é uma
regra separada e permanece registrada no contrato da DRE. Esta correção não
conclui BK-006 nem homologa o protótipo.
