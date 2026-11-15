# T03 — detalhes de Contas a Receber

**Data:** 15/11/2026  
**Estado:** correções implementadas localmente; aguardando aprovação visual da responsável.

## Pedido registrado

- Tirar a tabela de títulos da página principal e mostrá-la ao clicar em “Ver detalhes” do card correspondente.
- Fazer os demais cards manterem sua própria ação de detalhes.
- Explicar o gráfico por situação dos recebíveis e o significado de cada status.
- Não apagar o gráfico com base na pergunta sobre o seu significado.
- Não remover cards da T03 sem uma instrução inequívoca. Os cards PIX, Boletos, Cartão, Ambíguos e Chargebacks permanecem.
- Substituir Conciliação por um único card “TOTAL RECEBIDO”; seu valor é a soma de entradas conciliadas do relatório Caixa e Bancos no mês da data-base, sem duplicar o antigo card mensal.

## Alterações

- A tabela principal de recebíveis foi retirada do fluxo visual da T03. “TOTAL A RECEBER” abre uma janela com Cliente, Origem, Vencimento, Valor bruto, Valor líquido e Status. O detalhe de “EM ATRASO” filtra pela data-base do snapshot.
- “TOTAL RECEBIDO” abre Data, Conta, Origem e Valor. As linhas vêm de `bling_cash_movements` da mesma carga diária publicada, com `direction = INFLOW`, status `R` e as mesmas datas mensais do agregado `receipts_month`. O antigo card duplicado “RECEBIDO NO MÊS” foi removido.
- Auditoria da regra encontrou uma inconsistência: a consulta incluía situação vazia como conciliada, apesar do normalizador marcar situação vazia/desconhecida como não confirmada. Corrigido: o total e os detalhes de recebimentos/pagamentos realizados agora incluem somente situação explicitamente `R`. Situação vazia ou não mapeada não entra no total; uma carga com esses registros continua identificada como incompleta.
- As listas do dashboard aceitam até 200 linhas. O limite e os limites por campo/registro continuam protegendo a leitura e a memória; se houver mais de 200 itens, o detalhe não substitui uma exportação completa.
- O nome da forma de pagamento do Bling é passado ao frontend para identificar títulos de PIX, boleto e cartão sem expor o payload RAW.
- “Ver detalhes” dos outros cards permanece acionável. Quando o snapshot não possui uma lista de origem para aquele indicador, a janela informa essa ausência e não fabrica linhas.
- O gráfico continua visível e as definições ficam visíveis sem clique. Confirmado: título/valor conferidos com a origem; Provisório: valor registrado aguardando validação final; Pendente: sem conferência registrada; Divergente: registros diferentes que precisam de revisão. Sem distribuição numérica disponível, o círculo fica neutro; as cores são apenas a legenda e não representam fatias nem percentuais.
- O gráfico de distribuição permanece na T03 e não redireciona para Fluxo de Caixa; as definições visíveis respondem à pergunta sobre o significado do gráfico.
- Os cards Ambíguos e Chargebacks permanecem visíveis; seus detalhes explicam que o relatório atual não traz campos suficientes para classificar esses casos com segurança. Não são fabricados registros.

## Validação

- A expectativa do teste de integração existente foi corrigida para conferir que somente lançamentos `R` entram nos totais. Os testes não foram executados nesta revisão.
- `npm run build` do frontend passou após esta correção.
- A revisão visual final na T03 permanece pendente de aprovação.

Os resultados da T02 e a sincronização Bling permanecem conforme o estado documentado no README; esta evidência não declara a carga de Contas a Pagar atualizada nem a T02 homologada.
