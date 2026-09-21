# Nuvem Pago — recebimentos futuros

Este pacote lê somente o CSV controlado colocado em
`02_nuvem_pago/recebimentos_futuros`. Ele representa direito futuro a
receber B2C e produz registros `PROJECTED` com origem `NUVEM_PAGO`.

As colunas financeiras do arquivo são preservadas: bruto, taxas, juros,
custos totais e líquido. Para a representação normalizada, deduções negativas
do extrato são armazenadas como valores não negativos; o RAW mantém o sinal e
a rastreabilidade da origem.

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

`ParseFutureCSV` e `ImportFutureFolder` alimentam `FutureImportService`, que
abre uma transação PostgreSQL, cria o lote, grava o payload RAW versionado e
faz upsert idempotente em `receivables` como B2C/PROJECTED. O worker expõe
preview e importação pelos métodos IPC `nuvem_pago.future.preview` e
`nuvem_pago.future.import`; a tela de Importações/Integrações usa esses métodos
e apresenta apenas metadados sanitizados, sem nome ou PII. A importação falha
fechada quando banco, conexão ou confirmação operacional não estão configurados.
O campo `open_balance` permanece nulo porque este export não fornece saldo
aberto separado; nenhum valor foi inventado. Ainda faltam E2E contra
PostgreSQL, homologação visual e teste do instalador em VM limpa.

