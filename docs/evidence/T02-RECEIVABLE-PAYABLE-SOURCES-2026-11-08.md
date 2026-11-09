# T02 — origem e conferência das receitas e despesas — 08/11/2026

## Regra aprovada

- A T02 não usa o relatório preparado “Fluxo de Caixa” do Bling.
- **Receitas:** títulos em aberto do relatório **Contas a Receber do Bling**, agrupados pela data de vencimento e usando o valor líquido mostrado na coluna “Líquido”. A lista do Bling já contém títulos cuja forma de pagamento é Nuvem Pago; portanto o arquivo externo de recebimentos futuros não deve ser somado outra vez na T02.
- **Despesas:** títulos em aberto do relatório **Contas a Pagar do Bling**, agrupados pela data de vencimento e usando o saldo em aberto como saída.
- Incluir todas as contas financeiras que aparecem nos relatórios; não filtrar títulos pelo banco associado. O saldo bancário é uma medida separada e parte do último saldo conciliado do Itaú.
- Abertura + receitas líquidas − despesas = fechamento semanal. O fechamento de uma semana abre a semana seguinte. Sem prazo de liquidação nem float. Sem rótulo “Parcial” nos totais.

## Referência visual de 22/11/2026 a 28/11/2026

As capturas da responsável mostram “Todas as contas financeiras” e títulos “Em aberto”. Contas a Pagar: 14 títulos, total de **R$ 25.275,71**. Contas a Receber: 146 títulos, **R$ 45.482,22 bruto**, taxas de **R$ 1.691,56** e **R$ 43.790,66 líquido**. O resultado de receita da semana deve reconciliar com o líquido da lista para o mesmo intervalo e filtros.

O próprio Bling descreve o líquido como valor total menos as taxas da forma de pagamento e documenta a taxa percentual, fixa e arredondamento. A API de Contas a Receber não retorna a coluna “Líquido”; para reproduzi-la, o app usa as configurações oficiais da forma de pagamento obtidas na mesma carga diária. Se as taxas necessárias não forem conhecidas, não deve substituir o resultado por bruto.

## Auditoria e causa encontrada

A tela apresentava receita semanal diferente por dois motivos de composição:

1. A consulta aplicava a deduplicação entre Bling e o arquivo externo da Nuvem Pago. Isso podia remover uma linha do Bling, embora o relatório de Contas a Receber já contenha os recebimentos Nuvem Pago.
2. A consulta usava `open_balance` sem descontar as taxas da forma de pagamento, apresentando valor bruto onde a referência exige líquido.

A T02 foi alterada para ler as linhas do snapshot diário de Contas a Receber do Bling diretamente, sem acrescentar a fonte externa nem deduplicar entre fontes. O valor líquido é calculado com as taxas oficiais da mesma execução; a série de despesas permanece vinculada ao snapshot de Contas a Pagar. Um teste de regressão verifica essas fontes e impede o retorno do fallback bruto.

## Validação e limites

- `go test ./internal/dashboard ./internal/financial -count=1`: passou.
- `go test ./... -count=1`: passou.
- A integração com PostgreSQL real não foi executada porque `MAJUCAU_TEST_DATABASE_URL` não está configurada neste ambiente.
- A validação automatizada comprova a regra no código; não comprova que a execução diária real tenha a mesma lista de 146 linhas nem o total líquido da captura. A janela 22/11–28/11 precisa ser conferida no protótipo após atualizar o worker e sincronizar novamente.
- A auditoria posterior da ação Atualizar confirmou que o ícone de cabeçalho só relê o snapshot local; ainda não há nova sincronização automática pelo clique. Ver `T02-UPDATE-ACTION-AUDIT-2026-11-08.md`.
- O protótipo foi reaberto pelo roteiro oficial com o worker recompilado às 23:02 de 08/11 e a interface local respondeu HTTP 200. A captura visual da janela falhou e não foi feita nova sincronização. Não afirmar que o valor de R$ 43.790,66 já aparece na tela até a conferência visual e numérica do snapshot atualizado.
