# Auditoria T02 — divergência de saldo e atualização por conciliação

Atualizado em 08/11/2026. Este registro substitui a conclusão anterior de que
a abertura da primeira semana deveria sempre repetir o valor do card, sem
considerar movimentos conciliados ocorridos depois da data-base.

## O que os prints comprovam

- A fórmula de cada semana mostrada é internamente consistente:
  saldo de abertura + entradas + saídas (negativas) = saldo de fechamento.
- Ao mudar o início do relatório de 07/11 para 08/11, o movimento do dia
  anterior deixa de aparecer na tabela, mas o saldo de abertura mostrado ainda
  é o saldo anterior a esse movimento. A diferença se propaga nos fechamentos
  seguintes. Portanto, a divergência final não nasce apenas na última semana.
- O relatório com outro início de período não pode ser usado como comparação
  direta sem alinhar datas, conta e opções. O print da tela serve como evidência
  do problema de corte; não prova sozinho se o movimento anterior estava
  conciliado.

## Causa encontrada no Majucau

O código estava usando `situacao=R`, que na API significa **lançamento
registrado**. Isso não significa **lançamento conciliado**. A listagem de
`GET /caixas` não devolve o estado da conciliação dentro de cada item. Por isso,
o código não tinha como separar com segurança o movimento que pode atualizar o
saldo-base daquele que apenas aparece no fluxo.

O problema era técnico: faltava sincronizar a lista oficial filtrada por
conciliação e aplicar o saldo-base na data correta. Não era necessário ler a
tela do Bling ao vivo para descobrir a situação de cada movimento.

## Como o dado será obtido e atualizado

1. Às 09:30, em `America/Sao_Paulo`, o worker lê todas as páginas de
   `GET /caixas` usando o filtro oficial `situacaoConciliacao=1` (conciliados),
   desde o início do histórico aprovado.
2. O Majucau grava, em um retrato separado por execução diária, os IDs retornados
   pelo filtro. O conjunto contém apenas a identificação necessária para
   reconhecer os conciliados; os lançamentos e valores continuam vindo da
   leitura oficial geral de movimentos.
3. O card de saldo parte do último checkpoint de saldo do Itaú aprovado e soma
   somente os movimentos conciliados posteriores até a data consultada.
4. Na primeira semana visível, a abertura parte do mesmo checkpoint e carrega
   somente os movimentos conciliados anteriores ao início escolhido. A partir
   daí, cada fechamento semanal alimenta a semana seguinte. Movimentos do
   período continuam compondo entradas e saídas conforme o relatório.
5. Se a consulta não terminar, as páginas estiverem incompletas ou os retratos
   não pertencerem à mesma execução, o Majucau não avança o saldo com dados
   misturados: mantém o último dado válido e sinaliza estado parcial.

Conciliações feitas depois das 09:30 entram na próxima execução diária das
09:30. A execução é idempotente e relê o histórico completo; assim, não depende
de um evento de navegador nem de a usuária mandar outro print.

## Implementação entregue nesta alteração

- Filtro de conciliação validado pelo contrato OpenAPI oficial: `1` conciliado,
  `2` não conciliado, `3` todos. O parâmetro anterior `conciliados` foi
  substituído por `situacaoConciliacao` no changelog oficial.
- Nova sincronização paginada `caixas.conciliadas`, com a mesma data de
  referência/correlation ID da carga diária e publicação normalizada própria.
- Nova tabela de retrato de IDs conciliados, com privilégios explícitos de
  leitura e gravação do worker e migration forward-only `000024`.
- A carga diária só se considera concluída quando o retrato geral e o de
  conciliados estiverem prontos na mesma execução.
- O cálculo de saldo do card e da abertura da semana usa o checkpoint aprovado
  mais os movimentos conciliados da conta Itaú. A tabela e a fórmula semanal
  permanecem como aprovadas; nenhuma regra contábil foi alterada.
- A captura mais recente mostrou os rótulos de dia ainda cortados apesar do
  ajuste anterior. A correção atual ampliou o `viewBox` do SVG para 210 unidades
  e posicionou os rótulos em y=155. Testes verificam as datas no gráfico da T02
  e que o texto fica dentro do `viewBox`; a conferência visual nativa permanece
  pendente porque a captura da janela falhou por tempo limite.
- Em nova revisão, a responsável informou que o eixo ainda estava cortado. O
  rótulo de cada faixa passou a ocupar duas linhas (data inicial e final) para
  não comprimir nem sobrepor as semanas. A suíte de interface e o build foram
  repetidos após essa alteração; a conferência visual no aplicativo segue
  necessária.
- Como ainda aparecia cortado, os dois dias passaram a ser textos SVG separados
  em coordenadas verticais próprias, sem `tspan`. Os testes validam o início e o
  fim do intervalo em ambos os usos do gráfico (T01 e T02). A compilação Wails
  foi refeita; a captura de tela continua indisponível nesta sessão.

## Limites de validação

Os 73 testes frontend, o lint/build frontend, os testes Go focados e o build do
worker confirmam código e cálculo localmente. A migration 000024 ainda não foi
aplicada ao banco do protótipo. A nova consulta real com a
conta autorizada ainda precisa completar uma execução diária para
confirmar se o movimento anterior à janela estava conciliado e conferir a
T02 com o relatório Bling para as mesmas datas e opções. Até esse retorno, o
saldo calculado continua parcial, sem valor inventado. Nenhum saldo ou dado
financeiro real foi gravado neste arquivo ou em outros arquivos do Git.

Fontes oficiais: [referência OpenAPI do Bling](https://developer.bling.com.br/referencia),
[changelog do filtro de conciliação](https://developer.bling.com.br/changelogs).

## Atualização de apresentação do gráfico — 08/11/2026

As notas antigas sobre duas linhas de datas foram substituídas: um ponto representa uma faixa semanal e recebe um único intervalo em uma linha; em séries densas usa-se DD/MM–DD/MM, mantendo a data completa no detalhe. A semana sob o ponteiro recebe uma linha vertical alinhada às séries. Tooltip identifica Despesas, Receitas e Saldo. Gisele confirmou que a apresentação ficou correta; esse aceite não homologa o saldo nem a sincronização.
