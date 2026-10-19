# Leitura das subpastas e importação CSV do Google Drive — 18/10/2026

## Resultado desta etapa

- A raiz aprovada foi registrada como `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com as três estruturas já autorizadas. IDs privados ficam fora do Git.
- O inventário passa a percorrer subpastas descobertas dentro da raiz autorizada, como `03_nuvem_envio/Frete`. Cada pasta guarda seu próprio marcador e estado; a retomada continua no ponto salvo.
- A consulta de arquivos fica restrita ao pai direto já confirmado na árvore. O serviço registra a raiz associada e verifica essa relação antes de publicar metadados ou baixar um arquivo.
- A varredura tem limites explícitos por árvore: até 10.000 pastas, 10.000 páginas e profundidade 32. Se ultrapassar, falha com segurança e não publica uma lista incompleta.
- Páginas, estados de pasta e inventário permanecem separados até a leitura completa; somente então o retrato, cursor e sucesso são publicados juntos. O retrato anterior permanece disponível durante falhas.
- A migration `000009_google_drive_tree_checkpoints.up.sql` acrescenta o estado privado de percurso. A conta `majucau_runtime` recebeu somente as permissões necessárias para ler, criar, atualizar e limpar esses checkpoints temporários.
- O job diário já liga o staging privado verificado ao importador CSV de recebimentos futuros. Pastas e arquivos de outras naturezas são inventariados, mas continuam pendentes quando não há parser/contrato homologado; diretórios não são tratados como lançamentos.
- A data do arquivo é validada contra o horário diário aprovado. Arquivo antigo ou que chegou depois das 09:30 não é publicado como dado atual.

## Verificações

- `go test . ./cmd/... ./database/... ./internal/...` — passou localmente em Windows.
- `go vet . ./cmd/... ./database/... ./internal/...` — passou em Windows.
- O teste unitário confirma que somente descendentes encontrados por listagem do pai autorizado podem ser consultados e que a autorização antiga é removida antes de uma nova varredura.
- O teste PostgreSQL cobre leitura aninhada, falha após página confirmada, retomada, preservação do retrato anterior e publicação completa. Ele foi acrescentado à etapa `verify-postgres` da CI e ainda precisa passar no GitHub.
- O serviço Docker não estava disponível nesta sessão; não foi usado o PostgreSQL já instalado no computador. A aplicação da migration e a execução PostgreSQL ficam para o ambiente descartável da CI.

## Limites reais observados

A prévia autorizada do Drive mostrou a estrutura de pastas, tipos e cabeçalhos. O arquivo CSV visível estava desatualizado em relação à rotina diária; por isso o controle de frescor mantém a fonte pendente. Nenhum arquivo foi baixado para o projeto e nenhum valor financeiro foi importado. A identidade técnica, a credencial protegida DPAPI e os IDs locais ainda não estão configurados, então não houve sincronização pela API real.

XLS/XLSX, TSV/JSON, contratos de cada estrutura, medição de carga/limpeza em instalação Windows e homologação real continuam pendentes. Nenhuma tela aprovada ou regra financeira foi alterada.
