# Evidência — reconexão e desconexão segura do Bling — 2026-09-21

## Escopo

Este incremento fecha o comportamento local dos botões **Reconectar** e
**Desconectar** da integração Bling na tela de Integrações. Ele não substitui
a homologação de credenciais, payloads oficiais ou a sincronização produtiva.

## Comportamento implementado

- **Reconectar** reutiliza o fluxo OAuth já exposto pelo worker, sem enviar
  segredo ou token para o frontend.
- **Desconectar** revoga localmente a autorização: remove access token,
  refresh token, tipo, escopos e expirações do bundle protegido por DPAPI.
- O `Client ID`, `Redirect URI` e a referência do segredo são preservados para
  permitir nova autorização sem redigitação estrutural.
- O banco marca a conexão como `NOT_CONFIGURED`, limpa a identidade externa,
  escopos e datas de autorização, registra a tentativa e preserva as tabelas
  de RAW, recebíveis, obrigações, recebimentos e pagamentos já importados.
- Sessões OAuth pendentes são encerradas e a UI informa explicitamente que os
  dados importados foram preservados.
- Enquanto a desconexão está em andamento, testar/sincronizar fica bloqueado;
  falhas retornam código público estável sem expor detalhes técnicos.
- A integração Nuvemshop continua sem OAuth habilitado até que o relay e os
  escopos oficiais sejam homologados.

## Evidência executada

- `go test . ./cmd/... ./database/... ./internal/...`: PASS.
- `go vet . ./cmd/... ./database/... ./internal/...`: PASS.
- `frontend/npm run typecheck`: PASS.
- `frontend/npm run lint`: PASS.
- `frontend/npm run test:run`: PASS — 3 arquivos, 23 testes.
- `frontend/npm run build`: PASS — bundle Vite produzido.
- `git diff --check`: PASS.

Os checks do frontend exigiram execução elevada neste ambiente porque o
esbuild percorre o diretório pai do workspace durante a resolução do projeto;
isso é uma limitação do sandbox local, não uma falha de assertion. A execução
foi repetida com essa permissão explícita.

## Limitações e próximo gate

Esta evidência não prova OAuth real, reconciliação, payload oficial do Bling,
ACL/Named Pipe em máquina limpa, backup/restore, assinatura ou instalação em
Windows 10/11 workstation. O próximo gate é validar credenciais e fixtures
sanitizadas do Bling e executar a matriz de instalação/atualização em VM limpa.

