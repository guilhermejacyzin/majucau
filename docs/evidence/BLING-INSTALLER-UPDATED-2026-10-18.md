# Correção da configuração do Bling e instalador atualizado — 18/10/2026

## Causa encontrada

A tela carregava Client ID e endereço de retorno somente do banco local. Como esses dados ainda não tinham sido salvos nesta instalação, os dois campos apareciam vazios. Sem eles, o app não consegue montar o pedido de autorização OAuth. As bolinhas no campo antigo não permitem concluir se o segredo estava cadastrado ou se eram apenas texto de exemplo; esse estado não foi presumido.

O campo secreto não pode reaproveitar a senha de entrada no site Bling. Ele exige o Client Secret do aplicativo criado no Bling. O código mantém esse valor fora do banco e da interface de resposta, no cofre protegido do Windows.

## Correção aplicada

- A tela continua buscando os valores que já estejam salvos; esses valores prevalecem.
- Se o banco ainda estiver sem Client ID ou endereço de retorno, a tela agora os preenche com os dados públicos do aplicativo Majucau que a usuária forneceu anteriormente.
- Os valores aparecem nos campos já existentes e só são gravados ao pressionar “Salvar configuração”. Nenhum novo elemento visual foi criado.
- O Client Secret não recebe valor padrão, não é exibido após salvar e precisa ser digitado diretamente na tela do Majucau.
- A versão nova consulta o cofre local e mostra o estado do segredo sem ler a senha de acesso ao Bling. Se indicar “Informe o segredo do app”, a usuária deve obter o Client Secret na configuração do aplicativo Bling e inseri-lo no Majucau.
- Nenhuma regra financeira, regra de negócio ou layout aprovado foi alterado.

## Verificações e pacote local

`scripts/build-windows.ps1 -Installer` terminou com sucesso em 18/10/2026. Foram aprovados: TypeScript/Vite, 33 testes frontend, testes Go, `go vet`, `govulncheck` (nenhuma vulnerabilidade chamada encontrada), ESLint, npm audit, verificação de secrets, checksums e geração determinística SQLC, build do app/worker/helper e empacotamento NSIS/portátil. O Bootstrapper WebView2 tinha certificado Microsoft Authenticode válido; sua mudança de hash foi aceita pelo script de validação. O instalador produzido ainda não tem assinatura digital de produção e serve para o protótipo/homologação local.

| Arquivo local | Tamanho | SHA-256 |
|---|---:|---|
| `artifacts/windows/Majucau Financial Intelligence-amd64-installer.exe` | 20.050.679 bytes | `7e48cfc948cb8cdf8bcaf78856185c68f91d350651f03d2d3f1edbd8019d9968` |
| `artifacts/windows/majucau-windows-x64-portable.zip` | 18.467.309 bytes | `2f35e820c71cd729c766eb87e77521d4148d70479fbfb1096030b949aef983ef` |

Os artefatos são saídas locais consolidadas em `artifacts/windows/` e permanecem fora do Git conforme a política do projeto.

## O que ainda é necessário

1. Instalar e abrir o novo pacote no protótipo.
2. No Majucau, confirmar Client ID/endereço de retorno já preenchidos, digitar o **Client Secret do aplicativo** no campo correspondente e pressionar “Salvar configuração”. Não enviar essa chave no chat.
3. Depois de salvo, conectar, autorizar no site Bling e testar a conexão. Ainda não houve acesso nem leitura real dos dados financeiros.

O código pode preparar a autorização com os campos públicos, mas não pode inventar nem recuperar o Client Secret. A integração permanece não conectada até os passos manuais acima e a homologação real.
