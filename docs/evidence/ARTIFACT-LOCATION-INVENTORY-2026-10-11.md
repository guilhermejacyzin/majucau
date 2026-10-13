# Inventário de caches e saídas locais — 11/10/2026

## Resultado da conferência

`artifacts/` é a pasta central aprovada para caches, builds e arquivos locais que precisam ser preservados. O script `scripts/consolidate-artifacts.ps1` copia os executáveis de `build/bin` para `artifacts/windows/` e remove as saídas temporárias conhecidas ao final do empacotamento.

Na conferência deste checkout, `build/bin/`, `frontend/dist/` e `tmp/` não existiam. Não foi necessário removê-los.

## Categorias observadas

| Local | Conteúdo/função | Tratamento desta conferência |
|---|---|---|
| `artifacts/cache/` | Cache local de ferramentas e dependências; inclui junctions temporárias já anotadas no registro local de limpeza. | Preservar ferramentas necessárias. Não apagar nem alterar o destino compartilhado das junctions; seguir o registro local para limpeza manual. |
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
