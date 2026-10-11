# Evidência: falhas ao criar diretório temporário seguro

**Data:** 2026-10-11
**Área:** `internal/securetemp`, `internal/integrations/csvstream`
**Escopo:** cobertura automatizada de falhas no provisionamento do diretório temporário no Windows.

## Mudanças

- O teste `TestNewDirPreservesCreateAndDescriptorReleaseFailures` injeta uma falha na criação do diretório e outra na liberação do buffer que contém os atributos de segurança.
- O teste confirma que o erro primário e o erro de liberação são retornados juntos e que não se tenta remover o diretório, pois a criação falhou.
- O teste de IDs em disco do `csvstream` passou a obter o caminho pelo `Workspace`, conforme a interface segura atual, em vez de acessar um campo antigo removido.

## Validação executada

```text
go test -count=1 ./internal/securetemp ./internal/integrations/csvstream
go vet ./internal/securetemp ./internal/integrations/csvstream
```

Ambos passaram localmente no Windows em 2026-10-11. Não foi feita validação em VM limpa nem teste de ACL do instalador.

## Limites e rastreabilidade

Esta entrega adiciona cobertura de teste e atualiza um teste existente; não muda comportamento contábil, financeiro, telas, importação ou formato persistido. `DATA-04` continua parcial. BK-024 permanece em 35% até validações de ACL/limpeza em VM e outros cenários de falha. O acesso ao Bling continua pendente da liberação da preferência salva no Chrome.
