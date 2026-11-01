# Remoção de botões do protótipo — 01/11/2026

## Solicitação aprovada

- Remover `Exportar` de todas as telas e não disponibilizar arquivos nesta etapa.
- Remover por completo `Gerar pedidos` de Compras; a tela permanece somente de consulta.
- Preservar os cards, o desenho restante das telas, os dados e as regras financeiras aprovadas.
- Manter qualquer exportação futura adiada até depois da finalização atual.

## Alterações

- `frontend/src/mockupDashboard.tsx`: retirado o menu `Exportar` compartilhado pelos cabeçalhos do protótipo e o botão `Gerar pedidos` da tela Compras.
- `frontend/src/mockup.css`: removidos os estilos sem uso associados ao menu de exportação.
- `frontend/src/App.test.tsx`: criada regressão que percorre todas as telas navegáveis e confirma que não há botão `Exportar`, que Compras não mostra botão de geração de pedido e que os filtros permanecem desativados.
- O helper `internal/reportexport` e as permissões de banco existentes não têm rota de interface exposta e não foram apagados. Eles não disponibilizam exportações no protótipo; a função futura fica adiada conforme o registro de decisão.
- `G0-DECISION-REGISTER.md`, `AGENTS.md`, `IMPLEMENTATION-PLAN.md`, `REQUIREMENTS-TRACEABILITY.md`, backlog JSON/Markdown e registro da auditoria histórica foram alinhados à decisão vigente.
- A planilha oficial do backlog v2 foi recriada a partir do JSON oficial. Pré-visualizações temporárias foram revisadas e serão removidas de `artifacts/verification/`.

## Verificação

- `npm run test -- --reporter=dot` em `frontend`: 4 arquivos e 41 testes aprovados.
- `npm run lint` e `npm run typecheck` em `frontend`: aprovados.
- `git diff --check`: aprovado.
- `scripts/verify-backlog.ps1`: 126 tarefas ativas; 11,9% estimados concluídos e 88,1% restantes. A estimativa não foi alterada por esta remoção de botões.
- `python work/verify_package.py`, executado em `docs/planejamento`: 127 tarefas, 108 decisões, 90 links locais válidos e nenhuma fórmula com erro na planilha oficial.
- A planilha oficial v2 foi regenerada pelo builder existente e as linhas alteradas de BK-012, BK-086, BK-097, BK-104, BK-109 e as decisões de exportação foram revisadas em prévias temporárias.
- Busca no código de interface não encontrou botões/textos ativos `Exportar`, `Gerar pedidos` ou `Gerar compras`. Referências históricas e o helper interno de XLSX não são ações disponíveis em tela.
- Na conferência posterior, a árvore acessível confirmou que a navegação não inclui a página T05 — Conciliação. Gisele apontou que o alerta legado “Conciliação pendente aguardando fonte oficial” ainda dava a impressão de que a tela existia; esse alerta foi removido da T01. Os demais alertas/cards e os controles internos das telas ativas permanecem preservados.
- Após a correção residual da T05, o protótipo aberto recebeu a atualização por hot reload e a árvore acessível confirmou três alertas na T01, sem o alerta antigo. `npm run test -- --reporter=dot` passou com 46 testes em 5 arquivos; `npm run lint` e `npm run typecheck` passaram. A CI do commit `e2bdb90` [#41270518884](https://github.com/guilhermejacyzin/majucau/actions/runs/41270518884) passou em backlog, PostgreSQL e Windows. A conferência foi feita pela árvore acessível, sem captura visual.

## Limite

Esta validação cobre a interface local de protótipo e os testes do frontend. A exportação para Diretores fica adiada; esta mudança não comprova sincronização real com Bling/Drive nem homologação para produção.
