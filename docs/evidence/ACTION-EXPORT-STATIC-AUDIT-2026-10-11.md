# Auditoria estática de ações e exportações — 2026-10-11

## Escopo

Leitura do código de apresentação e do catálogo de chamadas IPC/Wails comparada às decisões aprovadas e aos documentos de referência. Nenhuma tela, botão, regra financeira, permissão ou operação foi alterada ou executada.

## O que está comprovado

| Área/controle | Estado observado | Limite ou decisão aplicável |
|---|---|---|
| Navegação do painel | A navegação abre as telas representadas pelo componente de mockup. | O acesso aos dados e ações de cada módulo continua em construção. |
| `Exportar` no cabeçalho | Abre um menu com uma opção desativada enquanto não há dados confirmados. | Não há método de exportação de relatório no catálogo atual do Wails/IPC; o botão não gera arquivo. |
| `Ver detalhes` dos cards | O controle aparece, mas não tem ação de abertura ligada no componente atual. | O drill-down até a origem ainda precisa de contrato e implementação. |
| Detalhes de fluxo, cálculo, distribuição e parâmetros | Os controles de detalhe existentes estão desativados. | Não foi implementada navegação para registros de origem. |
| T11 — Compras | “Gerar pedidos” permanece desativado e sem ação. A tela não oferece exportação ou download. | Decisão explícita: somente consulta; não criar/alterar pedidos no Bling. |
| T15 — Bling | Existem controles separados para salvar configuração, conectar, reconectar, testar, desconectar, abrir sincronização com filtros explícitos e pré-visualizar/importar CSV de recebimentos. | Importação e sincronização são ações distintas; a verificação efetiva de permissões por perfil ainda está pendente em BK-013. |
| T15 — Nuvem Pago | Existem prévia e importação manual de CSV de recebimentos futuros. | Esses dados permanecem projetados e não dão baixa como recebidos. A autorização por perfil ainda precisa ser aplicada no worker. |

## Decisões e lacunas

- Gisele aprovou que Diretores consultem, sincronizem e exportem backups e relatórios; também aprovou importar CSV do Bling e da Nuvem Pago dentro da sincronização. Isso define capacidades gerais, não os relatórios disponíveis em cada tela.
- Os registros L-22 e Q-13 continuam abertos para definir, nas telas T13–T15, se há exportação, quais dados e filtros entram, formato e permissões. As referências consultadas não fecham esses detalhes para todas as telas.
- Não foi encontrada uma rota de exportação de relatório no catálogo IPC/Wails atual. O botão de cabeçalho ainda não representa uma exportação funcional.
- Regras para resolver pendências também continuam abertas em L-23; não associar um botão de detalhe a uma ação que feche ou altere uma pendência sem contrato aprovado.
- Até receber a referência aprovada, manter a apresentação existente e não inventar formatos, colunas, downloads ou ações. T11 continua sem exportação.

## Próxima etapa

Completar a matriz tela por tela com as referências aprovadas para detalhes e exportação. Onde L-22/Q-13 ainda estiverem sem resposta, obter a definição funcional antes de implementar. Depois, vincular cada capacidade ao contrato do worker e à mesma consulta/filtros/snapshot exibidos na tela.

**Progresso de BK-012:** 10%, somente pelo inventário estático das ações atuais e das decisões encontradas. O critério de aceite da tarefa não está concluído.
