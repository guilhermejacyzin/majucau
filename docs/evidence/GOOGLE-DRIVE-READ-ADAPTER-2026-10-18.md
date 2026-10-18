# Google Drive — adapter local de leitura — 18/10/2026

## Resultado

Foi implementado o núcleo Go de leitura somente GET para a API oficial do Drive em `internal/integrations/googledrive`, incluindo um provedor de tokens de service account que lê a chave do cofre de segredos do worker. O cliente só pode ser criado com exatamente três IDs de pasta distintos; a construção de produção usa endpoints HTTPS fixos do Google. Nenhuma tela ou regra financeira foi alterada.

## Controles implementados

- Listagem paginada limitada a 100 itens por página e resposta JSON de controle limitada a 2 MiB.
- Filtro explícito para cada pasta autorizada e rejeição local de arquivos cuja pasta pai não está na lista.
- Retry comum do projeto para GETs repetíveis: até três tentativas, classificação de 408/429/5xx e falhas transitórias, respeito a `Retry-After` e cancelamento.
- Redirecionamentos HTTP desativados para impedir que um token Bearer seja encaminhado a outro domínio.
- Download copiado em fluxo para um destino de staging, limitado pelo tamanho máximo que o chamador aprova para aquele formato.
- Metadados consultados antes e depois; mudança de versão, tamanho, horário, hash ou pasta cancela a publicação.
- SHA-256 calculado enquanto o conteúdo é lido; checksum MD5 informado pelo Drive é conferido quando disponível.
- O staging só recebe `Commit` depois das verificações; falha técnica ou cancelamento aciona `Rollback`.
- Erros públicos não incluem token, nome de arquivo, URL com identificador ou corpo bruto de resposta.
- O provedor solicita apenas o escopo Google `drive.readonly`, guarda o token temporário em memória com validade curta e respeita cancelamento e timeout; a troca OAuth de token não usa retry.
- O transporte do token só permite POST ao endpoint de token configurado e não consegue enviar a afirmação JWT para um endereço de redirecionamento.

## Verificação executada

- `go test ./internal/integrations/googledrive ./internal/integrations/httpretry -count=1` — **PASS** em Windows.
- `go test ./internal/integrations/... -count=1` — **PASS** em Windows, incluindo Bling, Drive, política de retry e os demais adapters.
- `go vet ./internal/integrations/...` e `go mod verify` — **PASS**.
- `govulncheck v1.9.0 ./internal/integrations/googledrive ./internal/integrations/httpretry` — **nenhuma vulnerabilidade conhecida encontrada** nos caminhos analisados.
- Os testes usam servidor HTTP local e tokens simulados; cobrem filtro de pastas, paginação, retry 429, verificação de conteúdo, rollback após alteração do arquivo, limite de tamanho, bloqueio de redirecionamento e redação de erro de credencial.
- `git diff --check` — **PASS**.
- Nenhuma conexão ao Google Drive, credencial real, arquivo financeiro ou banco de produção foi usada.

## Incremento de leitura paginada — 18/10/2026

O cliente agora oferece `WalkFiles`, que entrega os arquivos um por vez ao processamento e mantém somente uma página limitada da API na memória. Cada chamada tem limite explícito de páginas e retorna um marcador para continuar em outra chamada; não existe limite acumulado de arquivos entre chamadas. O cliente detecta marcadores repetidos e interrompe a leitura em vez de entrar em ciclo.

O marcador só avança depois que todos os arquivos da página passam pelo callback. Se o processamento falhar no meio da página, a próxima tentativa relê essa mesma página. Por isso, o consumidor futuro deve salvar os resultados de modo transacional e idempotente, aceitando a repetição sem duplicar dados. A mensagem pública do erro do callback é sanitizada.

Verificações adicionais desta mudança:

- `go test ./internal/integrations/googledrive -count=1` — **PASS**.
- `go test ./internal/integrations/... -count=1` — **PASS**.
- `go vet ./internal/integrations/...` — **PASS**.
- A nova cobertura verifica continuação por marcador, repetição de página em falha, detecção de ciclo, limite e cancelamento. Ainda não há execução de CI para este incremento.
- CI do GitHub #39174646423 — **PASS** nos jobs `verify-windows`, `verify-postgres`, `verify-backlog` e `detect-changes`: https://github.com/guilhermejacyzin/majucau/actions/runs/39174646423.

Esta melhoria não configura credenciais nem IDs reais e não conecta o adapter aos jobs ou ao pipeline de importação. O resultado continua incompleto e não conclui BK-042.

## Callback transacional por página — 18/10/2026

Foi acrescentado `WalkFilePages` para integrações que precisam persistir cada página e seu marcador de continuação na mesma transação. Cada página tem no máximo 100 arquivos. O callback recebe os arquivos limitados e o próximo marcador; só depois de sucesso o leitor avança o progresso retornado. Se o callback falhar, a chamada preserva o marcador do início da página para repetição. O consumidor ainda precisa ser idempotente caso a gravação tenha sido confirmada, mas a resposta tenha se perdido.

`WalkFiles` permanece disponível para consumidores que preferem um callback por arquivo e mantém o mesmo comportamento de repetição segura. Testes cobrem avanço de marcador após callback bem-sucedido, falha sem avanço, erro sanitizado e proteção contra tokens cíclicos.

- `go test ./internal/integrations/... -count=1` — **PASS** localmente.
- `go vet ./internal/integrations/...` — **PASS** localmente.
- Verificação do backlog e `git diff --check` — **PASS** localmente.
- A CI deste incremento ainda está pendente.

Isso prepara a integração ao banco, mas não cria tabelas, job diário, conexão real ou interpretação de dados financeiros. BK-042 permanece incompleto.

## Dependência de autenticação

Foi adotado `golang.org/x/oauth2 v0.37.0`, biblioteca mantida pelo projeto Go, licença BSD-3-Clause, para interpretar e usar credenciais de service account conforme o fluxo Google documentado. A versão está acima da correção publicada para GO-2025-3488. O scan de vulnerabilidades desta execução não encontrou alertas nos pacotes analisados. Fontes: [documentação do pacote Google OAuth](https://pkg.go.dev/golang.org/x/oauth2/google) e [GO-2025-3488](https://pkg.go.dev/vuln/GO-2025-3488).

## O que continua pendente

- Configurar o projeto Google Cloud e criar a identidade técnica dedicada.
- Compartilhar somente as três pastas com essa identidade como Leitor e registrar os IDs reais.
- Ligar o provedor de token já implementado ao worker e ao cofre de segredos DPAPI/ACL em execução; provisionar com segurança a credencial da conta de serviço sem incluí-la no Git ou no chat.
- Integrar o cliente ao worker, aos jobs persistentes, ao staging seguro existente e ao pipeline de importação/idempotência.
- Aprovar contratos reais para CSV, TSV e JSON; XLS/XLSX continuam bloqueados até existir parser seguro avaliado e homologado.
- Homologar com dados sanitizados, medir memória/limpeza e testar em VM Windows limpa.
- O teste local `-race` não pôde iniciar porque este ambiente Windows está com CGO desativado; a CI Linux ainda deve executar a cobertura com race.

Este resultado é **implementado localmente, aguardando integração e validação real**. Não é prova de acesso ao Drive nem conclusão de BK-042/BK-043.

## Inventário persistente e retomada após falha — 18/10/2026

Foi criada a migration `database/migrations/000006_google_drive_file_inventory.up.sql` para guardar o progresso da leitura e os metadados do Drive. O runtime pode usar somente as permissões explicitamente listadas em `database/privileges/runtime.sql`; a auditoria correspondente foi ampliada em `database/privileges/verify_runtime.sql`.

- Cada página de até 100 itens é guardada com o marcador da próxima página na mesma transação. A carga percorre no máximo 64 páginas por chamada e continua em outra chamada quando necessário; o total do inventário não tem limite fixo no aplicativo.
- Cada payload temporário é limitado a 1 MiB e cada marcador a 4 KiB. IDs, escopos e referências de arquivos ficam limitados às três pastas aprovadas. Metadados e marcadores são tratados como sensíveis e o worker tem acesso mínimo necessário.
- As versões de metadados são imutáveis e deduplicadas por hash. Um ponteiro separado identifica a versão atual e o lote em que o arquivo apareceu. Arquivos ausentes numa nova leitura não são apagados; deixam de pertencer ao retrato atual quando o cursor passa para o lote completo.
- A leitura interrompida conserva checkpoints e não muda o retrato/cursor anterior. Na publicação final, metadados, ponteiros, cursor, sucesso do lote e limpeza dos checkpoints são confirmados juntos. Repetir um lote concluído com o mesmo identificador não duplica a publicação.
- O banco guarda somente metadados da listagem; nenhum conteúdo de arquivo é gravado nessa tabela. O streaming dos arquivos para o staging seguro e a integração com o importador permanecem pendentes.

Verificações desta etapa:

- `go test ./internal/integrations/... -count=1` — **PASS** nos conectores Bling, CSV, Drive, retry, Nuvem Pago e Nuvemshop.
- `go vet ./internal/integrations/...` — **PASS** localmente.
- `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate -f database/sqlc.yaml` — **PASS**; gerou os modelos Go das novas tabelas, versionados no commit `ef83bed`.
- `scripts/verify-migrations.ps1` e `git diff --check` — **PASS**.
- A CI completa `39177839087`, no commit `ef83bed` — **PASS** nos quatro jobs: `verify-windows`, `verify-postgres`, `verify-backlog` e `detect-changes`. PostgreSQL aplicou a migration, verificou permissões e executou testes de snapshot completo, repetição, falha, retomada, preservação do retrato anterior e arquivo ausente.
- A primeira tentativa identificou que os modelos gerados pelo SQLC ainda não estavam versionados. Eles foram gerados com SQLC v1.31.1, adicionados ao commit seguinte e a geração foi aprovada na CI completa. A integração PostgreSQL não foi executada localmente porque esta máquina não possui uma conexão descartável configurada.
- Nenhuma conexão ao Google Drive, credencial real, arquivo financeiro, tela aprovada ou regra financeira foi usada/alterada.

BK-042 passa de 47% para 52% estimados. Ainda faltam provisionar e compartilhar a identidade de leitura, informar os IDs reais das pastas por configuração protegida, ligar a sincronização ao worker diário das 09:30 e ao staging/importador, homologar contratos CSV/TSV/JSON e implementar parser XLS/XLSX seguro. A validação real e o aceite também continuam pendentes.
