# Investigação do smoke Windows do serviço worker — 31/10/2026

## Fato observado

- O commit `11f4205` foi validado na execução de CI
  [41262694158](https://github.com/guilhermejacyzin/majucau/actions/runs/41262694158).
- `detect-changes`, `verify-backlog` e `verify-postgres` passaram.
- `verify-windows` falhou somente em `Smoke Windows worker service lifecycle`,
  com o estágio `WORKER_SERVICE_WAIT_FOR_RUNNING`.
- O diagnóstico posterior do mesmo smoke encontrou o serviço como `Running`,
  `Status=OK`, código de saída zero e evento de instalação 7045. O relatório não
  contém o erro Win32 da consulta que falhou; portanto, a causa transitória exata
  não pôde ser provada pelos logs disponíveis.
- O smoke equivalente passou na execução anterior
  [41261894619](https://github.com/guilhermejacyzin/majucau/actions/runs/41261894619).
- A tentativa de reexecutar somente o job falho foi recusada pela integração do
  GitHub com HTTP 403 (`Resource not accessible by integration`).

## Alteração aplicada

A consulta do estado do serviço pelo Service Control Manager é somente leitura.
Agora ela é repetida em intervalos de 200 ms até o estado solicitado ser
observado ou o limite já aprovado de 45 segundos expirar. Uma resposta com erro
nunca é interpretada como sucesso. Se as consultas continuarem falhando, a
instalação continua falhando e o último erro fica encadeado para diagnóstico.

Foram adicionados testes para (1) duas falhas de consulta seguidas de observação
de `Running` e (2) erro persistente até o timeout. Nenhum controle de ACL,
identidade do serviço, tela ou regra financeira foi alterado.

## Validação e limite atual

- `scripts/verify.ps1`: passou localmente, incluindo Go, `go vet`, análise de
  vulnerabilidades e segredos, validações do instalador e frontend (39 testes,
  lint, typecheck, build e `npm audit`).
- O banco PostgreSQL dedicado de integração não foi configurado nesta máquina;
  por isso a validação local não substitui a etapa PostgreSQL da CI.
- O smoke que instala serviço Windows não foi executado localmente: há um
  processo de worker do protótipo em execução e o terminal não está elevado. A
  instalação local poderia interferir com o protótipo. A CI em runner limpo é
  o próximo teste desse ciclo.
- A causa nativa da falha anterior permanece não comprovada. A mudança melhora
  a tolerância a falhas temporárias sem esconder erros persistentes; a conclusão
  depende do resultado do próximo job Windows da CI.

## Próximo passo

Enviar esta correção para o branch de trabalho e verificar a nova CI. Não fechar
esta investigação nem avançar para outra fatia enquanto o job Windows não
passar ou enquanto uma nova falha não for investigada e corrigida.
