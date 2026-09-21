# Nuvem Pago — extrato futuro CSV — evidência 2026-09-20

## Escopo fechado neste incremento

- parser semicolon-CSV para o contrato do arquivo de lançamentos futuros;
- origem fixa `NUVEM_PAGO`, entidade `nuvem_pago_recebimentos_futuros_csv_v1`;
- status normalizado `PROJECTED`;
- preservação de data de pagamento, data prevista de recebimento, método,
  bandeira, parcelas, bruto, taxa, juros, custos totais e líquido;
- rejeição de linhas que não sejam `Entrada`/`Venda`, datas inválidas,
  recebimento anterior ao pagamento e IDs repetidos;
- leitura da pasta controlada em ordem lexical, SHA-256 por arquivo e
  deduplicação entre arquivos;
- persistência transacional opcional no worker: lote, RAW versionado e upsert
  idempotente em `receivables` com `business_type=B2C` e `status=PROJECTED`;
- testes unitários sem os arquivos reais e sem PII.

## Regra de origem

Este incremento não altera a regra validada: recebimentos realizados são
exclusivamente Bling. O CSV Nuvem Pago alimenta apenas o universo B2C futuro;
não grava `receipts`, não dá baixa e não recalcula valores com a tabela de
tarifas. Taxas e líquido efetivos do arquivo prevalecem quando houver fato
futuro exportado; a persistência RAW ainda é o próximo passo.

## Verificação executada

```text
go test ./internal/integrations/nuvempago     PASS
go vet ./internal/integrations/nuvempago      PASS
go test . ./cmd/... ./database/... ./internal/... PASS
go vet . ./cmd/... ./database/... ./internal/... PASS
git diff --check                              PASS
```

## Limitação restante

O serviço de persistência está implementado, mas ainda falta executar um E2E
contra PostgreSQL descartável, ligar o serviço ao IPC e à tela de
Importações/Integrações e homologar um fixture sanitizado pela operação.
Nenhum valor do arquivo real foi versionado.
