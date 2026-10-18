# Bling — aposentadoria de páginas RAW em blocos — 18/10/2026

## O que foi ajustado

Ao concluir uma releitura completa do Bling, o sistema remove do retrato atual as páginas antigas que passaram a estar fora do conjunto lido. Antes, ele primeiro guardava na memória a lista inteira de páginas antigas e só depois as marcava como substituídas.

Agora o worker busca e processa esses identificadores em blocos de até 256. Cada consulta é ordenada por ID e continua da última posição lida; assim, a memória usada para essa limpeza não cresce com a quantidade total de páginas antigas. Uma trava transacional por conexão e recurso também serializa sincronizações manuais e agendadas, para que duas substituições da mesma lista não disputem o retrato atual.

## Segurança e consistência preservadas

- A busca continua limitada à conexão Bling, ao recurso e às páginas posteriores à última página confirmada.
- Os IDs são validados e a ordem do cursor é conferida; resultado fora de ordem causa erro, sem avançar silenciosamente.
- Cancelamento é verificado antes de cada alteração.
- A trava de sincronização é liberada automaticamente quando a transação confirma ou reverte.
- A substituição permanece na mesma transação/savepoint que os dados RAW e o cursor. Falha ou cancelamento continua acionando o rollback do lote pela rotina chamadora.
- Nenhuma tela, interpretação de dado, regra financeira ou regra contábil foi alterada.

## Verificação executada

- `go test ./internal/integrations/bling -count=1` — passou no Windows contra um PostgreSQL 18 descartável com todas as migrações aplicadas.
- A verificação também aplicou a proteção de migração de papéis e `database/privileges/verify_runtime.sql`.
- Testes cobrem 529 IDs em três blocos limitados, erro durante a substituição, cancelamento, rollback/idempotência em PostgreSQL e duas sincronizações concorrentes do mesmo recurso: a segunda só consulta o Bling depois que a primeira confirma a transação.
- `go vet ./internal/integrations/bling` — passou.
- O banco foi iniciado somente em `127.0.0.1:55440`, com credencial temporária protegida. A instância que já usava a porta 5432 não foi acessada. O banco temporário e os caches recriáveis foram removidos ao terminar; `artifacts/cache/go-mod/` foi preservado.
- A CI do commit `6cddbcd` passou em `verify-backlog`, mas parou no verificador do formato da mensagem do commit; Windows e PostgreSQL foram ignorados. Um novo commit com os campos em linhas separadas deve disparar a validação completa.
- Nenhuma conta Bling foi conectada e nenhuma sincronização real ocorreu.

## Limites e próximo passo

O teste de PostgreSQL local confirma migrações, trava e aposentadoria de páginas em transação num banco descartável; a CI ainda precisa repetir a validação após a correção da mensagem de commit. A confirmação do fim da paginação continua pendente: a [OpenAPI oficial atual do Bling](https://developer.bling.com.br/build/assets/openapi-Dw6cY8yQ.json) documenta `pagina`, `limite` e a resposta `data[]`, mas não descreve um marcador de última página. O código segue uma página não vazia e usa `data: []` como marcador provisório; é necessário confirmar isso na conta autorizada antes da homologação.
