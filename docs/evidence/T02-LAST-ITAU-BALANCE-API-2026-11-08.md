# T02 — último saldo diário do Itaú

## Pedido registrado

O card **Saldo atual** deve receber o último fechamento que aparece em Caixa e
Bancos para a conta Itaú. O card **Menor saldo 30 / 60 dias** deve usar essa
posição como ponto de partida e a série diária aprovada da T02. O número visto
em uma captura serve apenas para conferência; não será gravado como valor fixo.

## O que a API oficial fornece

- `GET /caixas` devolve lançamentos paginados, com data, valor, débito/crédito,
  situação e conta financeira. Não devolve a linha de fechamento diário do
  relatório.
- Sem `dataInicial`, `GET /caixas` consulta por padrão desde o primeiro dia do
  mês atual. O sincronizador precisa enviar uma data inicial explícita.
- `GET /caixas/{idCaixa}` tem o campo `saldo`, mas a documentação o define como
  indicador `S`/`N` de ajuste após o lançamento; esse campo não contém o saldo
  bancário.
- O catálogo de contas financeiras identifica as contas, mas não traz o saldo.

Fonte: [referência oficial da API Bling](https://developer.bling.com.br/referencia),
OpenAPI v3, seção “Caixas e Bancos”.

## Implementação local

- O sincronizador envia `dataInicial=2026-01-01` e `dataFinal` do dia da
  execução. A data inicial fica explícita para evitar o padrão de mês atual.
- A leitura seleciona somente a conta identificada como Itaú e lançamentos
  registrados (`R`). O saldo de movimentos é calculado com créditos positivos
  e débitos negativos, em precisão decimal.
- Como a API não entrega o saldo de abertura nem o fechamento do relatório, o
  valor reconstruído fica como **Parcial** até ser comparado com a última linha
  “Saldo em [data]” do relatório. O print recebido mostrou R$ 62.548,33 em
  06/11/2026; esse é apenas um ponto de conferência, não uma entrada do sistema.
- Os menores saldos de 30 e 60 dias também permanecem parciais enquanto a base
  atual não for validada.

## Limite desta execução

O navegador recusou a leitura automatizada do relatório por uma permissão salva
do próprio navegador do Codex. O serviço local do Majucau também não está
disponível neste ambiente para executar uma sincronização real. Portanto, ainda
não há confirmação de que o valor reconstruído coincide com o último saldo do
relatório, nem conferência visual da tela atualizada.

## Próximo passo da mesma tarefa

Executar a sincronização real do histórico do Itaú, comparar o saldo calculado
com o fechamento mais recente do relatório e só então confirmar a exibição nos
dois cards. Se a comparação não bater, investigar abertura/ajustes ou histórico
anterior ao início configurado; manter os valores como parciais até resolver.
