# Relatórios do Majucau — rascunho para dimensionamento

- **Status:** rascunho; consulta somente em tela aprovada para a versão inicial. A lista e a prioridade dos relatórios ainda precisam de validação.
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
  G --> H["Tela e relatório"]
  H --> I["Exportação auditável<br/>formato ainda a aprovar"]
~~~

O mapa técnico das APIs já lista recursos do Bling para contas a receber, contas a pagar, contas financeiras, caixas e categorias. O processamento passa por leitura autenticada, RAW, normalização, validação/reconciliação e motor financeiro. Isso define o caminho técnico geral.

**Ainda falta o mapa campo a campo** que liga cada campo real da origem à conta, classificação e linha da DRE. O próprio projeto mantém essa etapa pendente até a homologação dos campos oficiais. A demonstração pronta do Bling não substitui a DRE calculada pelo Majucau. A API financeira do Nuvem Pago também não está homologada como fonte de valores líquidos, taxas e datas de recebimento.

## Relatórios candidatos

| Relatório | O que a gestão consegue responder | Caminho dos dados | Situação e pendência |
|---|---|---|---|
| DRE do período e DRE completa | Como receita, custos e despesas chegaram ao resultado líquido? | Origem oficial → conta/categoria → mapeamento aprovado → cálculo por competência → snapshot DRE → tabela e composição T6 | Desenho T6 aprovado. Mapeamento de campos e contas ainda precisa de homologação. |
| Indicadores EBIT e EBITDA | Qual o resultado antes de juros e tributos? E antes também de depreciação, amortização e exaustão? | Linhas conciliadas da DRE → fórmulas registradas → cards EBIT/EBITDA da T6 | Dois cards aprovados para substituir a posição inferior do card UNMAPPED; o alerta de pendências continua visível. Valores ficam indisponíveis enquanto faltarem componentes validados. |
| Fluxo de caixa | Quanto entrou, saiu e qual saldo se projeta por dia? | Contas/caixas e recebíveis/pagáveis oficiais → conciliação → saldo diário → tela de fluxo | O motor possui regras de saldo; homologação de campos/fontes e reconciliação ponta a ponta continuam pendentes. |
| Capacidade para aplicação | Qual valor pode ser aplicado sem usar lucro indisponível nem comprometer o caixa? | Lucro DRE até M-2 + principal líquido já investido + menor saldo diário projetado + reserva → cálculo D-003-A → memória de cálculo | Fórmula aprovada; componentes reais e decisão sobre recebimentos já contidos no fluxo ainda precisam ser confirmados. |
| Contas a receber e a pagar | O que está aberto, vencido, realizado e por qual origem? | Contas do Bling e fontes oficialmente homologadas → normalização e conciliação → tabelas de títulos | Telas previstas; relatório/exportação detalhados ainda precisam de escopo e validação. |
| Pendências e conciliação | O que está sem mapeamento, divergente ou aguardando conferência? | Origem → validação → estado da conciliação → fila de pendências → visão de revisão | Existem estados e alertas no produto; campos, filtros e formato de saída do relatório precisam de aprovação. |

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

- Manter a tela aprovada de cada módulo como resumo.
- Em “Ver detalhes”, abrir um relatório do mesmo período e filtro, com composição e origem rastreável.
- Mostrar a fórmula e a data-base quando o relatório apresentar indicadores calculados.
- Destacar dados parciais, pendentes ou desatualizados; ausência de informação não vira zero.
- Manter a consulta dentro da tela do sistema; qualquer exportação futura precisa de aprovação separada.

## Aprovações necessárias antes de implementar relatórios

- Priorizar quais relatórios entram primeiro no protótipo.
- Confirmar se os gestores com acesso externo terão somente consulta.
- Homologar o mapeamento campo a campo das fontes oficiais antes de preencher totais financeiros.
