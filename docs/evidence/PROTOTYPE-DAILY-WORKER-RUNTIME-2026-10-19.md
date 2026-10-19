# Inicialização local do protótipo e worker diário — 19/10/2026

## Verificação executada

- Iniciei `scripts/start-prototype.ps1` contra o cluster PostgreSQL 18.3 isolado do Majucau, limitado a `127.0.0.1:55438`. O serviço PostgreSQL existente na porta 5432 não foi usado nem alterado.
- O banco isolado iniciou após recuperar uma interrupção anterior e registrou que ficou pronto para aceitar conexões. Nenhum dado financeiro real foi lido ou usado.
- O worker respondeu `state=OK` e confirmou PostgreSQL `OK`. As rotinas Bling e Drive iniciaram com o horário `09:30 America/Sao_Paulo`.
- A interface em desenvolvimento consultou saúde, estado das integrações e painel com respostas bem-sucedidas.
- A integração Bling respondeu `NOT_CONFIGURED`. Não houve OAuth, chamada de leitura do Bling nem sincronização. Nenhuma leitura do Drive foi feita pelo worker.
- A verificação ocorreu às 08:34 de Brasília, antes do horário diário das 09:30; portanto, não comprova que a carga agendada foi executada.
- Em modo de desenvolvimento, o log padrão em ProgramData não existe; os eventos sanitizados foram direcionados aos arquivos locais do protótipo, fora do Git. O aviso de diagnóstico não impediu o worker de responder.

## Resultado e limites

O protótipo atual inicia com um banco Majucau separado e o worker saudável. Isso comprova o arranque local e que a rotina diária ficou armada; não comprova coleta real, conclusão de carga, paginação do Bling, identidade do Drive, reinício em VM limpa ou operação contínua. A carga real continua pendente até a conta Bling ser autorizada e a identidade técnica do Drive ser provisionada.

O protótipo foi deixado em execução para conferência local. Os logs e o cluster permanecem em `artifacts/retained/postgresql-dev/`, registrados no inventário local de limpeza; não apagar enquanto a execução estiver ativa. Ao encerrar o protótipo, `scripts/start-prototype.ps1` encerra o worker e o banco que iniciou.

Não foram alteradas telas aprovadas, regras financeiras, configuração do serviço PostgreSQL existente, arquivos do Drive ou credenciais. Nenhum segredo ou ID de pasta foi incluído nesta evidência.

## Conferência da interface e das janelas — 19/10/2026

- Às 08:43, confirmei duas janelas com o mesmo título. `artifacts/windows/majucau.exe` era o aplicativo antigo, iniciado às 07:53; `build/bin/majucau-dev.exe` era o protótipo atual, iniciado às 08:35. Fechei normalmente somente a janela antiga e deixei o protótipo atual em execução.
- Executei o teste focal existente `informa que o segredo do aplicativo Bling ainda precisa ser configurado`: 1 teste passou. Ele confirma que, sem segredo salvo, a interface atual preenche Client ID e Redirect URI já aprovados e mostra “Informe o segredo do app”. O teste não realiza conexão real com o Bling.
- A captura anterior com campos vazios e pontos mascarados não comprova que o segredo estava salvo; ela foi feita antes da sessão atual do protótipo. Nenhuma tela ou regra financeira foi alterada nesta conferência.
- O status real permanece `NOT_CONFIGURED` até que o Client Secret do aplicativo seja salvo no Majucau e o OAuth seja autorizado.
- Às 08:47, `npm run test` passou com 33 testes em 3 arquivos; `go test ./internal/integrations/bling ./cmd/worker` também passou. Esses testes verificam o comportamento local e a resposta segura quando falta configuração; não autorizam a conta nem comprovam chamadas reais ao Bling.
