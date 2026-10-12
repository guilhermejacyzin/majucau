# Preparação do ambiente de desenvolvimento — 2026-10-12

## O que foi preparado

- Go 1.26.9, Node.js 24.14.0 e npm 11.9.0 foram consultados neste computador.
- `scripts/bootstrap.ps1` instala as dependências listadas em `go.mod` e `frontend/package-lock.json`, além das ferramentas Go fixadas pelo projeto: Wails CLI 2.13.0, sqlc 1.31.1 e govulncheck 1.7.0.
- As ferramentas Go ficam em `artifacts/tools/go/bin`. Cache de compilação, módulos, arquivos temporários do Go e cache npm ficam sob `artifacts/cache/`.
- `scripts/verify.ps1` e `scripts/build-windows.ps1` localizam as ferramentas instaladas em `artifacts/tools` antes de procurar instalações globais.
- O arquivo NSIS 3.13 em `artifacts/retained/manual-cleanup/nsis-3.13-alt.zip` foi conferido pelo SHA-256 `ba63dffc4410ee89193e1cb5a41989991bd77c61068da17e3156d136b7b0b3d8`, igual ao valor oficial registrado para essa versão. O conteúdo foi extraído para `artifacts/tools/nsis-3.13`; `Bin/makensis.exe /VERSION` informou `v3.13`. Nenhum instalador foi executado.
- A execução de `scripts/bootstrap.ps1` terminou com sucesso. `npm ci` instalou as dependências do lockfile e reportou zero vulnerabilidades nessa consulta.

## O que ainda falta

- Criar uma configuração de exemplo sem credenciais reais e com dados sintéticos, depois de confirmar que ela é realmente carregada pelo aplicativo. Hoje o worker lê `MAJUCAU_DATABASE_URL` do ambiente; documentar um `.env.example` sem suporte de carregamento não configuraria o aplicativo e poderia confundir a pessoa desenvolvedora.
- Validar a inicialização em uma máquina limpa e confirmar que nenhum dado de produção é usado. Essa etapa depende de fechar a persistência/credencial em BK-018 e a operação local em BK-014.
- O serviço PostgreSQL `postgresql-x64-18` encontrado no Windows não foi acessado nem alterado por não estar comprovado como exclusivo do Majucau. A pasta do banco de desenvolvimento separado fica registrada localmente em `artifacts/retained/postgresql-dev/`; este documento não afirma que o serviço está ativo nem que a inicialização foi homologada.

## Limites e escopo

- Esta evidência cobre apenas a preparação deste computador. Não comprova ambiente limpo, homologação de instalação ou prontidão de produção.
- Nenhuma tela aprovada, regra de negócio, regra financeira, dado do Bling ou banco desconhecido foi alterado ou consultado.
- Não foram executados testes do produto nesta etapa.
- A tarefa BK-026 segue em andamento, com progresso registrado em 35%.
