# Auditoria do comando Atualizar na T02 — 08/11/2026

## Pedido

Gisele esclareceu que, ao clicar em Atualizar, espera que o aplicativo busque os
dados recentes e atualize os resultados visíveis. Recarregar apenas os números
que já estavam guardados não atende ao pedido.

## Auditoria inicial

- `frontend/src/mockupDashboard.tsx`: o ícone do cabeçalho chama a função
  recebida como `onRefresh`.
- `frontend/src/App.tsx`: `refreshDataViews` chama `adapter.getState()` e
  `snapshotAdapter.getSnapshot()` em paralelo e coloca esses resultados no
  estado React.
- Essas leituras consultam o worker/snapshot local. Elas não iniciam uma nova
  chamada de sincronização ao Bling.
- Em Integrações, “Sincronizar agora” abre o painel de filtros; a sincronização
  só começa depois da escolha do período/tipo de data e do clique em “Executar
  sincronização”. O worker também executa a carga automática aprovada às 09:30
  em `America/Sao_Paulo`.

Na revisão inicial, o botão apenas relia o snapshot local e não chamava a API.
Essa constatação descreve o estado anterior à correção registrada abaixo.

## Implementação concluída no código

- O botão chama `App.SyncBling` com a opção interna `full_refresh`.
- O método Wails encaminha uma requisição própria `bling.refresh` ao worker,
  com timeout de até dez minutos para a carga completa. O corpo IPC é vazio;
  nenhum filtro financeiro é inventado.
- O worker chama `BlingOAuthService.SyncNow`, que usa um identificador novo
  para cada solicitação e o mesmo bloqueio PostgreSQL da rotina diária. Assim,
  a carga manual espera a execução diária se ela estiver em andamento e não é
  descartada porque a carga das 09:30 já terminou.
- A carga usa o fluxo oficial completo já existente: Contas a Receber, Contas a
  Pagar, movimentos de Caixa e Bancos, movimentos conciliados, formas de
  pagamento e contas financeiras. Permanece somente leitura e usa o token
  protegido pelo worker.
- Em conclusão integral, a interface relê o retrato local publicado. Em falha
  integral, não relê o retrato; em conclusão incompleta, relê os recursos que
  concluíram e apresenta um aviso que não afirma que os dados antigos foram
  preservados. O texto “Parcial” não é acrescentado aos valores financeiros.
- O cabeçalho mostra o andamento, impede cliques simultâneos e exibe mensagem
  de falha/incompletude. O horário exibido continua vindo do horário real de
  sincronização gravado no retrato, não da consulta local.

## Limites e aceite pendente

- Testes de contrato confirmam que a opção interna escolhe o método IPC dedicado
  e que o worker aceita somente o payload vazio esperado.
- Isso não prova que a conta Bling real autorizada esteja conectada nesta
  instância, que a nova carga tenha sido executada ou que os valores batam com
  as telas de origem.
- A homologação exige executar Atualizar no protótipo e comparar o mesmo
  período/filtros com Contas a Receber (líquido) e Contas a Pagar, e conferir o
  saldo conciliado do Itaú. Não declarar BK-062 ou T02 concluída antes disso.

## Aceite de homologação real (ainda pendente)

- Clicar em Atualizar resulta numa nova tentativa observável de consulta à fonte
  autorizada, e não somente numa leitura local.
- O estado da interface distingue em andamento, sucesso e falha.
- Com o mesmo período e filtros do Bling, a semana 22/11–28/11 deve ser
  comparada com as referências já registradas: 146 recebíveis, R$ 43.790,66
  líquidos e 14 contas a pagar, R$ 25.275,71.
- A atualização deve ser conferida visualmente no protótipo e, em caso de
  divergência, detalhada por registro, vencimento, status e tarifa antes da
  homologação.

## Estado atual desta auditoria

A ação está implementada e `scripts/verify.ps1` passou em 08/11/2026,
incluindo os 75 testes de interface. A carga real, a conferência visual e a
equivalência dos valores continuam pendentes. BK-062 permanece em andamento.
