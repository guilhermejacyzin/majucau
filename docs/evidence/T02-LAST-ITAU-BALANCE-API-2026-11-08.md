# T02 — saldo conciliado do Itaú como base do gráfico

## Regra aprovada

- O card **Saldo atual** mostra o saldo da conta Itaú no último dia conciliado.
- O mesmo saldo, com a data de referência, inicia a projeção semanal da T02.
- Para cada semana: **saldo de abertura + entradas + saídas** (saídas têm sinal
  negativo) = saldo de fechamento. O fechamento de uma semana abre a seguinte.
- Movimentos registrados recebidos pela API posteriores à data do saldo-base
  podem compor a abertura projetada do primeiro período. O card continua
  mostrando somente o último fechamento conciliado aprovado.
- O valor do print aprovado é armazenado somente no banco local isolado. Ele
  não é incluído em código, documentação versionada, exemplos ou Git.
- Até o worker localizar um fechamento conciliado mais recente e a comparação
  com Caixa e Bancos ser concluída, os indicadores permanecem **Parcial**.

## Evidência técnica

A documentação oficial de `GET /caixas` lista movimentos, datas, valores,
contas e filtros de situação. A resposta não inclui o saldo de fechamento por
dia. A rota `GET /caixas/{idCaixa}` tem um campo `saldo` que representa se o
movimento ajusta o saldo, não o total bancário da conta. Portanto, a aplicação
usa um ponto de partida aprovado e não chama a variação dos movimentos de saldo
bancário.

Fonte: [referência oficial da API Bling](https://developer.bling.com.br/referencia),
seções “Caixas e Bancos”.

## Implementação desta etapa

- A migration `000023_bling_balance_checkpoints` cria armazenamento local
  separado para pontos de saldo aprovados, com conta, data e valor.
- O worker recebe somente permissão de leitura nessa tabela. O ponto aprovado
  é provisionado localmente, fora do repositório. A migration foi aplicada ao
  PostgreSQL isolado do protótipo e a permissão de leitura foi verificada; a
  permissão de gravação do worker foi confirmada como negada.
- O card lê o checkpoint mais recente da conta identificada como Itaú.
- A série do gráfico começa pelo checkpoint e soma movimentos registrados da
  API entre a data do checkpoint e o primeiro período exibido; em seguida
  acumula entradas e saídas negativas semana a semana. Como o modelo normalizado
  não confirma a situação de conciliação de cada movimento, a projeção continua
  marcada como **Parcial** e não atualiza o card por conta própria.
- Os cards de menor saldo de 30/60 dias partem da mesma projeção. A reserva
  mínima aprovada permanece R$ 50.000.

## Limites e próximo passo

A API continua sendo a origem dos movimentos seguintes. O valor do card só
avança quando houver um novo fechamento conciliado aprovado/identificado; não
se presume que um movimento registrado seja conciliado. O build do
worker/dashboard passou, a checagem de checksum das migrations passou e a T01
foi aberta com os dados do protótipo. A sincronização posterior à mudança e a
comparação com Caixas e Bancos ainda precisam ser concluídas. A tentativa de
captura/navegação até a T02 falhou nesta sessão; sua conferência visual e a
homologação continuam pendentes.

Andamento preservado: BK-062 em 30% (70% restante); backlog geral em 12,3%
concluído (87,7% restante). Esses percentuais não significam homologação.
