# Retorno OAuth do Bling ao Majucau local — 13/10/2026

## Resultado em linguagem simples

O Bling já está cadastrado para devolver a autorização ao endereço HTTPS do Site Majucau. O Majucau instalado no computador, porém, espera receber o retorno em `127.0.0.1`. O fluxo agora liga esses dois endereços: o Site entrega somente o código temporário ao aplicativo local, e o aplicativo troca o código pelo token usando a URL de retorno HTTPS registrada no Bling.

O Site não troca nem guarda o token nessa autorização iniciada pelo Majucau local. O token é validado e guardado pelo Majucau no cofre protegido do Windows. O fluxo iniciado pelo botão do próprio Site permanece compatível.

## Proteções implementadas

- O Majucau aceita automaticamente somente o callback HTTPS exato já cadastrado e conhecido: `https://majucau-bling-bridge.giselegalha.chatgpt.site/oauth/bling/callback`. Outros callbacks HTTPS não reconhecidos são rejeitados.
- O Majucau continua gerando `state` com 32 bytes aleatórios, confere a correspondência antes de aceitar o código e abre o listener apenas em `127.0.0.1:43821`. O listener limita o cabeçalho/request line a 16 KiB e o código a 4 KiB.
- O Site só reconhece o formato de `state` gerado pelo fluxo local; o endereço de destino é fixo e nunca vem da solicitação. O Site devolve somente o código de autorização e `state`, com limite de 4 KiB em bytes e recusa caracteres de controle.
- A resposta do Site não é armazenada em cache, não envia referenciador e não depende do banco do Site. Cancelamento volta ao Majucau como erro controlado.
- O Majucau troca o código de uso único usando o HTTPS cadastrado, valida a conexão com uma consulta de leitura e armazena os tokens no cofre local. A resposta do callback remove os parâmetros OAuth da barra do navegador, não é armazenada em cache e usa política de conteúdo restritiva.
- A sincronização de dados continua sendo uma ação separada; autorizar não altera nem inicia uma sincronização financeira.

## Verificações e publicação

- Testes locais do pacote Go Bling passaram após a mudança do callback e da limpeza dos parâmetros da URL.
- As 20 verificações do Site passaram. A compilação Worker e a validação do pacote também passaram.
- Site v13 publicado com sucesso em `https://majucau-bling-bridge.giselegalha.chatgpt.site`, a partir do commit de origem `4c8cf47262df49628136aa2908d7f69c8045c201`.
- A versão foi publicada mantendo o público e a URL que já estavam configurados. Nenhuma tela aprovada ou regra financeira foi alterada.
- O fluxo foi validado com testes automatizados, não com uma autorização real da conta Majucau. Ainda não há evidência de conta conectada ou dados importados.

## Próximo passo para a responsável

Na tela **Importações / Integrações** do Majucau, preencher Client ID, Client Secret e Redirect URI com as credenciais do aplicativo Bling e o endereço HTTPS acima; salvar; clicar **Conectar**; concluir a autorização no navegador. Depois, testar a conexão e iniciar a sincronização com um filtro aprovado. Não enviar o Client Secret, senha ou token pelo chat.

Após conectar, ainda será preciso validar os registros da conta, confirmar como o Bling sinaliza a última página e reconciliar a primeira sincronização. A atualização diária das 09:30 e a leitura dos arquivos do Drive também continuam pendentes. Na versão 13 do Site, a importação do Drive ainda acumulava a árvore de arquivos e o CSV em memória. Esse ponto foi corrigido na versão 15 com listagem paginada, download em fluxo, limites e conferência da versão do arquivo; os detalhes e os limites restantes estão em `docs/evidence/DRIVE-STREAMING-IMPORT-2026-10-13.md`. A homologação com a pasta real, a medição de carga e o ensaio em VM continuam pendentes.

BK-041 foi atualizado de 69% para 75% (25% restante): o caminho do retorno OAuth está implementado e publicado, mas conexão real, paginação e sincronização homologada faltam. A CI completa passou no run [38343452090](https://github.com/guilhermejacyzin/majucau/actions/runs/38343452090), com validações no Windows e PostgreSQL, análise de segurança e montagem do protótipo. A estimativa do projeto segue em 10,2% concluída (89,8% restante); não há prazo total confiável.
