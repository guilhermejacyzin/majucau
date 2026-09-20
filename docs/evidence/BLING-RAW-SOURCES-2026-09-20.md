# Evidência — fontes RAW Bling — 2026-09-20

## Escopo

O conector Bling agora cobre os dois recursos financeiros primários em leitura, sem adivinhar campos financeiros:

- `GET /Api/v3/contas/receber` — recebíveis e recebimentos realizados conforme filtros oficiais;
- `GET /Api/v3/contas/pagar` — obrigações e pagamentos realizados conforme filtros oficiais.

Ambos usam o mesmo contrato de paginação, limite máximo de 100, timeout, cancelamento, classificação de HTTP 401/429/5xx e preservação de `json.RawMessage`.

## Persistência

O serviço RAW genérico grava cada página em `raw_records` dentro de lote transacional, com:

- `source_entity = bling.contas_receber.page` ou `bling.contas_pagar.page`;
- `source_id = pagina:N`;
- hash SHA-256 do payload;
- nova versão quando o hash muda;
- idempotência quando a mesma página volta sem mudança;
- nenhum mapeamento de valor, status, B2B/B2C ou data antes do payload real homologado.

O conector não usa Nuvem para recebimentos realizados e não soma B2C do Bling ao futuro B2C, preservando as regras D-001-A/D-004-A.

## Evidência automatizada

- Teste do endpoint de contas a pagar e dos filtros `dataPagamentoInicial`/`dataPagamentoFinal`.
- Testes existentes de contas a receber, paginação, 401, 429, indisponibilidade, shape incompatível e corpo excessivo.
- Testes do serviço RAW de recebíveis preservados após a generalização.
- `go test ./internal/integrations/bling ./internal/ipc ./cmd/worker`: aprovado.

## Limite

Este marco não habilita ainda o botão de sincronização nem cria fatos normalizados. A próxima etapa é orquestrar o lote explícito pelo worker, com refresh do token, cursor/checkpoint e relatório sanitizado; a reconciliação financeira continua bloqueada até uma resposta real do Bling ser homologada.
