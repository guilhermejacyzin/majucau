# Auditoria da exportação do painel DRE — 21/11/2026

## Objetivo

Comparar a exportação local do painel consultado com a DRE de setembro/2026 que a responsável aprovou e que o padrão oficial identifica como matriz. Esta comparação é uma auditoria de versões e consistência; a exportação do painel não substitui a matriz aprovada.

## Arquivos e identidade

- **Matriz de comparação:** `Majucau_DRE_Setembro_2026_revisada.html`, indicada em `PADRAO_OFICIAL_DRE_MENSAL.md` como matriz oficial e aprovada pela responsável.
- **Exportação examinada:** `Majucau _ Dashboard de Gestão.pdf`, arquivo local de 322.998 bytes, criado/modificado em 15/11/2026 às 13:52:46 (hora de Brasília), 3 páginas, URL impressa `balancete-dashboard-siujgcjlfa-rj.a.run.app/dre`.
- **SHA-256 da exportação:** `E1BCA886BCF1794CD79A213D38235C9F91BCA915149D6AC323926111BAECBB61`.
- O PDF é uma exportação do painel com dados de setembro, mas não há prova de que sua base de dados, filtros ou versão de regras sejam os mesmos da matriz aprovada. Não foi acessado o banco do painel hospedado.

## Comparação dos totais publicados

| Linha | Matriz aprovada de setembro | PDF do painel | Diferença do PDF contra a matriz |
| --- | ---: | ---: | ---: |
| Receita bruta | R$ 432.799,67 | R$ 423.506,18 | −R$ 9.293,49 |
| Receita líquida | R$ 389.517,77 | R$ 376.190,11 | −R$ 13.327,66 |
| EBITDA gerencial | R$ 156.681,40 | R$ 117.027,63 | −R$ 39.653,77 |
| Resultado líquido gerencial | R$ 155.727,81 | R$ 114.345,79 na tabela | −R$ 41.382,02 |

Valores do PDF foram lidos na coluna **“Total do mês”** da tabela, que apresenta semanas de 01–06/09 a 28–30/09. Os totais de referência vêm da matriz identificada no padrão oficial.

## Inconsistências internas observadas no PDF

1. O cartão superior intitulado **“LUCRO LÍQUIDO PREPARADO”** mostra **R$ 65.688,55**, enquanto a linha `= Lucro Líquido` da tabela detalhada mostra **R$ 114.345,79**. Diferença interna: **R$ 48.657,24**. A exportação não explica se os indicadores usam seleções ou fórmulas diferentes.
2. O detalhamento da matriz oficial mostra pagamentos de tributos de **R$ 47.722,01** como componente de memória de deduções, mas a fórmula publicada e a receita líquida de **R$ 389.517,77** usam deduções de **R$ 43.281,90**. Essa ressalva já está registrada no padrão; a fórmula e o total da matriz são a regra de comparação, e o quadro de pagamentos de caixa deve permanecer separado.

## Conclusão operacional

- A exportação do painel **não valida nem atualiza** os resultados da T06. Ela diverge da matriz aprovada e tem uma diferença interna no lucro líquido.
- Para homologar a T06, usar como critério de setembro os totais e as regras do `PADRAO_OFICIAL_DRE_MENSAL.md` e da matriz citada por ele. A exportação do painel serve apenas como evidência auxiliar para investigar a lógica de cálculo; não copiar seus números para o Intelligence.
- A receita bruta da matriz não foi reproduzida com os candidatos locais inventariados. O `relatorio.csv` soma R$ 433.949,38 em vendas brutas e não oferece, no esquema observado, identificador de pedido ou status; as outras exportações de vendas não têm cobertura comprovada equivalente. Ainda é preciso reconciliar os registros originais por identificador, filtros e status contra a fonte do sistema de vendas.
- A próxima ação de implementação deve ser a integração das fontes e o manifesto de importação com IDs, filtros, contagens e totais de controle, mantendo a matriz como oráculo de comparação. Diferenças precisam ser explicadas por registros e fórmulas, nunca ajustadas por um total manual.

## Limitações da auditoria

Esta conferência não acessou o banco privado do painel hospedado, não executou OAuth/API Bling e não importou dados para o PostgreSQL local. A leitura do PDF confirma o que aquela exportação mostra, não a origem ou a completude de seus registros. A DRE não está homologada e a T06 não deve ser apresentada como corrigida até a conciliação por lançamento e a validação visual da responsável.
