# Governança do projeto Majucau

**Objetivo:** concluir o backlog sem quebrar decisões aprovadas, manter evidência rastreável e deixar cada execução fácil de revisar.

## Fontes de verdade

- `IMPLEMENTATION-PLAN.md`: sequência e gates do plano-raiz.
- `REQUIREMENTS-TRACEABILITY.md`: estado de cada requisito e evidência que falta.
- `G0-DECISION-REGISTER.md`: decisões e regras aprovadas.
- `AGENTS.md`: padrões obrigatórios de implementação, dados, segurança e operação.
- Planejamentos v2 conflitantes são material de análise; não substituem o plano-raiz nem aprovam regras.

## Proteção de telas e regras

- Telas aprovadas e regras de negócio não podem ser alteradas sem aprovação expressa da usuária.
- Uma mudança de implementação deve preservar a apresentação aprovada e as regras existentes. Quando a fonte, o comportamento esperado ou a compatibilidade persistente não forem claros, interromper apenas a parte dependente e pedir esclarecimento.
- Decisões aprovadas devem apontar para a regra, a evidência e os testes de regressão correspondentes.

## Fluxo de mudança e evidência

1. Escolher um item do plano-raiz que não dependa de uma decisão em aberto.
2. Preservar a rastreabilidade entre requisito, código, teste e execução da CI.
3. Fazer commits pequenos no branch de trabalho e enviar ao remoto autorizado.
4. Só marcar `VERIFIED` quando a evidência requerida pelo requisito estiver registrada; uma CI verde não aprova automaticamente um módulo inteiro.
5. Após cada execução, resumir em linguagem simples o que foi feito, o que falta, percentual restante da tarefa, percentual restante do projeto e estimativa de conclusão. Se o prazo total depender de credenciais, fonte oficial, decisão funcional ou VM indisponível, registrar a dependência e não inventar uma data.

## Mensagem obrigatória de commit

Cada commit deve conter estes campos:

```text
O que é: tarefa que está subindo

O que foi feito: resumo técnico detalhado

Próximo passo: detalhe o próximo passo

Quantos % da meta falta para terminar a tarefa: N%

Quantos % da meta falta para terminar o projeto: N%
```

## Dados, resiliência e segurança

- Fluxos de dados crescentes devem ser processados em streaming com buffers e limites explícitos; envelopes pequenos de controle também têm limite de bytes.
- Persistência de RAW, fatos normalizados, cursor e conclusão de lote mantém atomicidade; replay é idempotente.
- Retry é padrão apenas para falhas transitórias e chamadas que podem ser repetidas com segurança. Operações sem garantia oficial de replay seguro não são repetidas por presunção.
- Entrada externa deve ser validada, limitada, cancelável e protegida contra exposição de PII/segredos. Falhas nunca viram valores financeiros zero.

## Caches, outputs e retenção

- Outputs e caches gerados pelo trabalho local ficam sob `artifacts/`.
- Ao terminar cada execução, apagar dados regeneráveis que não estejam em uso e cuja remoção não interrompa a entrega.
- Não apagar dados do usuário, artefatos de release ainda necessários nem diretórios usados por processo ativo. Mover retenções autorizadas para `artifacts/retained/` e registrar origem, motivo, tamanho, data e caminho de exclusão manual em `artifacts/retained/cleanup-register.md`.
- O registro local é operacional e fica fora do controle de versão para não publicar caminhos locais ou inventário de dados da máquina.
