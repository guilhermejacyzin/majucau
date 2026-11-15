# T03 — detalhes de Contas a Receber

**Data:** 15/11/2026  
**Estado:** correções implementadas localmente; aguardando aprovação visual da responsável.

## Pedido registrado

- Tirar a tabela de títulos da página principal e mostrá-la ao clicar em “Ver detalhes” do card correspondente.
- Fazer os demais cards manterem sua própria ação de detalhes.
- Explicar o gráfico e usar categorias financeiras que a origem permita conferir.
- Não apagar o gráfico com base na pergunta sobre o seu significado.
- Remover os cards “Ambíguos” e “Chargebacks”, que não têm finalidade aprovada para esta tela.
- Substituir o card “Conciliação” por “TOTAL RECEBIDO”; seu valor é a soma de entradas conciliadas do relatório Caixa e Bancos no mês da data-base.
- Incluir no topo um período personalizado “De” e “Até” que controle os valores, o gráfico, a distribuição e os detalhes dos recebíveis.
- Fazer a distribuição da T01 espelhar a mesma distribuição líquida da T03, originada em Contas a Receber, e não no saldo do Itaú.

## Alterações

- A tabela principal de recebíveis foi retirada do fluxo visual da T03. “TOTAL A RECEBER” abre uma janela com Cliente, Origem, Vencimento, Valor bruto, Valor líquido e Status. O período De/Até controla os cards, o gráfico, a distribuição e os detalhes. “TOTAL A RECEBER”, PIX, Boletos e Cartão somam o líquido dos títulos com vencimento no intervalo; “EM ATRASO” considera vencimentos anteriores ao início; “TOTAL RECEBIDO” soma movimentos conciliados dentro do intervalo.
- “TOTAL RECEBIDO” abre Data, Conta, Origem e Valor. O total e o detalhe usam os movimentos de entrada cujos IDs constam no snapshot oficial `situacaoConciliacao=1`, vinculado à mesma carga diária de `Caixas e Bancos`, e filtram a data selecionada. Não há remoção de outro card de recebimentos além da substituição solicitada do card “Conciliação”.
- Auditoria da regra corrigiu a confusão entre os dois campos da API: `situacao=R` significa que o lançamento é um registro ativo; a conciliação é indicada separadamente por `situacaoConciliacao=1` e pela lista de IDs retornada por esse filtro. A consulta agora exige essa associação oficial para os totais e detalhes. A situação opcional não é usada como prova de conciliação nem exclui uma linha que esteja na lista oficial.
- A atualização do protótipo revelou e corrigiu outra falha: o detalhe de recebíveis parava porque a consulta tinha dois campos chamados `payment_method`. O nome associado à forma de pagamento agora tem alias próprio, eliminando a ambiguidade no banco real.
- A auditoria da tela reaberta encontrou um elo faltando entre o worker e a janela Majucau: `WailsDashboardSnapshot` descartava `receivables_by_method`. O backend já fazia o cálculo, mas os cards e a legenda recebiam traços. O DTO do aplicativo agora encaminha o campo, com teste de mapeamento e serialização.
- A conferência seguinte encontrou a divergência entre a soma por forma e “TOTAL A RECEBER”: o total de uma única origem era deduplicado contra registros da outra origem que não estavam atualizados. O cálculo do subtotal Bling agora usa seus registros elegíveis sem deduplicar contra uma origem Nuvem Pago desatualizada; a distribuição usa a mesma seleção da origem atual e aplica a deduplicação cruzada apenas quando ambas estão atualizadas. Um teste PostgreSQL verifica ambos os cenários e exige que PIX + boleto + cartão + não identificados feche exatamente com “TOTAL A RECEBER”.
- As listas do dashboard aceitam até 10.000 linhas por tipo, dentro do limite de 16 MiB do snapshot. A lista de recebíveis agora prioriza vencimentos a partir da data-base e depois os atrasados mais recentes, para que milhares de títulos antigos não ocultem as semanas selecionadas. O detalhamento por período usa a mesma lista e os valores dos cards são recalculados nela.
- O nome da forma de pagamento é mapeado no backend para somar valores líquidos de PIX, boleto, cartão e não identificados sobre o conjunto completo deduplicado. Os três cards inferiores e o gráfico usam esses mesmos totais, sem depender do limite de 200 linhas da janela de detalhes.
- “Ver detalhes” dos outros cards permanece acionável. Quando o snapshot não possui uma lista de origem para aquele indicador, a janela informa essa ausência e não fabrica linhas.
- O gráfico de distribuição permanece na T03 e agora mostra a participação por forma de recebimento, em valor líquido. “Não identificados” reúne as linhas sem uma forma reconhecida como PIX, boleto ou cartão. Quando não há dados atuais, o círculo fica neutro e não sugere percentuais.
- A distribuição exibida na T01 usa a mesma agregação de Contas a Receber da T03 e compartilha o período enquanto a navegação permanece aberta. Ela mostra PIX, Boletos, Cartão e Não identificados, sem usar saldo bancário do Itaú ou categorias genéricas de conciliação.
- Os cards “Ambíguos” e “Chargebacks” foram removidos da faixa inferior. PIX, Boletos e Cartão permanecem com valores líquidos e detalhes sob demanda.

## Validação

- O teste de integração existente foi atualizado para exigir o snapshot oficial de conciliação e conferir que apenas IDs presentes nele entram nos totais, mesmo com `situacao` opcional ausente.
- `go test ./internal/dashboard -count=1` passou; os testes focados de sincronização/normalização do Bling passaram.
- `npm test -- --run src/T03ReceivablesDetails.test.tsx` cobre o intervalo De/Até, os totais líquidos por forma, os detalhes sob demanda, o espelhamento T01/T03 e a ausência dos cards Ambíguos e Chargebacks.
- `npm run build` passou (TypeScript e Vite).
- Teste da camada Wails confirma o encaminhamento e a serialização de `receivables_by_method`.
- `go test ./internal/dashboard -run TestReadStreamSummaryPostgresRequiresNormalizedDailyStages -count=1` passou contra o banco isolado do protótipo; a divisão por forma reconcilia com o total no cenário de integração.
- A versão local da T03 foi reaberta para conferência visual; a aprovação final da responsável ainda está pendente.

## Estado da sincronização e limite observado

- A tabela principal de títulos fica recolhida. PIX, Boletos e Cartão mantêm a ação “Ver detalhes”; Ambíguos e Chargebacks não aparecem. A primeira reabertura revelou o DTO Wails sem a divisão calculada; uma conferência seguinte também apontou a diferença de escopo entre a origem atual e a desatualizada. Ambas as falhas foram corrigidas. Na T03 reconstruída, “TOTAL A RECEBER” e o total do gráfico são R$ 363.381,42; PIX R$ 26.372,06 + Boletos R$ 74.391,26 + Cartão R$ 222,79 + Não identificados R$ 262.395,31 = R$ 363.381,42. Essa tela mostra o snapshot local datado de 15/11/2026 11:03; a conferência não acionou uma nova sincronização do Bling.
- O botão de atualização foi acionado durante a conferência. A tentativa atual do Bling falhou para recebíveis com `BLING_API_SCOPE_INSUFFICIENT` (0 recebíveis lidos) e terminou parcial; por isso, a tela conserva o último snapshot publicado, datado de 15/11/2026 02:13. Não considerar esse snapshot como uma nova importação bem-sucedida.
- Após a correção da consulta SQL ambígua (`payment_method`), o snapshot do dashboard voltou a carregar no protótipo. O card “TOTAL RECEBIDO” e sua lista usam exclusivamente IDs do snapshot oficial de conciliação; o último snapshot disponível mostra R$ 0,00 e nenhuma linha para esse indicador. A causa exata é a ausência de entradas conciliadas nessa carga publicada, não um valor inventado ou o valor bruto de contas a receber.
- Para que a atualização diária produza novos recebíveis, a conexão precisa ter o escopo de leitura de contas a receber. A conexão atual não o concedeu; será necessário reautorizar a integração com esse escopo antes de declarar os novos valores como atualizados.
- O botão “Atualizar dados da tela” agora preserva e mostra o motivo de uma atualização parcial. Para o erro de escopo, informa que a leitura de Contas a Receber não foi autorizada, que esses valores continuam vindo da última atualização concluída e que a conexão precisa ser reautorizada. Teste dedicado em `App.test.tsx` passou (38 testes do arquivo) e `npm run build` passou.
- A atualização da T03 agora chama somente as fontes que alimentam essa tela: Contas a Receber, movimentos de Caixa e Bancos, snapshot de conciliação, formas de pagamento e contas financeiras. A carga detalhada de Contas a Pagar continua disponível na atualização completa das outras telas, sem bloquear a atualização da T03. Os critérios de cálculo e as fontes financeiras não mudam.
- A legenda do gráfico explica PIX, boletos, cartão e não identificados com os totais líquidos recebidos no snapshot. O círculo permanece neutro quando esses dados estão indisponíveis. Os cards “Ambíguos” e “Chargebacks” foram removidos da T03 conforme pedido.

## Pendência de aprovação

A T03 corrigida está aberta no protótipo para conferência visual e aprovação da usuária. O acesso de leitura da API para recebíveis permanece como dependência externa e foi registrado separadamente da correção local.

Os resultados da T02 e a sincronização Bling permanecem conforme o estado documentado no README; esta evidência não declara a carga de Contas a Pagar atualizada nem a T02 homologada.
