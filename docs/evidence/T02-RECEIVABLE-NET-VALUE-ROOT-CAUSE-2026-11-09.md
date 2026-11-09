# T02 — causa da receita líquida zerada — 09/11/2026

## Correção solicitada

Gisele encontrou a T02 aberta no protótipo com receitas em R$ 0,00 apesar de a
semana de 22 a 28/11/2026 ter títulos em aberto no Bling. A regra aprovada é
usar Contas a Receber pelo valor líquido da coluna “Líquido”, Contas a Pagar
pelos valores em aberto e o saldo conciliado do Itaú como base separada.

## Evidência no snapshot local

Na carga real com correlação `14d5fd50-3a09-460c-b36f-9b9dadcef846`:

- Contas a Receber concluiu `SUCCESS`/`NORMALIZED` às 01:58 BRT.
- Contas a Pagar concluiu `FAILED`/`RAW_ONLY` às 02:14 BRT, depois de ler
  10.301 registros. A carga completa expirou durante a preparação de detalhes.
- Não existia lote `formas-pagamentos` nessa correlação: a rotina fazia essa
  chamada depois da hidratação longa de Contas a Pagar.
- A consulta compartilhada da T01/T02 exigia que os termos de pagamento fossem
  da mesma correlação do recebível. Sem termos, `liquid_amount` ficava nulo;
  o totalizador o convertia em zero e ocultava a receita.

## Conferência do exemplo enviado

Com o snapshot de Contas a Receber e a última versão completa das taxas oficiais
da mesma conta Bling, a consulta encontrou:

| Período 22–28/11/2026 | Resultado |
| --- | ---: |
| Títulos em aberto | 146 |
| Valor bruto | R$ 45.482,22 |
| Taxas calculadas | R$ 1.691,56 |
| Valor líquido | R$ 43.790,66 |
| Títulos sem taxa conhecida | 0 |

O líquido calculado coincide com o total líquido do relatório Contas a Receber
mostrado por Gisele. A verificação é local e somente de leitura; nenhum dado
financeiro foi alterado manualmente.

## Alterações

- A T01/T02 escolhe o lote de formas de pagamento mais recente que concluiu
  `SUCCESS` e foi normalizado para a mesma conexão Bling. Se o lote da execução
  atual faltar, não descarta taxas oficiais já conhecidas por exigir correlação
  idêntica.
- A sincronização atualiza formas de pagamento e identidades das contas antes
  da hidratação extensa de Contas a Pagar, para preservar a atualização rápida
  das fontes necessárias ao líquido e à identificação de saldos.
- Teste PostgreSQL de dashboard alterado para separar a correlação de taxas e
  recebíveis; testes focados da integração Bling e worker também executados.
- Nenhuma tela, fórmula financeira, fonte dos títulos ou regra de saldo mudou.

## Validação, estado atual e limites

- `go test -count=1 ./internal/dashboard` com o PostgreSQL isolado: passou.
- `go test -count=1 ./internal/integrations/bling ./cmd/worker`: passou.
- Um teste exploratório indevido apontou `MAJUCAU_TEST_DATABASE_URL` para a
  conta restrita de runtime; ele falhou por permissão antes de mutar tabelas
  persistentes. A suíte foi repetida de forma correta: dashboard no banco
  isolado com tabelas temporárias, integração Bling/worker sem apontar ao banco
  operacional.
- Às 02:36 BRT, reconstruí e abri a T02 atualizada no protótipo real. A tela
  mostra entradas líquidas de R$ 43.790,66, saídas de -R$ 25.275,71 e saldo
  acumulado de R$ 898,42 para 22–28/11/2026. Entradas e saídas coincidem com os
  relatórios enviados pela responsável. O saldo usa a abertura semanal exibida.
- A nova carga de Contas a Pagar terminou `FAILED`/`RAW_ONLY` às 02:14; por
  segurança, o protótipo preserva o último snapshot normalizado dessa fonte. A
  semana conferida bate, mas outras semanas e uma nova carga completa ainda
  precisam ser verificadas antes de homologar a T02. T03 continua aguardando a
  aprovação da T02.
