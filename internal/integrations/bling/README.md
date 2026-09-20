# Importador Bling — Contas Recebidas

O importador usa o relatório do Bling como fonte de verdade para recebimentos realizados. A forma de pagamento pode conter `NUVEMPAGO 1X`, `NUVEMPAGO 2X`, `NUVEMPAGO 3X` ou `Nuvemshop PIX`, mas isso não muda `source_system=BLING`.

## Contrato

O parser espera exportação CSV separada por ponto e vírgula com os campos equivalentes a:

`Cliente`, `Histórico`, `Forma de pagamento`, `Nº documento`, `Vencimento`, `Liquidação`, `Situação`, `Valor taxa`, `Recebido`.

Somente `Situação=pago` vira `ReceiptCandidate` confirmado. Linhas abertas, canceladas ou sem documento ficam em `RowError`; não são convertidas em zero nem entram no total. Documentos duplicados no mesmo lote também ficam em erro para impedir dupla contagem.

O PDF `Bling - Relatório de Contas a Receber` fornecido em 20/09/2026 foi usado para confirmar visualmente os nomes e a semântica das colunas. A persistência em `raw_records`/`receipts` será feita pelo worker em uma etapa posterior, depois que o conector de arquivo e a pasta de entrada forem definidos.
