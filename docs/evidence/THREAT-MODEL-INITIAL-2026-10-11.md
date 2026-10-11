# Mapa inicial de riscos de segurança — BK-024

**Data:** 11/10/2026
**Escopo:** primeira leitura estática do repositório e das decisões vigentes.
**Estado:** planejamento inicial; não é teste de invasão nem aprovação para produção.

## Em palavras simples

Este documento lista o que precisa ser protegido, por onde os dados passam, quais proteções já aparecem no projeto e quais ainda precisam ser comprovadas. A lista ajuda a planejar o trabalho; escrever uma proteção aqui não significa que ela já esteja implementada.

Nenhuma tela aprovada ou regra financeira foi alterada. Não foram executados testes nesta revisão e nenhum ambiente de produção foi acessado.

## O que precisa ser protegido

| Código | Informação ou recurso | Motivo |
| --- | --- | --- |
| A-01 | Lançamentos, saldos, DRE, fluxo de caixa e resultados da calculadora | Acesso ou alteração indevidos podem levar a decisões incorretas. |
| A-02 | Senhas técnicas, tokens e credenciais do banco e das integrações | Quem obtiver esses dados pode tentar entrar nos serviços conectados. |
| A-03 | Relatórios e arquivos CSV recebidos de fontes externas | Um arquivo errado ou malicioso pode causar erro, duplicidade ou consumo excessivo do computador. |
| A-04 | Backups e material usado para restaurar o sistema | Cópias precisam continuar privadas, íntegras e recuperáveis. |
| A-05 | Registros de auditoria e diagnósticos | Ajudam a entender ações e falhas, mas podem revelar dados se forem gravados sem cuidado. |
| A-06 | Código, dependências e instalador | Uma alteração ou biblioteca comprometida pode afetar todos os dados. |
| A-07 | Arquivos temporários, caches e artefatos de CI | Podem guardar cópias de dados e segredos após o processamento. |

## Caminhos dos dados e limites de confiança

1. **Bling e arquivos externos → worker/API do Majucau → PostgreSQL.** A entrada externa deve ser tratada como não confiável. A leitura, validação e gravação precisam respeitar limites, cancelamento, idempotência e transação.
2. **Aplicativo Windows → Named Pipe → worker → PostgreSQL.** O projeto tem proteção do canal local por SID e DACL. Essa proteção identifica um processo autorizado; não substitui a verificação do perfil da pessoa em cada ação.
3. **Hospedagem privada centralizada → login próprio → serviço → PostgreSQL privado.** Esta é a direção aprovada, mas a hospedagem e o login central ainda não estão prontos. Provedor, região e detalhes da operação continuam pendentes.
4. **Arquivos temporários, backups e restauração.** Cópias de trabalho precisam ficar em área privada, ter tamanho controlado e ser removidas com falha de limpeza informada.
5. **GitHub e CI.** O repositório é público; segredos e dados reais não podem entrar no Git nem em logs ou artefatos públicos.

## Riscos encontrados e trabalho necessário

“Proteção encontrada” descreve evidência no código ou nos registros. “Falta comprovar” é trabalho futuro, não uma afirmação de que um controle foi aprovado ou implantado.

| ID / prioridade antes da produção | Situação que pode causar problema | Proteção já encontrada | O que falta comprovar |
| --- | --- | --- | --- |
| SEC-R01 — Bloqueador | Uma pessoa poderia executar uma ação além do perfil aprovado se o serviço não conferir a autorização em cada comando. | Há tabelas de perfis e permissões; o Named Pipe protege o processo local. A auditoria estática registra que o worker ainda não aplica a permissão por comando. | Implementar e validar o login próprio e a autorização no serviço/worker conforme as capacidades já aprovadas. Não mudar telas ou aumentar permissões por inferência. |
| SEC-R02 — Bloqueador | Uma hospedagem configurada incorretamente poderia deixar serviço ou banco acessível a pessoas de fora. | A direção aprovada exige hospedagem privada e PostgreSQL sem acesso público. | Definir operação do ambiente e demonstrar rede privada, transporte protegido, configuração segura e acesso ao banco restrito antes de usar dados reais. |
| SEC-R03 — Bloqueador | Token ou senha poderia vazar em arquivo, log, navegador, Git, diagnóstico ou backup. | Há armazenamento local com DPAPI e regras que proíbem segredos em frontend, Git e logs. A configuração central ainda não foi definida. | Comprovar ACL da conta de serviço, armazenamento e recuperação de credenciais no ambiente escolhido, rotação/revogação e ausência de segredos nos canais de saída. |
| SEC-R04 — Alta | Arquivo externo muito grande ou malformado poderia consumir memória, disco ou tempo e interromper a importação. | Existem leitura em fluxo, limites de campo/registro/colunas em CSV, limite de resposta do Bling e áreas temporárias privadas em fluxos já auditados. | Fazer inventário por fonte; verificar que cada leitura, transformação e exportação usa blocos limitados, com limite de linha, quantidade, descompactação, timeout e cancelamento. Os limites atuais não provam cobertura de todos os formatos. |
| SEC-R05 — Alta | Repetir o mesmo arquivo ou lote poderia duplicar lançamentos; falha técnica poderia deixar parte inconsistente. | O projeto registra idempotência, atomicidade e avanço do cursor somente após confirmação. Para CSV, linhas válidas podem ser gravadas com o lote `PARTIAL`; falha técnica/cancelamento deve reverter o lote. | Demonstrar esses comportamentos em cada fluxo real, inclusive replay, duplicidade, erro de linha, rollback técnico e cancelamento. Manter a regra de contagem já registrada. |
| SEC-R06 — Alta | Diagnósticos, logs, relatórios ou artefatos poderiam expor valor financeiro, dados pessoais ou credenciais. | Há log estruturado com campos e tamanhos limitados, rotação e omissão de payload, valores, token e caminho de arquivo; a CI faz verificações de segredo. | Revisar todas as saídas, definir retenção e acesso, e demonstrar que relatórios autorizados não expõem dados fora do escopo. |
| SEC-R07 — Alta | Dados sensíveis poderiam sobrar em temporários ou caches depois de uma falha. | `securetemp` cria diretório privado em fluxos revisados e propaga falha de remoção com mensagem comum sanitizada. | Testar falha de limpeza e ciclo completo em instalação/VM limpa; identificar resíduos, registrar caminho para operador autorizado e impedir que sigam para cache/CI. |
| SEC-R08 — Bloqueador | Um backup poderia ser lido, alterado ou não restaurar quando necessário. | Há leitura em fluxo para V1 e decisão aprovada de usar V2 em blocos autenticados para novos backups. Há validações e testes unitários registrados. | Completar integração operacional e executar restauração real em ambiente isolado, incluindo arquivo corrompido, interrompido, chave incorreta, compatibilidade V1 e confirmação do backup anterior. |
| SEC-R09 — Alta | Acesso direto ou excessivo ao banco poderia permitir alterar dados de origem ou trilha de auditoria. | A função de banco de runtime é não proprietária; `raw_records` e `audit_events` têm restrições de alteração para essa função. | Revisar privilégios no ambiente central, validar credenciais e conexões, e impedir acesso público ao PostgreSQL. A permissão por usuário ainda requer controle na aplicação. |
| SEC-R10 — Alta | Uma dependência comprometida ou vulnerável poderia inserir código malicioso. | CI executa verificações Go/npm e SBOMs; uma correção de dependências tem evidência recente. | Fechar inventário de componentes nativos do instalador, licenças e rotina de atualização/revisão; manter falhas alcançáveis bloqueando release. |
| SEC-R11 — Alta | Um arquivo que ainda não chegou às 09:30 ou um dado antigo poderia ser tratado como atualização completa. | A solicitação diária para 09:30 está registrada; a política exige preservar o último dado válido e não substituir falha por zero. | Definir fuso horário e comportamento quando faltar arquivo; exibir estado/frescor e validar que lote incompleto não pareça sincronização completa. Não presumir uma resposta para a rotina. |
| SEC-R12 — Alta | Retry ou paginação inadequados poderiam duplicar, perder ou deixar dados antigos sem indicação. | Existe uma regra padrão de retry limitado para chamadas repetíveis; o adapter Bling tem limites e checkpoint transacional registrados. | Confirmar cursor incremental oficial por consulta, cobrir todos os endpoints e provar que apenas operações seguras são repetidas. Erro permanente, cancelamento e resposta inválida não devem avançar cursor. |
| SEC-R13 — Bloqueador | Dados pessoais poderiam ficar guardados por tempo indefinido ou continuar aparecendo em cópias. | O requisito LGPD está identificado na matriz de rastreabilidade. | Definir com responsáveis a retenção, exclusão/anonimização, auditoria, alcance em RAW e backups, e a evidência de que uma exclusão foi aplicada. SEC-03 segue bloqueado. |
| SEC-R14 — Alta | Uma exportação poderia entregar dados a alguém sem permissão ou formar um arquivo perigoso para abrir. | Diretores têm capacidade aprovada para exportar backup e relatórios. A permissão de exportação de relatório existe na migration. | Implementar a rota/exportação e verificar perfil no serviço, filtros, limites, sanitização de campos e trilha de auditoria antes de liberar o recurso. |

## Verificações que devem acompanhar os riscos

Os itens abaixo são propostas de evidência para o plano de segurança. **Ainda não foram executados por este documento.**

| Verificação | O que deverá demonstrar | Riscos |
| --- | --- | --- |
| SEC-T01 — Acesso por perfil | Tentativas autorizadas e negadas em cada comando; a conta compartilhada dos Diretores aparece como tal, sem atribuir a ação a uma pessoa específica. | R01, R14 |
| SEC-T02 — Limite de rede | Banco não acessível publicamente; conexões da aplicação protegidas e permitidas somente nos pontos definidos. | R02, R09 |
| SEC-T03 — Segredos | Token e senha ausentes do frontend, Git, logs, diagnósticos e backup comum; acesso e recuperação compatíveis com o serviço. | R03, R06 |
| SEC-T04 — Arquivos grandes e malformados | Fluxo com memória limitada; limites respeitados; erro/cancelamento não deixa gravação parcial indevida nem temporário silencioso. | R04, R07 |
| SEC-T05 — Replay e transação | Reprocessamento idempotente, lote parcial conforme regra aprovada, rollback em falha técnica e cursor confirmado somente após commit. | R05, R12 |
| SEC-T06 — Backup/restauração | Integridade autenticada, falha fechada para pacote alterado/incompleto, compatibilidade V1 e restauração real isolada. | R08 |
| SEC-T07 — Privacidade e auditoria | Logs/relatórios sanitizados, acesso registrado e procedimento de retenção/exclusão verificado em cópias abrangidas. | R06, R13, R14 |
| SEC-T08 — Componentes e publicação | Scans e SBOM revisados, licenças tratadas e nenhum segredo/dado real publicado em artefatos. | R06, R10 |
| SEC-T09 — Operação diária | Execução no horário aprovado indica arquivo ausente, erro ou dado desatualizado; não apresenta sucesso nem zero indevido. | R11, R12 |
| SEC-T10 — Recuperação e instalação | Instalação/upgrade/reinício e restauração em ambiente isolado mantêm permissões, serviço e dados íntegros. | R03, R07, R08 |

## Evidências consultadas

- `../../AGENTS.md` — regras obrigatórias de dados, streaming, retry, segurança e governança.
- `../../REQUIREMENTS-TRACEABILITY.md` — estados SEC-01 a SEC-04.
- `docs/evidence/ACCESS-CONTROL-AUDIT-2026-10-11.md` — diferença entre proteção do processo local e autorização por perfil.
- `docs/evidence/STREAMING-SECURITY-CI-2026-10-03.md` — evidências de limites, temporários e Named Pipe; também registra o que a CI não cobre.
- `docs/evidence/WORKER-STRUCTURED-DIAGNOSTICS-2026-10-11.md` — conteúdo, tamanho, retenção e limites atuais dos logs.
- `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md` — verificações e pendências da cadeia de dependências.
- `../../database/privileges/runtime.sql`, `../../internal/integrations/csvlimit/reader.go`, `../../internal/securetemp/`, `../../internal/backup/`, `../../internal/ipc/` e `../../cmd/worker/handler.go` — leitura estática dos controles relacionados.
- `../../G0-DECISION-REGISTER.md` e `../../ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md` — decisões de produto e direção da hospedagem.

## Próximos passos da BK-024

1. Mapear, uma a uma, todas as fontes e saídas de dados para confirmar limites e processamento em fluxo.
2. Planejar autenticação, sessão e autorização para a hospedagem central usando apenas os perfis já aprovados.
3. Completar retenção, recuperação de credenciais, operação de backup/restore e a rotina para arquivo ausente; perguntas que mudem regra continuam pendentes de aprovação.
4. Relacionar cada risco aos testes automatizados e ensaios de ambiente, executá-los e anexar resultados verificáveis.

Esta primeira etapa cobre apenas o mapa inicial. BK-024 permanece em andamento e não libera o uso com dados reais nem a produção.
