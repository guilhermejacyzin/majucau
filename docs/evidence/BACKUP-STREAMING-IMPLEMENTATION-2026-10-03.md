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

## Validação pendente

Esta evidência descreve o estado local antes da CI: o ambiente atual não possui
Go/gofmt, portanto a mudança ainda não foi compilada aqui. O status permanece
`PARTIAL` até a CI Windows/PostgreSQL e a validação em VM Windows 10/11. A suíte
existente exercita criação, verificação e restauração do formato gravado; ainda
falta uma prova automatizada dedicada de compatibilidade V1 e autenticação de
blocos CNG.

## Referências técnicas

- Microsoft CNG documenta a cadeia de chamadas BCrypt para autenticação de
  mensagens grandes e a validação da tag na chamada final:
  <https://learn.microsoft.com/pt-br/windows/win32/api/bcrypt/ns-bcrypt-bcrypt_authenticated_cipher_mode_info>.
- Tink Go documenta Streaming AEAD, o template AES-256-GCM-HKDF de 1 MiB e
  leitores/escritores em fluxo: <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/streamingaead>.
- O keyset Tink é gravado/lido com AEAD e dados associados:
  <https://pkg.go.dev/github.com/tink-crypto/tink-go/v2/keyset>.
