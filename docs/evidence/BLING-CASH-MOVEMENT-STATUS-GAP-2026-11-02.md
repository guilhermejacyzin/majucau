# Registros do fluxo de caixa com situação vazia — 02/11/2026

## O que foi observado

Na instância local isolada do protótipo, a integração Bling aparece conectada. A última carga de `GET /caixas` terminou `SUCCESS`/`NORMALIZED` e guardou 31.032 movimentos, todos com `situacao` vazia: 14.385 entradas e 16.647 saídas. Os registros cobrem 01/01/2026 a 30/10/2026 e identificam nove contas financeiras. Não foram copiados valores, nomes de contas ou outros dados pessoais para este documento.

A referência oficial do Bling descreve `GET /caixas` como listagem de lançamentos de caixas e bancos e inclui `debCred`, `situacao`, `valor`, `data` e `contaFinanceira` no contrato. A descrição pública de `situacao` não explica o caso vazio: [referência oficial da API](https://developer.bling.com.br/referencia). A responsável confirmou que o fluxo de caixa do Bling é a fonte onde os movimentos de despesas estão registrados e aprovou exibir valores incompletos como **Parcial**.

## Causa da ausência dos valores

O adapter já normalizava os 31.032 movimentos, mas o resumo do painel somava somente linhas com `source_status = 'R'`. Como todas as linhas reais tinham situação vazia, o filtro descartava todas e produzia métricas sem valor, mesmo havendo registros válidos na tabela de movimentos.

## Correção nesta revisão

- Os indicadores mensais existentes passam a incluir situação `R` ou vazia.
- A carga sinaliza `PARTIAL` quando a situação não veio informada, para o painel não apresentar o resultado como integralmente confirmado.
- Entrada continua positiva; saída continua negativa. Nenhuma tela, texto, posição de card ou regra aprovada de sinal foi alterada.
- Situações diferentes de `R` ou vazia continuam fora dos totais até o significado ser confirmado.
- A nova versão ainda precisa ser instalada no protótipo e sincronizada para a conferência visual com os dados Bling.

## Verificações executadas

- `go test ./internal/dashboard ./internal/integrations/bling` — passou.
- `TestReadStreamSummaryPostgresRequiresNormalizedDailyStages` no PostgreSQL isolado do protótipo — passou. Esse teste cria as tabelas como temporárias dentro de uma transação e faz rollback; não altera os dados persistentes.
- `git diff --check` — passou.

## O que continua pendente

- Contas a Receber chegou em `RAW_ONLY`; ainda não existem registros normalizados de Contas a Receber no banco.
- Contas a Pagar tem um lote antigo com falha publicado; a tentativa mais recente ainda está numa transação sem publicação confirmada. O código atual aceita o sinal negativo como saída, mas falta uma carga final para comprovar com os dados reais.
- A build que está aberta foi iniciada antes desta correção; a tela só refletirá a mudança após reiniciar com a nova versão e conferir o resultado.
- O saldo atual por banco exige uma fonte própria; não pode ser calculado somente somando os movimentos históricos.
- O fluxo de caixa alimenta o realizado financeiro. Ele não substitui o mapeamento contábil próprio da DRE nem a identificação aprovada de CAPEX.

Esta alteração corrige um filtro de dados e registra a decisão operacional de mostrar os movimentos do Bling com aviso **Parcial** quando a API não informar a situação. Não declara o protótipo homologado.
