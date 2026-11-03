# Normalização de Contas a Receber do Bling — 02/11/2026

## O que foi preparado

- O lote diário completo de `/contas/receber` agora tem uma etapa própria para converter linhas RAW em `receivables`, contatos e vínculos de origem.
- A rotina lê detalhes por GET fora da transação de publicação. Os checkpoints existentes mantêm as páginas retomáveis; a publicação final grava fatos, snapshot, cursor e auditoria numa transação só. Falha antes do commit mantém o snapshot anterior.
- Para títulos em aberto ou com saldo potencial, a rotina busca o detalhe e valida que ID, situação e contato correspondem à linha da lista. Valores são convertidos em decimais exatos e datas validadas.
- O tipo de pessoa do contato não é usado para decidir B2B/B2C. Os títulos Bling são persistidos como `UNCLASSIFIED`.
- A leitura do painel inclui apenas títulos que pertencem ao último snapshot diário NORMALIZED de Contas a Receber.
- A migração 000018 cria a associação imutável de títulos ao lote e a função validada para cadastrar/atualizar o nome do contato, sem conceder escrita direta do worker na tabela de contatos.
- A regra aprovada para o total combinado compara nome normalizado do cliente e valor bruto exato; a data não participa. Correspondências entre fontes são pareadas uma a uma, sem apagar repetições dentro da mesma fonte. Quando há correspondência, a linha Nuvem Pago representa valor líquido e data prevista.
- Se a importação diária da Nuvem Pago não tiver sido concluída, o total combinado permanece visível como **Parcial**.

## Limites e verificações

- Esta alteração no código **não comprova que os dados reais foram atualizados no protótipo**. A sincronização só pode ser considerada confirmada depois que o worker executar com a migration 000018, publicar um lote `NORMALIZED` e as telas de origem e T01 mostrarem o mesmo resultado.
- Gisele informou que atualizou o arquivo de recebimentos futuros na pasta autorizada do Drive. Ainda não foi possível ler essa aba nesta execução: o controle do navegador encerrou a tentativa por não conseguir confirmar com segurança a URL atual. O arquivo e suas colunas não foram inspecionados nem importados nesta etapa.
- Permanecem necessários: aplicar a migração ao PostgreSQL isolado do protótipo, aplicar e conferir os privilégios exatos, rodar o teste PostgreSQL específico, publicar o novo worker, sincronizar Bling e Nuvem Pago, reconciliar contagens e somas agregadas e verificar as telas.
- Nenhuma tela, regra contábil, cálculo de DRE ou classificação B2B/B2C foi alterada.

## Verificações a executar

1. `sqlc generate` e confirmar árvore gerada sem divergências.
2. `go test ./internal/integrations/bling ./internal/dashboard`.
3. Aplicar todas as migrations no banco PostgreSQL descartável da CI e executar `privileges/verify_runtime.sql`.
4. Executar teste de integração do normalizador de receber em `majucau_test`.
5. Só então atualizar o protótipo conectado e validar o resultado real.

Até essas verificações e a conferência visual ocorrerem, o projeto continua **não homologado** e o percentual geral não aumenta por esta implementação isolada.
