# Leitura paginada e publicação atômica da API Bling — 02/11/2026

## Problema observado

Às 20:39, o processo do protótipo que começou às 19:57 ainda mantinha uma
transação PostgreSQL aberta havia aproximadamente 42 minutos, aguardando a
resposta do cliente. A última tentativa publicada de Contas a Pagar continuava
com o código antigo `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`; nenhum novo lote de
Contas a Pagar estava visível no banco. Essa execução usa o worker que já estava
aberto e não recebeu o código desta correção.

O erro antigo se referia à validação do saldo em aberto do título. Não era um
saldo bancário negativo. A regra financeira confirmada é: Contas a Pagar é uma
saída e seu valor aparece negativo no fluxo; isso não torna negativo o saldo de
Caixa e Bancos.

## Correção preparada

- A transação inicial registra o lote e os checkpoints e termina antes das
  chamadas de rede ao Bling.
- As páginas recebidas são guardadas com limite e checkpoint por página. A
  leitura de cada página e seus detalhes ocorre sem manter aberta a transação de
  publicação financeira.
- A validação confere número da página, limite, indicador de próxima página e
  quantidade de registros antes de reutilizar os checkpoints.
- Depois que todas as páginas passam pela leitura e validação, uma única
  transação publica RAW, dados normalizados, snapshot, cursor e auditoria. Se a
  publicação falha, ela é desfeita e o retrato anterior continua válido.
- Após o sucesso, os checkpoints temporários da execução são apagados. A
  sincronização mantém o bloqueio por recurso para impedir que duas cargas
  concorrentes publiquem snapshots incompatíveis.
- Contas a Pagar continua negativa no fluxo. Nenhuma tela aprovada ou regra
  financeira foi alterada.

## Evidência de teste

- `go test -count=1 ./internal/integrations/bling ./internal/financial ./internal/dashboard ./cmd/worker ./internal/scheduler` passou no Windows.
- Com um PostgreSQL 18 descartável, separado do banco do protótipo, passaram os testes de integração de `internal/dashboard`, `internal/integrations/bling`, `internal/integrations/googledrive` e `internal/integrations/nuvempago`.
- O teste de Contas a Pagar verificou que não há transação de publicação aberta durante as requisições à API, que o saldo aberto de R$ 150,25 é guardado como saída negativa, que categoria sem permissão produz lote parcial, e que uma resposta inválida não substitui o retrato anterior nem avança o cursor.
- `go vet` dos pacotes afetados e `git diff --check` passaram.
- O cluster PostgreSQL descartável foi parado e removido de `artifacts/cache`.

## Limites e próximo passo

No momento deste registro, a mudança está apenas na árvore local. Ainda não foi
enviada ao GitHub, a CI desta revisão ainda não executou e o worker antigo
continua ativo com uma chamada pendente. Portanto, ainda não há confirmação de
que a sincronização real de Contas a Pagar concluiu nem de que seus valores
aparecem nas telas. Depois do push e da CI, será necessário aplicar a nova versão
ao protótipo, concluir com segurança a execução antiga e verificar um novo lote
real antes de afirmar que a tela está alimentada.
