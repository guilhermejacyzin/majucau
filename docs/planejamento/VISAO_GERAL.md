# Majucau Intelligence

Planejamento do sistema de gestão e análise da Majucau. A **versão 2** é a referência vigente do backlog. O código do produto já está em desenvolvimento; a conclusão global ainda não foi auditada tarefa a tarefa.

## Comece aqui

| Arquivo | Conteúdo |
|---|---|
| [Backlog detalhado v2](outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md) | 127 tarefas, da BK-000 à BK-126: objetivo, o que fazer, como fazer, impacto, dependências e aceite. |
| [Planilha de evolução v2](outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx) | Tarefas, evolução calculada, decisões e cobertura dos módulos. |
| [Análise e decisões v2](outputs/MAJUCAU_ANALISE_E_DECISOES_v2.md) | Regras vigentes, alterações do fechamento e dez frentes de verificação técnica futura. |
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

### Estado atual do backlog — 13/10/2026

O backlog registra **7 tarefas concluídas, 18 em andamento e 102 não iniciadas**. As tarefas em andamento são BK-006 (30%), BK-007 (20%), BK-012 (30%), BK-013 (20%), BK-014 (15%), BK-016 (20%), BK-017 (90%), BK-024 (35%), BK-025 (70%), BK-026 (55%), BK-027 (25%), BK-031 (85%), BK-041 (75%), BK-042 (40%), BK-043 (40%), BK-056 (5%), BK-067 (20%) e BK-102 (1%). Pela média simples das 127 tarefas, o acompanhamento estima **10,8% concluído e 89,2% restante**. Essa média é uma estimativa de planejamento; não comprova prontidão para homologação ou produção.

BK-016 avançou para 20% com um mapa técnico inicial dos sete contextos. A propriedade final dos fatos, agregados e eventos de negócio ainda aguarda validação. BK-017 avançou com o registro dos limites arquiteturais e a propriedade explícita do pool PostgreSQL pelo worker; o fechamento formal continua dependendo de BK-016.

O worker agora verifica se o PostgreSQL responde e informa degradação com um código seguro quando a conexão falha. Ainda faltam validar schema e permissões em instalação/VM, conectar e homologar o Bling com a conta real, confirmar o fim da paginação, configurar acesso persistente ao Drive e revisar as tarefas não iniciadas. Não há prazo global confiável. Consulte o backlog v2 e `REQUIREMENTS-TRACEABILITY.md` antes de usar os números como critério de liberação.
