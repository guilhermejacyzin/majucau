# T02 — fontes de receitas, despesas e saldo — 08/11/2026

## Decisão confirmada

Gisele determinou que a T02 não use os valores do relatório “Fluxo de Caixa” do
Bling. A tela deve montar as linhas a partir das telas de origem:

- **Receitas:** Contas a Receber, títulos em aberto incluídos no período, pelo
  valor líquido mostrado na coluna “Líquido”. Se a origem não fornecer líquido,
  usar o mesmo fallback da tela de recebíveis (valor bruto).
- **Despesas:** Contas a Pagar, títulos em aberto incluídos no período.
- **Saldo atual e saldo inicial projetado:** último saldo conciliado do Itaú.
- Todas as contas financeiras dentro de Contas a Receber/Pagar entram nos títulos;
  a conta associada ao título não restringe a soma. Somente o saldo permanece
  restrito ao Itaú.

A fórmula semanal aprovada é saldo inicial + receitas − despesas. O saldo final
semanal inicia a semana seguinte. A série não deve adicionar movimentos da tela
Caixas e Bancos às linhas de receita/despesa da T02.

## Evidência de referência fornecida

As capturas enviadas mostram Contas a Pagar e Contas a Receber para o período
22/11/2026–28/11/2026, com “Todas as contas financeiras” e títulos “Em aberto”.
A captura de Contas a Pagar mostra 14 títulos e total de **R$ 25.275,71**. A de
Contas a Receber mostra 146 títulos: **R$ 45.482,22 bruto** e **R$ 43.790,66
líquido**. Para a entrada da T02, a referência correta é **R$ 43.790,66**,
antes de conferir como o conjunto deduplicado das fontes locais se reconcilia
com os 146 títulos exibidos.

Os números da T02 devem coincidir com os títulos e valores apresentados nas
telas de origem. Para recebimentos, usar o valor **líquido** por título da tela
Contas a Receber (com o mesmo fallback para bruto quando não houver líquido);
para pagamentos, usar o saldo em aberto de Contas a Pagar. As datas são as datas
de vencimento e a soma não recalcula tarifas nem prazos de liquidação. Não exibir
o marcador “Parcial” nesses totais. Aplicar a regra aprovada de correspondência
entre Nuvem Pago e Bling para não contar duas vezes a mesma venda.

## Mudança técnica nesta revisão

- A consulta de fluxo usa títulos normalizados de Contas a Receber e Contas a
  Pagar, consolida as fontes de recebimento com a regra aprovada de não
  duplicidade e agrupa pela data efetiva de vencimento.
- A entrada da T02 usa `net_amount` (coluna “Líquido”); se não houver valor
  líquido disponível para a linha, usa `gross_amount`, como na tabela da tela de
  recebíveis. Não usa o saldo bancário nem o valor bruto quando há líquido.
- Títulos não são mais descartados porque estão vinculados a outra conta
  financeira. A inclusão por status, saldo em aberto e janela de datas continua
  aplicada. O valor e o vencimento passam diretamente para o agrupamento semanal,
  mantendo a mesma origem e o mesmo valor da tela de Contas a Receber/Pagar.
- Removidos os cálculos de taxa, prazo de liquidação e float da série T02, que
  pertenciam à reprodução do relatório preparado de Fluxo de Caixa e alteravam
  os valores ou suas semanas.
- Removida a trava que transformava os totais em indisponíveis quando os lotes
  de origem não eram identificados na mesma execução. A série sempre retorna os
  valores numéricos calculados dos títulos normalizados, sem “Parcial”.
- O ícone de atualização no cabeçalho recarrega o retrato mais recente do
  worker. Atualiza a tela, mas não dispara uma nova sincronização com o
  provedor; sincronizar continua sendo uma ação separada em Integrações.
- Se a tabela auxiliar da migration 000024 ainda não existir no banco local, a
  consulta conserva o último checkpoint aprovado do Itaú e continua montando
  os totais de Receber/Pagar. A atualização do saldo com conciliações posteriores
  depende de instalar essa migration; a ausência dela não deve mais apagar a
  tela toda.
- O saldo-base do gráfico volta a usar o resultado de `readLatestItauBalance`;
  essa mesma consulta alimenta o card de saldo, mantendo o Itaú como única fonte
  de saldo.
- O uso do relatório “Fluxo de Caixa” e a seleção de contas Itaú + Nuvem Pago
  permanecem registrados como histórico anterior, mas estão substituídos para
  esta composição da T02.
- A Visão Executiva e a T02 limitam a apresentação do gráfico à data atual em
  São Paulo, removendo o dia anterior que aparecia como 07/11 quando a tela já
  devia iniciar em 08/11. O filtro da T02 continua controlando tabela e gráfico.

## Verificação e limites

- `go test ./internal/dashboard -count=1`: passou depois da correção dos totais;
  a integração real do PostgreSQL não executa neste ambiente sem
  `MAJUCAU_TEST_DATABASE_URL`.
- `npm run build`: passou após corrigir a regra de agrupamento semanal do
  protótipo.
- O teste PostgreSQL de integração foi ignorado porque
  `MAJUCAU_TEST_DATABASE_URL` não está configurada neste ambiente; a consulta
  precisa ser validada pela CI PostgreSQL.
- `go build ./cmd/...`: passou após a alteração da consulta da T02.
- `npm run build`: passou após ligar o botão de atualização ao recarregamento
  do retrato do painel.
- `git diff --check`: passou.
- As capturas comprovam o desenho dos filtros e a diferença entre valor bruto e
  líquido no relatório de origem. Ainda falta atualizar o worker em execução,
  sincronizar os dados e conferir numericamente a T02 com o intervalo
  22/11–28/11. Os testes automatizados e a integração PostgreSQL também não
  foram executados nesta revisão. Portanto, a integração real e a homologação
  continuam pendentes.
