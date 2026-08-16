# Orquestracao multiagente

## Sumario

- Selecao de especialistas
- Roteamento de modelo e raciocinio
- Contrato de delegacao
- Paralelizacao e ownership
- Sintese e confronto

## Selecao de especialistas

Convocar somente papeis que mudem a qualidade da decisao:

- **Explorer:** mapear repositorio, entry points, dependencias e testes; manter read-only por padrao.
- **Software engineer:** implementar de ponta a ponta, preservar contratos e criar testes associados.
- **Debugger:** reproduzir, formular e falsificar hipoteses, localizar causa raiz, corrigir e retestar.
- **Architect:** avaliar requisitos funcionais e nao funcionais, integracoes, custo, evolucao e operacao.
- **QA:** procurar edge cases, regressao, falhas de contrato, concorrencia e comportamento nao deterministico.
- **Security:** modelar ameacas, trust boundaries, identidade, autorizacao, secrets, supply chain e abuso.
- **Data/AI:** avaliar fontes, qualidade, lineage, leakage, metricas, drift, reprocessamento e governanca.
- **Performance/SRE:** medir gargalos, capacidade, disponibilidade, recuperacao e failure modes.
- **Researcher:** buscar fontes primarias, versoes, especificacoes e evidencia externa atual.
- **Adversarial reviewer:** tentar derrubar a implementacao e as premissas.

Nao convocar todos automaticamente. Um agente pode cobrir mais de um papel quando o escopo for coerente.

## Roteamento de modelo e raciocinio

Quando a plataforma permitir escolher:

- Usar Luna low/medium para descoberta, busca, comandos, lint e testes simples.
- Usar Luna high para implementacao normal, debugging, integracao e revisao de codigo.
- Usar Luna xhigh para concorrencia, baixo nivel, algoritmos dificeis, distribuicao ou performance critica.
- Escalar para modelo intermediario ou para o agente principal quando tentativas delimitadas falharem, a tarefa exigir integracao sistemica continua ou o risco for elevado.

Nao prometer um modelo que a plataforma nao disponibilize. O prompt nao controla o modelo principal escolhido pelo host.

## Contrato de delegacao

Definir para cada agente:

- objetivo e limites;
- contexto minimo necessario;
- arquivos ou componentes sob ownership;
- se pode escrever ou deve permanecer read-only;
- entregavel esperado;
- criterios de aceitacao e validacao;
- restricoes de seguranca e dados.

Exigir retorno compacto com:

1. objetivo recebido;
2. resultado;
3. evidencias;
4. alteracoes realizadas;
5. validacao executada;
6. riscos e incertezas;
7. recomendacao.

## Paralelizacao e ownership

Paralelizar pesquisa, exploracao, threat modeling e preparacao de testes quando forem independentes. Serializar decisoes dependentes e alteracoes sobre os mesmos arquivos.

Evitar:

- dois agentes editando a mesma superficie;
- implementacao antes de contrato ou arquitetura necessarios;
- dependencias circulares entre tarefas;
- compartilhar dados de um cliente com contexto de outro;
- delegar uma decisao irreversivel sem revisao do agente principal.

## Sintese e confronto

Comparar resultados com o problema e com evidencias observaveis. Quando agentes divergirem:

1. localizar a premissa divergente;
2. verificar versao, ambiente, metodologia e escopo;
3. buscar teste ou fonte que possa falsificar uma das hipoteses;
4. explicitar incerteza residual;
5. tomar a decisao sistemica no agente principal, envolvendo o usuario quando material.
