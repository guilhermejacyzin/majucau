# Majucau Intelligence

Planejamento do sistema de gestão e análise da Majucau. A **versão 2** é a referência vigente do backlog. O código do produto já está em desenvolvimento; a conclusão global ainda não foi auditada tarefa a tarefa.

## Comece aqui

| Arquivo | Conteúdo |
|---|---|
| [Backlog detalhado v2](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md) | 127 tarefas, da BK-000 à BK-126: objetivo, o que fazer, como fazer, impacto, dependências e aceite. |
| [Planilha de evolução v2](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx) | Tarefas, evolução calculada, decisões e cobertura dos módulos. |
| [Análise e decisões v2](../../artifacts/planejamento/outputs/MAJUCAU_ANALISE_E_DECISOES_v2.md) | Regras vigentes, alterações do fechamento e dez frentes de verificação técnica futura. |
| [Limites da arquitetura Go](../architecture/GO-MODULAR-BOUNDARIES.md) | Módulos, contratos, dependências e transações verificadas em BK-017. |
| [Mapa inicial de contextos](../architecture/DOMAIN-CONTEXT-MAP-v1.md) | Responsabilidades técnicas documentadas e pontos de negócio que ainda precisam de validação em BK-016. |
| [Índice completo](README.md) | Histórico, auditorias, dados estruturados, prévias e scripts. |
| [Fontes e rastreabilidade](FONTES.md) | Inventário e identificação das referências usadas. |

## Escopo confirmado

- Uso exclusivo da Majucau, por três diretores. O cliente Wails/Windows com PostgreSQL local é a etapa atual de transição; a direção aprovada é a hospedagem privada centralizada (ADR-002), com detalhes operacionais ainda pendentes.
- Monólito com DDD e arquitetura hexagonal, backend em Go e frontend em React + TypeScript + Vite (conforme ADR-001 e código atual).
- Primeira entrada em produção com os módulos confirmados no material. Balanço Patrimonial está excluído.
- Bling como fonte de lançamentos; o motor Majucau calcula a DRE conforme regras contábeis aprovadas e o relatório do Bling serve para comparação. Três estruturas específicas do Google Drive permanecem somente de leitura.
- O fechamento funcional vigente está registrado na v2. Regras antigas substituídas nas v0/v1 são histórico.

## Evolução registrada

O backlog contém 127 tarefas, de BK-000 a BK-126. Os números **2,4% por quantidade** e **0% de desenvolvimento** pertencem à revisão de 13/09/2026 e são históricos. Na revisão de 11/10/2026, a matriz de requisitos tinha 51 itens: 1 verificado, 4 documentados, 26 parciais, 11 bloqueados e 9 não iniciados. A auditoria tarefa a tarefa segue em andamento.

Na planilha, uma tarefa pontua 100% somente quando está concluída e possui evidência de aceite. Os pesos iniciais são iguais a 1. O acompanhamento não é sincronizado automaticamente com commits, issues ou projetos do GitHub.

### Estado atual do backlog — 19/10/2026

O backlog registra **7 tarefas concluídas, 20 em andamento e 100 não iniciadas**. As tarefas em andamento são BK-006 (30%), BK-007 (20%), BK-012 (30%), BK-013 (20%), BK-014 (15%), BK-016 (20%), BK-017 (95%), BK-024 (35%), BK-025 (70%), BK-026 (60%), BK-027 (40%), BK-031 (85%), BK-039 (50%), BK-041 (80%), BK-042 (70%), BK-043 (40%), BK-056 (5%), BK-066 (5%), BK-067 (25%) e BK-102 (1%). Pela média simples das 127 tarefas, o acompanhamento estima **11,8% concluído e 88,2% restante**. Essa média é uma estimativa de planejamento; não comprova prontidão para homologação ou produção.

BK-016 avançou para 20% com um mapa técnico inicial dos sete contextos. A propriedade final dos fatos, agregados e eventos de negócio ainda aguarda validação. BK-017 avançou com o registro dos limites arquiteturais e a propriedade explícita do pool PostgreSQL pelo worker; o fechamento formal continua dependendo de BK-016.

O worker verifica se o PostgreSQL responde e informa degradação com um código seguro quando a conexão falha. A rotina do Bling às 09:30, com retomada idempotente e proteção entre instâncias, passou localmente em Windows e usa PostgreSQL descartável na CI. A paginação de contas a receber/pagar e categorias agora usa o limite documentado: página cheia continua; página curta encerra, mesmo sem metadados adicionais. Em `/caixas`, sem limite documentado ou metadados, uma página vazia continua sendo falha segura. A conexão real e homologação do Bling ainda faltam. Para o Drive, a leitura do inventário e a importação CSV de recebimentos futuros já estão ligadas ao worker diário; ainda faltam a credencial protegida, a sincronização real, os parsers dos outros formatos e a homologação. Não há prazo global confiável. Consulte o backlog v2 e `REQUIREMENTS-TRACEABILITY.md` antes de usar os números como critério de liberação.

**Atualização de 18/10/2026:** essa posição foi avançada. O worker agora percorre, com checkpoints retomáveis e publicação atômica, as subpastas encontradas dentro das três raízes aprovadas. O caminho de CSV de recebimentos futuros está conectado ao staging privado e à importação parcial idempotente; arquivos antigos ficam pendentes. A pasta aprovada está em `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`; seus IDs foram identificados na conta, mas não foram publicados no Git. A identidade técnica e sua credencial DPAPI ainda não estão configuradas. O CSV encontrado estava desatualizado, então nenhum valor real foi baixado ou importado. XLS/XLSX e os contratos de todas as fontes continuam pendentes. A suíte PostgreSQL completa passou localmente em uma instância descartável; a correção das falhas encontradas na CI aguarda nova execução após o push.

O importador de recebimentos futuros também precisava da identidade interna `NUVEM_PAGO`; a migration 000008 agora a cria como `UNAVAILABLE`, sem indicar conexão real à API. O teste de integração PostgreSQL verifica a identidade migrada. Na data daquele registro, o Drive ainda não estava ligado ao importador; a atualização recursiva e a integração CSV estão registradas na seção de 18/10 acima e em `docs/evidence/GOOGLE-DRIVE-NESTED-IMPORT-2026-10-18.md`.

**Atualização de 19/10/2026:** foi implementada a base aritmética da DRE do Majucau para as linhas já aprovadas. Ela só recebe valores que outro componente já classificou, usa decimal exato e deixa totais dependentes como indisponíveis se faltar uma entrada. A tarefa BK-066 passou a 5%; nenhum dado real foi ligado à DRE, nenhuma conta foi mapeada e os resultados continuam provisórios. Permanecem pendentes a classificação contábil validada, a ligação com lançamentos reais, a conciliação com o Bling e a publicação rastreável dos resultados. A tela não foi alterada. O Drive confirmou a conta `financeiro@majucau.com.br`; a primeira conferência não abriu arquivos.

**Conferência do Drive em 19/10/2026:** em leitura somente, confirmei um `.xlsx` operacional em `01_perdas_operacionais` e CSVs nas estruturas de Recebimentos Futuros e Frete. O Excel operacional inclui contato e anotações de entrega, sem valor monetário por ocorrência; não serve sozinho como comprovação de perda na DRE. O leitor XLSX atual está conectado somente a Recebimentos Futuros, e Frete ainda não tem importador contratado. Os arquivos estavam desatualizados (modificados em 11/09/2026) e continuam pendentes. Nenhum dado de linha foi gravado no Git ou importado. Detalhes e grupos de colunas: `docs/evidence/GOOGLE-DRIVE-READONLY-SCHEMA-INSPECTION-2026-10-19.md`.

Em 13/10/2026, a revisão do código confirmou suporte ao cabeçalho JWT e à URL atual da API do Bling. Naquele diagnóstico local, o salvamento retornou `BLING_SECRET_REQUIRED`; a conta não foi autorizada e não houve leitura financeira. O Client Secret do aplicativo precisa ser informado diretamente no Majucau e é diferente da senha para entrar no site do Bling. Mesmo após conectar, a sincronização atual preserva páginas RAW, mas ainda não alimenta os dados normalizados do painel. Detalhes: `docs/evidence/BLING-CONNECTION-JWT-READINESS-2026-10-13.md`.

Em 18/10/2026, a tela passou a recuperar o Client ID e o endereço de retorno salvos; se estiverem vazios, agora mostra os dados públicos do aplicativo Majucau já fornecidos pela usuária. Esses dados só são gravados quando ela salva a configuração, e valores anteriormente registrados continuam sendo usados. A chave secreta continua sem preenchimento automático e protegida no cofre local. A conexão real ainda depende de informar o Client Secret do aplicativo no Majucau, autorizar OAuth, testar a conta e homologar a paginação; os dados ainda precisam ser ligados aos relatórios. A tela aprovada e as regras financeiras não foram alteradas. Evidências: `docs/evidence/BLING-CONFIG-READBACK-2026-10-18.md`, `docs/evidence/BLING-SECRET-STATUS-PLACEHOLDER-FIX-2026-10-18.md`, `docs/evidence/BLING-PUBLIC-CONFIG-PENDING-SECRET-FIX-2026-10-18.md`, `docs/evidence/BLING-INSTALLER-UPDATED-2026-10-18.md`.

Em 18/10/2026, identifiquei que o instalador NSIS distribuído localmente ainda era de 13/10, embora o código já tivesse recebido correções da configuração do Bling. Reconstruí o instalador e o pacote portátil; a versão nova também preenche os dados públicos do aplicativo Majucau quando o cadastro está vazio. Depois da instalação, ainda é preciso salvar a configuração e, se o Majucau informar que o segredo não está salvo, inserir o Client Secret do aplicativo diretamente nele. A senha de entrar no site Bling não é esse segredo. O pacote é para homologação local, não tem assinatura de produção e ainda não foi conectado à conta real. Evidência e hashes: `docs/evidence/BLING-INSTALLER-UPDATED-2026-10-18.md`.

Os arquivos externos diários também incluem Excel `.XLS` e `.XLSX`. Há leitores estruturais em fluxo para ambos os formatos no caminho de recebimentos futuros, com limites de leitura e sem execução de macros ou importação de valores. No XLSX, cada token XML também é limitado antes de ser acumulado pelo decoder. Os arquivos continuam pendentes: os leitores ainda não interpretam dados financeiros sem contrato aprovado, não estão homologados com amostras reais e a sincronização real do Drive ainda depende da identidade técnica. Ver `docs/evidence/DRIVE-STREAMING-IMPORT-2026-10-13.md`, `docs/evidence/GOOGLE-DRIVE-READ-ADAPTER-2026-10-18.md`, `docs/evidence/EXCEL-STREAMING-PARSER-PROTOTYPE-2026-10-18.md`, `docs/evidence/DRIVE-XLS-STREAMING-PENDING-2026-10-19.md` e `docs/evidence/DRIVE-XLSX-XML-TOKEN-LIMIT-2026-10-19.md`.
