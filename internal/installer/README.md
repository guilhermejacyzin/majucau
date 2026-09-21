# Preflight do instalador

`internal/installer` implementa somente a avaliação pré-instalação. O
preflight não cria diretórios, não escreve registro, não abre serviços, não
instala WebView2/PostgreSQL e não altera configurações da máquina.

## Contrato da CLI

O executável `installer-helper` aceita:

```text
installer-helper preflight [--install-dir PATH] [--data-dir PATH]
                           [--free-space-path PATH]
                           [--min-free-bytes N] [--port N]
```

Os caminhos são usados somente para probes e nunca aparecem no JSON. O
resultado tem `schema_version`, `status`, `exit_code`, `checks` e `issues`.
Os códigos de processo são estáveis:

* `0` — `READY`: todos os gates passaram;
* `2` — `BLOCKED`: a máquina não atende um ou mais requisitos recuperáveis;
* `3` — `INTERNAL_ERROR`: argumento inválido ou probe não configurado/falhou.

Para exportar diagnóstico sanitizado, use `installer-helper diagnostics
--output C:\\Temp\\majucau-diagnostic.zip` com os mesmos parâmetros opcionais
do preflight. O comando pode retornar `BLOCKED` e ainda assim produzir o ZIP;
o bundle contém apenas `diagnostic.json` e `README.txt`, sem caminhos completos,
segredos, tokens, DSNs ou payloads.

O pacote também fornece `JournalStore` para o `Install\\state.json`, com
escrita temporária, `fsync`, rename atômico, fases versionadas e validação de
hash/estado. `AcquireInstallLock` usa mutex nomeado no Windows (`Global\\MajucauInstaller`)
para impedir duas instalações simultâneas; a UI do instalador ainda precisa
integrar essas primitivas ao fluxo NSIS antes do gate G7.

Os códigos de bloqueio estáveis incluem `OS_UNSUPPORTED`,
`INSTALL_NOT_ELEVATED`, `REBOOT_PENDING`, `DISK_SPACE_LOW`,
`PATH_UNAVAILABLE`, `PATH_UNSAFE_LOCATION`, `DB_PORT_CONFLICT` e
`WEBVIEW2_MISSING`.

## Manifesto de pacote

`ValidateReleaseManifest` faz a validação determinística, sem efeitos
colaterais, de um `release-manifest.json` dentro de um pacote. O contrato
versionado registra a versão do aplicativo, a janela compatível de schema,
os checksums das migrations e o inventário de artefatos. A validação rejeita
entradas duplicadas, traversal, links simbólicos, diretórios,
checksums/tamanhos divergentes e caminhos que indiquem token, secret,
credencial ou senha.

Os códigos públicos são `PACKAGE_INVALID`, `CHECKSUM_MISMATCH` e
`SCHEMA_INCOMPATIBLE`. Essa etapa somente prova que o pacote é consistente e
compatível com o schema informado pelo chamador; ela **não** executa
`pg_dump`/`pg_restore`, não cifra backup, não instala serviço e não prova
restore, upgrade ou rollback. Esses fluxos continuam pendentes dos gates
OPS-02/OPS-03 e G6/G7.

O `installer-helper verify-package --package-dir PATH [--manifest PATH]
--current-schema VERSION` expõe essa validação como gate operacional. Emite
`PACKAGE_VERIFIED` e código `0` quando todos os arquivos conferem, ou
`PACKAGE_INVALID` e código `2` quando o pacote deve ser bloqueado. O JSON de
erro nunca devolve caminhos absolutos.

O piso do fresh install é 4 GiB livres. Upgrade e restore devem substituir
esse piso por uma necessidade calculada que inclua dados atuais, backup,
temporários e conjunto completo de rollback.

## Probes

No Windows, os probes consultam arquitetura/versão, token elevado, chaves de
reboot pendente, espaço livre no volume, estado e segurança dos diretórios,
bind TCP local da porta preferida e registro do WebView2 Evergreen. Caminhos
UNC/rede, reparse points e marcadores de diretórios sincronizados (OneDrive,
Dropbox e SharePoint) são bloqueados sem que o caminho completo apareça no
JSON. A baseline de sistema é
Windows 10 x64 21H2 (build 19044) ou posterior; Windows 11 também reporta
major version 10 e builds acima dessa baseline. No Windows, o teste da
porta abre e fecha um listener loopback; isso não cria regra de firewall nem
mantém uma porta reservada.

O pacote separa a política (`Evaluate`) dos probes (`Probes`). Os testes usam
implementações injetadas para cobrir sucesso, limites inclusivos, bloqueios,
falhas internas e ausência de plataforma. A implementação `!windows` mantém
o repositório compilável em CI e retorna `BLOCKED` sem alegar suporte.

O JSON não inclui paths completos, nome de usuário, mensagens de erro do
Windows, tokens ou valores de registro além de uma versão sanitizada do
WebView2.
