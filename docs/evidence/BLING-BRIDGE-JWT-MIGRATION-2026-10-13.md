# Revisão e correção da ponte Bling — 13/10/2026

## Resultado em linguagem simples

O Bling conseguiu concluir a autorização no Site público “Majucau Bling Bridge”. Isso confirma que o Site recebe o retorno de autorização. Não confirma que o Majucau aberto no computador recebeu dados: o Site guarda o token protegido e, no código revisado, não existe uma rota que consulte as contas do Bling e entregue os resultados ao Majucau local.

O programa local tem outro caminho de OAuth: ele aguarda o retorno em um endereço local `http://127.0.0.1:<porta>/<caminho>`. O aplicativo do Bling está cadastrado para retornar ao Site em HTTPS. Esses dois caminhos não se encontram. Por isso, a tela pode mostrar o Bling como não conectado e nenhuma conta financeira ser carregada, mesmo que a autorização do Site tenha sido concluída.

## Correção feita no Site

- Atualizada a troca e a renovação de token para `https://api.bling.com.br/Api/v3/oauth/token` e inclusão do cabeçalho `enable-jwt: 1`, seguindo o [guia oficial de migração JWT do Bling](https://developer.bling.com.br/migracao-jwt) e a [migração oficial do endereço da API](https://developer.bling.com.br/migracao-url).
- A leitura da resposta de token agora é feita em fluxo, com limite de 64 KiB, evitando carregar uma resposta sem limite na memória.
- Adicionados cabeçalhos de segurança à resposta da página. Depois de concluir a autorização, o Site redireciona para um endereço limpo, sem manter o código temporário na barra do navegador.
- A tela de sucesso e as telas aprovadas do Majucau não foram redesenhadas; nenhuma regra financeira foi alterada.

## Verificação e publicação

- Site v10 publicado em `https://majucau-bling-bridge.giselegalha.chatgpt.site` a partir do commit de código-fonte `bf1a76260691939d5be505474bf6941cb795b21a`.
- As 14 verificações automatizadas existentes do Site passaram; a compilação do pacote do Site e a validação do artefato também passaram.
- Em uma navegação anterior pelo Site v9, a autorização terminou na mensagem “Bling conectado”. Isso prova apenas o retorno ao Site, não a sincronização com o programa local. A versão v10 foi compilada, validada e publicada; não foi feita nova sincronização financeira após a publicação.
- O código do Site guarda o token cifrado no banco do próprio Site. Nenhum Client Secret ou token foi copiado para o repositório ou para este registro.

## O que falta para carregar dados no protótipo

1. Ligar a autorização ao Majucau local. Para o protótipo, o caminho mais simples tende a ser cadastrar no Bling o retorno local `http://127.0.0.1:43821/callback` e configurar o mesmo endereço no Majucau. A aceitação desse endereço local pelo cadastro do aplicativo Bling ainda precisa ser confirmada na própria tela do Bling; não foi presumida.
2. Depois da conexão local, testar a leitura de contas, confirmar o fim da paginação e conferir os valores importados contra o Bling.
3. Só então considerar o fluxo de atualização diária das 09:30 homologado. O Site, por si só, não executa essa atualização no Majucau local.

Alternativamente, pode-se projetar um encaminhamento seguro entre o Site e o Majucau local. Isso requer autenticação entre os dois, autorização para entregar somente o necessário, proteção contra repetição e limites de tamanho. Esse encaminhamento não foi implementado.

## Estado do projeto

BK-041 continua em 69% (31% restante): falta conexão utilizável pelo Majucau, confirmação de paginação e sincronização real homologada. O projeto permanece na estimativa do backlog de 10,2% concluído (89,8% restante); isso mede tarefas registradas e não significa que o protótipo esteja 10,2% pronto para produção. Não existe base suficiente para prometer uma data de conclusão.
