# Provisionamento protegido da identidade do Drive — 2026-10-19

## O que foi preparado

- O worker aceita uma chave JSON de conta de serviço do Google e os três IDs autorizados em uma única configuração. A credencial e os IDs são gravados juntos em um único arquivo DPAPI do Windows; uma configuração incompleta não substitui a anterior.
- A chave é validada antes da gravação. O acesso OAuth usa somente `drive.readonly`; os três IDs precisam ser válidos, diferentes entre si e corresponder às estruturas já aprovadas.
- O envio usa o Named Pipe local autenticado. O worker obtém o SID do processo cliente pelo próprio Windows, sem confiar em um SID enviado no texto da solicitação.
- O worker grava no PostgreSQL um evento de auditoria com o SID, a correlação da solicitação e um resumo sem chave privada nem IDs das pastas. A permissão de escrita é concedida somente a uma função SQL restrita; o runtime não recebe permissão direta de gravação em `audit_events`.
- Se o registro de auditoria falhar, o worker tenta restaurar a credencial anterior (ou retirar a nova). Nesse caso, a operação retorna erro e não informa sucesso.
- O comando auxiliar recebe o caminho local da chave e um arquivo JSON separado, pequeno e limitado, com os IDs das pastas. Os IDs não aparecem nos argumentos do processo. A chave é lida com limite de 1 MiB, enviada somente pelo pipe local, nunca impressa e removida dos buffers em memória. A remoção do arquivo de origem é opcional e, se solicitada, significa apenas remover o arquivo; não é uma promessa de apagamento seguro do disco.
- O agendamento diário permanece ativo antes do provisionamento e relê a configuração protegida às 09:30 de Brasília, sem exigir reiniciar o worker.

## Segurança e limites

- Não foram alteradas telas aprovadas nem regras financeiras.
- A conta humana confirmada no Chrome é `financeiro@majucau.com.br`. Ela é usada apenas para consultar o Drive; não é a identidade técnica que o worker usa.
- Em 19/10, a consulta do Drive pelo Chrome foi concluída em modo somente leitura. A pasta `MAJUCAU_INTELLIGENCE_INPUT` contém `01_perdas_operacionais`, `02_nuvem_pago/Recebimentos_futuros` e `03_nuvem_envio/Frete`, conforme a estrutura aprovada. A listagem mostrou arquivos Excel (`.xlsx`) e CSV; o arquivo mais recente visível nessas pastas tinha data de modificação 11/09/2026. Foram conferidos nomes de pastas, tipos e metadados; nenhum arquivo de dados foi aberto, baixado ou modificado. IDs reais de pasta não foram gravados no Git.
- Conferência adicional em 19/10/2026 no perfil Chrome `majucau.com.br`, conectado como `financeiro@majucau.com.br`: `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT/01_perdas_operacionais/Planilha sem título.xlsx`; `02_nuvem_pago/Recebimentos_futuros/extrato_lancamentos_futuros_11_09_2026.csv`; e `03_nuvem_envio/Frete/relatorio_nuvem_envio_2026_09.csv`. A interface do Drive mostra a pasta `Recebimentos_futuros` com “R” maiúsculo; a chave interna padronizada em minúsculas continua sendo apenas um identificador técnico, não o nome exibido da pasta. Os três arquivos apareciam modificados em 11/09/2026. Foram consultados apenas nomes, tipos, caminhos e datas; nenhum conteúdo financeiro foi aberto, baixado ou copiado.
- Nenhuma chave de conta de serviço foi recebida ou gravada. Nenhum segredo, token, cookie, senha ou ID real de pasta foi colocado no Git. A tentativa de abrir o Google Cloud Console solicitou uma nova confirmação de identidade da conta financeira; ela não foi concluída, então a existência de um projeto técnico ainda é desconhecida.
- A identidade técnica ainda precisa ser criada ou localizada no Google Cloud, receber permissão somente de leitura nas três raízes autorizadas e ser configurada localmente. O inventário real feito pelo serviço ainda não foi executado.
- O mapa dos três IDs confirmados no Drive foi preparado localmente em `artifacts/retained/google-drive-folder-ids.json`, fora do Git e registrado em `artifacts/retained/cleanup-register.md`. O arquivo contém apenas os IDs das pastas, não tem chave nem conteúdo financeiro, e será removido depois que a configuração for gravada com proteção DPAPI no worker.
- A migração `000011_google_drive_credential_audit.up.sql` adiciona a função de auditoria mínima. Ela foi aplicada e suas permissões foram verificadas no PostgreSQL da CI #680.

## Verificação desta alteração

- Suítes focadas de Google Drive, worker e IPC passaram em Windows em 19/10/2026.
- `scripts/verify.ps1` passou localmente em 19/10/2026: testes Go, `go vet`, `govulncheck` sem chamadas vulneráveis (permanece 1 aviso de módulo não chamado, já registrado), checksums das migrations, geração SQLC sem diferenças, scan de secrets, teste do instalador, lint, TypeScript, 33 testes de interface, build e `npm audit` com zero vulnerabilidades.
- A execução GitHub Actions #678 aplicou a migration e aprovou `database/privileges/verify_runtime.sql`, mas revelou que a comparação textual do caminho retornado pelo Windows não era portátil entre ambientes.
- A primeira normalização ainda falhou na execução #679. A correção final valida que o caminho resolvido continua na unidade local esperada e confere por identidade do arquivo que a leitura abre o mesmo arquivo regular observado inicialmente; UNC, dispositivos e links no próprio arquivo continuam recusados.
- `scripts/build-windows.ps1` passou localmente após a correção final, incluindo todos os testes Go, `go vet`, `govulncheck`, integridade das migrations, SQLC, scan de segredos, verificações do instalador, lint, TypeScript, 33 testes da interface, build e `npm audit` sem vulnerabilidades. O pacote portátil foi recompilado e conferido: manifesto e hashes correspondem, e o bundle inclui o worker e a migration 000011.
- O smoke test do worker sem argumentos retornou `GOOGLE_DRIVE_CREDENTIAL_ARGUMENTS_INVALID` e código 2, conforme esperado; nenhuma credencial ou sincronização foi usada. A execução #680 passou nos quatro jobs: backlog, detecção de mudanças, PostgreSQL e Windows; incluiu a migration e as permissões, testes da correção Windows e geração do pacote.

## Próximo passo

A CI está concluída. O commit `e4a740c` passou na CI #39215678392. Próximo passo: após a confirmação de identidade no Google Cloud Console, verificar se já existe um projeto/conta técnica do Majucau e, somente depois, provisionar a leitura somente leitura e executar o primeiro inventário real, mantendo arquivos sem contrato financeiro em estado pendente. A listagem visível no Drive está modificada pela última vez em 11/09/2026; confirmar atualidade dos arquivos faz parte da homologação.
