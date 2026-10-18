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
- Progresso estimado: **11,5% concluído; 88,5% restante**. Essa média serve para acompanhamento do backlog e não prova que o sistema esteja pronto para homologação ou produção.
- Próxima entrega técnica em andamento: retomada segura da sincronização diária do Bling, sem publicar dados parciais.
- Nesta atualização, o Majucau passou a guardar páginas confirmadas do Bling e retomá-las depois de uma interrupção. A publicação do novo retrato acontece inteira ou não acontece. Os testes do pacote passaram em PostgreSQL temporário.
- As telas aprovadas e as regras financeiras permaneceram inalteradas.

## O que ainda falta para homologar o protótipo

1. Autorizar e validar a conexão real do Bling e confirmar como a lista informa que chegou à última página.
2. Ligar a leitura autorizada das três pastas do Drive e tratar os arquivos recebidos, inclusive Excel, em fluxo seguro.
3. Rodar o protótipo com dados de homologação, conferir os resultados e registrar qualquer divergência antes de produção.

Além desses pontos, há outras tarefas do backlog para concluir a operação, segurança, instalação e produção. Ainda não há prazo total confiável: a estimativa depende de validações externas e de tarefas que ainda não começaram.

## Próximo passo

Executar a CI do GitHub para esta implementação. Se ela passar, publicar o commit e então homologar a conexão Bling usando a autorização já configurada, sem expor senha ou segredo no repositório.
