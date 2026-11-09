# Estado real do projeto — 08/11/2026

## T02 — divergência de receitas

**Problema relatado:** receita da semana 22/11–28/11 diferente da tela Contas a Receber.

**Referência conferida nas capturas de Gisele:** período 22/11/2026–28/11/2026; Contas a Receber, todas as contas financeiras, títulos em aberto: 146 títulos, R$ 45.482,22 bruto, R$ 1.691,56 de taxas e R$ 43.790,66 líquido. Contas a Pagar: 14 títulos, R$ 25.275,71.

**Causas identificadas no código:** (1) a T02 usava pareamento entre o retrato do Bling e o arquivo externo Nuvem Pago, que podia excluir títulos do relatório mesmo que ele já incluísse os recebimentos da forma Nuvem Pago; (2) a receita usava saldo em aberto/bruto sem descontar as taxas da forma de pagamento.

**Correção feita no checkout:** a série T02 agora inclui diretamente todos os títulos em aberto do snapshot de Contas a Receber do Bling, sem adicionar ou deduplicar a cópia externa; calcula o líquido com as taxas oficiais da forma de pagamento da mesma execução e não usa fallback bruto; mantém Contas a Pagar como origem das despesas. Todas as contas financeiras dos relatórios entram na projeção. O saldo continua separado e baseado no último saldo conciliado do Itaú.

**Validação local:** `go test ./... -count=1`, `go build ./cmd/...` e `git diff --check` passaram. Testes PostgreSQL de integração foram ignorados porque `MAJUCAU_TEST_DATABASE_URL` não foi configurada.

**Ainda não comprovado:** a instância real do protótipo e sua carga do Bling ainda não foram conferidas com a correção. Não afirmar que R$ 43.790,66 já aparece na tela. A etapa só fecha depois de reconstruir/reiniciar a instância com o código novo, atualizar o snapshot e comparar a mesma semana e os mesmos filtros. O ajuste depende das taxas da forma de pagamento presentes na mesma carga; se alguma não estiver configurada ou sincronizada, não deve exibir o valor bruto como se fosse líquido.

## Sequência imediata

1. Atualizar o worker e a interface ativos com o código corrigido.
2. Recarregar a fonte Bling e validar que os 146 títulos e taxas da janela estão no snapshot normalizado.
3. Conferir que a T02 mostra R$ 43.790,66 de entradas na semana 22/11–28/11, R$ 25.275,71 de saídas e os fechamentos semanais derivados do saldo inicial aprovado.
4. Se a soma não bater, decompor diferenças por quantidade de títulos, vencimento, status, taxa e valor líquido por registro; não ajustar por palpite.
5. Registrar resultado real, hash/ID de sincronização sem dados secretos, encerrar a conferência T02 e só então avançar para T03.
