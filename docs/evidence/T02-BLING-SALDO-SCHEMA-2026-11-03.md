# T02 — verificação do contrato oficial de saldo do Bling

**Data da conferência:** 03/11/2026  
**Fonte consultada:** especificação OpenAPI oficial da API Bling e Central de Ajuda oficial, conferida novamente em 03/11/2026.
**Escopo:** somente leitura; sem mudança de tela, regra financeira ou configuração da conta.

## Estado local adicional — 03/11/2026

Consulta sanitizada ao último lote normalizado de `caixas` no PostgreSQL isolado
do Majucau:

- estado do lote: **Parcial**; 31.032 movimentos lidos entre 01/01/2026 e
  30/10/2026;
- 9 contas distintas estão identificadas nos movimentos; nenhum movimento está
  sem ID de conta;
- `origin_type` está ausente nos 31.032 movimentos;
- 29.333 registros RAW contêm a propriedade `origem`; nas respostas guardadas,
  esse objeto contém a chave `id`, mas não contém a chave `tipo`.

Foram consultadas somente contagens, datas mínima/máxima e nomes de chaves JSON.
Nenhum nome de cliente, valor ou ID financeiro foi exibido ou gravado. A carga
confirma que o normalizador espera `origem.tipo`, mas essa propriedade não está
presente no RAW recebido. Não se deve preencher o tipo usando inferências.

A responsável autorizou a leitura das Contas Financeiras e a consulta do
histórico completo para validar saldos. As permissões ainda não foram aplicadas
no cadastro do app nem houve nova autorização OAuth. Próximo fluxo: aplicar
somente leitura, reautorizar, sincronizar catálogo e histórico completo,
identificar as duas contas aprovadas e reconciliar abertura/ajustes antes de
calcular qualquer saldo. A permissão de catálogo, sozinha, não fornece saldo
monetário.

## Revalidação do OpenAPI publicado — 03/11/2026

Conferi novamente o arquivo OpenAPI servido pela própria página de referência:
[openapi-CR5QY_N1.json](https://developer.bling.com.br/build/assets/openapi-CR5QY_N1.json).
O resultado confirma o limite descrito abaixo:

- `GET /contas-contabeis` e `GET /contas-contabeis/{idContaContabil}` devolvem
  `id`, `descricao`, `tipo` e `aliasIntegracao`; não há campo monetário de
  saldo na resposta;
- `GET /caixas` devolve lançamentos com valor e conta financeira, mas a linha
  não contém saldo da conta;
- `GET /caixas/{idCaixa}` contém o campo `saldo` como texto `S`/`N`, descrito
  como indicação de ajuste após o lançamento, e não como quantia;
- a lista atual de operações não contém rota de leitura de saldo bancário.

Portanto, a permissão somente de leitura de Contas Financeiras é útil para
identificar Itaú e Mercado Pago pelo cadastro, mas não basta para mostrar o
saldo em reais. A autorização da responsável está registrada; a permissão
ainda não foi aplicada porque o navegador do Codex bloqueou o acesso ao site.
Mesmo depois de aplicada, não se deve preencher saldo atual somando apenas os
movimentos desde 01/01/2026: sem saldo de abertura comprovado nessa data, isso
seria somente a variação daquele período.

## O que a API pública retorna

O contrato atual da referência oficial contém:

- `GET /caixas`: lista de lançamentos com data, valor, situação, descrição, origem e conta financeira. O contrato da lista não inclui um campo de saldo. Se não houver datas na consulta, a API limita a resposta ao mês atual; o Majucau informa 01/01/2026 como data inicial para buscar o histórico aprovado de CAPEX;
- `GET /caixas/{idCaixa}`: detalhe de um lançamento. Esse detalhe tem `saldo` (`S`/`N`), descrito como indicador de ajuste do saldo após o lançamento; não é um valor de saldo atual nem de saldo inicial;
- `GET /contas-contabeis` e `GET /contas-contabeis/{idContaContabil}`: identidade e configuração da conta (`id`, `descricao`, `tipo` e `aliasIntegracao`).

O DTO das contas financeiras não tem saldo inicial nem saldo atual. Portanto, o escopo Contas Financeiras desbloqueia a identidade/configuração das contas, mas a referência pública não define um campo de valor de saldo para esse recurso. A página oficial lista esses contratos em [Referência da API Bling](https://developer.bling.com.br/referencia).

### Origem dos movimentos

O contrato atual de `GET /caixas` também inclui `origem.id` e `origem.tipo`.
Os tipos descritos são `caixa`, `duplicata` (contas a receber/pagar), `bordero`
e `estoque`. A listagem inclui ainda valor, direção, data, situação e conta
financeira. No detalhe `GET /caixas/{idCaixa}`, a propriedade adicional
`saldo` é somente a indicação `S`/`N` de ajuste após o lançamento; não informa
o montante do saldo.

Na tabela normalizada, 31.032 movimentos entre 01/01/2026 e 30/10/2026 estão
sem `origin_type`. O normalizador atual lê `origem.tipo`, então não se sabe
ainda se o Bling omitiu a origem nesses lançamentos ou se a carga RAW mostra
uma variação que precisa de tratamento. Até conferir uma amostra sanitizada da
resposta RAW, esses registros não devem ser classificados por aproximação.
Essa lacuna pode afetar a comparação das movimentações com o relatório pronto.

## O que a Central de Ajuda confirma

Ao cadastrar uma conta financeira, o Bling pede um saldo inicial e uma data de início. Ajustar o saldo cria um registro de histórico dentro do Bling; isso não altera o saldo real do banco. O próprio Bling descreve suas contas como controle interno e informa que elas não sincronizam automaticamente os movimentos bancários. Fontes oficiais: [Como cadastrar contas financeiras no Bling](https://ajuda.bling.com.br/hc/pt-br/articles/34491885650839-Como-cadastrar-contas-financeiras-no-Bling) e [Como ajustar o registro de saldo da conta](https://ajuda.bling.com.br/hc/pt-br/articles/29903710138263-Como-ajustar-o-registro-de-saldo-da-conta-no-Bling).

O Bling também explica que o relatório Fluxo de Caixa projeta entradas e saídas futuras a partir de contas a pagar e a receber e considera as condições da forma de pagamento para calcular o valor líquido esperado: [Relatórios do Bling](https://ajuda.bling.com.br/hc/pt-br/articles/10448301650455-Relat%C3%B3rios-do-Bling) e [controle de taxas e previsão de recebíveis](https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling).

## Efeito na T02

O fluxo da T02 consegue usar os lançamentos e títulos da API para compor entradas e saídas futuras. A API pública consultada, porém, não entrega o saldo atual pronto nem o saldo inicial usado pelo relatório. O histórico sincronizado pela T02 começa em 01/01/2026 para atender ao histórico aprovado de CAPEX; ele não comprova, sozinho, o saldo de abertura anterior a essa data.

Na implementação atual, `readStreamSummary` define `cash_balance` como `UNAVAILABLE` e não popula `bank_balances`. O contrato desses campos já existe, mas não há um cálculo de saldo por conta conectado à leitura de `GET /caixas`. Somar somente os movimentos desde 01/01/2026 produziria a variação líquida daquele período, não o saldo atual, a menos que o saldo de abertura dessa data também fosse comprovado.

Também consultei a tabela normalizada local sem ler nem exibir valores, nomes ou IDs: havia 31.032 movimentos entre 01/01/2026 e 30/10/2026; o campo `origin_type` estava vazio em todos. Isso não prova que não existam saldos iniciais ou ajustes no Bling, mas mostra que o histórico já sincronizado não permite identificá-los por esse campo. A próxima conferência precisa verificar o início efetivo de cada conta e se a API fornece os respectivos registros de abertura/ajuste; não se deve somar o histórico desde 01/01/2026 como se fosse o saldo atual.

Assim, continuam indisponíveis o saldo atual do Itaú, o saldo atual do Mercado Pago, o menor saldo e o saldo projetado. A permissão de leitura aprovada para Contas Financeiras permitirá identificar as contas no catálogo, mas não supre, por si só, o saldo inicial/atual ausente no contrato documentado. Não usei o saldo do relatório como saldo bancário nem estimei valores. A composição da T02 segue Parcial e BK-062 permanece em 30% até demonstrar posição por conta e equivalência com o relatório.

## Próxima verificação

Depois de a permissão ser aplicada e o Majucau ser autorizado novamente:

1. sincronizar o catálogo para identificar as contas Itaú e Mercado Pago;
2. verificar, por evidência da API e do relatório do Bling, a data de início e o saldo inicial de cada conta e se os registros de abertura/ajuste aparecem em `GET /caixas`;
3. somente se a abertura estiver comprovada, calcular e reconciliar o saldo por conta; caso contrário, manter os saldos indisponíveis até existir uma fonte aprovada que os forneça.
