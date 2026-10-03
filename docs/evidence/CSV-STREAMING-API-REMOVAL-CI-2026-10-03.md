# Remoção de importadores CSV em memória — CI 2026-10-03

## Mudança

A auditoria encontrou helpers internos que acumulavam a importação inteira em
listas: `ParseReceiptsCSV`, `ImportReceiptsFolder`, `PersistReceipts`,
`ParseFutureCSV`, `ImportFutureFolder` e `PersistFutureReceivables`. A busca por
chamadores confirmou que esses caminhos eram usados somente pelos próprios
helpers e por testes; o worker já importava usando os fluxos streaming.

Removi os helpers de materialização total e redirecionei a cobertura existente
para `StreamReceiptsCSV`, `StreamFutureCSV`, `PreviewReceiptsFolder` e
`PreviewFutureFolder`. Os importadores de produção continuam enviando cada
registro ao callback de persistência dentro da transação. A deduplicação fica no
índice transacional HMAC; as prévias guardam no máximo 50 detalhes de arquivos e
50 problemas, mantendo as contagens completas. O resumo Nuvem Pago permanece
como contadores agregados, sem armazenar registros.

## Evidência

A [CI 37114031004](https://github.com/guilhermejacyzin/majucau/actions/runs/37114031004)
passou nos jobs `verify-postgres` e `verify-windows`, incluindo testes Go,
`go vet`, build do aplicativo/worker, verificação de saída e smoke de instalação
e desinstalação.

A primeira tentativa, [CI 37113884245](https://github.com/guilhermejacyzin/majucau/actions/runs/37113884245),
detectou que o serviço Nuvem Pago ainda usa `FuturePersistenceSummary`. O tipo
agregado foi restaurado no commit `c21bde6`; as APIs que carregavam arrays
completos continuaram removidas e a nova CI passou.

## Limites

Nenhuma tela ou regra de negócio foi alterada. Esta prova confirma os caminhos
CSV identificados e a CI; não encerra DATA-04. Ainda falta auditar todos os
outros fluxos de dados crescentes e validar ACL/limpeza e restauração em VMs
Windows reais.
