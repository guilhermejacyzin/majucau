# Estado real do projeto — 08/11/2026

## T02 — divergência de receitas

**Problema relatado:** receita da semana 22/11–28/11 diferente da tela Contas a Receber.

**Referência conferida nas capturas de Gisele:** período 22/11/2026–28/11/2026; Contas a Receber, todas as contas financeiras, títulos em aberto: 146 títulos, R$ 45.482,22 bruto, R$ 1.691,56 de taxas e R$ 43.790,66 líquido. Contas a Pagar: 14 títulos, R$ 25.275,71.

**Causas identificadas no código:** (1) a T02 usava pareamento entre o retrato do Bling e o arquivo externo Nuvem Pago, que podia excluir títulos do relatório mesmo que ele já incluísse os recebimentos da forma Nuvem Pago; (2) a receita usava saldo em aberto/bruto sem descontar as taxas da forma de pagamento.

**Correção feita no checkout:** a série T02 agora inclui diretamente todos os títulos em aberto do snapshot de Contas a Receber do Bling, sem adicionar ou deduplicar a cópia externa; calcula o líquido com as taxas oficiais da forma de pagamento da mesma execução e não usa fallback bruto; mantém Contas a Pagar como origem das despesas. Todas as contas financeiras dos relatórios entram na projeção. O saldo continua separado e baseado no último saldo conciliado do Itaú.

**Validação local:** `go test ./... -count=1`, `go build ./cmd/...` e `git diff --check` passaram. Testes PostgreSQL de integração foram ignorados porque `MAJUCAU_TEST_DATABASE_URL` não foi configurada.

**Protótipo atualizado:** em 08/11 às 23:02, encerrei normalmente a execução anterior e reabri o app pelo roteiro oficial. O worker novo e a janela Majucau iniciaram; o servidor local da interface respondeu HTTP 200. A captura visual automática da janela falhou, então ainda não confirmei pela tela quais valores aparecem. Também não executei uma nova sincronização do Bling nesta etapa.

**Ainda não comprovado:** não afirmar que R$ 43.790,66 já aparece na tela. A etapa só fecha depois de confirmar o snapshot sincronizado e comparar a mesma semana e os mesmos filtros. O ajuste depende das taxas da forma de pagamento presentes na mesma carga; se alguma não estiver configurada ou sincronizada, não deve exibir o valor bruto como se fosse líquido.

## O que o botão “Atualizar” faz hoje

A responsável esclareceu que espera que a atualização busque os dados e altere os resultados mostrados. A auditoria do frontend constatou que o ícone “Atualizar dados da tela” apenas chama `getState()` e `getSnapshot()` e relê o que já está salvo localmente. Ele **não inicia uma nova busca no Bling**. Portanto, clicar nele pode redesenhar os mesmos números antigos e não comprova sincronização. A busca manual ainda é uma ação separada em Integrações; existe também a rotina automática diária das 09:30.

Essa diferença explica por que uma atualização visual não corrigiu os totais. A interface precisa ligar a ação Atualizar ao fluxo oficial de sincronização; quando o lote terminar, deve recarregar o snapshot e apresentar o horário e o resultado reais. Isso está documentado como trabalho pendente, não como código concluído. Não alterei a tela ou regras financeiras nesta atualização de documentação.

Evidência: [`T02-UPDATE-ACTION-AUDIT-2026-11-08.md`](evidence/T02-UPDATE-ACTION-AUDIT-2026-11-08.md).

## Sequência imediata

1. Atualizar o worker e a interface ativos com o código corrigido.
2. Recarregar a fonte Bling e validar que os 146 títulos e taxas da janela estão no snapshot normalizado.
3. Conferir que a T02 mostra R$ 43.790,66 de entradas na semana 22/11–28/11, R$ 25.275,71 de saídas e os fechamentos semanais derivados do saldo inicial aprovado.
4. Se a soma não bater, decompor diferenças por quantidade de títulos, vencimento, status, taxa e valor líquido por registro; não ajustar por palpite.
5. Registrar resultado real, hash/ID de sincronização sem dados secretos, encerrar a conferência T02 e só então avançar para T03.
