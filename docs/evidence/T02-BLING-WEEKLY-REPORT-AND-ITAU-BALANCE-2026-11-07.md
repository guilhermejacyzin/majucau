# T02 — tabela semanal, comparação e saldo Itaú

**Data:** 07/11/2026
**Estado:** tabela publicada no protótipo local; conferência parcial concluída em 08/11/2026; ainda não homologada.

## Regra e referências da responsável

- A T02 deve apresentar o resultado do relatório **Fluxo de Caixa** do Bling.
- A referência enviada agrupa os dados em períodos semanais e mostra saldo
  inicial, receitas, despesas, ajustes e saldo acumulado.
- O card de saldo bancário da T02 é uma informação separada e usa apenas a
  conta Itaú. Não somar Mercado Pago nem presumir transferências da Nuvem.
- O período da tabela deve ser selecionável a partir do dia atual de Brasília;
  a tabela e a semana crítica acompanham as datas escolhidas.
- A T01 reaproveita o resultado da T02 no snapshot comum, em vez de fazer outra
  soma/consulta independente. Assim, o resumo executivo e a tela detalhada
  mostram a mesma fonte e os mesmos valores.

## Conferência numérica com a imagem do Bling

Relatório apresentado para 07/11/2026 a 30/11/2026:

| Período | Receitas Bling | Despesas Bling | Resultado API comparado |
|---|---:|---:|---|
| 07/11 | R$ 0,00 | -R$ 1.977,43 | O limite mínimo mudou para 08/11 quando virou o dia em Brasília; 07/11 não pode mais ser selecionado pela regra aprovada de mostrar só hoje em diante. A referência anterior mostrava -R$ 1.977,43. Não afirmo uma nova leitura deste dia passado. |
| 08–14/11 | R$ 12.362,61 | -R$ 9.416,24 | Igual na tela atual. |
| 15–21/11 | R$ 12.815,27 | -R$ 160.208,69 | Igual na tela atual. |
| 22–28/11 | R$ 34.205,22 | -R$ 27.012,27 | A tela mostra R$ 34.205,21 em receitas (diferença de R$ 0,01); despesas iguais. Não apliquei correção artificial. |
| 29–30/11 | R$ 1.068,87 | -R$ 9.180,00 | Igual na tela atual. |

No período selecionável atual, 08–30/11, a soma da tabela mostra receitas de
R$ 60.451,96 (R$ 0,01 abaixo do relatório) e despesas de -R$ 205.817,20 (igual
ao total correspondente do relatório, excluindo 07/11). Os quatro períodos e o
total ficaram visíveis sem rolagem horizontal na largura testada.

### Causa identificada e correção

A tabela de Caixa e Bancos enviada pela responsável mostra uma entrada de
R$ 151,23 em 05/11, o saldo de R$ 87.548,33 ao fim de 05/11, uma saída de
R$ 25.000,00 em 06/11 e o saldo de R$ 62.548,33 ao fim de 06/11. No relatório
Fluxo de Caixa iniciado em 07/11, R$ 62.548,33 já aparece como saldo inicial;
portanto, os movimentos anteriores não podem ser somados novamente às
despesas/receitas de 07/11.

A implementação anterior aplicava um deslocamento de um dia a todo tipo de
registro. A correção preserva a data de movimentos realizados, usa a janela do
dia anterior somente para previsões de recebimentos e inicia movimentos e
contas a pagar na data escolhida. Depois de reiniciar o protótipo pelo roteiro
normal, a semana visível começou em 08/11 e os valores das semanas selecionadas
bateram conforme a tabela acima, exceto a diferença de R$ 0,01 nas receitas de
22–28/11.

O deslocamento observado ainda é uma inferência de comparação, não uma leitura
da configuração salva do Itaú. A documentação oficial da API publica os
recursos de Caixas e Bancos, mas não publica uma operação pronta que devolva o
relatório semanal calculado: [Referência API do Bling](https://developer.bling.com.br/referencia).
A ajuda do Bling explica que o campo Float Bancário soma dias à data de baixa
de boletos retornados pelo banco; isso não confirma os dias configurados na
conta Itaú nem documenta o efeito geral na previsão do relatório:
[artigo oficial sobre retorno bancário](https://ajuda.bling.com.br/hc/pt-br/articles/360038873754-Como-inserir-o-retorno-do-banco-para-a-baixa-de-boletos).
Por segurança, a projeção continua **Parcial** até a comparação final confirmar
cada período com o mesmo filtro, conta e opções do relatório.

## Alteração visual da T02

- O conteúdo semanal principal tem tabela e gráfico, ambos no período escolhido;
  a semana crítica fica abaixo dos dois.
- Reduzi fontes, cartões, margens e alturas dos painéis somente na T02.
- A tabela preserva os valores em uma linha; quando o período tem muitas
  semanas, usa rolagem horizontal em vez de quebrar valores em pedaços. Valores
  negativos da tabela e da semana crítica ficam vermelhos.
- A tabela mostra valores semanais de receitas e despesas, variação percentual
  frente à semana anterior, total do período e as linhas de saldo inicial,
  ajustes e saldo acumulado.
- Saldo inicial, ajustes e saldo acumulado permanecem “—” enquanto a origem
  oficial não estiver confirmada; não são calculados por aproximação.
- Os dois cards de prazos (30/60 dias), o card “Menor saldo 30 / 60 dias” e a
  reserva de R$ 50.000 continuam separados da tabela semanal.

## Verificações técnicas

- `go test ./internal/dashboard ./internal/application -count=1` — aprovado.
- `go vet ./internal/dashboard ./internal/application` — aprovado.
- `npm run lint` — aprovado.
- `npm run typecheck` — aprovado.
- `npm test -- --run` — 73 testes aprovados.
- `npm run build` — aprovado.
- O protótipo estava parado e a URL local não respondia; isso deixou a aba antiga
  sem atualizar. Reiniciei pelo roteiro normal, abri o protótipo e conferi a T02
  com tabela, gráfico e semana crítica abaixo. A
  data mínima acompanha Brasília; ao mudar apenas o fim, a data inicial fica
  preservada. A comparação de 07/11 não pode ser repetida agora porque já é
  passado e o seletor não permite retornar.

**Andamento mantido:** BK-062 em 30% (70% restante); projeto em 12,3% (87,7%
restante). Não marcar como concluído enquanto persistirem diferença de R$ 0,01,
validação do saldo inicial/acumulado e conferência da versão corrigida.
