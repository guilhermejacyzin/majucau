# T02 — armazenamento tipado das condições de recebimento — 03/11/2026

## Objetivo

Guardar os dados de API que o relatório **Fluxo de Caixa** do Bling usa para
ajustar as entradas previstas. Essa etapa prepara a reprodução do relatório; não
publica valores estimados na tela nem altera a apresentação aprovada.

## Implementação

- A resposta individual de cada forma de pagamento já era consultada durante a
  sincronização, mas os campos calculados eram descartados após o lote.
- Foi criada `bling_payment_method_terms`, com vínculo à conexão, lote e registro
  RAW de origem.
- São guardados destino, taxa percentual, taxa fixa, prazo, configuração de dias
  úteis e indicadores `*_known`. Um valor ausente permanece distinguível de um
  zero configurado.
- O `UPSERT` usa a identidade Bling da forma de pagamento e o lote diário; a
  repetição da sincronização atualiza a mesma linha.
- A publicação tipada acontece na transação final do lote, mantendo o padrão
  atômico. A rotina diária agora só considera a leitura de formas de pagamento
  concluída quando essa etapa normalizada termina.

## Verificações

- `go test ./internal/integrations/bling ./internal/financial`: passou.
- A validação da migration, a execução real PostgreSQL e a CI ainda precisam
  confirmar a entrega completa.

## Limites ainda em aberto na T02

Esta mudança não compõe entradas e saídas por data, não resolve o saldo inicial
do relatório, não calcula semana crítica ou menores saldos, e não liga o gráfico
à tela. Ainda é necessário ler os termos normalizados junto aos títulos em
aberto, reproduzir as regras do relatório e conferir todos os períodos com uma
execução atual do relatório Bling. Dias úteis continuam indisponíveis sem um
calendário que trate feriados. Não inserir os valores da captura como dados
ativos.
