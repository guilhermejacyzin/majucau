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
