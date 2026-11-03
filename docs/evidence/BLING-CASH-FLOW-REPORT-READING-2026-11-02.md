# Leitura visual do relatório Fluxo de Caixa do Bling — 02/11/2026

## Fonte e limite

Leitura dos dados visíveis na captura enviada pela responsável. Os números abaixo
servem para conferir a origem e a apresentação; não foram gravados como dados do
Majucau nem como valores permanentes. A captura mostra totais semanais, não a
série diária completa.

## Opções visíveis no relatório

- Período: 02/11/2026 a 09/12/2026.
- Contas financeiras: 2 contas selecionadas.
- Considerar float bancário: ativado.
- Considerar prazo de recebimento da forma de pagamento: ativado.
- Considerar pagamentos e recebimentos: ativado.
- Ajuste de saldo em todas as semanas: R$ 0,00.

## Valores lidos da tabela

| Período | Saldo inicial/acumulado | Receitas | Despesas | Ajuste de saldo | Saldo acumulado ao fim |
|---|---:|---:|---:|---:|---:|
| 02/11 a 07/11 | R$ 74.463,36 | R$ 8.401,31 | -R$ 25.483,13 | R$ 0,00 | R$ 57.381,54 |
| 08/11 a 14/11 | R$ 57.381,54 | R$ 12.362,61 | -R$ 18.020,38 | R$ 0,00 | R$ 51.723,77 |
| 15/11 a 21/11 | R$ 51.723,77 | R$ 37.383,22 | -R$ 152.735,18 | R$ 0,00 | -R$ 63.628,19 |
| 22/11 a 28/11 | -R$ 63.628,19 | R$ 9.722,56 | -R$ 22.264,52 | R$ 0,00 | -R$ 76.170,15 |
| 29/11 a 05/12 | -R$ 76.170,15 | R$ 7.124,43 | -R$ 57.445,14 | R$ 0,00 | -R$ 126.490,86 |
| 06/12 a 09/12 | -R$ 126.490,86 | R$ 1.606,17 | -R$ 5.104,29 | R$ 0,00 | -R$ 129.988,98 |
| **Total mostrado pelo Bling** | **R$ 74.463,36** | **R$ 76.600,30** | **-R$ 281.052,64** | **R$ 0,00** | **-R$ 129.988,98** |

## Correções de ligação registradas

- A fonte dos resultados e da projeção da T02 é o relatório/resultado Fluxo de
  Caixa do Bling, conforme orientação da responsável.
- O saldo por banco é uma fonte distinta: Caixas e Bancos. Não deve ser
  substituído por contas a pagar/receber.
- Os cards “ENTRADAS 30 DIAS” e “SAÍDAS 30 DIAS” não podem reutilizar as métricas
  de Contas a Receber e Contas a Pagar. A captura agregada em semanas também não
  fornece precisão para recalcular uma janela móvel de 30 dias que corte uma
  dessas semanas.
- Nenhum valor da captura foi copiado para a base de dados do produto.

## Verificação adicional das opções automáticas — 02/11/2026

A página oficial do Bling sobre o MCP Server informa que um assistente conectado
pode consultar contas a pagar/receber e montar panoramas financeiros sob demanda.
Ela descreve conexão de assistente por OAuth e consultas em linguagem natural;
não documenta um contrato de resposta próprio para o relatório Fluxo de Caixa,
nem um mecanismo de execução diária para um aplicativo financeiro como o
Majucau. A própria página também lista ferramentas que criam e atualizam dados,
portanto não é uma fonte somente de leitura por definição. Não vamos usar uma
resposta gerada por assistente como valor financeiro canônico nem ligar esse MCP
ao worker de produção sem contrato, escopo somente leitura e avaliação de
segurança específicos. Referência: <https://developer.bling.com.br/mcp-server>.

Na documentação pública consultada, `GET /caixas` é o recurso de lançamentos de
Caixas e Bancos, enquanto Contas a Pagar e Contas a Receber são recursos
separados. Isso permite sincronizar movimentos e títulos, mas, por si só, não
prova que a soma reproduza as opções do relatório exibido (duas contas,
float bancário, prazo da forma de recebimento e pagamentos/recebimentos).
Referências: <https://developer.bling.com.br/referencia> e
<https://developer.bling.com.br/changelogs>.

Conclusão operacional: a T02 deve refletir o relatório Fluxo de Caixa que a
responsável indicou. A equipe pode investigar se o resultado é reproduzível a
partir de recursos oficiais, mas precisa comparar todas as faixas e saldos com
um relatório de referência antes de publicar como confirmado. Se não houver
contrato oficial ou equivalência demonstrada, a integração permanece pendente;
não se inventa endpoint nem se troca silenciosamente a fonte.

## Como o relatório é formado — confirmação na Central de Ajuda — 02/11/2026

A Central de Ajuda oficial do Bling descreve o Fluxo de Caixa como a projeção
futura das entradas e saídas de acordo com Contas a Pagar e Contas a Receber.
Também esclarece que o recebível previsto pode usar o valor líquido e que prazo
e taxa da forma de pagamento afetam data e valor. A API pública já expõe esses
componentes em recursos próprios: contas a pagar/receber, formas de pagamento
(taxa, prazo em dias, uso de dias úteis e destino) e lançamentos de caixas e
bancos.

Isso corrige uma conclusão anterior que poderia ser entendida como “não usar
Contas a Pagar/Receber”. Esses dados fazem parte da base do relatório Bling; os
totais simples dessas telas, porém, não são automaticamente o resultado do
relatório. A regra aprovada para a T02 é apresentar o resultado do Fluxo de
Caixa do Bling. A captura enviada é a referência para identificar esse resultado
e conferir a apresentação; ela não deve ser lida como imagem em cada atualização
nem usada para cadastrar valores fixos no Majucau. Fontes oficiais:
<https://ajuda.bling.com.br/hc/pt-br/articles/360038306714-Como-gerar-um-relat%C3%B3rio-de-futuras-entradas-e-sa%C3%ADdas-de-valores>,
<https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling>
e <https://developer.bling.com.br/referencia>.

Ainda falta confirmar o caminho automático e somente de leitura que entrega o
resultado do relatório. A OpenAPI pública consultada não lista uma rota
específica para esse relatório. A composição por Contas a Pagar/Receber,
formas de pagamento e movimentos é apenas uma hipótese técnica a validar: só
poderá alimentar a T02 depois de reproduzir as opções da captura e bater os
resultados do Bling. A especificação também não documenta o float do portador
nem o saldo inicial por conta. Até validar essas partes, não declarar a série
completa nem trocar silenciosamente a origem aprovada.

## Contrato público atual de caixas e bancos — 02/11/2026

Conferi a especificação OpenAPI publicada na página oficial da referência. Os
recursos de caixa/financeiro encontrados incluem `GET /caixas` (lista de
lançamentos), `GET /contas-contabeis` (lista contas financeiras), `GET
/contas/receber` e `GET /contas/pagar`. A especificação consultada não contém
rotas identificadas como relatório ou fluxo de caixa.

O item de `GET /caixas` inclui lançamento, valor, data, conta financeira e o
campo `saldo`. A descrição oficial desse campo é uma marcação `S`/`N` para
indicar se existe ajuste de saldo após o lançamento; ele não é o valor de saldo
da conta. Já a resposta de `GET /contas-contabeis` informa ID, descrição, tipo e
alias da conta financeira, sem valor de saldo. Portanto, os endpoints públicos
consultados não comprovam uma fonte para o saldo bancário atual do Itaú ou do
Mercado Pago. A tela continua mostrando esse resultado como indisponível até
uma fonte compatível ser encontrada; não se deve somar o histórico parcial de
movimentos como se fosse o saldo.

Referência oficial: <https://developer.bling.com.br/referencia> (especificação
OpenAPI e contratos dos recursos).

## Próxima etapa técnica

Investigar primeiro um caminho automático, oficial e somente de leitura para
obter o resultado do relatório Fluxo de Caixa do Bling. Não presumir que a T02
deva somar os componentes por conta própria. Se não houver uma rota do
relatório, validar a composição por Contas a Pagar/Receber, formas de pagamento
e movimentos realizados contra o relatório real, incluindo filtros, contas,
float e prazo. Só após a equivalência ser comprovada ligar os resultados à T02;
a T01 deverá reutilizar os mesmos resultados.
