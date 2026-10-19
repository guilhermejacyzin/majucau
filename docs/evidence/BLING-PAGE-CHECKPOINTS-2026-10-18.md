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
- A CI do GitHub passou para o commit `951d4a6`, incluindo geração SQL, testes Windows, testes PostgreSQL, migrations, verificações de segurança e builds do desktop, worker e instalador: [execução 39172042551](https://github.com/guilhermejacyzin/majucau/actions/runs/39172042551).

## Limites que permanecem

- Não foi feita chamada real à conta Bling nesta execução.
- A documentação pública consultada não confirma como a conta indica a última página. A confirmação real ainda é necessária para homologar que a leitura diária percorre a lista inteira.
- A leitura do Drive, o parser dos arquivos Excel, a execução contínua do worker e a homologação com dados seguem pendentes.

## Próximo passo

Homologar o fim da paginação com a conta Bling autorizada; a CI desta implementação já passou.

## Correção defensiva de paginação — 18/10/2026

A revisão encontrou uma suposição ainda não comprovada no adapter: uma resposta
vazia sem metadados era tratada como confirmação de que a lista terminou. A
documentação pública não garante isso. O adapter agora continua após páginas
não vazias e, se receber uma página vazia sem marcador explícito, encerra a
tentativa com `BLING_API_PAGINATION_UNKNOWN`. A transação mantém a versão RAW
publicada e o cursor anteriores; o checkpoint confirmado da página anterior
fica disponível para retry. Isso prioriza não publicar um retrato incompleto.

O teste de integração PostgreSQL também cobre esse caso: falha auditada, cursor
inalterado, retrato anterior preservado e apenas páginas confirmadas mantidas
em checkpoint. Os testes unitários, `go vet`, integração PostgreSQL e CI
Windows/PostgreSQL passaram no commit `aff28d3`: [execução 39173420786](https://github.com/guilhermejacyzin/majucau/actions/runs/39173420786). A conta real continua sem homologação.

## Regra atual de término — 18/10/2026

Após rechecagem da OpenAPI oficial, o adapter usa o limite efetivo nas listas
que documentam `pagina` e `limite`: página cheia continua; página curta ou
vazia encerra. Isso permite ler a resposta publicada como `data[]`, sem exigir
metadados que o schema não promete. Respostas maiores que o limite são
rejeitadas. Em rotas sem limite efetivo nem metadados de paginação, o caso
vazio permanece desconhecido e continua falhando sem publicar o retrato ou
avançar o cursor. A implementação foi validada com testes locais; a conexão e
homologação da conta Majucau ainda estão pendentes.

Validação local registrada nesta execução: `go test -count=1 ./internal/integrations/bling`,
conferência do backlog e o build completo Windows passaram. O build também
verificou frontend, segurança, migrations e pacote portátil. O `govulncheck` não
encontrou vulnerabilidades alcançáveis; registrou um alerta numa parte OpenPGP
não utilizada da dependência `x/crypto`. A conta real ainda não foi autorizada:
a senha de entrada do Bling não substitui o Client Secret do aplicativo, que
deve ser cadastrado diretamente no Majucau.
