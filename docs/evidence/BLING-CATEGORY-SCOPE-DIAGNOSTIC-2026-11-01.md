# Diagnóstico específico da permissão de categorias — 2026-11-01

## O que foi corrigido

Quando a consulta de detalhe de uma categoria financeira do Bling recebe HTTP
403, o Majucau agora identifica o erro como
`BLING_CATEGORY_SCOPE_INSUFFICIENT`. A mensagem orienta a configurar a
permissão somente de leitura **Categorias — Receitas e Despesas** e refazer a
autorização.

O erro continua sanitizado: o conteúdo enviado pelo Bling não é mostrado nem
registrado. Os demais erros de escopo insuficiente continuam usando o código
genérico `BLING_API_SCOPE_INSUFFICIENT`.

## Verificação

- `go test ./internal/integrations/bling ./cmd/worker` passou localmente em
  01/11/2026.
- O teste do adapter simula uma resposta HTTP 403, verifica a classificação
  específica e genérica e confirma que a resposta do provedor não vaza.
- O teste do worker confirma o código e a mensagem que serão apresentados ao
  Majucau.
- `git diff --check` passou.

## O que esta correção não comprova

Não muda o cadastro nem as permissões da conta Bling, não conclui OAuth e não
prova que uma sincronização real está conectada. O estado real já registrado
continua descrito em `BLING-AUTHORIZATION-STATE-2026-11-01.md` e
`BLING-LIVE-SYNC-SCOPE-2026-11-01.md`. A consulta real só poderá ser repetida
depois que o escopo de leitura for salvo no cadastro Bling e a autorização for
concluída no Majucau.

Nenhuma tela, regra financeira, dado, credencial ou permissão externa foi
alterada.
