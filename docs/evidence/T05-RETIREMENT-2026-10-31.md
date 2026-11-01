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

- Revisão estática: procurar referências ativas à tela nos arquivos do frontend e conferir que não restem rota, navegação ou configuração T05.
- `git diff --check` deve passar antes do commit.
- A suíte da CI do GitHub valida a compilação e os testes existentes após o push. Nenhuma validação da conta real Bling ou de dados financeiros é implicada por remover a página.
- O processo `majucau-dev` e o servidor Vite do projeto já estavam em execução, portanto a edição da tela é servida pelo protótipo local. A captura da janela falhou duas vezes por timeout; não foi possível confirmar visualmente a atualização nesta execução.

## Próximo passo

Confirmar no protótipo iniciado novamente que “Conciliação” não aparece no menu e acompanhar a CI do commit. A operação financeira de correspondência permanece no escopo interno das telas que dependem dela.
