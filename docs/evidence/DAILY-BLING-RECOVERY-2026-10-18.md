# Retomada e proteção contra duplicidade do Bling — 18/10/2026

## O que foi concluído

- A agenda permanece diária às 09:30 no fuso `America/Sao_Paulo` e relê todas as Contas a Receber e a Pagar, sem filtro de data.
- Se o worker iniciar depois das 09:30, ele verifica a carga daquele dia. Se os dois recursos já terminaram, registra que a execução foi pulada; se estiver incompleta, relê a carga integral para retomá-la.
- O identificador estável da data programada fica no campo `correlation_id` das auditorias de lote existentes. Não foi criada tabela ou alterada a estrutura do banco.
- Uma trava consultiva do PostgreSQL serializa o job entre instâncias conectadas ao mesmo banco. A trava é liberada ao terminar e a conexão é fechada se a liberação não puder ser confirmada.
- O Bling continua sendo lido em transação por recurso. Falha de leitura, cancelamento ou paginação não comprovada mantém os dados anteriores e não publica páginas parciais. Recomeçar após falha reconsulta a lista inteira; os hashes RAW existentes impedem criar uma versão repetida quando o conteúdo não mudou.
- A decisão de acesso ao Drive não mudou: o serviço Go será o leitor direto e somente de leitura das três pastas; o Site fica restrito ao fluxo OAuth do Bling. O client GET/streaming existe, mas ainda não foi ligado a este job.

## Verificações desta execução

- `go test ./internal/integrations/bling ./cmd/worker ./internal/scheduler -count=1` — passou no Windows.
- Os testes unitários cobrem a identidade diária estável em horário de Brasília, rejeição de horário incorreto, catch-up após 09:30 e comportamento de execução já concluída.
- O teste de integração PostgreSQL para comprovar que reiniciar não cria nova auditoria diária após sucesso passou na CI #640, em uma instância descartável PostgreSQL 18.6. O job de Windows também passou, incluindo testes, análise estática, verificações de segurança e montagem do pacote de protótipo.
- O commit verificado pela CI é `f85b3579498c8b9aae3a0039e4067c1cb38f6539`; a execução completa está registrada em [GitHub Actions #640](https://github.com/guilhermejacyzin/majucau/actions/runs/39159670619).
- Nenhuma chamada real ao Bling ou Google Drive foi realizada. Nenhuma credencial, informação financeira real ou tela foi acessada.

## O que ainda falta nesta parte

- Comprovar o bloqueio de concorrência com duas instâncias reais contra banco descartável.
- Persistir checkpoint de páginas confirmadas. Atualmente a transação é atômica por recurso e, após interrupção, o job relê todas as páginas desde o começo.
- Provar na conta real como o Bling indica o fim da lista. Até lá, a carga falha com `BLING_API_PAGINATION_UNKNOWN` e não avança como sucesso.
- Integrar as três pastas do Drive e seus estados de pendência, após provisionar a identidade de serviço somente leitura, os IDs das pastas e a credencial protegida.
- Operar o worker de forma contínua e validar a agenda em ambiente de homologação.

BK-039 passa a 35% estimados, ainda parcial. Isso mede o escopo da tarefa, não a prontidão total do protótipo.
