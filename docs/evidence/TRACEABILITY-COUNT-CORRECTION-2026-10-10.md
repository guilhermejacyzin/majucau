# Correção da contagem da matriz — 2026-10-10

## Contagem conferida

A recontagem considera somente linhas da matriz que começam com um identificador de requisito. O resultado é:

| Estado | Itens |
| --- | ---: |
| `VERIFIED` | 1 |
| `DOCUMENTED` | 5 |
| `PARTIAL` | 22 |
| `BLOCKED` | 11 |
| `NOT_STARTED` | 12 |
| **Total** | **51** |

- 1 de 51 requisitos está verificado: aproximadamente 2% verificado e 98% ainda sem verificação final.
- Esse percentual mede a quantidade de requisitos verificados, não o esforço, o tempo restante ou o percentual de código implementado.
- A mensagem do commit `a9ed574` informou incorretamente 2 de 56. Essa nota corrige o registro: a matriz contém 51 requisitos e 1 `VERIFIED`.

## Limites

- Esta correção é somente de governança e contagem; não muda a matriz, o código, as telas ou regras de negócio.
- A matriz continua sendo a fonte de estado. Cada requisito só passa a `VERIFIED` quando sua evidência de aceite correspondente estiver registrada.
