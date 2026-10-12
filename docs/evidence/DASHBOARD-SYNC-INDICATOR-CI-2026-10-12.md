# Validação dos indicadores de sincronização do painel — 12/10/2026

## Resultado

- A atualização aprovada mantém o mesmo desenho e os mesmos rótulos. O cabeçalho mostra o horário do último lote Bling concluído (`SUCCESS` ou `PARTIAL`), e mostra `—` quando não existe lote concluído.
- Recebíveis sem registros confirmados são apresentados como indisponíveis (`—`), sem sugerir um saldo confirmado de R$ 0,00.
- A regra financeira não foi alterada. Nenhuma informação financeira real do Bling foi consultada nesta validação.

## Validação automática

- Commit: `e5ce5bbe83bd0424ec463f004ed19ba169e084a7`.
- Execução da CI: [38315792124](https://github.com/guilhermejacyzin/majucau/actions/runs/38315792124), concluída com sucesso em 12/10/2026.
- Os quatro grupos passaram: `detect-changes`, `verify-backlog`, `verify-postgres` e `verify-windows`.
- A verificação Windows incluiu testes de Go e frontend, build do aplicativo e geração do instalador de teste. A verificação PostgreSQL também passou.
- O commit atualizou o simulador já existente no `internal/dashboard/stream_test.go` para aceitar a data opcional de sincronização e verificar que ela chega ao resumo. Não foram criados testes locais nesta execução; a CI foi executada automaticamente após o push.

## Limites e pendências

- A validação não comprova conexão autenticada nem importação de dados reais. O registro de 12/10 mostra o Bling como desconectado e sem `Client ID` ou `Redirect URI` configurados.
- A janela local do protótipo permaneceu ativa. Ela não foi reiniciada, então esta evidência não afirma que o executável aberto já contém a alteração visual; confirma a implementação versionada e seu build na CI.
- A navegação direta ao Bling continuou bloqueada pela política salva da ferramenta do navegador. Não foi usado caminho alternativo.
- O backlog segue com 127 tarefas: 7 concluídas, 13 em andamento e 107 não iniciadas; média registrada de 9,1% concluída e 90,9% restante. BK-102 permanece em 1%; esta validação automática não altera seu percentual.

## Artefato temporário

A CI gerou o instalador não assinado `majucau-g1-unsigned` (56.238.548 bytes; SHA-256 `03a206cf8db9a3a4fc450ddda19307f0c953e563a08f6542e82de810ab7b2524`). O GitHub o mantém até `2026-10-13T22:23:20Z`. Como a conexão não oferece exclusão de artefatos, seu caminho foi anotado no registro local ignorado `artifacts/retained/cleanup-register.md` para remoção manual ou expiração automática.
