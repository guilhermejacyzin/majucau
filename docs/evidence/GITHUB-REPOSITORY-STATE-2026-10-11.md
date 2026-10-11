# Revalidação do estado do repositório — BK-002

Consulta pública somente de leitura em 11/10/2026. Nenhuma branch, issue ou pull request foi alterada.

- Repositório: [guilhermejacyzin/majucau](https://github.com/guilhermejacyzin/majucau), público; branch padrão `main`.
- `main`: `39f8630390936717b2500c8b7ce923dc138cca72`.
- Branch de trabalho `codex/majucau-artifact-consolidation`: `1d90a0a86b7e6162c2348c865a0e2f91c39d2de5`, confirmada tanto no checkout local quanto no remoto.
- A API de comparação informa a branch de trabalho 206 commits à frente e 0 atrás de `main`; `main` é ancestral da branch de trabalho.
- A API retornou cinco pull requests abertos e nenhuma issue aberta. Os PRs #1–#5 permanecem em outras branches e nenhum foi mesclado nesta auditoria.
- O dado anterior de 178 commits à frente descrevia a fotografia de 10/10/2026; foi substituído nesta revalidação pela contagem atual de 206.
- A responsável confirmou em 10/10/2026 que não há outro clone, servidor ou ambiente de trabalho com código/dados além deste checkout; o controle do código fica no Git/GitHub.

Esta revalidação foi apenas documental. Nenhuma branch foi mesclada, nenhuma tela ou regra de negócio foi alterada e nenhum teste foi executado.

## Proteção da branch — consulta adicional em 11/10/2026

- A consulta de regras do repositório (`/repos/guilhermejacyzin/majucau/rulesets`) retornou uma lista vazia. Não há rulesets configurados que apareçam nessa consulta.
- A consulta de proteção clássica da branch padrão (`/branches/main/protection`) retornou HTTP 403: `Resource not accessible by integration`. Essa conexão não tem a permissão administrativa necessária para confirmar se `main` exige revisão ou verificações antes de aceitar mudanças.
- O erro 403 não significa que a branch esteja desprotegida. Nenhuma configuração foi alterada. Para fechar essa verificação, uma pessoa administradora precisa conferir **GitHub → repositório Majucau → Settings → Branches**.

### Confirmação pública do estado da `main` — 11/10/2026

- A API pública de metadados da branch (`GET /repos/guilhermejacyzin/majucau/branches/main`) retornou `protected: false` para `main`, no commit `39f8630390936717b2500c8b7ce923dc138cca72`. Portanto, a branch principal está sem proteção ativa neste momento; a consulta antiga com HTTP 403 não permitia concluir isso, mas esta consulta pública permite.
- A CI contém os jobs `verify-backlog`, `verify-windows` e `verify-postgres`; isso identifica verificações disponíveis, mas não as torna obrigatórias no GitHub. `verify-backlog` executa em todo push/PR, inclusive mudanças só documentais. A integração continua sem permissão para configurar proteção.
- Nenhuma regra de branch foi alterada. Para concluir BK-025, uma pessoa administradora precisa configurar proteção da `main` com revisão e verificações apropriadas em **GitHub → repositório Majucau → Settings → Branches**. Como essa configuração pode bloquear pushes/merges diretos, ela não foi aplicada automaticamente.
