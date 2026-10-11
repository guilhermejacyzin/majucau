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

Definir e aprovar uma tabela simples que diga, para cada perfil, quem pode consultar, sincronizar, exportar backup, editar dados financeiros/contábeis, fechar períodos, administrar usuários e restaurar backups. Depois da aprovação, implementar a verificação no worker e validar que cada ação permitida e negada é auditada. Até lá, não alterar os vínculos padrão da migration nem ampliar capacidades.

## Arquivos lidos

- `database/migrations/000001_init.up.sql`
- `cmd/worker/handler.go`
- `internal/ipc/namedpipe_windows.go`
- `REQUIREMENTS-TRACEABILITY.md`
