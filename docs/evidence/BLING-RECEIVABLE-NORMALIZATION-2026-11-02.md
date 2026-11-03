# Normalização de Contas a Receber do Bling — 02/11/2026

## Atualização posterior — importação da pasta 02

Em 02/11, a pasta `MAJUCAU_INTELLIGENCE_INPUT/02_nuvem_pago/recebimentos_futuros` foi retomada e processada antes da pasta 03. O arquivo visível no Drive chama-se `extrato_lancamentos_futuros_02_11_2026.csv`. A cópia local correspondente foi validada com 635 registros válidos (zero linhas com erro) e importada no banco isolado: 635 criados, zero atualizados, zero falhos e zero ignorados. Não gravamos dados de clientes nem valores financeiros nesta evidência.

Após aplicar a migration 000018 no PostgreSQL de desenvolvimento, o snapshot do worker passou a responder. O mesmo total **Parcial** foi conferido visualmente na T03 e na T01, sem alterar o layout aprovado. A parte Bling segue sem lote NORMALIZED para Contas a Receber, então a exibição parcial é esperada; não alegamos que o total combinado esteja completo.

**Limite de origem:** o Drive mostrou o arquivo e o aviso de upload concluído, mas bloqueou o download nesta sessão. Por isso não foi possível comparar o hash do objeto no Drive com a cópia local importada. A carga 635/0/0/0 comprova a importação da cópia local CSV, não uma leitura automatizada do Drive. A integração diária do Drive continua pendente de credencial técnica e configuração das pastas. Não processamos a pasta 03.

**Permissões:** a verificação real `database/privileges/verify_runtime.sql` passou no PostgreSQL local. Corrigi a allowlist para remover leitura ampla da tabela `receivable_sources`; o runtime conserva somente leitura das colunas de chave aprovadas. A execução do script ACL exige o administrador controlado devido às rotinas da extensão pgcrypto; isso foi documentado no README do banco.

**Verificações realizadas:** consulta do worker `snapshot_query_ok`; T01 e T03 abertas no protótipo conectado e conferidas; `go test ./internal/dashboard -count=1`; teste PostgreSQL focalizado; checksum da migration; verificação de privilégios; `git diff --check`.

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
- Gisele informou que atualizou o arquivo de recebimentos futuros na pasta autorizada do Drive. A verificação do worker em 02/11 confirmou que ele tentou recuperar a execução diária das 09:30, mas terminou antes de consultar o Drive: `GOOGLE_DRIVE_FOLDERS_NOT_CONFIGURED`, `folders_read=0` e `files_read=0`. Portanto, o arquivo atualizado e suas colunas não foram inspecionados nem importados.
- A causa foi confirmada no cofre local: o manifesto com os três IDs das pastas existe em `artifacts/retained/google-drive-folder-ids.json`, mas a integração exige que esses IDs sejam gravados junto com a chave de uma conta técnica Google em um único pacote protegido por DPAPI. O worker não lê esse manifesto desprotegido. O cofre atual não contém a credencial `google-drive-service-account`; não encontrei nem li a chave da conta técnica.
- O requisito de segurança aprovado é uma conta técnica separada da conta humana `financeiro@majucau.com.br`, com acesso Leitor apenas às três pastas autorizadas. Senha, cookie ou sessão do Chrome não substituem essa credencial e não devem ser copiados para o worker.
- Permanecem necessários: aplicar a migração ao PostgreSQL isolado do protótipo, aplicar e conferir os privilégios exatos, rodar o teste PostgreSQL específico, publicar o novo worker, sincronizar Bling e Nuvem Pago, reconciliar contagens e somas agregadas e verificar as telas.
- Nenhuma tela, regra contábil, cálculo de DRE ou classificação B2B/B2C foi alterada.

## Verificações a executar

1. `sqlc generate` e confirmar árvore gerada sem divergências.
2. `go test ./internal/integrations/bling ./internal/dashboard`.
3. Obter a chave JSON da conta técnica de leitura do Drive e garantir que o e-mail técnico tenha acesso Leitor somente às três pastas aprovadas. Não enviar a chave pelo chat nem armazená-la no Git.
4. Provisionar a chave e os IDs juntos pelo comando `majucau-worker provision-google-drive`/Named Pipe; confirmar o provisionamento sem imprimir a chave.
5. Aplicar todas as migrations no banco PostgreSQL descartável da CI e executar `privileges/verify_runtime.sql`; executar o teste de integração do normalizador em `majucau_test`.
6. Reiniciar o worker para recuperar a atualização diária das 09:30, importar o arquivo atualizado e reconciliar registros/somas agregadas nas telas T03 e T01.

Até essas verificações e a conferência visual ocorrerem, o projeto continua **não homologado** e o percentual geral não aumenta por esta implementação isolada.

## Rechecagem do Drive — 02/11/2026

O log `artifacts/retained/postgresql-dev/worker.stderr.log` mostra que o worker ativo iniciou o agendamento diário e tentou recuperar a execução das 09:30. A tentativa registrou `status=FAILED`, `folders_read=0`, `files_read=0` e `GOOGLE_DRIVE_FOLDERS_NOT_CONFIGURED`. Isso prova que a atualização ainda não chegou ao Majucau; não prova ausência do arquivo no Drive. Não houve acesso ao conteúdo do Drive nesta verificação.

## Diagnóstico da tela T03 no banco local — 02/11/2026

Uma consulta somente de leitura ao PostgreSQL isolado do protótipo encontrou 15 cargas bem-sucedidas de `contas.receber`, todas em `RAW_ONLY`; o total de fatos em `receivables` é zero e a tabela `receivable_sync_snapshot_rows` da migration 000018 ainda não existe. Isso explica por que T03 não apresenta o recebível Bling.

Também encontrei que `dailyBlingSyncCompleted` exigia `NORMALIZED` para Contas a Pagar e Caixa, mas não para Contas a Receber. Assim, a execução de 02/11 podia ser marcada como concluída mesmo com Contas a Receber somente RAW, e o worker seguinte a pulava. Corrigi a condição para exigir `NORMALIZED` nas três fontes e acrescentei regressão no teste PostgreSQL. Incluí o SHA-256 da migration 000018 no manifesto.

**Verificação local:** `go test ./internal/integrations/bling ./internal/dashboard ./cmd/worker -count=1` passou; `scripts/verify-migrations.ps1` passou. Os testes PostgreSQL reais ainda não foram executados localmente porque não há `MAJUCAU_TEST_DATABASE_URL`. A tentativa de preparar uma base descartável para executá-los foi bloqueada pela revisão automática; nenhum banco foi criado ou alterado. A CI remota ainda precisa confirmar a regressão.

**Ainda pendente:** atualizar o worker conectado, aplicar a migration 000018 e fazer a recuperação idempotente da sincronização Bling. O resultado Bling poderá aparecer como **Parcial** enquanto a credencial de leitura do Drive não estiver provisionada e o arquivo Nuvem não tiver sido importado.
