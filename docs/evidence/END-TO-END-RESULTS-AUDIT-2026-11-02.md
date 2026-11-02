# Auditoria do caminho dos dados até as telas — 02/11/2026

## Objetivo e limites

Identificar por que dados recebidos das integrações ainda não viram resultados
nas telas T01–T17. Esta auditoria é de código, contratos e evidências registradas;
não altera tela, regra financeira, dado real nem credencial. A janela visual do
protótipo expirou em duas tentativas, então não há nova confirmação visual nesta
execução.

## Conclusão curta

O problema principal não é simplesmente “a API não respondeu”. O Majucau tem
código para buscar e guardar respostas do Bling, mas várias respostas param em
RAW (dados originais sem transformação financeira). O painel consulta tabelas
normalizadas e o contrato do snapshot não contém saldos por conta nem série de
projeção. Assim, uma chamada HTTP bem-sucedida pode coexistir com “—”/“Sem dados”
na tela.

Na execução real registrada antes da correção de 02/11, o Bling concluiu a leitura de 14.891
registros de Contas a Receber como `SUCCESS / RAW_ONLY`: houve resposta, mas
esses dados ainda não foram transformados para alimentar os valores do painel.
Na mesma execução, Contas a Pagar leu 10.198 títulos e teve o lote revertido
porque um título tinha `saldo` negativo. Esse campo é o saldo em aberto de um
título individual a pagar. Ele não é o saldo bancário do Itaú ou do Mercado
Pago. A evidência não autoriza tratar o caixa/banco como negativo.

Gisele informou que ainda não colocou arquivo na pasta externa da Nuvem. Essa
origem deve ficar sem resultado até receber um arquivo e não deve bloquear os
resultados disponíveis do Bling.

## Caminho, etapa por etapa

| Etapa | O que o código faz | Resultado da auditoria |
|---|---|---|
| 1. Autorização | OAuth e tokens são tratados pelo worker; a tela e o servidor têm estados de conexão separados. | O worker e o protótipo estavam em execução nesta máquina. Isso, sozinho, não prova OAuth ativo. O registro real de 02/11 prova que pelo menos uma leitura de Contas a Pagar chegou ao Bling. O estado OAuth atual não foi relido nesta auditoria. |
| 2. Leitura Bling | A carga diária prevista para 09:30 consulta Contas a Receber, Contas a Pagar e `GET /caixas`. | Os recursos são chamados separadamente e com retry/checkpoint. A leitura de `/caixas` ainda não tem homologação real do fim da paginação. |
| 3. Guarda RAW | As páginas são armazenadas com versão, hash e lote para preservar o retorno original. | RAW é evidência, não resultado financeiro. Contas a Receber ainda fica em `RAW_ONLY`; `/caixas` mantém a resposta RAW para rastreabilidade e agora também publica movimentos normalizados. |
| 4. Normalização | O conversor valida os campos e grava fatos tipados usados pelo sistema. | Contas a Receber continua RAW_ONLY. Desde o commit `9696bd9`, `/caixas` tem conversão idempotente para movimentos tipados e os indicadores mensais de realizado da T02 usam essa fonte. Contas a Pagar converte saldo em aberto para saída negativa e inclui os valores nas consultas de T01/T04. A CI PostgreSQL ainda precisa validar a migration; a sincronização real com o código novo não foi repetida. |
| 5. Cálculo | O leitor do painel soma tabelas normalizadas; a projeção exige saldo inicial conciliado e movimentos. | O helper D0–D+60 existe, mas não é chamado por um fluxo de produção, não grava `daily_balances` e não fornece a projeção do T02. |
| 6. Snapshot | O worker transmite ao protótipo os resultados já calculados no banco. | `DashboardSnapshot` (`internal/application/contracts.go`) oferece totais de receber/pagar e alguns realizados, além de campos opcionais CAPEX. Não oferece saldo por banco, série diária, semana crítica ou menor saldo. |
| 7. Tela | Os cards escolhem campos do snapshot pelo nome e exibem o mesmo resultado no detalhe e na Visão Executiva. | `metricForCard` (`frontend/src/mockupDashboard.tsx`) liga somente alguns cards a métricas do snapshot. Gráficos e vários painéis continuam sem dados dinâmicos. T01 não tem dados de origem que possa consolidar nesses espaços. |

## Situação das fontes financeiras relevantes

### Correção confirmada sobre a origem do fluxo de caixa

Gisele confirmou que o Bling já é usado para registrar o fluxo de caixa e as
despesas. Portanto, essas despesas devem ser buscadas no Bling; a pasta externa
da Nuvem não é uma fonte substituta para elas.

Na referência oficial consultada em 02/11/2026, `GET /caixas` devolve os
lançamentos com identificador, direção débito/crédito, situação, valor, data,
conta financeira e origem. O detalhe do lançamento também informa competência,
categoria e indicador de transferência. A rota `/contas-contabeis` fornece
identificador, descrição e tipo das contas financeiras; o contrato consultado
não fornece um campo de saldo atual nessa listagem. O adapter Majucau já chama
`GET /caixas`, mas publica essas páginas como RAW_ONLY e ainda não consulta a
lista de contas para transformar o ID em nomes como Itaú e Mercado Pago.

Isso confirma que despesas e movimentos têm uma fonte no Bling; não confirma,
por si só, que o saldo atual de cada banco possa ser calculado sem uma base
inicial conciliada e a comprovação de histórico completo. Essa verificação faz
parte do próximo passo técnico. A referência consultada foi a [OpenAPI oficial
do Bling](https://developer.bling.com.br/referencia), rotas `GET /caixas`,
`GET /caixas/{idCaixa}` e `GET /contas-contabeis`.

| Fonte | Estado do caminho atual | O que falta para aparecer resultado |
|---|---|---|
| Itaú no Bling | Os movimentos de `/caixas` já são normalizados com o identificador de conta; o snapshot ainda não expõe saldo atual separado por instituição. | Confirmar o ID/nome da conta, provar paginação completa e obter saldo inicial conciliado ou saldo atual oficial da fonte. Não inferir saldo bancário a partir de contas a pagar. |
| Mercado Pago no Bling | Mesmo estado do Itaú; o identificador da conta é preservado no movimento. | Confirmar o ID/nome da conta, provar paginação completa e obter saldo inicial conciliado ou saldo atual oficial da fonte, mantendo Itaú e Mercado Pago separados. |
| Nuvem externa | Gisele informou que a pasta ainda está vazia. | Não exibir valor Nuvem agora. Quando o arquivo existir, validar sua estrutura e ligar a fonte conforme seu conteúdo; a falta do arquivo não bloqueia Itaú/Mercado Pago. |
| Contas a Receber | A execução real mais recente registrada leu 14.891 registros e concluiu `SUCCESS / RAW_ONLY`; o snapshot consulta a tabela normalizada `receivables`, não essas páginas RAW. | Ligar um normalizador da API, manter a classificação aprovada B2B/B2C, persistir idempotentemente e validar a reconciliação. |
| Contas a Pagar | O último lote real registrado antes da correção foi revertido ao encontrar saldo negativo. O commit `9696bd9` agora interpreta conta a pagar como saída negativa, preserva o valor original do título e permite o saldo assinado nas consultas de T01/T04. | Aplicar a migration `000016` no banco isolado, atualizar o worker, repetir a sincronização e comparar totais/linhas com o Bling. A decisão de sinal já está confirmada; a validação real permanece pendente. |
| Pagamentos/recebimentos realizados | A T02 calcula os totais mensais realizados a partir de `bling_cash_movements`: crédito entra positivo e débito sai negativo. O lote normalizado preserva origem, conta, data e valor. | Confirmar fim da paginação e reconciliar por conta/período com o Bling. Evitar duplicar o mesmo movimento em outras fontes ou tabelas. |

## O que cada grupo de tela recebe hoje

| Tela | Ligação encontrada | Situação do resultado |
|---|---|---|
| T01 — Visão Executiva | A Receber e A Pagar são ligados a métricas do snapshot. Saldo atual, menor saldo, gráfico e distribuição não têm campos no snapshot. | A RECEBER depende de Contas a Receber normalizada; A PAGAR depende do último snapshot válido. Caixa/projeção não chegam à tela. |
| T02 — Fluxo de Caixa | Recebido no Mês e Pago no Mês usam movimentos realizados normalizados de `/caixas`, com crédito positivo e débito negativo. | A mudança foi enviada, mas aguarda CI PostgreSQL, atualização/migration do protótipo e conferência real. Saldo atual por banco, entradas/saídas projetadas, semana crítica e menor saldo 30/60 dias seguem sem validação de ponta a ponta. |
| T03 — Contas a Receber | Usa totais e linhas do snapshot de recebíveis. | O pipeline da API Bling permanece RAW_ONLY; não há resultado B2B confirmado da API nesta tela. |
| T04 — Contas a Pagar | Tem totais, próximos vencimentos e linhas ligados ao snapshot normalizado; o código atualizado aceita a saída negativa confirmada pela responsável. | A carga real anterior foi revertida por esse caso. Falta aplicar `000016`, sincronizar novamente no protótipo e conferir total, vencimentos e linhas na tela, sem alterar seu desenho. |
| T05 — Conciliação | Tela removida por decisão aprovada. | Continua removida; nenhuma reintrodução proposta. |
| T06 — DRE | EBIT/EBITDA têm cards separados; CAPEX tem campos opcionais no snapshot. | Os cards EBIT/EBITDA são placeholders. O worker não fornece DRE calculada, depreciação ou CAPEX. Requisito contábil permanece bloqueado por fonte/mapeamento e reconciliação. |
| T07 — Balancete | Não há resultado contábil no snapshot. | Plano de contas, saldos iniciais e partidas balanceadas ainda não alimentam a tela. |
| T08 — Calculadora de Aplicação | Depende de lucro elegível e projeção de caixa. | Sem DRE real, menor saldo D0–D+60 e saldo transferível, não há cálculo de produção. |
| T09–T11 — Estoque, Produção e Compras | Não há métricas desses módulos no contrato do snapshot executivo. T11 é somente consulta conforme decisão aprovada. | T01 não consolida resultados que ainda não chegam das telas de origem. |
| T12–T14 — Forecast, Pendências e Alertas | Não há cobertura de resultados operacionais/alertas dinâmicos no snapshot financeiro atual. | Conteúdo financeiro continua sem fonte/snapshot suficiente; textos estáticos não comprovam cálculo ou alerta real. |
| T15–T17 — Integrações, Parâmetros e Usuários | Integrações pode apresentar estado/configuração; parâmetros e usuários têm seus próprios cadastros/fluxos. | Esses módulos não fornecem saldos, DRE, forecast ou outros resultados para T01. |

## Causa imediata de os números não aparecerem

1. A conexão/execução de API é uma etapa separada da autorização OAuth; o
   processo local pode estar rodando sem que cada recurso esteja autorizado ou
   sincronizado.
2. A leitura de uma página só cria evidência RAW. Ainda é preciso converter os
   campos para as tabelas financeiras tipadas.
3. A consulta da tela usa as tabelas tipadas e exige um lote normalizado recente;
   ela não lê diretamente todo o JSON RAW para inventar um total.
4. Não há no snapshot do painel os campos necessários para saldo por banco e
   fluxo D0–D+60.
5. A interface está ligada a apenas alguns campos; gráficos, distribuição e
   muitos cards são placeholders explícitos.
6. T01 consolida a origem, então não consegue preencher lacunas que T02/T03/T06
   ainda não produzem. Uma busca direta de T01 ao Bling contrariaria a regra
   aprovada.

## Ordem técnica necessária para destravar resultados reais

1. Confirmar o estado OAuth atual por leitura de status no worker, sem coletar
   senha ou segredo.
2. Conferir o contrato vigente de Contas Financeiras e o detalhe real sanitizado
   de `/caixas`; identificar account ID, datas, valores e marcador de fim de
   paginação.
3. Criar o normalizador por conta Itaú/Mercado Pago e publicar atomicamente só
   quando todas as páginas e campos obrigatórios forem válidos.
4. Tornar utilizável um saldo inicial conciliado por conta, sem pressupor que
   ausência seja zero.
5. Ligar pagamentos/recebimentos realizados e recebíveis/pagáveis ao motor de
   caixa; gravar os 61 saldos diários com lineage e status.
6. Expor saldo por conta e série no snapshot T02; ligar os componentes já
   desenhados sem redesenhar a tela. A fonte Nuvem fica ausente enquanto a pasta
   estiver vazia.
7. Fazer T01 consumir os mesmos resultados do T02, sem fórmula ou consulta
   paralela.
8. Validar totais por conta e por período contra Bling, depois demonstrar a
   mesma posição nas telas T02 e T01.

Nenhum desses passos permite mostrar um número presumido. O resultado só pode
ser confirmado depois da cadeia inteira **fonte → RAW → normalizado → cálculo →
snapshot → tela** estar conectada e reconciliada.

## Atualização após a auditoria — normalização mensal de Caixa/Bancos

Em 02/11/2026, foi implementada uma parte do caminho apontado acima: a carga
diária normaliza `GET /caixas` em movimentos tipados e idempotentes, e a T02
passa a calcular recebimentos realizados positivos e pagamentos realizados
negativos por mês. A sincronização desse recurso roda mesmo se a carga de
Contas a Pagar falhar. A implementação e seus limites estão detalhados em
`docs/evidence/BLING-CASH-MOVEMENTS-NORMALIZATION-2026-11-02.md`.

Esta atualização não altera o fato registrado na auditoria: o novo código ainda
precisa ser publicado, instalado no protótipo, executado contra a conta Bling e
comparado com os valores de origem. Também não implementa saldo bancário atual,
saldo por Itaú/Mercado Pago, projeção D0–D+60, normalização de Contas a Receber
nem os demais resultados T02–T17. A auditoria visual do protótipo continua sem
nova captura.

## Atualização — sinal de Contas a Pagar esclarecido e corrigido

Gisele confirmou em 02/11 que conta a pagar representa saída e deve aparecer
negativa no resultado financeiro. O normalizador e a regra do banco que
rejeitavam o saldo negativo foram ajustados para guardar a obrigação como
saída negativa e incluir esses valores nos totais, vencimentos e linhas da T04
e no indicador correspondente da T01. Isso não representa saldo bancário.
O commit `9696bd9` já está publicado na branch `codex/majucau-artifact-consolidation`.
As verificações locais passaram; na última consulta, os jobs Windows e
PostgreSQL da CI ainda estavam em execução. O protótipo aberto continuava usando
binários anteriores, e a migration/sincronização reais ainda não tinham sido
executadas. Portanto, a correção está no código publicado, mas os valores
reais nas telas ainda não foram comprovados.
