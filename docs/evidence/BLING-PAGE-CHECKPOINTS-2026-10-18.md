# Retomada por página e publicação atômica do Bling — 18/10/2026

## Resultado

O Bling agora grava cada página confirmada numa tabela de checkpoint privada. Se o processo parar, a próxima tentativa com a mesma identificação diária continua da página seguinte. As páginas guardadas ainda não aparecem nos relatórios enquanto a carga estiver incompleta.

Quando todas as páginas terminam, o Majucau publica as páginas RAW, encerra o retrato anterior, atualiza o cursor e marca a sincronização como concluída dentro de uma única transação. Se uma página falhar, a transação é revertida: o último retrato válido permanece disponível, a auditoria registra a falha e as páginas confirmadas podem ser reutilizadas na tentativa seguinte.

Os checkpoints de outra execução do mesmo recurso são removidos quando uma carga nova começa. O runtime do worker pode ler, inserir e excluir apenas os próprios checkpoints; a verificação confirmou que ele não recebeu permissão de atualização direta.

Nenhuma tela, regra financeira ou regra de negócio foi alterada.

## Verificações concluídas

- As cinco migrations foram aplicadas em uma instância PostgreSQL 18 temporária e isolada.
- `database/privileges/runtime.sql` e `database/privileges/verify_runtime.sql` passaram; o resultado foi `runtime database privileges: PASS`.
- `go test ./internal/integrations/bling -count=1` passou com PostgreSQL temporário configurado por `MAJUCAU_TEST_DATABASE_URL`. Isso executou o pacote completo, incluindo retomada de páginas após falha, preservação do RAW anterior, limpeza dos checkpoints após sucesso e serialização de duas sincronizações do mesmo recurso.
- A instância temporária e seus dados foram encerrados e removidos após confirmar que o processo PostgreSQL e a porta dedicada estavam livres.

## Limites que permanecem

- Não foi feita chamada real à conta Bling nesta execução.
- A documentação pública consultada não confirma como a conta indica a última página. A confirmação real ainda é necessária para homologar que a leitura diária percorre a lista inteira.
- A leitura do Drive, o parser dos arquivos Excel, a execução contínua do worker e a homologação com dados seguem pendentes.

## Próximo passo

Executar a CI do GitHub para validar a migration, as permissões do worker e os testes no ambiente automatizado. Em seguida, homologar o fim da paginação com a conta Bling autorizada.
