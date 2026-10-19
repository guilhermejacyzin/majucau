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
- A execução remota inicial chegou a falhar apenas na instalação do NSIS: a consulta ao repositório Chocolatey retornou HTTP 503. A repetição confirmou o mesmo erro externo.
- A primeira alternativa, baixar diretamente do SourceForge, foi recusada porque o conteúdo recebido não correspondeu ao hash esperado.
- A CI agora baixa o arquivo do espelho de releases do Tauri, verifica o SHA-256 canônico publicado pelo projeto NSIS antes de extrair e executar o compilador. O download foi conferido localmente: tamanho 2.363.925 bytes e hash esperado.
- A execução remota que valida esse novo caminho ainda precisa concluir antes de declarar a CI verde.

## Limites restantes

- O instalador ainda não registra nem inicia o serviço.
- O instalador ainda precisa criar a pasta pai com ACL para o SID `NT SERVICE\\MajucauWorker`, SYSTEM e Administradores, provisionar a credencial do banco sem texto aberto em configuração permanente e definir a dependência real do PostgreSQL.
- Não houve instalação, alteração no Windows, configuração de identidade Google, acesso técnico às pastas, sincronização real ou importação financeira.
- BK-039 permanece em 50%; esta alteração prepara o modo seguro do cofre, mas não prova que o worker execute diariamente após fechar o aplicativo ou reiniciar o Windows.
- A confirmação somente leitura no Drive encontrou `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com `01_perdas_operacionais`, `02_nuvem_pago` e `03_nuvem_envio`. Nenhum conteúdo de arquivo foi incluído nesta evidência.

Nenhuma tela aprovada ou regra financeira foi alterada.

## Proveniência da ferramenta de empacotamento

- Arquivo oficial NSIS 3.13: https://sourceforge.net/projects/nsis/files/NSIS%203/3.13/
- Catálogo canônico de releases e hashes do projeto NSIS: https://github.com/nsis-dev/release-data
- Espelho binário usado pela CI: https://github.com/tauri-apps/binary-releases/releases/tag/nsis-3.13
- SHA-256 fixado na CI: `ba63dffc4410ee89193e1cb5a41989991bd77c61068da17e3156d136b7b0b3d8`
- O hash do espelho conferiu com o hash do arquivo NSIS publicado no catálogo canônico. O instalador continua sendo produzido por NSIS; esta correção altera apenas como a ferramenta de build é obtida e verificada na CI.
