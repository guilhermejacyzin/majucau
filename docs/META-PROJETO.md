# Meta do Majucau

## O que queremos entregar

Concluir o backlog aprovado do Majucau e entregar primeiro um protótipo que possa ser testado e homologado. Depois da homologação, seguir para a finalização e entrada em produção previstas no próprio backlog.

## Regras que devem ser preservadas

- Não alterar telas já aprovadas.
- Não alterar regras financeiras ou de negócio aprovadas.
- Se houver dúvida sobre um requisito, dado ou interpretação, parar essa parte do trabalho e perguntar antes de implementar.
- Nunca alterar telas aprovadas nem regras financeiras ou de negócio sem autorização explícita.
- Registrar cada avanço, teste, pendência e próximo passo no projeto e enviar as atualizações ao GitHub.

## Andamento registrado

- 127 tarefas no backlog: 7 concluídas, 19 em andamento e 101 não iniciadas.
- Progresso estimado: **11,7% concluído; 88,3% restante**. Essa média serve para acompanhamento do backlog e não prova que o sistema esteja pronto para homologação ou produção.
- Corrigi a configuração do Bling: os campos Client ID e endereço de retorno agora aparecem preenchidos com os dados públicos do aplicativo já fornecidos pela usuária quando ainda não há valores salvos. Eles continuam editáveis e só são gravados ao clicar em “Salvar configuração”; valores já salvos continuam prevalecendo. O Client Secret não é preenchido, não é revelado e deve ser informado no Majucau, onde o worker o protege. Nenhuma tela aprovada ou regra financeira foi alterada. A compilação completa do app, worker, auxiliar, instalador e pacote portátil passou; 33 testes frontend, os testes Go e as verificações locais também passaram. A usuária precisa instalar/abrir o novo pacote, salvar o Client Secret do aplicativo no Majucau e autorizar a conexão. Não houve conexão real ao Bling. Evidência e hashes: `docs/evidence/BLING-INSTALLER-UPDATED-2026-10-18.md`.
- Próxima entrega em andamento: validar a integração real do Drive e do Bling; a leitura recursiva e o caminho do CSV Nuvem Pago já estão ligados ao worker, mas nenhum arquivo real foi importado.
- Nesta atualização, o Majucau passou a guardar páginas confirmadas do Bling e retomá-las depois de uma interrupção. A publicação do novo retrato acontece inteira ou não acontece. A CI do GitHub passou para o commit `951d4a6`.
- Nesta execução, corrigi uma suposição insegura: página vazia sem explicação do Bling não comprova que a lista terminou. O sistema agora falha com segurança, preserva os dados válidos anteriores e não avança o cursor. Os testes locais, `go vet`, teste PostgreSQL, verificação do backlog e CI Windows passaram no commit `aff28d3` ([CI #39173420786](https://github.com/guilhermejacyzin/majucau/actions/runs/39173420786)).
- No Drive, a listagem agora tem checkpoint persistente de cada página, versões de metadados e publicação atômica do inventário completo. Se a execução falhar, o retrato anterior permanece e a leitura pode continuar do último marcador confirmado. Depois de incluir os modelos SQLC, a CI completa passou no commit `ef83bed` nos quatro jobs Windows, PostgreSQL, backlog e detecção de mudanças. BK-042 está em 52%; não houve homologação real.
- Em 18/10, a leitura do Drive passou a percorrer subpastas dentro das mesmas três raízes aprovadas. O progresso de cada pasta e página fica salvo; se houver interrupção, o worker retoma e só publica o inventário completo. O staging privado está ligado ao importador de CSV de recebimentos futuros, que mantém arquivos atrasados como pendentes. O teste local completo passou; o teste de retomada com PostgreSQL foi acrescentado à CI. A pasta real é `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`; os IDs foram identificados, mas não foram gravados no Git nem configurados localmente. A identidade técnica/credencial ainda não foi provisionada, e o CSV disponível no Drive estava desatualizado; não houve download nem importação de valores reais. BK-042: 68% estimado; projeto: 11,7% estimado pela média das 127 tarefas. Isso não representa homologação nem prontidão de produção.
- A conferência da regra JWT do Bling confirmou que o código envia `enable-jwt: 1` ao trocar e renovar tokens e em todas as consultas autenticadas. Isso confirma compatibilidade do código, mas não a autorização da conta real.
- Em 19/10, preparei um caminho seguro para configurar a leitura do Drive sem alterar telas: a chave técnica e os três IDs são gravados juntos no cofre DPAPI; o pedido vem por Named Pipe com SID Windows autenticado; e a alteração gera auditoria sanitizada no PostgreSQL. A agenda das 09:30 continua ativa e relê a configuração protegida. As suítes focadas e `scripts/verify.ps1` passaram localmente, inclusive 33 testes de interface, build, `go vet`, `govulncheck`, scan de secrets e `npm audit`; a CI #680 passou nos quatro jobs: backlog, detecção de mudanças, PostgreSQL e Windows; a migration, as permissões e a correção dos caminhos locais foram validadas. A conta humana confirmada continua sendo `financeiro@majucau.com.br`; nenhuma credencial técnica foi criada nem configurada. Evidência: `docs/evidence/GOOGLE-DRIVE-CREDENTIAL-PROVISION-2026-10-19.md`.
- As telas aprovadas e as regras financeiras permaneceram inalteradas.

## O que ainda falta para homologar o protótipo

1. Autorizar e validar a conexão real do Bling e confirmar como a lista informa que chegou à última página.
2. Aguardar CI PostgreSQL da nova migration; depois provisionar a identidade técnica de leitura com acesso somente às três raízes e executar a primeira sincronização real. O caminho CSV de recebimentos futuros já está ligado, mas o arquivo visto estava antigo. TSV, JSON, XLS e XLSX continuam pendentes até cada parser e contrato serem homologados.
3. Rodar o protótipo com dados de homologação, conferir os resultados e registrar qualquer divergência antes de produção.

Além desses pontos, há outras tarefas do backlog para concluir a operação, segurança, instalação e produção. Ainda não há prazo total confiável: a estimativa depende de validações externas e de tarefas que ainda não começaram.

## Próximo passo

A CI da migration de auditoria e da correção Windows passou. Próximo passo: provisionar a identidade técnica somente leitura e executar a primeira sincronização real do Drive; manter pendentes os arquivos cujos contratos financeiros ainda não foram homologados. Finalizar também a autorização e a homologação da paginação do Bling. Não marcar esta frente como pronta antes da leitura real completa.
