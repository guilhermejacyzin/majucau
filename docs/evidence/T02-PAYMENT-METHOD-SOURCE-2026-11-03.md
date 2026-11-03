# T02 — fonte de configuração das formas de pagamento — 03/11/2026

## O que foi feito

Foi acrescentada uma leitura somente de consulta da lista de formas de pagamento
do Bling e dos detalhes de cada forma. A leitura preserva os registros como
evidência RAW (`RAW_ONLY`), incluindo os campos que a configuração pode trazer
para destino, taxa, prazo e uso de dias úteis. Ela não calcula valores nem os
trata como entradas ou saídas financeiras.

A sincronização diária executa essa fonte como etapa independente das contas a
pagar, contas a receber e movimentos de caixa. Uma falha nela não interrompe as
demais leituras; a carga diária fica incompleta e pode ser retomada. As páginas
continuam limitadas, paginadas e publicadas com a política transacional já usada
pelas fontes Bling.

Na revisão seguinte, conferi o OpenAPI público atual e completei a captura dos
vínculos que faltavam em cada título. Receber pode trazer `formaPagamento.id`,
`portador.id` e `contaContabil.id`; pagar traz `formaPagamento.id` e, no detalhe,
`portador.id`. Os IDs agora são normalizados separadamente e guardados como
referências numéricas opcionais nas tabelas de receber e pagar. Um identificador
ausente continua ausente; não é substituído por conta ou forma presumida.
Referência: <https://developer.bling.com.br/referencia>.

## Por que precisamos desta fonte

A T02 deve exibir o resultado do relatório **Fluxo de Caixa** do Bling. A
referência pública da API consultada não apresenta uma rota documentada para
esse relatório pronto. A API documenta as formas de pagamento, mas as listas
trazem somente a identidade; os detalhes individuais carregam configurações
que podem alterar quando e em qual conta a projeção aparece. Sem esses detalhes,
somar vencimentos ou movimentos não reproduz necessariamente o relatório.

Por isso o projeto precisa ler essas configurações para reproduzir e comparar o
relatório, sem assumir que só a forma de pagamento define o resultado. Esta
mudança é preparação: os cards, gráfico, saldo bancário e semana crítica da T02
continuam sem novos resultados até a composição completa ser conferida contra o
relatório real. **Caixas e Bancos segue como fonte apenas dos saldos por
instituição.**

## Proteções mantidas

- Não alteradas as telas aprovadas nem regras financeiras.
- Nenhuma taxa, prazo ou movimento foi convertido em valor projetado.
- Falta de autorização em formas de pagamento não cancela as outras fontes.
- O registro da configuração não expõe segredo, token ou resposta em logs.
- Os resultados não foram homologados; é necessário comparar período,
  instituições e opções do relatório do Bling com a composição Majucau.

## Verificações

- `go test ./internal/integrations/bling ./database/gen`: passou localmente,
  incluindo os novos testes de preservação dos IDs de forma de pagamento e
  conta financeira.
- `go vet ./internal/integrations/bling`: passou.
- `scripts/verify-go-boundaries.ps1`: passou.
- `scripts/verify-migrations.ps1`: passou depois do registro do checksum da nova
  migration.
- O teste de PostgreSQL depende da CI ou de um banco de teste configurado.
- Ainda falta executar a atualização real após o código novo e conferir a T02.

Na CI do commit `6a2ea755a1d946d5e0e0937b3fa4e6482bc04689`, a verificação de
Windows falhou apenas em `Smoke Windows worker service lifecycle`, com
`WORKER_SERVICE_WAIT_FOR_RUNNING`, embora o diagnóstico seguinte tenha
encontrado o serviço como `Running`. Os jobs PostgreSQL, backlog e verificações
anteriores passaram. Esta falha não tem relação com a captura dos IDs; deve ser
reexecutada/investigada antes de declarar a CI verde.

## Próximo passo

Completar a composição da projeção diária usando as fontes aprovadas e as
configurações efetivamente retornadas; obter o saldo bancário de Caixas e
Bancos; depois comparar totais e datas com o relatório de Fluxo de Caixa do
Bling antes de alimentar T02 e T01. Se a composição não reproduzir o relatório,
manter o resultado pendente e corrigir o mapeamento antes de publicar valores.

## Decisão e checagem da API direta — 03/11/2026

A responsável confirmou que a integração deve usar somente a API oficial direta,
sem carregar CSV/XLS do relatório pelo Drive. Em seguida, aprovou reproduzir o
mesmo relatório usando os recursos oficiais. A API pública não oferece uma rota
`fluxo`/`relatório` pronta; portanto, o Majucau deverá compor o resultado a
partir dos dados oficiais e provar que ele coincide com o relatório do Bling
antes de publicá-lo como confirmado.

Na consulta atual da especificação OpenAPI oficial, `GET /contas-contabeis`
retorna identidade, descrição, tipo e alias da conta financeira; não documenta
saldo. `GET /caixas` retorna lançamentos individuais, também sem um campo de
saldo atual por conta. A T02 deve manter o saldo bancário separado do saldo
projetado, conforme a regra aprovada. A reconstrução de um saldo de abertura a
partir do histórico e a cobertura histórica completa ainda precisam ser
provadas; não presumir que somar somente os movimentos desde 01/01/2026 forneça
o saldo atual.

Configuração de referência registrada a partir da captura da responsável:
período semanal selecionado no relatório, contas Itaú e Mercado Pago, float
bancário ativado, prazo da forma de recebimento ativado e pagamentos/recebimentos
incluídos. O comportamento equivalente dessas opções precisa ser validado com
as configurações e lançamentos efetivamente retornados pela API. A captura é
referência de conferência, não fonte de dados em produção.

A leitura do relatório pelo Chrome não foi necessária nem tentada nesta decisão:
foi escolhida a API oficial direta. T02 permanece incompleta até a composição,
a conferência dos totais/datas contra o mesmo relatório e a ligação dos
resultados à tela, sem substituir o saldo de Caixas e Bancos por projeção.

## Referências oficiais que orientam a composição

- O [artigo do Fluxo de Caixa do Bling](https://ajuda.bling.com.br/hc/pt-br/articles/360038306714-Como-gerar-um-relat%C3%B3rio-de-futuras-entradas-e-sa%C3%ADdas-de-valores)
  descreve o relatório como a previsão futura de contas a pagar e receber.
- O [artigo de taxas e prazos das formas de pagamento](https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling)
  confirma que a previsão pode considerar o valor líquido, a taxa, o prazo de
  recebimento, os dias úteis e o destino da forma de pagamento.
- O [cadastro de contas financeiras](https://ajuda.bling.com.br/hc/pt-br/articles/34491885650839-Como-cadastrar-contas-financeiras-no-Bling)
  informa que a conta possui uma data inicial de lançamentos e um saldo inicial.
  Esse saldo inicial não aparece nos campos documentados de `GET /contas-contabeis`;
  será necessário provar se a leitura completa de `/caixas` o expõe como
  lançamento recuperável antes de calcular saldo de abertura.

Essas referências confirmam quais dados precisam compor o resultado, mas não
definem uma fórmula alternativa à do relatório. A fonte oficial do relatório
continua sendo o padrão de comparação e a T02 só será confirmada após
confrontar a API com ele, usando as opções já escolhidas pela responsável.

## Validação automática da saída do relatório — 03/11/2026

Foi implementado `ValidateBlingCashFlowReport` e `CompareBlingCashFlowReport`
em `internal/financial/bling_flow_report.go`. O primeiro confere a conta de
cada período, a continuidade do saldo entre períodos e a linha de totais; o
segundo compara período, saldo inicial, entradas, saídas, ajustes e saldo final
da série composta via API com a referência do Bling. A saída é rejeitada se
qualquer valor divergir. As saídas mantêm sinal negativo, como no relatório.

O teste dourado usa somente os valores da captura já enviada pela responsável:
seis semanas, R$ 76.600,30 em entradas, -R$ 281.052,64 em saídas e
-R$ 129.988,98 de saldo ao fim. Outros testes provam que divergência e quebra
de continuidade são recusadas. A captura é um oráculo de teste, não dado em
produção.

`go test ./internal/financial` passou. Este componente permite verificar a
equivalência da futura projeção API, mas ainda não existe uma série composta
para comparar e ele não preenche cards/gráfico da T02. Permanecem pendentes a
composição real, saldo inicial separado via Caixas e Bancos, semana crítica,
ligação do snapshot à tela e conferência com a atualização real do Bling.

## Agrupamento do resultado diário para o relatório semanal — 03/11/2026

Foi acrescentada a função `BlingCashFlowReportFromDailyBalances`, que agrupa a
projeção diária em períodos de domingo a sábado, conforme a captura de
referência. As pontas do intervalo podem ser semanas parciais. A função conserva
entrada positiva, converte saída para sinal negativo e confere a conta diária,
a continuidade entre dias, os limites do período e os totais antes de devolver
os períodos.

Os testes cobrem o primeiro período parcial (segunda a sábado), a semana
seguinte, os sinais, o saldo de fechamento e a recusa de saldo inconsistente.
`go test ./internal/financial`, `go vet ./internal/financial` e `git diff
--check` passaram. Ainda é somente uma etapa de composição: a projeção não está
alimentada pelas consultas do banco/API, não tem o saldo inicial oficial e não
está no snapshot nem na T02. Não deve ser apresentada como resultado do Bling
até a comparação real passar.

## Cálculo Bling implementado no domínio — 03/11/2026

Foi acrescentado um primeiro componente reutilizável para reproduzir as
configurações de recebimento do relatório:

- detalhes RAW da forma de pagamento agora podem ser normalizados em
  taxa percentual, taxa fixa, prazo e indicação de dias úteis;
- a taxa é calculada por parcela, somando percentual e valor fixo, arredondando
  para centavos e subtraindo do valor bruto;
- prazo em dias corridos é somado ao vencimento;
- prazo em dias úteis exige um calendário explícito, que também avança a data
  final para o próximo dia útil.
- valor ausente no detalhe RAW continua marcado como desconhecido; o cálculo
  rejeita taxa, prazo ou indicação de dias úteis ausentes, em vez de assumir
  zero ou dia corrido.

O calendário é obrigatório quando a forma usa dias úteis porque a ajuda do Bling
manda avançar a data quando ela cai em fim de semana ou feriado. Não se usa um
calendário de segunda a sexta como substituto silencioso. O exemplo oficial
(`R$ 166,66`, alíquota `0,25%`, taxa fixa `R$ 1,99`) foi codificado como caso de
referência: taxa `R$ 2,41`, líquido `R$ 164,25`. O exemplo de prazo de 30 dias
corridos também foi coberto.

**Limite atual:** este componente ainda não foi conectado aos títulos recebidos
nem às consultas da T02; portanto ainda não gera valores na tela e não prova a
equivalência com o relatório real. Faltam integrar o leitor dos detalhes RAW,
confirmar o float por conta, definir um calendário oficial compatível com o
Bling, montar a série com os títulos e movimentos elegíveis, disponibilizar a
série no snapshot, ligar os cards/gráfico existentes sem mudar o desenho e
comparar os períodos com a referência real.

Referência oficial para a fórmula e as datas:
<https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling>.

## Destino da forma de pagamento — 03/11/2026

A especificação oficial atual de `GET /formas-pagamentos/{idFormaPagamento}`
documenta `destino` como: `1` Conta a receber/pagar, `2` Ficha financeira e `3`
Caixa e bancos. A ajuda do Bling explica que Conta a receber/pagar alimenta a
previsão futura do relatório, enquanto lançamento direto em Caixa não gera a
previsão de contas. Portanto, só o destino `1` é elegível para compor a série
futura do relatório; `2` e `3` não entram nessa parte da projeção. Destino ausente
fica desconhecido e não é presumido.

O parser agora conserva o destino tipado e oferece um predicado que só autoriza
previsão quando o campo é conhecido e igual a `1`. Foram cobertos os três valores
documentados, a ausência do campo e a recusa de enumeração inválida. Os testes
focados passaram com os caches Go centralizados em `artifacts/cache`; `go vet`
e `git diff --check` também passaram. A CI do commit anterior concluiu verde.

**Limite atual:** este campo ainda não é consumido por um leitor de títulos nem
pela T02. O cálculo de forma de pagamento não está ligado à série do banco, ao
saldo inicial, ao float ou ao gráfico; ainda não há comparação real da série
inteira com o relatório. Esta etapa aumenta a fidelidade da futura composição,
mas não significa que a T02 já exiba resultado.

Referências oficiais consultadas em 03/11/2026:

- [Referência da API do Bling](https://developer.bling.com.br/referencia) —
  schema `FormasPagamentosDadosDTO.destino`.
- [Como cadastrar uma forma de pagamento](https://ajuda.bling.com.br/hc/pt-br/articles/360035630474-Como-cadastrar-uma-forma-de-pagamento-no-Bling) — enumera os destinos.
- [Como controlar taxas e prazos das formas de pagamento](https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling) — distingue Conta a Receber/Pagar de lançamento direto em Caixa e explica a previsão no fluxo.
