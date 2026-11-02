# Contas a Pagar como saída negativa — 02/11/2026

## Regra confirmada

Gisele esclareceu que Conta a Pagar representa uma saída; nos resultados
financeiros do Majucau, portanto, o saldo em aberto deve aparecer negativo.
Essa decisão se aplica ao sinal dos valores nas telas existentes T04 e T01. O
desenho das telas não foi alterado.

## Causa encontrada

O normalizador rejeitava qualquer `saldo` negativo no detalhe de uma conta a
pagar. A tabela `payables` também impunha `open_balance >= 0`, e as consultas
das telas filtravam `open_balance > 0`. Com isso, uma obrigação negativa podia
reverter o lote ou ficar fora de totais, vencimentos e linhas.

## Correção implementada

- O `valor` original do título continua preservado como magnitude positiva.
- O saldo em aberto é normalizado como saída: valores positivos do Bling viram
  negativos; valores já negativos continuam negativos; zero permanece zero.
- A migration `000016_payables_signed_outflow.up.sql` remove a restrição não
  negativa e converte os saldos positivos já armazenados para negativos.
- Totais, vencimentos e linhas de T01/T04 agora incluem saldos diferentes de
  zero, inclusive saídas negativas.
- O fluxo realizado mensal continua usando `/caixas`: crédito é entrada
  positiva e débito é saída negativa. Não se confunde título a pagar com saldo
  bancário.

## Verificações

- `scripts/verify.ps1` passou em 02/11/2026: testes Go, `go vet`, `govulncheck`,
  checksum das migrations, geração `sqlc` determinística, scan de segredos,
  validação do instalador, lint, typecheck, 62 testes de frontend, build e
  `npm audit` sem vulnerabilidades.
- A primeira execução da verificação encontrou uma diferença legítima de
  geração `sqlc` causada pelos comentários de schema/migration; após registrar
  a saída gerada, a verificação completa passou.
- Os testes PostgreSQL não foram executados localmente porque
  `MAJUCAU_TEST_DATABASE_URL` não está configurada. A CI pós-push precisa cobrir
  a aplicação da migration e os testes de integração.
- A migration ainda não foi aplicada ao protótipo e os dados reais não foram
  sincronizados com esta versão. Portanto, a tela ainda não está comprovada
  com os valores reais do Bling.

## Próximo passo

Publicar a alteração, confirmar CI com PostgreSQL, aplicar as migrations 000015
e 000016 no ambiente isolado do protótipo, sincronizar os dados autorizados e
comparar os valores exibidos em Contas a Pagar e Visão Executiva com o Bling.
Depois disso, conferir visualmente as telas aprovadas sem redesenhá-las.
