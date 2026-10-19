# Validação estrutural de JSON do Drive — 2026-10-19

## Entrega

- O worker passou a reconhecer arquivos `.json` encontrados somente em `02_nuvem_pago/recebimentos_futuros` e a validar a sintaxe em fluxo, dentro do staging privado existente.
- O leitor não mantém o documento inteiro na memória: limita arquivos a 256 MiB, strings a 1 MiB, profundidade a 64 níveis e até 1 milhão de valores/entradas. Também confere UTF-8, escapes, tamanho anunciado, cancelamento e conteúdo extra depois do documento.
- A validação não escolhe campos, não interpreta valores e não cria fatos financeiros. Como o contrato do JSON não foi aprovado, arquivos válidos, inválidos e acima do limite continuam pendentes. O caminho de JSON não chama o importador financeiro.
- O arquivo temporário é removido após a conferência. A integração com a identidade real do Drive ainda não foi executada.
- Nenhuma tela, regra financeira ou regra de negócio foi alterada. `DATA-04` e BK-042/BK-043 permanecem parciais; esta entrega não conclui uma tarefa do backlog.

## Verificação executada

- `go test ./internal/integrations/jsonstream ./cmd/worker`: aprovado.
- `scripts/verify.ps1`: aprovado por inteiro em 2026-10-19. Incluiu testes Go, limites arquiteturais, `go vet`, `govulncheck`, migrations, geração SQLC determinística, scan de secrets, verificações do instalador, ESLint, TypeScript, 33 testes frontend, build e `npm audit` sem vulnerabilidades.
- Os testes verificam JSON válido e malformado, limites, cancelamento, tamanho anunciado e limpeza do staging. O teste do worker confirma que JSON não chama a importação financeira e permanece pendente.
- `govulncheck` não encontrou chamadas vulneráveis; o alerta de módulo declarado sem chamada vulnerável continua registrado em `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- CI do GitHub ainda depende do push deste commit; nenhum resultado remoto está sendo antecipado.

## O que falta

- Aprovar o contrato dos arquivos JSON antes de qualquer interpretação financeira.
- Criar o parser seguro de XLS, implementar e homologar a leitura de XLSX nas estruturas necessárias, provisionar a identidade técnica de leitura e validar o Drive real.
- Medir memória, disco/WAL e limpeza em uma VM limpa; conectar e homologar Bling/Drive com dados de teste.

## Limpeza desta execução

- Removidos com `scripts/SafeCleanup.psm1` o preview frontend em `artifacts/verification/frontend-dist-ede92ceca11b44288b9d04748e4eef11/` e os caches regeneráveis `artifacts/cache/go-build/`, `go-tmp/`, `npm/` e `vite/`.
- Mantidos `artifacts/cache/go-mod/`, ferramentas locais, dependências instaladas e o pacote oficial em `artifacts/windows/`, pois sustentam as próximas etapas.
- `artifacts/verification/` ficou vazio. Nenhum cache compartilhado fora do projeto foi apagado.
- Nenhum dado real do Drive, segredo ou arquivo externo foi usado; não há novo item para exclusão manual.

## Retificação posterior — escopo das três pastas

O registro original descreve o primeiro caminho implementado: recebimentos futuros. Em seguida, `DRIVE-ALL-SOURCE-STRUCTURAL-STREAMING-2026-10-19.md` ampliou e validou a verificação estrutural em fluxo para as três pastas aprovadas (`01_perdas_operacionais`, `02_nuvem_pago/recebimentos_futuros` e `03_nuvem_envio`), incluindo CSV, TSV, JSON, XLS e XLSX. Portanto, as pendências originais de parser desses formatos foram superadas para validação estrutural. Esse suporte não aprova contratos nem interpretação financeira: arquivos sem contrato continuam pendentes e não viram lançamentos.

