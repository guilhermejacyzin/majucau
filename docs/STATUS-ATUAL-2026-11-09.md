# Status atual — 09/11/2026

## Em que ponto está

O trabalho continua na T02 — Fluxo de Caixa. Corrigi uma conversão duplicada
de data que fazia o fluxo começar um dia antes. Isso podia puxar uma conta
vencida fora do período e deslocar os valores semanais. Os testes locais e a
verificação PostgreSQL passaram depois da correção.

O banco de desenvolvimento isolado recebeu a migration 000024 e as permissões
necessárias para guardar, por execução, quais movimentos o Bling devolveu
como conciliados. A conferência das permissões passou.

## O que ainda falta antes de aprovar a T02

1. Conferir visualmente a T02 atualizada. A aplicação nativa foi reconstruída
   e reaberta com o código novo, mas a captura para confirmar a tela expirou;
   não afirmo que a T02 foi visualmente revisada.
2. Rodar uma sincronização real do Bling com a nova estrutura e comparar
   Contas a Receber (líquido), Contas a Pagar e o saldo conciliado do Itaú.
3. Confirmar a atualização diária do saldo a partir do último saldo-base
   aprovado, sem misturar Itaú com Mercado Pago.
4. Só após esses pontos e a aprovação visual da responsável, iniciar T03.

## Saldo-base

Último valor aprovado: **R$ 62.548,33 em 06/11/2026**, copiado do saldo
conciliado do Itaú mostrado por Gisele. Esse é o ponto inicial autorizado. O
saldo só avança quando a API retorna movimentos conciliados do mesmo Itaú em
um lote normalizado; o fluxo projetado usa o saldo-base mais entradas líquidas
de Contas a Receber e saídas de Contas a Pagar.

## Verificações executadas

- `go test ./internal/dashboard -count=1` — passou.
- Teste de integração `TestReadStreamSummaryPostgresRequiresNormalizedDailyStages`
  no PostgreSQL 18 isolado — passou.
- Aplicação da migration 000024, scripts de permissão e verificação de
  privilégios no banco isolado — passou.

## O que ainda não se pode afirmar

A tela nativa foi reconstruída e reaberta com o código novo, mas não consegui
confirmar visualmente a T02 nesta sessão; a sincronização real também não foi
confirmada. Portanto, T02 não está homologada e T03 não começou. Os
percentuais existentes no documento de meta são estimativas antigas; não foram
recalculados nesta atualização.
