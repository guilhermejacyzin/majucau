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
- Nenhuma chave de conta de serviço foi recebida ou gravada nesta execução. Nenhum segredo, token, cookie, senha, ID real de pasta ou arquivo real foi colocado no Git.
- A conta de serviço ainda precisa ser criada no Google Cloud, receber permissão somente de leitura nas três raízes autorizadas e ser configurada localmente. O primeiro inventário real ainda não foi executado.
- A migração `000011_google_drive_credential_audit.up.sql` adiciona a função de auditoria mínima. Ela precisa passar na verificação PostgreSQL local e na CI antes de ser considerada validada.

## Verificação desta alteração

- Suítes focadas de Google Drive, worker e IPC passaram em Windows em 19/10/2026.
- `scripts/verify.ps1` passou localmente em 19/10/2026: testes Go, `go vet`, `govulncheck` sem chamadas vulneráveis (permanece 1 aviso de módulo não chamado, já registrado), checksums das migrations, geração SQLC sem diferenças, scan de secrets, teste do instalador, lint, TypeScript, 33 testes de interface, build e `npm audit` com zero vulnerabilidades.
- A execução GitHub Actions #678 aplicou a migration e aprovou `database/privileges/verify_runtime.sql`, mas apontou incompatibilidade dos testes de leitura de arquivos no Windows: `filepath.EvalSymlinks` pode devolver um caminho local no formato estendido `\\?\C:\...`.
- A correção normaliza esse formato somente para caminhos de unidade local e continua recusando UNC e caminhos de dispositivo. Foi adicionado um teste de regressão para esses formatos.
- Após a correção, `scripts/build-windows.ps1` passou localmente, incluindo todos os testes Go, `go vet`, `govulncheck`, integridade das migrations, SQLC, scan de segredos, verificações do instalador, lint, TypeScript, 33 testes da interface, build e `npm audit` sem vulnerabilidades. O pacote portátil foi recompilado e conferido: manifesto e hashes correspondem, e o bundle inclui o worker e a migration 000011.
- O smoke test do worker sem argumentos retornou `GOOGLE_DRIVE_CREDENTIAL_ARGUMENTS_INVALID` e código 2, conforme esperado; nenhuma credencial ou sincronização foi usada. A CI do GitHub precisa confirmar a correção Windows antes de a alteração ser considerada concluída.

## Próximo passo

Concluir a CI da correção Windows. Depois, provisionar a identidade técnica somente leitura e executar a primeira leitura real sem baixar ou importar dados financeiros não homologados.
