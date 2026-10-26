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
- O job `verify-windows` da CI #40245161243 falhou no smoke de remoção do serviço Windows. A mensagem original não indicava a etapa interna. O commit `f7ad842` adicionou códigos estáveis de etapa e apresentação sanitizada no smoke. A execução seguinte, CI #40245975900, passou por completo, inclusive no smoke Windows; a falha anterior não se repetiu e não teve causa determinística confirmada.

## Segurança e escopo

- Nenhum segredo ou credencial foi lido ou gravado pelo agente.
- Nenhuma tela aprovada, regra financeira ou dado de origem foi alterado.
- A cópia isolada de teste e o cache npm dedicado foram removidos com `scripts/SafeCleanup.psm1` depois da verificação. Os caches de projeto `artifacts/cache/go-mod/` e `artifacts/cache/go-build/` foram preservados para não atrapalhar a continuidade do desenvolvimento.
- A checagem final de Go foi executada diretamente e usou os caches compartilhados `%USERPROFILE%\go\pkg\mod` e `%LOCALAPPDATA%\go-build`; eles não foram excluídos porque podem servir a outros projetos e permanecem no registro de limpeza manual `artifacts/retained/manual-cleanup/go-global-caches-2026-10-18.txt`. A versão Go 1.26.10 baixada automaticamente também fica nesse cache compartilhado.

## Próximo passo

- O backlog oficial v2 separa BK-039 (trabalhos em segundo plano, 50%), BK-040 (cobertura de dados Bling, 0%), BK-041 (adaptador Bling, 80%) e BK-042 (Drive, 70%).

1. Manter os códigos de etapa sanitizados caso o smoke do serviço volte a falhar; a CI #40245975900 confirmou a execução atual.
2. Depois que a usuária salvar o Client Secret diretamente na tela do Majucau, conferir apenas o estado sanitizado, concluir OAuth e validar teste, paginação e sincronização do Bling (BK-041).
3. Continuar a validação da credencial técnica de leitura do Drive e da sincronização diária às 09:30 (BK-042 e BK-039).
4. Manter BK-039 (execução de trabalhos em segundo plano) em 50% até comprovar continuidade e integração; BK-041 permanece em 80% e aguarda autenticação/homologação Bling; BK-042 está em 70% e aguarda acesso técnico e homologação Drive; BK-040 continua em 0% até validar cobertura real dos dados do Bling. Média estimada do projeto: 11,8% concluída, 88,2% restante.
