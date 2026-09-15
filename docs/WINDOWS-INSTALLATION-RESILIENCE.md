# Resiliência de instalação e operação em Windows

- **Produto:** Majucau Financial Intelligence
- **Plataforma alvo:** Windows 10/11 x64 em computador de terceiros, com um usuário operacional
- **Empacotamento:** instalador NSIS assinado; majucau.exe (UI Wails) e majucau-worker.exe (serviço Windows)
- **Persistência:** PostgreSQL local dedicado
- **Status:** especificação operacional para implementação e validação; este arquivo não afirma que todos os cenários já foram testados

Este documento define como o produto deve reagir a restrições, instalações anteriores, falhas de energia, antivírus, políticas corporativas, indisponibilidade de rede e erros de dependências. A meta é que o usuário instale, repare, atualize, recupere e remova o programa sem perder dados, sem expor segredos e sem receber uma falsa confirmação.

## 1. Escopo e linguagem de decisão

Cada operação de instalação deve verificar pré-condições, registrar o que alterou, ser idempotente, desfazer apenas o que é seguro e parar em estado compreensível quando a recuperação automática não for segura.

| Marcador | Significado | Comportamento |
|---|---|---|
| **D** | Detectável | detectar e registrar em diagnóstico sanitizado |
| **A** | Automático | corrigir ou repetir, com limite, timeout e backoff |
| **U** | Usuário/TI | orientar ação e permitir retomar pelo journal |
| **B** | Bloqueio seguro | não prosseguir, não destruir estado e preservar evidências |

Uma condição pode ter mais de um marcador. Não apresentar “instalado com sucesso” enquanto arquivos íntegros, serviço, PostgreSQL, migrations compatíveis, Named Pipe protegido, smoke test e atalho não estiverem confirmados.

## 2. Layout e fronteiras de confiança

Dados mutáveis nunca devem ficar em Program Files:

| Categoria | Caminho padrão | Conteúdo e regra |
|---|---|---|
| Binários | %ProgramFiles%\Majucau\ | executáveis e bibliotecas; somente instalador/admin |
| Dados | %ProgramData%\Majucau\ | estado técnico e configuração não secreta |
| Banco | %ProgramData%\Majucau\PostgreSQL\data\ | cluster local dedicado; nunca copiar ativo |
| Backups | %ProgramData%\Majucau\Backups\ | pacotes versionados e protegidos |
| Cofre | %ProgramData%\Majucau\Secrets\ | blobs DPAPI e entropy, ACL mínima |
| Logs | %ProgramData%\Majucau\Logs\ | rotacionados, sanitizados e com retenção |
| Journal | %ProgramData%\Majucau\Install\ | fases, hashes, códigos; sem segredos |

O usuário pode escolher outro volume para dados/backups, mas o fluxo deve rejeitar UNC, mídia removível, junction/symlink não confiável e diretório dentro de OneDrive, Dropbox, SharePoint ou serviço de sincronização. PostgreSQL em diretório sincronizado pode corromper o cluster e expor dados.

O manifesto deve registrar install_id, versão, schema, caminhos efetivos, serviço, porta local e hashes. Não gravar senha, token, client secret, DSN com senha ou payload de API.

## 3. Máquina de estados, lock e retomada

O instalador grava um journal em Install\state.json, com ACL administrativa, antes de cada mudança:

~~~text
NEW -> PREFLIGHT_OK -> PACKAGE_VERIFIED -> DEPENDENCIES_READY
    -> DATABASE_SERVICE_READY -> DATABASE_INITIALIZED -> MIGRATIONS_APPLIED
    -> WORKER_REGISTERED -> DESKTOP_INSTALLED -> SMOKE_TEST_OK -> COMMITTED

PREFLIGHT_BLOCKED | PACKAGE_INVALID | DEPENDENCY_BLOCKED | DATABASE_BLOCKED
MIGRATION_BLOCKED | SERVICE_BLOCKED | SMOKE_TEST_FAILED | RECOVERY_REQUIRED
~~~

O journal contém fase, timestamp, versão, hash, caminhos, serviço, porta, código e próximo passo. Ao retomar, não repetir cegamente migration, não apagar banco e não trocar identidade DPAPI.

Antes de ação destrutiva:

1. obter lock global nomeado, por exemplo Global\MajucauInstaller;
2. impedir duas execuções em paralelo;
3. parar UI e solicitar parada ordenada do worker;
4. criar backup validado se houver dados;
5. gravar checkpoint durável;
6. executar ação atomically;
7. confirmar resultado e somente então avançar o journal.

Cancelamento, crash ou queda de energia devem permitir continuar ou abrir Repair. Artefatos temporários têm nome único, ACL restrita e limpeza posterior; nunca remover diretório de dados inteiro para limpar uma instalação parcial.

## 4. Pré-flight obrigatório

| Verificação | D/A | Correção ou orientação | Bloqueio |
|---|---|---|---|
| Windows, build e arquitetura | D/B | informar versão suportada | x86, ARM sem pacote compatível ou build abaixo do mínimo |
| Token elevado e UAC | D/U/B | solicitar elevação; explicar motivo | política impede elevação |
| Espaço em binário, dados, temp e backup | D/U/B | liberar espaço ou escolher volume | sem margem para rollback |
| NTFS, caminho absoluto e comprimento | D/U/B | sugerir caminho local padrão | path inválido, UNC, reparse point ou mídia removível |
| Outra versão instalada | D/A/U | continuar, reparar ou atualizar | manifesto/hash incoerente |
| Reboot pendente | D/U | pausar e retomar após reinício | operação não pode ser retomada |
| EDR/Defender/Controlled Folder Access | D/U/B | orientar TI com hash/assinatura | arquivo essencial foi quarentenado |
| WebView2 e arquitetura | D/A/U/B | usar runtime offline ou instalar oficialmente | runtime ausente e sem pacote offline |
| Serviços, processos e portas | D/A/U/B | parar somente processos próprios | conflito não pode ser alterado com segurança |
| PostgreSQL existente | D/U/B | identificar somente instância Majucau | nunca assumir posse de cluster de terceiros |
| Rede, proxy e TLS | D/U/B | instalar offline ou orientar rede | download obrigatório não verificável |
| Assinatura, hash e manifesto | D/A/B | redownload oficial uma vez | assinatura inválida ou pacote adulterado |
| Relógio do sistema | D/U/B | corrigir horário com TI | TLS/OAuth não validável |

O setup deve mostrar requisito quantitativo de espaço, o que será preservado e quais fases podem requerer reboot. Não desativar UAC, antivírus, SmartScreen, firewall ou políticas para forçar a instalação.

## 5. Cenários de terceiros

### 5.1 Arquitetura, políticas e UAC

- Instalar somente no Windows x64 declarado como suportado; “Windows” sem build não basta.
- Não instalar x64 em Windows 32-bit nem aceitar incompatibilidade como aviso.
- Se política corporativa bloquear serviço, elevação, execução ou ProgramData, informar a TI com código específico.
- Usuário padrão pode iniciar o setup e elevar por credencial administrativa. A senha do administrador nunca vai para log.
- O usuário autorizado deve ser identificado por SID, não pelo nome visível.
- Se outro usuário abrir a UI, mostrar não autorizado e oferecer fluxo administrativo; não abrir o cofre por inferência.
- Não substituir conta virtual do worker por LocalSystem sem decisão explícita e registro de risco.

### 5.2 Paths, Unicode e roaming

Validar acentos, nomes CJK, path longo, perfil móvel, Documents/Desktop em OneDrive, volume quase cheio, NTFS/FAT32/exFAT e junctions. Aceitar caminho local NTFS normalizado; rejeitar nome terminado em espaço/ponto, caminho não gravável e reparse point fora do escopo.

Atalhos podem acompanhar o usuário, mas banco, secrets, logs e journal devem permanecer em locais estáveis do computador. O diretório de dados não deve ser sincronizado nem compartilhado.

### 5.3 Antivírus, EDR e SmartScreen

Mostrar editor, assinatura e hash; orientar a TI a permitir o pacote assinado; nunca orientar desativação da proteção. Se arquivo essencial desaparecer, comparar o manifesto, registrar AV_QUARANTINED_FILE, bloquear e oferecer Repair. Não baixar outro binário silenciosamente.

SmartScreen pode alertar mesmo para assinatura válida e baixa reputação inicial. O UX deve explicar confirmação pelo usuário/TI, sem bypass.

### 5.4 WebView2 online/offline

Preferir runtime Evergreen offline embutido no release, com assinatura e hash. Online só com permissão:

| Condição | Resultado |
|---|---|
| runtime compatível | continuar e registrar versão |
| offline + bootstrapper offline | instalar e continuar |
| offline + somente bootstrapper online | pausar com WEBVIEW2_OFFLINE_REQUIRED |
| proxy com autenticação | orientar configuração; não capturar senha |
| runtime corrompido | oferecer repair oficial |
| arquitetura incompatível | bloquear antes da UI |
| runtime de outra aplicação | não remover nem fazer downgrade |

Se WebView2 falhar no primeiro uso, informar que dados não foram apagados. Worker e banco não devem ser removidos por falha de renderização.

### 5.5 NSIS interrompido ou duplicado

Segunda instância deve detectar lock e mostrar o status da primeira. Cancelar antes de DATABASE_INITIALIZED remove apenas artefatos próprios incompletos; depois disso preserva banco e oferece continuar/Repair. Queda de energia lê journal no próximo setup. Nunca usar remoção ampla sobre ProgramData.

## 6. PostgreSQL isolado e readiness

Uma instalação PostgreSQL de terceiros não pode ser reutilizada por inferência de versão, porta ou nome. Identificar a instância somente por marcador/cluster/serviço conhecido do Majucau.

O setup deve:

1. escolher porta local na faixa privada do produto;
2. testar bind e considerar corrida entre teste e bind real;
3. registrar porta sem senha;
4. configurar listen_addresses=127.0.0.1, SCRAM e pg_hba mínimo;
5. não abrir acesso de rede;
6. criar serviço e dependência real do worker;
7. aguardar readiness por polling com timeout;
8. confirmar que responde no cluster/database esperados.

Porta ocupada gera DB_PORT_CONFLICT. Repetir escolha dentro de limite; não matar processo, não mudar porta de instalação alheia e não abrir firewall público. Sem porta segura, bloquear.

Senhas iniciais são aleatórias e não aparecem em linha de comando/janela/log. O worker usa role runtime sem superuser; migrations usam identidade de instalação; client secrets ficam no DPAPI; ACLs são explicitamente aplicadas.

“Serviço iniciado” não significa “banco pronto”. Health deve distinguir:

~~~text
SERVICE_STOPPED
SERVICE_START_PENDING
DATABASE_NOT_REACHABLE
DATABASE_AUTH_FAILED
DATABASE_WRONG_CLUSTER
DATABASE_READY
MIGRATION_PENDING
WORKER_READY
WORKER_DEGRADED
~~~

Em timeout, verificar serviço, processo, porta, permissão e log; reiniciar somente serviço próprio uma vez. Depois, parar com DB_READINESS_TIMEOUT, preservar cluster e oferecer retry/diagnóstico.

## 7. Migrations e worker

Migrations são forward-only, checksumadas, executadas sob advisory lock e somente depois de backup validado. Verificar versão do schema, faixa suportada pelo binário, checksums já aplicados, espaço e ausência de execução concorrente.

| Situação | Conduta |
|---|---|
| banco novo | criar roles/DB e aplicar tudo |
| versão esperada | validar checksum, sem repetir |
| pendência compatível | backup, lock, migration e validação |
| checksum divergente | bloquear; nunca apagar histórico |
| schema mais novo | bloquear downgrade |
| backup não validado | não fazer alteração destrutiva |
| migration falha | rollback transacional quando possível; senão Recovery/restore |
| corrupção/constraints quebradas | preservar evidência; não “resetar” automaticamente |

O worker só fica pronto após migration, health IPC e smoke test. No boot, validar assinatura/versão, SID, ACL, PostgreSQL, Named Pipe e protocolo. Se cair repetidamente, limitar reinícios, registrar SERVICE_CRASH_LOOP, preservar último dado e mostrar modo degradado.

A UI deve distinguir pipe inexistente, acesso negado, versão incompatível, timeout, resposta inválida e worker em atualização. Ação material sempre é autorizada no worker e auditada; falha nunca zera indicador.

## 8. Rede, proxy, TLS e first run

A instalação base deve funcionar offline quando o pacote contém os binários e WebView2 necessários. APIs externas não são pré-condição para abrir o dashboard. O first run configura integrações depois.

- OAuth abre navegador externo; login nunca é capturado em WebView.
- Usar proxy documentado; não pedir senha de proxy no formulário da integração.
- Validar cadeia TLS, hostname e relógio; nunca aceitar certificado inválido.
- Detectar captive portal pela resposta inesperada; não interpretar HTML como token.
- Aplicar timeout, cancelamento, paginação, backoff e retry apenas transitório.
- Preservar último dado válido e marcar STALE/PARTIALLY_AVAILABLE.

| Cenário | Conduta |
|---|---|
| sem internet | abrir local; integrações NOT_CONFIGURED |
| proxy autenticado | orientar TI; nenhum segredo no log |
| captive portal | pedir navegador/rede correta |
| DNS indisponível | retry limitado e último estado |
| TLS inválido | bloquear chamada, corrigir relógio/rede |
| HTTP 429 | respeitar Retry-After |
| 5xx/timeout | retry com jitter, depois STALE |
| token expirado | AUTH_ERROR e reconexão no navegador |
| falha parcial | não zerar card nem misturar fonte |

## 9. Crash, energia e arquivos bloqueados

Journal, configuração, secrets e manifestos devem ser escritos em arquivo temporário no mesmo volume, sincronizados e substituídos atomically por API Windows. Nunca truncar o arquivo ativo.

Após queda de energia:

1. ler journal;
2. esperar recovery do PostgreSQL;
3. validar readiness;
4. retomar migration somente se o estado permitir;
5. retomar jobs por checkpoint/idempotência;
6. marcar lote interrompido para reprocessamento seguro.

Se arquivo estiver bloqueado por UI, worker, EDR ou indexador: parar processos próprios, esperar timeout, identificar processo quando possível, pedir fechamento e tentar novamente. Não forçar processo desconhecido nem apagar parcialmente dados.

## 10. Upgrade, downgrade e rollback

Upgrade normal: verificar assinatura/hash/compatibilidade; detectar processos; criar e validar backup; pausar jobs; parar UI/worker; trocar arquivos; iniciar PostgreSQL/worker; aplicar migration; smoke test; só então reabrir UI.

Se falhar, permanecer em RECOVERY_REQUIRED. Não instruir abrir binário antigo contra schema incompatível.

Downgrade isolado é bloqueado quando schema está acima do limite. O procedimento seguro restaura backup pré-upgrade e o conjunto compatível de binário, schema, roles e configuração. Migration irreversível não recebe promessa de rollback automático.

Cada release publica min_schema_version e max_schema_version. Reboot durante upgrade abre Repair, restaura último conjunto íntegro de binários e nunca restaura banco automaticamente sem critério.

## 11. Backup, restore e corrupção

Pacote de backup deve conter dump PostgreSQL, globals necessários sem senhas, manifesto com versão/schema/origem/data/hash, criptografia autenticada e indicação de que tokens não são exportados por padrão. Verificar que o dump pode ser lido antes de update.

Restore:

1. backup do estado atual se legível;
2. verificar assinatura/hash e compatibilidade;
3. parar worker;
4. restaurar cluster controlado;
5. aplicar somente migrations permitidas;
6. validar constraints, RAW, snapshots, auditoria e continuidade;
7. iniciar em RESTORED_NEEDS_RECONNECT;
8. pedir reconexão dos provedores.

Restore em máquina diferente não pode reutilizar tokens DPAPI; status deve ser AUTH_REQUIRED.

Falha de leitura, checksum ou constraint exige modo somente diagnóstico, cópia preservada quando possível e restore. Não executar DELETE, VACUUM FULL, recriação de tabelas ou reset sem confirmação administrativa e backup.

## 12. Repair, reinstall e uninstall

Repair deve permitir separadamente recuperar arquivos assinados, atalhos, serviço/dependência/ACL, WebView2, PostgreSQL/migrations e smoke test. Preservar banco, backups, secrets e logs por padrão. Reset de dados é outro fluxo, com backup e confirmação reforçada.

Reinstall da mesma versão preserva install_id, serviço, SID, banco, backups e cofre. Se DPAPI falhar após mudança de identidade, manter dados e marcar integrações para reconexão; não apagar cofre automaticamente.

Uninstall padrão fecha UI, para worker, remove serviço/atalhos/binários e preserva banco, backups, logs úteis e secrets, informando caminhos. Perguntar separadamente:

1. remover programa mantendo dados;
2. também excluir banco, backups, logs e credenciais protegidas.

Exclusão completa requer confirmação reforçada, resumo do que será perdido e backup recomendado/executado. Arquivo bloqueado gera cleanup pending; não declarar sucesso total nem apagar ProgramData à força.

## 13. Diagnóstico e códigos estáveis

“Exportar diagnóstico” inclui versão, build, arquitetura, WebView2/PostgreSQL, estado do serviço/pipe/porta sem credencial, journal, últimos erros sanitizados, hashes/assinaturas, espaço e status de integrações. Redigir usuário, e-mail, documento, nome de loja, URL com query sensível, token, secret, DSN e payload RAW. Exibir prévia de arquivos/tamanho antes de exportar.

| Código | Significado | Próxima ação |
|---|---|---|
| INSTALL_NOT_ELEVATED | elevação não obtida | executar como administrador |
| OS_UNSUPPORTED | Windows/arquitetura fora do suporte | usar máquina compatível |
| DISK_SPACE_LOW | espaço insuficiente | liberar ou escolher volume |
| PACKAGE_SIGNATURE_INVALID | assinatura/hash inválido | obter pacote oficial |
| PACKAGE_FILE_MISSING | EDR/arquivo ausente | TI/Repair |
| REBOOT_PENDING | reboot impede fase | reiniciar e continuar |
| WEBVIEW2_OFFLINE_REQUIRED | runtime não disponível | instalar pré-requisito |
| WEBVIEW2_MISSING | runtime não detectado no preflight | instalar runtime oficial/offline |
| PATH_UNSUPPORTED | caminho inseguro | escolher NTFS local |
| PATH_UNAVAILABLE | destino ou ancestral não acessível | corrigir caminho/permissão |
| DB_PORT_CONFLICT | porta ocupada | corrigir/retry |
| DB_SERVICE_FAILED | PostgreSQL não inicia | diagnóstico/TI |
| DB_READINESS_TIMEOUT | não ficou pronto | retry/Repair |
| DB_WRONG_CLUSTER | resposta de outro cluster | bloquear |
| MIGRATION_CHECKSUM_MISMATCH | histórico alterado | restore íntegro |
| MIGRATION_INCOMPATIBLE | schema incompatível | release/restore correto |
| SERVICE_ACCESS_DENIED | ACL/SID incorreto | Repair administrativo |
| SERVICE_CRASH_LOOP | worker cai repetidamente | suporte/modo degradado |
| PIPE_ACCESS_DENIED | usuário não autorizado | corrigir SID/DACL |
| DPAPI_DECRYPT_FAILED | cofre não abre | reconectar; preservar dados |
| AV_QUARANTINED_FILE | EDR removeu arquivo | TI com hash |
| NETWORK_PROXY_REQUIRED | proxy necessário | configurar rede |
| TLS_VALIDATION_FAILED | certificado/relógio inválido | corrigir; não ignorar TLS |
| BACKUP_VALIDATION_FAILED | backup não confiável | novo backup |
| RESTORE_INCOMPATIBLE | pacote incompatível | backup/release correto |
| LOCKED_FILE | arquivo em uso | fechar/retry |
| RECOVERY_REQUIRED | fase incompleta | abrir Repair |

Mensagem pública deve ser simples e não revelar SQL, stack trace, segredo ou caminho sensível. Correlation ID liga UX ao log sanitizado.

## 14. UX e operação

Instalação, Repair, Restore e Uninstall mostram fase atual, fases concluídas, progresso honesto, ação, Ver detalhes, Retry quando idempotente, Cancelar quando seguro, preservação de dados, reboot, código e correlation ID.

Exemplos:

- “O banco local já existe. Vamos preservá-lo e verificar se pertence ao Majucau.”
- “O Windows bloqueou um arquivo. Não desligue a proteção; peça à TI para liberar este pacote assinado.”
- “A internet não é necessária para abrir o programa. Integrações podem ser configuradas depois.”
- “A atualização não terminou. Seus dados foram preservados; abra Reparar para continuar.”
- “As credenciais não puderam ser abertas nesta conta. Seus dados continuam no computador; reconecte as integrações.”

Tooltips explicam schema, readiness, backup e repair em linguagem simples, sem esconder informação crítica.

## 15. Matriz mínima de VMs

São cenários planejados; a presença nesta lista não é evidência de execução.

| ID | Cenário | Aceite |
|---|---|---|
| VM-01 | Win10 x64 limpo, internet | instalação, serviço, WebView2, smoke, first run |
| VM-02 | Win11 x64 limpo, internet | mesmo fluxo e atalhos |
| VM-03 | versão mínima, sem internet | offline ou bloqueio explícito |
| VM-04 | sem WebView2, pacote offline | runtime e UI |
| VM-05 | WebView2 corrompido | repair/bloqueio |
| VM-06 | usuário padrão + admin | UAC, SID e serviço |
| VM-07 | UAC recusado | nada removido |
| VM-08 | espaço baixo | pré-flight antes de mudança |
| VM-09 | Unicode/path redirecionado | dados em caminho seguro |
| VM-10 | OneDrive em perfil | cluster fora da sincronização |
| VM-11 | PostgreSQL/porta existentes | não reutilizar; detectar |
| VM-12 | reboot pendente | pausa e retomada |
| VM-13 | EDR/quarentena | código, diagnóstico, Repair |
| VM-14 | SmartScreen | UX de assinatura, sem bypass |
| VM-15 | cancelamento em cada fase | journal e preservação |
| VM-16 | energia durante migration | recovery e estado seguro |
| VM-17 | arquivo bloqueado | parada/retry |
| VM-18 | ACL/worker incorretos | pipe negado e Repair |
| VM-19 | upgrade com backup | migration e dados |
| VM-20 | upgrade interrompido | recovery/rollback compatível |
| VM-21 | downgrade com schema novo | bloqueio sem alteração |
| VM-22 | restore mesmo computador | RAW/auditoria/continuidade |
| VM-23 | restore outra máquina | AUTH_REQUIRED |
| VM-24 | backup corrompido | não sobrescrever atual |
| VM-25 | Repair | dados e secrets preservados |
| VM-26 | reinstall mesma versão | install_id/SID/cofre preservados |
| VM-27 | uninstall mantendo dados | reinstalação lê dados |
| VM-28 | uninstall completo | confirmação e remoção verificável |
| VM-29 | proxy/captive/TLS inválido | sem vazamento e sem bypass |
| VM-30 | crash loop worker | limite, log, último dado |

Critérios comuns:

1. nenhum dado é perdido sem confirmação explícita;
2. desconhecido não vira zero, sucesso ou sincronizado;
3. toda falha tem código, orientação e journal;
4. nenhum segredo aparece em log, linha de comando, diagnóstico ou backup comum;
5. banco só é acessado pelo worker;
6. upgrade, Repair e uninstall preservam dados por padrão;
7. assinatura, manifesto e hash são verificáveis;
8. usuário sem conhecimento técnico entende a ação principal;
9. suporte reproduz por VM, versão e correlation ID;
10. cenário não executado permanece “não verificado”.

## 16. Checklist de release e atendimento

- [ ] Windows/build/arquitetura suportados publicados
- [ ] executáveis e instalador assinados com timestamp
- [ ] hash e manifesto publicados
- [ ] WebView2 online/offline definidos
- [ ] PostgreSQL dedicado, porta e serviço documentados
- [ ] DPAPI, identidade, ACL e Named Pipe validados
- [ ] migrations checksumadas e faixa de compatibilidade publicada
- [ ] backup pré-update e restore real executados
- [ ] Repair, reinstall e uninstall preservando dados validados
- [ ] logs sanitizados e códigos revisados
- [ ] secret scan, dependências, SBOM e licenças aprovados
- [ ] smoke test de serviço e pipe aprovado
- [ ] matriz de VMs preenchida com evidência
- [ ] limitações e bloqueios externos comunicados
- [ ] nenhum gate marcado como produção se assinatura, VM limpa, restore ou aceite funcional estiver pendente

Atendimento: pedir código/correlation ID, confirmar Windows/arquitetura/privilégio, coletar diagnóstico sanitizado, ler fase do journal, escolher retry/Repair/Restore/Uninstall conforme o estado e nunca pedir token/senha nem apagar ProgramData como primeira tentativa. Encerrar somente após confirmar serviço, banco, UI, último dado válido e status explícito das integrações.
