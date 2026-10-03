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

O leitor impõe 1 MiB por campo, 4 MiB por registro e 256 campos enquanto
consome o fluxo, sem limite total de linhas. Exceder qualquer teto é falha
técnica: a pasta inteira é recusada antes de gravar dados. Erros de validação
comuns por linha continuam permitindo salvar as linhas válidas e fechar o
lote como `PARTIAL`.

`PreviewFutureFolder` atende a prévia local sem depender de PostgreSQL. Ela
processa linhas em streaming, deduplica com HMAC em índice temporário cifrado,
ordena arquivos por runs cifrados e remove o diretório ao final. Mantém no
máximo 50 metadados de arquivo e 50 problemas, com `file_count` e contagens
totais para preservar a mesma apresentação na tela. Um preflight em streaming
mantém a validação de arquivo inteiro e detecta mudanças durante a leitura.
`ParseFutureCSV` e `ImportFutureFolder` permanecem como helpers internos.
`FutureImportService` usa `StreamFutureFolder` e `StreamFutureCSV`: valida cada
linha e grava payload RAW versionado e upsert idempotente em `receivables`
como B2C/PROJECTED dentro da transação do lote. A deduplicação de produção
fica numa tabela temporária do PostgreSQL com HMAC por execução; o identificador
externo em texto puro não é gravado nela, e a tabela some no commit ou rollback.
Falhas técnicas, cancelamento ou limite excedido revertem todo o lote; erros de
validação por linha preservam as válidas e fecham como `PARTIAL`. Na importação,
os nomes CSV são lidos em blocos de 256 e ordenados em tabela temporária. O worker expõe
preview e importação pelos métodos IPC `nuvem_pago.future.preview` e
`nuvem_pago.future.import`; a tela de Importações/Integrações usa esses métodos
e apresenta apenas metadados sanitizados, sem nome ou PII. A importação falha
fechada quando banco, conexão ou confirmação operacional não estão configurados.
O campo `open_balance` permanece nulo porque este export não fornece saldo
aberto separado; nenhum valor foi inventado. O E2E PostgreSQL roda na CI com
fixture sanitizada; homologação visual e teste do instalador em VM limpa ainda
faltam.

## Validação PostgreSQL

O teste opt-in `TestFutureImportServicePostgresE2E` roda na CI contra o mesmo
banco descartável da validação Bling. Para executá-lo manualmente, use uma
pasta sanitizada de recebimentos futuros:

Evidência da execução aprovada em CI: `docs/evidence/CSV-STREAMING-CI-2026-10-02.md`.

```powershell
$env:MAJUCAU_TEST_DATABASE_URL = 'postgres://postgres@127.0.0.1:55439/majucau_test?sslmode=disable'
$env:MAJUCAU_TEST_FUTURE_FOLDER = (Resolve-Path 'internal/integrations/nuvempago/testdata/e2e').Path
go test ./internal/integrations/nuvempago -run TestFutureImportServicePostgresE2E -count=1 -v
```

O E2E exige migration aplicada em PostgreSQL descartável. Ele verifica duas
linhas `B2C/PROJECTED`, uma rejeição, reexecução idempotente, dois RAW atuais e
zero linhas no ledger `receipts` do Bling. Em 21/09/2026 ele passou em um
cluster PostgreSQL 18.3 isolado em loopback.

