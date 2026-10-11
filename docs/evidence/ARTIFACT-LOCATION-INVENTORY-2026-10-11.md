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
| `docs/planejamento/outputs/` | Nove relatórios oficiais v0–v2 em Markdown/Excel, todos versionados no Git e referenciados pela documentação de planejamento. | Manter no local atual até esclarecer se “todos os outputs em um só lugar” inclui também esses documentos oficiais versionados. Mover exige atualizar os links, o manifesto e o processo de geração. |

Temporários sensíveis de execução são uma exceção necessária: devem permanecer em diretório privado e ter limpeza garantida, conforme `AGENTS.md`; colocá-los em uma pasta de artefatos compartilhada poderia reduzir a proteção dos dados.

## Limites

Esta é uma fotografia do checkout local em 11/10/2026, não uma prova de que todos os processos futuros usarão a pasta central. Nenhuma pasta foi apagada ou movida, nenhum segredo foi lido e nenhum teste de produto foi executado.

**Decisão pendente:** confirmar se os relatórios oficiais em `docs/planejamento/outputs/` também precisam ser fisicamente movidos para a pasta central de artefatos. A mudança pode quebrar links e a geração atual; por isso, não foi feita por inferência.
