# Retry-After sem overflow — CI 2026-10-03

## Mudança

O adapter Bling agora interpreta segundos de `Retry-After` em inteiro de 64 bits e aplica o teto aprovado de 30 segundos antes da multiplicação por `time.Second`. Valores decimais maiores que `int64` também saturam no teto, em vez de virar uma duração negativa por overflow. Datas HTTP continuam usando o parser padrão e o atraso final segue limitado a 30 segundos.

## Evidência

- `TestRetryAfterClampsHugeNumericValuesWithoutOverflow` cobre 30 segundos, o maior `int64` e um valor decimal acima do intervalo do tipo.
- A [CI 37119548600](https://github.com/guilhermejacyzin/majucau/actions/runs/37119548600), para o commit `8b27000`, passou nos jobs Windows e PostgreSQL. O pipeline inclui testes Go, `go vet`, análises de vulnerabilidades/segredos, frontend, builds do desktop/worker e smoke do instalador.

## Limites

O teste valida a saturação matemática e a política local de retry; não homologa rate limits, cursor ou janela incremental com credenciais reais do Bling. INT-04 permanece `PARTIAL` até a integração real e a confirmação oficial da paginação.
