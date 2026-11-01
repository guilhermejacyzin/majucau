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
- Em 31/10, foi confirmado que a janela aberta havia iniciado às 23:05, antes do commit de retirada de T05 às 23:27. A execução antiga foi fechada normalmente e o protótipo foi reaberto às 23:51 por `scripts/start-prototype.ps1`; Wails e Vite compilaram e iniciaram sem erro.
- A regressão automatizada agora verifica que não existe botão de navegação “Conciliação”. `npm run test -- --reporter=dot` passou com 41 testes em 4 arquivos. A tela independente também não existe no tipo de telas nem na configuração React.
- A CI [41267096669](https://github.com/guilhermejacyzin/majucau/actions/runs/41267096669), no commit `1e1f1a1`, terminou com sucesso nos quatro jobs: mensagem do commit, backlog, Windows e PostgreSQL.
- A captura visual da janela depois da reabertura ainda falhou por timeout. Assim, a remoção está confirmada no código, nos testes e na nova execução, mas a inspeção por imagem não pôde ser concluída nesta sessão.

## Próximo passo

O protótipo local já está aberto com a versão atual. Se a janela ainda mostrar uma página independente “Conciliação”, informar isso para investigarmos qual executável está sendo aberto. A operação financeira de correspondência permanece no escopo interno das telas ativas que dependem dela.

## Conferência da janela atual — 01/11/2026

- Consultei a janela ativa do executável `majucau-dev.exe` iniciado a partir deste checkout. A navegação contém Visão Executiva, Fluxo de Caixa, Contas a Receber, Contas a Pagar, DRE, Balancete, Calculadora de Aplicação, Estoque, Produção, Compras, Lançamentos & Pendências, Alertas, Integrações, Parâmetros e Usuários. Não existe item nem página independente chamada “Conciliação”.
- Na primeira conferência ainda havia o alerta “Conciliação pendente aguardando fonte oficial” na Visão Executiva. Ele não abria a tela antiga, mas mantinha uma referência visual à T05.
- A árvore acessível da janela foi lida com sucesso. A captura de imagem continuou expirando por timeout, portanto não afirmo validação visual por screenshot.

## Remoção do alerta residual — 01/11/2026

- Gisele pediu nova atualização do protótipo porque ainda via a tela de Conciliação. A conferência da janela mostrou que não havia página nem opção de navegação T05; a referência restante era o alerta antigo na T01.
- Removi da Visão Executiva somente o alerta `Conciliação pendente aguardando fonte oficial`, que remetia à tela retirada. Os três outros alertas, os cards principais, os indicadores e as regras internas permanecem iguais.
- O card `CONCILIAÇÃO` de Contas a Receber foi preservado porque pertence à tela ativa aprovada e mostra um controle operacional distinto da página T05.
- A regressão de interface verifica que nem a navegação oferece T05 nem a Visão Executiva exibe o alerta legado. Após o hot reload, a árvore acessível da janela aberta confirmou que o alerta sumiu e os três outros alertas continuam visíveis; esta conferência usou a árvore acessível, sem captura visual.
- Verificações da correção: `npm run test -- --reporter=dot` passou com 46 testes em 5 arquivos; `npm run lint` e `npm run typecheck` também passaram. A CI do commit desta correção será registrada após o push.
