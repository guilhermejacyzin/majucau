# Verificação das permissões do PostgreSQL do worker — 13/10/2026

## O que foi conferido

A CI agora cria um papel `majucau_runtime` sem privilégios administrativos num PostgreSQL temporário, aplica `database/privileges/runtime.sql` e executa `database/privileges/verify_runtime.sql`.

O verificador confirma que o papel:

- pode entrar no banco, mas não é superusuário, não cria bancos ou papéis e não herda permissões de outro papel;
- pode usar o esquema e ler as tabelas necessárias, sem criar objetos;
- pode inserir dados RAW e eventos de auditoria, mas não pode alterar nem apagar diretamente esses registros;
- consegue executar somente as funções aprovadas para fazer as transições de versão e redação RAW; o público em geral não pode executar essas duas funções.

As tentativas de alteração e exclusão usam `WHERE false` e são feitas numa transação revertida. A verificação não altera linhas, mesmo se uma permissão estiver incorreta; nesse caso, a CI falha.

## Resultado

- A CI [38379095626](https://github.com/guilhermejacyzin/majucau/actions/runs/38379095626) passou nos quatro grupos: backlog, detecção de mudanças, Windows e PostgreSQL.
- O passo novo foi executado no serviço PostgreSQL descartável da CI. O banco local de desenvolvimento e dados financeiros não foram acessados.
- BK-027 passa de 30% para 35%: faltam 65% desta tarefa. O projeto permanece com 11,0% concluído e 89,0% restante pela média simples do backlog.

## O que continua pendente

O teste cria o papel temporário diretamente para verificar a política SQL. Ele não prova que o instalador provisiona essa conta corretamente, nem substitui uma validação em VM limpa. Também não verifica a conexão real com o Bling ou dados financeiros.

Nenhuma tela aprovada, regra de negócio ou regra financeira foi alterada.
