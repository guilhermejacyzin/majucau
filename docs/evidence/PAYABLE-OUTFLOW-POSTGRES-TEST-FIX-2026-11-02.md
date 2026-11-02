# Teste PostgreSQL do sinal de Contas a Pagar — 02/11/2026

## Resultado da verificação

A migration `000016_payables_signed_outflow.up.sql` foi aplicada com sucesso no
PostgreSQL 18 descartável da CI. A etapa de integração encontrou uma expectativa
de teste desatualizada: o fixture enviava `saldo: 150.2500` e o teste ainda
esperava armazenar `150.2500`, embora a regra aprovada seja apresentar Contas a
Pagar como saída negativa.

## Correção e reprodução local

- Atualizei `payables_api_sync_integration_test.go` para esperar `-150.2500`;
  o valor original do título continua `250.0000`.
- Reproduzi a sequência em um PostgreSQL 18 isolado, na porta local `55439`,
  separado do banco `majucau_dev` do protótipo.
- Após corrigir a expectativa, passaram os testes de integração PostgreSQL de
  `internal/dashboard`, `internal/integrations/bling`,
  `internal/integrations/googledrive` e `internal/integrations/nuvempago`.
- O diretório temporário do banco de teste fica em
  `artifacts/cache/pg18-integration-20261102-2212` e deve ser encerrado e
  removido após a validação.

## Estado e limites

A execução CI `41533971213` passou por aplicar o schema, permissões, health do
worker e restore de backup. A etapa de integração falhou pela expectativa antiga
acima. Depois da correção do teste, a CI `41535352924` passou em Windows e
PostgreSQL, incluindo aplicação das migrations, suíte de integração, build e
empacotamento. A execução está registrada em
https://github.com/guilhermejacyzin/majucau/actions/runs/41535352924.

Esse teste usa dados inventados e não comprova sincronização real do Bling nem
valores visíveis no protótipo. A janela do protótipo não pôde ser capturada pelo
controle do computador; o banco de teste usado nesta reprodução é descartável e
separado do protótipo. Na verificação posterior, o protótipo continuava usando
binários anteriores ao commit que contém a correção.

Não foram alteradas telas nem regras financeiras. A correção implementa a
regra já confirmada: obrigação a pagar é saída negativa no resultado financeiro;
isso é diferente do saldo de caixa/banco.
