# Auditoria estática de controle de acesso — 11/10/2026

## Objetivo e limites

Registrar o que o código atual comprova sobre perfis e autorização, sem aprovar uma matriz nova e sem mudar permissões, telas ou regras de negócio. Esta é uma leitura estática dos arquivos citados abaixo; nenhum teste foi executado nesta auditoria.

## O que existe no projeto

- A migration `database/migrations/000001_init.up.sql` cria `roles`, `permissions`, `role_permissions` e `system_users`.
- A mesma migration cadastra os perfis `ADMIN`, `DIRECTOR`, `FINANCE`, `ACCOUNTING` e `VIEWER`, o catálogo de 11 permissões e associações iniciais entre perfis e permissões.
- Os valores iniciais dão todas as permissões a `ADMIN`; `DIRECTOR` recebe leitura do painel, leitura/sincronização de integrações e exportação de backup; `FINANCE` recebe leitura do painel, leitura/sincronização de integrações e escrita financeira; `ACCOUNTING` recebe leitura do painel, leitura de integrações, escrita contábil e fechamento de período; `VIEWER` recebe leitura do painel e de integrações.
- Esses vínculos são valores padrão presentes na migration. Esta auditoria não os considera uma matriz de acesso aprovada pela responsável.

## Lacuna encontrada

- Em `cmd/worker/handler.go`, `workerHandler.Handle` direciona os métodos da IPC para os respectivos handlers. Nesse caminho de despacho não foi encontrada uma verificação do papel/permissão do usuário antes da ação.
- Em `internal/ipc/namedpipe_windows.go`, `authorizePipeClient` valida o SID do processo cliente contra a lista permitida do Named Pipe. Essa proteção identifica se o processo pode conversar com o worker; ela não consulta `system_users`, `role_permissions` ou a capacidade do perfil para cada ação.
- O código tem um fallback de desenvolvimento quando a lista de SIDs autorizados está vazia. A presença desse fallback não comprova se a configuração de produção o utiliza; isso precisa ser verificado na configuração operacional.

## Consequência prática

O projeto tem estruturas de banco para perfis e permissões e controle de quem pode abrir o Named Pipe, mas a leitura do código não comprovou autorização individual por perfil dentro das ações do worker. Portanto, `SEC-02` permanece `PARTIAL` e BK-013 segue em andamento. Isso é uma lacuna a resolver antes do gate de produção, não uma autorização para alterar agora os acessos cadastrados.

## Próximo passo seguro

Em 11/10/2026, a responsável aprovou para a versão 1: Financeiro usa `ADMIN`; Diretores podem consultar, sincronizar e exportar backups e relatórios; `ACCOUNTING` e `VIEWER` ficam fora da versão 1. Essa decisão resolve a escolha dos perfis e as capacidades dos Diretores. A implementação ainda precisa definir o provisionamento/identidade de usuários e verificar cada capacidade no worker, registrando ações permitidas e negadas. Até essa validação, a migration e o despacho atual não garantem que o acesso real corresponda à decisão aprovada; não alterar telas.

## Arquivos lidos

- `database/migrations/000001_init.up.sql`
- `cmd/worker/handler.go`
- `internal/ipc/namedpipe_windows.go`
- `REQUIREMENTS-TRACEABILITY.md`

## Reauditoria após aprovação da conta compartilhada — 11/10/2026

Gisele confirmou que os Diretores usarão uma conta compartilhada `contato@majucau.com.br`, após ser informada de que o histórico não identificará qual Diretor fez cada ação. A decisão está registrada no `G0-DECISION-REGISTER.md` e substitui a exigência de identidade individual da ADR-002 somente para o perfil `DIRECTOR`. Financeiro usa o perfil `ADMIN` com `financeiro@majucau.com.br`. Não houve mudança de tela nem regra financeira.

### Mapa observado dos comandos do worker

Esta classificação registra o que os comandos fazem e confronta com as capacidades amplas aprovadas. Não é uma implementação de autorização e não amplia permissões:

| Comando observado | Efeito observado | Relação com o perfil aprovado |
| --- | --- | --- |
| `health.get`, `integration.status`, `dashboard.snapshot` | Consulta saúde, integração ou dados do painel | Consulta; o worker ainda não valida perfil |
| `bling.oauth.*`, `bling.config.save`, `nuvemshop.config.save` | Configura, conecta, testa, revoga ou desconecta credenciais | Administração de integração; não está entre as capacidades aprovadas para Diretores |
| `bling.sync` | Consulta a API do Bling e grava dados sincronizados | Corresponde a “sincronizar integrações”, aprovado para Diretores |
| `bling.receipts.preview`, `nuvem_pago.future.preview` | Lê e valida arquivos CSV sem persistir seus registros | Prévia de consulta; autorização por perfil não existe hoje |
| `bling.receipts.import`, `nuvem_pago.future.import` | Grava registros CSV no banco local | Gisele aprovou a importação manual desses arquivos como parte de `INTEGRATION_SYNC`; o worker ainda não valida o perfil |
| Exportação de relatórios | Não há método de exportação no catálogo IPC atual; a tela mantém a opção indisponível sem dados confirmados | Capacidade aprovada, ainda sem rota de serviço comprovada nesta auditoria |

### Lacunas técnicas atualizadas

- A migration local identifica `system_users` por `windows_sid`; a auditoria não encontrou consumo dessa tabela no despacho do worker nem um contexto de identidade/login Majucau por comando. Um SID do Windows identifica uma conta de máquina/sistema local, não prova quem entrou na hospedagem central.
- O catálogo da migration possui `BACKUP_EXPORT`, mas não possui permissão específica `REPORT_EXPORT`, apesar de a exportação de relatórios ter sido aprovada para Diretores.
- O Named Pipe autentica o processo Windows autorizado; não transforma a conta compartilhada de e-mail em identidade autenticada no serviço central. Na futura auditoria, registrar a ação como executada pela principal compartilhada, sem atribuir um Diretor individual.
- Não foram localizadas verificações de permissão por comando no worker. A conta compartilhada aprovada, por si só, não corrige essa lacuna.
- A importação manual de CSVs do Bling e da Nuvem Pago foi aprovada para Diretores como parte de `INTEGRATION_SYNC`. A autorização ainda precisa ser aplicada no serviço/worker, com registro de sucesso e negação.

Na auditoria original, o código de aplicação e as telas permaneceram inalterados. A permissão de exportação de relatórios foi acrescentada em migration aditiva, conforme registrado abaixo. Ainda faltam desenhar autenticação/sessão do login próprio e aplicar no serviço/worker as capacidades aprovadas, incluindo importação CSV sob `INTEGRATION_SYNC`.

## Atualização do catálogo — 2026-10-11

A migration aditiva `000003_director_report_export_permission.up.sql` registra
`REPORT_EXPORT` e associa a capacidade somente aos perfis `ADMIN` e `DIRECTOR`,
conforme decisão já aprovada no `G0-DECISION-REGISTER.md`. A migration não cria
usuários, não abre uma rota de exportação e não aplica autorização no worker.
Por isso, `SEC-02` e BK-013 continuam parciais; nenhum perfil foi provisionado
e nenhuma tela foi alterada. A CI [38113352327](https://github.com/guilhermejacyzin/majucau/actions/runs/38113352327)
aplicou as migrations e passou nos jobs PostgreSQL e Windows. A autorização
por comando no worker continua pendente.

## Matriz das capacidades aprovadas — 2026-10-11

A matriz consolidada de permissões da versão 1 está em
[`APPROVED-ACCESS-MATRIX-V1-2026-10-11.md`](APPROVED-ACCESS-MATRIX-V1-2026-10-11.md).
Ela mantém as capacidades já aprovadas, registra o limite da conta compartilhada
e destaca que cancelar uma sincronização por Diretores ainda precisa de resposta.
Nenhuma permissão ou tela foi alterada.

## Alinhamento do catálogo de acesso — 12/10/2026

- A migration aditiva `000004_v1_access_catalog_alignment.up.sql` restringe os papéis cadastráveis a `ADMIN` e `DIRECTOR`, em linha com a matriz aprovada. Ela falha antes de alterar dados se algum usuário ainda estiver ligado aos perfis antigos; não promove nem reatribui usuários.
- `INTEGRATION_SYNC` passa a descrever sincronização e importação manual de CSVs aprovadas. `INTEGRATION_CANCEL` fica separado e concedido somente a `ADMIN`; Diretores não recebem cancelamento até haver aprovação explícita.
- A migration confere suas próprias condições e permanece forward-only. Ela alinha o catálogo do banco, mas não cria usuários, não fornece login e não aplica as permissões por comando no worker. `SEC-02` permanece `PARTIAL`; a execução da migration em CI ainda precisa ser confirmada.
- Nenhuma tela ou regra financeira foi alterada.
