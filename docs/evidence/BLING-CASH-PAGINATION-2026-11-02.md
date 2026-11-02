# Paginação real de `GET /caixas` — 02/11/2026

## Resultado observado

A execução usou somente leitura na conta Bling autorizada, com `pagina` e `limite=100`. As páginas completas retornaram o limite solicitado; a última página com dados veio menor. A API não incluiu metadados de paginação para confirmar o encerramento.

O adapter antigo continuou para a página seguinte, recebeu uma resposta vazia sem metadados e encerrou o lote com `BLING_API_PAGINATION_UNKNOWN`. O lote ficou `FAILED`/`RAW_ONLY`; nenhum movimento foi publicado na tabela normalizada e o retrato/cursor anteriores foram preservados. Os checkpoints por página permitem retomar a leitura sem expor payload financeiro no log. Quantidades e IDs reais foram omitidos deste documento.

A documentação oficial geral do Bling define `pagina` e `limite` para listas GET e informa o padrão de 100 registros por página: [Boas práticas — paginação](https://developer.bling.com.br/boas-praticas). A própria conta confirmou o comportamento do limite na execução observada.

## Segundo bloqueio encontrado nos dados

Com a paginação corrigida, a validação passou a conferir os campos reais. A API retornou `situacao` vazio, que o schema público não declara como obrigatório, e valores de `valor` assinados: todos os registros debitados observados vieram negativos e os creditados vieram positivos. O adapter antigo rejeitava os dois comportamentos. O campo de situação agora é preservado como texto de origem; sinal e `debCred` são validados juntos; a magnitude é armazenada sem sinal e a consulta aplica saída negativa pelo campo de direção. Divergências entre sinal e direção continuam rejeitando o lote. Nenhum valor financeiro bruto foi registrado neste documento.

## Correção aplicada

- `GET /caixas` passa a tratar o limite pedido e documentado como efetivo.
- Uma página não cheia encerra a leitura mesmo se a resposta não repetir metadados.
- Situação vazia ou nova é preservada sem interpretação contábil implícita.
- Saída debitada com valor negativo vira magnitude positiva mais direção `OUTFLOW`; a tela calcula a saída como negativa. Entradas seguem positivas. Sinal conflitante com `debCred` falha com segurança.
- Respostas acima do limite, inválidas ou com falha continuam cancelando a publicação.
- O snapshot RAW, a normalização, o cursor e o sucesso continuam sendo publicados atomicamente.
- Foi incluído teste sintético para uma página curta sem metadados.

## Confirmação pendente

A execução real corrigida terminou `SUCCESS`/`NORMALIZED` e publicou movimentos no banco local, confirmando a paginação e a normalização técnica. Isso não homologa os totais financeiros: a situação veio vazia e os movimentos ainda não entram como realizados até confirmação da regra; a validação visual nas telas também permanece pendente.

A tela continua incluindo somente `source_status='R'` nos totais confirmados. Os registros com `situacao` vazia ficam guardados sem classificação até a regra de inclusão ser confirmada; não devem ser somados automaticamente como realizados.

## Atualização de decisão e filtro — 02/11/2026

Gisele esclareceu que o fluxo de caixa do Bling é a fonte onde as despesas estão registradas e aprovou que resultados incompletos sejam mostrados como **Parcial**. Por isso, a nova versão passa a considerar linhas com situação `R` ou vazia nos indicadores mensais existentes e a marcar a carga como `PARTIAL` quando a situação faltar. Entradas continuam positivas, saídas negativas; nenhuma tela ou regra aprovada de sinal mudou. Situações não vazias diferentes de `R` continuam fora dos totais até serem validadas. A correção está em código, ainda precisa ser aplicada e conferida no protótipo; ver `BLING-CASH-MOVEMENT-STATUS-GAP-2026-11-02.md`.
