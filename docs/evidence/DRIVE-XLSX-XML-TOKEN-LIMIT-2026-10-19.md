# Limite de tokens XML na leitura XLSX — 2026-10-19

## O que foi corrigido

O leitor `.xlsx` já percorria abas em fluxo e limitava campos depois de receber tokens do `encoding/xml`. Isso deixava uma lacuna: um único atributo, trecho de texto ou comentário muito grande podia ser acumulado pelo decoder antes de a validação do campo agir.

Foi adicionado um leitor intermediário que reconhece limites de tokens XML durante a leitura e impede que mais de 1 MiB de um mesmo token chegue ao decoder. A checagem cobre texto, tags/atributos, comentários, CDATA, instruções e declarações XML. O leitor guarda no máximo um bloco fixo de entrada, preserva o cancelamento e transforma estouro em erro de limite.

## Proteções comprovadas

- Casos acima do limite para atributo, texto, comentário, CDATA e declaração são recusados como `ErrLimitExceeded`.
- Nenhum desses arquivos produz linhas para o consumidor.
- A pasta temporária privada fica vazia após cada rejeição.
- O limite não altera interpretação de células, datas, fórmulas, cabeçalhos ou valores financeiros.
- Arquivos reais, credenciais e dados da Majucau não foram usados.

## Verificações locais

- `go test -count=1 ./internal/integrations/xlsxstream/...` — aprovado.
- `scripts/verify.ps1` — aprovado: testes Go, limites de arquitetura, `go vet`, `govulncheck`, migrations, SQLC determinístico, scan de secrets, contrato do instalador, ESLint, TypeScript, 33 testes frontend, build e `npm audit`.
- `scripts/verify-backlog.ps1` — aprovado: 127 tarefas, 7 concluídas, 19 em andamento, 101 não iniciadas, 11,7% concluído e 88,3% restante.
- `govulncheck` encontrou zero vulnerabilidades alcançáveis e uma vulnerabilidade em módulo requerido sem chamada alcançável; o alerta existente continua documentado em `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- A CI [#39204918731](https://github.com/guilhermejacyzin/majucau/actions/runs/39204918731), no commit `d777be5`, passou nos quatro trabalhos: `detect-changes`, `verify-backlog`, `verify-windows` e `verify-postgres`.

## O que continua pendente

Esta mudança melhora a segurança estrutural do parser, mas não conecta ou homologa os arquivos do Drive. Ainda faltam identidade técnica de leitura, contratos aprovados para os arquivos, amostras reais sanitizadas e validação de ponta a ponta antes de transformar qualquer linha em dado financeiro. XLSX continua pendente e BK-043 permanece em 40% (60% restante); o projeto permanece em 11,7% (88,3% restante).

Nenhuma tela aprovada, regra financeira ou regra de negócio foi alterada.

## Limpeza desta execução

- Removidos com `scripts/SafeCleanup.psm1` os caches regeneráveis `artifacts/cache/go-build/`, `artifacts/cache/go-tmp/`, `artifacts/cache/npm/` e `artifacts/cache/vite/`, além do bundle temporário `artifacts/verification/frontend-dist-53762993b49f4da8921559505499887e/` criado pela verificação.
- Mantido `artifacts/cache/go-mod/` para as dependências Go compartilhadas entre etapas. Ferramentas locais, dependências frontend necessárias e pacotes oficiais foram preservados.
- Nenhum cache global foi apagado; nenhum item novo depende de exclusão manual.
