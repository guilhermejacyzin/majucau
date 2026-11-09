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

1. Gisele conferir a T02 que está aberta no protótipo local e aprovar ou
   apontar correções visuais.
2. Rodar uma nova sincronização real do Bling e comparar
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

## Estado que aparece na tela aberta

O protótipo local está na T02. O último snapshot exibido indica atualização em
09/11/2026 às 00:11. Os cards mostram entradas de R$ 219.267,59 em 30 dias e
R$ 260.805,50 em 60 dias; saídas de -R$ 313.408,28 e -R$ 512.222,94,
respectivamente. O saldo atual mostrado é R$ 62.548,33. A tabela e o gráfico
trazem as mesmas semanas e os mesmos valores de receitas, despesas e saldo
acumulado. Esses valores vêm do último snapshot local; não comprovam uma nova
consulta ao Bling nesta conferência.

T02 aguarda a aprovação visual e uma nova comparação real com o Bling; portanto
não está homologada e T03 não começou. Os percentuais existentes na meta são
estimativas anteriores e não foram recalculados nesta atualização.
