# Auditoria estática de ações e exportações — 2026-10-11

## Escopo

Leitura das referências T13–T15 preservadas em `docs/source/`, das decisões vigentes e do catálogo de chamadas Wails/IPC. Nenhuma tela, botão, regra financeira, permissão ou operação foi alterada ou executada. Os números das imagens são exemplos visuais, não dados reais nem valores de teste.

## Decisões já confirmadas

- As referências aprovadas das telas T13, T14 e T15 mostram o botão **Exportar**. L-22 e Q-13 estão definidos: preservar a função de exportação aprovada e detalhar tecnicamente sua execução, sem perguntar novamente se o botão deve existir.
- Em 11/10/2026, Gisele aprovou Excel para T13–T15, com as linhas da tabela e a data exibida na tela. A decisão foi registrada em `G0-DECISION-REGISTER.md`; ela não define detalhes de filtros, escopo/paginação ou consultas de detalhe.
- A responsável aprovou que `ADMIN` e `DIRECTOR` possam exportar relatórios; `REPORT_EXPORT` foi incluída para esses perfis pela migration `000003_director_report_export_permission.up.sql`. A verificação efetiva da capacidade no worker continua pendente em BK-013.
- F-14 exige que o arquivo use os mesmos dados, filtros, período, snapshot e métricas da tela, e preserva Excel quando a referência o especificar. A aprovação geral não autoriza criar dados, colunas, filtros ou formatos que não estejam definidos.
- T11 Compras é a exceção expressa: somente consulta, sem exportação ou download.

## Matriz parcial das referências T02–T07 e T13–T15

### T02–T07 — telas financeiras

| Tela | Ação e dados visíveis na referência aprovada | Caso de uso/capacidade | Situação técnica e limites |
|---|---|---|---|
| **T02 — Fluxo de Caixa** | **Exportar** no cabeçalho. Posição/sincronização, saldo atual, entradas/saídas e menores saldos em 30/60 dias, série do gráfico, semana crítica, recebido/pago no mês, reserva, saldo transferível e alertas. | Exportar os mesmos indicadores e dados da posição exibida com `REPORT_EXPORT`; consultar cartões, gráfico e detalhes com `DASHBOARD_READ`. | O handoff registra que o menu não define formatos; a imagem não define filtros de período. A posição e as janelas de 30/60 dias precisam permanecer explícitas no relatório. Não presumir que exportar executa sincronização. |
| **T03 — Contas a Receber** | **Exportar** no cabeçalho. Tabela: Cliente, Origem, Vencimento, Valor bruto, Valor líquido e Status; cartões de recebimento/atrasos, distribuição e alertas. | Exportar a consulta aprovada com `REPORT_EXPORT`; abrir recebimentos, atrasos, conciliação, lista e regras em leitura com `DASHBOARD_READ`. | A imagem não abre o menu nem define formato, filtros, escopo de paginação ou colunas adicionais. Previsões, recebimentos e estados continuam conforme as regras de fonte e conciliação; exportar não pode dar baixa nem conciliar. |
| **T04 — Contas a Pagar** | **Exportar** no cabeçalho. Tabela: Fornecedor, Documento, Vencimento, Valor, Categoria e Status; cartões de vencimento, atraso, divergência e alertas. | Exportar a consulta aprovada com `REPORT_EXPORT`; abrir detalhes, vencimentos, atrasos, conciliação e regras em leitura com `DASHBOARD_READ`. | A imagem não define formato, filtros, escopo de paginação nem se os cartões entram como resumo. Não transformar a exportação em pagamento, baixa ou alteração da obrigação. |
| **T05 — Conciliação** | **Exportar** no cabeçalho. Tabela: Venda Bling, Transação Nuvem, Cliente, Valor bruto, Recebimento e Status; cartões de conciliados/pendentes/ambíguos/divergentes e alertas. | Exportar a posição consultada com `REPORT_EXPORT`; abrir situação, transações, regras e divergências em leitura com `DASHBOARD_READ`. | Formato, filtros, escopo e colunas adicionais não estão definidos. Exportação e detalhes não confirmam, desfazem, reutilizam nem alteram um vínculo de conciliação. |
| **T06 — DRE** | **Exportar** no cabeçalho. Tabela por competência com Descrição e Valor; valores de resumo, composição e alertas. A decisão posterior também aprovou cards EBIT/EBITDA e CAPEX pago no mês/acumulado; os cards são informativos, e CAPEX fica fora da tabela e do lucro. | Exportar a DRE própria do Majucau com `REPORT_EXPORT`; consultar detalhes da linha e composição com `DASHBOARD_READ`. | O formato e o conteúdo exato do arquivo não estão especificados. A exportação deve respeitar as linhas existentes e as decisões posteriores sobre cards; CAPEX não entra no resultado. Componentes sem dados confiáveis continuam `—`, sem valor ilustrativo ou zero. |
| **T07 — Balancete** | **Exportar** no cabeçalho. Tabela: Conta, Descrição, Débito, Crédito, Saldo e Status; composição por grupo, cartões e alertas. | Exportar a posição contábil consultada com `REPORT_EXPORT`; abrir detalhes, divergências, grupos e saldos em leitura com `DASHBOARD_READ`. | Formato, filtros, escopo e colunas adicionais não estão definidos. Manter sinais, estados OPEN/DIVERGENT e ausência de saldo inicial conforme o contrato; exportar não cria ajuste nem saldo zero. |

As referências de T02–T07 confirmam que o botão **Exportar** existe, mas os menus aparecem fechados. PDF e Excel continuam sem aprovação específica para essas seis telas. A decisão posterior aprovou Excel com tabela e data exibida somente para T13–T15. `G0-DECISION-REGISTER.md`, as imagens preservadas e F-14 definem que os arquivos devem acompanhar os dados, filtros, período, snapshot e métricas da tela; os limites de cada tela acima continuam pendentes onde o material não especifica o detalhe.

Na T06, seguir as decisões posteriores registradas no `G0-DECISION-REGISTER.md`: os cards aprovados substituem o card inferior UNMAPPED, o alerta de pendências permanece, e CAPEX permanece separado da tabela/lucro. A imagem T06 arquivada é uma versão anterior e não deve prevalecer sobre essas decisões posteriores.

### T13–T15 — exceções, alertas e integrações

| Tela | Ação e dados visíveis na referência aprovada | Caso de uso/capacidade | Situação técnica e limites |
|---|---|---|---|
| **T13 — Lançamentos e Pendências** | O cabeçalho tem **Exportar**. A tabela mostra Módulo, Origem, Descrição, Status, Responsável e Ação; há links como “Ver lista completa”, “Ver divergências”, “Ver mapeamento” e “Ver decisão”. | Exportar em Excel as linhas da tabela e a data exibida na tela, com `REPORT_EXPORT`; abrir lista e detalhes em leitura com `DASHBOARD_READ`. “Ver” não fecha pendência nem altera o Bling. | A opção do cabeçalho ainda não está ligada a uma rota Wails/IPC de relatório. Filtros e escopo/paginação das linhas, além do contrato de detalhe, permanecem pendentes. A transição de pendências segue L-23 por tipo; não implementar resolução genérica. |
| **T14 — Alertas** | O cabeçalho tem **Exportar**. A tabela mostra Módulo, Alerta, Impacto, Prioridade, Atualizado e Status; os recortes oferecem “Ver detalhes”. | Exportar em Excel as linhas da tabela e a data exibida na tela, com `REPORT_EXPORT`; consultar alerta e origem em leitura com `DASHBOARD_READ`. Prioridade e status continuam dimensões distintas; detalhe não resolve alerta. | A opção ainda não está ligada a uma rota Wails/IPC de relatório. Filtros, escopo/paginação e a associação do conjunto exportado ao snapshot correspondente permanecem pendentes. Não calcular novos limiares nem criar resolução automática. |
| **T15 — Importações e Integrações** | O cabeçalho tem **Exportar**. A tabela mostra Origem, Lote/arquivo, Data ref., Importação, Status, Aceitos, Rejeitados e Hash; os links incluem “Ver todos os lotes”, “Ver rejeições”, “Ver arquivo” e “Ver lotes”. | Exportar em Excel as linhas da tabela e a data exibida na tela, com `REPORT_EXPORT`; consultar lotes/erros com `INTEGRATION_READ`. Prévia, importação manual e sincronização continuam ações distintas, autorizadas por `INTEGRATION_SYNC`; exportar não pode iniciar nenhuma delas. | A opção ainda não está ligada a uma rota Wails/IPC de relatório. A paginação/escopo das linhas e os contratos de detalhe ainda não estão mapeados. Não exibir caminho sensível nem segredo. |

As imagens exibem posição e última sincronização, mas não definem todos os filtros, campos de origem, ordenação, paginação nem as opções abertas do menu **Exportar**. Em T15 o hash de exemplo aparece abreviado; isso não aprova truncar ou expandir o hash no arquivo. Esses detalhes precisam ser derivados do contrato já aprovado ou confirmados antes da implementação. Não usar os valores de exemplo das imagens como fonte ou oráculo.

## Estado da implementação observado

- O código atual tem uma opção de exportação desativada no menu do cabeçalho e não possui rota Wails/IPC para exportação de relatórios. Logo, a função aprovada ainda não está implementada.
- A migration de `REPORT_EXPORT` registra a capacidade para `ADMIN` e `DIRECTOR`, mas isso não comprova enforcement no worker; ver auditoria de acesso e BK-013.
- A implementação de leitura e exportação deverá preservar a mesma população e contexto apresentados na respectiva tela. Uma falha ou ausência de dados não pode virar zero, nem a exportação pode disparar sincronização, importação ou mudança de estado.
- A matriz acima cobre as referências T13–T15. T02–T07 e o vínculo completo entre controles aprovados, consultas e contratos de exportação ainda precisam ser conferidos para concluir BK-012.

## Próxima etapa

Concluir o mapeamento dos controles de T02–T07, incluindo formato e conteúdo, e definir para T13–T15 filtros, escopo/paginação e contratos de detalhe. Depois, desenhar o contrato das rotas Wails/IPC e vincular a capacidade aprovada. Não redesenhar nem retirar os botões das referências. T11 permanece sem exportação.

**Progresso de BK-012:** 25% (inventário e matriz parcial de T02–T07 e T13–T15; formato/conteúdo aprovados para T13–T15: Excel com tabela e data exibida. Faltam detalhes de filtros/escopo, contratos de detalhe, formato/conteúdo de T02–T07, rotas e enforcement de capacidades no worker). Nenhuma tela ou regra de negócio foi modificada.
