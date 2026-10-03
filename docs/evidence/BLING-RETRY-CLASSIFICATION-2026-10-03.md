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

## Validação CI em 2026-10-03

O commit `abaef3f611443b7d7253476f4d0132cc12640ff2` passou no run
[37130725887](https://github.com/guilhermejacyzin/majucau/actions/runs/37130725887),
nos jobs `verify-windows` e `verify-postgres`. No Windows passaram os testes Go,
`go vet`, govulncheck, checksums de migrations, geração SQL, secret scan, lint,
typecheck, testes e build do frontend, auditoria de dependências, builds do
desktop/worker/helper, smoke do helper, contrato NSIS, build do instalador,
smoke de instalação/desinstalação e consolidação dos outputs. O job PostgreSQL
também passou.

O artifact G1 unsigned `11276817282` tem 37.432.967 bytes, SHA-256
`8a573399e16b1063174dd6be1358787fa51c0f1b7e663b0d47e5f5a1de66597b` e expira
em 2026-10-04 14:53:01 UTC. Ele não é uma release assinada. A conexão GitHub não
oferece exclusão remota; o artifact foi registrado em
`artifacts/retained/cleanup-register.md`.
