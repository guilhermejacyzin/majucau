# T02 — retomada de sincronização Bling interrompida

**Data:** 15/11/2026  
**Estado:** correção de retomada implementada; execução real completa ainda pendente.

## Causa operacional registrada

A leitura de Contas a Pagar havia avançado pelas páginas da lista e da hidratação
de detalhes, mas parou antes de normalizar e publicar. O banco local manteve
checkpoints RAW da carga interrompida. Uma nova atualização manual criava outro
identificador de correlação, o que impedia a retomada desses checkpoints e
obrigava a reler as mesmas fontes.

## Correção

- A atualização completa manual procura a carga integral mais recente de Bling
  que tenha checkpoint de Contas a Pagar, ainda não tenha publicado essa fonte
  como `NORMALIZED`, e contenha as demais fontes necessárias da mesma execução.
- Quando encontra essa carga, reutiliza sua identidade diária e seus
  checkpoints.
- Fontes dessa mesma execução já publicadas e sem checkpoint pendente são
  reaproveitadas, evitando repetição de leituras concluídas.
- Se não houver carga retomável, a atualização usa uma nova correlação como
  antes.

## Validação e limites

- `go test -count=1 ./internal/integrations/bling ./cmd/worker` passou no ciclo
  anterior da correção.
- Foi criado um teste de integração PostgreSQL que seleciona a carga pendente
  mais nova, deixa de selecioná-la após publicação e confirma que uma fonte já
  publicada não é consultada de novo.
- O teste de integração só executa com `MAJUCAU_TEST_DATABASE_URL` apontando
  para a base descartável `majucau_test`; nesta máquina não foi comprovado que
  esse banco estava configurado e o teste pode ser ignorado.
- A rotina ainda não foi executada contra a conta real depois da mudança. O
  último snapshot publicado de Contas a Pagar permanece o de 08/11/2026 às
  11:00; o painel e os valores da T02 continuam sem homologação.

Esta evidência descreve a retomada do processo, não uma carga concluída, nem a
confirmação dos totais financeiros.
