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

O run [37101826025](https://github.com/guilhermejacyzin/majucau/actions/runs/37101826025)
passou nos dois jobs Windows e PostgreSQL. No Windows passaram `go mod tidy -diff`,
os testes Go (incluindo `internal/backup`), `go vet`, govulncheck, checksums de
migrations, secret scan, frontend lint/typecheck/tests/build, auditoria de
dependências, builds desktop/worker/helper, smoke do helper, contrato NSIS,
build do instalador e smoke de instalar/desinstalar. A integração PostgreSQL
também passou.

O bootstrapper atual tem SHA-256
`AA38A8CFCE6179B87181609B1C730A29EAF26138FC833AF5759E67576770F3A3`, diferente
do pin antigo. A CI aceitou a atualização somente depois de Authenticode `Valid`,
signatário `Microsoft Corporation` e EKU de assinatura de código; thumbprint
`4028CAD637509D4744B17EC5B42AED8D7A31E6AF`. O artefato G1 unsigned foi carregado
com 36,939,538 bytes e SHA-256
`b3fa9d268c9bc6352240cd59e3d123602725728766fd597bb2478652fb85ded6`; ele continua
sendo um artefato unsigned, não uma release assinada.

Foi adicionado ao job Windows um fixture que reproduz o ZIP/envelope V1 das
versões anteriores. O teste cobre restauração de dumps com vários blocos CNG e
confirma que uma tag adulterada, mesmo com o SHA-256 do manifesto atualizado, é
rejeitada antes de parar o worker. O primeiro run desse fixture,
[37103043862](https://github.com/guilhermejacyzin/majucau/actions/runs/37103043862),
detectou uma access violation na chamada encadeada a `BCryptDecrypt`; o job
PostgreSQL passou. A rotina agora mantém um buffer IV de bloco entre chamadas,
aplica AAD somente no primeiro bloco e mantém vivos os buffers enviados à API.
A correção aguarda a próxima execução da CI.
O status do trabalho de backup segue `PARTIAL` até a CI confirmar esse teste e
haver validação de restauração em VM Windows 10/11 limpa. O ambiente local não
possui Go/gofmt; a validação de compilação e testes veio da CI.

## Referências técnicas

- Microsoft CNG documenta a cadeia de chamadas BCrypt para autenticação de
  mensagens grandes e a validação da tag na chamada final:
  <https://learn.microsoft.com/pt-br/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_authenticated_cipher_mode_info>.
- Tink Go documenta Streaming AEAD, o template AES-256-GCM-HKDF de 1 MiB e
  leitores/escritores em fluxo: <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/streamingaead>.
- O keyset Tink é gravado/lido com AEAD e dados associados:
  <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/keyset>.
