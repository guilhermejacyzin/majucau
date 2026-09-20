# Bling — API oficial e importador controlado de apoio

## API oficial (caminho principal)

O Bling/API v3 é a fonte de verdade para fatos financeiros operacionais, em
modo somente leitura. O pacote agora contém um cliente técnico para:

- OAuth 2 Authorization Code e renovação por `refresh_token`;
- `GET /contas/receber` com paginação (`pagina`/`limite`), filtros explícitos e
  limite máximo de 100 registros;
- classificação segura de 401/403, 429, timeout, indisponibilidade e resposta
  fora do contrato;
- preservação dos itens retornados como JSON bruto até a homologação dos
  campos reais da conta Bling.

O cliente não grava credenciais, não escreve no Bling, não faz scraping e não
assume campos financeiros que ainda não foram confirmados em uma resposta
real. O segredo deve chegar somente ao worker, ser protegido por DPAPI e
nunca ser enviado ao frontend, colocado em URL ou incluído em logs/erros.

O teste de conexão faz uma leitura mínima de `contas/receber` com uma linha.
Isso prova transporte, autorização e formato básico, mas não equivale à
homologação funcional dos campos. A homologação BK-040 ainda precisa registrar
endpoint, campo, significado, ausência, limites e uma amostra sanitizada da
conta autorizada antes da normalização contábil.

## Importador CSV de apoio — Contas Recebidas

O importador usa o relatório do Bling como uma entrada local controlada para
fixtures, conferência e contingência operacional aprovada. Ele não substitui a
API oficial e não deve ser tratado como a implementação final de sincronização.
A forma de pagamento pode conter `NUVEMPAGO 1X`, `NUVEMPAGO 2X`, `NUVEMPAGO 3X`
ou `Nuvemshop PIX`, mas isso não muda `source_system=BLING`.

## Contrato

O parser espera exportação CSV separada por ponto e vírgula com os campos equivalentes a:

`Cliente`, `Histórico`, `Forma de pagamento`, `Nº documento`, `Vencimento`, `Liquidação`, `Situação`, `Valor taxa`, `Recebido`.

Somente `Situação=pago` vira `ReceiptCandidate` confirmado. Linhas abertas, canceladas ou sem documento ficam em `RowError`; não são convertidas em zero nem entram no total. Documentos duplicados no mesmo lote também ficam em erro para impedir dupla contagem.

O PDF `Bling - Relatório de Contas a Receber` fornecido em 20/09/2026 foi usado para confirmar visualmente os nomes e a semântica das colunas. O serviço `ReceiptImportService` abre a transação, carrega a conexão Bling, grava `raw_records`/`receipts` e finaliza o lote de sincronização de forma atômica.

## Pasta de entrada

`ImportReceiptsFolder` percorre uma pasta local em ordem alfabética e processa somente arquivos `.csv`. Cada arquivo recebe SHA-256 e contagens de importados/erros. Arquivos de Nuvemshop/Nuvem Pago não atendem ao cabeçalho Bling e são rejeitados explicitamente; nunca são reinterpretados como recebimentos realizados. Arquivos que não são CSV são ignorados e listados no resultado.

O resultado da pasta é uma prévia determinística em memória. O worker deve abrir a transação, criar o lote, chamar `PersistReceipts`, finalizar o lote e somente então avançar o cursor/estado da integração.

O método IPC `bling.receipts.preview` chama essa leitura pelo worker e retorna apenas nomes de arquivos, hashes, contagens e até 50 problemas sanitizados. A UI nunca recebe nomes de clientes, históricos ou valores de linhas.

O método IPC `bling.receipts.import` só funciona quando o worker tem o PostgreSQL dedicado configurado em seu ambiente seguro (`MAJUCAU_DATABASE_URL`, sem senha em logs ou argumentos). Ele cria o lote e comita RAW + ledger atomicamente; sem banco ou conexão Bling configurada, retorna código estável e não altera arquivos nem dados.

## Validação PostgreSQL

O teste opt-in `TestReceiptImportServicePostgresE2E` não abre banco no fluxo normal de CI. Para executá-lo contra um PostgreSQL descartável com o schema aplicado:

```powershell
$env:MAJUCAU_TEST_DATABASE_URL = 'postgres://postgres@127.0.0.1:55439/majucau_test?sslmode=disable'
$env:MAJUCAU_TEST_RECEIPTS_FOLDER = (Resolve-Path 'internal/integrations/bling/testdata/e2e').Path
go test ./internal/integrations/bling -run TestReceiptImportServicePostgresE2E -count=1 -v
```

O fixture contém duas linhas pagas e uma linha em aberto. A primeira execução deve criar dois recebimentos e terminar como `PARTIAL`; uma nova execução deve atualizar os dois, sem duplicar o RAW corrente.
