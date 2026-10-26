# Conferência da fonte dos cards CAPEX — 2026-10-11

## Escopo

Revisão local dos dois PDFs de Contas a Pagar enviados pela responsável e do
recorte de categorias compartilhado na conversa. Nenhum nome de fornecedor ou
valor individual foi copiado para este repositório.

## O que os arquivos mostram

- Na conferência inicial dos arquivos recebidos em 04/10, o relatório mensal
  cobria 01/09/2026 a 30/09/2026 e o geral cobria 01/01/2026 a 30/09/2026.
- O relatório detalhado atualizado, recebido em 11/10, cobre 01/01/2026 a
  30/11/2026; o arquivo geral continua cobrindo 01/01/2026 a 30/09/2026.
- Os dois relatórios detalhados mostram fornecedor, histórico, número do
  documento, vencimento, data de liquidação, situação e valor pago.
- Esses PDFs não mostram a categoria CAPEX em cada lançamento.
- O recorte separado de categorias mostra um total agrupado de CAPEX e
  subcategorias, mas não exibe período, data de liquidação nem número de
  documento para relacioná-lo às linhas dos PDFs.

## Consequência para os cards

As datas e os valores pagos dos relatórios detalhados, sozinhos, não permitem
separar CAPEX de despesas comuns. Não classificar um lançamento usando apenas o
fornecedor ou o texto do histórico. O card “CAPEX pago no mês” precisa de uma
fonte que relacione a categoria CAPEX com cada título e sua data de liquidação,
ou de uma chave de ligação aprovada entre relatórios.

Na redação original desta auditoria, o sentido financeiro do card “CAPEX
acumulado desde o início” ainda aguardava confirmação. Essa dúvida foi
respondida depois e está fechada no `G0-DECISION-REGISTER.md`: o card representa
investimentos em ativos reconhecidos como CAPEX desde o início do histórico
confiável disponível. O card “CAPEX pago no mês” continua limitado ao pagamento
efetivo no mês/período exibido e ligado a um ativo classificado como CAPEX.
Esta decisão substitui a pergunta antiga; não altera a regra aprovada.

Até essas condições serem atendidas, os cards permanecem sem valor confirmado.
Esta auditoria registra limites da fonte e não altera telas ou regras
financeiras aprovadas.

## Atualização posterior — fonte identificada e dúvida encerrada

A usuária já forneceu os relatórios do Bling e confirmou o acumulado desde o
início. Não é necessário pedir outro relatório para prosseguir. Os PDFs
detalhados não ligam cada conta à categoria CAPEX, mas a fonte técnica descrita
em `CAPEX-REPORT-STRUCTURE-REVIEW-2026-10-11.md` é a API oficial: cadastro de
categorias e movimentações por data/categoria. A implementação ainda precisa
homologar a árvore real de CAPEX da conta, associar lançamentos à categoria e
reconciliar os valores com os relatórios enviados. Até essa conferência, os
cards continuam mostrando dado indisponível. Nenhum dado financeiro real foi
copiado para esta evidência.

## Rechecagem dos arquivos disponíveis — 12/10/2026

Os dois PDFs foram conferidos novamente, incluindo a primeira página renderizada
e a estrutura textual das quatro páginas de cada arquivo. O relatório detalhado
mais recente mostra período de 01/01/2026 a 30/11/2026; o relatório identificado
como geral mostra 01/01/2026 a 30/09/2026. Ambos apresentam as colunas de
fornecedor, histórico, vencimento, liquidação, situação e valor pago. Nenhum
dos PDFs contém a categoria nos registros.

O print separado continua comprovando que existe uma visão agrupada de CAPEX e
subcategorias. Ele não informa o período do filtro nem associa cada categoria a
um documento, data de liquidação e valor pago. Portanto, os materiais confirmam
que CAPEX aparece no Bling, mas ainda não permitem conciliar o total agrupado
com os pagamentos individuais nem calcular os dois cards com segurança.

Não foram copiados nomes, números de documentos ou valores individuais. A regra
dos cards e a tela permanecem iguais; os valores seguem indisponíveis até haver
uma ligação confiável entre a categoria e cada pagamento, pela API homologada ou
por um relatório detalhado equivalente.

## Caminho de associação no esquema atual — 12/10/2026

A revisão do modelo local identificou um caminho possível, mas ainda não implementado:

- `payables.category` guarda a categoria da conta a pagar.
- `payments` guarda data e valor pagos, mas não guarda a categoria.
- `payment_allocations` pode ligar um pagamento à conta a pagar correspondente.
- A busca no adapter Bling não encontrou persistência de pagamentos, movimentações de caixa, nem preenchimento desse vínculo.

Assim, o valor CAPEX pago só poderá ser separado com segurança após importar o pagamento e sua ligação idempotente à conta categorizada. O acumulado continua seguindo a decisão registrada: ativos reconhecidos como CAPEX desde o início do histórico confiável. Não somar todos os pagamentos, nem usar fornecedor ou histórico como classificação. Até haver dados vinculados e conferidos com os relatórios, manter ambos os cards indisponíveis.

## Início aprovado para o histórico acumulado — 26/10/2026

Gisele aprovou `01/01/2026`, a data inicial do relatório geral, como início do histórico do card `CAPEX ACUMULADO`. A rotina de consulta já usa essa data como limite inferior. Isso fecha a data de início, mas não identifica por si só os lançamentos CAPEX: o card permanece indisponível até os registros da fonte serem associados à categoria e conciliados com datas e valores do relatório.
