# Leitura segura da hierarquia de categorias do Bling — 13/10/2026

## Contrato oficial consultado

A OpenAPI oficial do Bling descreve `GET /categorias/receitas-despesas` como
uma lista paginada com `pagina` e `limite`. A resposta documenta `data[]`, sem
`hasNext`, `total` ou outro marcador que prove quando a lista inteira terminou.
Cada item documenta `id`, `idCategoriaPai`, `descricao` e `tipo`; a raiz usa
`idCategoriaPai: 0`. O detalhe `GET /categorias/receitas-despesas/{idCategoria}`
devolve esses campos e a situação da categoria.

A API também lista categorias ativas e inativas. Não é seguro declarar a lista
completa porque uma página veio curta, e a sincronização não deve usar `grupoDRE`
do Bling como regra contábil do Majucau.

Referência: [OpenAPI oficial do Bling — categorias de receitas e despesas](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json), consultada em 13/10/2026.

## Mudança técnica

O cliente Bling agora oferece `WalkFinancialCategoryHierarchy`: começando pelo
ID de categoria que veio em um lançamento, consulta uma categoria de cada vez
e entrega cada objeto RAW ao chamador até alcançar a raiz. Isso permite obter
somente os pais relevantes sem carregar a lista inteira e sem depender do
marcador de fim que a OpenAPI não documenta.

O caminho valida que o ID retornado corresponde ao solicitado, exige os campos
de identidade e tipo documentados, detecta ciclos, propaga cancelamento e limita
a profundidade a 32 consultas. O limite impede que dados externos causem uma
cadeia de consultas ou retenção sem fim. Callbacks que persistirem os objetos
precisam fazê-lo dentro de uma transação e reverter se houver falha posterior.

O método só lê metadados do Bling e preserva cada objeto recebido. Ele não
classifica categorias como CAPEX, não calcula valores, não persiste dados e não
altera cards, telas ou regras financeiras. O mapeamento precisa continuar usando
IDs confirmados na conta real.

## Verificação e limites

- `go test -count=1 ./internal/integrations/bling` passou em 10,292 s.
- `go vet ./internal/integrations/bling` passou.
- Os testes usam respostas sintéticas sanitizadas; não houve chamada à conta
  Bling nem validação da hierarquia da Majucau.
- A sincronização RAW de categorias, o vínculo com contas/borderôs/pagamentos,
  a conciliação dos CAPEX pago no mês e acumulado e a confirmação da primeira
  página real continuam pendentes.

Nenhuma tela, regra financeira ou dado real foi alterado.
