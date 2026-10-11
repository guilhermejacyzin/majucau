# Compatibilidade com JWT do Bling — 11/10/2026

## Referência oficial

O guia [Migração para autenticação JWT](https://developer.bling.com.br/migracao-jwt) exige o cabeçalho HTTP `enable-jwt: 1` ao obter ou renovar tokens e em todas as chamadas autenticadas. O guia informa que os tokens podem ter aproximadamente 1.500–3.000 caracteres e precisam caber no armazenamento e nos cabeçalhos.

A documentação [Descontinuação de URL da API](https://developer.bling.com.br/migracao-url) define `https://api.bling.com.br/Api/v3` como host atual. O código já utilizava esse host.

## Correção

- OAuth Authorization Code e Refresh Token usam `enable-jwt: 1`.
- Chamadas autenticadas do adapter (listas financeiras e consulta de detalhes) usam `enable-jwt: 1` junto com `Authorization: Bearer ...`.
- Não foi introduzido limite curto para o token: o bundle protegido pelo worker aceita tokens maiores que o intervalo de 1.500–3.000 caracteres indicado pelo Bling.
- O retry de OAuth não foi alterado: a troca de código e a renovação continuam sem repetição automática, conforme o contrato de operação de uso único registrado em `AGENTS.md`.

## Validação

Em Windows, passaram `go test -count=1 ./internal/integrations/bling ./internal/securetemp`, `go vet ./internal/integrations/bling ./internal/securetemp`, `scripts/verify-backlog.ps1` e `git diff --check`. Os testes verificam o cabeçalho na troca do authorization code, na renovação, nas consultas autenticadas e a transmissão de JWT de 3.000 caracteres. Os caches Go foram isolados em `artifacts/cache/` e removidos após a validação.

A CI completa `38172628938` passou em `verify-backlog`, `verify-postgres` e `verify-windows`, incluindo testes Go, `go vet`, varreduras, frontend, instalador e consolidação de outputs. O instalador unsigned gerado (`11714099858`, SHA-256 `92a705e002ec3629a16f8d8f0ab90a9078dea910db7aa9ad3df332ef48066aa4`) não pôde ser apagado pela conexão disponível; está no registro `artifacts/retained/cleanup-register.md` e expira em 12/10/2026.

Esta evidência de código/testes não equivale a homologação com a conta real. OAuth real, permissões, escopos e contrato de paginação ainda precisam ser confirmados antes de declarar a integração Bling pronta.

Nenhuma tela aprovada ou regra de negócio/financeira foi alterada.
