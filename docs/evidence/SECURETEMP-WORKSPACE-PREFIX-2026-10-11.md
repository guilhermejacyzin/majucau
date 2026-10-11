# Validação do prefixo de diretório temporário — 11/10/2026

## Risco observado

`securetemp.NewWorkspace` recebe um prefixo para criar um diretório temporário.
Antes desta correção, barras no prefixo podiam influenciar o caminho montado e
fazer o diretório sair da pasta pai informada. Os chamadores atuais usam
prefixos fixos; ainda assim, a API reutilizável deve recusar separadores de
caminho na própria fronteira.

## Correção

- `NewWorkspace` agora rejeita `/` e `\\` no prefixo antes de acessar o sistema
  de arquivos, retornando `ErrUnsafeWorkspacePrefix`.
- Testes cobrem prefixos relativos e absolutos com ambos os separadores e
  confirmam que um prefixo válido cria a pasta diretamente dentro da pasta pai.
- Nenhuma regra financeira, tela, formato persistido ou importação foi alterada.

## Verificações executadas

- `go test -count=1 ./internal/securetemp` — aprovado no Windows.
- `go vet ./internal/securetemp` — aprovado.
- `git diff --check` — aprovado.
- Os caches Go foram criados em uma pasta exclusiva sob `artifacts/cache/` e
  removidos por `scripts/SafeCleanup.psm1` ao final. Os diretórios de teste
  foram removidos pelo próprio `go test`.

## Limites

Este teste cobre a validação do prefixo e o local de criação. Não demonstra a
ACL sob a conta real do serviço, a segurança de junctions no diretório pai,
falha de disco ou instalação em VM limpa. BK-024 permanece em 35% até esses
controles operacionais e as demais lacunas do mapa de riscos serem concluídos.
