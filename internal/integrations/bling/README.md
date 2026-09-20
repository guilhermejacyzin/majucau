# Importador Bling — Contas Recebidas

O importador usa o relatório do Bling como fonte de verdade para recebimentos realizados. A forma de pagamento pode conter `NUVEMPAGO 1X`, `NUVEMPAGO 2X`, `NUVEMPAGO 3X` ou `Nuvemshop PIX`, mas isso não muda `source_system=BLING`.

## Contrato

O parser espera exportação CSV separada por ponto e vírgula com os campos equivalentes a:

`Cliente`, `Histórico`, `Forma de pagamento`, `Nº documento`, `Vencimento`, `Liquidação`, `Situação`, `Valor taxa`, `Recebido`.

Somente `Situação=pago` vira `ReceiptCandidate` confirmado. Linhas abertas, canceladas ou sem documento ficam em `RowError`; não são convertidas em zero nem entram no total. Documentos duplicados no mesmo lote também ficam em erro para impedir dupla contagem.

O PDF `Bling - Relatório de Contas a Receber` fornecido em 20/09/2026 foi usado para confirmar visualmente os nomes e a semântica das colunas. O adapter `PersistReceipts` já grava `raw_records`/`receipts` na transação fornecida pelo worker; a ligação ao ciclo de vida do serviço e à configuração da pasta fica para a integração operacional.

## Pasta de entrada

`ImportReceiptsFolder` percorre uma pasta local em ordem alfabética e processa somente arquivos `.csv`. Cada arquivo recebe SHA-256 e contagens de importados/erros. Arquivos de Nuvemshop/Nuvem Pago não atendem ao cabeçalho Bling e são rejeitados explicitamente; nunca são reinterpretados como recebimentos realizados. Arquivos que não são CSV são ignorados e listados no resultado.

O resultado da pasta é uma prévia determinística em memória. O worker deve abrir a transação, criar o lote, chamar `PersistReceipts`, finalizar o lote e somente então avançar o cursor/estado da integração.

O método IPC `bling.receipts.preview` chama essa leitura pelo worker e retorna apenas nomes de arquivos, hashes, contagens e até 50 problemas sanitizados. A UI nunca recebe nomes de clientes, históricos ou valores de linhas.

O método IPC `bling.receipts.import` só funciona quando o worker tem o PostgreSQL dedicado configurado em seu ambiente seguro (`MAJUCAU_DATABASE_URL`, sem senha em logs ou argumentos). Ele cria o lote e comita RAW + ledger atomicamente; sem banco ou conexão Bling configurada, retorna código estável e não altera arquivos nem dados.
