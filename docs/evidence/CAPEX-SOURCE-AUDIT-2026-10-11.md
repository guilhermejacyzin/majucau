# Conferência da fonte dos cards CAPEX — 2026-10-11

## Escopo

Revisão local dos dois PDFs de Contas a Pagar enviados pela responsável e do
recorte de categorias compartilhado na conversa. Nenhum nome de fornecedor ou
valor individual foi copiado para este repositório.

## O que os arquivos mostram

- O relatório de setembro cobre 01/09/2026 a 30/09/2026.
- O relatório geral cobre 01/01/2026 a 30/09/2026.
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
