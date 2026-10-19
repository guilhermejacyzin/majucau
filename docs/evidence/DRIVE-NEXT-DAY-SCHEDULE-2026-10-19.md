# Regra de horário para arquivos atrasados do Drive — 19/10/2026

## Regra aprovada

O Drive é consultado diariamente às 09:30 no fuso `America/Sao_Paulo`. Se um arquivo não estiver disponível nesse horário ou chegar depois, permanece pendente e só entra na próxima execução das 09:30 do dia seguinte. O serviço não deve iniciar uma coleta do Drive fora desse horário por ter iniciado ou reiniciado depois das 09:30.

## Correção

O worker fazia uma recuperação imediata no início do processo sempre que iniciava depois das 09:30. Isso podia importar no mesmo dia um arquivo que deveria aguardar a próxima execução programada. Removi esse caminho de recuperação imediata para o Drive. O worker agora aguarda o próximo horário futuro na agenda diária; o agendamento Bling permanece independente.

Não alterei os formatos, contratos, regras financeiras, dados, telas ou permissões. A coleta continua sujeita à configuração do Drive, às pastas autorizadas e aos contratos pendentes.

## Verificações

- Teste do worker confirma que uma inicialização às 11:15 de Brasília calcula a próxima execução para 09:30 do dia seguinte.
- Os testes da agenda diária continuam cobrindo a borda de 09:30 e a execução no fuso aprovado.
- `go test ./cmd/worker ./internal/scheduler -count=1` — aprovado.
- `scripts/verify.ps1` — aprovado integralmente: Go, `go vet`, `govulncheck`, migrations, SQLC determinístico, scan de secrets, verificações do instalador, lint, TypeScript, 33 testes frontend, build e `npm audit` sem vulnerabilidades.
- A CI do GitHub será registrada depois do envio desta correção; ela ainda não foi confirmada.
- BK-039 permanece em 50% (50% restante): a correção fecha um desvio de horário, mas ainda faltam integração real do Drive, leitura persistente, operação contínua e homologação. A média do projeto permanece em 11,8% concluído (88,2% restante).

## Limpeza

Com `scripts/SafeCleanup.psm1`, foram removidos `artifacts/cache/go-build/`, `go-tmp/`, `npm/`, `vite/` e a prévia `artifacts/verification/frontend-dist-*` criada pela verificação. Permanecem `artifacts/cache/go-mod/`, ferramentas, dependências instaladas e pacotes oficiais em `artifacts/windows/`. Nenhum cache global foi apagado; nenhum item novo exige exclusão manual. O registro local está em `artifacts/retained/cleanup-register.md`.
