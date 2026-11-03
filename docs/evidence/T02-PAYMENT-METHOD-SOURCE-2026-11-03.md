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
sem carregar um CSV/XLS do relatório pelo Drive. A especificação OpenAPI vigente
consultada lista `GET /caixas`, `GET /contas/receber`, `GET /contas/pagar` e
`GET /formas-pagamentos`, mas não lista rota `fluxo` ou `relatório`. Portanto,
ela não oferece, hoje, uma resposta pronta do relatório que a T02 possa apenas
copiar. A aba do relatório já aberta no Chrome também não ficou acessível nesta
sessão: depois da autorização da responsável, o controle de navegador retornou
`Browser is not available: chrome`; não houve leitura do relatório nem tentativa
de contornar o bloqueio.

Ainda falta uma decisão funcional para seguir usando apenas a API: autorizar que
o Majucau reproduza o mesmo resultado a partir dos recursos oficiais e o
compare com o relatório real, ou manter a T02 pendente até o Bling disponibilizar
uma rota oficial para o relatório pronto. Não presumir que a autorização para
“mostrar o mesmo resultado” autoriza cálculos divergentes ou uma fonte por
arquivo. T02 continua em 5% e não foi homologada.
