# Escopo DPAPI separado para protótipo e serviço — 19/10/2026

## Correção

- O protótipo iniciado em modo console continua usando DPAPI no escopo do usuário atual.
- Quando o worker é iniciado pelo Service Control Manager, o contexto de execução seleciona DPAPI no escopo da máquina (`CRYPTPROTECT_LOCAL_MACHINE`).
- O cofre mantém DACL protegida e validada para a identidade atual do processo, SYSTEM e Administradores. DPAPI no escopo da máquina não substitui a ACL.
- Um marcador privado `.scope` impede que o mesmo diretório seja aberto com escopos incompatíveis. Cofres antigos sem marcador só podem ser reconhecidos como escopo do usuário atual, mantendo o acesso aos secrets anteriores; o construtor de escopo da máquina recusa diretórios antigos ou já preenchidos.

## Verificações

- `go test -count=1 ./internal/security ./cmd/worker`: passou no Windows local, incluindo teste do DPAPI no escopo da máquina, round-trip, marcador, ACL, preservação do cofre antigo e escolha explícita do modo de serviço.
- `scripts/verify.ps1`: passou por completo no Windows local, incluindo todos os testes Go, `go vet`, limites arquiteturais, `govulncheck`, checksum de migrations, geração SQLC determinística, scan de secrets, verificação do contrato do instalador, ESLint, TypeScript, 33 testes de interface, build e `npm audit` sem vulnerabilidades.
- `govulncheck` não encontrou vulnerabilidade alcançável; encontrou uma vulnerabilidade em módulo requerido sem chamada aparente, acompanhada em `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- `scripts/verify-backlog.ps1` e `git diff --check`: passaram.
- A CI completa do commit publicado ainda precisa concluir antes de declarar o avanço remoto validado.

## Limites restantes

- O instalador ainda não registra nem inicia o serviço.
- O instalador ainda precisa criar a pasta pai com ACL para o SID `NT SERVICE\\MajucauWorker`, SYSTEM e Administradores, provisionar a credencial do banco sem texto aberto em configuração permanente e definir a dependência real do PostgreSQL.
- Não houve instalação, alteração no Windows, configuração de identidade Google, acesso técnico às pastas, sincronização real ou importação financeira.
- BK-039 permanece em 50%; esta alteração prepara o modo seguro do cofre, mas não prova que o worker execute diariamente após fechar o aplicativo ou reiniciar o Windows.
- A confirmação somente leitura no Drive encontrou `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com `01_perdas_operacionais`, `02_nuvem_pago` e `03_nuvem_envio`. Nenhum conteúdo de arquivo foi incluído nesta evidência.

Nenhuma tela aprovada ou regra financeira foi alterada.
