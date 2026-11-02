# Normalização do fluxo de Caixa/Bancos do Bling — 02/11/2026

## O que foi corrigido

- A carga diária tenta Contas a Receber, Contas a Pagar e Caixa/Bancos de forma independente. Uma falha em um título a pagar deixa de impedir a tentativa de `/caixas`.
- As páginas de `/caixas` continuam protegidas por paginação, checkpoint, hash, retry e armazenamento RAW. Depois de uma leitura completa, cada movimento também é validado e gravado na tabela normalizada `bling_cash_movements`, na mesma transação de publicação do lote e do cursor.
- A chave única `(connection_id, source_id)` torna a regravação idempotente. O detalhe RAW permanece ligado ao movimento para rastrear origem e versão.
- O normalizador guarda o valor como magnitude decimal exata e mantém a direção separada: `debCred = D` é saída; `debCred = C` é entrada. No resultado de fluxo do mês, entradas realizadas aparecem positivas e pagamentos/saídas realizados aparecem negativos, como Gisele esclareceu em 02/11.
- Os totais mensais usam apenas movimentos realizados (`situacao = R`) do lote diário atual. Uma carga sem movimentos pode confirmar zero somente depois de normalizar completamente `/caixas`. Se faltar a conta financeira de um movimento, os valores continuam rastreáveis e o indicador fica `PARTIAL`.
- A tela e seu desenho não foram modificados. O painel recebe os indicadores de realizado pelos campos já existentes do snapshot.

## Regra de sinal confirmada para Contas a Pagar

Em 02/11, Gisele esclareceu que, na apresentação financeira do Majucau, conta a pagar é saída e deve aparecer negativa. A causa da carga rejeitada era o normalizador de Contas a Pagar e a restrição do banco proibirem `saldo` negativo. A correção normaliza o saldo aberto como saída negativa, mesmo quando o Bling envia sua magnitude positiva, e mantém o valor negativo quando já chega assinado. O valor original do título continua guardado como magnitude positiva para rastreabilidade.

As consultas das telas passam a incluir obrigações com saldo negativo nos totais, vencimentos e linhas; antes, filtros `open_balance > 0` escondiam essas saídas. O ajuste não confunde esse valor com saldo bancário: saldo de conta continua sendo uma métrica separada e ainda depende de saldo de abertura e movimentos completos.

## Verificações executadas

- `go test ./internal/integrations/bling ./internal/dashboard` — passou em 02/11/2026.
- `scripts/verify-migrations.ps1` — passou; checksum da migration `000015_bling_cash_movements.up.sql` incluído.
- Testes novos cobrem débito/saída, crédito/entrada, valor decimal inválido, conta ausente, normalização negativa do saldo a pagar e sinais de saída nos totais e pagamentos.
- A suíte PostgreSQL local não foi executada: `MAJUCAU_TEST_DATABASE_URL` não está configurada nesta sessão. A CI PostgreSQL precisa executar o teste de integração SQL após o push.
- O código ainda não foi carregado no protótipo aberto, a migration ainda não foi aplicada à instância local e nenhuma sincronização real com esta versão foi feita. Portanto, não há comprovação de valores visíveis na tela neste momento.

## O que falta para confirmar T02/T01

1. Publicar a mudança e obter CI verde, incluindo a consulta real PostgreSQL com fixtures de entradas positivas e saídas negativas.
2. Atualizar/reiniciar o protótipo com a migration e o worker novos.
3. Rodar a carga autorizada do Bling e provar que a paginação de `/caixas` chegou ao fim na conta real.
4. Conferir os totais recebidos/pagos na T02 contra o Bling, incluindo estado `Parcial` quando aplicável.
5. Aplicar a nova migration que libera saldos assinados, atualizar o protótipo e validar contas a pagar contra Bling. Depois completar contas/saldos por Itaú e Mercado Pago, saldo inicial, projeção D0–D+60 e ligação de T01 aos resultados da T02. Esse commit não declara essas partes prontas.

Nenhuma tela aprovada foi redesenhada. As saídas ficam negativas nos valores existentes das telas, conforme a regra esclarecida pela responsável.
