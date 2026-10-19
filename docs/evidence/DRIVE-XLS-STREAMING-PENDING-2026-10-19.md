# Validação estrutural de XLS no Drive — 2026-10-19

## O que foi implementado

- O worker reconhece `.xls` na pasta autorizada `02_nuvem_pago/recebimentos_futuros`, confere o tamanho e baixa o arquivo para o staging privado já existente.
- O leitor percorre planilhas BIFF8 (Excel 97 ou posterior) linha a linha, sem guardar todas as linhas em memória. Não abre nem executa macros.
- O limite de 256 colunas é aplicado dentro do leitor antes da criação da linha. Isso evita que um arquivo malformado solicite uma alocação enorme antes da validação do Majucau.
- Limites adicionais: 64 MiB por arquivo; 256 abas e 65.536 entradas em cada tabela de estilos; 65.536 linhas por aba; 1.048.576 linhas por arquivo; 256 células por linha; 8.388.608 células por arquivo; 256 KiB por conteúdo de célula; 4 MiB por linha; 256 bytes por nome de aba. Os limites de abas/estilos são aplicados durante a abertura. O leitor também verifica cancelamento.
- A implementação usa o trecho de leitura BIFF8 do projeto `github.com/jloor/go-excel-reader` v0.3.0, sob licença MIT preservada em `legacyxls/LICENSE`. O código necessário foi incorporado ao projeto e ajustado para recusar colunas fora do limite antes da alocação; a tabela de strings compartilhadas também limita a capacidade pela quantidade possível de bytes de entrada. Isso mantém o Go mínimo do projeto em 1.26.0, sem nova dependência externa.
- O callback descarta cada linha lida. Planilhas válidas, inválidas, acima do tamanho ou fora dos limites ficam pendentes; nenhuma célula é interpretada como dado financeiro e o importador de valores não é chamado. O arquivo temporário é removido após a validação.

## Verificações

- Fixtures pequenas e sintéticas do projeto de origem: um arquivo simples e um com três abas; não contêm dados da empresa ou de clientes.
- Testes cobrem leitura e resumo de linhas/abas, arquivo truncado ou inválido, arquivo acima do limite, cancelamento, erro do consumidor, cabeçalho CFB com quantidade forjada de setores e rejeição de células/ranges largos antes da alocação.
- O teste do worker cobre `.xls` válido, arquivo inválido e tamanho acima do limite; confirma que os três permanecem pendentes, não chamam o importador financeiro e deixam o staging vazio.
- Testes focados passaram: `go test -count=1 ./internal/integrations/xlsstream/... ./cmd/worker`.
- `scripts/verify.ps1` passou integralmente em Windows em 2026-10-19: testes Go, limites de arquitetura, `go vet`, `govulncheck`, migrations, geração SQLC determinística, scan de secrets, contrato do instalador, ESLint, TypeScript, 33 testes frontend, build e `npm audit`.
- `govulncheck` não encontrou chamadas vulneráveis. Um alerta em módulo requerido sem chamada alcançável continua registrado na evidência `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- O commit `deba395ccb787e028bce7bef99c7af8b7469575e` foi enviado à branch `codex/majucau-artifact-consolidation`. Após o push, a conexão GitHub disponível retornou `statuses: []` e `workflow_runs: []`; portanto, o estado da CI remota não foi confirmado e não é declarado como aprovado.

## O que continua pendente

- Contratos aprovados para as abas, cabeçalhos e significados dos arquivos antes de qualquer interpretação financeira.
- XLSX e os outros formatos nas três raízes autorizadas, leitura real do Drive com identidade técnica e validação em VM Windows limpa.
- Medição de memória/carga com arquivos representativos e homologação do fluxo completo.

Nenhuma tela aprovada ou regra financeira foi alterada. `DATA-04` e BK-043 permanecem parciais; BK-043 continua em 40% (60% restante) porque leitura estrutural não equivale a contrato nem a interpretação financeira.

## Limpeza desta execução

- Depois da verificação, `scripts/SafeCleanup.psm1` removeu os caches regeneráveis `artifacts/cache/go-build/`, `go-tmp/`, `npm/` e `vite/`, além da prévia `artifacts/verification/frontend-dist-*`; confirmei que esses itens não permaneceram.
- Mantidos `artifacts/cache/go-mod/`, ferramentas locais, dependências instaladas do frontend e artefatos oficiais necessários para continuar o desenvolvimento. Nenhum cache global do computador foi removido.
- Nenhum arquivo real do Drive, segredo ou dado financeiro foi usado. Nenhum item novo requer exclusão manual.
