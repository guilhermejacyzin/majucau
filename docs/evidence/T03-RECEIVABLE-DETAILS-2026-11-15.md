# T03 — detalhes de Contas a Receber

**Data:** 15/11/2026  
**Estado:** implementado localmente; aguardando aprovação visual da responsável.

## Pedido registrado

- Tirar a tabela de títulos da página principal e mostrá-la ao clicar em “Ver detalhes” do card correspondente.
- Fazer os demais cards manterem sua própria ação de detalhes.
- Explicar o gráfico por situação dos recebíveis e o significado de cada status.
- Não apagar o gráfico com base na pergunta sobre o seu significado.
- Não remover cards da T03 sem uma instrução inequívoca. Os cards PIX, Boletos, Cartão, Ambíguos e Chargebacks permanecem.
- Substituir Conciliação por um único card “TOTAL RECEBIDO”; seu valor é a soma de entradas conciliadas do relatório Caixa e Bancos no mês da data-base, sem duplicar o antigo card mensal.

## Alterações

- A tabela principal de recebíveis foi retirada do fluxo visual da T03. “TOTAL A RECEBER” abre uma janela com Cliente, Origem, Vencimento, Valor bruto, Valor líquido e Status. O detalhe de “EM ATRASO” filtra pela data-base do snapshot.
- “TOTAL RECEBIDO” abre Data, Conta, Origem e Valor. As linhas vêm de `bling_cash_movements` da mesma carga diária publicada, com `direction = INFLOW`, status `R` ou vazio e as mesmas datas mensais do agregado `receipts_month`. O antigo card duplicado “RECEBIDO NO MÊS” foi removido.
- As listas do dashboard aceitam até 200 linhas. O limite e os limites por campo/registro continuam protegendo a leitura e a memória; se houver mais de 200 itens, o detalhe não substitui uma exportação completa.
- O nome da forma de pagamento do Bling é passado ao frontend para identificar títulos de PIX, boleto e cartão sem expor o payload RAW.
- “Ver detalhes” dos outros cards permanece acionável. Quando o snapshot não possui uma lista de origem para aquele indicador, a janela informa essa ausência e não fabrica linhas.
- O gráfico continua visível. O texto explica que ele divide o total por situação; “O que significa cada situação?” abre definições de Confirmado, Provisório, Pendente e Divergente. Sem distribuição numérica disponível, o círculo fica neutro e não apresenta percentuais estimados.
- A ação do gráfico não redireciona para Fluxo de Caixa. Ela abre a explicação das situações, respondendo diretamente à pergunta da responsável sobre o significado do gráfico.
- Os cards Ambíguos e Chargebacks permanecem visíveis; a instrução mais recente questionou sua remoção, então não foram apagados.

## Validação

- `go test -count=1 ./internal/dashboard`
- Testes focados da interface T03, T01 e navegação T05.
- `npm run build` do frontend.
- Revisão visual final na T03 aberta antes da homologação.

Os resultados da T02 e a sincronização Bling permanecem conforme o estado documentado no README; esta evidência não declara a carga de Contas a Pagar atualizada nem a T02 homologada.
