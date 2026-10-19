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
- Corrigi a distribuição local da correção do Bling: o instalador NSIS estava datado de 13/10 e foi reconstruído a partir do código atual, que já recupera Client ID/Redirect URI e não usa bolinhas como prova de segredo salvo. O app precisa ser atualizado; em seguida, a usuária deve salvar o Client Secret do aplicativo diretamente no Majucau e autorizar a conexão. Build do aplicativo, worker, auxiliar e instalador concluído; não executei testes nem acessei a conta real nesta execução. Evidência: `docs/evidence/BLING-INSTALLER-UPDATED-2026-10-18.md`.
- Próxima entrega técnica em andamento: ligar o estágio privado de arquivos do Drive aos importadores homologados; a confirmação da última página do Bling também continua necessária para homologação.
- Nesta atualização, o Majucau passou a guardar páginas confirmadas do Bling e retomá-las depois de uma interrupção. A publicação do novo retrato acontece inteira ou não acontece. A CI do GitHub passou para o commit `951d4a6`.
- Nesta execução, corrigi uma suposição insegura: página vazia sem explicação do Bling não comprova que a lista terminou. O sistema agora falha com segurança, preserva os dados válidos anteriores e não avança o cursor. Os testes locais, `go vet`, teste PostgreSQL, verificação do backlog e CI Windows passaram no commit `aff28d3` ([CI #39173420786](https://github.com/guilhermejacyzin/majucau/actions/runs/39173420786)).
- No Drive, a listagem agora tem checkpoint persistente de cada página, versões de metadados e publicação atômica do inventário completo. Se a execução falhar, o retrato anterior permanece e a leitura pode continuar do último marcador confirmado. Depois de incluir os modelos SQLC, a CI completa passou no commit `ef83bed` nos quatro jobs Windows, PostgreSQL, backlog e detecção de mudanças. BK-042 está em 52%; não houve homologação real.
- Nesta execução, o inventário do Drive foi ligado ao worker diário às 09:30. Cada uma das três pastas tem correlação idempotente por data e falha independente. Os testes locais do worker e do adapter passaram. O agendamento só inicia depois de configurar os três IDs e a credencial DPAPI; a migration registra a integração como `NOT_CONFIGURED`. BK-042 está em 56%. Nenhum conteúdo foi baixado ou importado, e não houve conexão real com o Drive.
- Nesta atualização, o download verificado do Drive ganhou uma área temporária privada e um leitor em fluxo para o processador. O arquivo só fica disponível depois da validação de tamanho, checksum e versão; falha, retry e fim do processamento removem o temporário, e erro de limpeza mantém um diagnóstico seguro. Os testes do adapter passaram, inclusive download interrompido e tentativa novamente. BK-042 agora está estimada em 58% (42% restante); ainda não está ligado ao worker/importadores nem houve conexão real ou importação. O projeto continua em 11,6% (88,4% restante).
- A conferência da regra JWT do Bling confirmou que o código envia `enable-jwt: 1` ao trocar e renovar tokens e em todas as consultas autenticadas. Isso confirma compatibilidade do código, mas não a autorização da conta real.
- As telas aprovadas e as regras financeiras permaneceram inalteradas.

## O que ainda falta para homologar o protótipo

1. Autorizar e validar a conexão real do Bling e confirmar como a lista informa que chegou à última página.
2. Ligar o estágio privado aos importadores depois de homologar os contratos das três pastas; CSV, TSV, JSON e Excel ainda não foram processados pelo job diário.
3. Rodar o protótipo com dados de homologação, conferir os resultados e registrar qualquer divergência antes de produção.

Além desses pontos, há outras tarefas do backlog para concluir a operação, segurança, instalação e produção. Ainda não há prazo total confiável: a estimativa depende de validações externas e de tarefas que ainda não começaram.

## Próximo passo

Homologar os contratos dos arquivos e ligar o estágio privado ao processamento idempotente. Em seguida, configurar e validar o acesso real ao Drive e ao Bling; no Bling, não publicar a carga até confirmar como identificar a última página.
