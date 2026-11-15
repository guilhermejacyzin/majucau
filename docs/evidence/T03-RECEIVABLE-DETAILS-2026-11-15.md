# T03 — detalhes de Contas a Receber

**Data:** 15/11/2026  
**Estado:** correções implementadas localmente; aguardando aprovação visual da responsável.

## Pedido registrado

- Tirar a tabela de títulos da página principal e mostrá-la ao clicar em “Ver detalhes” do card correspondente.
- Fazer os demais cards manterem sua própria ação de detalhes.
- Explicar o gráfico por situação dos recebíveis e o significado de cada status.
- Não apagar o gráfico com base na pergunta sobre o seu significado.
- Não remover cards da T03 sem uma instrução inequívoca. Os cards PIX, Boletos, Cartão, Ambíguos e Chargebacks permanecem.
- Substituir o card “Conciliação” por “TOTAL RECEBIDO”; seu valor é a soma de entradas conciliadas do relatório Caixa e Bancos no mês da data-base.

## Alterações

- A tabela principal de recebíveis foi retirada do fluxo visual da T03. “TOTAL A RECEBER” abre uma janela com Cliente, Origem, Vencimento, Valor bruto, Valor líquido e Status. O detalhe de “EM ATRASO” filtra pela data-base do snapshot.
- “TOTAL RECEBIDO” abre Data, Conta, Origem e Valor. O total e o detalhe usam os movimentos de entrada cujos IDs constam no snapshot oficial `situacaoConciliacao=1`, vinculado à mesma carga diária de `Caixas e Bancos`, e aplicam o mesmo intervalo mensal. Não há remoção de outro card de recebimentos além da substituição solicitada do card “Conciliação”.
- Auditoria da regra corrigiu a confusão entre os dois campos da API: `situacao=R` significa que o lançamento é um registro ativo; a conciliação é indicada separadamente por `situacaoConciliacao=1` e pela lista de IDs retornada por esse filtro. A consulta agora exige essa associação oficial para os totais e detalhes. A situação opcional não é usada como prova de conciliação nem exclui uma linha que esteja na lista oficial.
- A atualização do protótipo revelou e corrigiu outra falha: o detalhe de recebíveis parava porque a consulta tinha dois campos chamados `payment_method`. O nome associado à forma de pagamento agora tem alias próprio, eliminando a ambiguidade no banco real.
- As listas do dashboard aceitam até 200 linhas. O limite e os limites por campo/registro continuam protegendo a leitura e a memória; se houver mais de 200 itens, o detalhe não substitui uma exportação completa.
- O nome da forma de pagamento do Bling é passado ao frontend para identificar títulos de PIX, boleto e cartão sem expor o payload RAW.
- “Ver detalhes” dos outros cards permanece acionável. Quando o snapshot não possui uma lista de origem para aquele indicador, a janela informa essa ausência e não fabrica linhas.
- O gráfico continua visível e as definições ficam visíveis sem clique. Confirmado: título/valor conferidos com a origem; Provisório: valor registrado aguardando validação final; Pendente: sem conferência registrada; Divergente: registros diferentes que precisam de revisão. Sem distribuição numérica disponível, o círculo fica neutro; as cores são apenas a legenda e não representam fatias nem percentuais.
- O gráfico de distribuição permanece na T03 e não redireciona para Fluxo de Caixa; as definições visíveis respondem à pergunta sobre o significado do gráfico.
- Os cards Ambíguos e Chargebacks permanecem visíveis; seus detalhes explicam que o relatório atual não traz campos suficientes para classificar esses casos com segurança. Não são fabricados registros.

## Validação

- O teste de integração existente foi atualizado para exigir o snapshot oficial de conciliação e conferir que apenas IDs presentes nele entram nos totais, mesmo com `situacao` opcional ausente.
- `go test ./internal/dashboard -count=1` passou; os testes focados de sincronização/normalização do Bling passaram.
- `npm test -- --run src/T03ReceivablesDetails.test.tsx` passou (6 testes).
- `npm run build` passou (TypeScript e Vite).
- A revisão visual final na T03 permanece pendente de aprovação.

## Estado da sincronização e limite observado

- O protótipo foi atualizado e aberto na T03; a lista principal de títulos fica recolhida e cada botão “Ver detalhes” abre seu próprio painel. Permanecem os cards PIX, Boletos, Cartão, Ambíguos e Chargebacks.
- O botão de atualização foi acionado durante a conferência. A tentativa atual do Bling falhou para recebíveis com `BLING_API_SCOPE_INSUFFICIENT` (0 recebíveis lidos) e terminou parcial; por isso, a tela conserva o último snapshot publicado, datado de 15/11/2026 02:13. Não considerar esse snapshot como uma nova importação bem-sucedida.
- Após a correção da consulta SQL ambígua (`payment_method`), o snapshot do dashboard voltou a carregar no protótipo. O card “TOTAL RECEBIDO” e sua lista usam exclusivamente IDs do snapshot oficial de conciliação; o último snapshot disponível mostra R$ 0,00 e nenhuma linha para esse indicador. A causa exata é a ausência de entradas conciliadas nessa carga publicada, não um valor inventado ou o valor bruto de contas a receber.
- Para que a atualização diária produza novos recebíveis, a conexão precisa ter o escopo de leitura de contas a receber. A conexão atual não o concedeu; será necessário reautorizar a integração com esse escopo antes de declarar os novos valores como atualizados.

## Pendência de aprovação

A T03 corrigida está aberta no protótipo para conferência visual e aprovação da usuária. O acesso de leitura da API para recebíveis permanece como dependência externa e foi registrado separadamente da correção local.

Os resultados da T02 e a sincronização Bling permanecem conforme o estado documentado no README; esta evidência não declara a carga de Contas a Pagar atualizada nem a T02 homologada.
