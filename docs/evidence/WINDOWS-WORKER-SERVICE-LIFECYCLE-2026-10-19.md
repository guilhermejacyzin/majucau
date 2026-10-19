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
- A primeira execução remota, [CI #39231983131](https://github.com/guilhermejacyzin/majucau/actions/runs/39231983131), passou backlog, verificação da mensagem e PostgreSQL, mas falhou no registro/início do serviço Windows. O retorno anterior do auxiliar e do smoke ocultava a fase e o motivo detalhado; o job não confirmou a instalação.
- Para a próxima execução, o auxiliar agora diferencia erros de configuração, transição do serviço e trava. O smoke registra, em caso de falha, o estado e os códigos do serviço e eventos recentes do Service Control Manager, sem incluir credenciais. O novo teste de contrato e o verificador completo local passaram.
- A segunda execução, [CI #39233455017](https://github.com/guilhermejacyzin/majucau/actions/runs/39233455017), ainda falhou antes de criar um serviço persistente, agora identificada como `WORKER_SERVICE_CONFIG_INVALID`; a limpeza automática removeu a tentativa incompleta e os outros três trabalhos passaram. O Windows não produziu eventos úteis para restringir a causa. Esta atualização acrescenta códigos estáveis por etapa da instalação, sem expor caminhos ou erros nativos.
- Depois de acrescentar os códigos por etapa, `scripts/verify.ps1`, `scripts/verify-backlog.ps1`, os testes focados do auxiliar/instalador e `git diff --check` passaram nesta máquina. O próximo CI ainda precisa confirmar qual etapa falha.
- A terceira execução, [CI #39234596899](https://github.com/guilhermejacyzin/majucau/actions/runs/39234596899), apontou `WORKER_SERVICE_PATH_VALIDATION`. A causa era a validação reaproveitar o analisador de linha de comando do Service Control Manager para um caminho de arquivo vindo do instalador; como o teste usa `C:\Program Files`, o espaço era incorretamente recusado. Esta atualização valida caminhos de arquivo separadamente e grava o executável entre aspas tanto na criação quanto na atualização do serviço.
- A quarta execução, [CI #39235534818](https://github.com/guilhermejacyzin/majucau/actions/runs/39235534818), passou pela validação de caminho, mas falhou em `WORKER_SERVICE_VALIDATE_EXISTING_SERVICE`. O evento 7045 mostrou aspas duplicadas: a biblioteca `mgr.CreateService` já escapa/coloca aspas no caminho por conta própria. Esta correção passa o caminho bruto apenas a `CreateService` e mantém a citação explícita para `UpdateConfig`, onde a biblioteca não o faz.
- A quinta execução, [CI #39236674784](https://github.com/guilhermejacyzin/majucau/actions/runs/39236674784), confirmou que o serviço registrou e iniciou, mas o smoke esperava uma subpasta `Vault` que o worker só cria quando há conexão de banco configurada. O CI de ciclo de serviço não configura um banco, como esperado. A instalação agora prepara também a pasta vazia do cofre com a mesma ACL restrita, sem criar nem gravar credenciais; o smoke verifica a ACL dessa pasta separadamente.
- A sexta execução, [CI #39237629665](https://github.com/guilhermejacyzin/majucau/actions/runs/39237629665), avançou até a conferência da DACL do Registro e o serviço permaneceu registrado/iniciado. O script falhou porque o PowerShell do runner não disponibiliza o tipo `Microsoft.Win32.RegistryRights`. Esta atualização compara as máscaras `KEY_READ` e `KEY_ALL_ACCESS` diretamente, com os mesmos valores verificados pelo código Go.

## O que ainda não foi comprovado

- O smoke privilegiado de registro/início/remoção do serviço ainda precisa passar no CI Windows após corrigir a leitura das máscaras de permissão no PowerShell. A nova execução deve confirmar as ACLs, iniciar o worker, remover o serviço e preservar os dados. O smoke não foi executado nesta máquina para não registrar serviço nem criar dados em `ProgramData\Majucau` do computador da usuária.
- A instalação e reinicialização devem ser verificadas em uma VM limpa Windows 10/11 antes da homologação do protótipo.
- O serviço ainda não recebeu configuração segura do PostgreSQL, a identidade/credencial técnica do Google Drive ainda não foi provisionada, e as conexões reais do Drive e do Bling continuam pendentes. A paginação do Bling precisa de homologação real.
- Portanto, o agendamento de 09:30 já está preparado no worker, mas ainda não pode ser considerado operacional de ponta a ponta. A execução real e os dados financeiros não foram confirmados.

## Progresso

BK-039 permanece em 50% (50% restante) até o smoke Windows, VM limpa, configuração segura das fontes e homologação. A média simples registrada do backlog permanece em 11,8% concluída e 88,2% restante (127 tarefas; 7 concluídas, 20 em andamento, 100 não iniciadas). Esta alteração não muda os percentuais. Nenhuma tela aprovada ou regra financeira foi alterada.

## Próxima etapa

Enviar a comparação de máscaras numéricas e aguardar uma nova execução do smoke Windows. Se apontar outra falha, corrigir e repetir até ACLs, registro, início e remoção passarem. Depois, validar o pacote em uma VM limpa Windows 10/11. Isso ainda não conclui BK-039.
