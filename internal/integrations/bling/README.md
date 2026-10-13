# Bling — API oficial e importador controlado de apoio

## API oficial (caminho principal)

O Bling/API v3 é a fonte de verdade para fatos financeiros operacionais, em
modo somente leitura. O pacote contém um cliente técnico para:

- OAuth 2 Authorization Code e renovação por `refresh_token`, com o cabeçalho
  `enable-jwt: 1` exigido pela migração JWT;
- `GET /contas/receber` e `GET /contas/pagar`, com paginação (`pagina`/`limite`)
  e filtros explícitos por rota;
- consulta de detalhe `GET /contas/pagar/{idContaPagar}`,
  `GET /borderos/{idBordero}`, `GET /caixas/{idCaixa}` e
  `GET /categorias/receitas-despesas/{id}`;
- leitura de uma página limitada de `GET /caixas`, com página e período; o
  cliente aplica até 100 registros somente como limite local. A OpenAPI
  pública atual não documenta o parâmetro `limite` nessa rota nem informa o
  marcador de última página; cada filtro de data aceita no máximo 64 bytes;
- leitura de uma página de `GET /categorias/receitas-despesas`;
- classificação segura de 401/403, 429, timeout, indisponibilidade e resposta
  fora do contrato;
- preservação dos itens retornados como JSON bruto até a homologação dos
  campos reais da conta Bling.

As leituras são limitadas a 4 MiB por resposta e 100 registros por página; 100
é um teto local de segurança, não um máximo comprovado pelo Bling. Consultas
`GET` repetíveis usam o retry comum do projeto, com até três tentativas totais
para falhas transitórias. A troca e a renovação OAuth usam `POST` e não recebem
retry automático.

O cliente não grava credenciais, não escreve no Bling, não faz scraping e não
assume campos financeiros que ainda não foram confirmados em uma resposta
real. O segredo deve chegar somente ao worker, ser protegido por DPAPI e
nunca ser enviado ao frontend, colocado em URL ou incluído em logs/erros.

Os filtros de data seguem o contrato por recurso: Contas a Receber usa
`tipoFiltroData` (`V` para vencimento, `R` para recebimento) junto de
`dataInicial`/`dataFinal`; Contas a Pagar usa `dataVencimentoInicial`/
`dataVencimentoFinal` ou `dataPagamentoInicial`/`dataPagamentoFinal`. A
sincronização manual por vencimento consulta os dois recursos; por recebimento,
somente Contas a Receber; por pagamento, somente Contas a Pagar. Uma consulta
com apenas situação mantém o comportamento anterior e consulta ambos. Estes
filtros de data não fornecem, por si só, um cursor de alterações desde a última
execução.

As consultas de detalhe e categorias aplicam o mesmo limite de resposta,
cancelamento, JWT e retry seguro. Elas retornam apenas o objeto `data` como JSON
bruto; ainda não persistem esses detalhes nem os transformam em fatos do
Majucau. A leitura de categorias retorna uma página por chamada.

O teste de conexão faz uma leitura mínima de `contas/receber` com uma linha.
Isso prova transporte, autorização e formato básico, mas não equivale à
homologação funcional dos campos. A homologação BK-040 ainda precisa registrar
endpoint, campo, significado, ausência, limites e uma amostra sanitizada da
conta autorizada antes da normalização contábil.

A integração não considera uma página curta ou metadado ausente como prova de
que a lista terminou: o chamador deve conferir `HasNextKnown` antes de avançar
ou fechar a sincronização. O contrato público de recebíveis não documenta um
marcador de última página nem um cursor incremental por alteração; por isso,
essa homologação continua pendente. O detalhe está em
`docs/evidence/BLING-PAGINATION-CONTRACT-2026-10-11.md`.

Os schemas públicos de categorias e de caixa descrevem `data[]`, sem informar
como identificar a última página; o cliente não pode declarar a lista completa
sem esse marcador. O adaptador agora oferece `ListCashMovements` e
`GetCashMovement`, que aplicam limite local de registros, JWT, cancelamento e
retry seguro, e preservam a resposta como RAW. Em `GET /caixas`, esse limite
não é enviado como parâmetro `limite`, pois a OpenAPI atual não o documenta.
A leitura de caixa ainda não é persistida nem integrada ao snapshot. A lista
não envia filtro por ID de categoria até que sua codificação exata seja
homologada. Os IDs/hierarquia de CAPEX e a
correspondência dos movimentos com os relatórios da Majucau precisam ser
confirmados antes de exibir valores. Consulte
`docs/evidence/CAPEX-REPORT-STRUCTURE-REVIEW-2026-10-11.md`.

Um `total` sem tamanho de página conhecido não basta para inferir se há próxima
página: o limite local de registros não é usado como se tivesse sido enviado
ao Bling. A leitura permanece incompleta até receber um marcador explícito ou
homologar o contrato real da conta.

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

O worker usa `PreviewReceiptsFolder` para a prévia local sem dependência de PostgreSQL. As linhas são processadas em streaming; a deduplicação exata usa HMAC com chave aleatória por execução em um índice temporário cifrado no disco, e os nomes são ordenados em runs temporários cifrados. O diretório é removido ao final. A prévia retém no máximo 50 metadados de arquivo e 50 problemas, mantendo contagens totais em `file_count` e `error_count`; a tela preserva a mesma apresentação. Um preflight em streaming mantém a validação de arquivo inteiro e detecta mudança do conteúdo durante a prévia.

`ImportReceiptsFolder` permanece disponível como helper legado para callers internos, mas não é usado pelo worker para a prévia. A importação de produção usa `StreamReceiptsFolder` e `StreamReceiptsCSV`: valida e persiste cada linha válida dentro da mesma transação do lote, sem guardar todas as linhas no heap. A deduplicação de produção fica numa tabela temporária do PostgreSQL com HMAC por execução; o identificador externo em texto puro não é gravado nessa tabela, e ela é removida no commit ou rollback.

Falha técnica, cancelamento ou limite excedido reverte o lote inteiro. Erros de validação por linha continuam rejeitando somente as linhas inválidas e fechando o lote como `PARTIAL`, conforme a regra aprovada. Na importação, a lista de CSVs é lida em blocos de 256 nomes e ordenada numa tabela temporária, sem crescer no heap. Na prévia, a fila de arquivos é ordenada externamente em blocos limitados e os registros temporários não guardam identificadores de origem em texto puro.

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
