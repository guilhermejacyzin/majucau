# Instalador atualizado para a configuração do Bling — 18/10/2026

## Causa e correção

O código que corrige a leitura/salvamento da configuração do Bling já estava no repositório, incluindo os commits `caed9d5` e `12901b5`. Porém, o instalador NSIS local tinha sido gerado em 13/10/2026. Assim, quem instalasse por esse arquivo poderia continuar usando uma versão anterior à correção.

Reconstruí o aplicativo Wails, o worker, o auxiliar e o instalador NSIS usando o código atual da branch. Também regenerei o pacote portátil. A tela aprovada e as regras financeiras/contábeis não foram alteradas.

## Arquivos atualizados

| Arquivo | Tamanho | SHA-256 |
|---|---:|---|
| `artifacts/windows/Majucau Financial Intelligence-amd64-installer.exe` | 20.049.553 bytes | `3d3d9d34d4dbfdc59ce4851bce22afd38d924eb31f9bac2533f62436f48a8bfa` |
| `artifacts/windows/majucau-windows-x64-portable.zip` | 18.465.830 bytes | `d3e89ec7738df1f08a9ad7561a275b382a30881a5510c6828592eb12b15dc02f` |

O manifesto do pacote registra os hashes dos executáveis e migrations. Os artefatos ficam em `artifacts/windows/`, o local consolidado definido pelo projeto, e são saídas locais ignoradas pelo Git.

## Limites e próximo passo

- Foi feita compilação do app, worker e auxiliar, além do empacotamento NSIS. **Não foram executados testes** nesta atualização.
- Não houve acesso à conta Bling nem leitura de dados financeiros.
- O instalador continua sem assinatura e serve para homologação local, não para produção.
- Este pacote não cria nem recupera um Client Secret. As bolinhas antigas eram apenas um exemplo visual; a senha usada para entrar no site do Bling não é a chave Client Secret do aplicativo.
- Após instalar/abrir a versão nova, conferir se Client ID e endereço de retorno aparecem. Se estiverem vazios, salvá-los no formulário existente. Depois, informar o Client Secret do aplicativo diretamente nessa tela, salvar, conectar, autorizar e testar a conexão.
- BK-041 permanece em **80%** (20% restante); autorização e leitura reais, homologação, integração de `/caixas`, mapeamento contábil de dados reais e validação final continuam pendentes. A média do projeto permanece em **11,6%** (88,4% restante).
