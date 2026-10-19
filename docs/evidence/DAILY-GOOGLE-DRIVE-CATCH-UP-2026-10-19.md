# Recuperação da atualização diária do Drive após reinício — 19/10/2026

## Problema

BK-039 prevê recuperar trabalhos agendados quando o worker volta após suspensão ou desligamento. O worker já fazia esse catch-up para o Bling, mas a rotina diária do Drive começava esperando até às 09:30 do dia seguinte. Assim, um desligamento do worker na hora programada deixava a leitura do Drive atrasada por mais um dia.

## Correção

- Quando o worker inicia antes das 09:30, a rotina do Drive aguarda o horário normal.
- Quando inicia às 09:30 ou depois, recupera a execução prevista para aquele mesmo dia e registra a data/horário agendados, não o horário tardio de inicialização.
- A identidade da execução usa a data, pasta aprovada e chave determinística já existente. Retomar no mesmo dia não cria uma nova identidade de carga.
- O filtro de importação continua comparando os arquivos com o corte aprovado das 09:30. Um arquivo alterado depois desse corte permanece pendente para a execução do próximo dia, inclusive durante catch-up.
- A rotina normal continua esperando a próxima execução diária; não foi criada uma repetição fora do horário aprovado.

Nenhuma tela, regra financeira, conteúdo do Drive ou permissão de conta foi alterada.

## Verificações

- `go test -count=1 ./cmd/worker ./internal/scheduler`: passou após a implementação final.
- `scripts/verify.ps1`: passou nesta execução; incluiu a suíte Go, `go vet`, `govulncheck`, limites de arquitetura, checksums de migração, geração SQL determinística, scan de segredos, verificações do instalador, ESLint, TypeScript, 33 testes de interface, build e `npm audit` sem vulnerabilidades.
- `govulncheck` não encontrou vulnerabilidades alcançáveis no código; reportou uma vulnerabilidade em módulo requerido que o código não parece chamar, já acompanhada em `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- A CI do GitHub para o commit desta correção ainda precisa ser executada após o push.
- Não foi feita sincronização real do Drive: a conta técnica não foi provisionada. A execução real com a pasta, o serviço Windows persistente e a homologação continuam pendentes.

## Próximo bloqueio de BK-039

O instalador NSIS atual copia `majucau-worker.exe`, mas não registra nem inicia o worker como serviço Windows. A execução diária não pode ser garantida depois que o processo local termina ou o Windows reinicia. Corrigir essa lacuna requer provisionar a identidade do serviço, suas ACLs/DPAPI e o acesso seguro ao PostgreSQL conforme `AGENTS.md` e ADR-001; validar em VM limpa antes de declarar BK-039 concluída. A confirmação de identidade do Google Cloud também é necessária para provisionar a conta técnica somente de leitura e efetuar a primeira leitura real.
