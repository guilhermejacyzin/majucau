# Ciclo de vida do serviço Windows do worker — 19/10/2026

## Objetivo

Garantir que o worker possa iniciar com o Windows, continuar executando as rotinas agendadas após fechar o aplicativo e ser removido sem apagar os dados protegidos. Esta etapa não muda telas, contratos financeiros nem regras de negócio.

## O que foi implementado

- O instalador mantém uma cópia versionada do worker dentro da pasta do Majucau. O auxiliar registra e inicia `MajucauWorker` como a conta virtual `NT SERVICE\MajucauWorker`, com SID de serviço restrito, início automático atrasado e recuperação por reinício após falha.
- Antes de iniciar o serviço, o auxiliar valida os caminhos, rejeita argumentos no caminho do executável e caminhos com reparse points, cria as pastas de cofre e logs com DACL protegida somente para o SID do serviço, SYSTEM e Administradores, e configura/valida a leitura restrita do SID da interface no registro.
- O worker executado pelo Service Control Manager descobre o próprio SID e lê o SID da interface do registro protegido. Não depende de variável de ambiente gravada no registro. O cofre continua usando DPAPI de máquina com ACL explícita; o modo de console não muda.
- O instalador chama a operação de serviço somente depois de copiar os arquivos e gravar o desinstalador. Falha na instalação interrompe o fluxo sem afirmar que a atualização automática está ativa. A desinstalação remove o serviço antes dos binários e preserva os diretórios de dados.
- O auxiliar serializa instalação e remoção com a trava existente e tenta restaurar a configuração anterior em caso de falha no upgrade.
- O CI ganhou um smoke de ciclo completo: registro, início, conta/SID, inicialização automática, ACLs de pastas e registro, diretórios DPAPI/log, remoção do serviço e preservação de dados. O smoke só limpa suas próprias pastas quando a lista de arquivos corresponde ao conteúdo conhecido; resíduos inesperados são preservados e registrados em `RETAINED-CLEANUP.json`.

## Verificações feitas nesta máquina

- `go test -count=1 ./cmd/installer-helper ./cmd/worker ./internal/diagnosticslog ./internal/installer`: passou.
- `scripts/test-installer-nsis.ps1`: passou.
- Análise sintática PowerShell dos dois smokes e do módulo de limpeza: passou.
- `scripts/verify.ps1`: passou por completo, incluindo Go, `go vet`, `govulncheck`, migrations, SQLC, scan de secrets, verificações do instalador, lint, typecheck, 33 testes de interface, build e `npm audit` sem vulnerabilidades.
- `git diff --check`: passou.

## O que ainda não foi comprovado

- O smoke privilegiado de registro/início/remoção do serviço ainda precisa passar no CI Windows. Ele foi acrescentado ao workflow, mas não foi executado nesta máquina para não registrar serviço nem criar dados em `ProgramData\Majucau` do computador da usuária.
- A instalação e reinicialização devem ser verificadas em uma VM limpa Windows 10/11 antes da homologação do protótipo.
- O serviço ainda não recebeu configuração segura do PostgreSQL, a identidade/credencial técnica do Google Drive ainda não foi provisionada, e as conexões reais do Drive e do Bling continuam pendentes. A paginação do Bling precisa de homologação real.
- Portanto, o agendamento de 09:30 já está preparado no worker, mas ainda não pode ser considerado operacional de ponta a ponta. A execução real e os dados financeiros não foram confirmados.

## Progresso

BK-039 permanece em 50% (50% restante) até o smoke Windows, VM limpa, configuração segura das fontes e homologação. A média simples registrada do backlog permanece em 11,8% concluída e 88,2% restante (127 tarefas; 7 concluídas, 20 em andamento, 100 não iniciadas). Esta alteração não muda os percentuais. Nenhuma tela aprovada ou regra financeira foi alterada.

## Próxima etapa

Enviar a mudança ao GitHub e aguardar a execução do smoke Windows. Se ele passar, continuar a mesma BK-039 com a validação de instalação em VM limpa e depois conectar, em modo seguro, as fontes autorizadas para homologação real.
