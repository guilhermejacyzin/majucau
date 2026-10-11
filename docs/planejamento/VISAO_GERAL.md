# Majucau Intelligence

Planejamento do sistema de gestão e análise da Majucau. A **versão 2** é a referência vigente do backlog. O código do produto já está em desenvolvimento; a conclusão global ainda não foi auditada tarefa a tarefa.

## Comece aqui

| Arquivo | Conteúdo |
|---|---|
| [Backlog detalhado v2](outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md) | 127 tarefas, da BK-000 à BK-126: objetivo, o que fazer, como fazer, impacto, dependências e aceite. |
| [Planilha de evolução v2](outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx) | Tarefas, evolução calculada, decisões e cobertura dos módulos. |
| [Análise e decisões v2](outputs/MAJUCAU_ANALISE_E_DECISOES_v2.md) | Regras vigentes, alterações do fechamento e dez frentes de verificação técnica futura. |
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

### Estado atual do backlog — 11/10/2026

O backlog detalhado registra **6 tarefas concluídas, 11 em andamento e 110 não iniciadas**. As tarefas em andamento são BK-001 (0%), BK-006 (30%), BK-007 (20%), BK-012 (10%), BK-013 (15%), BK-014 (15%), BK-024 (30%), BK-025 (70%), BK-031 (80%), BK-041 (60%) e BK-067 (20%). Pela média simples das evoluções registradas, o acompanhamento estima **7,5% concluído e 92,5% restante**. Essa média é uma estimativa de planejamento; não comprova prontidão para homologação ou produção.

Há código e evidências de CI para partes da fundação, mas ainda faltam validação com dados reais, homologação e revisão das tarefas não iniciadas. Não há prazo global confiável. Consulte o backlog v2 e `REQUIREMENTS-TRACEABILITY.md` antes de usar os números como critério de liberação.
