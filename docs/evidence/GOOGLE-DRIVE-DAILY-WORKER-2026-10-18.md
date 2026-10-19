# Rotina diária de inventário do Google Drive — 18/10/2026

## O que foi ligado

- O worker ganhou uma rotina diária às 09:30, no fuso `America/Sao_Paulo`, para atualizar o inventário das três pastas autorizadas.
- Cada pasta usa um identificador estável por data e é sincronizada de forma independente. Uma falha em uma pasta não interrompe as outras; repetir a tentativa não duplica o lote já concluído.
- A migration `000007_google_drive_connection.up.sql` cria a conexão `GOOGLE_DRIVE` como `NOT_CONFIGURED`. Isso fornece uma identidade estável no banco sem declarar que a conta já foi conectada.
- Os IDs das pastas são lidos de variáveis de ambiente dedicadas. A chave da service account continua sendo lida somente do cofre DPAPI do worker. O worker não registra IDs, tokens, nomes de arquivos ou conteúdo.
- O agendamento só inicia quando os três IDs e a credencial protegida estão disponíveis. Configuração incompleta fica registrada por código técnico sanitizado.

Variáveis esperadas para configurar o worker:

- `MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID`
- `MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID`
- `MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID`

## Limites desta etapa

Esta rotina atualiza somente o inventário de metadados. Ela ainda não baixa os conteúdos para staging, não interpreta CSV/TSV/JSON/XLS/XLSX e não publica fatos financeiros. Ainda não validamos se os arquivos reais são encontrados. O tratamento de arquivo que chega depois das 09:30 também continua pendente no pipeline de importação.

Não foram usados dados financeiros reais, credenciais, IDs reais ou conexão com o Drive. Nenhuma tela aprovada ou regra financeira foi alterada.

## Verificações locais

- `go test ./... -count=1` — passou em Windows. O teste de integração PostgreSQL foi ignorado porque `MAJUCAU_TEST_DATABASE_URL` não está configurado; nenhuma instância desconhecida foi usada.
- `go vet ./...`, `go mod verify`, SQLC v1.31.1 e `git diff --check` — passaram.
- `scripts/verify-migrations.ps1`, verificação e testes do backlog e `scripts/verify-go-boundaries.ps1` — passaram.
- A cobertura adicionada verifica configuração parcial/ausente/completa, identificador diário estável, leitura independente por pasta e códigos de erro sem detalhes sensíveis.
- O checksum local da migration foi verificado, mas a migration ainda não foi aplicada localmente. A execução em PostgreSQL descartável e a CI completa aguardam o push.

## Próximo passo

Conectar o download em fluxo ao staging seguro do Majucau, preservando hash, limites de tamanho, verificação da versão do arquivo, limpeza garantida e publicação atômica. Depois, ligar cada formato ao contrato homologado. Excel permanece pendente até existir parser seguro e amostras autorizadas.
