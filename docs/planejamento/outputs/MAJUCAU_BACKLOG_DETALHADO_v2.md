# Majucau — backlog de ponta a ponta, versão 2

Revisão de 13/09/2026: fechamento funcional enviado pelo usuário incorporado. Escopo desta rodada confirmado: somente atualizar o backlog. Nenhum código do produto, integração, serviço ou teste real foi executado.

Adendo aprovado em 04/10/2026, registrado no `G0-DECISION-REGISTER.md`: o Majucau calculará sua própria DRE com regras contábeis aprovadas; a DRE pronta do Bling deixa de ser tratada como resultado final. Este adendo substitui os trechos conflitantes abaixo. Princípios para estoque, produção e CPV/CMV foram recebidos e organizados em `docs/DRE-ACCOUNTING-POLICY-v1.md`; o mapeamento por conta/operação, os critérios específicos da empresa e os dados de origem continuam pendentes de validação contábil.

Foram revisadas 70 tarefas mantendo os IDs BK-000 a BK-126. As versões anteriores estão preservadas como histórico. Suas regras superadas não devem orientar a implementação.

## Base vigente

Somente Majucau, três diretores; Balanço Patrimonial permanece excluído. O cliente Wails/Windows e o PostgreSQL local são a etapa atual de transição. A direção aprovada é hospedar o serviço central em ambiente privado, com PostgreSQL sem acesso público (ADR-002). O login próprio do Majucau foi escolhido, com `financeiro@majucau.com.br` para `ADMIN` e a conta compartilhada `contato@majucau.com.br` para Diretores. Gisele confirmou essa escolha após ser informada de que o histórico não distinguirá qual diretor fez cada ação; a decisão posterior substitui a identidade individual da ADR-002 somente para Diretores. O provisionamento ainda depende da implementação e validação de autorização, MFA, sessão e recuperação. Provedor, orçamento, região e detalhes da operação diária continuam pendentes. A arquitetura usa backend Go, frontend React + TypeScript + Vite e monólito DDD hexagonal. O mockup aprovado define o visual; as decisões aprovadas definem fontes, cálculos e comportamento.

Bling é uma das fontes principais de lançamentos. A DRE é calculada pelo motor Majucau segundo regras contábeis aprovadas; o relatório pronto do Bling serve para comparação, não como resultado final. Drive somente de leitura conserva 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio. Não existe pasta adicional de investimentos.

Análise/conciliação normal usa movimentos até D-1; dia corrente requer solicitação expressa. A agenda de recebimentos futuros continua alimentando o caixa projetado. D-1 é corte da análise, não frequência de sincronização.

As regras funcionais existentes permanecem fechadas. A decisão posterior aprovou que o Majucau calcule sua própria DRE. Princípios contábeis gerais recebidos estão registrados em `docs/DRE-ACCOUNTING-POLICY-v1.md`; o mapeamento e as escolhas contábeis específicas da empresa ainda dependem de validação com a contabilidade. A engenharia ainda executará especificação técnica, mapeamento das fontes, implementação e validação. Artefato aprovado ausente no pacote será localizado, sem inventar seu conteúdo.

## Evolução

**Registro histórico de 13/09/2026 — não representa o progresso atual:** naquele levantamento, havia 127 tarefas, 3 levantamentos concluídos com evidência, 2,4% por quantidade e 0% de desenvolvimento do produto. O andamento atual exige auditoria tarefa a tarefa; não apresentar estes números históricos como estado de hoje.

Cada tarefa recebe 100% quando está Concluída e possui evidência de aceite. Peso inicial 1 mede quantidade. Evolução geral = soma(peso × conclusão) / soma(pesos). O denominador inclui tarefas condicionais ainda previstas; quando o escopo for efetivamente alterado, registrar a versão e revisar pesos/denominador. Não existe atualização automática pelo GitHub.

### Auditoria histórica — 04/10/2026

- O número de 2,4% e a indicação de desenvolvimento em 0% acima pertencem à revisão de 13/09/2026. Como o código e as evidências avançaram desde então sem atualização tarefa a tarefa neste arquivo, esses números são históricos e não devem ser apresentados como o progresso atual.
- A matriz `REQUIREMENTS-TRACEABILITY.md` tem 51 requisitos, distribuídos assim nesta auditoria: 1 `VERIFIED`, 5 `DOCUMENTED`, 22 `PARTIAL`, 11 `BLOCKED` e 12 `NOT_STARTED`.
- Apenas `VERIFIED` conta como requisito integralmente comprovado. Isso equivale a 1 de 51 (aproximadamente 2%) por contagem simples de requisitos; não é o percentual oficial do projeto, porque os requisitos têm tamanhos e pesos diferentes.
- Portanto, ainda não há percentual global nem prazo de conclusão confiável. A auditoria tarefa a tarefa começou em 10/10/2026; veja a atualização abaixo. Documentação isolada não fecha uma tarefa de implementação.
- Até essa revisão, comunicar o estado como “1 de 51 requisitos verificados; 22 parciais; 11 bloqueados; 12 não iniciados; 5 documentados”. Não converter automaticamente “não verificado” em “não feito”.

### Atualização da matriz — 10/10/2026

- Recontagem da matriz atual: 51 requisitos — 1 `VERIFIED`, 4 `DOCUMENTED`, 26 `PARTIAL`, 11 `BLOCKED` e 9 `NOT_STARTED`.
- Desde a conferência anterior, `ARC-01`, `ARC-03` e `OPS-01` avançaram de `NOT_STARTED` para `PARTIAL`; `UX-01` avançou de `DOCUMENTED` para `PARTIAL`, pois os testes e a CI foram comprovados, mas a comparação visual ainda falta.
- O resultado continua sendo 1 de 51 requisitos totalmente verificado (aproximadamente 2% por contagem simples). Isso não é uma estimativa de conclusão por esforço. O percentual e prazo global permanecem indisponíveis até a revisão individual das tarefas do backlog.
- Evidência da recontagem: `docs/evidence/TRACEABILITY-COUNT-CORRECTION-2026-10-10.md`. A matriz `REQUIREMENTS-TRACEABILITY.md` permanece como fonte atual dos estados.

### Auditoria do backlog — 10–11/10/2026

- BK-000 foi revisada contra ADR-001, ADR-002, `AGENTS.md`, `G0-DECISION-REGISTER.md`, a política contábil da DRE e a análise v2. Seu aceite foi atendido e a tarefa foi concluída com evidência.
- BK-004 foi revisada contra o inventário das fontes, o mapa MOCAP e as decisões vigentes: 15 imagens T1–T15/C11 estão presentes e seus hashes conferem; as correções de menu e T6 estão registradas. Nenhuma tela foi alterada.
- BK-002 foi reauditada contra a API pública do GitHub: main está em 39f8630; a branch de trabalho publicada está em b94bb3a, 178 commits à frente; há cinco PRs abertos e nenhuma issue. A nota anterior de repositório vazio é histórica.
- BK-001 foi reaberta: o inventário e as extrações estão disponíveis, mas os arquivos binários originais não estão no local privado; aguardamos a orientação da responsável sobre preservá-los.
- Cinco tarefas têm auditoria atual concluída: BK-000, BK-002, BK-003, BK-004 e BK-005. BK-001 aguarda os originais ou autorização para aceitar hashes/extrações; 121 tarefas ainda aguardam revisão.
- A planilha mostra 5 de 127 tarefas concluídas (3,9% por contagem simples). É estado do rastreador, não percentual real de conclusão do produto. Prazo global ainda indisponível.
- Esta auditoria não alterou código, telas aprovadas ou regras de negócio.

### Recontagem do andamento — 11/10/2026

- O backlog vigente contém 127 tarefas: 5 concluídas, 4 em andamento e 118 não iniciadas. As tarefas em andamento são BK-001 (0%), BK-006 (10%), BK-007 (10%) e BK-013 (15%).
- Como o plano atribui peso 1 a cada tarefa, a média dos percentuais registrados é `(5 × 100% + 0% + 10% + 10% + 15%) ÷ 127 = 4,2%`; portanto, o indicador do backlog tem 95,8% restante. Os percentuais de tarefas em andamento são estimativas, não medição de esforço nem prova de protótipo pronto.
- Os números 3,9% e 4,2% respondem perguntas diferentes: 3,9% é a proporção de tarefas totalmente concluídas; 4,2% inclui o avanço parcial registrado nas tarefas em andamento.
- Prazo global continua sem estimativa confiável: BK-001 aguarda materiais/decisão da responsável, outros módulos dependem de fontes e aprovações, e a revisão técnica das 127 tarefas ainda não terminou.

### Recontagem após auditorias de BK-025 e BK-067 — 11/10/2026

- BK-025 foi revisada contra `README.md`, `.gitignore`, `AGENTS.md`, `.github/workflows/ci.yml` e `CONTRIBUTING.md`. A base do repositório e as instruções estão disponíveis; a proteção efetiva de branches e a exigência de revisores no GitHub ainda não foram confirmadas. BK-025 passa a `Em andamento`, com estimativa de 60%; nenhum ajuste em tela ou regra de negócio foi feito.
- A rechecagem estática de BK-067 confirmou que a tela já renderiza EBIT, EBITDA, CAPEX pago no mês e CAPEX acumulado como cards de consulta, remove o card numérico UNMAPPED e preserva o alerta de lançamentos pendentes. A tabela não recebe CAPEX. Os valores continuam indisponíveis: EBIT/EBITDA ainda são traços e o backend não carrega os campos CAPEX a partir do relatório. BK-067 passa a `Em andamento`, com estimativa conservadora de 20%; não houve alteração de tela ou regra nesta auditoria.
- O backlog continua com 127 tarefas: 5 concluídas, 6 em andamento e 116 não iniciadas. As tarefas em andamento são BK-001 (0%), BK-006 (10%), BK-007 (10%), BK-013 (15%), BK-025 (60%) e BK-067 (20%).
- Pela média simples configurada no backlog, o indicador passa a `(5 × 100% + 0% + 10% + 10% + 15% + 60% + 20%) ÷ 127 = 4,8%`; restam 95,2%. É uma estimativa do rastreador, não uma medição de esforço nem prova de prontidão para produção.
- Sete tarefas tiveram auditoria atual concluída: BK-000, BK-002, BK-003, BK-004, BK-005, BK-025 e BK-067. BK-001 aguarda os materiais originais ou a decisão da responsável; 119 tarefas ainda aguardam revisão individual.
- Prazo global continua sem estimativa confiável, pois ainda faltam essa revisão, fontes reais, validações e decisões operacionais.

### Recontagem após conclusão de BK-030 — 11/10/2026

- A CI de Go/Windows e PostgreSQL executa testes, lint, análise de dependências e segredos, builds e checagens de artefatos em runners hospedados pelo GitHub. A etapa do instalador unsigned usa `if: success()`. A CI [38115370056](https://github.com/guilhermejacyzin/majucau/actions/runs/38115370056), no commit `c5eac57`, passou em Windows e PostgreSQL; a verificação da consolidação e o envio do artefato também passaram. A documentação oficial define que `success()` retorna verdadeiro somente quando todas as etapas anteriores passaram; ver a evidência `docs/evidence/CI-ARTIFACT-PUBLICATION-GATE-2026-10-11.md`. BK-030 foi concluída. Não houve execução local de testes.
- A proteção de branch e os status checks obrigatórios continuam sem confirmação remota e são acompanhados em BK-025; não fazem parte do critério de aceite da BK-030.
- O backlog continua com 127 tarefas: 6 concluídas, 6 em andamento e 115 não iniciadas. As tarefas em andamento são BK-001 (0%), BK-006 (10%), BK-007 (10%), BK-013 (15%), BK-025 (60%) e BK-067 (20%).
- Pela média simples configurada no backlog, o indicador permanece `(6 × 100% + 0% + 10% + 10% + 15% + 60% + 20%) ÷ 127 = 5,6%`; restam 94,4%. É uma estimativa do rastreador, não uma medição de esforço nem prova de prontidão para produção.
- Oito tarefas tiveram auditoria atual concluída: BK-000, BK-002, BK-003, BK-004, BK-005, BK-025, BK-030 e BK-067. BK-001 aguarda os materiais originais ou a decisão da responsável; 118 tarefas ainda aguardam revisão individual.
- Prazo global continua sem estimativa confiável, pois ainda faltam essa revisão, fontes reais, validações e decisões operacionais.

### Recontagem após CI de BK-031 — 11/10/2026

- BK-031 avançou para 75%: o worker registra em JSON a operação, origem, identificador da requisição, tempo, código de erro, lote e contagens aprovadas; os arquivos têm rotação limitada e permissões privadas. Não são gravados payloads, tokens, mensagens técnicas brutas, dados financeiros ou caminhos de arquivos.
- Permanecem pendentes a referência curta visível junto aos erros (aguardando decisão sobre alteração mínima da mensagem/tela) e a métrica de tamanho de fila. O worker não possui fila de lotes monitorável hoje; não registrar fila inexistente como zero. A ligação dos registros técnicos a uma referência exibida ao usuário também não foi aprovada.
- A CI `38116971454` passou em Windows e PostgreSQL, incluindo testes Go, `go vet`, builds do desktop, worker e instalador e consolidação dos outputs. O novo package de logs não tem testes próprios; runtime de ACL em uma instalação/VM ainda não foi demonstrado. Não foram executados testes locais.
- O trabalho não altera telas nem regras financeiras. O artefato unsigned publicado pela CI está registrado para expiração/exclusão manual em `artifacts/retained/cleanup-register.md`.
- O backlog tem 127 tarefas: 6 concluídas, 7 em andamento e 114 não iniciadas. Em andamento: BK-001 (0%), BK-006 (10%), BK-007 (10%), BK-013 (15%), BK-025 (60%), BK-031 (75%) e BK-067 (20%).
- Pela média simples, o rastreador indica `(6 × 100% + 0% + 10% + 10% + 15% + 60% + 75% + 20%) ÷ 127 = 6,2%`; restam 93,8%. É estimativa do backlog, não prontidão nem conclusão por esforço.
- Prazo de conclusão do projeto não é confiável enquanto 114 tarefas ainda não foram iniciadas e a revisão global, as fontes reais e decisões externas permanecem pendentes.

### Revalidação do GitHub — 11/10/2026

- BK-002 foi revalidada em modo de leitura. `main` continua em `39f8630`; a branch `codex/majucau-artifact-consolidation` está em `1d90a0a`, 206 commits à frente e nenhum atrás; há cinco PRs abertos e nenhuma issue aberta. Evidência: `docs/evidence/GITHUB-REPOSITORY-STATE-2026-10-11.md`.
- Os números de 178 commits à frente e ponta `b94bb3a` continuam registrados como a fotografia anterior de 10/10, não como estado atual.

### Rechecagem das proteções GitHub — 11/10/2026

- A consulta de rulesets retornou lista vazia. A consulta de proteção clássica da branch `main` foi negada com HTTP 403 porque a integração GitHub não tem permissão administrativa; isso não prova que a branch esteja desprotegida.
- BK-025 permanece em 60% até que uma pessoa administradora confirme em **Settings → Branches** se `main` exige revisão e verificações. Nenhuma configuração foi alterada. A auditoria foi documental e não executou testes.
- A estimativa atual do rastreador continua em 6,2% concluído e 93,8% restante, entre 127 tarefas; isso não representa prontidão para produção nem esforço medido. Evidência: `docs/evidence/GITHUB-REPOSITORY-STATE-2026-10-11.md`.

### Rechecagem do material de origem — 11/10/2026

- Foram encontrados quatro arquivos complementares de outubro na pasta privada; os hashes conferem com o registro local de limpeza, e o Git os ignora. Eles não substituem os 17 arquivos do pacote histórico.
- A busca pelos 17 nomes conhecidos não encontrou os arquivos em Downloads nem na pasta privada; a extração das vinte abas também está ausente. BK-001 segue em 0% aceito e aguarda o pacote histórico. Se ele não estiver disponível, a tarefa permanece pendente até que a responsável decida expressamente como tratar essa ausência; não será marcada como concluída só com os arquivos complementares.
- Não houve alteração de tela ou regra financeira, nem execução de testes. A estimativa geral continua em 6,2% concluído e 93,8% restante.

## Dependências e ordem prática

### Recontagem após auditoria de BK-006 — 11/10/2026

- BK-006 avançou de 10% para 15%: a cópia atual da calculadora foi mapeada em leitura somente para consulta e comparada com D-003-A. Os conflitos e campos que não alimentam as fórmulas foram registrados; nenhuma regra, tela ou planilha foi alterada. A ligação com os dados consolidados e os exemplos reproduzíveis continuam pendentes.
- O backlog mantém 127 tarefas: 6 concluídas, 7 em andamento e 114 não iniciadas. Em andamento: BK-001 (0%), BK-006 (15%), BK-007 (10%), BK-013 (15%), BK-025 (60%), BK-031 (75%) e BK-067 (20%).
- Pela média simples do rastreador, `(6 × 100% + 0% + 15% + 10% + 15% + 60% + 75% + 20%) ÷ 127 = 6,3%`; restam 93,7%. É uma estimativa do backlog, não uma medição de esforço nem prova de prontidão para produção.
- Não há prazo confiável para o projeto completo enquanto 114 tarefas não foram iniciadas e faltam fontes, integrações e decisões operacionais.

### Reauditoria estática do adaptador Bling — 11/10/2026

- BK-041 avançou de 0% para 35%: foram conferidos no código o cliente GET somente de leitura para contas a receber e a pagar, a paginação limitada, o corpo de resposta limitado, as três tentativas com backoff/jitter e `Retry-After` limitado a 30 segundos. OAuth usa POST sem repetição automática. A gravação RAW e o cursor são transacionais, com rollback de dados e auditoria de falha; evidências anteriores da CI cobrem também replay idempotente.
- Permanecem pendentes a homologação dos endpoints/campos da conta real, o cursor incremental oficial e a cobertura das fontes/recursos requeridos. A renovação concorrente ainda não tem exclusão no serviço: a conexão IPC atende uma chamada por vez, mas a sessão OAuth roda em goroutine e pode disputar gravações no armazenamento de credenciais. Nenhuma regra financeira ou tela foi alterada.
- O registro da calculadora continua preservando a estrutura da planilha recebida; essa confirmação não substitui homologação das fontes do Bling nem autoriza alterar a fórmula financeira.
- O backlog tem 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. Em andamento: BK-001 (0%), BK-006 (15%), BK-007 (10%), BK-013 (15%), BK-025 (60%), BK-031 (75%), BK-041 (35%) e BK-067 (20%). Onze tarefas tiveram auditoria atual concluída.
- Pela média simples do rastreador, `(6 × 100% + 0% + 15% + 10% + 15% + 60% + 75% + 35% + 20%) ÷ 127 = 6,5%`; restam 93,5%. É estimativa do backlog, não medição de esforço nem prontidão para produção. Não há prazo global confiável enquanto permanecem 113 tarefas não iniciadas, homologações e validações externas.

### Correção da concorrência das credenciais Bling — 11/10/2026

- BK-041 avançou de 45% para 50% após a CI completa confirmar a correção. O worker agora compartilha uma proteção entre salvar a configuração Bling e os fluxos de OAuth. Teste, renovação, conclusão do login, desconexão e leitura do token no início da sincronização não disputam o mesmo segredo; a sincronização libera a proteção antes de ler as páginas de dados.
- Um teste cobre o login do Bling em andamento enquanto `Test` é solicitado; outro cobre desconexão enquanto a troca do código por token está em andamento, verificando que a desconexão termina por último e os tokens ficam removidos. `go test ./internal/integrations/bling ./cmd/worker`, `go vet` e scanner de segredos passaram localmente com caches centrais. A CI completa `38120125096` passou nos jobs Windows e PostgreSQL, incluindo os cenários `-race` no Ubuntu; o alerta anterior do scanner foi corrigido com a marcação da fixture de teste.
- A trava protege os serviços dentro do processo worker atual. Uso futuro com vários processos/instâncias precisa de mecanismo coordenado entre processos. Homologação dos campos e do cursor oficial permanece pendente. Nenhuma tela ou regra financeira foi alterada.
- O backlog tem 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. Em andamento: BK-001 (0%), BK-006 (15%), BK-007 (10%), BK-013 (15%), BK-025 (60%), BK-031 (75%), BK-041 (50%) e BK-067 (20%).
- Pela média simples do rastreador, `(6 × 100% + 0% + 15% + 10% + 15% + 60% + 75% + 50% + 20%) ÷ 127 = 6,7%`; restam 93,3%. Não existe prazo global confiável enquanto há 113 tarefas não iniciadas, homologações e validações externas.

### Proteção contra paginação não confirmada no Bling — 11/10/2026

- BK-041 avançou de 50% para 55%. A sincronização agora exige que a resposta informe `hasNext` ou `total`; quando nenhum desses campos está disponível, registra `BLING_API_PAGINATION_UNKNOWN`, reverte os dados do lote e não avança o cursor. Assim, uma resposta válida mas sem marcador conhecido não é tratada como sincronização completa.
- O teste unitário cobre respostas sem metadados e com somente `page`/`limit`. O teste PostgreSQL prova que a falha preserva páginas RAW válidas e o cursor anterior. `go test ./internal/integrations/bling`, `go vet ./internal/integrations/bling` e `git diff --check` passaram localmente; a CI `38121237550` passou nos jobs Windows e PostgreSQL, incluindo esse teste integrado.
- Isso é uma proteção contra dados incompletos, não uma homologação do contrato do Bling. A confirmação real de como a resposta encerra a lista e a retomada incremental continuam pendentes. Nenhuma tela, regra financeira ou estrutura da calculadora foi alterada.
- O backlog permanece com 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. Em andamento: BK-001 (0%), BK-006 (15%), BK-007 (10%), BK-013 (15%), BK-025 (60%), BK-031 (75%), BK-041 (55%) e BK-067 (20%).
- Pela média simples do rastreador, `(6 × 100% + 0% + 15% + 10% + 15% + 60% + 75% + 55% + 20%) ÷ 127 = 6,7%`; restam 93,3%. Não há prazo global confiável enquanto há 113 tarefas não iniciadas, homologações e validações externas.

### Testes próprios dos logs técnicos — 11/10/2026

- BK-031 avançou de 75% para 80% após acrescentar testes de limites, rotação e retenção dos arquivos recentes, registro JSON e permissões privadas. O gravador também rejeita limites de rotação inválidos e um registro maior que o próprio limite do arquivo antes de gravar.
- A CI `38122175947` passou nos jobs Windows e PostgreSQL/Linux. O job Linux executou os testes de arquivo e `go vet` do pacote; no Windows passaram a suíte Go, `go vet`, verificações de segurança, frontend, build e instalador. Testes locais de `internal/diagnosticslog` e `go vet` passaram no Windows.
- Continuam pendentes o teste de ACL em uma instalação/VM limpa, uma referência que ligue o erro mostrado ao log (qualquer mudança na tela aguarda aprovação) e métricas de fila, que só serão definidas se o worker passar a ter uma fila real. Nenhuma tela ou regra financeira foi alterada; a estrutura da calculadora continua sendo a da planilha enviada.
- O backlog tem 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. Em andamento: BK-001 (0%), BK-006 (15%), BK-007 (10%), BK-013 (15%), BK-025 (60%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 15% + 10% + 15% + 60% + 80% + 55% + 20%) ÷ 127 = 6,7%`; restam 93,3%. É uma estimativa do backlog, não uma medição de esforço nem prova de prontidão para produção.
- Não há prazo global confiável enquanto 113 tarefas não foram iniciadas e permanecem decisões, fontes reais e validações operacionais pendentes. Evidência: `docs/evidence/WORKER-STRUCTURED-DIAGNOSTICS-2026-10-11.md`.

### Testes do fluxo diário em streaming — 11/10/2026

- A `DailyProjectionFrom` agora tem testes próprios: o leitor gera 1.000 movimentos sob demanda, produz uma série limitada a D0–D+60 e mantém os totais corretos; erros da fonte, leitor ausente e estouro de moeda falham sem devolver uma série parcial.
- `go test ./internal/financial -count=1`, `go vet ./internal/financial` e `git diff --check` passaram no Windows. O teste confirma o helper de cálculo em fluxo; ainda não há chamada de produção, consulta ao saldo diário real nem conexão com a calculadora.
- A CI completa `38123193176`, no commit `6983a76`, passou nos jobs Windows e PostgreSQL/Linux. Passaram testes e `go vet`, verificações de segurança, migrations/sqlc, frontend, builds, instalador e versão portátil sem assinatura, smoke de instalação/desinstalação, consolidação e upload dos outputs. O artefato temporário está no registro local de limpeza. Essa aprovação não comprova integração da calculadora com dados financeiros reais.
- BK-006 permanece em 15%, pois os dados reais de entrada e a ligação ao fluxo consolidado continuam pendentes. A estrutura da planilha enviada, telas e regras financeiras não foram alteradas.
- O backlog continua com 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. A média simples permanece em 6,7% concluído e 93,3% restante; não há prazo global confiável.
- Evidência: `docs/evidence/APPLICATION-RULE-D003A-2026-09-20.md`.

### Recontagem após conferência das regras da DRE — 11/10/2026

- A política `docs/DRE-ACCOUNTING-POLICY-v1.md` registra a estrutura das linhas já aprovadas, o que entra e não entra, o caminho de estoque/produção/CPV, o tratamento de máquinas/depreciação, tributos recuperáveis, as proibições de classificação automática e as informações que serão necessárias das fontes. Ela preserva as telas e linhas aprovadas. Não é parecer contábil final.
- BK-007 avança de 10% para 20% pelo registro organizado dos princípios que Gisele enviou. Não houve mudança na regra nem na tela. Validação com a contabilidade, critérios específicos da empresa, mapeamento de contas/eventos e dados reais continuam pendentes.
- O backlog mantém 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. BK-007 passa a 20%; as demais tarefas em andamento mantêm seus percentuais.
- Pela média simples, `(6 × 100% + 0% + 15% + 20% + 15% + 60% + 80% + 55% + 20%) ÷ 127 = 6,8%`; restam 93,2%. O número é estimativa do rastreador, não prontidão para produção.
- Não há prazo global confiável enquanto 113 tarefas não foram iniciadas e as validações contábeis, fontes reais e decisões operacionais permanecem pendentes.

### Conferência automática do resumo do backlog — 11/10/2026

- `scripts/verify-backlog.ps1` foi criado para ler as tarefas uma por vez, conferir IDs repetidos, estado e percentual de cada uma e comparar contagens e média com a última recontagem antes da lista detalhada.
- A conferência e seu teste de rejeição de resumo incorreto passaram localmente e na CI Windows. A CI completa do commit `a61dd71` terminou com sucesso no run `38124546590`, incluindo Windows e PostgreSQL/Linux.
- BK-025 avança de 65% para 70% com a validação remota; a confirmação administrativa das regras de proteção da branch ainda está pendente. Nenhuma tela ou regra financeira foi alterada.
- O backlog mantém 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. Em andamento: BK-001 (0%), BK-006 (15%), BK-007 (20%), BK-013 (15%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 15% + 20% + 15% + 70% + 80% + 55% + 20%) ÷ 127 = 6,9%`; restam 93,1%. É estimativa do rastreador, não prontidão para produção.

### Sugestão de reserva implementada isoladamente — 11/10/2026

- `internal/financial.SuggestedReserve` implementa a fórmula aprovada D-003-C usando uma série validada de 61 saldos e aritmética decimal exata. A função é separada do máximo D-003-A.
- Os testes locais validam o exemplo de R$ 80.000, provam que o máximo D-003-A permanece R$ 20.000 nessa série e rejeitam série incompleta, reserva negativa e estouro. `go test ./internal/financial -count=1`, `go vet ./internal/financial` e `git diff --check` passaram. A CI completa `38125567483`, no commit `0c74b1e`, também passou em Windows e PostgreSQL/Linux.
- BK-006 avança de 20% para 25% com a validação remota. A ligação ao fluxo consolidado, a DRE própria, o principal líquido já investido e a apresentação permanecem pendentes. Nenhuma tela ou regra financeira foi alterada.
- O backlog continua com 127 tarefas: 6 concluídas, 8 em andamento e 113 não iniciadas. Em andamento: BK-001 (0%), BK-006 (25%), BK-007 (20%), BK-013 (15%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 25% + 20% + 15% + 70% + 80% + 55% + 20%) ÷ 127 = 7,0%`; restam 93,0%. Não há prazo global confiável enquanto as fontes reais e validações operacionais continuam pendentes.

### Inventário parcial do ambiente Windows de desenvolvimento — 11/10/2026

- BK-014 avançou de 0% para 10% com a leitura somente para consulta do Windows, CPU, memória, espaço em disco, conectividade reportada e serviços relacionados. O PostgreSQL 18 encontrado não foi presumido exclusivo nem alterado. A evidência omite nomes de host, contas, endereços de rede e credenciais.
- O inventário descreve somente o computador usado no desenvolvimento. Não confirma o host operacional ou de produção; energia, estabilidade da rede, permissões e rotinas de backup/recuperação para o ambiente escolhido ainda precisam ser validadas. A direção aprovada de hospedagem privada centralizada permanece inalterada.
- Nenhum serviço, banco, porta, tela ou regra financeira foi modificado. Evidência: `docs/evidence/WINDOWS-ENVIRONMENT-INVENTORY-2026-10-11.md`.
- O backlog tem 127 tarefas: 6 concluídas, 9 em andamento e 112 não iniciadas. Em andamento: BK-001 (0%), BK-006 (25%), BK-007 (20%), BK-013 (15%), BK-014 (10%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 25% + 20% + 15% + 10% + 70% + 80% + 55% + 20%) ÷ 127 = 7,0%`; restam 93,0%. Não há prazo global confiável enquanto permanecem fontes reais, decisões operacionais e validações externas pendentes.

### Auditoria parcial das ações e exportações — 11/10/2026

- BK-012 avançou de 0% para 10% após conferir, em leitura somente, os controles atuais do painel e as rotas IPC/Wails disponíveis. A evidência separa os controles funcionais dos controles desativados ou ainda sem ação conectada.
- A exportação de relatórios não tem rota no IPC/Wails atual. Embora a capacidade geral de Diretores exportarem relatórios esteja aprovada, os formatos, dados, filtros e permissões por tela continuam indefinidos em L-22/Q-13. A tela T11 permanece somente para consulta, conforme a decisão explícita.
- Os métodos de configuração, sincronização e importação CSV foram inventariados, mas isso não comprova autorização por perfil; essa verificação permanece em BK-013. Nenhuma tela ou permissão foi alterada. Evidência: `docs/evidence/ACTION-EXPORT-STATIC-AUDIT-2026-10-11.md`.
- O backlog tem 127 tarefas: 6 concluídas, 10 em andamento e 111 não iniciadas. Em andamento: BK-001 (0%), BK-006 (25%), BK-007 (20%), BK-012 (10%), BK-013 (15%), BK-014 (10%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 25% + 20% + 10% + 15% + 10% + 70% + 80% + 55% + 20%) ÷ 127 = 7,1%`; restam 92,9%. Ainda não há prazo global confiável: faltam 111 tarefas não iniciadas e dependências de dados e aprovação.

## Dependências e ordem prática

### Verificação do ciclo de suporte do Windows — 11/10/2026

- A documentação oficial confirma que o suporte regular do Windows 10 Pro 22H2 terminou em 14/10/2025. O programa ESU exige inscrição e é temporário; instalar apenas o pacote preparatório não inscreve a máquina. A situação de ESU deste computador não foi verificada.
- BK-014 avança de 10% para 15%. A máquina inventariada segue classificada como ambiente de desenvolvimento, não como host operacional. O equipamento final, a comprovação de ESU ou atualização para Windows suportado e as demais condições operacionais permanecem pendentes.
- Nenhuma atualização, inscrição, instalação ou configuração do Windows foi realizada. Evidência: `docs/evidence/WINDOWS-ENVIRONMENT-INVENTORY-2026-10-11.md`; fontes oficiais Microsoft estão vinculadas nela.
- O backlog tem 127 tarefas: 6 concluídas, 10 em andamento e 111 não iniciadas. Em andamento: BK-001 (0%), BK-006 (25%), BK-007 (20%), BK-012 (10%), BK-013 (15%), BK-014 (15%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 25% + 20% + 10% + 15% + 15% + 70% + 80% + 55% + 20%) ÷ 127 = 7,2%`; restam 92,8%. Não há estimativa confiável de conclusão enquanto permanecem 111 tarefas não iniciadas, decisões e validações externas.

### Verificação do backlog em mudanças só documentais — 11/10/2026

- A CI anterior ao commit `9c4c727` executou apenas `detect-changes` em uma atualização documental e ignorou `verify-windows` e `verify-postgres`; por isso, o validador de contagens não rodou. Evidência: execução `38130699588` e `docs/evidence/BACKLOG-CI-ALWAYS-ON-2026-10-11.md`.
- A CI agora possui um job `verify-backlog` separado, que roda o verificador e seu teste de rejeição em cada push e pull request, inclusive quando só documentos mudam. A execução `38130879386` passou nesse job e também nos jobs completos de Windows e PostgreSQL.
- BK-025 permanece em 70%: a conferência automatizada agora também cobre documentação, mas a branch `main` continua sem proteção e exige ação de uma pessoa administradora. Nenhuma tela ou regra financeira foi alterada.
- O backlog tem 127 tarefas: 6 concluídas, 10 em andamento e 111 não iniciadas. Em andamento: BK-001 (0%), BK-006 (25%), BK-007 (20%), BK-012 (10%), BK-013 (15%), BK-014 (15%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 25% + 20% + 10% + 15% + 15% + 70% + 80% + 55% + 20%) ÷ 127 = 7,2%`; restam 92,8%. Ainda não há prazo global confiável.

### Conferência do pacote local do instalador — 11/10/2026

- A inspeção somente para leitura do NSIS confirmou que ele valida o pacote e instala o aplicativo, worker e helper, além de copiar a migration inicial. O pacote não contém PostgreSQL nem provisiona, inicia ou verifica um serviço PostgreSQL dedicado.
- `cmd/worker/database.go` lê `MAJUCAU_DATABASE_URL` do ambiente do processo, mas o instalador ainda não configura essa entrada nem há um fluxo aprovado/testado para entregar a senha do banco ao serviço com proteção de segredo. Não colocar uma DSN com senha na configuração persistente do serviço antes de fechar esse desenho.
- As ferramentas PostgreSQL presentes no computador pertencem à instalação já existente, que não foi presumida exclusiva e continua intocada. Não foi criado cluster/serviço adicional nesta execução.
- Esta evidência identifica uma dependência para o protótipo instalado completo: o provisionamento isolado do banco precisa entrar no caminho suportado de instalação e passar por validação de ciclo de vida/reboot. Ela não autoriza reaproveitar o serviço atual nem muda as telas ou regras de negócio/financeiras.
- BK-014 continua em 15%; documentar o estado do instalador melhora o inventário, mas o critério de aceite da proposta operacional ainda não foi atendido. Evidência: `docs/evidence/WINDOWS-ENVIRONMENT-INVENTORY-2026-10-11.md`.
- O total e a média do backlog não mudam: 6 concluídas, 11 em andamento e 110 não iniciadas; estimativa simples de 7,4% concluído e 92,6% restante. Não há prazo global confiável.

### Contrato público de paginação do Bling — 11/10/2026

- A documentação OpenAPI oficial descreve `GET /contas/receber` com `pagina` e `limite`, mas o schema de resposta publicado declara somente `data` como lista; não declara `hasNext`, `total` ou indicador de página final. O FAQ confirma o padrão de até 100 registros e o uso de `pagina`.
- A rota também declara `tipoFiltroData`, `dataInicial` e `dataFinal`; não declara `dataAlteracaoInicial/Final`. A menção genérica desses filtros na página de limites não prova suporte nesta rota específica.
- BK-041 permanece em 55%: a regra que falha com rollback quando não reconhece o fim da paginação continua necessária; não foi presumido que uma página curta encerra a lista. A resposta autenticada da conta e a retomada incremental permanecem pendentes porque a preferência salva do Chrome ainda bloqueia a página do Bling; não contornar essa restrição.
- Nenhuma tela, regra financeira ou comportamento do adapter foi alterado; não foram executados testes. Evidência e fontes: `docs/evidence/BLING-PAGINATION-CONTRACT-2026-10-11.md`.
- O backlog permanece com 127 tarefas: 6 concluídas, 11 em andamento e 110 não iniciadas; estimativa simples de 7,4% concluído e 92,6% restante. Evidência documental não altera o aceite nem o percentual da tarefa.

### Testes de integração PostgreSQL no Windows — 11/10/2026

- Um PostgreSQL 18.3 descartável foi criado sob `artifacts/cache`, limitado a loopback, sem mexer no serviço existente e sem instalar serviço Windows. A verificação de checksums passou e as três migrations foram aplicadas ao banco vazio.
- `go test -count=1 ./internal/integrations/bling ./internal/integrations/nuvempago` passou com as fixtures sanitizadas e variáveis de ambiente dos testes opt-in. A execução incluiu os E2Es de importação de recebimentos do Bling, atomicidade/cursor da sincronização de API e repetição idempotente dos recebíveis futuros.
- O processo foi encerrado; a porta 55439 foi liberada e `artifacts/cache` removida pela rotina segura. Nenhuma tela ou regra financeira mudou. Evidência: `docs/evidence/LOCAL-POSTGRES-INTEGRATION-2026-10-11.md`.
- BK-041 permanece em 55%, pois paginação/campos oficiais do Bling ainda precisam de homologação. BK-014 permanece em 15%, pois o instalador continua sem provisionar o banco dedicado. As contagens e a estimativa global permanecem 6 concluídas, 11 em andamento, 110 não iniciadas; 7,4% concluído e 92,6% restante.

### Direcionamento de temporários Go — 11/10/2026

- A revisão das rotinas oficiais identificou que `bootstrap.ps1`, `verify.ps1` e `build-windows.ps1` já centralizavam cache de compilação e módulos, mas deixavam o diretório temporário do Go no padrão global do Windows. Os três scripts agora definem `GOTMPDIR` em `artifacts/cache/go-tmp` e criam a pasta antes de executar Go; `AGENTS.md` torna o padrão obrigatório.
- Uma consulta manual anterior foi executada fora desses scripts e baixou três versões de dependências para o cache de módulos compartilhado do perfil. Como esse cache pode ser usado por outros projetos, não foi apagado; os caminhos exatos foram registrados no registro local ignorado `artifacts/retained/cleanup-register.md` para decisão manual. Os comandos seguintes usarão os caminhos de cache do repositório.
- O parser PowerShell e as verificações de caminho dos três scripts passaram. A evidência sanitizada está em `docs/evidence/CACHE-CLEANUP-2026-10-11.md`. BK-025 continua em 70%, pois a proteção da branch `main` e os checks obrigatórios ainda dependem de uma administradora. Percentual global permanece 7,4% concluído e 92,6% restante.

### Retry compartilhado nas consultas Bling — 11/10/2026

- As duas implementações repetidas de retry foram consolidadas em `withBlingReadRetry`. A função só aceita GET e rejeita POST antes de executar a operação. As consultas GET preservam três chamadas no máximo, espera exponencial com jitter, `Retry-After` limitado a 30 segundos e encerramento por cancelamento/erro permanente. A troca de token OAuth continua sem retry automático.
- `TestBlingReadRetryRejectsNonGETBeforeCallingRequest`, `go test -count=1 ./internal/integrations/bling` e `go vet ./internal/integrations/bling` passaram. `GOCACHE`, `GOMODCACHE` e `GOTMPDIR` foram limitados a `artifacts/cache`; a pasta foi removida ao final.
- BK-041 permanece em 55%, pois ainda faltam homologação real dos campos e cursor/paginação oficial do Bling. A mudança não altera regra financeira, tela ou chamada à conta real. Evidência: `docs/evidence/BLING-ADAPTER-STATIC-AUDIT-2026-10-11.md`.

IDs e grupos temáticos foram preservados; sua ordem numérica não é cronograma. As dependências de cada tarefa governam a execução. Em particular, a projeção de insumos BK-094 precede a recomendação BK-082/BK-084.

1. Transcrever contratos fechados e localizar referências; desenhar domínios, acesso, persistência e operação local.
2. Preparar fundações e adaptadores; comprovar campos reais do Bling/arquivos.
3. Definir com aprovação as regras e o mapeamento da DRE própria do Majucau; calcular por competência e comparar com a origem. Construir conciliação, caixa e disponibilidade que alimentam a calculadora.
4. Construir estoque/produção realizados e forecast; calcular produção líquida, trajetória de estoque e insumos por data.
5. Implementar compras líquidas e prioridade com cobertura parcial/temporal; sinais adicionais de cacau são independentes do motor principal.
6. Integrar telas fiéis aos mockups, controles, exportações, capacidades, vigência e histórico.
7. Executar verificações, validar com a Majucau, ensaiar instalação/backup/restauração, entrar em produção e acompanhar a operação.

Tarefas condicionais de fontes externas do cacau permanecem previstas, sem bloquear o funcionamento do motor principal. Nenhuma ausência de dados permite valor fictício: componentes obrigatórios ausentes produzem indisponibilidade, e componentes independentes comprovados permanecem utilizáveis.

## Critério comum de conclusão

Além do aceite específico: regras vigentes e capacidades aplicadas no backend, visual congelado preservado, métricas compartilhadas, ausência distinta de zero, origem e competência rastreáveis, períodos fechados preservados, verificações proporcionais executadas e evidência registrada. Os 167 cenários da planilha original são previstos e precisam ser atualizados nas regras superadas; não foram executados nesta etapa.

### Recontagem após a revisão de memória e handles do Windows — 11/10/2026

- BK-024 segue em 30% e `Em andamento`. Além do mapa de riscos, foram corrigidos e verificados fechamentos de conexão pendente, evento, processo/token, buffers DPAPI, descritores temporários, mutex do instalador e recursos de importação CSV em streaming. A primeira leva de injeções controladas em `securetemp` cobriu falha ao liberar memória nativa e falha combinada de liberação e remoção; a evidência está em `docs/evidence/THREAT-MODEL-INITIAL-2026-10-11.md`. A segurança geral não está concluída; faltam outras falhas controladas, fechar o mapa de riscos, VM limpa, rede e acesso por perfil.
- O backlog tem 127 tarefas: 6 concluídas, 11 em andamento e 110 não iniciadas. Em andamento: BK-001 (0%), BK-006 (30%), BK-007 (20%), BK-012 (10%), BK-013 (15%), BK-014 (15%), BK-024 (30%), BK-025 (70%), BK-031 (80%), BK-041 (55%) e BK-067 (20%).
- Pela média simples, `(6 × 100% + 0% + 30% + 20% + 10% + 15% + 15% + 30% + 70% + 80% + 55% + 20%) ÷ 127 = 7,4%`; restam 92,6%. É uma estimativa do backlog, não uma medida de esforço nem prova de prontidão. O prazo global continua sem estimativa confiável.

## 00 Descoberta e decisões

### BK-000 — Confirmar a base de escopo e as decisões da Majucau

**Situação:** Concluída. **Evolução:** 100%. **Tipo:** Planejamento. **Responsável sugerido:** Majucau + Arquitetura.

**Objetivo:** Ter uma referência única para construir o produto.

**O que fazer:** Registrar o fechamento vigente e sua precedência sem reabrir regras de negócio.

**Como fazer:**

1. Preservar a Majucau como empresa única e os três diretores. Registrar separadamente a base atual Wails/Windows com PostgreSQL local, Go, React + TypeScript + Vite e monólito DDD hexagonal, e a direção aprovada de hospedagem privada centralizada conforme ADR-002; detalhes operacionais continuam pendentes.
2. Aplicar o fechamento mais recente a cálculos, fontes, compras, produção, permissões e operação. O mockup aprovado governa o visual e as decisões aprovadas governam comportamento e cálculos. Mudanças posteriores aprovadas, como DRE própria, cards EBIT/EBITDA e CAPEX e exclusão do Balanço Patrimonial, permanecem rastreadas sem redesenhar telas ou reabrir regras.
3. Manter como registro histórico que em 13/09/2026 o usuário autorizou somente atualizar o backlog naquela rodada. Decisões posteriores e atividades futuras de mapeamento, implementação e validação ficam registradas nas fontes de governança e nas tarefas correspondentes.

**Critério de aceite:** Referência vigente, mudanças e atividades futuras estão rastreadas; regra definida não é apresentada como nova pergunta à Majucau.

**Impacto:** Evita construir uma regra ou módulo diferente do esperado.

**Evidência da conclusão (10/10/2026):** BK-000 auditado em 10/10/2026. Arquitetura/destino: ADR-001/002. Telas e regras aprovadas: G0. Pendências mantidas; sem alterações.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** Pedido do usuário; 00_Resumo; C_Conflitos; D_Lacunas; Consolidação 13/09/2026, consolidação integral; Fechamento vigente 13/09/2026, integral

**Regras e verificações vinculadas:** Todas as decisões em aberto,F-13

### BK-001 — Inventariar os materiais e suas versões

**Situação:** Em andamento. **Evolução:** 0% aceito. **Tipo:** Planejamento. **Responsável sugerido:** Produto + UX.

**Objetivo:** Saber exatamente o que foi analisado e preservar as fontes.

**O que fazer:** Catalogar imagens, abas, documentos citados, versões e arquivos originais.

**Como fazer:**

1. Conferir o registro histórico de 16 PNGs (15 únicos), 20 abas e a duplicata T10; preservar a extração disponível.
2. Conferir o inventário atual dos mockups aprovados e seus hashes, sem confundir versões atuais e materiais históricos.
3. Manter identificados documentos citados mas não recebidos, hashes dos fechamentos e as mudanças entre v0/v1/v2.
4. Guardar os originais no local privado aprovado ou obter autorização expressa para considerar hashes e extrações suficientes.

**Critério de aceite:** Pacote histórico, abas, duplicatas, ausências e versões ficam rastreáveis; os originais são preservados ou a responsável aprova expressamente a retenção somente de hashes e extrações.

**Impacto:** Evita omissões e mistura de versões.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** docs/planejamento/FONTES.md; docs/source/README.md; extrações do pacote; revision_v1.json; revision_v2.json.

**Evidência e pendência:** Os registros e os documentos v0/v1/v2 foram preservados. Em 11/10/2026, a conferência local encontrou os quatro materiais complementares de outubro na pasta privada e confirmou correspondência com o registro de limpeza; eles não são o pacote histórico. Nenhum dos 17 nomes do pacote inicial foi encontrado em Downloads ou em `sources/private/material-original/`, e `work/source_extract/` não existe neste checkout. Assim, não foi possível recalcular os hashes do pacote original nem revisar as vinte abas. A tarefa continua em 0% aceito e aguarda o pacote original; se ele não estiver disponível, o inventário permanece aberto até decisão expressa sobre a ausência, sem tratar os quatro arquivos complementares como substitutos.

### BK-002 — Verificar o ponto de partida no GitHub

**Situação:** Concluída. **Evolução:** 100%. **Tipo:** Planejamento. **Responsável sugerido:** Arquitetura.

**Objetivo:** Separar trabalho existente de trabalho futuro.

**O que fazer:** Inspecionar repositório, commits, branches, issues e pull requests.

**Como fazer:**

1. Preservar a resposta de repositório vazio da versão 0 como fotografia histórica.
2. Consultar em modo de leitura o estado atual: repositório público, branch principal, branches remotas, issues e pull requests.
3. Registrar a diferença entre a branch de trabalho publicada e main; não mesclar por inferência.
4. Registrar a confirmação da responsável de que não há outros clones, servidores ou ambientes de trabalho com código/dados além deste checkout; o controle do código fica no GitHub.

**Critério de aceite:** O histórico e o estado atual do GitHub são distinguíveis; branches, issues e pull requests têm evidência; a confirmação da responsável de que não há outros ambientes de trabalho fora deste checkout está registrada.

**Impacto:** Permite planejar o trabalho sem confundir código existente, histórico e trabalho futuro.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** `docs/evidence/GITHUB-REPOSITORY-STATE-2026-10-10.md` (fotografia histórica); `docs/evidence/GITHUB-REPOSITORY-STATE-2026-10-11.md` (estado atual); endpoints de leitura da API do GitHub.

**Evidência:** Na revalidação de 11/10/2026, `main` está em `39f8630` e a branch de trabalho publicada está em `1d90a0a`, 206 commits à frente e nenhum atrás; há cinco PRs abertos e nenhuma issue. A observação antiga de repositório vazio e a contagem anterior de 178 commits são históricas. A responsável confirmou em 10/10/2026 que não há outro clone, servidor ou ambiente de trabalho fora deste checkout; o código é controlado somente pelo Git/GitHub. Nenhuma branch foi mesclada e nenhum teste foi executado.

### BK-003 — Auditar cada tela e seus cálculos ilustrativos

**Situação:** Concluída. **Evolução:** 100%. **Tipo:** Planejamento. **Responsável sugerido:** UX + Produto + QA.

**Objetivo:** Transformar imagens em comportamentos verificáveis.

**O que fazer:** Mapear cartões, tabelas, gráficos, filtros, links, mensagens e inconsistências.

**Como fazer:**

1. Percorrer cada imagem e suas notas
2. Conferir somas, percentuais, datas, unidades e estados
3. Registrar conflitos sem escolher silenciosamente uma referência

**Critério de aceite:** Cada tela possui inventário e achados, com destino no backlog.

**Impacto:** Reduz erros que passariam despercebidos numa reprodução visual.

**Dependências:** BK-001

**Fonte:** Todas as imagens; CEIs T11–T17

**Evidência:** MAJUCAU_ANALISE_E_DECISOES_v0.md — inventário e análise; inspeção dos materiais e GitHub nesta sessão. Evidência original preservada na versão 0; esta revisão não altera a conclusão desse levantamento.

### BK-004 — Localizar e versionar os mockups e contratos aprovados

**Situação:** Concluída. **Evolução:** 100%. **Tipo:** Planejamento. **Responsável sugerido:** Majucau + UX.

**Objetivo:** Disponibilizar à implementação as referências visuais oficiais de cada tela.

**O que fazer:** Manter localizadas e versionadas as referências visuais oficiais e as decisões de escopo que governam sua implementação.

**Como fazer:**

1. Usar o inventário de `docs/source/README.md` e a evidência `docs/evidence/MOCAP-T1-T15-REALIGNMENT-2026-09-20.md` para vincular as referências T1–T15/C11 preservadas no repositório; os 15 hashes foram conferidos nesta auditoria.
2. Aplicar as correções aprovadas registradas no G0: Balanço Patrimonial fora da versão 1, Balancete mantido, somente Calculadora de Aplicação no menu e cards T6 nos locais aprovados. “Aplicações” não é uma rota separada.
3. Manter as imagens como referência histórica/visual sem copiar números ilustrativos. Regras de cálculo pendentes seguem suas próprias tarefas (BK-006/BK-007); esta conclusão não as presume nem altera telas.

**Critério de aceite:** Referências T1–T15/C11 estão presentes, seus hashes correspondem ao inventário e o mapa liga as telas às decisões vigentes; rotas excluídas não são reintroduzidas; nenhum mockup ou tela foi alterado.

**Impacto:** Impede declarar produção completa com módulos faltantes.

**Dependências:** BK-001

**Fonte:** `docs/source/README.md`; `docs/evidence/MOCAP-T1-T15-REALIGNMENT-2026-09-20.md`; `docs/FUNCTIONAL-DECOMPOSITION-v1.md`; `G0-DECISION-REGISTER.md`; hashes das 15 referências conferidos em 10/10/2026

**Evidência:** As 15 referências do inventário existem e os 15 SHA-256 conferem; o MOCAP mapeia 15 rotas e registra Balanço Patrimonial fora do escopo, ausência de rota separada “Aplicações” e correções aprovadas da T6. A inspeção visual das imagens T1–T15 foi revalidada em 10/10/2026; discrepâncias e ambiguidades têm destino registrado em [MOCAP-SCREEN-AUDIT-2026-10-10.md](../../evidence/MOCAP-SCREEN-AUDIT-2026-10-10.md). Os valores dos mockups são ilustrativos, não regras. A responsável confirmou que Compras deve apenas mostrar os dados, sem exportar, baixar ou criar/alterar pedidos; a aparência do botão histórico “Gerar pedidos” não será alterada sem aprovação visual explícita. Contratos de cálculo continuam vinculados às tarefas financeiras específicas.

**Regras e verificações vinculadas:** Q-01,Q-02,F-13

### BK-005 — Detalhar os casos de pagamento divergente após decidir C-01

**Situação:** Concluída. **Evolução:** 100%. **Tipo:** Planejamento. **Responsável sugerido:** Majucau / Financeiro.

**Objetivo:** Ter uma única semântica de baixa de títulos.

**O que fazer:** Formalizar a decisão: valor diferente gera divergência e mantém título sem baixa até análise.

**Como fazer:**

1. Comparar o valor do título com o movimento efetivamente registrado e comprovado no Bling. Valor igual pode seguir a conciliação normal se o vínculo for único e sustentado por evidência.
2. Usar estes exemplos de aceite para C-01: título de R$ 1.000,00 com movimento de R$ 999,99, R$ 1.000,01 ou R$ 600,00 = `DIVERGENT`; manter o título sem baixa até análise. A diferença de um centavo já é divergência; não criar tolerância.
3. Preservar desconto, tarifa e estorno como eventos separados conforme a origem. Se houver diferença líquida, reversão ou dúvida de vínculo, deixar pendente para análise; não compensar ou ajustar automaticamente.
4. Se mais de um movimento puder corresponder ao título, não escolher por aproximação; manter a conciliação pendente de revisão humana. Reprocessar o mesmo arquivo/evento não deve duplicar o movimento.
5. Registrar título, movimento e origem, valores comparados, estado, motivo da divergência e revisão realizada. O Majucau não altera nem baixa títulos no Bling.

**Critério de aceite:** Valor igual e vínculo único seguem a conciliação normal com evidência. Qualquer diferença, inclusive R$ 0,01, mantém `DIVERGENT` e o título sem baixa. Desconto/tarifa/estorno não são compensados por hipótese; vínculo ambíguo fica para revisão humana; replay não duplica fatos; não há escrita no Bling.

**Impacto:** Afeta contas a pagar, vencidos, caixa e conciliação.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** C_Conflitos!A3:L3; T04; Fechamento vigente 13/09/2026, §8.6–8.8; C-01 mantida

**Evidência:** Cenários de aceite e tratamento de ambiguidade registrados nesta tarefa a partir da regra C-01 aprovada; não houve alteração de código nem execução de testes.

**Regras e verificações vinculadas:** C-01

### BK-006 — Especificar tecnicamente a regra fechada da calculadora

**Situação:** Em andamento. **Evolução:** 30% (estimativa de planejamento; a estrutura da planilha enviada foi preservada. D-003-A e a sugestão separada D-003-C continuam vigentes. O Data Dictionary foi alinhado à regra aprovada: menor fechamento diário D0–D+60, fluxo consolidado e reserva mínima de R$ 50.000. O mapa estático localizou as estruturas previstas para DRE, projeção e saldo bancário, mas não encontrou chamadas de produção. O schema não contém histórico de aportes/resgates; a origem do principal líquido investido e a classificação do saldo transferível do Itaú continuam pendentes. Nenhum dado ausente será convertido em zero. A ligação com fontes reais e exemplos reproduzíveis ainda não existem). **Tipo:** Planejamento. **Responsável sugerido:** Majucau / Financeiro.

**Objetivo:** Permitir um cálculo reproduzível e compreensível.

**O que fazer:** Transcrever o contrato consolidado da calculadora e os estados das entradas.

**Como fazer:**

1. Usar o resultado líquido mensal calculado pela DRE do Majucau; preservar a janela de janeiro até dois meses antes do mês selecionado (M−2). A planilha atual recebe os meses manualmente e não calcula esse acumulado; o sistema deve fazê-lo a partir dos resultados mensais comprovados.
2. Calcular Já Investido = aportes de principal − resgates de principal; manter rendimentos separados. O campo manual “Valor já investido” da planilha atual não é usado por suas fórmulas e não substitui o cálculo aprovado. A reserva mínima continua em R$ 50.000. Exibir, apenas como sugestão informativa, `reserva_sugerida_hoje = R$ 50.000 + MAX(0; caixa projetado em D0 − menor saldo diário projetado em D0–D+60)`. A sugestão não altera D-003-A nem o valor máximo para aplicar.
3. Seguir D-003-A: `worst_week_cash` representa o menor saldo diário de fechamento entre D0 e D+60. Usar somente o fluxo de caixa consolidado, com fonte oficial e data válida; não somar recebimentos da Nuvemshop separadamente. Na planilha atual, E15 é um saldo de semana digitado e E16 é uma previsão da Nuvemshop somada separadamente; nenhum deles comprova os 61 saldos diários da regra aprovada. A equivalência desses campos continua pendente e não substitui D-003-A.
4. Calcular `available_profit = MAX(0; lucro líquido acumulado de janeiro até M−2 − Já Investido)`, `financial_limit = MAX(0; menor saldo diário D0–D+60 − reserva mínima)` e `Máximo Aplicável = MAX(0; MIN(available_profit; financial_limit))`, com moeda decimal exata.
5. Calcular separadamente `Valor Executável Hoje = MIN(Máximo Aplicável; saldo transferível do Itaú)`, conforme R-08; exibir o saldo projetado restante. A tela é somente para consulta e não executa transferência ou aplicação.
6. A calculadora recebe resultados financeiros consolidados, com período, fonte e qualidade. Bandeira, MDR, parcelamento e tarifas de cartão não entram nela.
7. Mapear os campos técnicos que comprovam o saldo diário de fechamento, D0, D+60 e recebimentos elegíveis. Se uma entrada obrigatória não puder ser comprovada, mostrá-la como indisponível e pedir orientação; não inventar equivalência com a planilha nem substituir dado ausente por zero.

**Critério de aceite:** Contrato financeiro fechado transcrito com referências e exemplos técnicos; sugestão de reserva aparece separada e não interfere no máximo; nenhuma entrada ausente é inventada. A execução do cálculo depende das entradas comprovadas, sem exigir integração pronta para concluir a especificação.

**Impacto:** Evita uma capacidade de aplicação baseada em critérios indefinidos.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** `G0-DECISION-REGISTER.md` (D-003-A e D-003-C); `docs/evidence/APPLICATION-RULE-D003A-2026-09-20.md`; `docs/evidence/INVESTMENT-RESERVE-SUGGESTION-2026-10-10.md`; `docs/source/T8-mockup-2026-09-20.png`; C_Conflitos!A4:L5; A_Matriz_Mestra!A32:S36; T02/T08; Consolidação 13/09/2026, §§6–9; Fechamento vigente 13/09/2026, §§1–2

**Regras e verificações vinculadas:** L-07,Q-10,Q-21,R-06,R-08,F-01

### BK-007 — Especificar as regras da DRE própria do Majucau

**Situação:** Em andamento. **Evolução:** 20% (estimativa de planejamento; os princípios enviados por Gisele estão organizados em `docs/DRE-ACCOUNTING-POLICY-v1.md`, incluindo estrutura de classificação, estoque/CPV, máquinas/depreciação, tributos recuperáveis e dados mínimos. Essa documentação preserva as linhas e a tela aprovadas. Ainda faltam validação com a contabilidade, critérios específicos da empresa, mapeamento de contas/eventos e conferência com dados reais). **Tipo:** Planejamento. **Responsável sugerido:** Arquitetura + Integrações + Contabilidade.

**Objetivo:** Calcular uma DRE própria, com classificação e competência aprovadas pela contabilidade.

**O que fazer:** Validar e completar as regras contábeis específicas da empresa e definir como os lançamentos de origem alimentam as linhas já aprovadas da DRE, usando `docs/DRE-ACCOUNTING-POLICY-v1.md` como base de trabalho.

**Como fazer:**

1. Conferir com a contabilidade os princípios recebidos e decidir critérios pendentes (inclusive valoração do estoque, rateio, tributos e imobilizado); mapear tipos de lançamento às linhas já existentes da DRE, sem inventar classificação ou fórmula.
2. Mapear os campos e contas de origem realmente disponíveis no Bling e nas demais fontes homologadas, sem alterar os sistemas de origem.
3. Comparar o resultado calculado pelo Majucau com a DRE do Bling por período e conta; registrar diferenças, sem copiar o relatório pronto como resultado final.
4. Campo necessário não acessível pela fonte deve aparecer como Dado indisponível; ausência nunca vira zero.
5. Preservar competência dos rendimentos; aplicação e resgate de principal não são automaticamente receita/despesa. Não alterar resultado por fontes externas fora de seu escopo.

**Critério de aceite:** Regras específicas da empresa são validadas/versionadas e mapeadas às linhas existentes; CPV/CMV concilia estoque, produção e vendas; entradas, cálculos e diferenças podem ser rastreados até a fonte. Sem regra, documento ou campo aprovado, o valor dependente permanece indisponível. Tela aprovada não é alterada.

**Impacto:** Previne resultado incompleto apresentado como definitivo.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** D_Lacunas; M-017..M-029; Consolidação 13/09/2026, pendências 4–6; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** L-03,L-04,L-05,L-08,L-09,L-13,L-14,Q-09,HT-04,HT-06,F-02,F-03

### BK-008 — Especificar necessidade líquida e acompanhamento da produção

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Planejamento. **Responsável sugerido:** Majucau / Operação.

**Objetivo:** Eliminar critérios operacionais ambíguos.

**O que fazer:** Traduzir as regras funcionais fechadas em contratos e exemplos verificáveis.

**Como fazer:**

1. Usar estoque utilizável, comprometimento aplicável, demanda, BOM, produção válida, saldo pendente dos pedidos e chegada em tempo.
2. Necessidade líquida desconta somente cobertura efetiva: 100 kg − 20 kg − 30 kg = 50 kg. Pedido parcial não exclui o item e pedido tardio não cobre a necessidade daquela data.
3. Separar estado temporal (imediata/futura) de cobertura (coberta/parcial/não coberta); priorizar risco operacional e impacto de produção, com critérios explicáveis.
4. Meta atingida exige realizado >= meta; em andamento e abaixo da meta dependem do avanço esperado no período/plano. Sem granularidade temporal, não inventar atraso ou tolerância.

**Critério de aceite:** Exemplos cobrem pedido parcial, tardio, cancelado, atendido e ausência de dado; custo/tarifa/cacau não determinam quantidade necessária.

**Impacto:** Evita compra duplicada e avaliação incorreta da produção.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** T09/T10; Q_CEI_Tela_11; Consolidação 13/09/2026, §3; Fechamento vigente 13/09/2026, §§3–4

**Regras e verificações vinculadas:** L-17,L-18,Q-07,Q-08,F-04,F-05,F-09

### BK-009 — Especificar forecast, metas e planejamento de reposição

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Planejamento. **Responsável sugerido:** Arquitetura + Dados.

**Objetivo:** Definir quando uma projeção pode ser utilizada.

**O que fazer:** Separar previsão baseada em vendas, meta comercial, produção e trajetória de estoque.

**Como fazer:**

1. Usar histórico confiável de 2025 em diante com cobertura real por SKU; não completar 24 meses nem converter ausência em venda zero.
2. Manter meta de faturamento como cenário comercial; não distribuí-la automaticamente em unidades por SKU. Meta direta por produto só é entrada quando cadastrada explicitamente.
3. Normalizar tecnicamente a expressão de reposição de §4.5 como demanda + estoque-alvo − estoque utilizável − produção válida ainda não incorporada ao estoque. O marcador de lista antes de estoque-alvo no texto não é interpretado como multiplicação; registrar esta normalização.
4. Calcular cronologicamente, sem descontar duas vezes produção que já integra o saldo. Estoque-alvo, produção válida e demais entradas exigem dados efetivos; ausência não assume zero.
5. Propor e testar modelos, métricas e critérios técnicos com histórico real; registrar limitações. T01 e T10 compartilham a métrica; cenário T12 não sobrescreve automaticamente meta operacional.

**Critério de aceite:** Contratos separam fatos, previsão e planejamento; fórmula técnica tem unidade e período coerentes; receita alvo não cria demanda física fictícia.

**Impacto:** Impede metas irreconciliáveis e métricas de qualidade fictícias.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** R_CEI_Tela_12; T12; M-045..M-049; Consolidação 13/09/2026, pendência 9; Fechamento vigente 13/09/2026, §4

**Regras e verificações vinculadas:** L-20,L-21,Q-11,Q-12,R-13,HT-09,F-10,F-11

### BK-010 — Homologar contratos do Bling e das três estruturas do Google Drive

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Planejamento. **Responsável sugerido:** Engenharia de Integrações + responsável pelos arquivos.

**Objetivo:** Conhecer os formatos que o sistema realmente receberá.

**O que fazer:** Mapear Bling/API e arquivos de perdas, Nuvem Pago e Nuvem Envio na pasta compartilhada Majucau.

**Como fazer:**

1. Manter as três estruturas do Drive somente de leitura e contratos específicos: perdas, agenda Nuvem e Nuvem Envio.
2. Mapear amostras reais, colunas, IDs, sinais, período e versões; registrar processamento e rejeições internamente, sem mutar o Drive.
3. Definir a detecção técnica de novas versões e idempotência por arquivo, hash, chave de negócio e lote.
4. Aplicar D-1 à análise/conciliação normal de movimentos realizados. Frequência de sincronização é configuração técnica distinta; agenda futura preserva seus vencimentos.

**Critério de aceite:** Contratos identificam somente as três estruturas externas aprovadas, seus arquivos reais e o Bling. Amostras e acesso ainda precisam ser obtidos; não há nova pergunta sobre a escolha da pasta ou permissão de escrita.

**Impacto:** Evita importações que duplicam ou perdem dados.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** H_Contrato_Dados; U_CEI_Tela_15; Consolidação 13/09/2026, §§1–2,10–12; Fechamento vigente 13/09/2026, §8

**Regras e verificações vinculadas:** L-11,L-12,L-24,Q-06,R-01,R-02,R-10,HT-08,F-18,F-19

### BK-011 — Definir fontes e critérios dos sinais do cacau

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Planejamento. **Responsável sugerido:** Engenharia de Dados.

**Objetivo:** Tornar os sinais externos comparáveis e rastreáveis.

**O que fazer:** Homologar fontes de preço, câmbio, clima, safra e estoques globais.

**Como fazer:**

1. Relacionar fornecedores pequenos produtores às localidades efetivamente usadas pela Majucau; não criar regiões fictícias.
2. Refinar tecnicamente fontes de preços/clima, unidades, datas, cobertura e comparabilidade, mantendo contexto separado da necessidade operacional.
3. Especificar ausência/atraso dos sinais sem impedir o funcionamento do motor principal de compras; não ampliar as estruturas do Drive por inferência.

**Critério de aceite:** Contrato separa necessidade, histórico/preço atual e sinais externos; refinamento dos sinais não é bloqueio da recomendação operacional.

**Impacto:** Evita oportunidade ou risco calculados sobre bases incompatíveis.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** Q_CEI_Tela_11; H_Contrato_Dados!A21:H22; Consolidação 13/09/2026, pendência 10; Fechamento vigente 13/09/2026, §3.4

**Regras e verificações vinculadas:** L-19,R-14,HT-10,F-07

### BK-012 — Especificar ações e exportações já previstas

**Situação:** Em andamento. **Evolução:** 10% (auditoria estática inicial dos controles atuais e das chamadas disponíveis, registrada em `docs/evidence/ACTION-EXPORT-STATIC-AUDIT-2026-10-11.md`. A matriz tela por tela, as rotas de exportação e os contratos de detalhe/atualização ainda faltam; para T13–T15, L-22/Q-13 continuam abertos). **Tipo:** Planejamento. **Responsável sugerido:** Majucau + UX.

**Objetivo:** Especificar o que cada botão pode fazer.

**O que fazer:** Mapear os controles aprovados a casos de uso, dados e capacidades.

**Como fazer:**

1. Preservar botões, navegação e exportações dos mockups/contratos congelados; não pedir nova aprovação da função existente. Exceção aprovada em 10/10/2026: T11 Compras somente mostra dados e não exporta nem baixa arquivos.
2. Definir tecnicamente contratos de detalhes, atualização e downloads somente onde a exportação estiver aprovada, preservando filtros, período, snapshot, dados e Excel onde especificado. T11 não oferece download.
3. Resolução de pendência registra responsável, data, alteração, origem e justificativa; resolver exige tratar a causa e não apenas esconder o alerta.

**Critério de aceite:** Cada controle aprovado tem caso de uso e capacidade associados; exportações autorizadas reproduzem a tela. T11 permanece somente para consulta, sem exportação/download. Ausência de arquivo de referência é localização documental, sem função inventada.

**Impacto:** Evita botões decorativos e encerramento indevido de problemas.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** S/T/U CEIs; T02–T07; T13–T15; `G0-DECISION-REGISTER.md` (decisão posterior de T11 em 10/10/2026); Fechamento vigente 13/09/2026, §§5–6

**Regras e verificações vinculadas:** L-22,L-23,Q-13,F-14

### BK-013 — Especificar perfis, capacidades, vigência e auditoria

**Situação:** Em andamento. **Evolução:** 15% (estimativa documental: schema, permissões iniciais, caminho de autorização do worker e capacidades aprovadas; login próprio do Majucau, endereços por perfil e uso aprovado de uma conta compartilhada para Diretores, com limitação de auditoria registrada. A migration `000003_director_report_export_permission.up.sql` inclui `REPORT_EXPORT` para `ADMIN` e `DIRECTOR` e foi aplicada pela CI `38113352327`. A migration ainda usa `windows_sid` e a autorização não é aplicada no worker. A importação manual de CSV do Bling e da Nuvem Pago foi aprovada para Diretores como parte de `INTEGRATION_SYNC`; MFA, sessão, recuperação, provisionamento e enforcement continuam pendentes). **Tipo:** Planejamento. **Responsável sugerido:** Arquitetura + Segurança.

**Objetivo:** Definir quem pode consultar ou alterar cada recurso.

**O que fazer:** Definir tecnicamente autorização por usuário, perfil e capacidade, com menor privilégio.

**Como fazer:**

1. Separar visualizar, exportar, alterar metas, alterar parâmetros, resolver pendências, reprocessar, administrar usuários e consultar auditoria; não hardcodar nomes de diretores.
2. Restringir criação, perfil, ativação e inativação à capacidade administrativa; usuário comum não promove a própria autoridade.
3. Especificar login, sessão, revogação e provisionamento com atribuição explícita de perfis na implantação, sem tornar nomes pessoais requisitos de arquitetura.
4. Alterações usam vigência e histórico antes/depois. Período fechado só muda por reprocessamento explícito, autorizado e auditado.

**Critério de aceite:** Matriz por capacidade e protocolo de atribuição definidos tecnicamente; consulta não implica alteração e ações sensíveis conservam evidência.

**Impacto:** Reduz acesso indevido e mudança de regra sem rastreio.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** W/Y CEIs; H_Contrato_Dados!A34:H39; Fechamento vigente 13/09/2026, §6; auditoria estática `docs/evidence/ACCESS-CONTROL-AUDIT-2026-10-11.md`

**Regras e verificações vinculadas:** L-25,L-26,Q-14,F-15

### BK-014 — Inventariar o Windows e propor a operação local

**Situação:** Em andamento. **Evolução:** 15% (inventário somente para leitura do ambiente Windows usado nesta sessão e verificação do fim do suporte regular do Windows 10 Pro 22H2. A situação de inscrição ESU não foi verificada, e a máquina não foi presumida como host de produção; capacidade de energia/rede, direitos de instalação e escolha/validação do host operacional permanecem pendentes). **Tipo:** Planejamento. **Responsável sugerido:** Infraestrutura + Segurança.

**Objetivo:** Dimensionar uma instalação que a Majucau consiga operar.

**O que fazer:** Levantar o host existente e documentar solução operável por serviços.

**Como fazer:**

1. Verificar tecnicamente host, Windows, CPU, RAM, disco, energia e rede; não inventar características não acessíveis.
2. Propor serviços independentes de login humano, inicialização controlada, banco, portas, dependências, atualização e tratamento de falta de internet.
3. Propor backup, retenção, recuperação, monitoramento, suporte e acesso autenticado. Acesso externo permanece condicionado a habilitação explícita e canal seguro.
4. Registrar requisitos de funcionamento para validação da Majucau na implantação; obter permissões/acessos materiais quando necessários, sem devolver decisões de infraestrutura como perguntas financeiras.

**Critério de aceite:** Inventário real e proposta técnica documentados; limitações de acesso/hardware são evidenciadas, sem depender de sessão de funcionário.

**Impacto:** Evita depender de um computador sem capacidade ou sem recuperação.

**Dependências:** Sem tarefa prévia; observar contrato e evidências aplicáveis.

**Fonte:** Pedido: hospedagem local; Fechamento vigente 13/09/2026, §7

**Regras e verificações vinculadas:** Q-03,Q-04,Q-05,Q-15,Q-16,F-17

### BK-015 — Organizar responsáveis, estimativas e marcos de entrega

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Planejamento. **Responsável sugerido:** Majucau + Liderança técnica.

**Objetivo:** Manter evolução visível sem prometer prazos sem base.

**O que fazer:** Revisar dependências e estimar esforço com a equipe.

**Como fazer:**

1. Propor esforço, sequência e responsáveis por papel conforme capacidade técnica conhecida; não prometer prazo inventado.
2. Preservar IDs, subdividir entregas quando necessário e registrar pesos de esforço apenas com base explícita.
3. Planejar marcos internos até a primeira produção e critérios de aceite; refinamento dos sinais de cacau não impede testar/usar o motor principal de compras.

**Critério de aceite:** Backlog sem dependências circulares, dono por tarefa e cronograma apoiado em capacidade real.

**Impacto:** Permite acompanhar avanço e antecipar bloqueios.

**Dependências:** BK-000, BK-004, BK-014

**Fonte:** Pedido do usuário; G_Ordem; Fechamento vigente 13/09/2026, §7

**Regras e verificações vinculadas:** Q-17

## 01 Arquitetura e UX

### BK-016 — Modelar os domínios e o vocabulário do negócio

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Arquitetura.

**Objetivo:** Separar regras sem fragmentar a aplicação em serviços.

**O que fazer:** Desenhar contextos do monólito e suas responsabilidades.

**Como fazer:**

1. Modelar Tesouraria, Conciliação, Resultados, Operação, Planejamento, Integrações e Acesso
2. Definir entidades, agregados, invariantes e eventos
3. Compartilhar contratos e IDs, evitando modelos únicos gigantes
4. Separar eventos de perdas operacionais de rendimento industrial; separar venda, direito a receber, liquidação, aporte, resgate, rendimento realizado e posição comprovada.
5. Separar compras por quantidade, apropriação de custo e inteligência de cacau. O motor Majucau calcula a DRE segundo regras contábeis aprovadas, e o Bling é fonte dos lançamentos e referência de comparação. A calculadora recebe apenas contratos financeiros consolidados. Uma métrica tem um dono e um cálculo para todos os consumidores.

**Critério de aceite:** Mapa mostra dono de cada fato, dependências direcionais e transações de negócio.

**Impacto:** Evita regras duplicadas entre telas e módulos.

**Dependências:** BK-000, BK-005, BK-006, BK-007, BK-008, BK-009

**Fonte:** A_Matriz_Mestra; B_Intersecoes; H_Contrato_Dados; Consolidação 13/09/2026, §§6,10–12; Fechamento vigente 13/09/2026, §§1–4

### BK-017 — Definir a estrutura hexagonal do monólito Go

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Arquitetura.

**Objetivo:** Manter regras testáveis e independentes das integrações.

**O que fazer:** Registrar portas, casos de uso e adaptadores.

**Como fazer:**

1. Separar domínio, aplicação e adaptadores por módulo
2. Definir portas de repositório, relógio, arquivos e provedores
3. Documentar composição única, fronteiras de transação e dependências permitidas

**Critério de aceite:** Domínio não depende de HTTP, React, ORM ou clientes Bling; módulos se comunicam por contratos.

**Impacto:** Facilita manutenção e substituição de integração.

**Dependências:** BK-016

**Fonte:** Stack definida pelo usuário

### BK-018 — Definir persistência, migrações e precisão dos dados

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Arquitetura + Dados.

**Objetivo:** Preservar dinheiro, quantidades e histórico corretamente.

**O que fazer:** Avaliar PostgreSQL local e modelo relacional por domínio.

**Como fazer:**

1. Definir valores monetários exatos e arredondamento por regra
2. Modelar chaves, índices, constraints, versões e transações
3. Escolher migrações e política de compatibilidade de versões

**Critério de aceite:** Modelo diferencia ausência de zero e proíbe duplicatas de negócio; decisão de banco é registrada.

**Impacto:** Evita erros de centavos e corrupção lógica de fatos.

**Dependências:** BK-016, BK-010

**Fonte:** H_Contrato_Dados; proposta técnica

### BK-019 — Descrever contratos de API e estados de leitura

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Full Stack.

**Objetivo:** Unir React e Go com respostas previsíveis.

**O que fazer:** Documentar endpoints, ações, filtros e erros em OpenAPI.

**Como fazer:**

1. Definir paginação, ordenação, datas, moeda e unidade
2. Incluir posição, origem, qualidade, versão e autorização
3. Diferenciar validação, falta de dado, conflito, acesso negado e falha de fonte
4. Definir portas distintas para demonstrativos Bling, caixa projetado e disponibilidade financeira. A porta da calculadora recebe lucro elegível, caixa crítico, reserva, aportes líquidos e saldo transferível, com versão/qualidade; detalhes de cartão ficam nos módulos de origem.

**Critério de aceite:** Exemplos por tela permitem testar cliente e servidor antes da integração real.

**Impacto:** Reduz retrabalho de interface e interpretação incorreta.

**Dependências:** BK-017, BK-012, BK-013

**Fonte:** H_Contrato_Dados; todos CEIs; Fechamento vigente 13/09/2026, §1

**Regras e verificações vinculadas:** F-01

### BK-020 — Modelar estado, proveniência e consistência entre telas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Explicar a confiabilidade de cada número.

**O que fazer:** Separar status do negócio, qualidade do dado e processamento.

**Como fazer:**

1. Mapear CONFIRMED, PROJECTED, STALE, UNAVAILABLE e demais estados
2. Definir cobertura, data de posição e snapshot consistente
3. Definir propagação de qualidade sem transformar ausência em zero
4. Aplicar uma função e um conjunto de dados/snapshot por métrica em card, tabela, gráfico, detalhe e exportação. PENDING_RECONCILIATION fica fora de indicadores que exigem reconciliação.

**Critério de aceite:** Mesma consulta e posição produzem mesmos totais; dados parciais e antigos ficam identificados. Ausência não vira zero; ausência de campo opcional não invalida componentes independentes comprovados.

**Impacto:** Impede que uma tela pareça atualizada quando sua fonte falhou.

**Dependências:** BK-018, BK-019

**Fonte:** M-002,M-037,M-038; H_Contrato_Dados; Fechamento vigente 13/09/2026, §§4.3,8.7–8.8

**Regras e verificações vinculadas:** F-08,F-20

### BK-021 — Transcrever o design congelado em componentes

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** UX.

**Objetivo:** Traduzir os mockups para componentes consistentes.

**O que fazer:** Implementar especificação técnica fiel ao mockup aprovado, sem redesenho.

**Como fazer:**

1. Capturar estrutura, blocos, cabeçalho, menu, cards, gráficos, tipografia, cores, espaçamentos, botões e hierarquia da última versão comprovadamente aprovada.
2. Separar contrato visual do funcional: números demonstrativos não são regra; fonte e cálculo seguem o fechamento vigente.
3. Preparar componentes reutilizáveis mantendo posição e hierarquia; erro futuro de implementação se corrige no código, sem reabrir a tela congelada.

**Critério de aceite:** Especificação preserva o visual oficial; diferenças entre arquivos exigem localizar a versão aprovada, sem escolher versão antiga ou redesenhar por conveniência.

**Impacto:** Corrige inconsistência de navegação sem redesenho silencioso.

**Dependências:** BK-003, BK-004, BK-012

**Fonte:** Todas as imagens; M-039,M-050; Fechamento vigente 13/09/2026, §5

**Regras e verificações vinculadas:** Q-01,Q-02,Q-18,F-13

### BK-022 — Especificar acessibilidade, telas pequenas e estados de exceção

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** UX + QA.

**Objetivo:** Permitir uso além da imagem estática.

**O que fazer:** Desenhar loading, vazio, erro, bloqueado e dados antigos.

**Como fazer:**

1. Especificar carregamento, vazio, falha, indisponibilidade e acesso negado dentro dos blocos e padrões aprovados.
2. Implementar teclado, foco, rótulos, zoom e leitura de tabelas preservando hierarquia e disposição.
3. Conferir referência congelada; não autorizar mudança estrutural ou responsiva que redesenhe a tela por iniciativa técnica.

**Critério de aceite:** Fluxos completos cobrem falhas e permissões; nenhuma informação depende só da cor.

**Impacto:** Reduz erros de operação e dificuldade no uso diário.

**Dependências:** BK-021, BK-014

**Fonte:** Todas as imagens; gaps UX; Fechamento vigente 13/09/2026, §5

**Regras e verificações vinculadas:** Q-18,F-13

### BK-023 — Decidir a necessidade de Python e Streamlit

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Arquitetura + Dados.

**Objetivo:** Adicionar ferramentas somente quando houver benefício verificável.

**O que fazer:** Planejar avaliação de importação e projeções com dados representativos.

**Como fazer:**

1. Comparar processamento Go com Python para cargas e séries reais
2. Avaliar pandas ou Polars conforme memória, formato e tempo medidos
3. Manter React como interface principal; documentar eventual uso interno de Streamlit

**Critério de aceite:** Decisão registra ganho, custo de operação e contrato de entrada/saída; sem segundo motor financeiro.

**Impacto:** Controla complexidade e mantém uma fonte de regras.

**Dependências:** BK-017, BK-009, BK-010, BK-014

**Fonte:** Sugestão permitida pelo usuário; proposta técnica

### BK-024 — Planejar segurança e proteção dos dados

**Situação:** Em andamento. **Evolução:** 30%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Segurança + Arquitetura.

**Objetivo:** Identificar falhas possíveis antes de expor o sistema.

**O que fazer:** Mapear acessos, arquivos, credenciais e fronteiras da rede.

**Como fazer:**

1. Classificar informações e privilégios
2. Avaliar importação maliciosa, acesso direto à API e exposição de segredos
3. Definir retenção, restauração e controle de dados usados em testes
4. Projetar identidade técnica do Drive com permissão somente de leitura, restrição às estruturas autorizadas e credenciais apenas no servidor; escrita interna não concede escrita na origem.

**Critério de aceite:** Riscos têm tratamento e testes associados; nenhuma credencial entra no Git ou no navegador. O desenho não usa a pasta de fontes como destino de exportação, logs ou backup.

**Evidência atual:** [Mapa inicial de ativos, riscos, controles, lacunas e verificações](../../evidence/THREAT-MODEL-INITIAL-2026-10-11.md). Inclui revisões estáticas e verificações no instalador, backups, Named Pipe, DPAPI, temporários, diagnósticos, mutex e propagação de erros no encerramento de importações em streaming. Os testes com PostgreSQL descartável passaram nas [CI #403](https://github.com/guilhermejacyzin/majucau/actions/runs/38135368926) e [CI #408](https://github.com/guilhermejacyzin/majucau/actions/runs/38138263712); a revisão anterior do Named Pipe passou na [CI #406](https://github.com/guilhermejacyzin/majucau/actions/runs/38137190197). A CI #410 passou nos quatro jobs, com testes Go, `go vet`, análise de vulnerabilidades, builds Windows do desktop/worker, SBOM, smoke do instalador, verificação do backlog e consolidação de outputs. Os testes locais de `securetemp` e `csvstream` também passaram. Ainda faltam outras falhas controladas, fechar o mapa de riscos e validar instalação/atualização em VM limpa, rede e acesso remoto por perfil.

**Impacto:** Protege os dados financeiros e operacionais.

**Dependências:** BK-013, BK-014, BK-019

**Fonte:** M-058; Y_CEI_Tela_17; Consolidação 13/09/2026, §2

**Regras e verificações vinculadas:** R-02

## 02 Fundação do projeto

### BK-025 — Preparar o repositório e as convenções do projeto

**Situação:** Em andamento. **Evolução:** 70% (estimativa de planejamento; README, `.gitignore`, estrutura e `AGENTS.md` já registram instalação, separação do projeto, proteção de segredos e regras essenciais. `CONTRIBUTING.md` explica o fluxo seguro de mudanças, commits e evidências. `scripts/verify-backlog.ps1` e seu teste de divergência passaram na CI Windows, e a CI completa do commit `a61dd71` concluiu com sucesso. A consulta pública à API do GitHub confirmou `protected: false` para `main`; os jobs de CI existentes são `verify-windows` e `verify-postgres`, mas ainda não são exigidos pela branch. Falta uma pessoa administradora configurar as proteções e verificações obrigatórias). **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Criar uma base reproduzível para a equipe.

**O que fazer:** Inicializar estrutura, documentação e políticas do Git.

**Como fazer:**

1. Criar README, ignore de segredos e regras de contribuição
2. Organizar backend, frontend, docs e infraestrutura conforme decisão
3. Configurar revisão e permissões do repositório

**Critério de aceite:** Novo desenvolvedor encontra instruções e convenções; histórico contém somente artefatos do projeto.

**Impacto:** Evita configuração informal e exposição de arquivos locais.

**Dependências:** BK-002, BK-017, BK-015

**Fonte:** GitHub vazio; BK arquitetura

**Evidência da auditoria de 2026-10-11:** `README.md`, `.gitignore`, `AGENTS.md`, `.github/workflows/ci.yml`, `CONTRIBUTING.md` e `docs/evidence/GITHUB-REPOSITORY-STATE-2026-10-11.md`. `git ls-files` não encontrou caches, artefatos gerados, arquivos de ambiente ou chaves nos caminhos pesquisados. A API pública confirmou que `main` está sem proteção; configurar a proteção e os status checks obrigatórios ainda depende de uma pessoa administradora.

### BK-026 — Preparar o ambiente de desenvolvimento local

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir execução reproduzível fora da produção.

**O que fazer:** Documentar versões e comandos de instalação, execução e limpeza.

**Como fazer:**

1. Fixar dependências e arquivos de lock aplicáveis
2. Criar configuração exemplo sem segredos e dados sintéticos
3. Validar início em máquina limpa

**Critério de aceite:** Ambiente sobe seguindo documentação e não acessa dados de produção por padrão.

**Impacto:** Diminui tempo perdido com diferenças entre máquinas.

**Dependências:** BK-025, BK-018, BK-014

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-027 — Criar a fundação do backend Go

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Dispor de processo executável com as fronteiras planejadas.

**O que fazer:** Implementar composição, configuração, servidor e ciclo de vida.

**Como fazer:**

1. Validar configuração ao iniciar
2. Implementar timeouts, cancelamento, desligamento e health/readiness
3. Montar módulos por injeção explícita de dependências

**Critério de aceite:** Processo inicia e encerra sem interromper operações silenciosamente; configuração inválida é identificada.

**Impacto:** Sustenta todas as funcionalidades.

**Dependências:** BK-026, BK-017

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-028 — Criar a fundação do frontend React JS

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Frontend.

**Objetivo:** Dispor de navegação e layout reutilizáveis.

**O que fazer:** Montar rotas, shell, estilos e cliente de API.

**Como fazer:**

1. Implementar menu e componentes comuns aprovados
2. Configurar tratamento de erro e sessão
3. Padronizar moeda, datas, unidade e tradução de status

**Critério de aceite:** Rotas e estados base funcionam por teclado; seleção do menu corresponde à página.

**Impacto:** Evita refazer componentes em cada tela.

**Dependências:** BK-026, BK-021, BK-022, BK-019

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-029 — Criar o esquema inicial e o controle de migrações

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Evoluir banco sem ajustes manuais improvisados.

**O que fazer:** Implementar migrações base, usuários técnicos e constraints.

**Como fazer:**

1. Separar permissões de aplicação e migração
2. Criar índices e versões iniciais
3. Validar aplicação e recuperação de migração em base descartável

**Critério de aceite:** Banco vazio fica pronto automaticamente; migração falha de modo diagnosticável.

**Impacto:** Reduz risco na instalação e atualização local.

**Dependências:** BK-018, BK-027

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-030 — Configurar verificações automáticas do projeto

**Situação:** Concluída. **Evolução:** 100%. A CI executa verificações de Go, frontend, PostgreSQL, segurança, dependências, builds e consolidação em runners hospedados pelo GitHub. A etapa do instalador unsigned só publica após sucesso das etapas anteriores; `success()` tem semântica oficial documentada e a CI `38115370056` comprovou o caminho de sucesso. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia + Infraestrutura.

**Objetivo:** Detectar regressões antes da entrega.

**O que fazer:** Criar pipeline de lint, testes, análise de dependências e build.

**Como fazer:**

1. Executar Go, React e testes contratuais apropriados
2. Bloquear publicação de artefato com verificações obrigatórias falhas
3. Separar CI de acesso à máquina de produção

**Critério de aceite:** Commit válido produz resultado rastreável; PR não confiável não executa na rede da Majucau.

**Impacto:** Reduz falhas e exposição da rede local.

**Dependências:** BK-027, BK-028

**Fonte:** Repositório público; recomendação GitHub Actions

**Reauditoria de 2026-10-11:** `.github/workflows/ci.yml` limita as permissões do token a `contents: read`, executa o job Windows em `windows-2025` e o PostgreSQL em `ubuntu-24.04`, sem runner da rede da Majucau. O workflow roda em `push` e `pull_request`. A etapa `Upload unsigned G1 build` usa `if: success()`. A CI `38115370056` passou nas etapas `Consolidate generated outputs`, `Verify output consolidation` e `Upload unsigned G1 build`; a semântica oficial de `success()` garante que uma falha anterior impede o envio. Evidência: `docs/evidence/CI-ARTIFACT-PUBLICATION-GATE-2026-10-11.md`.

### BK-031 — Criar logs e rastreio técnico básicos

**Situação:** Em andamento. **Evolução:** 80% (estimativa; logs do worker, correlação técnica básica, limites de tamanho, rotação e testes próprios passaram na CI Windows/PostgreSQL-Linux. Falta validar ACL em instalação/VM, aprovar uma referência de diagnóstico junto ao erro na tela e definir/medir fila somente se uma fila real for criada). **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Diagnosticar falhas sem expor dados sensíveis.

**O que fazer:** Padronizar IDs de requisição, erro, lote e operação.

**Como fazer:**

1. Implementar logs estruturados com mascaramento
2. Correlacionar API, job, lote e fonte
3. Definir métricas de latência, falhas e tamanho das filas

**Implementação parcial (2026-10-11):** `docs/evidence/WORKER-STRUCTURED-DIAGNOSTICS-2026-10-11.md`. O worker registra requisições com campos limitados e IDs de lote. A duração e falhas podem ser agregadas dos eventos; a fila não existe neste caminho síncrono e, portanto, ainda não possui métrica. O usuário vê códigos de erro, mas ainda não vê o ID único que permite filtrar uma requisição específica no log.

**Critério de aceite:** Erro visível ao usuário possui referência que encontra diagnóstico técnico sem tokens ou extratos completos.

**Impacto:** Facilita suporte e investigação.

**Dependências:** BK-027, BK-024

**Fonte:** Requisito técnico para entregar e operar o produto.

## 03 Identidade e configurações

### BK-032 — Implementar autenticação e administração inicial

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir entrada segura no ambiente local.

**O que fazer:** Construir login e integração com identidade aprovada.

**Como fazer:**

1. Criar primeiro administrador sem senha fixa no pacote
2. Implementar sessão, logout e expiração
3. Aplicar recuperação e MFA conforme fluxo homologado
4. Compatibilizar convites e recuperação com as ferramentas locais e o canal aprovado para três diretores.

**Critério de aceite:** Sessão expirada não acessa API; recuperação não permite tomar conta de outro usuário.

**Impacto:** Protege a entrada no sistema.

**Dependências:** BK-013, BK-029, BK-028, BK-024

**Fonte:** L-26; Y_CEI_Tela_17

**Regras e verificações vinculadas:** F-15

### BK-033 — Aplicar permissões no backend e na interface

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Garantir que cada perfil faça somente o autorizado.

**O que fazer:** Implementar matriz de recursos e ações.

**Como fazer:**

1. Aplicar negação por padrão nas rotas e casos de uso
2. Restringir consultas, exports e operações internas
3. Refletir permissões na UI sem depender dela para segurança
4. Controlar capacidades por usuário/perfil, inclusive reprocessar período fechado e solicitar inclusão do dia corrente na conciliação; nunca vincular autoridade ao nome de uma pessoa.

**Critério de aceite:** Chamada direta proibida é negada; perfil LEITURA não modifica metas, usuários ou parâmetros.

**Impacto:** Evita escalada de privilégios.

**Dependências:** BK-032, BK-019

**Fonte:** M-057,M-058; H_Contrato_Dados!A38:H38; Fechamento vigente 13/09/2026, §6

**Regras e verificações vinculadas:** F-15

### BK-034 — Registrar auditoria de acesso e mudanças

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Saber quem executou cada alteração material.

**O que fazer:** Persistir eventos com ator, alvo, data, resultado e motivo.

**Como fazer:**

1. Registrar mudanças de perfil, bloqueio, metas e parâmetros
2. Preservar antes/depois e vínculo com requisição
3. Restringir acesso à auditoria e aplicar retenção aprovada
4. Persistir usuário, data/hora, ação, entidade, antes/depois, origem e identificador da operação sensível; usuário comum não altera logs.

**Critério de aceite:** Histórico permanece consultável após bloqueio e rollback; nenhum segredo é registrado.

**Impacto:** Permite investigar ações e explicar alterações.

**Dependências:** BK-033, BK-031

**Fonte:** H_Contrato_Dados!A39:H39; Fechamento vigente 13/09/2026, §6.8

**Regras e verificações vinculadas:** F-16

### BK-035 — Implementar T17 Usuários e o ciclo de vida das contas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Administrar acessos de forma compreensível.

**O que fazer:** Construir lista, perfis, detalhes e convites aprovados.

**Como fazer:**

1. Calcular cartões e distribuição com população definida
2. Implementar convite, ativação, bloqueio e desbloqueio
3. Revogar sessões quando o bloqueio exigir e proteger o último administrador

**Critério de aceite:** Contagens conferem com filtros; bloqueado perde acesso; convite vencido não ativa conta.

**Impacto:** Mantém controle centralizado do acesso.

**Dependências:** BK-033, BK-034, BK-022

**Fonte:** T17; Y_CEI_Tela_17; T-CEI17-01..08

**Regras e verificações vinculadas:** F-15

### BK-036 — Implementar registro e vigência dos parâmetros

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Reproduzir resultados com a regra vigente em cada data.

**O que fazer:** Persistir parâmetros tipados, origem e histórico.

**Como fazer:**

1. Validar tipo, unidade e faixa por parâmetro
2. Versionar valor, regra, vigência e evidência
3. Proibir sobreposição de vigências e edição sem permissão
4. Versionar parâmetros e metas por vigência, preservando os períodos encerrados. Alteração atual não recalcula mês fechado; reprocessamento anterior exige ação explícita, autorizada e auditada, com versão anterior consultável.

**Critério de aceite:** Cálculo histórico recupera versão utilizada; zero válido não é confundido com ausência.

**Impacto:** Evita alterações retroativas silenciosas.

**Dependências:** BK-018, BK-033, BK-034

**Fonte:** M-056; H_Contrato_Dados!A34:H35; Fechamento vigente 13/09/2026, §6.3–6.7

**Regras e verificações vinculadas:** F-16

### BK-037 — Implementar T16 Parâmetros e fluxos autorizados de edição

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar regras vigentes e suas fontes.

**O que fazer:** Construir central de consulta e edições específicas aprovadas.

**Como fazer:**

1. Preservar o layout T16 e mostrar valores, origem, estado e vigência reais.
2. Permitir alteração somente pela capacidade correspondente, validando tipo/unidade e registrando usuário, valor anterior/novo e data.
3. Aplicar a vigência registrada, mantendo períodos encerrados; reprocessamento retroativo passa por ação explícita e auditada.
4. Apresentar fontes atuais Bling e três estruturas Drive; eliminar a informação funcional superada de investimento em pasta externa.

**Critério de aceite:** Mudanças autorizadas preservam histórico e vigência; usuário de leitura não altera regras; a tela conserva o design congelado.

**Impacto:** Ajuda a Majucau a entender o efeito de cada configuração.

**Dependências:** BK-036, BK-021

**Fonte:** T16; W_CEI_Tela_16; T-CEI16-01..08; Consolidação 13/09/2026, §§1,4–5; Fechamento vigente 13/09/2026, §6

**Regras e verificações vinculadas:** F-16

## 04 Dados e integrações

### BK-038 — Implementar proveniência e snapshots de fatos

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Rastrear indicadores até a origem.

**O que fazer:** Criar metadados e consultas de linhagem.

**Como fazer:**

1. Persistir origem, IDs, lote, datas, posição e versão
2. Vincular fatos derivados às entradas
3. Preservar histórico de correções e versões
4. Guardar internamente nome do arquivo, identificador disponível, origem, hash, lote, detecção, processamento, período, status, total/aceitos/rejeitados, motivos e vínculo entre reprocessamentos.
5. Preservar snapshots e registros de fechamento de períodos. Reprocessamento explícito cria nova versão vinculada à anterior com ator, motivo, origem e operação; não apagar o histórico encerrado.

**Critério de aceite:** Um indicador chega ao registro e à regra que o produziram; reprocessamento preserva versões. Correções mantêm versão anterior e explicam seu efeito sobre os fatos; metadados não são gravados no Drive.

**Impacto:** Permite conferir qualquer número relevante.

**Dependências:** BK-020, BK-029, BK-034

**Fonte:** M-037; H_Contrato_Dados!A3:H3; Consolidação 13/09/2026, §2; Fechamento vigente 13/09/2026, §6.6–6.8

**Regras e verificações vinculadas:** R-02,F-16,F-20

### BK-039 — Implementar execução controlada de trabalhos em segundo plano

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Processar integrações sem travar consultas.

**O que fazer:** Criar agendamento, tentativas e recuperação de jobs.

**Como fazer:**

1. Persistir estado, checkpoint e tentativa
2. Impedir execução concorrente duplicada da mesma carga
3. Aplicar cancelamento, timeout e retomada após reinício
4. Separar agendamento da coleta, data da competência e corte de análise D-1. Movimentos de hoje podem ser ingeridos e armazenados sem compor a conciliação normal; inclusão requer solicitação expressa, capacidade e trilha.
5. Programar a atualização diária para 09:30, após o tempo para colocar os arquivos externos nas pastas. Confirmar o fuso horário e a ação se algum arquivo estiver ausente antes de implementar esse comportamento.

**Critério de aceite:** Queda do processo não perde o trabalho nem duplica fatos; limites são configuráveis; a atualização diária começa às 09:30 no fuso aprovado e o sistema trata explicitamente arquivos externos ausentes.

**Impacto:** Sustenta sincronização e reprocessamento local.

**Dependências:** BK-027, BK-029, BK-031

**Fonte:** U_CEI_Tela_15; Fechamento vigente 13/09/2026, §8.11; requisito operacional registrado em `G0-DECISION-REGISTER.md` em 2026-10-11

**Regras e verificações vinculadas:** F-19

### BK-040 — Validar acesso e cobertura dos dados do Bling

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Provar que os campos necessários estão disponíveis.

**O que fazer:** Fazer descoberta técnica em conta e permissões autorizadas.

**Como fazer:**

1. Mapear autenticação, recursos e campos reais da API para DRE por competência, balancete, saldos/contas, movimentos, custos, produção, BOM, compras, resgates e posição.
2. Registrar endpoint, campo, amostra protegida, competência, significado, data, limites e correspondência com as regras contábeis aprovadas; não alterar o plano de contas do Bling.
3. Para a DRE, comparar o cálculo do Majucau com o relatório do Bling na mesma competência e posição; diferenças seguem as regras aprovadas e ficam explicadas. Existência funcional informada não comprova endpoint específico disponível.
4. Para Conta Caixa, usar API oficial ou alternativa de exportação estruturada oficial de investimentos já permitida, sem scraping nem pasta adicional no Drive.
5. Campo específico não acessível = Dado indisponível no Bling. Produzir evidência técnica por componente, sem inferir, estimar, preencher ausência ou fabricar zero.

**Critério de aceite:** Matriz de campos e resultados de validação registrados; nenhuma prova foi executada nesta atualização de backlog. As regras gerais foram recebidas; critérios específicos e mapeamento da empresa precisam de validação contábil antes da implementação da DRE.

**Impacto:** Evita construir telas apoiadas em dados inexistentes.

**Dependências:** BK-010, BK-014, BK-024

**Fonte:** Bling documentação oficial; H_Contrato_Dados; Consolidação 13/09/2026, §§3–9; pendências 1–8; Fechamento vigente 13/09/2026, §§2,8

**Regras e verificações vinculadas:** R-03,R-04,HT-01,HT-02,HT-03,HT-04,HT-05,HT-06,HT-07,HT-08,F-02,F-03

### BK-041 — Implementar o adaptador de leitura do Bling

**Situação:** Em andamento. **Evolução:** 55%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Ingerir dados oficiais sem alterar o ERP.

**O que fazer:** Construir cliente e autenticação no backend.

**Como fazer:**

1. Armazenar credenciais fora do frontend
2. Renovar tokens com exclusão de concorrência
3. Aplicar paginação, limites vigentes, retry e erros de autenticação
4. Manter Conta Caixa, Itaú e Mercado Pago dentro da leitura oficial do Bling. Se a homologação apontar necessidade de exportação estruturada oficial de investimentos, implementar recepção local interna desse arquivo com o parser de BK-043; não usar o Drive.
5. Implementar extração dos lançamentos e campos necessários, com recursos reais homologados; calcular a DRE por competência no Majucau conforme as regras aprovadas e expor limitações como Dado indisponível, sem copiar a DRE pronta do Bling.

**Critério de aceite:** Adaptador usa somente operações aprovadas; 401/429/falha de rede geram diagnóstico e retomada segura. Não contém scraping nem integração direta presumida com banco ou Mercado Pago. A alternativa oficial necessária tem recepção, autenticação e proveniência implementadas; repetição e versão corrigida são tratadas por BK-044.

**Evidência da auditoria estática e correção parcial (11/10/2026):** já existem OAuth e armazenamento protegido no worker, cliente somente de leitura para contas a receber e a pagar, limites de resposta/página, retry seguro para GET, persistência RAW atômica e replay idempotente. A trava compartilhada entre os serviços de configuração e OAuth foi implementada; os testes locais dos pacotes Bling/worker e `go vet` passaram. A CI completa `38120125096` passou nos jobs Windows e PostgreSQL, incluindo os cenários `-race` no Ubuntu. O alerta anterior do scanner Windows era uma fixture de teste sem marcador e foi corrigido. A sincronização agora falha com segurança se não consegue saber se a página recebida é a última; a CI `38121237550` passou no teste PostgreSQL que confirma rollback e preservação do cursor. A cobertura real de campos/recursos e o cursor oficial não foram homologados. A trava é local ao processo worker. Ver `docs/evidence/BLING-ADAPTER-STATIC-AUDIT-2026-10-11.md`. Estimativa parcial não representa integração pronta.

**Impacto:** Mantém integração estável e credenciais protegidas.

**Dependências:** BK-040, BK-039, BK-038

**Fonte:** M-003,M-035; documentação Bling; Consolidação 13/09/2026, §§4,8–9; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** R-04

### BK-042 — Implementar leitura das três estruturas autorizadas do Google Drive

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Importar arquivos preservando integralmente a pasta compartilhada.

**O que fazer:** Construir adaptador de leitura do Drive e área de processamento dentro do Majucau.

**Como fazer:**

1. Configurar a identidade e os identificadores reais da pasta/estruturas; limitar leitura a 01_perdas_operacionais, 02_nuvem_pago/recebimentos_futuros e 03_nuvem_envio.
2. Detectar arquivos, obter conteúdo consistente e calcular hash local; se houver alteração durante a leitura, não publicar lote parcial.
3. Guardar controle e cópia de evidência necessária no armazenamento interno, conforme retenção homologada; nunca criar, mover, renomear, excluir, sobrescrever ou editar arquivos na origem.
4. Tratar permissão revogada, indisponibilidade de rede e versão corrigida com diagnóstico e retomada interna.

**Critério de aceite:** Leitura funciona com permissão somente de leitura; processar e reprocessar não altera nenhum arquivo do Drive. Arquivo parcial ou fora do contrato não produz fatos publicados.

**Impacto:** Protege os originais e mantém a trilha de importação dentro do Majucau.

**Dependências:** BK-010, BK-039, BK-038, BK-024

**Fonte:** M-009,M-010; L-11,L-12; Consolidação 13/09/2026, §§1–2

**Regras e verificações vinculadas:** R-01,R-02

### BK-043 — Implementar validação e interpretação dos arquivos

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Converter entradas heterogêneas em registros confiáveis.

**O que fazer:** Construir parsers por contrato e versão.

**Como fazer:**

1. Validar colunas, encoding, separadores, datas, moeda e sinais
2. Distinguir célula vazia, zero e conteúdo inválido
3. Registrar rejeições por linha e rejeição total quando o contrato exigir
4. Vincular parsers do Drive exclusivamente às três estruturas aprovadas. Se necessária após BK-040, implementar também interpretação da exportação estruturada oficial de investimentos do Bling recebida internamente por BK-041, fora do Drive; não habilitar fontes ilustrativas antigas.

**Critério de aceite:** Amostras válidas produzem registros esperados; arquivo incompatível explica o erro sem produzir números parciais ocultos. A alternativa oficial, quando necessária, tem parser e recepção atribuídos, com proveniência e deduplicação por BK-044; não cria uma quarta estrutura no Drive.

**Impacto:** Evita erros silenciosos de importação.

**Dependências:** BK-042

**Fonte:** H_Contrato_Dados; T-FILE-*; Consolidação 13/09/2026, §§1–2,4

### BK-044 — Deduplicar arquivos e fatos de negócio

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Garantir que repetir uma carga não altere os totais.

**O que fazer:** Implementar chaves e transações de importação.

**Como fazer:**

1. Deduplicar por hash e chave de evento
2. Preservar eventos diferentes com o mesmo número de transação
3. Tratar arquivo corrigido como nova versão rastreável
4. Distinguir identidade do arquivo, hash do conteúdo, período e lote; detectar duplicatas de eventos mesmo em arquivos diferentes. Correção cria nova versão interna com diferença rastreável, sem substituição silenciosa.

**Critério de aceite:** Carga repetida e duas cargas concorrentes não duplicam fatos; estorno legítimo é preservado. O efeito da correção sobre relatórios é demonstrável e nenhuma versão externa é alterada.

**Impacto:** Protege todos os indicadores contra duplicidade.

**Dependências:** BK-043, BK-041, BK-018

**Fonte:** M-038; T-IDEMP-*; Consolidação 13/09/2026, §2

**Regras e verificações vinculadas:** R-02,F-20

### BK-045 — Ingerir produtos, clientes e vendas realizadas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Criar base oficial de vendas por item e canal.

**O que fazer:** Mapear entidades Bling e estados da venda.

**Como fazer:**

1. Preservar IDs por pedido/item/SKU
2. Tratar cancelamentos, devoluções e alterações
3. Identificar B2C/B2B com regra aprovada e preservar unidade

**Critério de aceite:** Venda Bling é contada uma vez; evento Nuvem não cria outra venda.

**Impacto:** Sustenta resultados e projeções sem duplicidade.

**Dependências:** BK-044, BK-040

**Fonte:** M-003; H_Contrato_Dados!A4:H4

### BK-046 — Ingerir recebíveis e eventos do Nuvem Pago

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Separar direito a receber de venda e recebimento.

**O que fazer:** Ler a agenda Nuvem/Nuvem Pago na estrutura 02_nuvem_pago/recebimentos_futuros e relacioná-la aos fatos do Bling.

**Como fazer:**

1. Ler a agenda em 02_nuvem_pago/recebimentos_futuros preservando IDs, nomes, valores, datas e parcelas que existirem; ela representa previsão.
2. Obter fatos realizados no Bling e manter datas de ocorrência, competência, coleta e corte distintas; análise/conciliação normal usa D-1.
3. Relacionar previsão e liquidação correspondente para evitar duplicidade. Uma previsão futura não é excluída somente por ter vencimento depois de D-1.
4. Preservar taxas, juros, chargebacks e estornos comprovados no módulo pertinente, sem enviar lógica de cartão à calculadora.

**Critério de aceite:** Venda, direito futuro e dinheiro recebido não duplicam caixa; o corte D-1 não elimina a agenda futura necessária à projeção.

**Impacto:** Melhora a confiabilidade das contas e da conciliação.

**Dependências:** BK-044, BK-045

**Fonte:** M-004,M-005,M-011,M-012; H_Contrato_Dados!A5:H6; Consolidação 13/09/2026, §10; Fechamento vigente 13/09/2026, §8.2,8.5,8.11

**Regras e verificações vinculadas:** R-10,HT-08,F-18

### BK-047 — Ingerir obrigações e movimentos financeiros

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Formar títulos reais e suas evidências de pagamento.

**O que fazer:** Ler títulos e movimentos financeiros disponíveis no Bling, incluindo contas Itaú e Mercado Pago.

**Como fazer:**

1. Mapear documento, fornecedor, vencimento, valor e eventos
2. Preservar datas de competência e caixa separadas
3. Não transformar recomendação ou forecast em obrigação
4. Mapear identidade real de cada conta, sinal, transferência interna e evidência de liquidação; não prever arquivo bancário separado na pasta compartilhada.

**Critério de aceite:** Cada título possui chave de origem e histórico; previsão de compra não aparece como dívida constituída.

**Impacto:** Evita compromissos financeiros fictícios.

**Dependências:** BK-044, BK-005

**Fonte:** M-006,M-007; T04; Consolidação 13/09/2026, §§8–10

**Regras e verificações vinculadas:** R-08,R-09

### BK-048 — Ingerir estoque, produção, compras e notas de entrada

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Criar uma posição operacional única.

**O que fazer:** Mapear snapshots e estados operacionais do Bling.

**Como fazer:**

1. Ler saldos por SKU/unidade/depósito e mínimo
2. Mapear OP planejada, em andamento e concluída
3. Vincular pedido atendido, timestamp e NF-e de entrada
4. Homologar no cadastro operacional ID/descrição do depósito, saldo físico e, se existirem, reservado/disponível/outros estados. Ingerir também BOM por produto, componente, quantidade e unidade.

**Critério de aceite:** Fatos mantêm IDs, posição e unidade; campo ausente é indisponível e não zero. Sem identificação inequívoca, mostrar Dado indisponível no Bling; não hardcodar nome/ID nem escolher automaticamente Bloqueado.

**Impacto:** Faz Estoque, Produção e Compras concordarem.

**Dependências:** BK-044, BK-040

**Fonte:** M-035; H_Contrato_Dados!A11:H14,A19:H20; Consolidação 13/09/2026, §3; pendência 7

**Regras e verificações vinculadas:** R-03,HT-03,HT-07

### BK-049 — Ingerir etiquetas, faturas e frete da nota fiscal

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Preservar o custo logístico realizado e provisório.

**O que fazer:** Relacionar Nuvem Envio, envio, pedido e NF de saída.

**Como fazer:**

1. Mapear cotação e fatura por envio
2. Substituir cotação somente quando sua fatura chegar
3. Diferenciar NF válida sem frete de falha na leitura
4. Usar 03_nuvem_envio e preservar separadamente custo efetivo, frete cobrado do cliente, diferença/subsídio, pedido, transportadora, etiqueta e data quando disponíveis.

**Critério de aceite:** Fatura e cotação do mesmo envio não se somam; status e origem do custo permanecem visíveis. Frete cobrado do cliente não é compensado silenciosamente com custo; ausência de coluna fica explícita.

**Impacto:** Evita duplicar frete e distorcer margens.

**Dependências:** BK-044, BK-045, BK-010

**Fonte:** M-013..M-016; H_Contrato_Dados!A7:H9; Consolidação 13/09/2026, §12

**Regras e verificações vinculadas:** R-12,F-18

### BK-050 — Ingerir eventos de perdas operacionais

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Receber eventos de perda com origem e vínculo financeiro verificáveis.

**O que fazer:** Interpretar arquivos da estrutura 01_perdas_operacionais.

**Como fazer:**

1. Mapear evento, identificador, datas, valor, item e vínculo com Bling conforme amostra homologada.
2. Preservar ocorrência, confirmação, baixa e recuperação como fatos distintos; impedir duplicidade entre arquivo e Bling.
3. Manter a fonte no domínio Perdas Operacionais; não extrair percentual de scrap, rendimento de receita, yield ou fator automático de aumento de matéria-prima.

**Critério de aceite:** Cada evento tem proveniência e efeito financeiro conforme regra aprovada. Arquivo não cria venda, dinheiro ou movimento de estoque; não existe ingestão de extrato externo de investimento ou planilha genérica de ajustes nesta tarefa.

**Impacto:** Evita misturar perdas financeiras com parâmetros industriais.

**Dependências:** BK-043, BK-044, BK-007

**Fonte:** M-024..M-029,M-033; T15; Consolidação 13/09/2026, §11

**Regras e verificações vinculadas:** R-11

### BK-051 — Controlar atualização e falha por fonte

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar claramente quando os dados ficaram antigos.

**O que fazer:** Implementar política de freshness e snapshots válidos.

**Como fazer:**

1. Registrar separadamente tentativa/sucesso da coleta, data de referência/competência e corte D-1 da análise.
2. Definir frequência, timeout, repetição e validade por fonte como configuração técnica; D-1 não é periodicidade de sincronização.
3. Falha preserva o último snapshot válido, com aviso e data. Não generalizar o caso de 04/09/2026 em regra de sexta-feira ou feriado.

**Critério de aceite:** Coleta recente não mascara referência antiga; sem primeira carga, componente é indisponível. Agenda futura permanece projetada e não é cortada como movimento realizado.

**Impacto:** Evita decisões sobre dados antigos apresentados como atuais.

**Dependências:** BK-039, BK-038, BK-010

**Fonte:** L-24; U_CEI_Tela_15; Consolidação 13/09/2026, §10; Fechamento vigente 13/09/2026, §8.11

**Regras e verificações vinculadas:** R-10,F-19

### BK-052 — Implementar T15 Importações e Integrações

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir diagnosticar a entrada de dados.

**O que fazer:** Construir listagem de fontes, lotes, rejeições e tentativas.

**Como fazer:**

1. Calcular cartões com período explícito
2. Mostrar detalhes, hash, aceitações e rejeições sem segredos
3. Executar reprocessamento autorizado pelo mesmo motor de importação
4. Substituir fontes ilustrativas antigas pelas reais: Bling e as três estruturas aprovadas; exibir metadados, rejeições e versões internas sem controles de escrita no Drive.

**Critério de aceite:** Reprocessar mantém idempotência; fontes reais e estados conferem com lotes; export depende da decisão por tela. Não exibir extrato de investimentos, arquivo Itaú/Mercado Pago ou Planilha Ajustes/Cacau como fonte ativa. Quantidade de fontes declara o critério e não é número fixo do mockup.

**Impacto:** Reduz tempo para corrigir uma integração.

**Dependências:** BK-051, BK-044, BK-033, BK-028, BK-012

**Fonte:** T15; U_CEI_Tela_15; T-CEI15-01..08; Consolidação 13/09/2026, §§1–2

**Regras e verificações vinculadas:** R-01,F-18

## 05 Tesouraria e conciliação

### BK-053 — Implementar conciliação por evidência e unicidade

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Relacionar venda, transação e recebimento sem falso vínculo.

**O que fazer:** Construir regras de matching versionadas.

**Como fazer:**

1. Priorizar identificador inequívoco do Bling; na ausência, identificação/nome + valor, com normalização conservadora documentada e chaves reais homologadas.
2. Não conciliar automaticamente quando houver mais de um candidato; bloquear reutilização e preservar eventos distintos.
3. Aplicar D-1 aos movimentos realizados da análise normal; dia corrente exige solicitação expressa e auditada.
4. Manter PENDING_RECONCILIATION fora dos indicadores que exigem reconciliação, preservando dados e origem; pendência não significa valor zero.

**Critério de aceite:** Identificador prioritário e fallback seguem regra definida; ambiguidade vira conferência. Valor diferente em A Pagar segue C-01, sem tolerância inventada. Nenhuma regra especial fixa de sexta-feira é criada.

**Impacto:** Evita baixa ou recebimento associados à operação errada.

**Dependências:** BK-046, BK-047, BK-045, BK-036

**Fonte:** M-008; T03/T05; Consolidação 13/09/2026, §10; pendência 8; Fechamento vigente 13/09/2026, §8.6–8.8,8.11

**Regras e verificações vinculadas:** Q-19,R-10,HT-08,F-19,F-20

### BK-054 — Implementar revisão, confirmação e desfazimento de conciliação

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir corrigir exceções com trilha de decisão.

**O que fazer:** Criar fluxo contextual conforme permissões.

**Como fazer:**

1. Exibir evidências das duas pontas e regra aplicada
2. Exigir motivo nas decisões manuais aprovadas
3. Impedir conflito concorrente e registrar reversão sem apagar histórico
4. Implementar solicitação expressa para incluir movimentos do dia corrente, com capacidade, escopo, solicitante, data e auditoria; a exceção não altera silenciosamente a rotina padrão ou períodos fechados.

**Critério de aceite:** Revisão não altera origem Bling indevidamente; desfazer restaura vínculos consistentes.

**Impacto:** Resolve casos que não podem ser automatizados.

**Dependências:** BK-053, BK-034, BK-012

**Fonte:** L-23; T05/T13; Fechamento vigente 13/09/2026, §8.11

**Regras e verificações vinculadas:** F-19,F-20

### BK-055 — Implementar T05 Conciliação

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Dar visibilidade aos vínculos e exceções.

**O que fazer:** Construir cartões, tabela, distribuição e detalhamento.

**Como fazer:**

1. Adicionar busca/filtros aprovados e paginação
2. Exibir ID Bling, origem Nuvem, bruto, recebimento e situação
3. Conectar pendentes, ambíguos, divergentes, chargebacks e regras

**Critério de aceite:** Contagens totalizam a mesma população; cada linha permite consultar evidência e revisão autorizada.

**Impacto:** Facilita resolver divergências sem procurar em vários sistemas.

**Dependências:** BK-054, BK-028

**Fonte:** T05; T-RECON-*

### BK-056 — Implementar regras e leitura de Contas a Receber

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar direitos abertos, recebidos e vencidos corretamente.

**O que fazer:** Calcular saldos e agrupamentos com contratos aprovados.

**Como fazer:**

1. Separar valor bruto/líquido, vencimento, liquidação, canal e estado de reconciliação conforme origem.
2. Calcular previsto, aberto, realizado e atraso no contexto temporal correspondente. Movimentos realizados seguem D-1, salvo solicitação expressa do dia corrente.
3. Quando a liquidação correspondente integrar o universo realizado, retirar sua participação duplicada na previsão desse mesmo cálculo; preservar vínculo/histórico.
4. Indicadores confirmados exigem dados reconciliados; manter pendências visíveis separadamente, sem convertê-las em zero.

**Critério de aceite:** Recebido no mês usa data de liquidação; vencido usa posição e regra temporal; totais não duplicam fontes.

**Impacto:** Apoia cobrança e planejamento de caixa.

**Dependências:** BK-046, BK-053, BK-020

**Fonte:** T03; M-004,M-005; Consolidação 13/09/2026, §10; Fechamento vigente 13/09/2026, §8.5–8.11

**Regras e verificações vinculadas:** Q-20,F-19,F-20

### BK-057 — Implementar T03 Contas a Receber

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Expor valores e exceções de recebimento de modo rastreável.

**O que fazer:** Construir cartões, lista, distribuição e meios de pagamento.

**Como fazer:**

1. Conectar detalhes e alertas aos filtros corretos
2. Mostrar cliente, origem, vencimento, bruto, líquido e status
3. Traduzir rótulos técnicos conforme padrão aprovado

**Critério de aceite:** Lista, cartões e gráfico usam a mesma posição; usuário chega à origem de qualquer valor.

**Impacto:** Reduz dúvidas sobre o que ainda será recebido.

**Dependências:** BK-056, BK-055

**Fonte:** T03

### BK-058 — Implementar regras de Contas a Pagar

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar obrigações e efeitos de pagamento de forma única.

**O que fazer:** Aplicar a decisão C-01 e o calendário de vencimentos.

**Como fazer:**

1. Calcular aberto, vence hoje, atrasado e próximos sete dias
2. Tratar valor diferente, estorno e cancelamento conforme regra
3. Separar divergência de valor de status temporal

**Critério de aceite:** Exemplo R$1.000/R$600 reproduz decisão aprovada; título futuro não aparece vencido na data anterior.

**Impacto:** Evita saldo a pagar incorreto.

**Dependências:** BK-047, BK-005, BK-053

**Fonte:** T04; C-01

### BK-059 — Implementar T04 Contas a Pagar

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir acompanhar fornecedores, vencimentos e divergências.

**O que fazer:** Construir cartões, obrigações, categorias e detalhamento.

**Como fazer:**

1. Mostrar documento e vínculo financeiro
2. Conectar vencidos, impostos, fretes e conciliação
3. Reconciliar distribuição e total de divergências

**Critério de aceite:** Valor de divergências é o mesmo nos componentes com igual filtro; ações obedecem permissões.

**Impacto:** Ajuda a priorizar pagamentos e análise.

**Dependências:** BK-058, BK-028

**Fonte:** T04

### BK-060 — Implementar fluxo de caixa realizado, previsto e projetado

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Calcular a trajetória do caixa sem dupla contagem.

**O que fazer:** Combinar posição bancária e eventos temporais aprovados.

**Como fazer:**

1. Usar saldos e movimentos do Bling com o mesmo corte/posição; análise normal dos movimentos usa D-1. Não combinar saldo de hoje com movimento de ontem como se fossem uma posição conciliada.
2. Incluir agenda futura Nuvem como previsão, preservando vencimentos; realizado correspondente substitui sua participação futura no universo aplicável, sem dupla contagem.
3. Separar consolidado comprovado de saldo transferível Itaú. Transferências internas não produzem nova receita ou aumento líquido consolidado.
4. Calcular série temporal, horizonte, reserva e disponibilidade com o contrato consolidado; propagar ausência e pendência de reconciliação somente aos indicadores dependentes.

**Critério de aceite:** Trajetória de caixa é rastreável e consistente por corte; a calculadora consome disponibilidade consolidada sem conhecer regras de cartão. Nenhum saldo desconhecido é zero.

**Impacto:** Sustenta decisões de liquidez.

**Dependências:** BK-056, BK-058, BK-020, BK-006, BK-047

**Fonte:** T02; M-002; C-02; L-07; Consolidação 13/09/2026, §§8–10; Fechamento vigente 13/09/2026, §§1,8.5–8.11

**Regras e verificações vinculadas:** Q-21,R-08,R-09,F-18,F-19

### BK-061 — Implementar seleção e memória da semana crítica

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Explicar qual semana foi selecionada e por quê.

**O que fazer:** Aplicar algoritmo, calendário e posição de caixa aprovados.

**Como fazer:**

1. Calcular entradas e despesas por semana
2. Resolver empates e semanas incompletas
3. Preservar indicadores de mínimo de caixa como medida distinta quando aprovados
4. Selecionar a semana de maior total de despesas e vincular Caixa Semana Crítica ao contrato vigente do fluxo. Localizar a posição oficial nos artefatos/campos; não eleger inicial/final/mínimo por conveniência técnica.

**Critério de aceite:** Dataset com maior despesa e menor saldo em semanas diferentes produz decisão esperada.

**Impacto:** Impede divergência entre fluxo e calculadora.

**Dependências:** BK-060, BK-006

**Fonte:** T02/T08; C-02,L-07; Fechamento vigente 13/09/2026, §1

**Regras e verificações vinculadas:** F-01

### BK-062 — Implementar T02 Fluxo de Caixa

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Apresentar a posição e os próximos compromissos.

**O que fazer:** Construir cartões, gráfico combinado, semana crítica e alertas.

**Como fazer:**

1. Distinguir realizado, previsto e projetado com legendas
2. Conectar detalhes, reserva, saldo transferível e períodos
3. Mostrar posição da fonte e indisponibilidade de cálculos parciais
4. Identificar a origem e data do caixa consolidado e distinguir o valor transferível do Itaú para aplicação; recebíveis futuros permanecem previstos.

**Critério de aceite:** Valores, folgas, datas e gráfico conferem com motor; export mantém filtros aprovados.

**Impacto:** Torna a situação do caixa compreensível.

**Dependências:** BK-061, BK-028, BK-012

**Fonte:** T02; Consolidação 13/09/2026, §§8–10

## 06 Resultados, perdas e aplicações

### BK-063 — Preservar o custo Bling válido para cada venda

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Sustentar CMV e margem histórica auditáveis.

**O que fazer:** Implementar custo por SKU com validade e evidência.

**Como fazer:**

1. Mapear snapshot histórico confirmado
2. Vincular custo à data da venda
3. Aplicar decisão de indisponibilidade quando só houver custo atual
4. Comparar respostas reais de venda antiga e custo atual para provar semântica temporal do campo; guardar endpoint, campo e evidência do teste em HT-05.

**Critério de aceite:** Venda antiga não recebe custo atual como se fosse histórico confirmado; BOM e NF não substituem a metodologia financeira. Se só houver custo atual, histórico permanece indisponível, sem retroagir valores novos.

**Impacto:** Evita margem histórica artificial.

**Dependências:** BK-045, BK-007, BK-038

**Fonte:** M-019,M-020; L-03; Consolidação 13/09/2026, pendência 5

**Regras e verificações vinculadas:** HT-05

### BK-064 — Preservar contas, classificações e competência do Bling

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Traduzir recursos da API sem criar uma contabilidade paralela.

**O que fazer:** Implementar correspondência técnica entre campos/contas da origem e contratos de leitura.

**Como fazer:**

1. Mapear IDs, contas, categorias, natureza, hierarquia, competência e linhas existentes no Bling.
2. Preservar classificação e competência da origem; correspondência técnica não reclassifica receita, custo, perda, juros, frete ou tributo.
3. Campo sem correspondência ou não retornado gera diagnóstico e Dado indisponível no Bling, nunca conta coringa.
4. Versionar a correspondência técnica e preservar período fechado; mudança de integração não altera histórico silenciosamente.

**Critério de aceite:** Cada valor apresentado é rastreável ao registro, à conta e à classificação originais do Bling e ao mapeamento contábil aprovado no Majucau; não há mapa paralelo sem aprovação nem troca de competência por inferência.

**Impacto:** Permite conferir a composição do resultado.

**Dependências:** BK-041, BK-007, BK-038

**Fonte:** M-017,M-018,M-029; H_Contrato_Dados!A18:H18; Consolidação 13/09/2026, pendência 6; §12; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** R-05,R-12,HT-06,F-02

### BK-065 — Implementar perdas e recuperações sem dupla contagem

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Reconhecer eventos uma única vez no período correto.

**O que fazer:** Relacionar planilha, origem oficial e decisão financeira.

**Como fazer:**

1. Registrar eventos de perdas operacionais e recuperações de sua fonte específica, separando ocorrência, confirmação e liquidação.
2. Vincular fatos correspondentes do Bling e preservar sua classificação original como dado de origem; o tratamento na DRE do Majucau segue o mapeamento contábil aprovado e a competência comprovada.
3. Não acrescentar receita/despesa à DRE apenas por existir um evento no arquivo, nem dar baixa de caixa sem movimento financeiro comprovado.
4. Manter perdas operacionais separadas de yield, scrap e necessidade industrial de insumos.

**Critério de aceite:** Mesmo evento não reduz resultado duas vezes; ressarcimento sem recebimento comprovado não vira caixa. Diagnóstico externo não altera a DRE sem documento, competência e mapeamento aprovado.

**Impacto:** Protege resultado e rastreabilidade das perdas.

**Dependências:** BK-050, BK-064

**Fonte:** M-024..M-026; L-04,L-14; Consolidação 13/09/2026, §11; Fechamento vigente 13/09/2026, §§2,8.14

**Regras e verificações vinculadas:** R-11,F-18

### BK-066 — Calcular e conferir a DRE própria por competência

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Apresentar no Majucau o resultado calculado conforme regras contábeis aprovadas.

**O que fazer:** Consumir os lançamentos de origem necessários e calcular as linhas da DRE pelo motor Majucau.

**Como fazer:**

1. Usar regras e mapeamento contábil aprovados para calcular as linhas existentes por competência.
2. Obter campos e lançamentos pelos recursos reais homologados; manter origem, conta, sinal, data e competência rastreáveis.
3. Comparar linhas e totais com o relatório do Bling como referência, guardando período e evidência das diferenças.
4. Sem regra ou dado necessário, mostrar Dado indisponível. Preservar snapshots fechados e bloquear alteração retroativa silenciosa.

**Critério de aceite:** DRE no Majucau aplica as regras aprovadas, tem origem e cálculo rastreáveis e diferenças explicadas contra o Bling. Não estima nem completa linhas com fatos fictícios. Custo SKU e arquivos externos só entram conforme escopo e regra aprovados.

**Impacto:** Oferece resultado verificável.

**Dependências:** BK-041, BK-064, BK-038, BK-020

**Fonte:** T06; M-017..M-029; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** HT-06,F-02

### BK-067 — Implementar T06 DRE

**Situação:** Em andamento. **Evolução:** 20% (estimativa de planejamento; os cards informativos EBIT, EBITDA, CAPEX pago no mês e CAPEX acumulado estão desenhados no espaço aprovado, somente para consulta, e o card numérico UNMAPPED foi retirado. A tabela e o lucro da DRE não incluem CAPEX. Ainda faltam cálculo da DRE própria com lineage, integração dos dados e valores confiáveis; hoje EBIT/EBITDA e CAPEX ficam indisponíveis). **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Explicar como o resultado foi formado.

**O que fazer:** Construir linhas, cartões, composição e navegação.

**Como fazer:**

1. Preservar o layout T06 e as linhas aprovadas; apresentar os valores calculados pelo Majucau com competência e contas rastreáveis.
2. Usar a mesma camada de leitura/cálculo para cards, tabela, gráfico e detalhe em tela, com base de percentual explícita. Os cards são somente para consulta.
3. Apresentar ausência de campo como Dado indisponível no espaço previsto; não copiar números ilustrativos nem inferir classificação fora das regras contábeis aprovadas.
4. No topo da tela, no espaço aprovado para os indicadores onde hoje aparece `UNMAPPED`, exibir cards separados, somente para consulta. Dar destaque a `EBITDA` e `EBIT`, calculados a partir da DRE própria sem mudar suas linhas: EBITDA antes da depreciação; EBIT depois. Incluir também `CAPEX PAGO NO MÊS` e `CAPEX ACUMULADO` como informações separadas da tabela e do lucro. CAPEX pago no mês considera o pagamento efetivo no mês filtrado; CAPEX acumulado soma a categoria CAPEX dos relatórios desde o início do histórico confiável. Reduzir os cards apenas o necessário para caber no espaço aprovado; não alterar o restante da tela. Até a integração, mostrar Dado indisponível.

**Critério de aceite:** Tabela e gráfico usam o cálculo aprovado e não se contradizem; EBIT/EBITDA aparecem em cards separados e destacados no topo, derivados da DRE sem alterar as linhas dela; os cards substituem `UNMAPPED` no espaço aprovado; CAPEX pago no mês e acumulado são consultas separadas, vêm da categoria CAPEX dos relatórios e não entram no lucro; nenhuma outra parte do layout aprovado é alterada.

**Impacto:** Evita interpretação enganosa da rentabilidade.

**Dependências:** BK-066, BK-028, BK-021

**Fonte:** T06; Fechamento vigente 13/09/2026, §§2,5

**Regras e verificações vinculadas:** Q-22,F-02

**Reauditoria estática de 2026-10-11:** `frontend/src/mockupDashboard.tsx` contém os quatro cards e filtra o card `UNMAPPED` dessa área; o alerta de lançamentos pendentes continua separado. `frontend/src/bootstrap.ts` e `internal/application/contracts.go` aceitam as métricas de CAPEX, mas a busca de código de produção não encontrou fonte que as preencha. EBIT/EBITDA continuam com valor indisponível. A evidência de CI anteriormente registrada em `docs/evidence/MOCAP-T1-T15-REALIGNMENT-2026-09-20.md` cobre a presença da interface; esta reauditoria não executou testes e não altera telas ou regras.

### BK-068 — Reproduzir o balancete a partir da base do Bling

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Preservar partidas e saldos auditáveis.

**O que fazer:** Extrair e compor os dados contábeis/financeiros existentes no Bling.

**Como fazer:**

1. Mapear contas, abertura, débitos, créditos, contrapartidas e natureza dos saldos conforme recursos reais disponíveis.
2. Preservar competência/classificação e vínculos da origem; não solicitar nova estrutura contábil à Majucau.
3. Conferir totais e saldos contra o próprio Bling para o mesmo período e posição.
4. Elemento não obtido pela API = Dado indisponível no Bling; não estimar partida, criar conta de fechamento ou transformar ausência em zero.

**Critério de aceite:** Balancete reproduz a informação Bling; ausência pontual fica explícita, sem demonstração fictícia ou classificação paralela.

**Impacto:** Impede fechamento artificial do balancete.

**Dependências:** BK-041, BK-064, BK-038, BK-020

**Fonte:** T07; M-027,M-028; L-08; Consolidação 13/09/2026, pendência 4; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** HT-04,F-03

### BK-069 — Implementar T07 Balancete

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir conferir saldos e divergências por conta.

**O que fazer:** Construir tabela, grupos, cartões e detalhamento.

**Como fazer:**

1. Exibir saldo inicial, movimentos e saldo final conforme contrato
2. Explicar natureza devedora/credora e pendências
3. Ligar contrapartida ao documento original
4. Mostrar natureza, saldos e contas provenientes do Bling, preservando competência e campo indisponível; cartão e tabela não tratam diferença entre movimentos como prova de saldo final conhecido.

**Critério de aceite:** Saldo não nulo não é divergência por si só; composição usa grupos e base aprovados.

**Impacto:** Facilita investigação financeira.

**Dependências:** BK-068, BK-028

**Fonte:** T07; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** Q-23,F-03

### BK-070 — Definir e implementar PMR, PMP, PME e ciclo financeiro

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Financeiro + Backend.

**Objetivo:** Explicar os indicadores de prazo presentes na calculadora.

**O que fazer:** Homologar e calcular cada indicador separadamente.

**Como fazer:**

1. Definir período, fontes, numeradores e denominadores de prazo médio de recebimento, pagamento e estoque
2. Implementar médias e ciclo conforme regra aprovada, preservando ausência e denominador zero
3. Exibir memória e função diagnóstica, sem alterar a capacidade de aplicação por inferência

**Critério de aceite:** Cada indicador tem fonte e fórmula própria; os 28 e 41 dias da imagem não viram constantes; NCG só entra se confirmado no escopo.

**Impacto:** Evita confundir três medidas diferentes ou influenciar indevidamente a calculadora.

**Dependências:** BK-056, BK-058, BK-077, BK-066

**Fonte:** T08 indicadores PMR/PMP/PME e ciclo; B_Intersecoes!A12:I12

**Regras e verificações vinculadas:** Q-27

### BK-071 — Implementar movimentos de investimentos e posição comprovada no Bling

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Separar principal, rendimentos e posição com evidência de origem.

**O que fazer:** Usar Conta Caixa via API oficial ou exportação estruturada oficial homologada.

**Como fazer:**

1. Identificar Transferência destinada à aplicação como aporte de principal e Rendimentos como rendimento realizado; não classificar toda transferência como investimento.
2. Homologar categoria, sinal, origem/destino, aplicação e resgates parciais/totais. Já Investido = aportes de principal − resgates de principal; período/cobertura precisam ser comprovados.
3. Obter posição, data e instituição somente se a fonte demonstrar esses campos; principal líquido não é automaticamente posição atual.
4. Exibir componentes independentes comprováveis e declarar os ausentes. Não presumir reinvestimento de rendimento nem resgate zero por falha de identificação.
5. Principal aplicado/resgatado não vira automaticamente receita/despesa. Preservar como dado a classificação e a competência informadas pelo Bling; a classificação na DRE Majucau segue as regras contábeis aprovadas, sem inventar uma segunda regra.

**Critério de aceite:** Aportes e resgates reconciliam com fatos reais. Posição ausente não é inventada. Indicador dependente de resgate desconhecido permanece incompleto; indicador independente comprovado pode ser exibido.

**Impacto:** Dá base correta à calculadora e à consulta de aplicações.

**Dependências:** BK-041, BK-044, BK-006, BK-064, BK-047

**Fonte:** M-030..M-034; T08; Consolidação 13/09/2026, §§4–7; Fechamento vigente 13/09/2026, §2

**Regras e verificações vinculadas:** Q-10,R-04,R-05,R-06,R-07,HT-01,HT-02

### BK-072 — Implementar o cálculo de capacidade de aplicação

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Reproduzir a decisão financeira homologada.

**O que fazer:** Calcular lucro elegível, reserva, máximo e executável.

**Como fazer:**

1. Consumir contratos financeiros consolidados de resultado elegível, disponibilidade, caixa crítico, reserva e principal líquido; o resultado elegível vem da DRE calculada no Majucau conforme regras aprovadas.
2. Aplicar a janela e fórmulas transcritas em BK-006, deduzindo principal líquido uma única vez e limitando o executável ao saldo transferível Itaú.
3. Propagar corte, versão e qualidade das entradas; detalhes de bandeira, MDR, taxas e parcelas não integram este domínio.
4. Expor componentes comprováveis e indicar ausência dos demais sem zero fictício; nenhuma transferência ou aplicação é executada.

**Critério de aceite:** Memória de cálculo reproduz o contrato consolidado com entradas verificáveis; falta apenas da posição total não bloqueia componentes independentes. Não reabre critérios de negócio.

**Impacto:** Evita valor disponível aparente sem sustentação.

**Dependências:** BK-071, BK-061, BK-066, BK-006

**Fonte:** T08; M-034; C-02,C-03,L-07; Consolidação 13/09/2026, §§6–9; Fechamento vigente 13/09/2026, §§1–2

**Regras e verificações vinculadas:** R-06,R-08,R-09,HT-01,F-01

### BK-073 — Implementar T08 Calculadora de Aplicação

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar cálculo, condições e motivos de bloqueio.

**O que fazer:** Construir indicadores e memória detalhada.

**Como fazer:**

1. Exibir fonte, período, fórmula e valores intermediários
2. Distinguir máximo calculado de executável
3. Conectar parâmetros, DRE, fluxo e aplicações sem ação financeira
4. Identificar Itaú na memória do executável, apresentar aportes/resgates/rendimentos separadamente e explicar exatamente qual entrada falta.

**Critério de aceite:** Estado OPEN não apresenta capacidade confirmada; tela não exporta nem executa dinheiro conforme contrato a validar. Não copiar posição ou capacidade fictícia do mockup; componentes verificados conservam sua visibilidade.

**Impacto:** Torna o resultado compreensível e auditável.

**Dependências:** BK-072, BK-028, BK-012, BK-070

**Fonte:** T08; T16 regra Tela 08 não exporta; Consolidação 13/09/2026, §§6–9

**Regras e verificações vinculadas:** R-07

### BK-074 — Implementar a consulta de Aplicações

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Cobrir o módulo citado no menu e na calculadora.

**O que fazer:** Construir a tela e o detalhamento após completar a especificação.

**Como fazer:**

1. Homologar visual, campos e filtros usando os dados que Conta Caixa realmente disponibilizar, sem exigir extrato externo no Drive.
2. Mostrar aportes, resgates, rendimentos realizados e suas datas; mostrar posição atual somente quando comprovada, com instituição/aplicação.
3. Reconciliar a consulta com Bling e calculadora, mantendo explicação de ausência por componente.

**Critério de aceite:** Layout validado e totais reconciliados com a fonte oficial homologada; nenhum rendimento é presumido reinvestido e não existe scraping ou pasta de investimento.

**Impacto:** Evita entregar um menu sem funcionalidade.

**Dependências:** BK-004, BK-071, BK-021

**Fonte:** Menus; T08 Ver aplicações; Consolidação 13/09/2026, §§4–7

**Regras e verificações vinculadas:** Q-02,R-04,R-07,HT-02

### BK-075 — Implementar análise de preço e margem no local aprovado

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Cobrir Pricing citado na matriz sem alterar o ERP.

**O que fazer:** Construir cálculo analítico por SKU e canal.

**Como fazer:**

1. Manter valor bruto, custo variável, tarifa/custo fixo, critério de rateio e custo final apropriado por produto em componentes rastreáveis.
2. Usar a apropriação fornecida pelo Bling quando existente; ausência de informação suficiente não autoriza distribuir por um critério inventado.
3. Garantir que a tarifa fixa de uma operação não seja multiplicada integralmente por cada item; reconciliar soma apropriada com custo original.
4. Preservar análise de preço/margem e layout aprovados, separada do motor de quantidade de compras; não gravar preço no Bling.

**Critério de aceite:** Apropriação não duplica custo fixo; parcela indisponível é declarada. Necessidade operacional de compra funciona independentemente de ausência de rateio/custo econômico.

**Impacto:** Permite analisar preço com regra rastreável.

**Dependências:** BK-004, BK-063, BK-064, BK-036

**Fonte:** M-021..M-023; L-06,L-10; Fechamento vigente 13/09/2026, §3.3

**Regras e verificações vinculadas:** L-06,L-10,F-06

### BK-076 — Implementar indicadores de logística no local aprovado

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Cobrir os custos e subsídios de frete citados na matriz.

**O que fazer:** Calcular resultado do frete por pedido e transportadora.

**Como fazer:**

1. Aplicar custo confirmado/provisório e frete da NF
2. Calcular subsídio e resultado preservando sinal
3. Apresentar agregados comparáveis sem afirmar melhor transportadora só pela média
4. Exibir custo efetivo do frete, valor cobrado do cliente e diferença/subsídio em componentes separados, com rastreio para pedido/etiqueta quando disponíveis.

**Critério de aceite:** Fatura substitui cotação por envio; origem e qualidade acompanham o indicador.

**Impacto:** Evita análise logística baseada em custo duplicado.

**Dependências:** BK-004, BK-049, BK-064

**Fonte:** M-013..M-016; L-16; Consolidação 13/09/2026, §12

**Regras e verificações vinculadas:** L-16,R-12

## 07 Estoque, produção e compras

### BK-077 — Implementar classificação e posição de estoque

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar quantidade por item e unidade correta.

**O que fazer:** Aplicar saldo, tipo, mínimo e pedido aberto aplicável.

**Como fazer:**

1. Separar produto pronto e insumo
2. Aplicar igualdade ao mínimo conforme regra aprovada
3. Calcular abaixo do mínimo sem pedido e com reposição
4. Consumir somente depósito operacional inequivocamente identificado por HT-03; preservar diferença entre saldo físico, reservado e disponível retornados.
5. Distinguir saldo atual, comprometimento, mínimo cadastrado, necessidade futura e cobertura por pedido. Existência de pedido não comprova cobertura completa; universos distintos de T09/T11 não têm contagens forçadas a coincidir.

**Critério de aceite:** Produtos prontos não recebem mínimo/alerta de insumo; kg e unidades não são somados. Identificação ambígua ou dado obrigatório ausente produz Dado indisponível no Bling e impede apenas resultados dependentes. A ausência de reservado/disponível opcionais não invalida saldo físico comprovado se a regra homologada usar esse saldo; estados opcionais não são inferidos.

**Impacto:** Evita indicação incorreta de necessidade de compra.

**Dependências:** BK-048, BK-008, BK-020

**Fonte:** T09; Q_CEI_Tela_11; Consolidação 13/09/2026, §3; Fechamento vigente 13/09/2026, §3

**Regras e verificações vinculadas:** R-03,HT-03

### BK-078 — Implementar T09 Estoque

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Exibir a posição operacional informativa.

**O que fazer:** Construir filtros, lista, cartões e distribuição.

**Como fazer:**

1. Mostrar pedido, quantidade, mínimo e situação
2. Calcular distribuição por população de SKUs definida
3. Conectar lista completa e origem Bling

**Critério de aceite:** Contagens reconciliam com Compras; não há escrita de estoque no Bling.

**Impacto:** Permite identificar insumos que exigem atenção.

**Dependências:** BK-077, BK-028

**Fonte:** T09

### BK-079 — Implementar metas de produção por SKU e período

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Guardar o planejamento interno da Majucau.

**O que fazer:** Criar cadastro específico de metas versionadas.

**Como fazer:**

1. Armazenar meta explícita por produto/período como planejamento, separada de fato Bling e de forecast.
2. Registrar usuário, data/hora, período, valor anterior/novo e versão, com capacidade correspondente; não apagar histórico.
3. T01 e T10 usam o mesmo registro e métrica para resumo/detalhe. T12 preserva cenário/forecast; dependência técnica não sincroniza conceitos diferentes automaticamente.

**Critério de aceite:** Alteração autorizada mantém histórico; meta ausente é diferente de zero.

**Impacto:** Permite comparar produção real e objetivo sem alterar OP.

**Dependências:** BK-008, BK-036, BK-034

**Fonte:** T10; T16 Meta por SKU; Fechamento vigente 13/09/2026, §§4,6.4

**Regras e verificações vinculadas:** Q-08,Q-12,F-11,F-16

### BK-080 — Implementar indicadores de produção realizada

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Separar ordens, quantidades e alcance de metas.

**O que fazer:** Calcular planejado, andamento, concluído e percentual por SKU.

**Como fazer:**

1. Usar produção efetivamente informada pelo Bling por produto, unidade, período e estado; ausência não vira zero.
2. Calcular uma vez realizado, meta, percentual e status para card, tabela, gráfico e Produzido Geral na executiva.
3. Meta atingida: realizado >= meta correspondente. Em andamento/abaixo: comparar com avanço esperado comprovado do plano no ponto temporal.
4. Sem granularidade temporal suficiente, informar avaliação de atraso indisponível, mantendo componentes conhecidos; não assumir progresso linear, tolerância ou atraso por <100%.

**Critério de aceite:** Mesma métrica e mesmo universo produzem os mesmos resultados em T01/T10; quantidades incompatíveis não são somadas e produção ausente permanece desconhecida.

**Impacto:** Evita indicadores contraditórios como os do mockup.

**Dependências:** BK-048, BK-079, BK-008

**Fonte:** T10; T10 duplicata; Fechamento vigente 13/09/2026, §4.1–4.4

**Regras e verificações vinculadas:** F-08,F-09

### BK-081 — Implementar T10 Produção

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Apresentar acompanhamento e acesso às metas.

**O que fazer:** Construir período, busca, filtros, tabela e cartões.

**Como fazer:**

1. Exibir unidade e posição de sincronização
2. Conectar Definir metas ao fluxo autorizado
3. Preservar natureza informativa e ausência de exportação/alertas conforme regra aprovada
4. Compartilhar a métrica canônica com Produzido Geral da T01; apresentar quantidade/meta/realizado/progresso/status no layout aprovado, sem copiar números do mockup.

**Critério de aceite:** Tabela e distribuição usam a mesma população; nenhuma OP é criada, cancelada ou editada no Bling.

**Impacto:** Ajuda a acompanhar produção com confiança.

**Dependências:** BK-080, BK-028

**Fonte:** T10; M-052; Fechamento vigente 13/09/2026, §4.2–4.4

**Regras e verificações vinculadas:** F-08,F-09

### BK-082 — Avaliar necessidade e cobertura temporal das compras

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Usar uma regra única entre Estoque e Compras.

**O que fazer:** Identificar necessidade por insumo/data e cobertura efetiva de estoque e pedidos.

**Como fazer:**

1. Consumir demanda bruta de insumos da produção/BOM, estoque utilizável, comprometimento aplicável e pedidos pendentes.
2. Considerar somente quantidade ainda pendente do pedido cuja chegada é válida antes da necessidade; recebida/cancelada não é novamente contada como pendente.
3. Manter pedido parcial e tardio visíveis, sem excluir o insumo da avaliação. Classificar cobertura integral, parcial ou inexistente e horizonte imediato/futuro.

**Critério de aceite:** Pedido aberto não elimina item automaticamente. T09 e T11 coincidem somente em métricas realmente equivalentes, com mesmo universo e posição; recomendação não cria pedido.

**Impacto:** Evita compras duplicadas e números distintos entre telas.

**Dependências:** BK-077, BK-008, BK-094

**Fonte:** M-042; Q_CEI_Tela_11; Fechamento vigente 13/09/2026, §3.1–3.2

**Regras e verificações vinculadas:** F-04

### BK-083 — Implementar preço faturado e prazo histórico de compras

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar fatos de compra comprováveis por insumo.

**O que fazer:** Calcular último preço e tempo até ATENDIDO.

**Como fazer:**

1. Vincular última NF-e válida ao pedido atendido
2. Preservar preço, moeda, unidade e quantidade
3. Usar timestamp real de atendimento com janela/estatística homologadas

**Critério de aceite:** Preço não vem de cotação; prazo não usa média genérica de fornecedor sem a regra individual.

**Impacto:** Melhora a base de priorização e previsão de compra.

**Dependências:** BK-048, BK-008

**Fonte:** M-043; H_Contrato_Dados!A19:H20

**Regras e verificações vinculadas:** F-05

### BK-084 — Calcular necessidade líquida e prioridade operacional de compra

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Aplicar o planejamento de compra homologado.

**O que fazer:** Calcular quanto comprar e por que aquele insumo tem prioridade.

**Como fazer:**

1. Necessidade líquida = necessidade bruta − estoque utilizável − cobertura pendente que chega em tempo; necessidade não positiva gera nenhuma compra adicional e excedente permanece visível.
2. Reproduzir 100 kg − 20 kg − 30 kg = 50 kg; pedido tardio não reduz cobertura na data de necessidade e não pode esconder ruptura.
3. Priorizar risco imediato de ruptura e impacto na produção usando cobertura, lead time, demanda e datas disponíveis; documentar ordenação e motivo, sem pontuação/tolerância arbitrária.
4. Manter custo, tarifa, rateio e inteligência cacau separados da quantidade; ausência desses custos não bloqueia recomendação operacional comprovável.

**Critério de aceite:** Quantidade e prioridade são explicáveis por insumo/data; pedidos parciais/tardios não mascaram necessidade, cobertura não é contada duas vezes e nenhuma ordem é criada no Bling.

**Impacto:** Evita excesso ou fragmentação de compras.

**Dependências:** BK-082, BK-083, BK-008

**Fonte:** L-15,L-17,L-18; Q_CEI_Tela_11; Fechamento vigente 13/09/2026, §3.1–3.3

**Regras e verificações vinculadas:** F-04,F-05,F-06

### BK-085 — Implementar sinais externos do cacau

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Apresentar risco e oportunidade com fontes verificáveis.

**O que fazer:** Construir adaptadores e avaliação de sinais homologados.

**Como fazer:**

1. Relacionar sinais somente a localidades reais dos fornecedores da Majucau e registrar origem, unidade e data.
2. Separar necessidade operacional, preço histórico/atual e contexto de preço/clima; não criar região ou comparação fictícia.
3. Refinar e validar fontes externas independentemente do motor principal; manter indisponibilidade explícita quando um sinal não estiver habilitado.

**Critério de aceite:** Sinais habilitados são rastreáveis e comparáveis; sinal ausente não impede cálculo principal de compra e nenhum sinal executa pedido.

**Impacto:** Apoia análise de suprimento sem automatizar compra indevida.

**Dependências:** BK-011, BK-039, BK-036, BK-083

**Fonte:** M-044; Q_CEI_Tela_11; Consolidação 13/09/2026, pendência 10; Fechamento vigente 13/09/2026, §3.4

**Regras e verificações vinculadas:** R-14,HT-10,F-07

**Condição:** Fontes externas específicas habilitadas após refinamento técnico; não bloqueia compras operacionais.

### BK-086 — Implementar T11 Compras para consulta

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Exibir as informações de compras para consulta, sem executar operações.

**O que fazer:** Construir a tela conforme a referência visual aprovada, com recomendação operacional e contexto separado, respeitando a decisão posterior de consulta somente.

**Como fazer:**

1. Apresentar necessidade, quantidade, cobertura, prazo e prioridade pelo contrato funcional vigente nos blocos aprovados.
2. Preservar bloco de inteligência do cacau com componentes disponíveis; ausência/refinamento de fonte externa não bloqueia motor principal.
3. Somente mostrar os dados; não disponibilizar exportação/download, criação, alteração ou envio de pedidos ao Bling.
4. Manter “Gerar pedidos” no mesmo lugar e desativado, conforme confirmação visual de 11/10/2026. Não associar handler nem ação de exportação, download, criação, alteração ou envio de pedido. A referência histórica permanece arquivada.

**Critério de aceite:** A tela exibe dados consultivos consistentes com a mesma base e versão; não existe ação funcional para exportar/baixar ou gerar/alterar pedidos. O botão “Gerar pedidos” permanece no lugar aprovado e desativado. Qualquer outra mudança visual depende de aprovação explícita.

**Impacto:** Permite discutir compras com evidência sem alterar o ERP.

**Dependências:** BK-004, BK-084, BK-028

**Fonte:** Q_CEI_Tela_11; T-CEI11-01..12; Fechamento vigente 13/09/2026, §§3.4,5

**Regras e verificações vinculadas:** F-05,F-07

## 08 Projeções e cenários

### BK-087 — Preparar histórico de vendas por SKU para projeções

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Construir séries temporais confiáveis.

**O que fazer:** Preparar histórico confiável de 2025 em diante, declarando meses disponíveis por SKU.

**Como fazer:**

1. Buscar histórico em janelas compatíveis com API
2. Tratar cancelamentos, devoluções, períodos sem venda e lacunas
3. Definir SKU novo, descontinuado e mudanças de unidade
4. Não buscar dados anteriores a 2025 para completar artificialmente a referência de 24 meses; registrar insuficiência e corte temporal antes do backtest.

**Critério de aceite:** Cobertura real começa em 2025 ou depois conforme a série; mês ausente não é venda zero. Treino e avaliação declaram quantidade de observações, limitações e origem.

**Impacto:** Evita treinar projeções sobre histórico incompleto.

**Dependências:** BK-045, BK-009, BK-023

**Fonte:** R_CEI_Tela_12; Consolidação 13/09/2026, pendência 9

**Regras e verificações vinculadas:** R-13,HT-09,F-12

### BK-088 — Implementar calendário sazonal versionado

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Representar eventos relevantes do negócio de chocolates.

**O que fazer:** Persistir eventos, períodos e versões por ano.

**Como fazer:**

1. Cadastrar Páscoa, Natal e demais eventos aprovados
2. Usar datas corretas de eventos móveis
3. Separar dados conhecidos na data de cada simulação

**Critério de aceite:** Modelo registra calendário utilizado e reproduz projeção sem informação futura indevida.

**Impacto:** Melhora a interpretação da sazonalidade.

**Dependências:** BK-009, BK-036

**Fonte:** H_Contrato_Dados!A23:H23; T12

### BK-089 — Construir e comparar modelos de projeção

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Selecionar método pelo desempenho observado.

**O que fazer:** Implementar referência simples e candidatos pertinentes.

**Como fazer:**

1. Comparar baseline sazonal e métodos adequados ao volume
2. Manter divisão temporal e parâmetros versionados
3. Medir tempo/memória em hardware local

**Critério de aceite:** Método escolhido tem resultado comparativo reproduzível e atende restrições locais.

**Impacto:** Evita complexidade sem ganho demonstrado.

**Dependências:** BK-087, BK-088, BK-023

**Fonte:** M-046; R_CEI_Tela_12

**Regras e verificações vinculadas:** F-12

### BK-090 — Validar WAPE, Bias e MASE sem vazamento temporal

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Medir qualidade real e condições de utilização.

**O que fazer:** Executar backtest e diagnósticos por período e SKU.

**Como fazer:**

1. Definir fórmulas, sinal do Bias e baseline MASE
2. Tratar denominadores zero e histórico insuficiente
3. Avaliar sazonalidades separadas conforme volume e thresholds aprovados
4. Executar backtest apenas com histórico confiável de 2025 em diante, registrando janelas de treino/teste e restrições de comparação sazonal.

**Critério de aceite:** Métricas vêm de execução real; versão não vira CONFIRMED sem critérios atendidos. Nenhum backtest foi executado na elaboração deste backlog; a tarefa requer os dados reais.

**Impacto:** Evita confiança em percentuais copiados do mockup.

**Dependências:** BK-089, BK-009

**Fonte:** L-20; T-CEI12-01,02,10; Consolidação 13/09/2026, pendência 9

**Regras e verificações vinculadas:** R-13,HT-09,F-12

### BK-091 — Implementar forecast e metas explícitas como planejamento distinto

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Preservar estudo e planejamento como versões distintas.

**O que fazer:** Versionar estudo por SKU, meta financeira e metas diretas existentes sem distribuição inventada.

**Como fazer:**

1. Preservar forecast fundamentado no histórico e identificar o modelo/versão utilizados.
2. Registrar meta de faturamento como referência comercial/cenário; não convertê-la automaticamente em unidades por SKU por mix, preço ou rateio.
3. Quando existir meta direta de produto, registrá-la explicitamente por período e autor, mantendo histórico.
4. Comparar planejamento e forecast sem forçar soma dos SKUs a atingir uma meta financeira por redistribuição automática.

**Critério de aceite:** Meta financeira não fabrica demanda ou produção por SKU; estudo, metas explícitas e realizado têm identidades e versões distintas.

**Impacto:** Evita metas que não fecham com o detalhe.

**Dependências:** BK-090, BK-009, BK-079

**Fonte:** M-048; L-21; R_CEI_Tela_12; Fechamento vigente 13/09/2026, §4.6

**Regras e verificações vinculadas:** F-11

### BK-092 — Implementar produção projetada a partir das vendas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Calcular necessidade futura pela regra aprovada.

**O que fazer:** Calcular reposição/produção líquida por SKU e horizonte considerando o estoque.

**Como fazer:**

1. Usar demanda prevista, estoque-alvo necessário, estoque utilizável e produção válida ainda não incorporada ao saldo; aplicar normalização técnica registrada em BK-009.
2. Respeitar datas de disponibilidade e unidades. Produção concluída já incluída no estoque não é descontada novamente.
3. Executar cálculo cronológico por período usando posição inicial/saldo anterior e entradas válidas; BK-093 apresenta a trajetória desse mesmo plano, sem dependência circular.
4. Necessidade não positiva não cria produção negativa; excesso permanece na projeção de estoque. Dado obrigatório ausente não assume zero.

**Critério de aceite:** Estoque utilizável reduz a necessidade conforme a regra vigente, substituindo o contrato anterior. Resultado é planejamento, sem OP real, e não desconta duas vezes estoque/produção.

**Impacto:** Mantém o planejamento coerente com a decisão operacional.

**Dependências:** BK-091, BK-009, BK-048, BK-077

**Fonte:** M-047; R_CEI_Tela_12; Fechamento vigente 13/09/2026, §4.5

**Regras e verificações vinculadas:** Q-11,F-10

### BK-093 — Implementar evolução do estoque projetado

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Mostrar trajetória de saldo por cenário.

**O que fazer:** Acumular posição inicial, produção e vendas.

**Como fazer:**

1. Usar posição inicial comprovada e o mesmo plano cronológico da necessidade de produção.
2. Evoluir saldo por período com entradas válidas de produção e demanda, garantindo que cada evento conte uma vez.
3. Manter saldo do período anterior como abertura seguinte, com unidade coerente e ausência explicitada; não alimentar BK-092 com seu próprio resultado futuro.

**Critério de aceite:** Trajetória reconcilia saldo anterior + entradas de produção únicas − demanda; produção já incorporada não reaparece como entrada futura.

**Impacto:** Explica o impacto das metas no estoque.

**Dependências:** BK-092, BK-048

**Fonte:** M-047; H_Contrato_Dados!A28:H28; Fechamento vigente 13/09/2026, §4.5

**Regras e verificações vinculadas:** F-10

### BK-094 — Implementar necessidade de insumos e OP Projetada

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Cobrir a aba de insumos para um ano.

**O que fazer:** Converter produção em insumos com composição aprovada.

**Como fazer:**

1. Validar fonte, versão e vigência de BOM/ficha técnica
2. Converter unidades e rendimento/perda conforme regra
3. Marcar OP Projetada e necessidades como planejamento informativo
4. Usar BOM/ficha técnica homologada do Bling, com componente, quantidade e unidade. Yield ou scrap exigem dado e regra industrial próprios; jamais derivar da planilha de perdas operacionais.
5. Entregar demanda bruta de cada insumo por data a Compras, usando necessidade de produção e BOM comprovada; informar quantidades, unidades e dependência do produto que motivou a demanda.

**Critério de aceite:** Composição ausente bloqueia o SKU afetado; projeção não altera estoque, compras ou OP no Bling. BOM incompleta impede resultado confirmado para o SKU afetado; zero ou fator presumido não preenche a lacuna.

**Impacto:** Evita entregar uma aba de planejamento sem base técnica.

**Dependências:** BK-092, BK-048, BK-009

**Fonte:** H_Contrato_Dados!A12:H12; R_CEI_Tela_12; Consolidação 13/09/2026, §11; pendência 7; Fechamento vigente 13/09/2026, §3.1,4.5

**Regras e verificações vinculadas:** Q-24,R-11,HT-07,F-04,F-10

### BK-095 — Implementar resumo financeiro dos cenários

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Relacionar quantidades projetadas a preços e valores.

**O que fazer:** Calcular faturamento e componentes aprovados por cenário.

**Como fazer:**

1. Calcular faturamento previsto somente com quantidades e preços de referência comprováveis, separados da meta financeira comercial.
2. Exibir diferença entre forecast e meta sem rateio automático para forçar igualdade.
3. Identificar custos/parcelas disponíveis e ausentes; cenário não modifica competência, resultado realizado, caixa ou obrigação no Bling.

**Critério de aceite:** Resumo reconcilia com os SKUs que realmente pertencem ao forecast; meta financeira permanece planejamento específico, sem criar fato realizado.

**Impacto:** Permite comparar cenários sem misturar planejamento e resultado.

**Dependências:** BK-091, BK-093, BK-064

**Fonte:** T12 aba Resumo financeiro; R_CEI_Tela_12; Fechamento vigente 13/09/2026, §4.6

**Regras e verificações vinculadas:** Q-25,F-11

### BK-096 — Implementar T12 Projeções e ajuste de metas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Oferecer estudo e planejamento em uma interface completa.

**O que fazer:** Construir gráfico, cenário, filtros e quatro abas.

**Como fazer:**

1. Preservar layout T12, abas, cenário, filtros, gráficos e controles aprovados, usando cobertura histórica real e métricas calculadas.
2. Permitir edição de metas explícitas com capacidade e histórico; a meta financeira total não é distribuída automaticamente em unidades.
3. Mostrar forecast, metas diretas, produção necessária e estoque projetado conforme seus contratos e dados, com ausência explícita.

**Critério de aceite:** Todas as abas funcionam; cartões conferem com contexto total ou subtotal explicitamente indicado.

**Impacto:** Permite testar metas entendendo suas consequências.

**Dependências:** BK-090, BK-091, BK-093, BK-094, BK-095, BK-028

**Fonte:** T12; R_CEI_Tela_12; Fechamento vigente 13/09/2026, §§4–5

**Regras e verificações vinculadas:** F-11

### BK-097 — Exportar o cenário de projeções para Excel

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Produzir artefato de planejamento reproduzível.

**O que fazer:** Gerar XLSX estruturado do snapshot selecionado.

**Como fazer:**

1. Criar Resumo_Cenario, Projecao_Vendas_SKU, Projecao_Producao_SKU e Estoque_Projetado_SKU
2. Incluir Metas, Qualidade_Estudo, OP_Projetada e Metadados
3. Preservar filtros, versões, unidades e marcação PROJECTED

**Critério de aceite:** Valores conciliam com tela; OP é explicitamente projetada e nenhum fato é escrito no Bling.

**Impacto:** Permite discutir planejamento fora do sistema com contexto.

**Dependências:** BK-096, BK-033

**Fonte:** M-049; R_CEI_Tela_12; T-CEI12-11..13

**Regras e verificações vinculadas:** F-14

## 09 Controle e visão executiva

### BK-098 — Implementar registro único de pendências

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Concentrar problemas sem duplicar suas regras.

**O que fazer:** Criar ocorrências e vínculos aos módulos de origem.

**Como fazer:**

1. Modelar estados, responsável, dependência e evidência
2. Deduplicar pela condição e origem
3. Registrar resolução somente pelo fluxo autorizado do tipo
4. Distinguir arquivo rejeitado, origem atrasada, campo indisponível no Bling, resgate não identificado e posição não comprovada; registrar os indicadores afetados sem transformar ausência em zero.

**Critério de aceite:** Um evento mantém um pending_id rastreável; visualizar não fecha a pendência.

**Impacto:** Evita problemas ocultos ou duplicados.

**Dependências:** BK-038, BK-012, BK-034

**Fonte:** M-051,M-054; S_CEI_Tela_13; Consolidação 13/09/2026, §§2–7

**Regras e verificações vinculadas:** F-16

### BK-099 — Implementar T13 Lançamentos e Pendências

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Permitir localizar e resolver exceções.

**O que fazer:** Construir filtros, lista, distribuição e ações contextuais.

**Como fazer:**

1. Definir contagens exclusivas ou sobrepostas com rótulo explícito
2. Abrir origem e fluxo de resolução correto
3. Exibir dependência de regra/dado e responsável

**Critério de aceite:** Total não repete o erro 53 versus soma 58 do mockup; estado não muda sem evidência.

**Impacto:** Facilita tratamento diário das exceções.

**Dependências:** BK-098, BK-028, BK-033

**Fonte:** T13; S_CEI_Tela_13

### BK-100 — Implementar eventos de alerta com regra única de origem

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Chamar atenção para condições reais.

**O que fazer:** Publicar alertas a partir dos módulos responsáveis.

**Como fazer:**

1. Definir identidade, prioridade, posição e ciclo de vida
2. Deduplicar e preservar histórico de resolução/recorrência
3. Não criar thresholds próprios na central ou alertas T10 por inferência

**Critério de aceite:** Mesmo fato liga alert_id, pending_id e origem; nenhuma central recalcula regra divergente.

**Impacto:** Evita ruído e falsos alertas.

**Dependências:** BK-098, BK-051, BK-077, BK-060, BK-012

**Fonte:** M-052,M-054; T_CEI_Tela_14

### BK-101 — Implementar T14 Alertas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Consolidar o que exige atenção.

**O que fazer:** Construir prioridades, módulos, filtros e detalhes.

**Como fazer:**

1. Separar ativos e resolvidos no período
2. Reconciliar distribuição com mesma população
3. Conectar alerta ao módulo de origem e à pendência associada

**Critério de aceite:** Cartões não usam população diferente sem explicar; resolução mantém histórico e responsabilidade.

**Impacto:** Permite priorizar problemas reais.

**Dependências:** BK-100, BK-028

**Fonte:** T14; T_CEI_Tela_14

### BK-102 — Implementar T01 Visão Executiva

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia.

**Objetivo:** Consolidar indicadores aprovados para gestão.

**O que fazer:** Construir a tela a partir da referência aprovada localizada, reutilizando métricas existentes.

**Como fazer:**

1. Preservar blocos, menu e indicadores da última versão aprovada/congelada.
2. Produzido Geral e detalhamento T10 compartilham função, período, snapshot e universo; quantidades com unidades diferentes não formam soma sem significado.
3. Reutilizar resultados de caixa, DRE calculada pelo Majucau, estoque e planejamento; origem atrasada, parcial ou indisponível permanece identificada.

**Critério de aceite:** Cada indicador chega ao módulo e ao detalhe; nenhuma fórmula paralela na dashboard.

**Impacto:** Oferece uma visão geral confiável.

**Dependências:** BK-004, BK-062, BK-067, BK-078, BK-081, BK-096, BK-101

**Fonte:** Menus: Visão Executiva; Fechamento vigente 13/09/2026, §4.2,5

**Regras e verificações vinculadas:** Q-02,F-08

### BK-103 — Implementar os módulos adicionais confirmados no escopo

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Produto + Full Stack.

**Objetivo:** Cobrir todos os itens que a Majucau mantiver para a primeira versão.

**O que fazer:** Localizar contratos aprovados e decompor a implementação dos itens confirmados.

**Como fazer:**

1. Vincular comportamento de Produtos, Simuladores e Relatórios às últimas referências aprovadas; localizar artefato ausente sem redesenhar.
2. Decompor fluxos de implementação e critérios de aceite conforme os contratos existentes.
3. Preservar navegação, permissões e fontes; não inventar rota ou comportamento por texto ilustrativo.

**Critério de aceite:** Nenhum módulo confirmado permanece apenas como link; exclusão exige decisão registrada.

**Impacto:** Evita falsa completude do produto.

**Dependências:** BK-004, BK-021, BK-016

**Fonte:** Menus T12/T13; matriz Pricing/Logística; Fechamento vigente 13/09/2026, §5

**Regras e verificações vinculadas:** Q-02

## 10 Verificação integrada

### BK-104 — Implementar exportações financeiras e de controle aprovadas

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Full Stack.

**Objetivo:** Transformar botões homologados em relatórios completos.

**O que fazer:** Construir contratos de exportação por tela, endpoint, gerador e integração do botão.

**Como fazer:**

1. Implementar os formatos/colunas definidos nas referências aprovadas para T02–T07 e T13–T15, incluindo Excel quando especificado.
2. Consumir mesmos dados, filtros, snapshot, unidade e período da tela, com capacidade de exportação.
3. Conferir totais, tipos, tratamento seguro de conteúdo e falhas de download; ausência de dado não é zero.

**Critério de aceite:** Botões de exportação aprovados permanecem funcionais e geram conteúdo coerente com a tela; não exigir nova decisão sobre uma função já congelada.

**Impacto:** Evita aprovar uma interface com botão de exportação sem implementação.

**Dependências:** BK-012, BK-062, BK-057, BK-059, BK-055, BK-067, BK-069, BK-099, BK-101, BK-052

**Fonte:** T02–T07; T13–T15; L-22; Fechamento vigente 13/09/2026, §5.7

**Regras e verificações vinculadas:** F-14

### BK-105 — Rastrear cada requisito até tarefa e teste

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** QA + Arquitetura.

**Objetivo:** Garantir cobertura do material de ponta a ponta.

**O que fazer:** Relacionar regras da matriz e CEIs às entregas.

**Como fazer:**

1. Mapear M-001..M-058, C-01..03, L-01..26 e testes I_Testes
2. Complementar contratos incompletos de caixa, metas e investimentos
3. Marcar regras sem evidência de aceite e corrigir lacunas
4. Vincular R-01 a R-14 e HT-01 a HT-10 aos contratos, tarefas e cenários novos; identificar cenários da versão anterior superados pela mudança de fontes.
5. Vincular F-01 a F-20 ao código futuro e critérios de aceite; regras antigas superadas permanecem apenas em histórico, sem bloquear novamente o negócio.

**Critério de aceite:** Toda regra do escopo tem responsável, implementação prevista e validação específica.

**Impacto:** Evita perder requisitos espalhados na planilha.

**Dependências:** BK-016, BK-019

**Fonte:** A_Matriz_Mestra; H_Contrato_Dados; I_Testes; Consolidação 13/09/2026, consolidação integral; Fechamento vigente 13/09/2026, integral

### BK-106 — Validar regras de domínio e casos de fronteira

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** QA + Backend.

**Objetivo:** Provar os cálculos fora da interface.

**O que fazer:** Executar testes unitários e de invariantes relevantes.

**Como fazer:**

1. Testar centavos, arredondamento, zero/ausente, datas e estados
2. Validar conflitos resolvidos com exemplos homologados
3. Cobrir idempotência, estornos e real versus projetado
4. Cobrir principal aportado menos resgates, rendimento separado, limite executável pelo Itaú, frete sem compensação silenciosa e perdas operacionais sem efeito automático na BOM.
5. Cobrir D-1/exceção corrente, DRE/balancete Bling, compra parcial/tardia, produção líquida, plano temporal ausente, meta financeira sem rateio e histórico fechado preservado.

**Critério de aceite:** Falhas materiais são corrigidas; números de mockup não servem como oráculo quando inconsistentes. Tratar ausência, ambiguidade e cobertura temporal insuficiente com resultados explícitos.

**Impacto:** Detecta erros de cálculo antes da operação.

**Dependências:** BK-105, BK-072, BK-068, BK-084, BK-093, BK-065

**Fonte:** I_Testes: catálogo previsto, ainda não executado; Consolidação 13/09/2026, §§6–12; Fechamento vigente 13/09/2026, §§2–4,6,8

### BK-107 — Validar integrações, concorrência e reinício

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** QA + Integrações.

**Objetivo:** Provar comportamento sob falhas reais de sistema.

**O que fazer:** Testar banco, arquivos, adaptadores e jobs em ambiente controlado.

**Como fazer:**

1. Simular 401,429,timeout, arquivo corrompido e versão nova
2. Repetir carga e interromper processo durante processamento
3. Disputar reprocessamento e mudança manual concorrente
4. Verificar em ambiente controlado leitura com permissões restritas, versão corrigida, mesmo conteúdo com outro nome, períodos sobrepostos, reprocessamento e interrupção; conferir fonte intocada antes/depois.
5. Validar serviços e retomada após reinício sem sessão humana logada; repetir eventos/arquivos sem duplicar recebimento, perda, frete, lançamento ou KPI.

**Critério de aceite:** Último snapshot válido sobrevive; fatos não duplicam; retomada possui evidência. Cobrir também resgate parcial/total, posição ausente e venda/recebível/liquidação sem duplicidade; registrar evidências na implementação futura.

**Impacto:** Reduz falhas de sincronização em produção.

**Dependências:** BK-052, BK-044, BK-041, BK-054

**Fonte:** ; Consolidação 13/09/2026, §§1–2,6–10; Fechamento vigente 13/09/2026, §§7–8

**Regras e verificações vinculadas:** F-20

### BK-108 — Validar fidelidade visual e acessibilidade

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** UX + QA.

**Objetivo:** Conferir a experiência completa nos dispositivos previstos.

**O que fazer:** Comparar telas implementadas e decisões UX aprovadas.

**Como fazer:**

1. Comparar cada tela com o mockup aprovado/congelado, preservando blocos, posições, cores, tipografia e navegação.
2. Conferir card/tabela/gráfico/detalhe/exportação com a mesma métrica e conjunto de dados; números do mockup são ilustrativos.
3. Verificar teclado, foco, mensagens e estados de ausência sem redesenho; corrigir implementação divergente.

**Critério de aceite:** Todas as telas do escopo passam pelos critérios acordados; alterações de estrutura têm decisão registrada.

**Impacto:** Evita entregar apenas uma aparência semelhante sem usabilidade.

**Dependências:** BK-102, BK-103, BK-074, BK-075, BK-076, BK-086, BK-037, BK-035, BK-055, BK-057, BK-059, BK-069, BK-073, BK-099

**Fonte:** Todas as imagens e decisões UX; Fechamento vigente 13/09/2026, §5

**Regras e verificações vinculadas:** F-13

### BK-109 — Validar arquivos exportados e proteção do conteúdo

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** QA.

**Objetivo:** Garantir que relatórios correspondam à tela e ao perfil.

**O que fazer:** Conferir contratos XLSX/CSV/PDF apenas onde aprovados.

**Como fazer:**

1. Validar colunas, tipos, filtros, totais e metadados
2. Proteger células de texto contra injeção de fórmulas
3. Abrir amostras no programa usado pela Majucau

**Critério de aceite:** Exports não incluem segredos nem linhas sem permissão; dados conferem com snapshot.

**Impacto:** Evita relatório incorreto ou exposição de informação.

**Dependências:** BK-097, BK-086, BK-012, BK-104

**Fonte:** Requisito técnico para entregar e operar o produto.

**Regras e verificações vinculadas:** F-14

### BK-110 — Executar jornadas completas de ponta a ponta

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** QA + Majucau.

**Objetivo:** Provar que os módulos trabalham juntos.

**O que fazer:** Testar entradas reais controladas até resultados e resolução.

**Como fazer:**

1. Validar fluxo Bling + agenda futura → conciliação → caixa projetado/disponibilidade → calculadora, com D-1 e exceção explícita.
2. Validar em caminho independente a DRE calculada pelo Majucau contra a referência do Bling na mesma competência/posição, sem reconstrução a partir do caixa.
3. Validar forecast por SKU → estoque/produção necessária → BOM/insumos → compras parciais e temporais → exportação.
4. Validar permissões, metas, vigência, período fechado, reprocessamento auditado e recuperação sem alterar arquivos de origem.
5. Validar perda operacional → vínculo ao fato Bling ou pendência → correção/reprocessamento idempotente; manter a trilha externa sem criar um segundo cálculo de DRE fora do motor Majucau.

**Critério de aceite:** Jornadas conferem IDs, valores, permissões, estados e histórico, com evidência reproduzível.

**Impacto:** Demonstra funcionamento do produto completo.

**Dependências:** BK-106, BK-107, BK-108, BK-109

**Fonte:** ; Fechamento vigente 13/09/2026, §§1–8; Fechamento vigente 13/09/2026, §8.14

### BK-111 — Validar controles de segurança antes da produção

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** Segurança + QA.

**Objetivo:** Verificar que restrições resistem a acesso indevido.

**O que fazer:** Executar testes direcionados do modelo de ameaças.

**Como fazer:**

1. Testar bypass de API, troca de perfil, sessão revogada e recuperação
2. Revisar uploads, exportações, path traversal e exposição de logs
3. Verificar dependências e segredos no pacote/artefatos
4. Verificar em ambiente de teste que a identidade e o adaptador do Drive não conseguem mutar a origem; inspecionar escopos, caminhos e operações, sem tentar escrita destrutiva na pasta real.

**Critério de aceite:** Nenhuma falha crítica sem tratamento; perfil somente leitura não consegue mutar via chamada direta. Nem usuário administrador do Majucau ganha escrita no Drive por reprocessar um lote.

**Impacto:** Reduz risco de comprometimento de dados.

**Dependências:** BK-035, BK-037, BK-052, BK-024

**Fonte:** ; Consolidação 13/09/2026, §2

### BK-112 — Validar desempenho no hardware local

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** Engenharia + Infraestrutura.

**Objetivo:** Demonstrar capacidade de uso e processamento.

**O que fazer:** Medir jornadas com volume e concorrência esperados.

**Como fazer:**

1. Medir percentis de resposta, importação, forecast, CPU, memória e disco
2. Validar carga simultânea sem congelar consultas
3. Otimizar índices e consultas somente nos gargalos medidos

**Critério de aceite:** Metas aprovadas de tempo e capacidade são atingidas; limites conhecidos ficam documentados.

**Impacto:** Evita lentidão percebida após a instalação.

**Dependências:** BK-110, BK-014, BK-023, BK-117

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-113 — Homologar todos os módulos com a Majucau

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** Majucau + QA.

**Objetivo:** Confirmar que o produto atende às regras e ao uso diário.

**O que fazer:** Executar roteiro com dados e cenários reconhecidos pelos responsáveis.

**Como fazer:**

1. Conferir amostras com fontes oficiais e responsáveis do negócio
2. Resolver discrepâncias e repetir apenas verificações afetadas
3. Registrar aceite de cada módulo e de cada conflito encerrado

**Critério de aceite:** Todos os módulos confirmados têm aceite; não se substitui pendência crítica por zero ou tela vazia.

**Impacto:** Evita lançar um sistema tecnicamente ativo, mas incorreto para a operação.

**Dependências:** BK-110, BK-111, BK-112

**Fonte:** Requisito técnico para entregar e operar o produto.

## 11 Instalação, produção e sustentação

### BK-114 — Definir a topologia final da instalação local

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Infraestrutura.

**Objetivo:** Tornar operação, acesso e recuperação viáveis.

**O que fazer:** Desenhar aplicação, banco, arquivos e rede no equipamento escolhido.

**Como fazer:**

1. Projetar aplicação, banco e jobs como serviços independentes de login de funcionário no Windows local.
2. Documentar host, sistema, serviços, inicialização, portas, dependências, dados persistentes e atualização.
3. Usar autenticação e credenciais somente no backend; acesso externo, se habilitado expressamente, usa canal seguro sem exposição direta de banco, segredos ou interfaces internas.
4. Documentar saída para Bling/Drive, armazenamento interno, ausência de internet e recuperação.

**Critério de aceite:** Topologia opera sem sessão humana e tem inicialização controlada; este backlog não habilita acesso externo nem instala serviço.

**Impacto:** Evita exposição acidental e dependência não planejada.

**Dependências:** BK-014, BK-017, BK-024, BK-023

**Fonte:** ; Consolidação 13/09/2026, §§1–2; Fechamento vigente 13/09/2026, §7

**Regras e verificações vinculadas:** F-17

### BK-115 — Preparar o servidor e a rede local

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Infraestrutura.

**Objetivo:** Disponibilizar um ambiente estável para instalação.

**O que fazer:** Configurar sistema, contas técnicas, armazenamento e acesso.

**Como fazer:**

1. Aplicar atualizações e permissões mínimas
2. Configurar DNS/TLS/firewall e acesso remoto aprovado
3. Validar energia, espaço e reinício automático conforme necessidade

**Critério de aceite:** Serviços não executam com privilégio excessivo; somente acessos planejados alcançam a aplicação.

**Impacto:** Reduz paradas e riscos da máquina local.

**Dependências:** BK-114

**Fonte:** Requisito técnico para entregar e operar o produto.

**Regras e verificações vinculadas:** F-17

### BK-116 — Gerar pacote versionado de instalação

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Engenharia + Infraestrutura.

**Objetivo:** Entregar uma versão reproduzível e identificável.

**O que fazer:** Empacotar Go, build React, migrações e dependências aprovadas.

**Como fazer:**

1. Gerar artefato em CI confiável
2. Registrar versão, hash e inventário de dependências
3. Incluir configuração exemplo e instruções sem credenciais

**Critério de aceite:** Pacote instala em ambiente limpo; versão visível corresponde ao artefato validado.

**Impacto:** Facilita suporte e retorno à versão anterior.

**Dependências:** BK-030, BK-114

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-117 — Instalar ambiente local de homologação

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Infraestrutura.

**Objetivo:** Ensaiar instalação e atualização sem usar produção.

**O que fazer:** Criar ambiente e dados separados no hardware previsto ou equivalente.

**Como fazer:**

1. Instalar pacote e executar migrações
2. Configurar fontes de teste e credenciais separadas quando disponíveis
3. Verificar que nenhum job de teste altera a operação real

**Critério de aceite:** Ambiente reproduz instalação documentada e permite validar desempenho e restauração.

**Impacto:** Reduz surpresas na entrada em produção.

**Dependências:** BK-115, BK-116

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-118 — Configurar backup de banco, arquivos e configuração

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Infraestrutura.

**Objetivo:** Preservar tudo que é necessário para recuperar o serviço.

**O que fazer:** Implantar cópias automáticas conforme perda tolerável de dados.

**Como fazer:**

1. Incluir banco, fontes importadas, versões e configurações críticas
2. Guardar cópia separada do disco e da máquina de produção
3. Definir retenção, proteção das cópias e verificação de falha
4. Copiar banco, metadados e evidências internas necessárias para destino de backup próprio; as três estruturas do Drive são somente leitura e não recebem cópias de segurança do sistema.
5. Incluir banco, configurações, parâmetros, metas, usuários/capacidades, registros de processamento, períodos fechados, logs necessários e dados locais não reconstruíveis pela origem.

**Critério de aceite:** Rotina gera evidência de backup completo e alerta quando falha; cópia no mesmo disco não é a única proteção. Restauração do Majucau não precisa escrever, mover ou renomear arquivos na origem.

**Impacto:** Permite recuperar o sistema após falha do equipamento.

**Dependências:** BK-014, BK-117, BK-018

**Fonte:** ; Consolidação 13/09/2026, §2; Fechamento vigente 13/09/2026, §7.5

**Regras e verificações vinculadas:** R-02,F-17

### BK-119 — Testar restauração e recuperação do ambiente

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** Infraestrutura + QA.

**Objetivo:** Provar que o backup funciona no prazo necessário.

**O que fazer:** Restaurar em instalação separada e conferir dados.

**Como fazer:**

1. Simular perda de banco/arquivo e indisponibilidade da máquina
2. Reinstalar versão compatível e restaurar cópias
3. Conferir totais, arquivos, permissões e jobs antes de reativar integrações
4. Restaurar histórico de hashes, lotes e versões antes da retomada; verificar que reler o Drive não duplica fatos nem altera arquivos externos.
5. Validar integridade de banco, configuração, permissões, fechamentos e histórico após restauração; documentar como subir os serviços sem sessão humana.

**Critério de aceite:** Tempo e perda real medidos atendem os limites aprovados; procedimento é executável pelo responsável.

**Impacto:** Evita descobrir backup inválido durante uma falha.

**Dependências:** BK-118, BK-116

**Fonte:** ; Consolidação 13/09/2026, §2; Fechamento vigente 13/09/2026, §7.6

**Regras e verificações vinculadas:** F-17

### BK-120 — Configurar monitoramento operacional local

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Infraestrutura.

**Objetivo:** Perceber falhas antes que invalidem a operação.

**O que fazer:** Monitorar serviço, banco, disco, sincronização e backups.

**Como fazer:**

1. Monitorar serviço, banco, disco, processamento e backup; registrar fonte, execução, horário, sucesso/falha, arquivo/lote e identificadores.
2. Diagnosticar aplicação parada, API indisponível, token vencido, rejeição, duplicidade e processamento bloqueado sem revelar credenciais.
3. Documentar procedimentos de resposta e suporte; aplicar limites técnicos justificados e alertas pelo canal habilitado.

**Critério de aceite:** Falha simulada chega ao responsável pelo canal acordado; diagnóstico identifica ação necessária.

**Impacto:** Reduz tempo de indisponibilidade e dados desatualizados.

**Dependências:** BK-031, BK-051, BK-118

**Fonte:** ; Fechamento vigente 13/09/2026, §7.7–7.8

**Regras e verificações vinculadas:** Q-26,F-17

### BK-121 — Preparar e reconciliar a carga inicial

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Dados + Majucau.

**Objetivo:** Começar a produção com base conhecida e validada.

**O que fazer:** Planejar corte temporal, importação histórica e snapshots iniciais.

**Como fazer:**

1. Definir data de corte e ordem das cargas
2. Importar histórico, aberturas, parâmetros, usuários e metas
3. Reconciliar amostras e totais com fontes sem criar duplicata
4. Carregar somente fontes aprovadas; projeções usam histórico confiável de 2025 em diante. Registrar ausência de posição, custos históricos, BOM ou aberturas sem preenchê-los por hipótese.

**Critério de aceite:** Carga inicial tem cobertura, pendências conhecidas e aceite por domínio; fonte ausente permanece explícita.

**Impacto:** Evita lançar sobre dados incompletos sem perceber.

**Dependências:** BK-113, BK-117, BK-044

**Fonte:** ; Consolidação 13/09/2026, §§1–2; pendência 9

### BK-122 — Ensaiar atualização e retorno à versão anterior

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Infraestrutura + Engenharia.

**Objetivo:** Controlar mudanças futuras sem perder informação.

**O que fazer:** Criar roteiro de release, migração e recuperação.

**Como fazer:**

1. Verificar compatibilidade de esquema e pacote
2. Fazer backup prévio e definir ponto de retorno
3. Testar rollback de aplicação e recuperação de dados quando migração não for reversível

**Critério de aceite:** Procedimento identifica quando rollback simples é possível e quando requer restauração; sem prometer reversão destrutiva automática.

**Impacto:** Reduz risco de atualização local.

**Dependências:** BK-116, BK-119

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-123 — Documentar e treinar a rotina da Majucau

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Desenvolvimento. **Responsável sugerido:** Produto + Infraestrutura + Majucau.

**Objetivo:** Permitir uso e suporte sem depender de memória da equipe.

**O que fazer:** Produzir guia por perfil e manual de operação.

**Como fazer:**

1. Documentar uso, capacidades, metas, vigência, período fechado, reprocessamento e auditoria.
2. Explicar Drive somente de leitura, arquivo corrigido, rejeição e repetição segura; movimentos normais usam D-1 e dia corrente exige solicitação expressa.
3. Entregar diagnóstico de aplicação, banco, API, token, arquivos, processamento e recuperação dos serviços; realizar exercício com responsáveis na implantação.

**Critério de aceite:** Usuários executam tarefas críticas; operador sabe diagnosticar falha e acionar recuperação.

**Impacto:** Reduz erros e dependência informal.

**Dependências:** BK-113, BK-120, BK-122

**Fonte:** ; Consolidação 13/09/2026, §§1–2,10; Fechamento vigente 13/09/2026, §§7–8

**Regras e verificações vinculadas:** F-17,F-19

### BK-124 — Verificar os critérios de entrada em produção

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Qualidade. **Responsável sugerido:** Majucau + Liderança técnica.

**Objetivo:** Confirmar que toda a primeira versão está pronta.

**O que fazer:** Reunir evidências funcionais, técnicas e operacionais.

**Como fazer:**

1. Conferir todos os módulos e conflitos do escopo
2. Validar segurança, carga inicial, desempenho, restauração e responsável
3. Registrar versão, decisão de entrada e plano de contingência

**Critério de aceite:** Nenhum requisito obrigatório permanece sem evidência; Majucau autoriza a entrada da versão concreta.

**Impacto:** Evita liberar com módulo faltante ou recuperação não testada.

**Dependências:** BK-121, BK-111, BK-119, BK-122, BK-123, BK-120

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-125 — Instalar e ativar a primeira versão em produção local

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Produção. **Responsável sugerido:** Infraestrutura + Engenharia.

**Objetivo:** Colocar a versão homologada em uso real.

**O que fazer:** Executar o roteiro aprovado no ambiente de produção.

**Como fazer:**

1. Instalar pacote, configurar segredos e aplicar migrações
2. Ativar fontes e jobs na sequência planejada
3. Executar verificações rápidas de acesso, dados, módulos e backup

**Critério de aceite:** Versão esperada está acessível aos usuários autorizados e produz dados reconciliados; roteiro de contingência disponível.

**Impacto:** Inicia a operação completa da Majucau.

**Dependências:** BK-124

**Fonte:** Requisito técnico para entregar e operar o produto.

### BK-126 — Acompanhar os primeiros ciclos reais e encerrar a entrega

**Situação:** Não iniciada. **Evolução:** 0%. **Tipo:** Produção. **Responsável sugerido:** Majucau + Engenharia + Infraestrutura.

**Objetivo:** Confirmar estabilidade no uso cotidiano.

**O que fazer:** Observar conciliação, fechamento, projeção e atualização de fontes.

**Como fazer:**

1. Validar ciclos acordados com o negócio
2. Registrar e corrigir incidentes com evidência
3. Transferir responsabilidade operacional e rotina de manutenção

**Critério de aceite:** Critérios de estabilização acordados são atendidos; suporte, backups e atualizações têm responsáveis.

**Impacto:** Consolida a produção e evita abandono após a instalação.

**Dependências:** BK-125

**Fonte:** Requisito técnico para entregar e operar o produto.
