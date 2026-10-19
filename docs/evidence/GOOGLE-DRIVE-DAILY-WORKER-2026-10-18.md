# Rotina diária de inventário do Google Drive — 18/10/2026

## O que foi ligado

- O worker ganhou uma rotina diária às 09:30, no fuso `America/Sao_Paulo`, para atualizar o inventário das três pastas autorizadas.
- Cada pasta usa um identificador estável por data e é sincronizada de forma independente. Uma falha em uma pasta não interrompe as outras; repetir a tentativa não duplica o lote já concluído.
- A migration `000007_google_drive_connection.up.sql` cria a conexão `GOOGLE_DRIVE` como `NOT_CONFIGURED`. Isso fornece uma identidade estável no banco sem declarar que a conta já foi conectada.
- Na implementação inicial, os IDs das pastas eram lidos de variáveis de ambiente dedicadas, enquanto a chave da service account ficava no cofre DPAPI. Esse caminho alternativo foi removido em 19/10/2026; a correção está registrada em `docs/evidence/GOOGLE-DRIVE-PROTECTED-FOLDER-CONFIG-2026-10-19.md`.
- O agendamento diário permanece ativo antes do provisionamento. Cada execução relê a configuração protegida e não chama a API do Drive até encontrar os três IDs e a credencial no mesmo pacote DPAPI. Configuração ausente fica pendente; erros usam códigos sanitizados.

Variáveis usadas pela implementação inicial e agora desativadas:

- `MAJUCAU_GOOGLE_DRIVE_FOLDER_01_PERDAS_OPERACIONAIS_ID`
- `MAJUCAU_GOOGLE_DRIVE_FOLDER_02_NUVEM_PAGO_RECEBIMENTOS_FUTUROS_ID`
- `MAJUCAU_GOOGLE_DRIVE_FOLDER_03_NUVEM_ENVIO_ID`

Definir essas variáveis não configura nem habilita a leitura do Drive na versão atual.

## Limites desta etapa

Esta rotina atualiza somente o inventário de metadados. O adapter agora também oferece um componente de download para estágio privado, mas a rotina diária ainda não o chama. CSV/TSV/JSON/XLS/XLSX não são interpretados e nenhum fato financeiro é publicado. Ainda não validamos se os arquivos reais são encontrados. O tratamento de arquivo que chega depois das 09:30 também continua pendente no pipeline de importação.

Não foram usados dados financeiros reais, credenciais, IDs reais ou conexão com o Drive. Nenhuma tela aprovada ou regra financeira foi alterada.

## Verificações locais

- `go test ./... -count=1` — passou em Windows. O teste de integração PostgreSQL foi ignorado porque `MAJUCAU_TEST_DATABASE_URL` não está configurado; nenhuma instância desconhecida foi usada.
- `go vet ./...`, `go mod verify`, SQLC v1.31.1 e `git diff --check` — passaram.
- `scripts/verify-migrations.ps1`, verificação e testes do backlog e `scripts/verify-go-boundaries.ps1` — passaram.
- A cobertura adicionada verifica configuração parcial/ausente/completa, identificador diário estável, leitura independente por pasta e códigos de erro sem detalhes sensíveis.
- O checksum local da migration foi verificado, mas a migration ainda não foi aplicada localmente. A execução em PostgreSQL descartável e a CI completa aguardam o push.

## Próximo passo

Ligar o estágio privado ao pipeline depois que os contratos das três pastas forem homologados. O componente valida hash, limites, versão do arquivo e limpeza; a integração ao worker, os parsers e a homologação continuam pendentes. Excel permanece pendente até existir parser seguro e amostras autorizadas.
