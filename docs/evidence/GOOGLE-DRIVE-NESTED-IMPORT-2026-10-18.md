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

## Conferência adicional — conta financeira e Excel

- A usuária confirmou `financeiro@majucau.com.br` como a conta de referência. O Chrome estava autenticado nessa conta e mostrou a raiz aprovada `MAJUCAU_INTELLIGENCE_INPUT`, com as três subpastas previstas; `03_nuvem_envio/Frete` também foi conferida. Os IDs observados não são registrados no Git.
- A listagem mostrou uma planilha `.xlsx` em `01_perdas_operacionais`, um CSV em `02_nuvem_pago/Recebimentos_futuros` e um CSV em `03_nuvem_envio/Frete`. Todos exibiam última modificação em 11/09/2026, então permanecem desatualizados para a execução diária de 18/10. Foram consultados apenas nomes, tipos e datas; nenhum conteúdo foi baixado ou lido e nenhum dado financeiro foi importado.
- O importador existente de recebimentos futuros agora valida `.xlsx` pelo parser em fluxo: confere metadados e hash no staging privado, limita o arquivo a 256 MiB, lê uma linha por vez sem calcular fórmulas, remove o temporário e mantém a fonte pendente. XLS corrompido ou acima do limite também fica pendente sem chegar ao importador financeiro. Isso ainda não integra os arquivos das outras duas estruturas nem aprova contratos financeiros.
- Testes específicos do worker e do parser XLSX, `go vet` nesses pacotes, geração sqlc, checksums das migrations, limites arquiteturais e `git diff --check` passaram localmente. A CI #39191128295 (#674) falhou somente porque faltava versionar o modelo sqlc da migration 000009; `database/gen/models.go` foi regenerado. A nova CI após este envio ainda precisa passar.
- A identidade técnica de serviço, o segredo protegido por DPAPI e a configuração local dos IDs continuam ausentes. A sessão `financeiro@majucau.com.br` não é reutilizada pelo worker. Sem esses itens não há sincronização real pelo aplicativo.

## Revisão de continuidade — 19/10/2026

- A conta de referência continua confirmada como financeiro@majucau.com.br; a usuária confirmou que é essa mesma conta. Isso não substitui nem compartilha a credencial técnica do worker.
- A validação segura de XLSX foi ligada ao caminho existente de recebimentos futuros. Ela reconhece a estrutura do arquivo sem importar valores nem executar fórmulas, e mantém o arquivo pendente até aprovação do contrato financeiro.
- Os arquivos vistos em 18/10 tinham data de modificação 11/09/2026; permanecem antigos para a execução diária de 19/10. Nenhum conteúdo foi aberto, importado ou alterado no Drive.
- A sincronização real segue bloqueada por identidade técnica, credencial protegida DPAPI e configuração local dos IDs. Não há evidência de conexão ou importação real nesta etapa.
- Verificação ampla local em 19/10: scripts/verify.ps1 concluiu com sucesso, incluindo suíte Go, go vet, sqlc determinístico, migrations, scan de secrets, checks do instalador, lint, typecheck, 33 testes da interface, build e npm audit sem vulnerabilidades. govulncheck não achou chamadas a código vulnerável; o aviso de módulo GO-2026-5932 para openpgp não usado e sem correção permanece documentado em docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md. A CI após o novo commit ainda é necessária.
