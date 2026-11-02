# Contas a Pagar do Bling com categoria temporariamente indisponível — 02/11/2026

## Estado real confirmado antes da mudança

- A autorização do Bling estava conectada e o teste de leitura respondia.
- A recuperação diária das 09:30 falhava com `BLING_CATEGORY_SCOPE_INSUFFICIENT`: o Bling negava a leitura do detalhe de categorias.
- Por causa dessa consulta, a carga transacional inteira de Contas a Pagar era revertida. O painel mostrava seus valores como indisponíveis.
- Contas a Receber e movimentações de Caixa também tinham leitura RAW sem transformação completa para os resultados do painel.

## Correção implementada

- Quando apenas a consulta de detalhe de categoria retorna o erro de escopo já identificado, a normalização continua com os campos financeiros validados da conta: valor, saldo em aberto, vencimento, fornecedor, documento e situação.
- O registro permanece sem nome de categoria e o lote é publicado como `PARTIAL`, com mensagem sanitizada. A transação inclui o retrato financeiro, sua associação à execução diária e o cursor, sem publicar uma carga como completa.
- Valores de Contas a Pagar chegam ao painel acompanhados do estado `Parcial`. A tela e os cálculos existentes não foram redesenhados.
- Erros em dados financeiros obrigatórios, detalhe de título/fornecedor ou paginação continuam revertendo a tentativa e preservando o último retrato válido.
- Uma carga parcial normalizada é reconhecida pelo mecanismo de retomada idempotente. Uma nova falha de sincronização ainda pode ser repetida com segurança.
- A categoria ausente continua identificada como pendência para resultados que precisem dessa classificação, segundo as regras aprovadas. Nenhuma categoria ou regra contábil é inferida.

## Verificação local

- `go test -count=1 ./internal/dashboard ./internal/integrations/bling ./cmd/worker`: passou em 02/11/2026. Os testes PostgreSQL que dependem de `MAJUCAU_TEST_DATABASE_URL` foram compilados, mas não executaram localmente sem essa configuração.
- `go vet ./internal/dashboard ./internal/integrations/bling ./cmd/worker`: passou.
- `git diff --check`: passou.
- A CI com PostgreSQL ainda precisa confirmar a publicação parcial, a retomada idempotente e a preservação do retrato após erros em outros campos.

## Homologação real ainda pendente

Essa alteração está preparada para a execução real do Bling, mas o painel só pode ser considerado funcionando após reiniciar o worker com o código novo, concluir a leitura paginada e confirmar valores e estado `Parcial` na interface. Valores financeiros reais não foram copiados para este arquivo ou para o Git.

A leitura de categorias continua necessária para preencher essa coluna sem pendência. Recebimentos e Caixa ainda precisam de normalização/homologação para alimentar os respectivos resultados. Nenhuma tela aprovada ou regra financeira foi alterada.

## Reexecução no protótipo — 02/11/2026, 17:10 BRT

- Reiniciei o protótipo e confirmei pelo worker `BLING=CONNECTED`. A leitura diária percorreu 14.884 contas a receber e concluiu esse recurso.
- Contas a Pagar leu 10.198 registros, mas o primeiro detalhe não foi normalizado. O lote terminou `FAILED / RAW_ONLY / BLING_API_NORMALIZATION_FAILED`, com rollback; portanto, nenhum número novo foi publicado nos cards e a tela permaneceu sem resultados.
- Essa falha não foi a falta de permissão de categoria: com o código executado, esse caso teria sido classificado como `BLING_CATEGORY_SCOPE_INSUFFICIENT` e publicado como `PARTIAL`. A causa desta execução permanece desconhecida porque os erros internos não podem ser gravados em texto bruto.
- A classificação segura foi ampliada para distinguir ausência/rejeição de detalhe e falhas de persistência PostgreSQL sem guardar conteúdo de respostas ou registros. Os testes unitários e `go vet` focais passaram. Uma nova execução com esta versão ainda é necessária para confirmar a categoria exata da falha.
- A carga real continua incompleta e a T02 não está homologada. `/caixas` segue sem normalização, saldos conciliados e série projetada; não declarei resultados reais na tela.

## Correção do identificador de categoria e repetição real — 02/11/2026

- A execução das 17:47 terminou com o código sanitizado `BLING_PAYABLE_CATEGORY_ID_INVALID`. Isso localizou a falha no identificador da categoria dentro do detalhe do título; nenhum valor, nome ou payload foi registrado para diagnóstico.
- Consultei o contrato OpenAPI oficial do Bling, publicado na [referência da API](https://developer.bling.com.br/referencia): a categoria de uma conta a pagar contém `id` inteiro obrigatório, mas o contrato não define um valor mínimo. O parser agora rejeita ID ausente ou negativo, e trata ID zero como categoria não atribuída: conserva valor, saldo, vencimento, fornecedor e documento validados, deixa categoria sem classificação e fecha o lote como `PARTIAL`. Nenhuma classificação contábil é inferida.
- Se a categoria tiver ID positivo e a consulta de seu detalhe retornar 403, o comportamento parcial aprovado continua o mesmo. Falhas em campos financeiros, fornecedor, detalhes do título, categoria positiva inválida ou paginação continuam revertendo a publicação.
- A referência oficial de [limites da API](https://developer.bling.com.br/limites) documenta o teto global de três requisições por segundo. Como a lista de contas a pagar não traz saldo em aberto e esse valor é obrigatório para a tela, a normalização consulta o detalhe de cada título respeitando o limite; uma atualização completa pode levar mais de uma hora. Nenhum limite ou atalho que altere a cobertura foi introduzido.
- Verificações locais: testes Go focais e `go vet` passaram; o teste de integração `TestDailyPayablesAPIIsNormalizedAtomicallyWithCategoryScopeGap` passou no banco PostgreSQL separado `majucau_test`; `frontend/src/App.test.tsx` passou com 32 testes, incluindo valor parcial e categoria em branco. O teste de integração confirma transação, estado parcial e preservação dos dados financeiros; não representa dados reais do Bling.
- Repetição com a conta real iniciada às 17:53 BRT: Contas a Receber concluiu `SUCCESS / RAW_ONLY` com 14.888 registros. Contas a Pagar terminou `FAILED / RAW_ONLY` com 10.198 registros lidos e rollback; nenhum dado normalizado desse lote foi publicado.

## Diagnóstico do saldo negativo em Contas a Pagar — 02/11/2026

- Com o diagnóstico sanitizado atualizado, a repetição real das 18:03 BRT concluiu Contas a Receber como `SUCCESS / RAW_ONLY` com 14.891 registros e Contas a Pagar como `FAILED / RAW_ONLY / BLING_PAYABLE_OPEN_BALANCE_NEGATIVE` com 10.198 registros lidos. A sincronização de Contas a Pagar foi revertida integralmente, sem substituir o último retrato válido.
- O código `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE` se refere ao campo `saldo` de um título individual de Contas a Pagar. Não significa que o saldo total de banco/caixa esteja negativo; esse total pode estar positivo ao mesmo tempo. O log deliberadamente não guarda valor nem identificador do título, portanto ainda falta comparar o item específico com a tela do Bling.
- Consultei o contrato OpenAPI atual na [referência oficial da API Bling](https://developer.bling.com.br/referencia). `ContasPagarDadosDTO.saldo` é um número descrito como o valor da conta subtraído dos valores de pagamentos; o contrato não declara valor mínimo. Portanto, o formato negativo existe na resposta observada, mas a documentação não define se deve ser tratado como crédito de fornecedor, divergência a corrigir ou outra situação contábil.
- O Majucau e a restrição do banco atualmente exigem saldo em aberto não negativo. Não alterei esse tratamento: permitir o sinal negativo ou convertê-lo em zero mudaria o resultado financeiro. A T04 segue sem nova publicação e não está homologada até uma regra aprovada e uma sincronização válida.
- Diagnósticos passaram a registrar somente códigos estáticos de formato (`MISSING`, `TYPE_INVALID`, `NEGATIVE`, `PRECISION_EXCEEDED`, entre outros). O saldo real e qualquer payload continuam fora de logs, testes versionados e documentação.
- A janela do protótipo permanece ativa. A inspeção da tela T04 não foi concluída nesta execução, então não há evidência visual de publicação no painel.

## Verificação integral local — 02/11/2026, 18:18–18:20 BRT

- `scripts/verify.ps1` terminou com sucesso: testes Go, limites entre camadas, `go vet`, `govulncheck`, checksums de migrations, geração sqlc determinística, scan local de secrets, smoke do instalador, lint, TypeScript, 62 testes frontend, build e `npm audit` (zero vulnerabilidades npm).
- O banco de teste não foi configurado (`MAJUCAU_TEST_DATABASE_URL` ausente); portanto, os testes de integração PostgreSQL foram compilados, mas pulados nesta execução local. A validação PostgreSQL precisa ocorrer na CI acionada pelo push.
- `govulncheck` encontrou zero vulnerabilidades alcançáveis; relatou uma vulnerabilidade em módulo requerido sem caminho de chamada identificado pelo scanner.
- A leitura real de Contas a Pagar continua recusando a publicação por `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`; o comportamento financeiro não foi alterado. A regra aprovada para esse caso ainda é necessária antes de homologar T04.
- O protótipo/worker e a instância PostgreSQL local permanecem em uso. Os caches centralizados e `frontend/dist` foram preservados; conferir `artifacts/retained/cleanup-register.md` antes de qualquer limpeza.
