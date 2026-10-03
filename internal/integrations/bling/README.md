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

O `ReceivablesAPISyncService` já fecha a etapa seguinte de infraestrutura: lê
as páginas e grava cada página como evidência RAW append-only em uma transação
PostgreSQL, com hash, versão corrente, deduplicação e lote de sincronização.
Ele não transforma a página em recebimento nem em indicador enquanto o ID e os
campos financeiros do payload real não estiverem homologados.

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

O leitor impõe 1 MiB por campo, 4 MiB por registro e 256 campos enquanto consome o fluxo, sem limite total de linhas. Exceder qualquer teto é falha técnica do arquivo: a pasta não é importada. Erros de validação comuns por linha continuam permitindo salvar as linhas válidas e fechar o lote como `PARTIAL`.

O PDF `Bling - Relatório de Contas a Receber` fornecido em 20/09/2026 foi usado para confirmar visualmente os nomes e a semântica das colunas. O serviço `ReceiptImportService` abre a transação, carrega a conexão Bling, grava `raw_records`/`receipts` e finaliza o lote de sincronização de forma atômica.

## Pasta de entrada

`ImportReceiptsFolder` percorre uma pasta local em ordem alfabética e processa somente arquivos `.csv`. Cada arquivo recebe SHA-256 e contagens de importados/erros. Arquivos de Nuvemshop/Nuvem Pago não atendem ao cabeçalho Bling e são rejeitados explicitamente; nunca são reinterpretados como recebimentos realizados. Arquivos que não são CSV são ignorados e listados no resultado.

`ImportReceiptsFolder` continua sendo o adaptador de prévia legado e materializa o relatório para manter o contrato atual da tela. A importação de produção usa `StreamReceiptsFolder` e `StreamReceiptsCSV`: valida e persiste cada linha válida dentro da mesma transação do lote, sem guardar todas as linhas no heap. A deduplicação entre arquivos fica numa tabela temporária do PostgreSQL com HMAC por execução; o identificador externo em texto puro não é gravado nessa tabela, e ela é removida no commit ou rollback.

Falha técnica, cancelamento ou limite excedido reverte o lote inteiro. Erros de validação por linha continuam rejeitando somente as linhas inválidas e fechando o lote como `PARTIAL`, conforme a regra aprovada. Na importação, a lista de CSVs é lida em blocos de 256 nomes e ordenada numa tabela temporária, sem crescer no heap. A prévia ainda usa `os.ReadDir` e coleta linhas/erros em memória; esse fluxo permanece no backlog DATA-04.

O método IPC `bling.receipts.preview` chama essa leitura pelo worker e retorna apenas nomes de arquivos, hashes, contagens e até 50 problemas sanitizados. A UI nunca recebe nomes de clientes, históricos ou valores de linhas.

O método IPC `bling.receipts.import` só funciona quando o worker tem o PostgreSQL dedicado configurado em seu ambiente seguro (`MAJUCAU_DATABASE_URL`, sem senha em logs ou argumentos). Ele cria o lote e comita RAW + ledger atomicamente; sem banco ou conexão Bling configurada, retorna código estável e não altera arquivos nem dados.

## Validação PostgreSQL

O teste opt-in `TestReceiptImportServicePostgresE2E` roda no job PostgreSQL da CI com banco descartável e fixture sanitizada. Para executá-lo manualmente com o schema aplicado:

Evidência da execução aprovada em CI: `docs/evidence/CSV-STREAMING-CI-2026-10-02.md`.

```powershell
$env:MAJUCAU_TEST_DATABASE_URL = 'postgres://postgres@127.0.0.1:55439/majucau_test?sslmode=disable'
$env:MAJUCAU_TEST_RECEIPTS_FOLDER = (Resolve-Path 'internal/integrations/bling/testdata/e2e').Path
go test ./internal/integrations/bling -run TestReceiptImportServicePostgresE2E -count=1 -v
```

O fixture contém duas linhas pagas e uma linha em aberto. A primeira execução deve criar dois recebimentos e terminar como `PARTIAL`; uma nova execução deve atualizar os dois, sem duplicar o RAW corrente.
