# Nuvem Pago — recebimentos futuros

Este pacote lê somente o CSV controlado colocado em
`02_nuvem_pago/recebimentos_futuros`. Ele representa direito futuro a
receber B2C e produz registros `PROJECTED` com origem `NUVEM_PAGO`.

As colunas financeiras do arquivo são preservadas: bruto, taxas, juros,
custos totais e líquido. Para a representação normalizada, deduções negativas
do extrato são armazenadas como valores não negativos; o próximo incremento
de persistência gravará o arquivo original em `raw_records`, mantendo o sinal
e a rastreabilidade da origem.

O pacote não importa recebimentos realizados, não dá baixa, não alimenta a
tabela `receipts` e não usa a tabela contratual de tarifas para recalcular o
líquido. Recebimentos realizados continuam vindo exclusivamente do Bling.
O caminho da pasta é uma seleção operacional explícita; o parser não tenta
inferir que um arquivo de recebidos seja futuro apenas pelo nome.

## Contrato aceito

CSV separado por ponto e vírgula, com cabeçalho equivalente a:

`Data do pagamento`, `Data de recebimento`, `Tipo de movimentação`, `Tipo de transação`, `Nº transação`, `Nome`, `Forma de pagamento`, `Bandeira`, `Nº Parcelas`, `Valor Bruto (R$)`, `Taxas (R$)`, `Juros (R$)`, `Custos Totais (R$)`, `Valor Liquido (R$)`.

Somente linhas `Entrada` e `Venda` são aceitas. Transações repetidas dentro
da mesma pasta são rejeitadas. O processamento é lexicalmente determinístico
e cada arquivo recebe SHA-256 para o lote posterior.

`ParseFutureCSV` e `ImportFutureFolder` são prévias em memória. Ainda falta
ligar o resultado ao worker e à transação PostgreSQL que grava RAW append-only
e faz upsert idempotente em `receivables`; esse é o próximo incremento antes
de considerar o fluxo homologado.
