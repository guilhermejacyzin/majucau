# Confirmação de estado do serviço Windows após consulta lenta — 2026-11-02

## Falha observada

Na CI `41355117965`, o instalador retornou erro ao aguardar o worker. O
diagnóstico colhido logo em seguida mostrou o serviço como `Running`, saudável,
com código de saída 0. A falha ocorria porque o código rejeitava o estado
correto quando a chamada de consulta ao Windows terminava após o prazo de
espera.

## Correção

A espera agora termina com sucesso quando o Windows confirma o estado esperado,
mesmo se a consulta que trouxe essa confirmação demorou. Se o estado esperado
não for confirmado até o prazo, o timeout continua falhando como antes. Um
teste de regressão reproduz uma consulta lenta que retorna `Running`.

## Verificação

- `go test ./internal/installer` passou localmente em 02/11/2026.
- `git diff --check` passou.
- A CI anterior passou no backlog e PostgreSQL, mas a verificação Windows
  falhou no smoke do serviço. A correção deve ser confirmada pela nova CI.

Esta correção não altera telas, regras financeiras, permissões ou dados reais.
