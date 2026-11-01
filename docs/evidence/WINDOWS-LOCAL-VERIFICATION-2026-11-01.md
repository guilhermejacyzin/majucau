# Correções do verificador local no Windows — 01/11/2026

## Correções

- O verificador de limites arquiteturais montava `go list -f` com aspas embutidas no template. O Windows PowerShell 5.1 dividia o argumento e o Go recebia um template incompleto. O script agora lê `.Imports` sem aspas internas e separa a lista retornada.
- O teste do instalador NSIS comparava mensagens em português, mas o arquivo PowerShell UTF-8 sem BOM era interpretado como ANSI pelo Windows PowerShell 5.1. Salvei esse script com BOM UTF-8 para que os textos sejam lidos corretamente.
- `verify.ps1` agora só move `frontend/dist` quando essa execução o criou e tenta a movimentação até três vezes. Isso evita mover uma saída anterior do usuário.

## Verificação

`powershell -NoProfile -ExecutionPolicy Bypass -File scripts/verify.ps1` terminou com sucesso quando `frontend/dist` já existia. Passaram:

- `go test` nos pacotes do projeto;
- limites de arquitetura Go, `go vet` e `govulncheck` (nenhuma vulnerabilidade alcançável; uma dependência declarada foi reportada sem uso vulnerável pelo código);
- checksums das migrations, geração SQLC determinística, scan local de secrets e contrato do instalador NSIS;
- lint, TypeScript, 61 testes de interface, build e `npm audit` (0 vulnerabilidades).

Também executei a mesma verificação sem `frontend/dist`: todas as verificações passaram, mas o Windows recusou mover a pasta gerada mesmo após três tentativas. Não interrompi os processos locais ativos para liberar o caminho. O bundle permanece ignorado pelo Git em `frontend/dist` e está registrado para limpeza manual em `artifacts/retained/manual-cleanup/CLEANUP-2026-11-01.md`. Os caches do Go, npm e Vite ficam em `artifacts/cache/`.

Nenhuma tela aprovada ou regra financeira foi alterada. A autenticação/sincronização real do Bling continua dependendo da permissão de categorias na conta e da autorização OAuth completa; esta correção é somente do roteiro de verificação local.

BK-025 permanece estimada em 70% (30% restante); a estimativa do projeto continua em 11,9% concluído e 88,1% restante. Esta manutenção de verificação não conclui uma tarefa funcional do produto.
