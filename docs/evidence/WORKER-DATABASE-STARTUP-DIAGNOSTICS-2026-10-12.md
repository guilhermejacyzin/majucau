# Diagnóstico seguro de inicialização do PostgreSQL — 2026-10-12

## Alteração

`cmd/worker/database.go` agora retorna um código estável para as três condições que impedem o pool local de iniciar:

- `DATABASE_NOT_CONFIGURED`: a configuração do banco não foi fornecida;
- `DATABASE_CONFIG_INVALID`: o endereço configurado não foi aceito pelo parser;
- `DATABASE_UNAVAILABLE`: não foi possível criar ou confirmar a conexão.

`cmd/worker/process.go` registra apenas esse código no log estruturado. Não registra DSN, senha nem erro bruto do driver. A rotina de conexão continua limitada a cinco segundos.

O worker continua disponível para operações que não precisam do banco, como a prévia local de arquivos. Esta informação não declara o banco pronto e não transforma ausência de dados em valor zero. Nenhuma tela, schema, regra de negócio ou cálculo financeiro foi alterado.

## Limites

`DATABASE_UNAVAILABLE` ainda agrupa falhas de abertura e de confirmação da conexão; não classifica DNS, porta ou autenticação. A saúde dinâmica do banco e a validação em instalação/VM continuam pendentes. A execução desta mudança não rodou testes locais; a CI do commit é a verificação automatizada subsequente.
