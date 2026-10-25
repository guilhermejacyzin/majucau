# Registro técnico — tentativas de sincronização e dependências — 25/10/2026

## O que foi corrigido

- Bling e Google Drive agora contam as tentativas adicionais feitas pela política compartilhada de retry HTTP.
- A contagem é anexada ao lote de sincronização e salva no campo `integration_sync_batches.retry_count`, tanto em sucesso quanto em falha.
- A tentativa inicial não conta como retry. A contagem só aumenta quando a próxima chamada realmente começa; cancelamento antes dela não soma uma tentativa.
- Os testes de integração PostgreSQL foram ampliados para conferir que a contagem persiste no banco. O job `verify-postgres` da CI #40245161243 passou em 25/10/2026.
- Atualizei o toolchain Go para 1.26.10 e os pins do workflow. Atualizei Vite para 7.3.7 e Vitest para 4.1.12, com lockfile revisado.

## Verificações realizadas

- `scripts/verify.ps1`: passou localmente antes da última correção descrita abaixo; a auditoria npm encontrou 0 vulnerabilidades.
- Cópia isolada em `artifacts/verification/frontend-patched-deps-20261025`, mantendo o protótipo aberto sem reinstalar seus módulos: versões confirmadas Vite 7.3.7 e Vitest 4.1.12; `npm audit` encontrou 0 vulnerabilidades; lint, verificação TypeScript, 33 testes em 3 arquivos e build passaram.
- Depois da revisão final, corrigi também o campo `RetryCount` retornado pela função de sincronização quando a carga termina com sucesso. `go test ./internal/integrations/httpretry ./internal/integrations/bling ./internal/integrations/googledrive` passou localmente; os cenários que precisam de PostgreSQL foram ignorados por não haver banco de teste configurado localmente. A CI #40245161243 executou e aprovou o job PostgreSQL.
- `scripts/verify-backlog.ps1` confirmou 127 tarefas, 7 concluídas, 20 em andamento e 100 não iniciadas; 11,8% estimado concluído e 88,2% restante.
- `git diff --check`: passou antes do commit `351d167`.
- A verificação Go encontrou 0 vulnerabilidades chamadas pelo código. Um aviso de módulo não chamado permanece explicitamente fora do bloqueio da ferramenta; não foi declarado como ausência total de avisos em dependências.
- O job `verify-windows` da mesma CI falhou no smoke de remoção do serviço Windows. A mensagem original não indicava a etapa interna. Foi adicionada telemetria de erro com códigos estáveis e sem caminhos locais; essa alteração ainda precisa ser enviada e executada no runner Windows.

## Segurança e escopo

- Nenhum segredo ou credencial foi lido ou gravado pelo agente.
- Nenhuma tela aprovada, regra financeira ou dado de origem foi alterado.
- A cópia isolada de teste e o cache npm dedicado foram removidos com `scripts/SafeCleanup.psm1` depois da verificação. Os caches de projeto `artifacts/cache/go-mod/` e `artifacts/cache/go-build/` foram preservados para não atrapalhar a continuidade do desenvolvimento.
- A checagem final de Go foi executada diretamente e usou os caches compartilhados `%USERPROFILE%\go\pkg\mod` e `%LOCALAPPDATA%\go-build`; eles não foram excluídos porque podem servir a outros projetos e permanecem no registro de limpeza manual `artifacts/retained/manual-cleanup/go-global-caches-2026-10-18.txt`. A versão Go 1.26.10 baixada automaticamente também fica nesse cache compartilhado.

## Próximo passo

1. Enviar a telemetria de etapa para a branch `codex/majucau-artifact-consolidation`, repetir o smoke no runner Windows e corrigir a etapa que ele identificar.
2. Depois que a usuária salvar o Client Secret diretamente na tela do Majucau, conferir apenas o estado sanitizado, concluir OAuth e validar teste, paginação e sincronização do Bling.
3. Continuar a validação da credencial técnica de leitura do Drive e da sincronização diária às 09:30.
4. Manter BK-039 em 50% até esses critérios operacionais serem comprovados; média estimada do projeto: 11,8% concluída, 88,2% restante.
