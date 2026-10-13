# Inventário de caches e saídas locais — 11/10/2026

## Resultado da conferência

`artifacts/` é a pasta central aprovada para caches, builds e arquivos locais que precisam ser preservados. O script `scripts/consolidate-artifacts.ps1` copia os executáveis de `build/bin` para `artifacts/windows/` e remove as saídas temporárias conhecidas ao final do empacotamento.

Na conferência deste checkout, `build/bin/`, `frontend/dist/` e `tmp/` não existiam. Não foi necessário removê-los.

## Categorias observadas

| Local | Conteúdo/função | Tratamento desta conferência |
|---|---|---|
| `artifacts/cache/` | Cache local de ferramentas e dependências; inclui junctions temporárias já anotadas no registro local de limpeza. | Preservar ferramentas necessárias. Não apagar nem alterar o destino compartilhado das junctions; seguir o registro local para limpeza manual. |
| `artifacts/cache/sites/majucau-bling-bridge/` | Cópia de trabalho do código-fonte publicado do Site Majucau Bling Bridge, versão 16; necessária para concluir a avaliação e a implementação do leitor XLS/XLSX. | Manter durante esse trabalho; não é saída descartável enquanto a análise do importador estiver em andamento. Registrar para limpeza manual quando a substituição estiver concluída e o código-fonte necessário tiver sido preservado. |
| `artifacts/windows/` | Executáveis e pacote de build conservados para validação manual do protótipo. | Manter por enquanto, como já registrado em `artifacts/retained/cleanup-register.md`; podem ser recriados. |
| `artifacts/retained/postgresql-dev/` | Banco de desenvolvimento isolado, credenciais protegidas por DPAPI e logs locais. | Preservar os dados da instalação. Há processos PostgreSQL ativos no Windows, mas não os associei ao diretório nem os parei. |
| `frontend/node_modules/` | Dependências instaladas do frontend. | Preservar; havia processos Node ativos e não foi possível associá-los com segurança a uma aplicação específica. Não mover nem apagar durante essa execução. |
| `docs/planejamento/artifacts/cache/` | Junction temporária antiga para o runtime compartilhado, registrada para limpeza manual. | Não apagar nem alterar o destino compartilhado; a remoção manual da junction já está documentada no registro local. |
| `artifacts/planejamento/outputs/` | Nove relatórios oficiais v0–v2 em Markdown/Excel, versionados no Git e referenciados pela documentação de planejamento. Caches e prévias continuam em subpastas dedicadas de `artifacts/`. | Consolidado em 13/10/2026, conforme instrução da responsável. Links, manifesto, verificador e gerador foram atualizados; os arquivos continuam versionados. |

Temporários sensíveis de execução são uma exceção necessária: devem permanecer em diretório privado e ter limpeza garantida, conforme `AGENTS.md`; colocá-los em uma pasta de artefatos compartilhada poderia reduzir a proteção dos dados.

## Limites

O inventário começou em 11/10/2026. Na atualização de 13/10/2026, os nove relatórios oficiais foram movidos para a pasta central; nenhum banco, cache ativo ou dado local foi removido. As prévias temporárias e o atalho de dependências criados para gerar/revisar a planilha foram removidos com a rotina segura. O pacote foi conferido localmente: 127 tarefas, 108 decisões, 80 links locais e zero erros de fórmula.

**Registro da decisão — 13/10/2026:** a responsável determinou que todos os caches e outputs gerados sejam consolidados em um único local. Os relatórios oficiais permanecem versionados e foram transferidos para `artifacts/planejamento/outputs/`; caches e prévias temporárias permanecem em `artifacts/cache/` e `artifacts/verification/`. Dados locais e caches ativos não foram removidos nesta alteração.

**Verificação remota — 13/10/2026:** a CI [38381106600](https://github.com/guilhermejacyzin/majucau/actions/runs/38381106600) passou nos quatro grupos depois que o caminho padrão do verificador foi atualizado para a pasta central.

## Limpeza local — 13/10/2026

- Removidos com `scripts/SafeCleanup.psm1` o cache local `artifacts/cache/npm/` (logs temporários e marcador de atualização), a pasta vazia `artifacts/tmp/go/` e o pacote antigo de publicação `artifacts/working/bling-bridge-v0.5.7.tar` (26.300 bytes; SHA-256 `BE8CBBCA0978438A5ABBCE6185954990D4105996FCCCDF19ABBC422F75AE5CB3`). A versão publicada atual da ponte e seu checkout-fonte permanecem registrados e preservados.
- `artifacts/verification/` estava vazio. A limpeza confirmou que os três caminhos removidos não existem mais.
- Mantidos temporariamente enquanto o protótipo está em uso: `frontend/dist/`, `build/bin/majucau-dev.exe` e `artifacts/windows/`. A janela Wails, o aplicativo de desenvolvimento e o worker estavam em execução; apagar a saída embutida ou o executável poderia interromper a sessão de validação. `frontend/dist/` e `build/bin/` serão recriados pelo próximo build e devem ser limpos quando a sessão do protótipo for encerrada.
- Mantidos em `artifacts/cache/` o runtime Go, módulos Go, cache de compilação, cache Vite e runtime WebView2, pois o ambiente local continua ativo e esses itens evitam reinstalação ou recompilação longa. O cache npm foi removido porque nenhum processo npm estava em execução.
- As Junctions `artifacts/tmp/node_modules/`, `artifacts/cache/backlog-sync-20261011-01/node_modules/`, `artifacts/cache/bk017-workbook-tooling/node_modules/` e `artifacts/cache/spreadsheet-tool/node_modules/` continuam apontando para dependências compartilhadas do Codex. Permanecem registradas para remoção manual; não apagar nem alterar o destino compartilhado.
- Não houve alteração no banco, nos dados locais, nas telas aprovadas ou nas regras financeiras. O arquivo pessoal `scripts/smoke-installer.ps1` permaneceu intocado.
