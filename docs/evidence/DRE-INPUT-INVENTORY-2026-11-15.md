# Inventário inicial das fontes da DRE — 2026-11-15

## Estado e escopo

Esta é a etapa 1 prevista em `docs/source/Manual_DRE_Intelligence_Majucau_Dev.md`: localizar arquivos candidatos, registrar esquema, período, quantidade e identidade sem alterar os originais. **Setembro de 2026 é a base pronta de comparação e validação da DRE**, conforme esclarecimento de Gisele em 2026-11-15. Isso não fixa a tela nesse mês: o seletor da T06 deve governar todos os resultados e permitir qualquer competência mensal selecionada.

O inventário abaixo cobre o checkout do Intelligence, os arquivos da pasta Downloads e o conteúdo do pacote local `Setembro.zip`. Não consulta a API Bling nem lê o PostgreSQL local; portanto, a ausência de um arquivo local não prova que a mesma fonte não exista no sistema conectado.

## Estado do código e regras existentes

- `internal/financial/dre.go` contém cálculo aritmético para lançamentos já classificados. Não há, nesta inspeção, evidência de importador DRE, mapa carregado das 76 categorias ou composição de linhas DRE a partir das fontes Bling.
- O código atual possui leitura paginada e sincronização RAW para Contas a Receber, Contas a Pagar e Caixa; também possui normalizadores das listas de Receber/Pagar/Caixa. Há leitura paginada e detalhe de categorias financeiras, mas não foi encontrado um serviço de sincronização de categorias ligado à execução DRE. Isso comprova capacidade no código, não que uma execução real recente esteja completa ou que a T06 consuma esses dados.
- A auditoria de ponta a ponta de 2026-11-02 registra a T06 sem DRE calculada pelo worker e com EBIT/EBITDA de apresentação ainda sem dados confirmados.
- O novo manual especifica duas visões: espelho do painel de referência e DRE gerencial ajustada por correções aprovadas.
- A política contábil anterior continua protegendo o reconhecimento por competência/fato econômico, CMV pelo custo dos itens efetivamente vendidos, CAPEX fora do lucro e EBIT/EBITDA nos cards aprovados. O espelho da referência não substitui essa política; diferenças precisam permanecer visíveis na ponte entre as duas visões.

## Arquivos locais candidatos

Os nomes abaixo são referências locais; nenhum arquivo de dados foi copiado para o repositório.

| Fonte candidata | Evidência estrutural observada | Limite/pendência |
|---|---|---|
| `Setembro.zip` | Pacote de 1.068.582 bytes, SHA-256 `0FFA6C32F83B4247D91528EC6B0ECDF1694203B5D16FAA1F6DD67177A4EB613C`. Contém arquivos de setembro de recebimentos, Caixa e Bancos, cartão, extratos e notas fiscais. | Fonte de comparação local para setembro; reconciliar os arquivos com a API e com a DRE pronta. Não importar cópias ou fontes externas como se fossem movimento original do Bling. |
| `DRE_Gerencial_Majucau_Setembro_2026_Em_Conciliacao.xlsx` | Pasta de trabalho com abas `01 Painel`, `02 DRE`, `03 Logistica` e `04 Fontes`. | Nome indica que estava em conciliação. As abas e seus resultados não foram homologados contra as fontes de origem nesta etapa. |
| `Majucau_DRE_Setembro_2026_exportavel.zip` e cópias `(1)` e `(2)` | Os três ZIPs têm o mesmo tamanho (16.765 bytes) e o mesmo SHA-256; cada um contém `Majucau_DRE_Setembro_2026_revisada.html` (81.131 bytes). | São cópias idênticas de um resultado de referência, não três fontes independentes. Não usar o total pronto como entrada da DRE. |
| `relatorio.csv` | 4.637 registros; cabeçalhos de venda incluem data do pedido, cliente, produto, quantidade, venda bruta, desconto, frete, outras despesas, IPI e venda líquida. Os dados observados abrangem 01/09/2026 a 30/09/2026. | O vínculo de cada registro com pedido, nota, recebimento e competência contábil precisa ser validado; não somar automaticamente com outras exportações de vendas. |
| `Vendas-5a03afcc-128e-455e-b4a3-6ac407ad646a.csv` | 4.485 linhas exportadas, com datas de venda em setembro de 2026; inclui pedido, status, pagamento, envio, SKU, taxas, estorno e total líquido. | Há outra exportação da mesma competência com quantidade diferente. A sobreposição por pedido/item ainda precisa ser medida; não somar as duas. |
| `Vendas-7862b48b-dbd3-47ca-81c7-5f69729be5b1.csv` | 6.074 linhas exportadas, também com datas de venda em setembro de 2026 e campos semelhantes. | Possível cobertura ou extração diferente da exportação anterior; registros repetidos ou exclusivos precisam ser conciliados por identificadores. |
| `Vendas-076f1475-686f-4bd6-8acb-5a4548b2107d.csv` | 448 linhas, com venda entre 08/10/2026 e 10/10/2026. | É cobertura de outubro deste arquivo; não altera o período que a usuária selecionar na T06. |
| `Vendas-a0151071-e67b-413f-adac-4885ecff6b70.csv` | 12.798 linhas, com vendas em agosto de 2026. | É cobertura de agosto deste arquivo; pode ser histórico para reconstrução somente quando a regra exigir e sua cobertura for validada. |
| `Setembro/Contas a receber/Contas a Receber - recebidos (1).csv` e `(2).csv` dentro de `Setembro.zip` | 1.474 e 1.476 linhas de dados; ambos trazem vencimento, liquidação, número, nome, valor, acréscimo, juros, desconto, taxa de marketplace, recebido, saldo, conta financeira, situação e forma de pagamento. | Os arquivos diferem em duas linhas. Conferir IDs, cobertura, competência e valores contra a fonte oficial antes de decidir qual lote representa a extração completa. |
| `Setembro/CaixasBancos_Setembro2026_bling.csv` dentro de `Setembro.zip` | 106 linhas de dados; traz data, cliente/fornecedor, categoria, histórico, tipo, valor, banco, período e ID. | Confrontar com a carga API, evitar somar baixa novamente e validar `tipo_lancamento`/competência exigidos no manual. |
| `Setembro/Contas a pagar/Bling - Relatório de Contas a Pagar - Pagos.pdf` dentro de `Setembro.zip` | Relatório PDF de Contas a Pagar pagos. | Ainda não foi extraído/conciliado por lançamento; não é evidência suficiente para afirmar cobertura completa de todas as categorias da competência. |
| `Setembro/Relatório de Notas Fiscais de Saída.pdf` dentro de `Setembro.zip` | Relatório de notas fiscais de saída, 3.367.175 bytes. | Precisa ser ligado aos pedidos e às vendas sem somar NF e pedido como duas receitas. |
| `pedidos_compra_2026-11-15-17-36-57.csv` | 58 linhas, com pedido, datas, SKU, quantidade e valor unitário; datas observadas entre 01/09/2026 e 30/09/2026. | Conferir pedido/item contra notas de entrada para não duplicar compra e distinguir competência de pagamento. |
| `produtos_estrutura_2026-11-15-17-37-33.csv` e `produtos_estrutura_2026-11-15-17-38-13.csv` | Cada arquivo contém 1.254 linhas e os dois têm o mesmo SHA-256: são duplicatas byte a byte. Trazem produto final, componente, quantidade, custo unitário e operação. | A estrutura precisa de vigência histórica e validação de componentes repetidos. A duplicata deve ser importada uma única vez. |
| `relatorio_15_11_2026 (3).csv` | 276 linhas de ordens de produção, com início/término previstos, produto, quantidade e situação. | O arquivo traz datas previstas; confrontar com datas reais antes de tratar produção prevista como evento realizado. |
| `relatorio_15_11_2026 (2).csv` | 49 linhas com produto, código, data prevista, quantidade e valor. | O tipo exato de relatório e o significado dos campos precisam ser identificados antes de classificar. |
| `relatorio_15_11_2026 (4).csv` | 120 linhas; cabeçalhos indicam saldo antes de 01/09/2026, entradas, saídas e saldo em 30/09/2026 por produto. | É um resumo de estoque do mês candidato; não comprova a âncora histórica de 09/01/2026 nem os eventos detalhados de custo. |
| `relatorio_15_11_2026.csv` e `relatorio_15_11_2026 (1).csv` | 335 linhas cada, com produto, quantidade, valor unitário e valor total. | Os arquivos têm hashes diferentes. A natureza de cada relatório e a cobertura ainda não foram identificadas; não escolher um deles por suposição. |
| `Majucau_Margem_Produtos_Validada_Setembro_2026.xlsx` | Abas `Margem Produtos` (51 × 27), `Premissas` (31 × 2) e `Auditoria Custos` (50 × 13). | É apoio de margem de produtos, não substitui os eventos auditáveis de compras, produção, estoque e venda requeridos pelo CMV. |
| `PGDASD-DECLARACAO-22000026202608001..pdf` e cópia `(1)` | Os nomes identificam declaração da competência 08/2026. | Não é o PGDAS de setembro; obter a declaração da competência candidata e conferir período/valores. |

## Cobertura das 12 famílias de dados exigidas

| Família | Situação neste inventário local |
|---|---|
| Categorias e hierarquia | Nenhuma exportação local identificada. O cliente já oferece listagem paginada e detalhe de categoria; falta ligar a sincronização, conferir catálogo real e validar IDs/hierarquia. |
| Situações financeiras | Nenhum catálogo local identificado; precisa ser confirmado na API ou no lote real. |
| Recebimentos | Duas exportações locais do Bling de títulos recebidos (1.474/1.476 linhas), além de sincronização e normalização no código. Falta reconciliar versões, cobertura e competência com a API/RAW. |
| Pagamentos | Relatório PDF de contas pagas local e sincronização/normalização de Pagar no código. Falta extrair, conferir cobertura completa e confirmar campos DRE por competência/categoria. |
| Caixa e bancos | CSV local do Bling com 106 linhas; o código sincroniza/normaliza Caixa. Falta reconciliar registros, categoria, competência e baixa para impedir dupla contagem. Não reutilizar a série da T02 como cálculo DRE. |
| Vendas | Há exportações locais para setembro, com contagens diferentes e possível sobreposição; a conciliação por pedido/item ainda não foi feita. |
| NF de saída | PDF local encontrado; vínculo com vendas/pedidos e cobertura ainda precisam ser conciliados. |
| Pedidos de compra | Exportação com 58 linhas e datas de setembro localizada; conciliação com documentos fiscais ainda pendente. |
| NF de entrada | Não localizada no pacote `Setembro.zip` nem na seleção local pesquisada. |
| Produção | Exportação com 276 linhas localizada, mas contém previsões; data real e ficha precisam ser conciliadas. |
| Estruturas/fichas técnicas | Arquivo com 1.254 linhas localizado em duas cópias idênticas; exige validação e vigência. |
| Produtos/estoque e âncora | Catálogo de produtos e resumo de setembro existem localmente; não foi localizado o relatório de posição de estoque/custo da âncora de 09/01/2026 citado no manual. |

PGDAS e complementos de tributos: há somente documento identificado como competência agosto de 2026 nesta seleção. A declaração de setembro não foi localizada.

## Próximos passos seguros

1. Usar a DRE pronta de setembro de 2026 como caso-base de comparação; não transformá-la em valor fixo/default da T06.
2. Fazer o seletor de competência da T06 governar cards, tabela, gráfico, detalhes, status, conciliações e snapshot; ao mudar a seleção, consultar/recalcular os dados daquela competência.
3. Completar o manifesto com as respostas reais da API/RAW e os totais de controle das fontes de setembro usadas na comparação, sem importar duplicatas e sem alterar originais.
4. Apresentar fontes, lacunas e diferenças contra a DRE pronta de setembro antes de publicar resultado confirmado. A categoria de venda e a reconstrução do CMV têm aprovações explícitas no manual.
5. Preservar a separação entre o espelho de referência e as regras contábeis anteriores da DRE ajustada; nenhuma divergência deve ser escondida por um total forçado.

## Evidência do checkout

Na inspeção de 2026-11-15, `internal/financial/dre.go` contém somente a camada aritmética da DRE já classificada. A auditoria `docs/evidence/END-TO-END-RESULTS-AUDIT-2026-11-02.md` registra que o worker não fornece DRE calculada, depreciação ou CAPEX. Esta etapa não alterou código de produção, arquivos fonte financeiros ou valores da DRE.
