# Retorno OAuth do Bling ao Majucau local — 13/10/2026

> **Revisão de evidência — 26/10/2026:** Gisele confirmou que o aplicativo Bling
> está cadastrado com o retorno
> `https://majucau-bling-bridge.giselegalha.chatgpt.site/oauth/bling/callback`.
> O conector Sites confirma que a ponte está publicada na versão 16, criada a
> partir do commit `fdce727d8f3d868b5bea418d52208039054e7cf8`. Abri essa versão no
> cache consolidado `artifacts/cache/sites/majucau-bling-bridge` e confirmei que
> a rota `/oauth/bling/callback` encaminha o código temporário e o `state` para
> `http://127.0.0.1:43821/callback`, destino fixo no código. Os 28 testes da
> versão 16, a compilação e a validação do pacote passaram localmente. Isso
> confirma a implementação da ponte, mas não uma autorização real: o cofre do
> protótipo continua sem o Client Secret e não houve OAuth ou sincronização.

## Resultado em linguagem simples

O Bling está cadastrado para devolver a autorização ao endereço HTTPS do Site Majucau. O código da versão 16 encaminha o retorno ao endereço local fixo `127.0.0.1:43821`. O Majucau confere o `state` e troca o código usando a URL HTTPS registrada. A ponte foi comprovada no código e nos testes; falta concluir uma autorização real pela conta da empresa.

O Site não troca nem guarda o token nessa autorização iniciada pelo Majucau local. O token é validado e guardado pelo Majucau no cofre protegido do Windows. O fluxo iniciado pelo botão do próprio Site permanece compatível.

## Proteções implementadas

- O Majucau aceita automaticamente somente o callback HTTPS exato já cadastrado e conhecido: `https://majucau-bling-bridge.giselegalha.chatgpt.site/oauth/bling/callback`. Outros callbacks HTTPS não reconhecidos são rejeitados.
- O Majucau continua gerando `state` com 32 bytes aleatórios, confere a correspondência antes de aceitar o código e abre o listener apenas em `127.0.0.1:43821`. O listener limita o cabeçalho/request line a 16 KiB e o código a 4 KiB.
- A versão 16 só aceita para o encaminhamento local um `state` no formato hexadecimal de 32 bytes; o destino é fixo no código e nunca vem da solicitação. O Site devolve somente o código de autorização e o `state`, limita o código a 4 KiB e recusa caracteres de controle. Os testes da versão 16 verificam esses casos.
- O retorno local da versão 16 não é armazenado em cache, não envia referenciador e não depende do banco do Site; cancelamento volta ao Majucau como erro controlado. Testes locais verificam o comportamento, mas não simulam o navegador ou a conta real.
- O Majucau troca o código de uso único usando o HTTPS cadastrado, valida a conexão com uma consulta de leitura e armazena os tokens no cofre local. A resposta do callback remove os parâmetros OAuth da barra do navegador, não é armazenada em cache e usa política de conteúdo restritiva.
- A sincronização de dados continua sendo uma ação separada; autorizar não altera nem inicia uma sincronização financeira.

## Verificações e publicação

- Testes locais do pacote Go Bling passaram após a mudança do callback e da limpeza dos parâmetros da URL.
- As 20 verificações do Site passaram. A compilação Worker e a validação do pacote também passaram.
- Site v13 publicado com sucesso em `https://majucau-bling-bridge.giselegalha.chatgpt.site`, a partir do commit de origem `4c8cf47262df49628136aa2908d7f69c8045c201`. A versão atualmente publicada é a 16, criada a partir do commit `fdce727d8f3d868b5bea418d52208039054e7cf8`; seu código-fonte foi aberto no cache consolidado sem alterações nem publicação.
- Na cópia exata da versão 16, `npm test` passou com 28 testes; `npm run build` e `npm run validate` também passaram. `npm audit --omit=dev` não encontrou vulnerabilidades nas dependências de produção. A auditoria completa registrou quatro alertas moderados em ferramentas de desenvolvimento: `drizzle-kit` inclui um `esbuild` antigo usado pelo servidor de desenvolvimento; isso não aparece nas dependências do pacote de produção. A atualização dessa ferramenta continua registrada para avaliação, sem aplicar downgrade automático.
- A versão foi publicada mantendo o público e a URL que já estavam configurados. Nenhuma tela aprovada ou regra financeira foi alterada.
- O fluxo foi validado com testes automatizados, não com uma autorização real da conta Majucau. Ainda não há evidência de conta conectada ou dados importados.

## Próximo passo para a responsável

Na tela **Importações / Integrações** do Majucau, preencher Client ID, Client Secret e Redirect URI com as credenciais do aplicativo Bling e o endereço HTTPS acima; salvar; clicar **Conectar**; concluir a autorização no navegador. Depois, testar a conexão e iniciar a sincronização com um filtro aprovado. Não enviar o Client Secret, senha ou token pelo chat.

Após conectar, ainda será preciso validar os registros da conta, confirmar como o Bling sinaliza a última página e reconciliar a primeira sincronização. A atualização diária das 09:30 e a leitura dos arquivos do Drive também continuam pendentes. Na versão 13 do Site, a importação do Drive ainda acumulava a árvore de arquivos e o CSV em memória. Esse ponto foi corrigido na versão 15 com listagem paginada, download em fluxo, limites e conferência da versão do arquivo; os detalhes e os limites restantes estão em `docs/evidence/DRIVE-STREAMING-IMPORT-2026-10-13.md`. A homologação com a pasta real, a medição de carga e o ensaio em VM continuam pendentes.

BK-041 foi atualizado de 69% para 75% (25% restante): o caminho do retorno OAuth está implementado e publicado, mas conexão real, paginação e sincronização homologada faltam. A CI completa passou no run [38343452090](https://github.com/guilhermejacyzin/majucau/actions/runs/38343452090), com validações no Windows e PostgreSQL, análise de segurança e montagem do protótipo. A recontagem mais recente do backlog aponta 10,9% concluído (89,1% restante); não há prazo total confiável.

## Reconfirmação do endereço de retorno — 13/10/2026

- Gisele enviou novamente o endereço configurado no aplicativo Bling. Ele corresponde exatamente ao callback HTTPS fixo aceito pelo Majucau em `internal/integrations/bling/oauth_service.go` e pela ponte publicada.
- Esta conferência confirma a correspondência do endereço, não uma autorização real nem uma sincronização. A tentativa de abrir o cadastro do Bling no Chrome foi recusada por uma preferência de segurança salva; nenhum caminho alternativo foi usado.
- Continuam pendentes salvar no Majucau as credenciais do aplicativo, concluir a autorização, testar a conexão e homologar uma sincronização real. Client Secret, senha e tokens não foram consultados nem registrados.
- BK-041 permanece em 75% (25% restante). A média do backlog é 10,9% concluída (89,1% restante). Nenhuma tela, regra financeira ou configuração da conta foi alterada.

## Conferência da atividade local — 13/10/2026

- Revisei novamente o log sanitizado do worker local. Ele tem 28 registros válidos; as operações registradas são consultas de saúde, estado das integrações e painel. Não há início de OAuth, teste de conexão ou sincronização do Bling nesta sessão.
- Portanto, os dados ainda não foram pedidos à API nessa execução. Isso não demonstra falha da API: mostra que o fluxo de conectar/testar/sincronizar ainda não foi iniciado no protótipo.
- O próximo passo operacional é salvar no Majucau as configurações do aplicativo Bling e usar **Conectar**, **Testar conexão** e **Sincronizar agora**, nessa ordem. A senha usada para entrar no site do Bling é diferente do Client Secret da integração; não enviar o Client Secret pelo chat.
- BK-041 permanece em 75% (25% restante), porque ainda faltam uma autorização real, teste, leitura paginada e conferência dos dados importados. Não alterei telas, regras financeiras, credenciais ou configurações externas.

## Diagnóstico da configuração local — 13/10/2026

- A tela local do Majucau estava com Client ID e Redirect URI vazios e Bling em “Não conectado”. Preenchi os dois campos com os dados públicos enviados pela responsável; o campo Client Secret ficou intocado.
- Ao salvar, o worker respondeu `BLING_SECRET_REQUIRED`. A configuração não foi salva, não houve tentativa de OAuth nem chamada à API do Bling. O Client Secret não foi lido, copiado ou registrado.
- A senha de entrada do site Bling é diferente do Client Secret do aplicativo. Para continuar, a responsável precisa preencher o Client Secret diretamente no campo correspondente do Majucau e salvar. Depois disso, o fluxo pode seguir com autorização, teste e sincronização.
- BK-041 permanece em 75% (25% restante). Nenhuma tela aprovada ou regra financeira foi alterada; os dados públicos estão apenas preenchidos no formulário local até o segredo ser informado.
