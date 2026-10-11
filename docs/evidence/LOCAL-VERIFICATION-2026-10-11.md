# Verificação local completa — 2026-10-11

## Escopo

Execução de `scripts/verify.ps1` no Windows, usando PowerShell 7.6.5, no commit `be9cbae`. O roteiro verificou testes Go, `go vet`, `govulncheck`, checksums das migrações, geração determinística do sqlc, scan de segredos, contrato do instalador NSIS, lint, typecheck, testes, build e auditoria npm do frontend.

## Resultado

- Go: `go test . ./cmd/... ./database/... ./internal/...` passou em todos os pacotes; `go vet` passou.
- Segurança Go: `govulncheck` não encontrou vulnerabilidade alcançável pelo código.
- Dependências Go: `GO-2026-5932` foi reportada no módulo `golang.org/x/crypto@v0.56.0`, para o pacote `openpgp` sem manutenção; não há versão corrigida indicada. O projeto importa `golang.org/x/crypto/argon2` para derivação de chave e não importa `openpgp`. Não há caminho de chamada vulnerável. Manter como item monitorado; remover ou trocar a dependência exige avaliar a criptografia usada no backup.
- Banco/instalador: checksum das migrações, geração sqlc sem diff, scan local de segredos e contrato NSIS passaram.
- Frontend: lint, typecheck, 24 testes em 3 arquivos, build de produção e `npm audit` sem vulnerabilidades passaram.
- O primeiro ensaio foi iniciado por `powershell.exe` 5.1 e falhou na leitura de um texto acentuado. A CI do projeto usa PowerShell 7 (`pwsh`); no runtime correspondente o contrato NSIS passou. A verificação completa foi repetida em PowerShell 7 e terminou com sucesso.

## Limites

Esta é validação local do roteiro, não uma execução de CI do GitHub. Não valida ACL/limpeza em VM Windows 10/11 limpa, PostgreSQL ponta a ponta real, assinatura de release, aceite funcional, acesso à conta Bling nem os gates G2–G8. Nenhuma tela ou regra financeira foi alterada.

O roteiro gerou caches, ferramentas locais e dois bundles frontend em `artifacts/`; esses itens foram removidos ao final da execução. A edição local preexistente de `scripts/smoke-installer.ps1` foi preservada e não entrou em commit.
