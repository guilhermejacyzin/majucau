# Matriz de acesso aprovada — versão 1

**Data:** 11/10/2026

**Status:** capacidades aprovadas documentadas; implementação e autenticação ainda pendentes.

Esta matriz apenas organiza decisões já registradas pela responsável. Ela não altera telas nem amplia permissões.

## Capacidades aprovadas

| Ação | ADMIN — Financeiro (`financeiro@majucau.com.br`) | DIRECTOR — conta compartilhada (`contato@majucau.com.br`) | Base da decisão |
| --- | --- | --- | --- |
| Consultar as informações disponíveis | Sim | Sim | “Financeiro é o ADMIN”; Diretores podem consultar. |
| Sincronizar integrações oficiais | Sim | Sim | Diretores podem sincronizar. |
| Importar CSV manual do Bling e da Nuvem Pago | Sim | Sim | A importação foi aprovada como parte da sincronização (`INTEGRATION_SYNC`). |
| Exportar backup | Sim | Sim | Diretores podem exportar backups. |
| Exportar relatórios | Sim | Sim | Diretores podem exportar relatórios. |
| Configurar, conectar ou revogar credenciais e integrações | Sim, como ADMIN | Não concedido à conta DIRECTOR | A aprovação de sincronização/importação não autoriza administrar credenciais ou integrações. |
| Outras ações administrativas | Sim, como ADMIN | Não concedidas pela decisão atual | Manter o menor privilégio até existir aprovação explícita. |
| Cancelar uma sincronização em andamento | Sim, como ADMIN | **Pendente de confirmação** | A permissão técnica existente agrupa iniciar e cancelar; a aprovação dos Diretores ainda não esclareceu o cancelamento. Até a confirmação, bloquear para DIRECTOR. |

Os perfis `ACCOUNTING` e `VIEWER` ficam fora da versão 1. A conta compartilhada de Diretores não permite identificar qual diretor realizou cada ação; a auditoria registra a conta compartilhada e a ação, sem atribuir a ação a uma pessoa específica.

## Situação técnica verificada

- As migrations existentes registram `DASHBOARD_READ`, `INTEGRATION_READ`, `INTEGRATION_SYNC` e `BACKUP_EXPORT` para `DIRECTOR`; a migration `000003_director_report_export_permission.up.sql` acrescenta `REPORT_EXPORT` para `ADMIN` e `DIRECTOR`.
- `INTEGRATION_SYNC` ainda descreve iniciar **ou cancelar** uma sincronização. Essa descrição não será usada para presumir a aprovação de cancelamento por Diretores.
- A leitura estática do worker não encontrou checagem de perfil/permissão em cada comando. Portanto, o conteúdo da migration não comprova que as ações estejam protegidas durante a execução.
- O login próprio do Majucau foi escolhido, mas o mecanismo de sessão, MFA, recuperação, provisionamento e identidade para hospedagem continuam pendentes. A estrutura local vinculada ao SID do Windows não comprova autenticação na hospedagem.

## Próximo passo

Confirmar se Diretores podem cancelar uma sincronização em andamento. Depois, especificar e implementar a autenticação/provisionamento e aplicar cada permissão no serviço/worker, com auditoria de ações permitidas e negadas. Nenhuma permissão foi alterada nesta atualização.

## Fontes

- `G0-DECISION-REGISTER.md` — perfis de acesso, conta compartilhada e importação CSV.
- `database/migrations/000001_init.up.sql` e `database/migrations/000003_director_report_export_permission.up.sql`.
- `docs/evidence/ACCESS-CONTROL-AUDIT-2026-10-11.md`.
