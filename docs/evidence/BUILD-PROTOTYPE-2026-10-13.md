# Build do protótipo para homologação local — 13/10/2026

## Resultado

Foi gerado o pacote Windows x64 a partir do código atual do Majucau, incluindo o caminho de retorno OAuth do Bling. O instalador e o pacote portátil ficam em `artifacts/windows/`, junto com o manifesto e os executáveis auxiliares. Os caches e saídas intermediárias permaneceram dentro de `artifacts/`.

## Correção necessária

Na primeira execução, o NSIS terminou de gerar o instalador, mas o script falhou ao guardar o resultado. No PowerShell, nomes de variáveis não diferenciam maiúsculas de minúsculas: o parâmetro `$Installer` colidia com a variável `$installer`. O script agora guarda o arquivo em `$installerArtifact`. O build seguinte terminou e publicou os artefatos no local central.

## Verificações realizadas

`scripts/build-windows.ps1 -Installer` concluiu todas as etapas locais:

- Testes Go em todos os pacotes do projeto.
- `go vet`, análise de vulnerabilidades, checksum das migrations e verificação determinística de `sqlc`.
- Scanner de segredos e verificação do contrato do instalador NSIS.
- ESLint, TypeScript, 26 testes frontend, build Vite e `npm audit`.
- Build do desktop Wails, worker, helper e instalador NSIS.

O scanner de vulnerabilidades encontrou zero vulnerabilidades alcançáveis pelo código; reportou uma vulnerabilidade em dependência indireta que não parece ser chamada. O `npm audit` não encontrou vulnerabilidades.

`scripts/smoke-windows.ps1` terminou com `Windows smoke PASS`. Ele conferiu o ZIP, os hashes dos arquivos, o manifesto, o preflight e o pacote de diagnóstico; iniciou e encerrou o worker sem conectar a um banco de produção.

## Artefatos atuais

| Arquivo | Tamanho | SHA-256 |
|---|---:|---|
| `artifacts/windows/Majucau Financial Intelligence-amd64-installer.exe` | 19.856.540 bytes | `500ce1ba3a77781feb80e433ca809ddc4765cee65103a9127b118fb141ecde06` |
| `artifacts/windows/majucau-windows-x64-portable.zip` | 18.287.462 bytes | `44978f2fc392b6b985d41598992f6964962b4259789acc3f0242ace29ef6a977` |

O manifesto também registra hashes de `majucau.exe`, `majucau-worker.exe`, `installer-helper.exe` e das migrations incluídas. O instalador é uma versão não assinada para homologação local; não atende ao gate de produção G7.

## Limites e próximo passo

Nenhuma conta Bling foi autorizada, nenhum dado financeiro real foi importado e a interface não foi instalada nesta máquina. A próxima validação depende de a responsável abrir **Integrações → Bling → Conectar**, aprovar o acesso na página do Bling e depois solicitar **Testar conexão** e **Sincronizar agora**. Não pedir nem registrar senha, Client Secret ou token no chat.

Faltam também a instalação e execução em VM Windows limpa, a homologação visual/funcional e o aceite da integração real. Nenhuma tela aprovada ou regra financeira foi alterada.
