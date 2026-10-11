# Limite de logs antigos na inicialização — 2026-10-11

## Risco encontrado

O logger já limitava cada gravação nova a 32 KiB, cada arquivo a 10 MiB e a
rotação a cinco arquivos. Porém, na inicialização ele aceitava um arquivo antigo
maior que 10 MiB. Uma rotação posterior poderia manter esse arquivo acima do
limite anunciado de 50 MiB por instalação.

## Correção

Antes de abrir o arquivo ativo, o logger inspeciona os cinco nomes de arquivo
que administra. Se um deles não for um arquivo regular ou exceder o limite
configurado, a inicialização do logger falha sem apagar, truncar ou regravar o
arquivo existente. No Windows, a inspeção também recusa links e reparse points.

## Validação local

- `go test -count=1 ./internal/diagnosticslog` passou no Windows em
  2026-10-11, incluindo casos de arquivo ativo e arquivo rotacionado acima do
  limite. Os testes confirmam que os bytes antigos permanecem intactos.
- `go vet ./internal/diagnosticslog` passou no Windows em 2026-10-11.
- `git diff --check` passou.
- O cache Go foi colocado em `artifacts/cache` e removido com o módulo de limpeza
  segura após a validação.

## Limites restantes

Esta alteração não valida a instalação do serviço em VM Windows 10/11 e não
mostra uma referência de diagnóstico na tela. BK-031 continua parcial. Nenhuma
tela aprovada ou regra financeira foi alterada.
