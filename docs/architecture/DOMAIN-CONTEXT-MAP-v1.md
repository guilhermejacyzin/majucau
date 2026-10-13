# Mapa inicial de contextos do Majucau

**Data:** 12/10/2026
**Estado:** mapa técnico inicial para BK-016; pontos marcados como pendentes não estão aprovados como regra de negócio.

## Finalidade

Este documento organiza os sete contextos já listados no backlog e registra o que as decisões e o código atuais permitem afirmar. Ele não define fórmulas, lançamentos contábeis, transições financeiras, telas ou novas funcionalidades. As lacunas ficam explícitas para aprovação antes de virarem comportamento.

Fontes: BK-016 em `docs/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md`, `G0-DECISION-REGISTER.md`, `AGENTS.md` e `GO-MODULAR-BOUNDARIES.md`.

## Contextos e responsabilidades registradas

| Contexto | Responsabilidade registrada | Fronteira que deve ser preservada | Estado |
| --- | --- | --- | --- |
| **Tesouraria** | Reúne os conceitos de caixa, contas a receber, contas a pagar e disponibilidade financeira usados nos contratos aprovados. | Não confundir saldo/recebimento/pagamento com receita, despesa ou lucro. Não duplicar uma liquidação no previsto e no realizado. | Os conceitos estão documentados; propriedade exata de alguns fatos e transições ainda precisa de validação. |
| **Conciliação** | Mantém o estado de comparação entre os dados de origem e a posição financeira correspondente. | Uma divergência ou ausência de conciliação não pode virar saldo confirmado nem ser apagada por uma nova leitura. | A fronteira funcional está identificada; critérios completos por recurso ainda estão pendentes. |
| **Resultados** | Calcula a DRE própria do Majucau conforme as regras contábeis aprovadas; usa o relatório do Bling como fonte/comparação, não como resultado final. | Não copiar a DRE pronta do Bling, não reconhecer compra de estoque automaticamente como CPV/CMV e manter CAPEX fora do lucro. | A direção foi aprovada; mapeamento contábil por conta, competência e origem ainda está pendente. |
| **Operação** | Agrupa os temas operacionais de perdas, estoque, produção e compras presentes no escopo documentado. | Separar perda operacional de rendimento industrial; compras por quantidade de apropriação contábil de custo. | Contexto listado no backlog; propriedade detalhada dos fatos e relações entre eventos ainda não foi aprovada. |
| **Planejamento** | Consome contratos financeiros consolidados para projeções e calculadora de aplicação. | A calculadora não deve recalcular ou combinar por conta própria recebíveis, liquidações, aportes, resgates ou posição comprovada. | A fronteira de consumo está registrada; fontes, estados e regras detalhadas de cada componente permanecem vinculados às decisões aprovadas específicas. |
| **Integrações** | Conecta o Majucau ao Bling e às fontes de arquivo autorizadas, preservando origem, estado de sincronização e proveniência. | Adaptadores leem e traduzem dados; não inventam classificação financeira. RAW, fatos, cursor e conclusão seguem os limites transacionais de `AGENTS.md`. | A fronteira técnica está descrita; contrato real do Bling e acesso persistente ao Drive continuam pendentes. |
| **Acesso** | Trata identidade, perfis, capacidades e auditoria das ações. | A autorização acontece no worker; conhecer uma conta ou estar autenticado não concede capacidades adicionais. | `ADMIN` e `DIRECTOR` estão decididos para a versão 1; autenticação, sessão, MFA, provisionamento e aplicação das permissões ainda precisam ser concluídos. |

Os nomes e responsabilidades acima vêm do backlog, das decisões registradas e das fronteiras de implementação atuais. Onde a tabela diz “pendente”, isso não autoriza uma inferência de negócio.

## Fatos que precisam continuar distintos

O backlog manda separar estes conceitos antes de modelar seus relacionamentos:

- venda;
- direito a receber;
- liquidação/recebimento;
- aporte;
- resgate;
- rendimento realizado;
- posição de aplicação comprovada;
- perda operacional;
- rendimento ou perda industrial;
- compra por quantidade;
- apropriação contábil do custo;
- sinal de inteligência de cacau.

Esta lista não escolhe qual contexto é dono de cada evento, não define quando um evento é criado ou revertido e não determina como ele afeta uma demonstração. Atribuição final exige a política aprovada, os contratos das fontes e a validação da Majucau.

## Limites técnicos já registrados

- O produto continua um monólito Go com arquitetura hexagonal; a lista de contextos não cria serviços separados.
- `internal/domain` e `internal/financial` não dependem de HTTP, React, banco de dados ou adaptadores concretos, conforme `GO-MODULAR-BOUNDARIES.md`.
- O frontend usa IPC; não acessa PostgreSQL nem APIs externas diretamente.
- Bling e arquivos são fontes externas não confiáveis. Dados em crescimento são processados em fluxo, com limites, cancelamento e validação.
- A ingestão preserva a regra aprovada de atomicidade por recurso/lote, idempotência e checkpoint confirmado; uma falha não substitui último dado válido.
- A DRE é do contexto de Resultados, mas as linhas, a competência, a classificação e os critérios de estoque permanecem os aprovados em `G0-DECISION-REGISTER.md` e `docs/DRE-ACCOUNTING-POLICY-v1.md`.

## Itens necessários para fechar BK-016

1. Validar com a Majucau o contexto responsável por cada fato listado acima, sem duplicar proprietários.
2. Definir agregados, IDs compartilhados, eventos e invariantes a partir das regras já aprovadas; os itens ainda sem regra ficam pendentes.
3. Mapear dependências e transações entre os contextos sem introduzir ciclos nem alterar regras financeiras.
4. Confirmar como a inteligência de cacau permanece separada do escopo inicial do MVP.
5. Só então revisar os contratos de domínio e dependências dos módulos; qualquer impacto em tela ou regra exige aprovação própria.

## Escopo desta versão

Esta versão conclui somente o mapa inicial dos sete contextos, correspondente ao primeiro item de BK-016. Agregados, eventos, invariantes de negócio novos, atribuição final de cada fato, migrações e mudanças de código continuam fora desta versão.
