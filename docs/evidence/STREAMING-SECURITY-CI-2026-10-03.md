# Auditoria incremental de limites e streaming — 2026-10-03

## Escopo validado

- A resposta JSON financeira do Bling é lida com `json.Decoder` sobre `io.LimitedReader`. O corpo tem teto de 4 MiB; a lista mantém no máximo 100 `json.RawMessage` por página e rejeita página acima desse contrato, truncamento, JSON adicional e bytes além do limite. Os campos RAW por registro continuam preservados para persistência idempotente.
- O cofre Windows DPAPI lê o arquivo de entropia em fluxo até 33 bytes e aceita apenas o formato exato de 32 bytes. Valores armazenados ficam limitados a 1 MiB, coerente com o quadro IPC; blobs cifrados aceitam até 64 KiB adicionais para o envelope DPAPI e falham fechados acima disso.
- A decisão aprovada para backup permanece: manter leitura/restauração V1 e usar V2 em blocos autenticados para novos arquivos, conforme `G0-DECISION-REGISTER.md` e `BACKUP-STREAMING-IMPLEMENTATION-2026-10-03.md`.

## Evidência de CI

- DPAPI: commit `e2157bf7fccef6c60408239eb03a53a4aea69f98`; [run 37108239734](https://github.com/guilhermejacyzin/majucau/actions/runs/37108239734) passou em Windows e PostgreSQL, incluindo testes para entropia, valores e blobs acima dos limites.
- Bling: commit `83952a6c60f1e3b2abf531025b0fedc0a0406f49`; [run 37108913837](https://github.com/guilhermejacyzin/majucau/actions/runs/37108913837) passou em Windows e PostgreSQL. A suíte validou arrays com 101 registros e corpo acima do limite após JSON válido; testes Go, `go vet`, verificação de vulnerabilidades/secrets, frontend, builds, smoke do helper/instalador e consolidação dos outputs também concluíram com sucesso.
- Outputs locais: `artifacts/cache/` e `artifacts/verification/` não existiam. O pipeline centralizou os outputs do runner em `artifacts/`. Como a conexão GitHub não oferece exclusão de artifacts, seus IDs, hashes, expiração e instrução manual constam em `artifacts/retained/cleanup-register.md`.

## Limites desta evidência

- A página da API fica limitada a 100 registros e 4 MiB; não se mantém o histórico inteiro em memória. Os `json.RawMessage` de uma única página ainda ficam retidos juntos para gravar o RAW transacionalmente.
- O DPAPI é uma API de sistema baseada em buffer; o código agora limita estritamente o bloco antes de chamá-la. Ainda faltam validar ACL/SDDL e recuperação de credenciais em instalação Windows limpa.
- Esta auditoria cobre os fluxos citados e não certifica todos os formatos, integrações nem a operação completa. DATA-04 e SEC-01 permanecem `PARTIAL`; resta auditar os demais caminhos e executar os gates em VMs reais.

## Temporários privados — commit `d17b66f`

- A prévia CSV e os fluxos de backup/restauração compartilham agora `internal/securetemp`: no Windows a pasta nasce com DACL protegida e herança restrita aos arquivos; nos demais sistemas recebe permissão de proprietário (`0700`).
- A CI [37116232138](https://github.com/guilhermejacyzin/majucau/actions/runs/37116232138) passou nos jobs Windows e PostgreSQL e executou os testes existentes do projeto.
- A CI valida compilação e regressões dos fluxos que usam o helper, mas não verifica as ACEs efetivas numa VM Windows limpa. A prova de ACL/limpeza operacional continua pendente em VM Windows 10/11; DATA-04 e SEC-01 permanecem `PARTIAL`.

## Manifesto de pacote — commits `3b43388`, `27c990b` e `2f4e022`

- O manifesto limita a 4.096 entradas, 1.024 bytes por caminho e 128 bytes para cada versão de schema; o parser rejeita componentes numéricos que excedam o tipo inteiro.
- A leitura do manifesto e dos arquivos referenciados usa `os.Root`, rejeita links simbólicos e componentes que não sejam diretórios, e compara o arquivo aberto com o item previamente inspecionado.
- A CI [37117232676](https://github.com/guilhermejacyzin/majucau/actions/runs/37117232676) passou nos jobs Windows e PostgreSQL. Incluiu testes Go, `go vet`, verificações de vulnerabilidades e segredos, lint/typecheck/test/build do frontend, builds do desktop/worker/helper, smoke de instalação e desinstalação e consolidação dos outputs.
- A CI confirma os testes/builds no runner, não substitui a validação de instalação, links simbólicos e permissões numa VM limpa Windows 10/11. OPS-04, DATA-04 e SEC-01 permanecem `PARTIAL` onde essas evidências operacionais ainda faltam.

## Snapshot do painel e limite IPC — auditoria estática

- A consulta do painel retorna métricas agregadas e no máximo 50 linhas de recebíveis mais 50 de contas a pagar. Os nomes em `contacts.name` e os campos `payables.document`/`payables.category` são `text` sem limite de bytes no schema atual.
- O PostgreSQL agrega as linhas em JSON; o worker lê e decodifica esse JSON, `NewResponse` o serializa e o servidor serializa a resposta novamente. O limite de 1 MiB do IPC só é aplicado ao frame final, depois dessas alocações.
- Portanto, a quantidade de linhas é limitada, mas o tamanho da resposta em memória ainda depende do conteúdo textual armazenado. Um payload grande falha ao ultrapassar o frame, porém pode consumir memória antes da rejeição.
- Nenhuma regra financeira ou tela foi alterada. Foi solicitada decisão sobre carregar os detalhes em partes mantendo a mesma aparência, ou definir limites por campo e rejeitar entradas acima deles. DATA-04 permanece `PARTIAL` até fechar esse contrato e validar o fluxo.

## Propagação de falhas de limpeza temporária — CI em 2026-10-03

- `internal/securetemp.CleanupError` conserva o caminho para operador autorizado e omite esse caminho da mensagem comum. A remoção abrangente recusa caminhos relativos e a raiz do volume.
- Backup/restauração, writes DPAPI, bundles de diagnóstico, journal do instalador e preview CSV propagam falhas de limpeza; o pacote V2 incompleto é montado dentro do workspace privado antes do rename final.
- A CI [37123188191](https://github.com/guilhermejacyzin/majucau/actions/runs/37123188191) passou nos jobs Windows e PostgreSQL, incluindo Go tests/vet, verificações de segurança, build e smoke do instalador.
- A execução não força uma falha de permissão/remoção e não substitui ensaio de DACL e limpeza numa VM Windows 10/11 limpa. `OPS-02`, `SEC-01` e `DATA-04` permanecem `PARTIAL` até as evidências operacionais correspondentes.

## DACL efetiva de temporários privados — CI Windows em 2026-10-03

- O teste `TestNewDirAppliesProtectedDACLAndRestrictsInheritedAccess` lê o descritor real do diretório criado por `securetemp.NewDir`, confirma DACL protegida e exatamente três ACEs `FA` para Owner Rights, LocalSystem e Administrators.
- O teste também cria um arquivo filho e verifica que as permissões esperadas foram herdadas; permissões mais amplas ou ausentes fazem o teste falhar.
- A CI [37132621657](https://github.com/guilhermejacyzin/majucau/actions/runs/37132621657) passou nos jobs Windows e PostgreSQL, incluindo testes Go, vet, scans de segurança, build e smoke de instalação/desinstalação.
- Esta prova cobre a ACL efetiva no runner Windows. Não substitui a validação da instalação e limpeza numa VM Windows 10/11 limpa nem verifica toda a configuração operacional de DPAPI/serviço; `SEC-01` e `DATA-04` continuam `PARTIAL`.

## DACL efetiva do Named Pipe — CI Windows em 2026-10-03

- `TestNamedPipeAppliesProtectedDACLToCreatedPipe` lê o descritor do pipe criado, usando [`GetSecurityInfo`](https://learn.microsoft.com/en-us/windows/win32/api/aclapi/nf-aclapi-getsecurityinfo), que a documentação do Windows indica para recuperar a segurança de named pipes.
- O teste valida DACL protegida e exatamente três ACEs de acesso total para o SID explicitamente autorizado, BUILTIN\Administrators e LocalSystem. Compara os SIDs binários e a máscara `FILE_ALL_ACCESS` efetiva após a normalização de `GENERIC_ALL`; rejeita ACE extra, tipo ou máscara diferentes e flags de herança.
- A CI [37134721442](https://github.com/guilhermejacyzin/majucau/actions/runs/37134721442) passou em Windows e PostgreSQL. No Windows passaram os testes Go, vet, govulncheck, secret scan, verificações do frontend, builds desktop/worker/instalador, smoke de instalação/desinstalação e consolidação de outputs.
- Esta prova confirma a DACL efetiva no runner Windows Server 2025. Ainda faltam validar o processo cliente autorizado separado, UI fechada, reboot, concorrência e instalação em VM Windows 10/11 limpa; `ARC-02` e `SEC-01` permanecem `PARTIAL`.

## Direitos mínimos e disponibilidade do Named Pipe — CI Windows em 2026-10-03

- O cliente IPC recebe somente `FILE_READ_DATA | FILE_WRITE_DATA | SYNCHRONIZE` (`0x00100003`); não pode criar instâncias do pipe. O SID do serviço recebe esses direitos mais `FILE_CREATE_PIPE_INSTANCE` (`0x00100007`), necessários para o worker criar a próxima instância. BUILTIN\Administrators e LocalSystem mantêm acesso total para administração e operação do serviço.
- `TestNamedPipeAppliesProtectedDACLToCreatedPipe` lê o descritor efetivo com `GetSecurityInfo` e compara SIDs binários, máscaras e quantidade exata de ACEs. Assim, verifica a DACL protegida e rejeita principal, permissão ou ACE extra inesperada.
- `TestNamedPipeHealthRoundTripAndCancellation` executa chamadas de saúde em sequência. O cliente aguarda 10 ms e tenta abrir novamente somente quando o pipe ainda está ocupado ou a instância seguinte ainda não foi criada; a espera respeita cancelamento. O retry ocorre antes de enviar o frame da solicitação, então não repete uma operação de negócio.
- A CI [37136211472](https://github.com/guilhermejacyzin/majucau/actions/runs/37136211472) passou nos jobs Windows e PostgreSQL, incluindo testes Go, vet, govulncheck, secret scan, frontend, builds desktop/worker/instalador, smoke de instalação/desinstalação e consolidação de outputs.
- O runner é Windows Server 2025. Ainda faltam usuário autorizado em processo separado, UI fechada, reboot, ensaio de concorrência e instalação em VM Windows 10/11 limpa; `ARC-02` e `SEC-01` permanecem `PARTIAL`.

## Cliente autorizado em processo separado — CI Windows em 2026-10-03

- `TestNamedPipeAcceptsAuthorizedClientProcess` inicia uma cópia separada do executável de teste. O servidor consulta o PID conectado, lê o SID do token desse processo e aceita somente o SID de teste explicitamente autorizado; o filho completa uma chamada de saúde pelo Named Pipe.
- O teste também cancela o servidor e confirma seu encerramento após a chamada. A execução separada valida o caminho de autenticação por PID/token, enquanto o teste da DACL verifica em separado os SIDs e máscaras de menor privilégio.
- O filho e o servidor usam o mesmo SID de teste no runner. Isso não substitui o ensaio do par instalado com UI e serviço sob identidades distintas nem verifica a conta `NT SERVICE\MajucauWorker` e seu ciclo após reboot.
- A CI [37137961615](https://github.com/guilhermejacyzin/majucau/actions/runs/37137961615) passou nos jobs Windows e PostgreSQL, incluindo Go tests, vet, govulncheck, secret scan, frontend, builds, smoke do instalador e consolidação de outputs.
- O runner é Windows Server 2025. Ainda faltam UI fechada, reboot, concorrência e validação em VMs limpas Windows 10/11; `ARC-02` e `SEC-01` permanecem `PARTIAL`.

## Clientes simultâneos e fechamento de instância — CI Windows em 2026-10-03

- O primeiro ensaio de oito clientes independentes falhou na CI [37139225972](https://github.com/guilhermejacyzin/majucau/actions/runs/37139225972): `concurrent-03` recebeu EOF antes do frame de resposta. A inspeção apontou uma corrida possível: o servidor iniciava a desconexão/fechamento em segundo plano e podia criar a próxima instância antes de o handle anterior terminar de fechar.
- A documentação oficial do Windows descreve que clientes recebem `ERROR_PIPE_BUSY` quando todas as instâncias estão ocupadas e que `DisconnectNamedPipe` descarta dados ainda não lidos; uma instância precisa ser desconectada antes de ser reutilizada ([cliente](https://learn.microsoft.com/en-us/windows/win32/ipc/named-pipe-client), [DisconnectNamedPipe](https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-disconnectnamedpipe)).
- O commit `c50d134` dá ao objeto de instância a propriedade explícita do `os.File`, torna o fechamento idempotente e, no caminho normal, conclui desconexão e fechamento antes de publicar a instância seguinte. Isso evita tanto a janela observada de EOF quanto um finalizer fechar depois um handle já reutilizado.
- `TestNamedPipeConcurrentClientHealthRequests` libera oito clientes independentes ao mesmo tempo, confere cada `request_id` e payload, rejeita resposta duplicada e impõe prazo para conclusão/cancelamento. O servidor ainda processa um handler por vez; esta evidência prova que as chamadas concorrentes terminam sem perder ou misturar respostas nesse cenário.
- A CI [37139715109](https://github.com/guilhermejacyzin/majucau/actions/runs/37139715109) passou nos jobs Windows e PostgreSQL, incluindo testes Go, vet, scans de segurança, frontend, builds, smoke do instalador e consolidação de outputs.
- Isto não substitui concorrência entre operações do worker, ensaio do serviço instalado com UI fechada, reboot ou VMs limpas Windows 10/11. `ARC-02` e `SEC-01` continuam `PARTIAL`.

## Estado atual da CI Windows — 2026-10-04

- As execuções dos commits `86b275a` ([CI 37235671944](https://github.com/guilhermejacyzin/majucau/actions/runs/37235671944)), `667eb5a` ([CI 37236624188](https://github.com/guilhermejacyzin/majucau/actions/runs/37236624188)) e `3ad1e2a` ([CI 37237958548](https://github.com/guilhermejacyzin/majucau/actions/runs/37237958548)) falharam no job Windows, na etapa `Go tests`; os jobs PostgreSQL passaram.
- O commit `3ad1e2a` só alterou documentação. A repetição confirma que o conjunto Windows ainda não está verde, mas não identifica a causa.
- A consulta disponível mostrou apenas o estado do job e uma anotação genérica de saída 1. Os logs detalhados não foram acessíveis nesta execução; por isso, não atribuir a falha a um teste específico nem afrouxar controles de segurança com base nisso.
- Próximo passo: obter a saída detalhada dos testes Windows, corrigir a causa e voltar a exigir sucesso nos dois jobs antes de marcar `ARC-02` concluído.
