# T02 — verificação do contrato oficial de saldo do Bling

**Data da conferência:** 03/11/2026  
**Fonte consultada:** especificação OpenAPI oficial da API Bling e Central de Ajuda oficial.  
**Escopo:** somente leitura; sem mudança de tela, regra financeira ou configuração da conta.

## O que a API pública retorna

O contrato atual da referência oficial contém:

- `GET /caixas`: lista de lançamentos com data, valor, situação, descrição, origem e conta financeira;
- `GET /caixas/{idCaixa}`: detalhe de um lançamento;
- `GET /contas-contabeis` e `GET /contas-contabeis/{idContaContabil}`: identidade e configuração da conta (`id`, `descricao`, `tipo` e `aliasIntegracao`).

No lançamento, o campo `saldo` é uma marcação `S`/`N` que indica se houve ajuste de saldo após aquele lançamento. **Não é o valor do saldo bancário.** O DTO das contas financeiras também não tem saldo inicial nem saldo atual. A página oficial lista esses contratos em [Referência da API Bling](https://developer.bling.com.br/referencia).

## O que a Central de Ajuda confirma

Ao cadastrar uma conta financeira, o Bling pede um saldo inicial e uma data de início. A ajuda também informa que não existe integração automática de movimentações bancárias para essas contas: o controle é interno ao ERP. Fonte: [Como cadastrar contas financeiras no Bling](https://ajuda.bling.com.br/hc/pt-br/articles/34491885650839-Como-cadastrar-contas-financeiras-no-Bling).

## Efeito na T02

O fluxo da T02 consegue usar os lançamentos e títulos da API para compor entradas e saídas futuras. A API pública consultada, porém, não entrega o saldo atual pronto nem o saldo inicial usado pelo relatório. O histórico sincronizado pela T02 começa em 01/01/2026 para atender ao histórico aprovado de CAPEX; ele não comprova, sozinho, o saldo de abertura anterior a essa data.

Na implementação atual, `readStreamSummary` define `cash_balance` como `UNAVAILABLE` e não popula `bank_balances`. O contrato desses campos já existe, mas não há um cálculo de saldo por conta conectado à leitura de `GET /caixas`. Somar somente os movimentos desde 01/01/2026 produziria a variação líquida daquele período, não o saldo atual, a menos que o saldo de abertura dessa data também fosse comprovado.

Assim, continuam indisponíveis o saldo atual do Itaú, o saldo atual do Mercado Pago, o menor saldo e o saldo projetado. A permissão de leitura aprovada para Contas Financeiras permitirá identificar as contas no catálogo, mas não supre, por si só, o saldo inicial/atual ausente no contrato documentado. Não usei o saldo do relatório como saldo bancário nem estimei valores. A composição da T02 segue Parcial e BK-062 permanece em 30% até demonstrar posição por conta e equivalência com o relatório.

## Próxima verificação

Depois de a permissão ser aplicada e o Majucau ser autorizado novamente:

1. sincronizar o catálogo para identificar as contas Itaú e Mercado Pago;
2. verificar, por evidência da API e do relatório do Bling, se há movimento de abertura utilizável antes de 01/01/2026;
3. somente se a abertura estiver comprovada, calcular e reconciliar o saldo por conta; caso contrário, manter os saldos indisponíveis até existir uma fonte aprovada que os forneça.
