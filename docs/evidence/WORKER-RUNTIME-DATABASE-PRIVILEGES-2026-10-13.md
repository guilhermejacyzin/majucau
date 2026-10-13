# Verificação das permissões do PostgreSQL do worker — 13/10/2026

## O que foi conferido

A CI agora cria um papel `majucau_runtime` sem privilégios administrativos num PostgreSQL temporário, aplica `database/privileges/runtime.sql` e executa `database/privileges/verify_runtime.sql`.

O verificador confirma que o papel:

- pode entrar no banco, mas não é superusuário, não cria bancos ou papéis, não herda permissões e não é dono de objetos;
- pode usar o esquema e ler somente as relações de integração e painel que o worker atual consulta;
- grava somente configuração/estado das integrações, lotes, cursores, RAW, recebimentos e recebíveis futuros que já têm fluxo implementado;
- não pode apagar linhas, alterar RAW/auditoria diretamente, inserir diretamente na auditoria, alterar o catálogo de acesso ou escrever demonstrativos;
- só pode executar a função controlada de encerramento da versão RAW usada pela sincronização; a função de redação continua sem acesso do papel de execução até existir um caminho autenticado no aplicativo;
- não pode executar funções, procedimentos ou outras rotinas públicas arbitrárias nem usar sequências.

As tentativas diretas de alteração, exclusão e inserção na auditoria usam `WHERE false` e uma transação revertida. O verificador percorre todas as relações e funções do esquema e compara as permissões efetivas com a lista permitida; uma tabela ou função nova não recebe acesso silenciosamente.

## Resultado da primeira validação

- A CI [38379095626](https://github.com/guilhermejacyzin/majucau/actions/runs/38379095626) passou nos quatro grupos: backlog, detecção de mudanças, Windows e PostgreSQL.
- O passo foi executado no serviço PostgreSQL descartável da CI. O banco local de desenvolvimento e dados financeiros não foram acessados.
- BK-027 passa de 30% para 35%: faltam 65% desta tarefa. O projeto permanece com 11,0% concluído e 89,0% restante pela média simples do backlog.

## O que continua pendente

O teste cria o papel temporário diretamente para verificar a política SQL. Ele não prova que o instalador provisiona essa conta corretamente, nem substitui uma validação em VM limpa. Também não verifica a conexão real com o Bling ou dados financeiros.

Nenhuma tela aprovada, regra de negócio ou regra financeira foi alterada.

## Restrição de privilégios por operação — 13/10/2026

A revisão das consultas de produção encontrou uma concessão ampla: `majucau_runtime` recebia leitura, inclusão, alteração e exclusão em todas as tabelas do esquema, embora o código atual só grave dados de integração. O papel agora recebe somente a leitura das dez relações usadas pelos fluxos existentes e as operações de escrita específicas de configuração/status, lotes, cursores, RAW, recebimentos e recebíveis futuros. Exclusão, escrita em auditoria, tabelas contábeis e catálogo de acesso ficam bloqueados. A função de redação RAW não é executável pelo worker até haver uma operação autenticada que a use.

O verificador foi ajustado para conferir, relação por relação, as permissões exatas de leitura e escrita, inclusive acessos concedidos por coluna; propriedade de objetos; execução de funções, procedimentos e demais rotinas; ausência de privilégios em sequências; e bloqueio de inclusão direta na auditoria. A nova versão aguarda a primeira execução completa da CI depois do push; ainda não foi aplicada ao banco local, a qualquer instalação nem a banco com dados da Majucau.
