# Engenharia e entrega de software

## Sumario

- Descoberta e criterios de sucesso
- Arquitetura e tecnologia
- Implementacao e mudancas
- Debugging e incidentes
- Dados e inteligencia artificial
- Sistemas distribuidos e concorrencia
- Performance, confiabilidade e operacao
- Software house e handoff

## Descoberta e criterios de sucesso

Identificar problema de negocio, usuario afetado, fluxo atual, resultado desejado, escopo, prazo, dependencias, restricoes legais ou contratuais e requisitos nao funcionais. Converter termos vagos em criterios observaveis sem inventar requisitos.

## Arquitetura e tecnologia

Comparar a menor arquitetura que atende corretamente com alternativas relevantes. Avaliar aderencia funcional, seguranca, disponibilidade, latencia, throughput, consistencia, observabilidade, recuperacao, manutencao, maturidade, interoperabilidade, lock-in, custo total, equipe e evolucao.

Registrar decisoes materiais, premissas e condicoes que fariam a recomendacao mudar. Nao escolher linguagem, cloud, banco, framework ou paradigma por preferencia ou hype.

## Implementacao e mudancas

Inspecionar convencoes e testes existentes. Aplicar a menor alteracao suficiente para a causa ou requisito. Evitar refatoracao nao relacionada, dependencia desnecessaria, reescrita gratuita e quebra de contrato sem decisao explicita.

Planejar rollout, compatibilidade, migracao e rollback quando a mudanca atingir dados, contratos ou producao. Manter logs uteis sem secrets nem dados pessoais desnecessarios.

## Debugging e incidentes

Usar o ciclo:

1. capturar evidencia;
2. definir esperado e observado;
3. reproduzir quando seguro;
4. formular hipotese falsificavel;
5. testar a hipotese;
6. localizar causa raiz;
7. alterar a menor superficie;
8. retestar e verificar regressao.

Em incidente, priorizar seguranca humana e de dados, contencao, restauracao do servico, preservacao de evidencia, comunicacao e rollback. Separar mitigacao imediata de correcao definitiva e post-mortem.

## Dados e inteligencia artificial

Para dados, verificar fonte de verdade, schema, contratos, qualidade, lineage, temporalidade, duplicacao, idempotencia, ordering, particionamento, retencao, auditoria, reprocessamento e recuperacao.

Para IA/ML, considerar dados, split, baseline, treino, validacao, metricas, error analysis, inferencia, deployment, monitoramento, drift, retreinamento e governanca. Verificar leakage, vies de amostragem, desbalanceamento, overfitting, calibracao, threshold, features espurias, robustez e reprodutibilidade. Uma unica metrica alta nao demonstra utilidade ou seguranca.

## Sistemas distribuidos e concorrencia

Considerar partial failure, timeout, retry, idempotencia, duplicacao, ordering, eventual consistency, backpressure, transacao, particionamento, split brain e cascading failure. Nao presumir rede confiavel.

Para concorrencia, investigar estado mutavel compartilhado, atomicidade, visibilidade, sincronizacao, locks, contention, race, deadlock, livelock, starvation, cancelamento e ownership de recursos. Nao presumir thread safety sem evidencia.

## Performance, confiabilidade e operacao

Medir antes de otimizar. Observar latencia por percentil, throughput, CPU, memoria, I/O, rede, storage, alocacao, GC, cache, locks e paralelismo conforme a stack.

Definir SLI/SLO quando pertinente. Avaliar retries, timeouts, circuit breakers, backpressure, capacidade, SPOFs, observabilidade, alertas, rollback, recovery e disaster recovery. Testar failure modes de maior impacto.

## Software house e handoff

Manter separacao estrita de clientes, repositorios, ambientes, acessos e dados. Considerar contrato, SLA, prazo, capacidade da equipe, suporte futuro e custo de manutencao.

Entregar contexto suficiente para continuidade: decisao, escopo, arquivos alterados, comandos de validacao, configuracao, operacao, migracao, rollback, riscos conhecidos e trabalho futuro. Evitar documentacao ornamental; preservar o conhecimento que reduz risco de manutencao.
