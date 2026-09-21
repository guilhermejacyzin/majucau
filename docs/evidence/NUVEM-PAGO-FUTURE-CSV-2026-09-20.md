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
- métodos IPC `nuvem_pago.future.preview` e `nuvem_pago.future.import`, métodos
  Wails correspondentes e painel sanitizado na tela de Integrações;
- testes unitários sem os arquivos reais e sem PII.

## Regra de origem

Este incremento não altera a regra validada: recebimentos realizados são
exclusivamente Bling. O CSV Nuvem Pago alimenta apenas o universo B2C futuro;
não grava `receipts`, não dá baixa e não recalcula valores com a tabela de
tarifas. Taxas e líquido efetivos do arquivo prevalecem quando houver fato
futuro exportado. O painel usa preview antes da gravação e o worker falha
fechado sem banco/conexão/configuração válida.

## Verificação executada

```text
go test ./internal/integrations/nuvempago     PASS
go vet ./internal/integrations/nuvempago      PASS
go test . ./cmd/... ./database/... ./internal/... PASS
go vet . ./cmd/... ./database/... ./internal/... PASS
frontend: typecheck, lint, 3 arquivos/19 testes e build Vite PASS
git diff --check                              PASS
E2E PostgreSQL opt-in                        IMPLEMENTADO, ainda sem ambiente executado neste checkout
```

## Limitação restante

O fluxo de preview/importação está ligado ao IPC, Wails e à tela, mas ainda
falta executar o teste opt-in contra PostgreSQL descartável, homologar fixture
sanitizado pela operação e validar o fluxo em instalador/VM limpa. O teste e o
fixture foram versionados sem PII; nenhum valor do arquivo real foi versionado.

