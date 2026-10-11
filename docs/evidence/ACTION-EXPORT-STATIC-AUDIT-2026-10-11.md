# Auditoria estática de ações e exportações — 2026-10-11

## Escopo

Leitura das referências T13–T15 preservadas em `docs/source/`, das decisões vigentes e do catálogo de chamadas Wails/IPC. Nenhuma tela, botão, regra financeira, permissão ou operação foi alterada ou executada. Os números das imagens são exemplos visuais, não dados reais nem valores de teste.

## Decisões já confirmadas

- As referências aprovadas das telas T13, T14 e T15 mostram o botão **Exportar**. L-22 e Q-13 estão definidos: preservar a função de exportação aprovada e detalhar tecnicamente sua execução, sem perguntar novamente se o botão deve existir.
- A responsável aprovou que `ADMIN` e `DIRECTOR` possam exportar relatórios; `REPORT_EXPORT` foi incluída para esses perfis pela migration `000003_director_report_export_permission.up.sql`. A verificação efetiva da capacidade no worker continua pendente em BK-013.
- F-14 exige que o arquivo use os mesmos dados, filtros, período, snapshot e métricas da tela, e preserva Excel quando a referência o especificar. A aprovação geral não autoriza criar dados, colunas, filtros ou formatos que não estejam definidos.
- T11 Compras é a exceção expressa: somente consulta, sem exportação ou download.

## Matriz parcial das referências T13–T15

| Tela | Ação e dados visíveis na referência aprovada | Caso de uso/capacidade | Situação técnica e limites |
|---|---|---|---|
| **T13 — Lançamentos e Pendências** | O cabeçalho tem **Exportar**. A tabela mostra Módulo, Origem, Descrição, Status, Responsável e Ação; há links como “Ver lista completa”, “Ver divergências”, “Ver mapeamento” e “Ver decisão”. | Exportar o relatório aprovado com `REPORT_EXPORT`; abrir lista e detalhes em leitura com `DASHBOARD_READ`. “Ver” não fecha pendência nem altera o Bling. | A opção do cabeçalho não está ligada a uma rota Wails/IPC de relatório. O período/filtro efetivo, o conjunto de linhas, o nome/formato do arquivo e o contrato de detalhe ainda não estão mapeados. A transição de pendências segue L-23 por tipo; não implementar resolução genérica. |
| **T14 — Alertas** | O cabeçalho tem **Exportar**. A tabela mostra Módulo, Alerta, Impacto, Prioridade, Atualizado e Status; os recortes oferecem “Ver detalhes”. | Exportar o relatório aprovado com `REPORT_EXPORT`; consultar alerta e origem em leitura com `DASHBOARD_READ`. Prioridade e status continuam dimensões distintas; detalhe não resolve alerta. | A opção não está ligada a uma rota Wails/IPC de relatório. O intervalo dos alertas, o conjunto de linhas, o nome/formato do arquivo e o contrato de detalhe ainda precisam ser ligados ao snapshot correspondente. Não calcular novos limiares nem criar resolução automática. |
| **T15 — Importações e Integrações** | O cabeçalho tem **Exportar**. A tabela mostra Origem, Lote/arquivo, Data ref., Importação, Status, Aceitos, Rejeitados e Hash; os links incluem “Ver todos os lotes”, “Ver rejeições”, “Ver arquivo” e “Ver lotes”. | Exportar o relatório aprovado com `REPORT_EXPORT`; consultar lotes/erros com `INTEGRATION_READ`. Prévia, importação manual e sincronização continuam ações distintas, autorizadas por `INTEGRATION_SYNC`; exportar não pode iniciar nenhuma delas. | A opção do cabeçalho não está ligada a uma rota Wails/IPC de relatório. A paginação/escopo da lista, o conteúdo completo do hash, o nome/formato do arquivo e os contratos de detalhe ainda não estão mapeados. Não exibir caminho sensível nem segredo. |

As imagens exibem posição e última sincronização, mas não definem todos os filtros, campos de origem, ordenação, paginação nem as opções abertas do menu **Exportar**. Em T15 o hash de exemplo aparece abreviado; isso não aprova truncar ou expandir o hash no arquivo. Esses detalhes precisam ser derivados do contrato já aprovado ou confirmados antes da implementação. Não usar os valores de exemplo das imagens como fonte ou oráculo.

## Estado da implementação observado

- O código atual tem uma opção de exportação desativada no menu do cabeçalho e não possui rota Wails/IPC para exportação de relatórios. Logo, a função aprovada ainda não está implementada.
- A migration de `REPORT_EXPORT` registra a capacidade para `ADMIN` e `DIRECTOR`, mas isso não comprova enforcement no worker; ver auditoria de acesso e BK-013.
- A implementação de leitura e exportação deverá preservar a mesma população e contexto apresentados na respectiva tela. Uma falha ou ausência de dados não pode virar zero, nem a exportação pode disparar sincronização, importação ou mudança de estado.
- A matriz acima cobre as referências T13–T15. T02–T07 e o vínculo completo entre controles aprovados, consultas e contratos de exportação ainda precisam ser conferidos para concluir BK-012.

## Próxima etapa

Concluir o mapeamento dos controles de T02–T07 e confirmar, em T13–T15, os detalhes técnicos que as referências não especificam (escopo/lista exportada, filtros/período, formato e conteúdo completo). Depois, desenhar o contrato da rota Wails/IPC e vincular a capacidade aprovada. Não redesenhar nem retirar os botões das referências. T11 permanece sem exportação.

**Progresso de BK-012:** 15% (inventário inicial mais matriz técnica das referências T13–T15; ainda falta mapear T02–T07, fechar os contratos de detalhe/exportação e verificar capacidades). Nenhuma tela ou regra de negócio foi modificada.
