# Relatórios do Majucau — rascunho para dimensionamento

- **Status:** rascunho; consulta somente em tela aprovada para a versão inicial. As telas já aprovadas guiam o planejamento. Ainda falta confirmar quais consultas são necessárias em cada tela.
- **Público previsto:** gestão e operação financeira em acesso de consulta. A arquitetura alvo é privada e centralizada.
- **Objetivo:** definir o que cada relatório responde, quais dados exige, de onde vêm os dados e como conferir as informações exibidas.

## Caminho de dados já mapeado

~~~mermaid
flowchart LR
  A["Fontes oficiais<br/>Bling e fontes homologadas"] --> B["Conector de leitura<br/>autenticação, paginação e retry"]
  B --> C["RAW preservado<br/>lote e origem"]
  C --> D["Normalização<br/>campos e contas"]
  D --> E["Classificação e conciliação<br/>regras aprovadas"]
  E --> F["Cálculos Majucau<br/>DRE e fluxo de caixa"]
  F --> G["Snapshots conferidos"]
  G --> H["Telas aprovadas<br/>consulta em tela"]
~~~

O mapa técnico das APIs já lista recursos do Bling para contas a receber, contas a pagar, contas financeiras, caixas e categorias. O processamento passa por leitura autenticada, RAW, normalização, validação/reconciliação e motor financeiro. Isso define o caminho técnico geral.

**Ainda falta o mapa campo a campo** que liga cada lançamento real da origem às regras contábeis aprovadas e às linhas existentes da DRE. A DRE pronta do Bling não será copiada como resultado final; os dados de origem poderão ser usados no cálculo e para comparação. As regras detalhadas ainda serão fornecidas e aprovadas. A API financeira do Nuvem Pago também não está homologada como fonte de valores líquidos, taxas e datas de recebimento.

## Ordem do painel executivo

O planejamento das consultas seguirá a ordem em que as informações aparecem na tela aprovada “Visão Executiva”, de cima para baixo:

1. **Cards do topo:** saldo atual, a receber, a pagar e menor saldo em 30/60 dias.
2. **Área central:** fluxo de caixa projetado e distribuição do saldo por conta/meio.
3. **Cards inferiores:** vendas do mês, margem bruta, estoque/cobertura, produção prevista e compras previstas.
4. **Alertas:** saldo abaixo da reserva, obrigações vencidas, recebimentos atrasados e conciliação pendente.

As telas de Fluxo de Caixa, Contas a Receber, Contas a Pagar, Estoque, Produção e Compras continuam sendo consultas próprias. A DRE é uma tela separada que consolida lançamentos financeiros depois do mapeamento contábil aprovado; suas linhas não serão alteradas. EBIT e EBITDA permanecem como os dois cards aprovados dentro da DRE.

Essa ordem organiza o detalhamento das consultas. Ela não cria telas novas nem muda fórmulas ou regras de negócio.

## Telas e consultas que precisam de detalhamento

| Relatório | O que a gestão consegue responder | Caminho dos dados | Situação e pendência |
|---|---|---|---|
| DRE do período, com cards EBIT, EBITDA e CAPEX | Como receita, custos e despesas chegaram ao resultado líquido? Quanto de CAPEX foi pago no mês e quanto foi acumulado? | Lançamentos e valores informados nos relatórios → integração ao snapshot → tabela T6 e cards fora da tabela | A DRE mantém suas linhas. EBIT/EBITDA e os dois cards CAPEX ficam fora do lucro; Gisele confirmou que os relatórios existentes informam CAPEX. O snapshot atual ainda não importa esses valores; conectar os dados do relatório é a pendência. |
| Fluxo de caixa | Quanto entrou, saiu e qual saldo se projeta por dia? | Contas/caixas e recebíveis/pagáveis oficiais → conciliação → saldo diário → tela de fluxo | O motor possui regras de saldo; homologação de campos/fontes e reconciliação ponta a ponta continuam pendentes. |
| Capacidade para aplicação | Qual valor pode ser aplicado sem usar lucro indisponível nem comprometer o caixa? | Lucro DRE até M-2 + principal líquido já investido + menor saldo diário projetado + reserva → cálculo D-003-A → memória de cálculo | Fórmula aprovada; componentes reais e decisão sobre recebimentos já contidos no fluxo ainda precisam ser confirmados. |
| Contas a receber e a pagar | O que está aberto, vencido, realizado e por qual origem? | Contas do Bling e fontes oficialmente homologadas → normalização e conciliação → tabelas das telas | Telas previstas; detalhes, filtros e conferência de dados ainda precisam de validação. |
| Pendências e conciliação | O que está sem mapeamento, divergente ou aguardando conferência? | Origem → validação → estado da conciliação → fila de pendências → tela de revisão | Existem estados e alertas no produto; campos e filtros da consulta ainda precisam de aprovação. |

## O que cada relatório precisa especificar

1. **Quem vai consultar:** chefias, financeiro ou operação.
2. **Pergunta respondida:** por exemplo, “qual foi o resultado de agosto?”.
3. **Período e filtros:** competência, data, origem, situação e conta, quando aplicáveis.
4. **Detalhe:** somente totais, linhas da DRE, ou também abertura até conta e documento de origem.
5. **Comparação:** período anterior, orçamento ou realizado versus projetado.
6. **Saída:** consulta somente na tela na versão inicial. PDF e Excel ficam fora do escopo até nova aprovação.
7. **Conferência:** total, origem, atualização, incompletudes e explicação de valores indisponíveis.
8. **Acesso:** consulta sem permissão para lançar ou alterar dados; pessoas e perfis precisam ser definidos.

## Proposta inicial de desenho

- Usar as telas já aprovadas como guia das consultas; não propor telas novas sem aprovação.
- A DRE apresenta o resultado calculado pelo Majucau segundo regras contábeis aprovadas; a DRE pronta do Bling serve apenas para comparação, não como resultado final.
- Cada tela mantém sua finalidade. Fluxo de caixa e calculadora de aplicação, por exemplo, são consultas diferentes.
- Em “Ver detalhes”, quando já existir essa ação aprovada, abrir os dados do mesmo período e filtro, com composição e origem rastreável.
- Mostrar a fórmula e a data-base quando o relatório apresentar indicadores calculados.
- Destacar dados parciais, pendentes ou desatualizados; ausência de informação não vira zero.
- Manter a consulta dentro da tela do sistema; qualquer exportação futura precisa de aprovação separada.

## Aprovações necessárias antes de implementar relatórios

- Confirmar quais detalhes e filtros são necessários dentro de cada tela existente.
- Confirmar se os gestores com acesso externo terão somente consulta.
- Homologar o mapeamento campo a campo das fontes oficiais antes de preencher totais financeiros.
