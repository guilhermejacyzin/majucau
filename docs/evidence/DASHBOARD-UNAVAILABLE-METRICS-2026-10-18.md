# Visão Executiva sem valores não confirmados — 18/10/2026

## Objetivo

Proteger a Visão Executiva contra exibir `R$ 0,00` quando a fonte ainda não
forneceu um valor confirmado. O leiaute, os rótulos e as regras financeiras
permanecem iguais; a tela mostra o traço previsto quando a métrica está
indisponível.

## Verificação local

- `go test ./internal/dashboard -run TestReadStreamSummaryKeepsEmptySourcesUnavailableInsteadOfZero -count=1` — PASS. Com todas as fontes vazias, o resumo mantém as métricas como `UNAVAILABLE`, sem valor/contagem e sem data de sincronização Bling.
- `go vet ./internal/dashboard` — PASS.
- `npm test -- src/App.test.tsx` — PASS; 1 arquivo e 20 testes. Um cenário injeta `0.0000` acompanhado de estado `UNAVAILABLE` e confirma que o cartão “A RECEBER” exibe `—`, não `R$ 0,00`; também confirma que não aparece horário de consulta como sincronização.
- A CI do GitHub [39169340120](https://github.com/guilhermejacyzin/majucau/actions/runs/39169340120), no commit `d9a9f1a`, passou nos quatro jobs: mensagens, backlog, Windows e PostgreSQL.

## Limites

Os testes usam dados artificiais. Eles comprovam a apresentação segura para um
snapshot sem confirmação, mas **não** comprovam que a conta real Bling já
autorizou, sincronizou ou alimentou o banco local. A causa do painel vazio no
protótipo ainda precisa ser observada na execução autenticada. BK-102 permanece
em 1% porque a ligação funcional das métricas e os detalhes aprovados continuam
pendentes.

Nenhuma tela aprovada, regra financeira, dado persistido ou credencial foi
alterado.
