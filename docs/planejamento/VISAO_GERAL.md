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
- Monólito com DDD e arquitetura hexagonal, backend em Go e frontend em React JavaScript.
- Primeira entrada em produção com os módulos confirmados no material. Balanço Patrimonial está excluído.
- Bling como fonte principal; três estruturas específicas do Google Drive, somente de leitura.
- O fechamento funcional vigente está registrado na v2. Regras antigas substituídas nas v0/v1 são histórico.

## Evolução registrada

O backlog contém 127 tarefas, de BK-000 a BK-126. Os números **2,4% por quantidade** e **0% de desenvolvimento** pertencem à revisão de 13/09/2026 e são históricos. A matriz atual tem 51 requisitos: 1 verificado, 4 documentados, 26 parciais, 11 bloqueados e 9 não iniciados. A contagem simples (1/51, cerca de 2%) não mede o avanço real do projeto; as 127 tarefas ainda precisam ser auditadas contra o código e as evidências.

Na planilha, uma tarefa pontua 100% somente quando está concluída e possui evidência de aceite. Os pesos iniciais são iguais a 1. O acompanhamento não é sincronizado automaticamente com commits, issues ou projetos do GitHub.

Já existem código e evidências de CI para partes da fundação. Isso não equivale à homologação com dados reais nem significa que o produto esteja pronto para uso. Consulte `REQUIREMENTS-TRACEABILITY.md` para o estado dos requisitos. Ainda não há prazo global confiável, pois falta revisar cada tarefa do backlog.
