# T02 — relatório semanal do Bling, variação e saldo Itaú

**Data:** 07/11/2026
**Escopo:** análise e preparação do gráfico de fluxo de caixa; preservação da
tela aprovada e das regras financeiras.

## O que foi confirmado

- A imagem do relatório fornecida pela responsável mostra os períodos em
  semanas, saldo inicial, receitas, despesas, ajuste de saldo e saldo acumulado.
- A imagem do gráfico mostra uma marca em cada semana; ao passar o mouse, o
  Bling apresenta o período e os valores de receitas, despesas e saldo.
- A captura mais recente do relatório marca somente uma conta financeira. No
  contexto da regra aprovada, essa conta é Itaú. O código exigia Itaú e Mercado
  Pago juntos, então podia retornar dados indisponíveis para essa configuração;
  o seletor da composição foi corrigido para uma conta Itaú.
- A tela Majucau deve acompanhar esse padrão e acrescentar a variação percentual
  semanal para entradas e saídas. Entradas e saídas são comparadas com a semana
  anterior; para saídas, a comparação usa o valor absoluto da despesa, para
  que a variação expresse o aumento ou a redução do gasto.
- Gisele determinou que o saldo bancário considerado na T02 seja somente o do
  Itaú. Não incluir Mercado Pago e não presumir transferência de recebimentos
  da Nuvem ao Itaú. Esse saldo continua separado do cálculo das entradas e
  saídas exibidas pelo relatório.
- Gisele esclareceu que o saldo desenhado no gráfico é outro indicador: o
  saldo acumulado do relatório. Pela tabela enviada, ele segue `saldo inicial +
  receitas + despesas (negativas) + ajuste de saldo`; o saldo inicial da semana
  seguinte é o saldo acumulado da semana anterior. O saldo bancário atual do
  card continua sendo somente Itaú. A trajetória do gráfico acompanha a conta
  selecionada no relatório; se o filtro for somente Itaú, o saldo acumulado
  projetado será dessa conta, sem transformar esse número no saldo bancário
  atual do card.
- A referência oficial da API publica recursos de caixas e contas financeiras,
  mas não uma operação que devolva o relatório Fluxo de Caixa pronto. A
  integração precisa compor as informações pelos recursos oficiais e comparar
  período, conta e opções com o relatório apresentado no Bling.

## Gráfico semanal entregue no protótipo

- O gráfico da T02 foi aberto no protótipo em 07/11/2026. A tela mostrou dados
  reais parciais da integração Bling em 10 faixas semanais, com pontos e linhas
  de entradas e saídas.
- O tooltip acessível de cada semana mostra período, valores de entradas e
  saídas, e a variação percentual em relação à semana anterior. A variação de
  saídas considera o valor absoluto para indicar aumento ou redução dos gastos.
- Os cards de entradas e saídas de 30 dias também mostraram valores parciais
  recebidos pelo fluxo real de dados. A tela e sua disposição aprovada foram
  preservadas.
- O contrato de dados aceita opcionalmente o saldo acumulado de fechamento de
  cada semana. Quando esse valor for fornecido por uma origem validada, o
  gráfico consegue mostrar a linha e a variação em reais. A legenda e o
  tooltip identificam esse valor como saldo acumulado do relatório, sem
  confundi-lo com o card separado de saldo bancário Itaú.
- A conversão dos valores e percentuais usa aritmética decimal exata, sem
  conversão financeira por ponto flutuante.
- Os textos de governança separam o card de saldo bancário atual do saldo
  acumulado do relatório. O gráfico usa entradas/saídas da conta selecionada no
  relatório; na captura mais recente há uma conta, identificada como Itaú no
  contexto aprovado.
- A leitura da composição do relatório agora exige somente a conta Itaú
  identificada na captura mais recente; não exige Mercado Pago da captura
  anterior.

## Limite atual — saldo acumulado ainda não concluído

O backend ainda não preenche o campo `closing_balance`. A legenda continua
identificando o saldo acumulado, mas a série azul e seus valores permanecem
indisponíveis no gráfico. As capturas confirmam a fórmula e a continuidade
semanal, mas o sistema ainda não tem uma origem validada para o saldo inicial e
os ajustes do relatório. Não foi inventado um saldo pela soma de movimentos.
Portanto, está concluída a primeira parte verificável — fazer o gráfico semanal
receber e mostrar entradas, saídas e suas variações reais —, mas o gráfico T02
como um todo e o BK-062 continuam em andamento até a série de saldo ser
validada e a comparação integral com o relatório ser feita.

A equivalência completa com o relatório também não está homologada: as imagens
permitem confirmar a apresentação e as opções visíveis, mas não substituem a
comparação de todas as semanas com dados sincronizados na mesma configuração.
O navegador do Codex recusou acesso à página do Bling por uma preferência salva
da sessão. Essa restrição não foi contornada.

## Verificações

- `go test ./internal/dashboard -count=1` — passou.
- `npm test -- --run src/bootstrap.test.ts src/T01ExecutiveNavigation.test.tsx`
  (na pasta `frontend`) — passou, 28 testes.
- `git diff --check` — passou.
- Verificação visual e de acessibilidade no protótipo: T02 selecionada, gráfico
  com 10 pontos/faixas semanais; tooltips continham valores e percentuais; fonte
  indicada como “Parcial · fonte: Bling”. A série do saldo acumulado e o card de
  saldo bancário continuam como “—”, sem mostrar valores não confirmados.

## Próximos passos

1. Comparar cada faixa do gráfico com o relatório do Bling usando o mesmo
   período, conta e opções.
2. Com a integração autorizada, obter pela API oficial as entradas e saídas da
   mesma janela e opções do relatório e comparar cada semana.
3. Identificar a origem oficial do saldo inicial e dos ajustes usados no
   relatório do Bling e reconciliar os fechamentos semanais antes de preencher
   `closing_balance`. O card de saldo bancário atual deve continuar limitado ao
   Itaú.
4. Atualizar a T01 a partir do resultado validado da T02, sem criar cálculo
   paralelo.

**Andamento:** a primeira etapa do gráfico (série semanal de entradas/saídas e
variação percentual visíveis no protótipo) está concluída. BK-062 permanece em
30% (70% restante) e o projeto em 12,3% (87,7% restante), pois ainda faltam a
série validada do saldo acumulado, a comparação integral e a homologação.

**Referências:** [Referência oficial da API Bling](https://developer.bling.com.br/referencia); capturas do relatório e do gráfico enviadas pela responsável nesta conversa.
