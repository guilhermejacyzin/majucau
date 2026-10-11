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
| SEC-R07 — Alta | Dados sensíveis poderiam sobrar em temporários ou caches depois de uma falha. | `securetemp` cria diretório privado em fluxos revisados e propaga falha de remoção com mensagem comum sanitizada. A prévia CSV preserva o caminho para nova tentativa de limpeza; backup/restauração e pacote de diagnóstico informam falhas ao remover arquivos temporários. | Testar falha de limpeza e ciclo completo em instalação/VM limpa; identificar resíduos, registrar caminho para operador autorizado e impedir que sigam para cache/CI. Outros fluxos ainda precisam da mesma auditoria. |
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
- `../../database/privileges/runtime.sql`, `../../internal/integrations/csvlimit/reader.go`, `../../internal/integrations/csvstream/disk_source_ids.go`, `../../internal/securetemp/`, `../../internal/backup/`, `../../internal/ipc/` e `../../cmd/worker/handler.go` — leitura estática dos controles relacionados.
- `../../G0-DECISION-REGISTER.md` e `../../ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md` — decisões de produto e direção da hospedagem.

## Próximos passos da BK-024

1. Ao criar novos fluxos ainda não existentes — especialmente exportação de relatórios — acrescentá-los ao inventário e exigir limites/streaming desde o desenho. Os principais caminhos já implementados foram conferidos no complemento abaixo.
2. Planejar autenticação, sessão e autorização para a hospedagem central usando apenas os perfis já aprovados.
3. Completar retenção, recuperação de credenciais, operação de backup/restore e a rotina para arquivo ausente; perguntas que mudem regra continuam pendentes de aprovação.
4. Relacionar cada risco aos testes automatizados e ensaios de ambiente, executá-los e anexar resultados verificáveis.

O mapa de riscos e a conferência estática dos principais caminhos implementados estão registrados. BK-024 permanece em andamento e não libera o uso com dados reais nem a produção.

## Complemento — conferência dos fluxos de dados no código

**Data:** 11/10/2026

**Tipo:** leitura estática adicional; nenhum teste de produto ou ensaio em máquina foi executado.

Esta conferência rastreou os principais pontos em que dados entram, são transformados, consultados, transmitidos, guardados ou exportados. Ela confirma limites já registrados, mas não comprova consumo real de memória/disco nem proteção efetiva em uma instalação limpa.

| Caminho | O que o código faz hoje | Limite e ponto ainda aberto |
|---|---|---|
| CSV do Bling e Nuvem Pago | `internal/integrations/bling/receipts_report.go` e `internal/integrations/nuvempago/future_csv.go` leem e validam uma linha por vez; os serviços chamadores controlam deduplicação e persistência. | Cada campo tem teto de 1 MiB, cada registro 4 MiB e são aceitos até 256 campos. Não há teto total de linhas/arquivo, conforme a decisão da usuária de aceitar importação total sem limite fixo, usando blocos e área temporária segura. Ainda faltam política operacional de espaço/monitoramento do disco e validação da ACL/limpeza do temporário do PostgreSQL em Windows. Não criar um teto total sem aprovação.
| Descoberta de arquivos e prévia | `internal/integrations/csvstream/file_queue.go` lê nomes em blocos de 256. A prévia usa índice e ordenação auxiliares no espaço temporário privado; o importador usa tabela temporária PostgreSQL. | O volume de linhas e o número de arquivos podem crescer. Os blocos evitam acumular a lista inteira em memória, mas não substituem limite operacional de disco nem validação real das ACLs do diretório de dados do PostgreSQL.
| API do Bling | `internal/integrations/bling/api_client.go` limita o corpo HTTP; `api_resources.go` decodifica JSON por fluxo; o sincronizador percorre páginas sequencialmente e mantém um bloco limitado. | Corpo até 4 MiB e até 100 registros por página. Retry é limitado às falhas transitórias de chamadas repetíveis. Faltam as provas integradas para cada endpoint e a confirmação dos contratos reais de paginação já listadas nas evidências do adapter.
| Banco para o painel | `internal/dashboard/stream.go` calcula resumos no banco e lê os detalhes linha por linha; as tabelas do painel são limitadas a 50 linhas cada, com campo até 4 KiB e registro até 16 KiB. | O resumo fica pequeno; detalhes acima dos limites falham fechados. O painel segue com a mesma informação e tela aprovadas.
| Worker até a interface | `internal/ipc/transport.go` transmite o painel em blocos de 512 KiB; cada mensagem tem teto de 1 MiB. `app.go` recebe os blocos, mas monta um objeto de painel completo, limitado a 16 MiB, antes de devolvê-lo à tela. | A transmissão entre processos é em blocos e tem teto. A última etapa mantém um objeto limitado em memória; não é uma transmissão registro a registro até o componente visual. O limite e a tela atual permanecem preservados. Medição de pico de memória continua pendente.
| Backup e restauração | `internal/backup/backup_stream.go` copia e cifra/descriptografa os dados em fluxo. Manifesto, chave do pacote e diretório ZIP têm limites próprios; o pacote aceita quantidade fixa de arquivos. O leitor antigo V1 continua compatível. | Os arquivos de dados não são carregados inteiros na memória. A compatibilidade V1 conserva o limite técnico histórico; a V2 não define um teto total de arquivo no aplicativo, conforme a decisão de preservar restauração grande. Falta validar espaço em disco, falhas de limpeza e restauração real em ambiente isolado.
| Segredos | `internal/security/dpapi_windows.go` usa DPAPI e rejeita blobs acima do limite; não envia o segredo ao frontend. | Buffer protegido até 1 MiB; arquivo de credencial cifrado tem leitura limitada. ACL e ciclo de provisionamento para a hospedagem central ainda não existem/foram comprovados.
| Instalador e integridade de arquivos | `internal/installer/manifest.go` decodifica manifesto limitado, valida quantidade/caminhos e calcula hash em fluxo. | Limites do manifesto e dos caminhos são estáticos; não são a política de retenção de dados do produto. Validação de instalação real depende de VM limpa.
| Logs técnicos | `internal/diagnosticslog/logger.go` limita cada registro a 32 KiB e gira até cinco arquivos de 10 MiB; `cmd/worker/diagnostics.go` grava campos permitidos, sem payload financeiro ou token. | Retenção máxima local aproximada de 50 MiB. Controle de acesso/retensão na hospedagem central ainda precisa ser definido e verificado.
| Pacote de diagnóstico | `internal/installer/diagnostics.go` gera um ZIP com dois arquivos fixos contendo o preflight sanitizado e instruções; calcula seu hash em fluxo. | Não inclui logs nem arquivos financeiros. O destino é indicado pelo chamador; validar caminho e proteção do destino em execução real continua pendente.
| Exportação de relatórios | A auditoria `ACTION-EXPORT-STATIC-AUDIT-2026-10-11.md` não encontrou rota funcional de exportação de relatórios no catálogo atual. | Quando for implementada conforme a permissão aprovada dos Diretores, deve ser projetada desde o início com leitura em blocos, limites, filtro/snapshot consistente, sanitização, autorização no serviço e auditoria. Não foi criada rota nem arquivo nesta revisão.

### Lacunas que continuam abertas

- **Sem limite total de linhas CSV:** isso preserva a decisão aprovada. Ainda é necessário definir monitoramento de espaço, comportamento seguro quando o disco estiver cheio e eventual retenção de dados. Nenhum desses pontos autoriza descartar linhas ou mudar a regra da importação.
- **Banco PostgreSQL:** o app pode manter o CSV sem teto total, mas não foi comprovada uma política operacional de capacidade/alerta/backup para crescimento contínuo. Essa política depende da hospedagem e não deve ser inventada para alterar o comportamento.
- **Autorização:** a identidade do Named Pipe protege o processo local, mas não substitui perfil por ação; o serviço central e o login próprio ainda são trabalhos pendentes.
- **Retenção e LGPD:** duração e alcance de exclusão em RAW e backups permanecem sem decisão funcional.
- **Verificação prática:** memória máxima, consumo de disco, ACLs Windows/PostgreSQL, cancelamento e falha de limpeza precisam ser validados nos ambientes previstos. Esta revisão não executou testes.

O mapa está mais detalhado, mas SEC-R01, SEC-R02, SEC-R03, SEC-R07, SEC-R08, SEC-R09 e SEC-R13 ainda impedem considerar a segurança pronta para uso real. Nenhuma tela, regra financeira, fórmula, importação ou permissão foi alterada.

## Complemento — falhas ao limpar temporários da prévia CSV

**Data:** 11/10/2026

**Tipo:** correção de caminho de erro e revisão estática. Não foram executados testes.

- Em `internal/integrations/csvstream/disk_source_ids.go`, o redimensionamento do índice temporário e a intercalação dos nomes de arquivo agora propagam falhas ao fechar/remover arquivos parciais usando `securetemp.RemoveFile`. O erro original continua visível e um `securetemp.CleanupError` mantém o caminho disponível para tratamento autorizado sem mostrá-lo na mensagem comum.
- `DiskSourceIDSet.Close` não apaga da memória o caminho do diretório quando a primeira remoção falha. Uma nova chamada a `Close` pode tentar a limpeza outra vez; após sucesso, o caminho é descartado e a chamada continua idempotente.
- A correção cobre o índice temporário usado pela prévia CSV. Os caminhos de backup e restauração foram revisados em complemento separado. A cobertura prática de falha de disco/ACL e a limpeza em VM continuam pendentes.
- A revisão executada foi `gofmt` e `git diff --check`. Não altera telas, regras financeiras, fórmulas ou a regra de importação de linhas válidas.

## Complemento — falhas ao limpar temporários de backup/restauração

**Data:** 11/10/2026

**Tipo:** correção de caminhos de erro e revisão estática. Não foram executados testes.

- `writeV2Package` agora fecha o arquivo ZIP e o arquivo de saída antes de tentar remover um pacote incompleto. Falhas ao fechar ou remover são anexadas ao erro original; a remoção usa `securetemp.RemoveFile`, que preserva o caminho para tratamento autorizado sem incluí-lo na mensagem comum.
- A restauração propaga falha ao remover um arquivo parcialmente descriptografado. A cópia temporária do pacote também informa se não conseguiu remover uma cópia parcial.
- O arquivo temporário de senha do PostgreSQL fica dentro do diretório privado da operação. A limpeza desse diretório já é verificada por `securetemp.RemoveAll`; removi uma remoção duplicada que escondia erro.
- Esta revisão cobre os temporários de pacote, cópia de restauração e senha do backup. Outros caminhos, inclusive instalação, ainda precisam ser auditados. A verificação prática em falha de disco/ACL e em VM continua pendente.
- Foi feita apenas revisão estática, `gofmt` em `backup_stream.go` e `git diff --check`. `backup.go` recebeu uma alteração localizada, sem reformatar o arquivo. Não foram executados testes. O formato V1/V2, o conteúdo do backup, as telas e as regras financeiras não mudaram.

## Complemento — fechamento do ZIP temporário de diagnóstico

**Data:** 11/10/2026

**Tipo:** correção de caminho de erro e revisão estática. Não foram executados testes.

- `WriteDiagnosticBundle` agora verifica o fechamento do ZIP também quando uma das entradas falha. Em qualquer erro, fecha o pacote e o arquivo temporário, tenta remover a cópia incompleta e combina falhas de limpeza com o erro original.
- A mensagem pública continua sanitizada por `ErrDiagnosticOutput`; a falha de remoção mantém o caminho disponível via `securetemp.CleanupError`. O ZIP final só é publicado por rename depois de fechado e sincronizado.
- Foi feita revisão estática, `gofmt` em `internal/installer/diagnostics.go` e `git diff --check`. Não foram executados testes. Nenhuma tela, regra financeira ou dado de produção foi alterado.

## Complemento — falhas ao fechar recursos da validação do instalador

**Data:** 11/10/2026

**Tipo:** correção de caminhos de erro e revisão estática. Não foram executados testes.

- A validação do manifesto lê arquivos em fluxo. Se o fechamento do arquivo, da pasta do pacote ou de uma pasta interna falhar, o pacote não é marcado como válido e nenhum arquivo aberto é devolvido ao chamador.
- A leitura do journal não devolve estado se não conseguir fechar o arquivo. A geração do hash do pacote de diagnóstico propaga a falha ao fechar a leitura. No Windows, a verificação do token de administrador também falha com erro se não conseguir fechar o handle.
- As mensagens públicas continuam genéricas; não incluem caminho completo nem detalhe do sistema operacional. O conteúdo do manifesto, journal e pacote não mudou.
- A revisão foi estática, com `gofmt` e `git diff --check`. Não foram executados testes nem build Windows. Falhas de fechamento ainda precisam de cenário controlado antes de liberar o instalador.

## Complemento — testes automatizados do pacote do instalador

**Data:** 11/10/2026

**Tipo:** teste automatizado local no Windows; não é uma homologação de instalação em máquina limpa.

- Comando executado sem reutilizar resultado anterior do teste: `go test -count=1 ./internal/installer`.
- Resultado: **passou** (`ok majucau.local/financial-intelligence/internal/installer 0.400s`), usando Go `go1.26.9 windows/amd64`.
- O cache de compilação e os módulos do Go foram direcionados temporariamente para `artifacts/cache` e removidos depois pelo módulo `scripts/SafeCleanup.psm1`; a remoção foi confirmada.
- O teste valida os cenários automatizados já existentes no pacote. Ele não comprova instalação/atualização em VM limpa nem injeta falhas de fechamento de arquivo, token, permissão ou disco. Esses ensaios continuam pendentes.
- Nenhuma tela, regra financeira, fórmula, permissão ou arquivo de dados de produto foi alterado nesta validação.

## Complemento — falhas ao fechar recursos do pré-teste Windows

**Data:** 11/10/2026

**Tipo:** correção de caminhos de erro, revisão estática e teste do pacote no Windows.

- `probePendingReboot` e `probeWebView2` agora propagam erro ao fechar chaves do Registro; quando a leitura do valor também falha, os dois erros são preservados juntos.
- `probePath` agora propaga falha ao fechar os handles usados para verificar a pasta pai ou um destino existente. A falha continua sendo convertida por `Evaluate` em `PROBE_FAILED` sem expor caminho ou mensagem do Windows.
- Falhas ao abrir um caminho continuam sendo representadas como inacessibilidade, como antes. Não foram alteradas as condições funcionais de prontidão nem o conteúdo exibido em caso normal.
- `gofmt`, `git diff --check` e `go test -count=1 ./internal/installer` foram executados no Windows; o pacote passou (`0.437s`). Os caches foram direcionados a `artifacts/cache` e removidos pelo módulo seguro do projeto.
- Não foi possível injetar uma falha real de fechamento com os testes existentes. Instalação/atualização em VM limpa e ensaios controlados de falha continuam pendentes.

## Complemento — fechamento de recursos na verificação e restauração de backups

**Data:** 11/10/2026

**Tipo:** correção de caminhos de erro e teste automatizado local no Windows.

- A leitura de pacotes V1/V2 combina erros de leitura, descriptografia e fechamento dos arquivos. A verificação e a restauração também tratam falhas ao fechar o pacote, as entradas ZIP e os arquivos temporários restaurados.
- As mensagens de fechamento usam rótulos genéricos; o erro original continua encadeado para diagnóstico interno sem incluir seu caminho na mensagem comum. A verificação marca o pacote como inválido se não conseguir fechar um recurso.
- Os testes locais de backup passaram com `go test -count=1 ./internal/backup` (`1.541s`). A integração local foi pulada porque as duas URLs de banco descartável não estavam configuradas neste computador. Depois, a CI [#403](https://github.com/guilhermejacyzin/majucau/actions/runs/38135368926) executou a restauração real com dois PostgreSQL 18.6 descartáveis; o passo de backup/restauração e todos os jobs do workflow terminaram com sucesso para o commit `0e4a400`.
- Os caches do Go ficaram temporariamente em `artifacts/cache` e foram removidos pelo módulo `scripts/SafeCleanup.psm1`. `gofmt`, `git diff --check` e a busca por fechamentos ignorados no código de produção de `internal/backup` também foram concluídos.
- O formato dos backups, a compatibilidade V1, o conteúdo dos dados, as telas e as regras financeiras não foram alterados. Ainda faltam falhas de fechamento injetadas e ensaios de instalação/atualização em VM limpa; o teste integrado de backup/restore passou na CI, mas não substitui esses ensaios.

## Complemento — fechamento do canal Named Pipe no Windows

**Data:** 11/10/2026

**Tipo:** correção de encerramento do canal local e teste automatizado no Windows.

- O servidor agora espera fechar as instâncias abertas e devolve erros de desconexão/fechamento ao chamador; `runWorkerPipe` combina esse erro ao resultado do worker.
- O fechamento do cliente combina erros de fechamento ao resultado da chamada ou ao cancelamento. Fechar novamente um arquivo já encerrado continua sendo tratado como operação idempotente.
- O código 233 (`ERROR_PIPE_NOT_CONNECTED`) é tratado como encerramento normal porque significa que já não há cliente conectado; os demais erros de desconexão e fechamento são propagados. Referências: [Microsoft Learn — DisconnectNamedPipe](https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-disconnectnamedpipe) e [Microsoft Learn — códigos de erro do sistema](https://learn.microsoft.com/en-us/windows/win32/debug/system-error-codes--0-499-).
- `go test -count=1 ./internal/ipc` passou no Windows (`0.689s`), incluindo o ensaio de fechamento do servidor que interrompe uma conexão pendente. Os testes agora verificam também os erros de fechamento nos recursos que usam. Caches direcionados a `artifacts/cache` foram removidos pelo módulo seguro do projeto.
- Permanecem para auditoria os handles do evento de conexão, processo/token e o descritor de segurança liberado por `LocalFree`, além dos resultados da rotina que cancela uma conexão sobreposta. Login, permissões por perfil e configuração de acesso remoto continuam pendentes conforme SEC-R01.
- Não foram alteradas telas, regras financeiras, direitos da DACL nem dados persistidos.

## Complemento — cancelamento seguro e liberação de recursos do Named Pipe

**Data:** 11/10/2026

**Tipo:** correção de encerramento de conexão e propagação de falhas de limpeza.

- Ao cancelar uma conexão pendente, o worker agora pede o cancelamento e espera
  o resultado final da operação sobreposta antes de liberar a estrutura e o
  evento. `ERROR_NOT_FOUND` no cancelamento é aceito como ausência de operação
  pendente; falhas reais são devolvidas junto com o erro original. A espera é
  necessária porque `CancelIoEx` só solicita o cancelamento e não aguarda sua
  conclusão; a Microsoft orienta consultar o resultado antes de reutilizar ou
  liberar a estrutura `OVERLAPPED` ([CancelIoEx](https://learn.microsoft.com/en-us/windows/win32/fileio/cancelioex-func),
  [GetOverlappedResult](https://learn.microsoft.com/en-us/windows/win32/api/ioapiset/nf-ioapiset-getoverlappedresult)).
- As falhas ao fechar o evento, os handles de processo/token e ao liberar o
  descritor de segurança ou o texto do SID agora são verificadas e propagadas.
  Uma conexão não autorizada continua sendo recusada sem derrubar o worker;
  somente falha real ao limpar recurso é reportada como erro do serviço.
- `LocalFree` é considerado bem-sucedido quando retorna nulo; retorno não nulo
  agora gera erro registrado, conforme a [documentação da Microsoft](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-localfree).
- `gofmt`, `git diff --check`, `go test -count=1 ./internal/ipc` (Windows,
  aprovado em `0.666s`) e `go vet ./internal/ipc` passaram. O cache ficou em
  `artifacts/cache` e foi removido pelo módulo seguro do projeto.
- Continua pendente revisar handles e liberações em outros módulos (`DPAPI`,
  `securetemp` e trava do instalador), testar falhas de limpeza de forma
  controlada, terminar o mapa de riscos e instalar/atualizar em VM limpa.
- Nenhuma tela, regra financeira, permissão, DACL ou dado persistido foi
  alterado.

## Complemento — recursos Windows em segredos, temporários e instalador

**Data:** 11/10/2026

**Tipo:** revisão de memória nativa e handles do Windows.

- O resultado de `CryptProtectData`/`CryptUnprotectData` é copiado para memória
  Go e a memória nativa agora é liberada com `LocalFree` verificado. Se a
  liberação falhar, a cópia é zerada e a falha é devolvida; buffers vazios ou
  acima do limite configurado são recusados.
- A criação de pastas temporárias protegidas agora devolve erros ao liberar o
  descritor. Se a pasta vazia já foi criada, o código tenta removê-la; se a
  remoção também falhar, registra o caminho em `securetemp.CleanupError` para
  limpeza manual autorizada.
- A trava do instalador combina falhas ao liberar o mutex e fechar seu handle,
  e também propaga eventual erro ao fechar o handle quando outra instalação já
  está ativa.
- Foi corrigida a liberação indevida de descritores retornados pelos wrappers
  `SecurityDescriptorFromString`, `GetNamedSecurityInfo` e `GetSecurityInfo` do
  `golang.org/x/sys/windows`: esses wrappers copiam os dados para a memória Go
  e liberam o buffer nativo internamente. Não se chama `LocalFree` nos objetos
  Go retornados.
- `gofmt`, `git diff --check`, `go test -count=1` e `go vet` passaram no Windows
  para `internal/security`, `internal/securetemp`, `internal/installer`,
  `internal/diagnosticslog` e `internal/ipc`. Os caches ficaram em
  `artifacts/cache`; foram removidos pelo módulo seguro ao final.
- Falhas de limpeza ainda precisam de injeção controlada, e a auditoria maior
  ainda requer instalação/atualização em VM limpa, validação de rede e testes
  de acesso por perfil.
- Nenhuma tela ou regra financeira foi alterada.

## Complemento — erros de encerramento nos fluxos CSV em streaming

**Data:** 11/10/2026

**Tipo:** correção de propagação de falhas durante encerramento e limpeza.

- Filas de nomes CSV agora informam falhas ao fechar. A fila temporária em disco,
  os leitores das partes de ordenação e os diretórios de importação também
  devolvem os erros de fechamento ao chamador.
- As prévias e importações de recebimentos do Bling e de recebimentos futuros do
  Nuvem Pago propagam falhas ao fechar arquivos, diretório raiz e fila, inclusive
  quando há cancelamento, tamanho de linha excedido ou falha ao reiniciar a
  leitura. O motivo original e o erro de fechamento são preservados juntos.
- O worker agora retorna ao serviço eventual erro ao fechar o arquivo de log
  técnico, em vez de descartá-lo.
- `gofmt`, `git diff --check`, os testes e `go vet` dos quatro pacotes alterados
  passaram localmente. Também passaram os testes de todos os pacotes em
  `cmd`, `database` e `internal`, além de `go vet ./cmd/... ./database/... ./internal/...`.
  O comando `go test ./...` não conseguiu preparar somente o pacote principal:
  `main.go` incorpora `frontend/dist`, que não estava gerada nesta cópia. A CI
  prepara esse arquivo antes dos testes completos. Os caches desta execução
  foram centralizados em `artifacts/cache` e removidos pelo módulo seguro após
  as verificações; a CI desta revisão será registrada após o push.
- Continuam pendentes ensaios controlados que provoquem falhas reais de limpeza,
  a auditoria mais ampla, VM limpa, validação de rede e testes de acesso por
  perfil.
- Nenhuma tela, regra financeira, permissão ou dado persistido foi alterado.
