# Leitura das subpastas e importação CSV do Google Drive — 18/10/2026

## Resultado desta etapa

- A raiz aprovada foi registrada como `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com as três estruturas já autorizadas. IDs privados ficam fora do Git.
- O inventário passa a percorrer subpastas descobertas dentro da raiz autorizada, como `03_nuvem_envio/Frete`. Cada pasta guarda seu próprio marcador e estado; a retomada continua no ponto salvo.
- A consulta de arquivos fica restrita ao pai direto já confirmado na árvore. O serviço registra a raiz associada e verifica essa relação antes de publicar metadados ou baixar um arquivo.
- A varredura tem limites explícitos por árvore: até 10.000 pastas, 10.000 páginas e profundidade 32. Se ultrapassar, falha com segurança e não publica uma lista incompleta.
- Páginas, estados de pasta e inventário permanecem separados até a leitura completa; somente então o retrato, cursor e sucesso são publicados juntos. O retrato anterior permanece disponível durante falhas.
- A migration `000009_google_drive_tree_checkpoints.up.sql` acrescenta o estado privado de percurso. A conta `majucau_runtime` recebeu somente as permissões necessárias para ler, criar, atualizar e limpar esses checkpoints temporários.
- A migration `000010_google_drive_folder_id_checks.up.sql` mantém os limites e caracteres aceitos para IDs de pasta em uma expressão compatível com PostgreSQL 18. A CI reproduziu e confirmou que a validação anterior recusava o primeiro checkpoint.
- O job diário já liga o staging privado verificado ao importador CSV de recebimentos futuros. Pastas e arquivos de outras naturezas são inventariados, mas continuam pendentes quando não há parser/contrato homologado; diretórios não são tratados como lançamentos.
- A data do arquivo é validada contra o horário diário aprovado. Arquivo antigo ou que chegou depois das 09:30 não é publicado como dado atual.

## Verificações — atualização de 18/10/2026

- `go test . ./cmd/... ./database/... ./internal/... -count=1` passou em Windows com PostgreSQL 18 descartável, migrations 000001–000010 aplicadas e permissões runtime verificadas. Inclui leitura aninhada, falha após página confirmada, retomada, preservação do retrato anterior e publicação completa.
- `go vet`, `go mod verify`, `go mod tidy -diff`, verificações de backlog, limites Go, checksums e varredura local de segredos passaram.
- A CI #39189971597 encontrou defeitos reais antes desta correção: expressão regular incompatível com PostgreSQL 18, uma consulta com nome de coluna ambíguo e validação dos marcadores que tratava páginas de subpastas como se fossem da raiz. Também foram corrigidas expectativas antigas do teste de retomada.
- A migration 000010 preserva as migrations já registradas e valida IDs com comprimento de 1 a 256 e caracteres permitidos. O resultado da nova CI após enviar essa correção ainda precisa ser conferido.

## Limites reais observados

A prévia autorizada do Drive mostrou a estrutura de pastas, tipos e cabeçalhos. O arquivo CSV visível estava desatualizado em relação à rotina diária; por isso o controle de frescor mantém a fonte pendente. Nenhum arquivo foi baixado para o projeto e nenhum valor financeiro foi importado. A identidade técnica, a credencial protegida DPAPI e os IDs locais ainda não estão configurados, então não houve sincronização pela API real.

XLS/XLSX, TSV/JSON, contratos de cada estrutura, medição de carga/limpeza em instalação Windows e homologação real continuam pendentes. Nenhuma tela aprovada ou regra financeira foi alterada.
