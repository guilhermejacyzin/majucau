# Meta do Majucau

## O que queremos entregar

Concluir o backlog aprovado do Majucau e entregar primeiro um protótipo que possa ser testado e homologado. Depois da homologação, seguir para a finalização e entrada em produção previstas no próprio backlog.

## Regras que devem ser preservadas

- Não alterar telas já aprovadas.
- Não alterar regras financeiras ou de negócio aprovadas.
- Se houver dúvida sobre um requisito, dado ou interpretação, parar essa parte do trabalho e perguntar antes de implementar.
- Nunca alterar telas aprovadas nem regras financeiras ou de negócio sem autorização explícita.
- Registrar cada avanço, teste, pendência e próximo passo no projeto e enviar as atualizações ao GitHub.

## Andamento registrado

- 127 tarefas no backlog: 7 concluídas, 19 em andamento e 101 não iniciadas.
- Progresso estimado: **11,6% concluído; 88,4% restante**. Essa média serve para acompanhamento do backlog e não prova que o sistema esteja pronto para homologação ou produção.
- Próxima entrega técnica em andamento: ligar o download do Drive ao staging seguro e ao fluxo de arquivos; a confirmação da última página do Bling também continua necessária para homologação.
- Nesta atualização, o Majucau passou a guardar páginas confirmadas do Bling e retomá-las depois de uma interrupção. A publicação do novo retrato acontece inteira ou não acontece. A CI do GitHub passou para o commit `951d4a6`.
- Nesta execução, corrigi uma suposição insegura: página vazia sem explicação do Bling não comprova que a lista terminou. O sistema agora falha com segurança, preserva os dados válidos anteriores e não avança o cursor. Os testes locais, `go vet`, teste PostgreSQL, verificação do backlog e CI Windows passaram no commit `aff28d3` ([CI #39173420786](https://github.com/guilhermejacyzin/majucau/actions/runs/39173420786)).
- No Drive, a listagem agora tem checkpoint persistente de cada página, versões de metadados e publicação atômica do inventário completo. Se a execução falhar, o retrato anterior permanece e a leitura pode continuar do último marcador confirmado. Depois de incluir os modelos SQLC, a CI completa passou no commit `ef83bed` nos quatro jobs Windows, PostgreSQL, backlog e detecção de mudanças. BK-042 está em 52%; não houve homologação real.
- Nesta execução, o inventário do Drive foi ligado ao worker diário às 09:30. Cada uma das três pastas tem correlação idempotente por data e falha independente. Os testes locais do worker e do adapter passaram. O agendamento só inicia depois de configurar os três IDs e a credencial DPAPI; a migration registra a integração como `NOT_CONFIGURED`. BK-042 está em 56%. Nenhum conteúdo foi baixado ou importado, e não houve conexão real com o Drive.
- A conferência da regra JWT do Bling confirmou que o código envia `enable-jwt: 1` ao trocar e renovar tokens e em todas as consultas autenticadas. Isso confirma compatibilidade do código, mas não a autorização da conta real.
- As telas aprovadas e as regras financeiras permaneceram inalteradas.

## O que ainda falta para homologar o protótipo

1. Autorizar e validar a conexão real do Bling e confirmar como a lista informa que chegou à última página.
2. Ligar o download dos arquivos do Drive ao staging seguro e processar os formatos recebidos em fluxo, inclusive Excel após parser e contratos homologados.
3. Rodar o protótipo com dados de homologação, conferir os resultados e registrar qualquer divergência antes de produção.

Além desses pontos, há outras tarefas do backlog para concluir a operação, segurança, instalação e produção. Ainda não há prazo total confiável: a estimativa depende de validações externas e de tarefas que ainda não começaram.

## Próximo passo

Ligar o download do Drive ao staging seguro e ao processamento idempotente dos arquivos. Em seguida, configurar e validar o acesso real ao Drive e ao Bling; no Bling, não publicar a carga até confirmar como identificar a última página.
