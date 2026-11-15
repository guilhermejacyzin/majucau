# T04 — período personalizado e detalhes de Contas a Pagar

**Data:** 15/11/2026  
**Estado:** publicada no branch de execução e aberta no protótipo para revisão visual; não homologada pela responsável.

## Regra aplicada

- A usuária pediu uma faixa “De/Até” que determina os dados dos cards e dos detalhes.
- O período é inclusivo: entram títulos com vencimento igual ou posterior a “De” e igual ou anterior a “Até”.
- A fonte é o relatório normalizado de Contas a Pagar do Bling, limitado à última carga diária disponível. Não usar a série pronta de Fluxo de Caixa nem registros de outras origens.
- O total do período soma os saldos em aberto das linhas elegíveis. O card “Vence na data” mostra a quantidade de títulos no mesmo intervalo. “Em atraso” soma os títulos vencidos antes do início do intervalo.
- Títulos com situação `5`, cancelada ou estornada são excluídos, seguindo o filtro da T02.
- Os cards de total/período e atraso abrem uma janela com os fornecedores que compõem cada número. O card de quantidade abre a mesma lista do período.
- A tabela sem filtro não fica mais na página principal; o painel de distribuição por status também foi retirado da página principal, conforme a instrução sobre mover a composição para detalhes e remover os blocos inferiores.

## Causa observada

Na tela aberta em 15/11, os cards mostravam traços porque a configuração visual da T04 não ligava seus novos títulos aos dados do snapshot. Ao mesmo tempo, a página desenhava todas as linhas disponíveis sem aplicar intervalo, incluindo registros históricos de 2021 e situações numéricas canceladas. Não havia seletor de datas.

## Alterações locais

- Incluído seletor personalizado “De/Até”, iniciado na data-base corrente e com fim na janela de 60 dias disponível.
- Ligados os três cards à lista de Contas a Pagar filtrada por vencimento.
- Valores negativos agora ficam em vermelho em todas as telas, incluindo os cards da T04.
- Corrigida a edição do período: mudar “Até” preserva o “De” já exibido; mudar “De” preserva “Até” quando a ordem continua válida. A mesma regra está aplicada na T03.
- Os detalhes usam exatamente as linhas e somas do período selecionado; o detalhe de atraso usa vencimentos anteriores ao início.
- A consulta detalhada do snapshot agora seleciona apenas `BLING / bling.conta_pagar`, associada à última carga normalizada, e exclui cancelados/estornados.
- Removidos da tela principal a tabela permanente de obrigações e a distribuição genérica por status; fornecedores ficam em “Ver detalhes”.

## Limites e validação

- O protótipo foi aberto em T04. O intervalo exibido é 15/11/2026 a 13/01/2027; nessa carga local, os cards mostram -R$ 538.158,66, 328 títulos e -R$ 14.484,18 em atraso. O detalhe do card de total abriu e apresentou fornecedores com vencimentos no mesmo intervalo. A soma do card executivo “A PAGAR” para 60 dias também é -R$ 538.158,66.
- A tela aberta anteriormente explicava o erro da imagem: os cards ainda estavam sem vínculo com os dados e a tabela fixa ignorava o período. A tela atual não exibe essa tabela fixa.
- Os valores calculados passam a depender das linhas detalhadas contidas no snapshot ativo. Se a fonte não disponibilizar a lista ou estiver indisponível, os cards continuam sem valor; não se substitui por números de exemplo.
- Validação executada: `npm test -- --reporter=dot` (7 arquivos, 88 testes aprovados, incluindo a cor vermelha para valores negativos); `npm run lint`; `npm run build`; `go test ./...`; `go vet ./...`.
- O protótipo mostra a última sincronização já existente, 15/11/2026 às 11:03. Não foi iniciada nova sincronização do Bling. Falta a aprovação visual da responsável e a conferência financeira no relatório de origem para homologar a T04.
