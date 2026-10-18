# Agendamento diário e recuperação do Bling — 18/10/2026

## O que foi concluído

- A agenda permanece diária às 09:30 no fuso `America/Sao_Paulo` e relê todas as Contas a Receber e a Pagar, sem filtro de data.
- Se o worker iniciar depois das 09:30, ele verifica a carga daquele dia. Se os dois recursos já terminaram, registra que a execução foi pulada; se estiver incompleta, relê a carga integral para retomá-la.
- O identificador estável da data programada fica no campo `correlation_id` das auditorias de lote existentes.
- Uma trava consultiva do PostgreSQL serializa o job entre instâncias conectadas ao mesmo banco. A trava é liberada ao terminar e a conexão é fechada se a liberação não puder ser confirmada.
- Cada página confirmada é salva em checkpoint privado. Após uma interrupção, o mesmo lote retoma pela página seguinte, sem publicar o conteúdo parcial.
- Ao terminar a paginação, páginas RAW, aposentadoria do retrato anterior, cursor e auditoria de sucesso são confirmados juntos. Falha ou cancelamento reverte a publicação, mantém o retrato anterior e preserva a auditoria de falha; páginas confirmadas ficam disponíveis para retry.
- A decisão de acesso ao Drive não mudou: o serviço Go será o leitor direto e somente de leitura das três pastas; o Site fica restrito ao fluxo OAuth do Bling. O client GET/streaming existe, mas ainda não foi ligado a este job.

## Verificações desta execução

- `go test ./internal/integrations/bling ./cmd/worker ./internal/scheduler -count=1` — passou no Windows.
- Os testes unitários cobrem a identidade diária estável em horário de Brasília, rejeição de horário incorreto, catch-up após 09:30 e comportamento de execução já concluída.
- As cinco migrations passaram em PostgreSQL 18 temporário. `runtime.sql` e `verify_runtime.sql` confirmaram o acesso mínimo necessário para os checkpoints.
- `go test ./internal/integrations/bling -count=1` passou no Windows com PostgreSQL temporário; o pacote completo comprovou retry da página interrompida, publicação completa, preservação do retrato anterior e concorrência serializada.
- A instância de teste foi encerrada e removida depois de confirmar que a porta dedicada ficou livre.
- A CI do GitHub passou para o commit `951d4a6`, incluindo `verify-backlog`, `verify-postgres` e `verify-windows`: [execução 39172042551](https://github.com/guilhermejacyzin/majucau/actions/runs/39172042551).
- Nenhuma chamada real ao Bling ou Google Drive foi realizada. Nenhuma credencial, informação financeira real ou tela foi acessada.

## Proteção adicional de término de paginação — 18/10/2026

A revisão de segurança identificou que uma resposta vazia sem metadados não é
um marcador de término garantido pela documentação pública. Agora, se isso
ocorrer, o lote recebe `BLING_API_PAGINATION_UNKNOWN`, sem publicar dados nem
avançar o cursor. A página confirmada anterior pode ser retomada após corrigir
ou homologar o marcador. Esta mudança mantém a atomicidade do snapshot. Os
testes e a CI desta correção ainda estão pendentes.

## O que ainda falta nesta parte

- Provar na conta real como o Bling indica o fim da lista. Até lá, a carga falha com `BLING_API_PAGINATION_UNKNOWN` e não avança como sucesso.
- Integrar as três pastas do Drive e seus estados de pendência, após provisionar a identidade de serviço somente leitura, os IDs das pastas e a credencial protegida.
- Operar o worker de forma contínua e validar a agenda em ambiente de homologação.

BK-039 passa a 50% estimados, ainda parcial. Isso mede o escopo da tarefa, não a prontidão total do protótipo. Evidência detalhada: `docs/evidence/BLING-PAGE-CHECKPOINTS-2026-10-18.md`.
