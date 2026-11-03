# Leitura visual do relatório Fluxo de Caixa do Bling — 02/11/2026

## Fonte e limite

Leitura dos dados visíveis na captura enviada pela responsável. Os números abaixo
servem para conferir a origem e a apresentação; não foram gravados como dados do
Majucau nem como valores permanentes. A captura mostra totais semanais, não a
série diária completa.

## Opções visíveis no relatório

- Período: 02/11/2026 a 09/12/2026.
- Contas financeiras: 2 contas selecionadas.
- Considerar float bancário: ativado.
- Considerar prazo de recebimento da forma de pagamento: ativado.
- Considerar pagamentos e recebimentos: ativado.
- Ajuste de saldo em todas as semanas: R$ 0,00.

## Valores lidos da tabela

| Período | Saldo inicial/acumulado | Receitas | Despesas | Ajuste de saldo | Saldo acumulado ao fim |
|---|---:|---:|---:|---:|---:|
| 02/11 a 07/11 | R$ 74.463,36 | R$ 8.401,31 | -R$ 25.483,13 | R$ 0,00 | R$ 57.381,54 |
| 08/11 a 14/11 | R$ 57.381,54 | R$ 12.362,61 | -R$ 18.020,38 | R$ 0,00 | R$ 51.723,77 |
| 15/11 a 21/11 | R$ 51.723,77 | R$ 37.383,22 | -R$ 152.735,18 | R$ 0,00 | -R$ 63.628,19 |
| 22/11 a 28/11 | -R$ 63.628,19 | R$ 9.722,56 | -R$ 22.264,52 | R$ 0,00 | -R$ 76.170,15 |
| 29/11 a 05/12 | -R$ 76.170,15 | R$ 7.124,43 | -R$ 57.445,14 | R$ 0,00 | -R$ 126.490,86 |
| 06/12 a 09/12 | -R$ 126.490,86 | R$ 1.606,17 | -R$ 5.104,29 | R$ 0,00 | -R$ 129.988,98 |
| **Total mostrado pelo Bling** | **R$ 74.463,36** | **R$ 76.600,30** | **-R$ 281.052,64** | **R$ 0,00** | **-R$ 129.988,98** |

## Correções de ligação registradas

- A fonte dos resultados e da projeção da T02 é o relatório/resultado Fluxo de
  Caixa do Bling, conforme orientação da responsável.
- O saldo por banco é uma fonte distinta: Caixas e Bancos. Não deve ser
  substituído por contas a pagar/receber.
- Os cards “ENTRADAS 30 DIAS” e “SAÍDAS 30 DIAS” não podem reutilizar as métricas
  de Contas a Receber e Contas a Pagar. A captura agregada em semanas também não
  fornece precisão para recalcular uma janela móvel de 30 dias que corte uma
  dessas semanas.
- Nenhum valor da captura foi copiado para a base de dados do produto.

## Próxima etapa técnica

Implementar e validar a ingestão do resultado de Fluxo de Caixa do Bling,
preservando as opções de relatório que definem o resultado e a trilha de origem.
Só então ligar as métricas e a série da T02, e fazer a T01 reutilizar esses
mesmos resultados. A documentação pública da API consultada enumera recursos de
Caixas e Bancos, Contas a Pagar e Contas a Receber; ela não lista um endpoint de
relatório de Fluxo de Caixa. Portanto, não presumir uma rota nem substituir o
resultado do relatório por uma soma parcial de outras telas.
