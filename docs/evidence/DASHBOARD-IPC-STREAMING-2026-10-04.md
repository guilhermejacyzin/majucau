# Streaming do snapshot do painel — 2026-10-04

## Decisão aprovada

Em 2026-10-04, a usuária autorizou enviar os dados do painel em blocos menores,
mantendo a mesma tela e as mesmas informações.

## Implementação

- O protocolo IPC passou para v2 e acrescentou sequência, sinalizador de fim e
  conteúdo binário em cada quadro de resposta.
- Cada quadro de streaming carrega no máximo 512 KiB de conteúdo, abaixo do
  limite geral de 1 MiB do Named Pipe. O cliente valida versão, request id,
  sequência, limite e quadro final.
- O worker consulta os valores do resumo e as linhas de detalhe no PostgreSQL
  dentro de uma transação somente leitura `REPEATABLE READ`.
- Recebíveis e pagáveis são lidos com cursor `Rows.Next`, até 50 linhas por
  tabela, e codificados para o fluxo uma linha por vez.
- Campos textuais acima de 4 KiB ou linha acima de 16 KiB fazem a leitura falhar
  sem enviar o texto excessivo. O cliente só entrega o snapshot ao Wails depois
  do quadro final e da decodificação JSON completa.
- O JSON resultante conserva o contrato `DashboardSnapshot` consumido pela
  interface existente. Não houve alteração de tela ou regra financeira.

## Validação local — 2026-10-04

- O Go 1.26.6 definido em `go.mod` foi instalado de forma portátil em
  `artifacts/cache/dev/go-1.26.6`; o arquivo oficial foi conferido pelo SHA-256
  publicado em `go.dev/dl`.
- `go test ./internal/dashboard ./cmd/worker` passou. Os testes conferem o JSON
  final, transação somente leitura `REPEATABLE READ`, datas de negócio, limites
  de ambos os tipos de tabela e rejeição de linha excedente sem gravar a linha.
- A CI Windows/PostgreSQL ainda precisa compilar e validar a mudança antes de
  atualizar `DATA-04` para além de `PARTIAL`.

## Revalidação arquitetural

- O método legado `ReadDashboardSnapshot`, que fazia agregação e decodificação
  integral das linhas, foi removido. O contrato de aplicação agora exige
  `WriteDashboardSnapshot`; o worker não aceita uma implementação sem streaming.
- As consultas de detalhe continuam limitadas a 50 linhas por tabela. Cada
  campo tem limite de 4 KiB e cada linha, de 16 KiB; os dois fluxos rejeitam
  explicitamente registro excedente antes de escrevê-lo.

## Escopo restante

O streaming do snapshot foi implementado e testado localmente, mas `DATA-04`
continua parcial até a CI passar, a auditoria dos outros fluxos terminar e a
limpeza/ACL ser verificada em VM Windows real.

## Fechamento da rota sem streaming na fronteira Wails — 2026-10-04

A revisão encontrou uma alternativa em `app.go`: se o cliente IPC não
implementasse `StreamClient`, o painel ainda chamava `Client.Call` e decodificava
o snapshot JSON inteiro. A alternativa foi removida. O painel agora exige
`StreamClient`; se o cliente não oferecer quadros, retorna
`WORKER_STREAM_UNAVAILABLE` sem tentar a chamada não-streaming. A tela e os
dados esperados não mudaram.

Verificação local direcionada no Windows:

```text
go test app.go app_test.go
```

Resultado: aprovado (`command-line-arguments`, 0.417 s). O novo caso confirma
que um cliente não-streaming é rejeitado antes de chamar o handler de dashboard.
Este teste da fronteira Wails não substitui a suíte completa, a CI
Windows/PostgreSQL nem a validação em VM Windows limpa. `DATA-04` continua
`PARTIAL` até esses gates e a auditoria dos demais fluxos/ACLs serem concluídos.

## Revalidação após integração — 2026-10-04

- A execução [CI 37243820980](https://github.com/guilhermejacyzin/majucau/actions/runs/37243820980), no commit `ff9e5ce` (somente documentação), terminou com `verify-postgres` aprovado e `verify-windows` reprovado na etapa `Go tests`. O build da interface passou antes dessa etapa; verificações posteriores do job Windows foram ignoradas.
- A reprodução local de `TestNamedPipeHealthRoundTripAndCancellation` também falhou em `internal/ipc/namedpipe_windows_test.go:53` com `Acesso negado.`. Um diagnóstico separado confirmou que o Windows consegue abrir um Named Pipe com permissões padrão; isso não identifica por que o canal com DACL restrita está negando acesso.
- A leitura do diff confirmou que a revisão de streaming adicionou o caminho de quadros, mas não mudou as funções que montam e aplicam o DACL. A causa exata ainda não está provada; não ampliar permissões como tentativa de fazer os testes passarem.
- Os logs detalhados da execução não foram disponibilizados pela API pública do GitHub (`403 Forbidden`). É necessário obter a saída detalhada de `Go tests` para comparar todos os testes que falharam.
- `DATA-04` permanece `PARTIAL`; o streaming, por si só, não está homologado até corrigir a falha e obter CI verde.

## Correção local do Named Pipe — 2026-10-04

- A reprodução local confirmou duas causas no transporte Windows: a abertura do
  cliente era negada com a ACL mínima anterior; depois de incluir
  `FILE_READ_ATTRIBUTES`, a leitura síncrona usada para detectar desconexão
  podia bloquear a escrita da resposta no mesmo pipe.
- O servidor agora cria instâncias com `FILE_FLAG_OVERLAPPED` e aguarda
  `ConnectNamedPipe` por evento, cancelando a conexão pendente quando o contexto
  termina. A UI continua com direitos mínimos e não recebe
  `FILE_CREATE_PIPE_INSTANCE`; esse direito permanece no SID do serviço.
- A leitura que detecta desconexão foi mantida para que cancelamento do cliente
  chegue ao handler. O protocolo e os quadros limitados de streaming não foram
  alterados.
- `go test ./internal/ipc -count=1 -v` passou localmente com Go 1.26.6 no
  Windows. A execução incluiu ida e volta de saúde, oito chamadas concorrentes,
  fechamento do servidor durante a conexão, cancelamento do cliente até o
  handler, streaming de mais de um quadro, validação de DACL e
  rejeição/autorização de processo.
- A mudança ainda precisa da CI Windows/PostgreSQL. `DATA-04` permanece
  `PARTIAL` até essa validação e a auditoria dos demais fluxos/ACLs em VM real.

## Limite do leitor Wails — 2026-10-10

- A fronteira Wails já recebia o snapshot por IPC em blocos, mas o decoder não
  tinha limite total de bytes e aceitava dados após o primeiro documento JSON.
- `app.go` agora limita a leitura do snapshot a 16 MiB e rejeita conteúdo acima
  do teto ou JSON adicional após o documento esperado. Esse teto considera as
  duas tabelas com até 50 registros de até 16 KiB cada; inclui margem para a
  expansão de caracteres especiais no JSON e para o resumo fixo. A resposta
  inválida usa o código público já existente `WORKER_INVALID_RESPONSE`.
- Nenhum campo, cálculo, regra contábil ou elemento de tela foi alterado. O
  limite protege a entrada do processo Wails e preserva o contrato do snapshot.
- `go test app.go app_test.go` e `go vet app.go app_test.go` passaram com Go
  1.26.9; os testes cobrem JSON dentro do limite, excesso de bytes e documento
  JSON adicional. A tentativa de `go test .` não chegou à compilação porque o
  diretório gerado `frontend/dist` estava ausente; o script de verificação do
  projeto prepara esse diretório e a CI completa do commit de código ainda é
  necessária.
- `DATA-04` continua `PARTIAL` até a CI Windows/PostgreSQL do código atualizado,
  a conclusão da auditoria de todos os fluxos e a verificação em VM Windows
  limpa.

## CI Windows e PostgreSQL do limite Wails — 2026-10-10

- A CI [38076672784](https://github.com/guilhermejacyzin/majucau/actions/runs/38076672784)
  do commit `d0462b8` concluiu com sucesso nos jobs `verify-windows` e
  `verify-postgres`.
- No Windows, a execução passou pelos testes e `go vet`, verificações do
  frontend, builds do aplicativo/worker e instalador, smoke do instalador e
  consolidação dos outputs. No PostgreSQL, passou pelo schema, pacote de backup
  e testes de integração.
- Isso valida o limite de 16 MiB e a rejeição de JSON adicional no processo de
  CI. Não comprova pico de memória em execução nem substitui a validação de
  ACL/limpeza em VM Windows 10/11 limpa.
- `DATA-04` permanece `PARTIAL` até a auditoria completa dos fluxos e os gates
  operacionais em VM limpa.
