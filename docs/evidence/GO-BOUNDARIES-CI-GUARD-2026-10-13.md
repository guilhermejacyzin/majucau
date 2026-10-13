# Proteção das fronteiras da arquitetura Go — 13/10/2026

## O que foi conferido

O repositório documenta uma arquitetura modular para o monólito Go. O domínio e os cálculos não devem depender de banco, rede, sistema operacional ou interface. A aplicação depende do domínio; os adaptadores e pontos de composição conectam as implementações concretas.

## Mudança

Foi criado `scripts/verify-go-boundaries.ps1`. Ele consulta o grafo de imports de todos os pacotes Go do módulo e compara os pacotes encontrados com as dependências aceitas:

- `internal/domain`: somente biblioteca padrão do Go, sem rede, banco, sistema operacional, caminhos de arquivo, chamadas de sistema, plugins ou `unsafe`.
- `internal/financial` e `internal/application`: biblioteca padrão e `internal/domain`, com as mesmas restrições de infraestrutura.
- Todos os pacotes: não podem importar código em `cmd/`; esses pacotes são pontos de composição, não bibliotecas internas.

A verificação foi ligada a `scripts/verify.ps1` e ao job Windows da CI. Assim, um novo import fora das regras faz a verificação falhar antes de ser integrado.

## Verificação desta execução

- `scripts/verify-go-boundaries.ps1`: passou localmente em 13/10/2026.
- A CI completa será executada após o push; este documento não afirma que ela já passou.
- O endereço de retorno OAuth informado pela responsável coincide com o endereço HTTPS fixo que o Majucau aceita. Isso não comprova uma autorização real nem uma sincronização.

## Limite e impacto

BK-017 passa de 90% para 95%. Os 5% restantes dependem de BK-016 e da validação final dos contratos entre módulos. A recontagem também corrigiu o resumo de BK-067, que já estava em 25%, mas ainda aparecia como 20% no panorama. A média atual do backlog de 127 tarefas é 11,0% concluído e 89,0% restante, arredondada a uma casa decimal.

Nenhuma tela aprovada, regra financeira, dado persistido ou credencial foi alterado. A conexão real do Bling, a paginação confirmada e a sincronização homologada continuam pendentes.
