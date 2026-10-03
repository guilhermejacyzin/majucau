# Retry HTTP seguro — adapter Bling — 2026-10-03

## Regra aplicada

O adapter só repete o `GET` idempotente em falha transitória de transporte,
HTTP 408, 429 ou 5xx (500–599), com até três chamadas totais, backoff com jitter
e `Retry-After` limitado a 30 segundos. Erros permanentes, cancelamento,
resposta inválida e schema inesperado param sem nova chamada. O fluxo OAuth usa
`POST` de uso único e não herda retry.

## Mudanças e cobertura

- A classificação de erro de servidor agora limita-se a 500–599; status fora da
  faixa não são tratados como 5xx nem repetidos.
- Testes cobrem 408, 429, 500, 599 e 600; verificam que 400, 401, JSON inválido
  e schema inesperado geram somente uma chamada.
- O teste OAuth confirma que um erro 503 no pedido de refresh não repete o
  `POST` de uso único.

A validação da suíte e dos builds está pendente da CI deste commit.
