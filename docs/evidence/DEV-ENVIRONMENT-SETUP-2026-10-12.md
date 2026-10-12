# Preparação do ambiente de desenvolvimento — 2026-10-12

## O que foi preparado

- Go 1.26.9, Node.js 24.14.0 e npm 11.9.0 foram consultados neste computador.
- `scripts/bootstrap.ps1` instala as dependências listadas em `go.mod` e `frontend/package-lock.json`, além das ferramentas Go fixadas pelo projeto: Wails CLI 2.13.0, sqlc 1.31.1 e govulncheck 1.7.0.
- As ferramentas Go ficam em `artifacts/tools/go/bin`. Cache de compilação, módulos, arquivos temporários do Go e cache npm ficam sob `artifacts/cache/`.
- `scripts/verify.ps1` e `scripts/build-windows.ps1` localizam as ferramentas instaladas em `artifacts/tools` antes de procurar instalações globais.
- O arquivo NSIS 3.13 em `artifacts/retained/manual-cleanup/nsis-3.13-alt.zip` foi conferido pelo SHA-256 `ba63dffc4410ee89193e1cb5a41989991bd77c61068da17e3156d136b7b0b3d8`, igual ao valor oficial registrado para essa versão. O conteúdo foi extraído para `artifacts/tools/nsis-3.13`; `Bin/makensis.exe /VERSION` informou `v3.13`. Nenhum instalador foi executado.
- A execução de `scripts/bootstrap.ps1` terminou com sucesso. `npm ci` instalou as dependências do lockfile e reportou zero vulnerabilidades nessa consulta.
- Foi criado `scripts/start-prototype.ps1` para abrir o protótipo usando exclusivamente o banco PostgreSQL 18 local em `artifacts/retained/postgresql-dev/`, limitado a `127.0.0.1:55438`, e a credencial protegida por DPAPI desta conta do Windows.
- O roteiro localiza Wails em `artifacts/tools/go/bin`, compila o worker para `artifacts/cache/dev`, inicia Wails e encerra somente os processos que ele próprio iniciou. O atalho local `artifacts/retained/postgresql-dev/start-prototype.ps1` encaminha para esse roteiro versionado.
- `scripts/start-prototype.ps1` foi executado. O PostgreSQL isolado iniciou em `127.0.0.1:55438`, o Vite ficou disponível e o WebView2 do Wails foi criado. Depois, a execução do Wails terminou; a conferência visual e funcional continua pendente.
- Durante a geração de bindings, o Wails exibiu avisos repetidos `Not found: time.Time`, mas concluiu a compilação e abriu o WebView2. O efeito desses avisos sobre as funções que usam datas ainda precisa ser investigado.
- Na primeira finalização, o roteiro informou erro ao tentar encerrar um worker que já não existia; a consulta confirmou que o processo havia terminado e que o PostgreSQL isolado estava parado. O tratamento de encerramento foi ajustado para verificar o processo novamente; esse ajuste ainda precisa ser conferido em uma nova execução.

## O que ainda falta

- Criar uma configuração de exemplo sem credenciais reais e com dados sintéticos, depois de confirmar que ela é realmente carregada pelo aplicativo. Hoje o worker lê `MAJUCAU_DATABASE_URL` do ambiente; documentar um `.env.example` sem suporte de carregamento não configuraria o aplicativo e poderia confundir a pessoa desenvolvedora.
- Validar a inicialização em uma máquina limpa e confirmar que nenhum dado de produção é usado. Essa etapa depende de fechar a persistência/credencial em BK-018 e a operação local em BK-014.
- Conferir visualmente as telas aprovadas sem alterá-las, verificar as funções principais e confirmar o encerramento seguro. Validar o roteiro em máquina limpa.
- O serviço PostgreSQL `postgresql-x64-18` encontrado no Windows não foi acessado nem alterado por não estar comprovado como exclusivo do Majucau. A pasta do banco de desenvolvimento separado fica registrada localmente em `artifacts/retained/postgresql-dev/`; este documento não afirma que o serviço está ativo nem que a inicialização foi homologada.

## Limites e escopo

- Esta evidência cobre apenas a preparação deste computador. Não comprova ambiente limpo, homologação de instalação ou prontidão de produção.
- Nenhuma tela aprovada, regra de negócio, regra financeira, dado do Bling ou banco desconhecido foi alterado ou consultado.
- Não foram executados testes do produto nesta etapa.
- A tarefa BK-026 segue em andamento, com progresso estimado em 55%. A inicialização local ocorreu, mas a homologação visual/funcional e em máquina limpa continua pendente.
- Os logs gerados nesta execução permanecem somente em `artifacts/retained/postgresql-dev/` para análise do diagnóstico e da inicialização. Eles não foram enviados ao GitHub; a lista de arquivos está no registro local de limpeza.

## Segunda inicialização e correção dos tipos de data — 2026-10-12

- O Wails antes mostrava `Not found: time.Time` ao gerar os vínculos entre o aplicativo e a interface. A ponte agora converte datas para texto ISO/RFC3339Nano, como a interface já esperava; os contratos internos continuam usando `time.Time`.
- O Wails regenerou os vínculos: `checked_at`, `last_success_at`, `last_attempt_at` e `as_of` passaram a ser `string` no TypeScript. A compilação do aplicativo e da interface terminou com sucesso, sem os avisos de `time.Time`; o WebView2 foi criado.
- A inicialização local confirmou respostas de saúde do worker, estado das integrações e consulta do resumo do painel. Usou somente o PostgreSQL separado do Majucau, em `127.0.0.1:55438`. Não houve conexão ao Bling nem consulta de dados financeiros reais.
- A CI do commit `b69e20ce5fd8ce33f4609c0612169077e9456e81` passou nos quatro grupos: `detect-changes`, `verify-backlog`, `verify-postgres` e `verify-windows`. A etapa Windows concluiu os testes Go e frontend e compilou o aplicativo/worker/instalador; a etapa PostgreSQL também passou. Execução: [CI 38304223467](https://github.com/guilhermejacyzin/majucau/actions/runs/38304223467).
- A janela do protótipo permanece aberta para conferência. Por isso, worker, banco isolado e logs continuam ativos/retidos dentro de `artifacts/`; o roteiro fará a limpeza prevista quando a janela for fechada. A conferência visual e funcional ainda não foi feita.
- O worker registrou `DIAGNOSTICS_LOG_UNAVAILABLE`: nesta execução, o arquivo de diagnóstico protegido não foi aberto. O motivo exato não foi exposto no log comum. O worker continuou respondendo e enviou registros técnicos sanitizados para `worker.stderr.log` em `artifacts/retained/postgresql-dev/`. O log do serviço instalado ainda não foi validado.
- O percentual da BK-026 permanece em 55% (45% restante), pois ainda faltam conferir visualmente as telas, validar o encerramento e testar em uma máquina limpa. O backlog geral permanece em 9,1% (90,9% restante), sem prazo total confiável.
- Nenhuma tela aprovada, regra financeira ou regra de negócio foi modificada. Não foram executados testes do produto.
