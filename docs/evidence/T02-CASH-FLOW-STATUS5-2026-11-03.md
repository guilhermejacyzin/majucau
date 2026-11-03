# Evidência T02 — exclusão de título cancelado e comparação com o Bling

**Data:** 03/11/2026, 06:30 BRT  
**Escopo:** somente consulta às APIs oficiais do Bling, correção da projeção e verificação do protótipo. Sem alteração de tela ou regra financeira aprovada.

## Achado e correção

A API preserva `situacao` de Contas a Pagar como código numérico no banco. Um título da carga normalizada tinha situação `5`, saldo aberto R$ 1.794,15 e vencimento 20/11/2026. O SQL da T02 excluía valores textuais (`CANCELLED`, `CANCELED`, `REFUNDED`, `REVERSED`, `ESTORNADO`), mas não o código `5`; portanto, somava o título cancelado às saídas.

Incluí o código `5` nas exclusões do fluxo da T02 e nos filtros de períodos de Contas a Pagar. A regressão PostgreSQL cria um título `5` de R$ 1.794,15 com vencimento dentro de 30 e 90 dias e verifica que ele não altera a série, nem os totais dos dois períodos.

## Valores visíveis após reiniciar o worker

- Entradas próximas de 30 dias: **R$ 69.994,12 — Parcial**.
- Saídas próximas de 30 dias: **-R$ 262.136,32 — Parcial**.
- Antes da correção, o card de saída mostrava **-R$ 263.930,47**; a diferença removida foi exatamente **R$ 1.794,15**.
- Saldo atual, menor saldo e semana crítica continuam indisponíveis porque a composição não tem saldo inicial comprovado. Não calculei esses valores por estimativa.

## Comparação com a captura do relatório

A captura de referência foi gerada para 02/11–09/12/2026 com Itaú e Mercado Pago, float ativado e pagamentos/recebimentos considerados. O artigo oficial do Bling explica que o relatório de fluxo de caixa consulta contas futuras a pagar/receber e aplica as configurações de taxas e prazo da forma de pagamento para a previsão líquida. Fontes: [como gerar o relatório Fluxo de Caixa](https://ajuda.bling.com.br/hc/pt-br/articles/360038306714-Como-gerar-um-relat%C3%B3rio-de-futuras-entradas-e-sa%C3%ADdas-de-valores), [taxas e prazo de recebimento no Bling](https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling) e [referência oficial da API](https://developer.bling.com.br/referencia).

| Período semanal | Entradas API | Entradas do relatório | Saídas API | Saídas do relatório |
|---|---:|---:|---:|---:|
| 08–14/11 | R$ 12.362,61 | R$ 12.362,61 | -R$ 18.020,38 | -R$ 18.020,38 |
| 15–21/11 | R$ 37.383,21 | R$ 37.383,22 | -R$ 152.735,18 | -R$ 152.735,18 |
| 22–28/11 | R$ 9.722,56 | R$ 9.722,56 | -R$ 22.264,52 | -R$ 22.264,52 |
| 29/11–05/12 | R$ 7.124,43 | R$ 7.124,43 | -R$ 57.445,14 | -R$ 57.445,14 |
| 06–09/12 | R$ 1.606,17 | R$ 1.606,17 | -R$ 5.104,29 | -R$ 5.104,29 |

Cinco de cinco faixas de saída conferem; quatro de cinco faixas de entrada conferem. Resta uma diferença de **R$ 0,01** nas entradas de 15–21/11. A faixa de 02–07/11 não foi incluída na comparação porque a carga atual começa em 03/11, um dia depois da referência.

O relatório exibe saldo inicial de **R$ 74.463,36**. A consulta local aos movimentos históricos das duas contas não conseguiu demonstrar esse saldo inicial; por isso, a T02 mantém o estado Parcial e os indicadores que dependem dele vazios. Isso segue a regra registrada em `AGENTS.md`: o saldo inicial ausente/histórico não comprovado fica pendente, sem estimativa.

## Verificações executadas

- `go test ./internal/dashboard -count=1` com o PostgreSQL de desenvolvimento na porta isolada 55438 — passou, incluindo a regressão de status `5` e os totais 30/90 dias.
- `go vet ./internal/dashboard` — passou.
- Worker de desenvolvimento recompilado e reiniciado; a janela T02 mostrou a saída corrigida em **-R$ 262.136,32 — Parcial**.
- Não promovi os valores a Confirmado nem avancei o percentual: BK-062 continua em **25%**; o projeto fica em **12,2%**.

## Próximos passos

1. Obter um recorte atualizado do mesmo relatório para investigar o centavo e completar 02–07/11.
2. Identificar uma fonte oficial comprovável para saldo inicial/float e comparar a posição por conta.
3. Só depois retirar Parcial e validar os indicadores de menor saldo e semana crítica.
