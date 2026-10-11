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
| SEC-R03 — Bloqueador | Token ou senha poderia vazar em arquivo, log, navegador, Git, diagnóstico ou backup. | Os tokens são cifrados com DPAPI e há regras que proíbem segredos em frontend, Git e logs. `NewDPAPIStore` agora exige caminho absoluto, rejeita reparse points, confere DACL protegida e aplica permissões restritas à identidade atual do processo, SYSTEM e Administradores na pasta de secrets, arquivo de entropia, arquivo do segredo e temporário da gravação atômica. Operações do cofre são serializadas e buffers cifrados lidos/produzidos são zerados após uso. O teste Windows verificou essas ACLs sob a identidade de desenvolvimento. A proteção depende de uma pasta pai com a mesma DACL privada; o instalador ainda não cria essa pasta nem registra/configura a conta do serviço. O worker também lê `MAJUCAU_DATABASE_URL`, mas o instalador não configura essa variável nem há provisionamento seguro da senha do banco. | Implementar o registro transacional do serviço e provisionar, sob a identidade real do serviço, a pasta pai com ACL exata antes de iniciar o cofre; comprovar acesso mínimo, recuperação/rotação e ausência de segredo em ambiente persistente, argumentos, logs, diagnóstico ou backup. Configurar a credencial do banco com proteção sob a identidade do serviço; nunca em texto aberto na configuração. Validar o ciclo e os casos de falha em VM Windows limpa antes de usar dados reais. |
| SEC-R04 — Alta | Arquivo externo muito grande ou malformado poderia consumir memória, disco ou tempo e interromper a importação. | Existem leitura em fluxo, limites de campo/registro/colunas em CSV, limite de resposta do Bling e áreas temporárias privadas em fluxos já auditados. | Fazer inventário por fonte; verificar que cada leitura, transformação e exportação usa blocos limitados, com limite de linha, quantidade, descompactação, timeout e cancelamento. Os limites atuais não provam cobertura de todos os formatos. |
| SEC-R05 — Alta | Repetir o mesmo arquivo ou lote poderia duplicar lançamentos; falha técnica poderia deixar parte inconsistente. | O projeto registra idempotência, atomicidade e avanço do cursor somente após confirmação. Para CSV, linhas válidas podem ser gravadas com o lote `PARTIAL`; falha técnica/cancelamento deve reverter o lote. | Demonstrar esses comportamentos em cada fluxo real, inclusive replay, duplicidade, erro de linha, rollback técnico e cancelamento. Manter a regra de contagem já registrada. |
| SEC-R06 — Alta | Diagnósticos, logs, relatórios ou artefatos poderiam expor valor financeiro, dados pessoais ou credenciais. | Há log estruturado com campos e tamanhos limitados, rotação e omissão de payload, valores, token e caminho de arquivo; a CI faz verificações de segredo. | Revisar todas as saídas, definir retenção e acesso, e demonstrar que relatórios autorizados não expõem dados fora do escopo. |
| SEC-R07 — Alta | Dados sensíveis poderiam sobrar em temporários ou caches depois de uma falha. | `securetemp` cria diretório privado nos fluxos de dados grandes revisados e propaga falha de remoção com mensagem comum sanitizada. A prévia CSV preserva o caminho para nova tentativa de limpeza; backup/restauração e pacote de diagnóstico informam falhas ao remover temporários. Uma busca estática em código de produção encontrou remoção recursiva concentrada na API privada de `securetemp`; CSV, backup e restauração usam workspace privado. Os demais `CreateTemp` revisados são arquivos de controle limitados ou a troca atômica do arquivo protegido do DPAPI. | Testar falha de limpeza e ciclo completo em instalação/VM limpa; verificar ACL efetiva, disco cheio/interrupção, registrar resíduos para operador autorizado e impedir que sigam para cache/CI. A busca é estática e não comprova o comportamento do sistema em execução; outros fluxos ainda precisam de auditoria. |
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
| Segredos | `internal/security/dpapi_windows.go` usa DPAPI, rejeita blobs acima do limite e não envia o segredo ao frontend. `internal/security/dpapi_acl_windows.go` valida caminho sem reparse point e protege diretório, entropia, segredo e temporário com DACL explícita para a identidade atual, SYSTEM e Administradores. | Buffer protegido até 1 MiB; arquivo de credencial cifrado tem leitura limitada. Teste e `go vet` passaram sob a identidade de desenvolvimento, mas a conta real do worker, a DACL da pasta pai criada pelo instalador e a instalação em VM limpa não foram validadas. O provisionamento do banco e o ciclo operacional de rotação/recuperação também continuam pendentes.
| Instalador e integridade de arquivos | `internal/installer/manifest.go` decodifica manifesto limitado, valida quantidade/caminhos e calcula hash em fluxo. | Limites do manifesto e dos caminhos são estáticos; não são a política de retenção de dados do produto. Validação de instalação real depende de VM limpa.
| Logs técnicos | `internal/diagnosticslog/logger.go` limita cada registro a 32 KiB e gira até cinco arquivos de 10 MiB; `cmd/worker/diagnostics.go` grava campos permitidos, sem payload financeiro ou token. | Retenção máxima local aproximada de 50 MiB. Controle de acesso/retensão na hospedagem central ainda precisa ser definido e verificado.
| Pacote de diagnóstico | `internal/installer/diagnostics.go` gera um ZIP com dois arquivos fixos contendo o preflight sanitizado e instruções; calcula seu hash em fluxo. | Não inclui logs nem arquivos financeiros. O destino é indicado pelo chamador; validar caminho e proteção do destino em execução real continua pendente.
| Exportação de relatórios | A auditoria `ACTION-EXPORT-STATIC-AUDIT-2026-10-11.md` não encontrou rota funcional de exportação de relatórios no catálogo atual. | Quando for implementada conforme a permissão aprovada dos Diretores, deve ser projetada desde o início com leitura em blocos, limites, filtro/snapshot consistente, sanitização, autorização no serviço e auditoria. Não foi criada rota nem arquivo nesta revisão.

### Escopo transacional da importação CSV

Leitura estática de `ReceiptImportService.Import` e `FutureImportService.Import` em 11/10/2026:

- Cada serviço abre uma transação PostgreSQL antes de percorrer a pasta e só confirma depois de gravar o resultado do lote. Os parsers percorrem os arquivos linha a linha e a deduplicação guarda fingerprints HMAC em uma tabela temporária da mesma transação; o PostgreSQL descarta essa tabela no commit ou rollback.
- Erros de validação por linha são contabilizados; as linhas válidas e o lote `PARTIAL` são confirmados juntos. Uma falha técnica ou cancelamento retorna antes do commit e o rollback descarta as alterações daquela importação.
- Portanto, hoje a atomicidade técnica cobre a pasta selecionada inteira, embora a leitura e a persistência aconteçam registro a registro. Uma pasta muito grande pode manter a transação e a conexão abertas por bastante tempo. Dividir a confirmação em blocos mudaria o que permanece salvo após uma falha; isso precisa seguir a decisão de atomicidade pendente e não deve ser alterado por inferência.
- A inspeção foi somente leitura; nenhum teste, build ou mudança de regra/tela foi feito nesta revalidação.

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

## Auditoria estática adicional de leituras e streaming — 11/10/2026

Foi feita uma busca estática no código de produção Go e nos scripts por leitura integral de arquivos/respostas, montagem de buffers e decodificação de dados. Testes, dependências, documentação e outputs foram excluídos da busca.

- A leitura de CSV continua registro a registro, com limites por campo e registro; a quantidade total de linhas permanece sem teto, conforme a decisão aprovada. O armazenamento temporário e a limpeza em caso de disco cheio/ACL ainda precisam de validação operacional.
- As listas e detalhes do Bling passam `resp.Body` diretamente para `json.Decoder` através de `io.LimitedReader`; a resposta tem limite de 4 MiB e cada página aceita até 100 registros. A resposta OAuth usa leitura completa limitada a 4 MiB porque contém um único envelope de tokens.
- As consultas de linhas do dashboard e da fila CSV usam `Rows.Next()` e percorrem os resultados em fluxo. A lista de estados de integração é um conjunto pequeno de controle.
- As leituras completas restantes encontradas são limitadas: arquivo de credencial DPAPI, manifesto/chave do pacote de backup e resposta de token. `Read-BoundedTextFile` limita os arquivos dos scripts antes de `ReadToEnd`; os usos encontrados ficam entre 4 KiB e 4 MiB.
- A busca não encontrou `os.ReadFile`, `File.ReadAllBytes`, `response.json()`, `response.text()` ou `Buffer.concat` nos caminhos pesquisados. Isso é evidência estática delimitada; não comprova consumo de memória em execução, resposta real do Bling, disco cheio, ACL ou máquina limpa.

BK-024 permanece em 30%: esta conferência documenta mais evidência de streaming, mas não substitui os ensaios controlados, o fechamento do mapa de riscos, a validação em VM limpa, rede e perfis. Nenhuma tela ou regra financeira foi alterada; nenhum teste ou build foi executado nesta conferência.

## Complemento — escopo da transação CSV e cobertura de retry HTTP — 11/10/2026

**Tipo:** revisão estática de dois fluxos CSV e de chamadas HTTP de produção; sem mudança de comportamento.

- Os importadores de recebimentos do Bling e de recebíveis futuros do Nuvem Pago percorrem arquivos e linhas em fluxo, com limites CSV de 1 MiB por campo, 4 MiB por registro e 256 campos. A fila de nomes lê 256 itens por vez; a prévia usa área temporária privada/cifrada e mantém até 50 arquivos e 50 problemas na resposta.
- A deduplicação da importação fica numa tabela temporária PostgreSQL usando HMAC por execução. O serviço abre **uma única transação por pasta selecionada**. Erros de validação por linha permitem gravar as linhas válidas com estado `PARTIAL`; falha técnica ou cancelamento devolve erro e desfaz a transação da pasta.
- O streaming limita o uso de memória do processo, mas não limita o tamanho total da pasta nem a duração/volume da transação, das tabelas temporárias do banco e do WAL. Como o limite total de linhas foi definido como inexistente, a regra de atomicidade para pastas enormes precisa dizer se o lote é a pasta inteira ou se cada bloco pode ser confirmado e retomado separadamente. A pergunta foi enviada à usuária; não alterar esse comportamento até a resposta. Evidência: `internal/integrations/bling/receipt_service.go`, `internal/integrations/nuvempago/future_service.go`, `internal/integrations/csvstream/file_queue.go` e `internal/integrations/csvstream/source_ids.go`.
- Inventário estático de `http.Client.Do` e criação de requests no Go de produção encontrou consultas GET do Bling e o POST OAuth. Os GET passam pelo retry compartilhado, limitado a operações repetíveis e falhas transitórias; o POST OAuth não é repetido. Não foi encontrada outra chamada HTTP externa de produção neste inventário. Um futuro adapter de API deve adotar o padrão aprovado, classificando antes quais operações podem ser repetidas.
- Não houve alteração de código de produto, telas, permissões ou regras financeiras. Nenhum teste ou build foi executado nesta revisão; a constatação é estática e não mede crescimento real de WAL/temporários, consumo do banco, memória ou duração sob carga.

BK-024 permanece em 30% e o projeto mantém a estimativa registrada de 7,5% concluído: este complemento identifica uma decisão de arquitetura pendente, não fecha a validação operacional nem o mapa de riscos.

## Complemento — conferência de caches e outputs locais — 11/10/2026

- `artifacts/cache`, `artifacts/verification` e `frontend/dist` não existiam no início desta conferência. O `artifacts/retained` foi preservado porque contém a instância de PostgreSQL de desenvolvimento e seus arquivos de operação.
- O NSIS 3.13 foi restaurado de um arquivo já registrado para limpeza manual. O SHA-256 do ZIP coincide com o hash oficial anotado no registro local; `makensis.exe /VERSION` respondeu `v3.13`. A ferramenta fica em `artifacts/cache/nsis-3.13`, local central necessário para gerar o instalador. Go 1.26.9 também foi confirmado instalado no computador.
- Tentei remover somente três ZIPs antigos já classificados como dispensáveis no registro local (instalador Go redundante, cópia NSIS válida agora extraída e cópia NSIS com hash divergente). A revisão automática bloqueou a remoção (`blocked by policy`); nenhum desses arquivos foi removido e não tentei outro método. Eles continuam guardados em `artifacts/retained/manual-cleanup` para remoção manual.
- A Junction antiga em `docs/planejamento/artifacts/cache/backlog-tracker/node_modules` continua apontando para dependências compartilhadas do Codex e permanece registrada para limpeza manual; não foi seguida nem removida. Nenhuma saída de build foi criada nesta revisão.
- Não foram executados testes ou build do produto. A conferência de cache não alterou telas, dados de negócio ou código funcional.

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
  `main.go` incorpora `frontend/dist`, que não estava gerada nesta cópia. Por
  isso, o teste geral local parou na preparação do pacote principal. A [CI do
  commit e565c39](https://github.com/guilhermejacyzin/majucau/actions/runs/38138263712)
  preparou o bundle e passou em Go tests, Go vet, análise de vulnerabilidades,
  build frontend, builds do desktop/worker, consolidação de outputs e nos quatro
  jobs Windows/PostgreSQL/backlog. Os caches locais foram centralizados em
  `artifacts/cache` e removidos pelo módulo seguro após as verificações.
- Continuam pendentes ensaios controlados que provoquem falhas reais de limpeza,
  a auditoria mais ampla, VM limpa, validação de rede e testes de acesso por
  perfil.
- Nenhuma tela, regra financeira, permissão ou dado persistido foi alterado.

## Complemento — testes controlados de limpeza temporária

**Data:** 11/10/2026

**Tipo:** testes de falha e confirmação de evidência de CI.

- Os testes de `internal/securetemp` agora cobrem falha ao liberar memória
  nativa depois de uma operação bem-sucedida e o caso combinado em que a
  liberação e a remoção do temporário falham. A mensagem comum permanece
  sanitizada; quando a remoção falha, `securetemp.CleanupError` preserva o
  caminho para tratamento autorizado.
- `go test -count=1 ./internal/securetemp ./internal/integrations/csvstream`,
  `go vet` para os mesmos pacotes e `git diff --check` passaram localmente.
- A [CI #410](https://github.com/guilhermejacyzin/majucau/actions/runs/38139211957)
  passou nos jobs de backlog, Windows e PostgreSQL, incluindo testes Go, `go
  vet`, análise de vulnerabilidades, builds do desktop/worker, SBOM, smoke do
  instalador e consolidação dos outputs.
- A CI publicou o instalador temporário unsigned `majucau-g1-unsigned`, ID
  `11702500102`, SHA-256
  `bf93e59bd8ef372774c3cc175d4372dd05990774474c19245af5b08d8ce1d00a`, com
  expiração em `2026-10-12T12:32:30Z`. Como a conexão disponível não oferece
  exclusão de artifacts, os dados para remoção manual estão no registro local
  `artifacts/retained/cleanup-register.md`.
- Continuam pendentes outras injeções controladas de falha, fechar o mapa de
  riscos e validar instalação/atualização em VM limpa, rede e acesso por perfil.
- Nenhuma tela ou regra financeira foi alterada.

## Correção da remoção de arquivos intermediários CSV — 11/10/2026

- A revisão encontrou dois pontos da intercalação de nomes de arquivos em que a remoção direta propagava a falha, mas não preservava o tipo `securetemp.CleanupError` nem ocultava o caminho na mensagem comum.
- Os dois caminhos agora usam `securetemp.RemoveFile`, mantendo o caminho disponível para tratamento autorizado e a mensagem comum sanitizada. Os testes exercitam as junções em `add` e `finish`; também injetam uma falha de remoção em cada caminho e verificam o tipo do erro, o caminho protegido para suporte e a mensagem sem caminho. A suíte de `securetemp` cobre o comportamento do helper.
- `go test -count=1 ./internal/securetemp ./internal/integrations/csvstream` e `git diff --check` passaram localmente. O cache Go foi criado em `artifacts/cache` e removido pelo módulo `scripts/SafeCleanup.psm1`; a ausência da pasta foi conferida.
- A CI [38145917293](https://github.com/guilhermejacyzin/majucau/actions/runs/38145917293) passou em backlog, Windows e PostgreSQL para o commit `93a7ffb`. Os novos ensaios de falha serão verificados na CI do próximo commit.
- Não houve alteração de dados persistidos, telas ou regras financeiras. A simulação determinística não substitui falhas reais de disco/ACL nem validação em VM limpa.

## Complemento — exportações de relatórios e pacote de diagnóstico — 11/10/2026

**Tipo:** revisão estática das saídas atuais; nenhuma exportação de produto foi criada.

- A busca nas rotas Wails/IPC e no frontend não encontrou uma operação funcional de exportação de relatórios. O menu do mockup mantém a opção desativada. Os detalhes aprovados por tela, formatos, filtros e permissões continuam pendentes em L-22/Q-13; não alterar a tela nem criar uma saída por inferência.
- A exportação atualmente implementada no helper do instalador é um ZIP de suporte com somente dois itens: o contrato sanitizado de pré-teste e instruções. No uso atual da CLI, o contrato contém campos fixos de verificação e duas pastas configuradas; não lê logs, variáveis de ambiente, dados financeiros ou payloads externos.
- O helper escreve o ZIP num temporário ao lado do destino, fecha e sincroniza o arquivo, publica o destino com `os.Rename` e calcula o SHA-256 com `io.Copy`. Se a operação falhar, remove o temporário; a tentativa e seus limites constam em `internal/installer/diagnostics.go`.
- Os backups continuam sendo a saída de maior porte: V1 é lido em fluxo para compatibilidade e V2 cifra/decifra blocos autenticados, conforme `docs/evidence/BACKUP-STREAMING-IMPLEMENTATION-2026-10-03.md`. Essa revisão não repetiu o teste operacional de backup/restore.
- Permanecem pendentes o contrato dos relatórios por tela e a revisão operacional de ACL/limpeza em VM limpa. Nenhum teste ou build foi executado nesta revisão; BK-024 continua em 30%, sem mudança de telas ou regras financeiras.

## Revisão estática das chamadas de limpeza recursiva — 11/10/2026

**Tipo:** leitura de código e inventário das chamadas; nenhum teste foi executado.

- A busca por `os.RemoveAll` encontrou uma única chamada de produção, encapsulada por `securetemp.RemoveAll`. Os chamadores de produção revisados passam diretórios retornados por `securetemp.NewDir`: prévia CSV, criação de backup e restauração. Nenhum chamador de produção fornece diretamente um caminho digitado por usuário.
- `securetemp.RemoveAll` recusa caminho vazio, relativo ou raiz do volume e transforma falha em `securetemp.CleanupError`; a mensagem comum não revela o caminho. A implementação delega a remoção ao `os.RemoveAll` da biblioteca padrão Go.
- No Windows, `securetemp.NewWorkspace` cria pasta com DACL protegida para proprietário, SYSTEM e administradores. Nas outras plataformas, cria pasta temporária e restringe a permissão ao proprietário (`0700`). A ACL efetiva de uma máquina não foi medida nesta revisão.
- A busca por exclusão recursiva em PowerShell encontrou o `Remove-Item -Recurse` dentro de `scripts/SafeCleanup.psm1`. O módulo valida o destino esperado dentro do projeto, rejeita reparse points nos pais e no conteúdo, e interrompe se a remoção falhar.
- Para reduzir erro acidental de chamada, os usos de produção em backup, restauração e prévia CSV passaram a usar `securetemp.Workspace`, criado por `NewWorkspace`. A operação `Workspace.RemoveAll()` não recebe o caminho como argumento; o caminho fica privado no objeto criado pelo pacote.
- As funções públicas que recebiam o caminho da pasta como `string` foram removidas. O teste Windows de ACL também passou a limpar pelo objeto `Workspace`; a limpeza recursiva usa somente o caminho privado que o pacote guardou nesse objeto.
- A compilação `go build ./internal/securetemp ./internal/backup ./internal/integrations/csvstream` terminou com sucesso em Windows. Não foram executados testes. O cache dedicado em `artifacts/cache/workspace-type-guard` foi removido com `scripts/SafeCleanup.psm1`; os caches existentes permaneceram intactos.
- Depois da remoção das APIs públicas baseadas em caminho `string`, a mesma compilação terminou novamente com sucesso em Windows. O cache temporário `artifacts/cache/workspace-type-remove` foi removido com `SafeCleanup`; o `artifacts/cache/go-mod` existente foi reutilizado. Nenhum teste foi executado.
- Esta revisão não prova comportamento em falhas reais de disco/ACL, encerramento forçado ou ambiente instalado. Continuam necessários ensaios em VM e a verificação operacional da ACL.
- A mudança de API não altera dados persistidos, telas ou regras financeiras. BK-024 permanece em 30% até validação dos fluxos e controles operacionais.

## Proteção de ACL do cofre DPAPI no Windows — 11/10/2026

**Tipo:** implementação focalizada, teste Windows do pacote e `go vet`; sem instalação de serviço ou VM limpa.

- `NewDPAPIStore` exige caminho absoluto e deixa de criar a árvore inteira com `MkdirAll`. A pasta pai precisa existir e já ter DACL protegida contendo somente a identidade atual do processo, SYSTEM e Administradores. Se a condição não for atendida, o cofre recusa iniciar.
- O novo `dpapi_acl_windows.go` cria a pasta do cofre com DACL protegida, aplica e confere três permissões explícitas nas pastas/arquivos esperados e rejeita componentes que sejam reparse points. Arquivo de entropia, segredo existente e temporário da troca atômica também recebem a ACL restrita.
- `Put`, `Get` e `Delete` são serializados por instância. Buffers cifrados da gravação e leitura são sobrescritos após o uso; isso reduz a permanência desses bytes em memória, sem prometer limpeza de todas as cópias internas do runtime.
- `go test ./internal/security`, `go vet ./internal/security` e `git diff --check` passaram no Windows de desenvolvimento. O teste verificou a DACL da pasta do cofre, de `.entropy` e do arquivo de segredo. Não foi executado como `NT SERVICE\MajucauWorker`, nem em VM limpa ou pelo instalador.
- **Limitação para instalar:** o NSIS ainda não cria `ProgramData\Majucau` com ACL da identidade que executará o worker, não registra o serviço nem configura de forma segura a credencial de `MAJUCAU_DATABASE_URL`. Portanto, a implementação do cofre é fail-closed, mas ainda não está pronta para uso na hospedagem ou em uma instalação de produção.
- Nenhuma tela, regra financeira, fórmula ou dado persistido foi alterado. O progresso de BK-024 sobe apenas parcialmente; continuam bloqueantes o provisionamento pelo instalador, prova sob a identidade real do serviço, falhas controladas e validação em VM/rede/perfis.

## Divergência pendente sobre o escopo DPAPI — 11/10/2026

- O `ADR-001-ARQUITETURA-E-EMPACOTAMENTO.md` contém dois trechos diferentes: um diz que o worker protege dados no contexto da conta de serviço; outro pede DPAPI no escopo da máquina. A implementação em `internal/security/dpapi_windows.go` chama `CryptProtectData` com `dwFlags = 0`, e a migration `000001_init.up.sql` registra `DPAPI_CURRENT_USER`; `AGENTS.md` também vincula o cofre à identidade `NT SERVICE\MajucauWorker`.
- A documentação oficial da Microsoft explica que o modo padrão restringe a descriptografia à mesma conta, enquanto `CRYPTPROTECT_LOCAL_MACHINE` permite que qualquer usuário do mesmo computador descriptografe o conteúdo se conseguir obtê-lo. Fonte: [Microsoft Learn — CryptProtectData](https://learn.microsoft.com/pt-br/windows/win32/api/dpapi/nf-dpapi-cryptprotectdata).
- Foi solicitada à responsável a escolha do escopo. Até a resposta, o código e o schema permanecem sem alteração nesse ponto; não implementar o provisionamento final do cofre nem declarar SEC-R03 fechado.
- A pergunta é técnica e não muda nenhuma tela, regra financeira ou dado persistido.
