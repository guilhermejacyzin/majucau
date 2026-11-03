# T02 — protótipo atualizado com o quadro Semana Crítica — 03/11/2026

## O que foi corrigido

O worker que estava ativo havia sido iniciado às 07:00, antes do commit que
incluiu `critical_week` no snapshot. Por isso, embora a tela já tivesse o
quadro no código, o serviço não enviava os dados para preenchê-lo.

Compilei a revisão atual em uma cópia isolada dentro de `artifacts`, atualizei o
worker e abri a versão Wails atualizada. A instância antiga foi encerrada depois
que a nova janela abriu normalmente. O worker novo iniciou com o PostgreSQL
Majucau isolado conectado; a verificação de saúde retornou `OK`.

## Conferência real pelo IPC local

Em 03/11/2026 às 08:06 BRT, o snapshot lido do worker retornou:

- Entradas em 30 dias: **R$ 69.994,12 — Parcial**.
- Saídas em 30 dias: **-R$ 262.136,32 — Parcial**.
- Série: 61 dias, agrupados em 9 semanas.
- Semana de maior saída: **15–21/11/2026**; entradas **R$ 37.383,21** e
  saídas **-R$ 152.735,18**, ambas parciais.
- Última sincronização do Bling: **03/11/2026 às 05:25 BRT**. A atualização do
  executável não buscou dados novos do Bling.

O contrato do worker agora inclui o quadro “Semana Crítica”. A janela Wails
atualizada abriu com o título normal **Majucau Financial Intelligence**. A
captura visual do conteúdo não foi feita nesta conferência.

## Validação

- `go test . ./internal/dashboard -count=1` — passou.
- `go vet . ./internal/dashboard` — passou.
- Interface: 70 testes passaram.
- `npm run build` — passou.
- Build desktop Wails 2.13.0 — passou.
- Build do worker Windows — passou.
- CI `41625219772` — passou integralmente: backlog, Windows, PostgreSQL,
  build do instalador e pacote portátil, verificações de segurança, e criação
  e restauração de backup.
- Leitura read-only via Named Pipe confirmou métricas, estados e período da
  semana crítica, sem imprimir nomes de clientes nem conteúdo RAW.

Os executáveis ativos foram mantidos em `artifacts/cache/dev/` para a tela
atualizada continuar aberta. Os logs de inicialização do worker permanecem em
`artifacts/verification/` enquanto o processo os mantém abertos. A remoção da
cópia isolada `artifacts/verification/majucau-live-refresh-20261103` foi
barrada pela revisão automática ao solicitar a exclusão recursiva. Ela continua
centralizada em `artifacts/` e fica registrada para limpeza manual posterior.

## Limites atuais

Os valores continuam `PARTIAL`; não estão homologados contra o relatório
completo do Bling. A última carga continua sendo às 05:25 BRT. O escopo de
leitura de Contas Financeiras foi autorizado pela responsável, mas ainda não
foi alterado no cadastro do aplicativo Bling nem houve nova autorização OAuth.
Embora a responsável tenha autorizado esse escopo somente de leitura, não foi
possível alterar permissões de uma conta online pela interface de controle
disponível nesta execução. Próximo passo operacional: a responsável habilita
somente a consulta a Contas Financeiras no cadastro do aplicativo Bling, salva
a configuração e autoriza novamente o Majucau. Depois disso, é preciso fazer
uma nova carga e comparar os dados com o relatório Fluxo de Caixa do Bling.
O saldo atual por banco continua sem fonte confirmada. Nenhuma tela aprovada ou
regra financeira foi alterada.
