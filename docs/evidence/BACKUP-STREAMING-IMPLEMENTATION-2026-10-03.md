# Backups em fluxo — decisão e implementação em revisão — 2026-10-03

## Decisão aprovada

Gisele aprovou manter a leitura/restauração dos pacotes V1 existentes e gravar
novos pacotes no formato V2 com blocos autenticados. Nenhuma tela aprovada ou
regra financeira/de negócio foi alterada.

## Escopo implementado localmente

- Nova geração V2 usa Tink Streaming AEAD AES-256-GCM-HKDF com segmentos de
  1 MiB; o keyset aleatório do pacote é cifrado sob chave derivada por Argon2id.
- Verificação calcula SHA-256 durante a leitura e autentica o fluxo completo
  antes de declarar o pacote válido.
- V1 preserva o envelope e o AAD anteriores e é descriptografado em blocos por
  Windows CNG; a compatibilidade mantém o limite inerente ao AES-GCM legado.
- A restauração copia o ZIP cifrado para uma área temporária privada, valida a
  cópia e só então prepara os dumps descriptografados em disco. Nenhum comando
  PostgreSQL consome payload antes da autenticação final.
- As áreas temporárias Windows são criadas com DACL protegida para o
  proprietário, LocalSystem e administradores. O cancelamento é propagado pela
  cópia, criptografia, verificação e preparação dos arquivos.

## Validação CI em 2026-10-03

No run [37100892358](https://github.com/guilhermejacyzin/majucau/actions/runs/37100892358),
passaram `go mod tidy -diff`, os testes Go (incluindo `internal/backup`), `go vet`,
govulncheck, secret scan, frontend lint/typecheck/tests/build, auditoria de
dependências, os builds desktop/worker/helper, smoke do helper e a integração
PostgreSQL. O job Windows falhou depois, ao preparar o WebView2 Bootstrapper:
o hash recebido (`AA38A8CFCE6179B87181609B1C730A29EAF26138FC833AF5759E67576770F3A3`)
não coincide com o hash fixado. A execução rejeitou o arquivo e o smoke NSIS de
instalar/desinstalar foi pulado.

O script agora exige assinatura Authenticode válida, signatário Microsoft e
EKU de assinatura de código para aceitar atualizações legítimas do bootstrapper;
essa mudança ainda aguarda CI. O status segue `PARTIAL` até a nova CI e a validação
em VM Windows 10/11. Também falta uma prova automatizada dedicada de restauração
V1 e autenticação de blocos CNG. O ambiente local não possui Go/gofmt.

## Referências técnicas

- Microsoft CNG documenta a cadeia de chamadas BCrypt para autenticação de
  mensagens grandes e a validação da tag na chamada final:
  <https://learn.microsoft.com/pt-br/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_authenticated_cipher_mode_info>.
- Tink Go documenta Streaming AEAD, o template AES-256-GCM-HKDF de 1 MiB e
  leitores/escritores em fluxo: <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/streamingaead>.
- O keyset Tink é gravado/lido com AEAD e dados associados:
  <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/keyset>.
