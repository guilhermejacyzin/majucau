# Estado real do projeto — 08/11/2026

## T02 — divergência de receitas

**Problema relatado:** receita da semana 22/11–28/11 diferente da tela Contas a Receber.

**Referência conferida nas capturas de Gisele:** período 22/11/2026–28/11/2026; Contas a Receber, todas as contas financeiras, títulos em aberto: 146 títulos, R$ 45.482,22 bruto, R$ 1.691,56 de taxas e R$ 43.790,66 líquido. Contas a Pagar: 14 títulos, R$ 25.275,71.

**Causas identificadas no código:** (1) a T02 usava pareamento entre o retrato do Bling e o arquivo externo Nuvem Pago, que podia excluir títulos do relatório mesmo que ele já incluísse os recebimentos da forma Nuvem Pago; (2) a receita usava saldo em aberto/bruto sem descontar as taxas da forma de pagamento.

**Correção feita no checkout:** a série T02 agora inclui diretamente todos os títulos em aberto do snapshot de Contas a Receber do Bling, sem adicionar ou deduplicar a cópia externa; calcula o líquido com as taxas oficiais da forma de pagamento da mesma execução e não usa fallback bruto; mantém Contas a Pagar como origem das despesas. Todas as contas financeiras dos relatórios entram na projeção. O saldo continua separado e baseado no último saldo conciliado do Itaú.

**Validação local:** `go test ./... -count=1`, `go build ./cmd/...` e `git diff --check` passaram. Testes PostgreSQL de integração foram ignorados porque `MAJUCAU_TEST_DATABASE_URL` não foi configurada.

**Protótipo atualizado:** em 08/11 às 23:02, encerrei normalmente a execução anterior e reabri o app pelo roteiro oficial. O worker novo e a janela Majucau iniciaram; o servidor local da interface respondeu HTTP 200. A captura visual automática da janela falhou, então ainda não confirmei pela tela quais valores aparecem. Também não executei uma nova sincronização do Bling nesta etapa.

**Ainda não comprovado:** não afirmar que R$ 43.790,66 já aparece na tela. A etapa só fecha depois de confirmar o snapshot sincronizado e comparar a mesma semana e os mesmos filtros. O ajuste depende das taxas da forma de pagamento presentes na mesma carga; se alguma não estiver configurada ou sincronizada, não deve exibir o valor bruto como se fosse líquido.

## Ação “Atualizar” — implementação no checkout

A auditoria inicial encontrou que o botão apenas relia os dados locais. Essa
falha foi corrigida no checkout atual: o clique agora chama uma operação
dedicada do worker para a carga oficial completa; concluindo, relê o snapshot e
atualiza a tela. Durante a operação, o botão mostra o andamento e bloqueia
cliques duplicados. Falha completa e execução incompleta têm avisos diferentes;
não se afirma que dados antigos foram preservados se algumas fontes tiverem
atualizado.

**Verificação local concluída:** `scripts/verify.ps1` passou, incluindo Go,
`go vet`, análise de segurança, integridade de migrations, sqlc determinístico,
lint, TypeScript, 75 testes de interface, build de produção e `npm audit`.
Isso valida o código, não a execução real da API. A operação ainda precisa ser
executada com o Bling real e conferida contra Contas a Receber (líquido),
Contas a Pagar e saldo conciliado do Itaú. Portanto, a ação está implementada
e verificada localmente, mas a T02 não está homologada e não se afirma que os
valores reais já estejam atualizados na tela.
Evidência: [`T02-UPDATE-ACTION-IMPLEMENTATION-2026-11-08.md`](evidence/T02-UPDATE-ACTION-IMPLEMENTATION-2026-11-08.md).

## Sequência imediata

1. **Concluído:** reconstruir/reabrir o protótipo com o código atualizado. A janela Majucau e a página local estão abertas.
2. Executar Atualizar e confirmar que a nova sincronização termina sem erro. A tela observada ainda mostra a última sincronização de 08/11/2026 às 23:16; portanto, os valores exibidos são o retrato anterior e não comprovam que a nova carga ocorreu.
3. Conferir que a T02 mostra R$ 43.790,66 líquidos de entradas na semana 22/11–28/11 e R$ 25.275,71 de saídas, usando os mesmos filtros das telas Bling.
4. Se a soma não bater, decompor diferenças por quantidade de títulos, vencimento, status, taxa e valor líquido por registro; não ajustar por palpite.
5. Registrar resultado real, hash/ID de sincronização sem dados secretos, aguardar a aprovação da T02 e só então avançar para T03.
