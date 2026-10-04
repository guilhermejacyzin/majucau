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
- Campos textuais acima de 1 MiB ou linha acima de 4 MiB fazem a leitura falhar
  sem enviar o texto excessivo. O cliente só entrega o snapshot ao Wails depois
  do quadro final e da decodificação JSON completa.
- O JSON resultante conserva o contrato `DashboardSnapshot` consumido pela
  interface existente. Não houve alteração de tela ou regra financeira.

## Validação pendente

Esta execução não encontrou `go`/`gofmt` no `PATH` nem nos caminhos padrão do
Windows, portanto não foi possível formatar ou compilar localmente. Os testes
não foram executados nesta execução. A CI Windows/PostgreSQL ainda precisa
compilar e validar o caminho novo antes de `DATA-04` poder avançar de `PARTIAL`.

## Escopo restante

O streaming do snapshot foi implementado, mas `DATA-04` continua parcial até a
CI passar, a auditoria dos outros fluxos terminar e a limpeza/ACL ser verificada
em VM Windows real.
