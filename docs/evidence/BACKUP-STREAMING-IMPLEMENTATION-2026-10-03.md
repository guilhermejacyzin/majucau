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
versões anteriores. O primeiro run desse fixture,
[37103043862](https://github.com/guilhermejacyzin/majucau/actions/runs/37103043862),
detectou uma access violation na chamada encadeada a `BCryptDecrypt`; o job
PostgreSQL passou. A correção passou na CI seguinte,
[37103530380](https://github.com/guilhermejacyzin/majucau/actions/runs/37103530380),
nos dois jobs Windows e PostgreSQL. O teste restaura dumps V1 com vários blocos
CNG e confirma que uma tag adulterada, mesmo com SHA-256 do manifesto atualizado,
é rejeitada antes de parar o worker. A chamada mantém um buffer IV mutável entre
blocos, aplica AAD na primeira chamada da cadeia e mantém vivos os buffers Go
usados pela API.

O run `37103530380` executou `go mod tidy -diff`, testes Go, `go vet`,
govulncheck, checksums de migrations, secret scan, frontend lint/typecheck/tests/
build, auditoria de dependências, builds desktop/worker/helper, smoke do helper,
contrato e build NSIS e smoke de instalação/desinstalação. O job PostgreSQL
também passou. O artefato unsigned G1 `11267780358` tem 36,939,552 bytes e SHA-256
`08521b50fdd2edf8964b35038868816bf1d05217f66ed00b696b43ebd3840aa1`; continua
sendo um artefato unsigned, não uma release assinada.

O status do trabalho de backup segue `PARTIAL` até haver validação de restauração
em VM Windows 10/11 limpa, restore real de PostgreSQL e integração com
worker/UI/instalador. O ambiente local não possui Go/gofmt; a validação de
compilação e testes veio da CI.

## Validação adicional V2 — CI em 2026-10-03

O run [37105079618](https://github.com/guilhermejacyzin/majucau/actions/runs/37105079618)
passou nos jobs `verify-windows` e `verify-postgres`. O teste
`TestV2StreamingBackupAuthenticatesLargeSegmentedPayload` gera um dump de 3 MiB,
confirma que a verificação aceita o pacote V2 com vários segmentos e depois
altera o texto cifrado, atualiza o SHA-256 no manifesto e confirma que a
autenticação rejeita o pacote. Isso prova que a validação não depende somente
do hash do manifesto.

No mesmo run passaram testes Go, `go vet`, govulncheck, checksums de migrations,
geração SQL, secret scan, lint/typecheck/testes/build frontend, auditoria de
dependências, builds de desktop/worker/helper, smoke do helper e NSIS, smoke de
instalação/desinstalação e consolidação de outputs. O artefato G1 unsigned
`11267901629` tem 36,939,597 bytes, SHA-256
`29c117d5a72a320ccce30651a340965801cc6aae719fa6814350cfe71364864d` e expira em
2027-01-01. Ele continua sendo um artefato unsigned, não uma release assinada.

Esta prova cobre autenticação e verificação V2 em payload maior que os segmentos;
não substitui restore real PostgreSQL, integração do worker/UI/instalador nem
ensaio em VM limpa. `OPS-02` e `DATA-04` continuam `PARTIAL`.

## Referências técnicas

- Microsoft CNG documenta a cadeia de chamadas BCrypt para autenticação de
  mensagens grandes e a validação da tag na chamada final:
  <https://learn.microsoft.com/pt-br/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_authenticated_cipher_mode_info>.
- Tink Go documenta Streaming AEAD, o template AES-256-GCM-HKDF de 1 MiB e
  leitores/escritores em fluxo: <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/streamingaead>.
- O keyset Tink é gravado/lido com AEAD e dados associados:
  <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/keyset>.
