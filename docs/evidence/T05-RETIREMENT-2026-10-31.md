# Retirada da tela T05 — Conciliação

**Data:** 31/10/2026
**Solicitação:** Gisele pediu para excluir a tela independente de Conciliação do projeto porque não vê necessidade de uma página própria. Depois reforçou que o protótipo ainda a mostrava e pediu a atualização do protótipo, do código e da documentação, com publicação no Git.

## Decisão e limite

- T05 deixa de ser uma tela ativa da versão 1.
- Não há página substituta, novos cards ou redistribuição de informações para outras telas.
- Cards e desenho das telas ativas permanecem como estavam aprovados.
- Regras internas de correspondência, estados, prevenção de duplicidade e rastreabilidade permanecem protegidas. Elas ainda sustentam controles financeiros e indicadores de telas ativas; removê-las poderia alterar os resultados financeiros.
- T01 agrega somente resultados das telas de origem ativas. Não ganha consulta direta ao Bling por causa desta retirada.
- A imagem T5 original continua preservada com seu hash como histórico; ela não é mais referência de uma página a implementar.

## Alterações aplicadas

- Removido o tipo de tela `conciliation`, seu painel/configuração e seu item da navegação do protótipo React.
- Removido o item correspondente da navegação legada no `App.tsx` e seu tooltip de navegação.
- Atualizados `AGENTS.md`, `G0-DECISION-REGISTER.md`, `IMPLEMENTATION-PLAN.md`, `REQUIREMENTS-TRACEABILITY.md`, o inventário de imagens de origem, as evidências de auditoria visual e o backlog v2.
- BK-055 passa a registrar a retirada do escopo. Tarefas que dependiam da página deixam de depender dela; a implementação das regras internas segue pelas tarefas próprias.

## Validação e publicação

- A revisão do frontend confirmou que o tipo de tela, o painel/configuração e os itens de navegação T05 foram retirados. Permanecem apenas textos de conciliação em telas ativas que descrevem controles internos já aprovados.
- `git diff --check` passou antes da publicação.
- Commits publicados na branch `codex/majucau-artifact-consolidation`: `86cd8bd` (retirada da tela) e ajustes pequenos nos verificadores e no resumo ativo do backlog (`421bb4a`, `6068c95`, `9c7c3d0`).
- A CI do GitHub [41265605349](https://github.com/guilhermejacyzin/majucau/actions/runs/41265605349), no commit `9c7c3d0`, passou nos quatro jobs: detecção/mensagem de commit, backlog, Windows e PostgreSQL. Os testes/builds automatizados do Windows e a verificação de PostgreSQL passaram. Isso não equivale à homologação de dados reais do Bling.
- O protótipo local e o servidor Vite já estavam em execução. A captura visual da janela falhou duas vezes por timeout; por isso, não consegui confirmar a tela diretamente no computador, embora a versão tenha passado pela compilação automatizada de Windows.

## Próximo passo

Na próxima conferência visual, confirmar que “Conciliação” não aparece no menu. Se a janela ainda mostrar o estado anterior, reiniciar o protótipo local para carregar a versão atualizada. A operação financeira de correspondência permanece no escopo interno das telas que dependem dela.
