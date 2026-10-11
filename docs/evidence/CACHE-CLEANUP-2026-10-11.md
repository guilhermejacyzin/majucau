# Rechecagem de caches locais — 11/10/2026

Esta conferência verificou apenas os caminhos de cache registrados para o projeto. Não removeu arquivos: os caches antigos já estavam ausentes.

- `frontend/node_modules/.vite`: ausente.
- `frontend/node_modules/.vite-temp`: ausente.
- `artifacts/cache`: existe e está vazia nesta conferência.
- A configuração atual do Vite direciona o cache persistente para `artifacts/cache/vite` (`frontend/vite.config.ts`).
- O registro local de limpeza em `artifacts/retained/cleanup-register.md` foi atualizado. Essa pasta é ignorada pelo Git; o documento preserva os caminhos da máquina para controle local e não deve ser publicado.

Nenhuma tela, regra financeira ou código de processamento foi alterado. Não foram executados testes nem processos de build nesta conferência documental.

## Atualização após validação local — 11/10/2026

- O teste PostgreSQL temporário e seus caches Go ficaram dentro de
  `artifacts/cache`; o servidor foi encerrado e a pasta removida pela rotina
  segura do projeto.
- Uma consulta manual `go list ./...`, usada para contar pacotes, foi executada
  antes de direcionar os caches Go e baixou três versões de dependências para
  o cache global compartilhado. Elas não foram removidas porque outros
  projetos podem depender desse cache; os locais exatos ficaram registrados
  em `artifacts/retained/cleanup-register.md`, ignorado pelo Git, para decisão
  manual.
- Para evitar repetição nos fluxos oficiais do projeto, `bootstrap.ps1`,
  `verify.ps1` e `build-windows.ps1` agora direcionam também `GOTMPDIR` para
  `artifacts/cache/go-tmp` e criam essa pasta antes dos comandos Go.
- A consulta não executou build nem teste; os resultados de integração estão
  registrados em `LOCAL-POSTGRES-INTEGRATION-2026-10-11.md`.
