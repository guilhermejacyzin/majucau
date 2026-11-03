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

Assim, mantive em branco os valores dependentes de saldo inicial (saldo atual, menor saldo, saldo projetado e semana crítica). Não usei o saldo do relatório como saldo bancário atual nem estimei um valor. A composição da T02 segue Parcial e BK-062 permanece em 25% até demonstrar a posição por conta e a equivalência com o relatório.

## Próxima verificação

Verificar se a API de lançamentos contém um registro explícito de saldo inicial/ajuste que permita reconstruir o saldo das contas aprovadas, sem alterar o intervalo usado para CAPEX. Se a resposta pública não contiver esse registro, o saldo continua pendente até existir uma fonte aprovada que o forneça.
