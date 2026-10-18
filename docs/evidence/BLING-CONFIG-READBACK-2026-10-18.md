# Recuperação da configuração pública do Bling — 18/10/2026

## Resultado

A tela Integrações agora consulta o worker ao abrir o cartão do Bling e recupera o Client ID, o endereço de retorno e um indicador booleano de que existe Client Secret válido no cofre do Windows. Isso evita deixar os campos públicos vazios quando a configuração já foi salva.

O segredo e os tokens permanecem no worker/cofre: a consulta não os devolve à tela. O formulário continua sem preencher o campo Client Secret. Nenhuma posição, dimensão, texto visual ou regra financeira foi alterada.

## Limites

- Esta correção recupera configuração já salva; não inicia autorização OAuth nem sincronização.
- O teste não acessou a conta real do Bling nem leu dados financeiros reais.
- A presença do segredo no cofre não comprova que ele está correto ou que o OAuth funciona.
- A paginação real e a transformação dos dados RAW para os relatórios permanecem pendentes.
- A integração de ponta a ponta com o Bling continua aguardando homologação real.

## Verificações

- `go test -count=1 ./internal/integrations/bling ./internal/ipc ./cmd/worker .` — aprovado.
- `npm test` — 30 testes aprovados.
- `npm run typecheck` — aprovado.
- `npm run lint` — aprovado.
- `npm run build` — aprovado.
- `git diff --check` — aprovado.

Na primeira execução dos testes do frontend, um teste encontrou que a resposta do carregamento inicial sobrescrevia os campos logo após salvar. A atualização foi corrigida e a suíte completa passou na segunda execução.
