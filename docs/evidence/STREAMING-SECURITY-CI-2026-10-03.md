# Auditoria incremental de limites e streaming — 2026-10-03

## Escopo validado

- A resposta JSON financeira do Bling é lida com `json.Decoder` sobre `io.LimitedReader`. O corpo tem teto de 4 MiB; a lista mantém no máximo 100 `json.RawMessage` por página e rejeita página acima desse contrato, truncamento, JSON adicional e bytes além do limite. Os campos RAW por registro continuam preservados para persistência idempotente.
- O cofre Windows DPAPI lê o arquivo de entropia em fluxo até 33 bytes e aceita apenas o formato exato de 32 bytes. Valores armazenados ficam limitados a 1 MiB, coerente com o quadro IPC; blobs cifrados aceitam até 64 KiB adicionais para o envelope DPAPI e falham fechados acima disso.
- A decisão aprovada para backup permanece: manter leitura/restauração V1 e usar V2 em blocos autenticados para novos arquivos, conforme `G0-DECISION-REGISTER.md` e `BACKUP-STREAMING-IMPLEMENTATION-2026-10-03.md`.

## Evidência de CI

- DPAPI: commit `e2157bf7fccef6c60408239eb03a53a4aea69f98`; [run 37108239734](https://github.com/guilhermejacyzin/majucau/actions/runs/37108239734) passou em Windows e PostgreSQL, incluindo testes para entropia, valores e blobs acima dos limites.
- Bling: commit `83952a6c60f1e3b2abf531025b0fedc0a0406f49`; [run 37108913837](https://github.com/guilhermejacyzin/majucau/actions/runs/37108913837) passou em Windows e PostgreSQL. A suíte validou arrays com 101 registros e corpo acima do limite após JSON válido; testes Go, `go vet`, verificação de vulnerabilidades/secrets, frontend, builds, smoke do helper/instalador e consolidação dos outputs também concluíram com sucesso.
- Outputs locais: `artifacts/cache/` e `artifacts/verification/` não existiam. O pipeline centralizou os outputs do runner em `artifacts/`. Como a conexão GitHub não oferece exclusão de artifacts, seus IDs, hashes, expiração e instrução manual constam em `artifacts/retained/cleanup-register.md`.

## Limites desta evidência

- A página da API fica limitada a 100 registros e 4 MiB; não se mantém o histórico inteiro em memória. Os `json.RawMessage` de uma única página ainda ficam retidos juntos para gravar o RAW transacionalmente.
- O DPAPI é uma API de sistema baseada em buffer; o código agora limita estritamente o bloco antes de chamá-la. Ainda faltam validar ACL/SDDL e recuperação de credenciais em instalação Windows limpa.
- Esta auditoria cobre os fluxos citados e não certifica todos os formatos, integrações nem a operação completa. DATA-04 e SEC-01 permanecem `PARTIAL`; resta auditar os demais caminhos e executar os gates em VMs reais.

## Temporários privados — commit `d17b66f`

- A prévia CSV e os fluxos de backup/restauração compartilham agora `internal/securetemp`: no Windows a pasta nasce com DACL protegida e herança restrita aos arquivos; nos demais sistemas recebe permissão de proprietário (`0700`).
- A CI [37116232138](https://github.com/guilhermejacyzin/majucau/actions/runs/37116232138) passou nos jobs Windows e PostgreSQL e executou os testes existentes do projeto.
- A CI valida compilação e regressões dos fluxos que usam o helper, mas não verifica as ACEs efetivas numa VM Windows limpa. A prova de ACL/limpeza operacional continua pendente em VM Windows 10/11; DATA-04 e SEC-01 permanecem `PARTIAL`.
