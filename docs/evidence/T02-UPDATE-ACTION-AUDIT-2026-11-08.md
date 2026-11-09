# Auditoria do comando Atualizar na T02 — 08/11/2026

## Pedido

Gisele esclareceu que, ao clicar em Atualizar, espera que o aplicativo busque os
dados recentes e atualize os resultados visíveis. Recarregar apenas os números
que já estavam guardados não atende ao pedido.

## O que o código faz hoje

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

## Conclusão

O botão de cabeçalho atualiza a visualização do retrato local, mas não atualiza
a origem financeira. Assim, depois de uma nova alteração no Bling, o clique pode
redesenhar os mesmos dados antigos. Este é um limite funcional confirmado pela
leitura do código e explica a expectativa frustrada da responsável.

## O que precisa ser implementado

1. Fazer a ação Atualizar solicitar uma sincronização oficial compatível com o
   módulo aberto, usando o mesmo worker e as credenciais protegidas existentes.
2. Durante a consulta, impedir clique duplicado e indicar que a atualização está
   em andamento.
3. Só após a conclusão bem-sucedida, reler o snapshot financeiro publicado e
   atualizar os números na tela.
4. Em falha, mostrar uma mensagem simples e manter a última cópia válida sem
   chamá-la de atualizada.
5. Informar o horário da carga concluída de verdade; uma consulta local ao
   snapshot não pode avançar o horário de sincronização.
6. Preservar idempotência, atomicidade, retry seguro, filtros e regras
   financeiras vigentes. Na T02, Receitas seguem Contas a Receber pelo líquido,
   Despesas seguem Contas a Pagar, e saldo segue o último saldo conciliado do
   Itaú.

## Aceite

- Clicar em Atualizar resulta numa nova tentativa observável de consulta à fonte
  autorizada, e não somente numa leitura local.
- O estado da interface distingue em andamento, sucesso e falha.
- Com o mesmo período e filtros do Bling, a semana 22/11–28/11 deve ser
  comparada com as referências já registradas: 146 recebíveis, R$ 43.790,66
  líquidos e 14 contas a pagar, R$ 25.275,71.
- A atualização deve ser conferida visualmente no protótipo e, em caso de
  divergência, detalhada por registro, vencimento, status e tarifa antes da
  homologação.

## Estado desta auditoria

Foi feita apenas leitura do código e atualização da documentação. O botão ainda
não foi ligado ao fluxo de sincronização, não houve nova carga real do Bling e
os valores apresentados na janela não foram confirmados. Não marcar T02 como
homologada. O trabalho continua dentro de BK-062; o próximo passo é implementar
o fluxo de atualização, validar os testes pertinentes, publicar e conferir os
valores no protótipo.
