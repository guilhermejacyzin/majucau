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
