# Gerador Excel em fluxo — 12/10/2026

## O que foi construído

- `internal/reportexport` cria um arquivo `.xlsx` de uma planilha sem reunir
  todas as linhas na memória. A origem entrega uma linha por chamada e respeita
  o cancelamento do contexto.
- O arquivo leva a data exatamente como exibida na tela, os cabeçalhos e as
  linhas recebidas. As células são texto inline; conteúdo externo não vira
  fórmula do Excel.
- Foram aplicados limites de 64 colunas, 64 KiB por linha, 32.767 unidades
  UTF-16 por célula e o máximo de linhas do formato Excel. Texto inválido para
  XML, linhas incompatíveis e cancelamento encerram a geração com erro.
- A origem precisa aplicar esses limites antes de montar a linha em memória.
  O gerador consegue rejeitar uma linha grande depois que a recebe, mas não
  consegue evitar uma alocação excessiva feita por uma fonte defeituosa. Por
  isso o contrato agora expõe os limites e orienta futuras consultas do banco a
  verificar o tamanho antes de trazer campos para a aplicação. Essa proteção
  ainda não foi comprovada por uma consulta real, pois a origem do relatório
  não foi integrada.
- A implementação usa somente a biblioteca padrão do Go; nenhuma dependência
  nova foi adicionada.

## Verificações

- `go test ./internal/reportexport -count=1`: aprovado. Verifica o pacote XLSX,
  data, cabeçalhos, leitura linha a linha, texto externo inerte, limites e
  cancelamento.
- `go vet ./internal/reportexport`: aprovado.
- `git diff --check`: aprovado.

Na revisão de segurança desta etapa, não foram executados novos testes locais.
Os resultados acima correspondem ao commit anterior, cuja CI passou nos jobs
Windows, PostgreSQL e verificação do backlog. A CI de `81cc18a` passou em
PostgreSQL e backlog, mas a compilação Windows encontrou um teste que ainda
referenciava o nome interno anterior de um limite. A referência foi atualizada
para o nome público. A CI [38291839210](https://github.com/guilhermejacyzin/majucau/actions/runs/38291839210),
do commit `979a8e0`, passou nos jobs Windows, PostgreSQL e verificação do
backlog. Nenhum teste foi executado localmente nesta revisão.

## O que ainda falta

- O gerador ainda não está ligado às telas T13–T15 nem consulta as linhas reais.
- Antes de liberar o botão, o serviço precisa verificar `REPORT_EXPORT` para o
  usuário autenticado e produzir um arquivo temporário privado, publicando o
  resultado somente quando toda a geração terminar sem erro.
- T11 Compras continua sem exportação. Esta etapa não muda qualquer tela ou
  regra financeira.
