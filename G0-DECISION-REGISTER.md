# Registro de decisão G0 — Majucau Financial Intelligence

- **Status:** aprovado pela responsável pelo produto
- **Data da proposta:** 2026-08-15
- **Efeito da aprovação:** autoriza a fundação técnica e somente os módulos não bloqueados; não equivale a declarar o aplicativo pronto para produção
- **Regra:** ausência de fonte oficial nunca será substituída por estimativa silenciosa

## Como aprovar

A aprovação deve registrar uma opção para D-001 a D-005. A resposta curta recomendada é:

> Aprovo o G0 com as recomendações D-001-A, D-002-A, D-003-A, D-004-A e D-005-A.

## D-001 — Classificação de recebíveis B2B no Bling

### Problema

A API pública do Bling não fornece um campo `business_type=B2B`. Sem regra determinística, o total futuro B2B não é reproduzível.

### D-001-A — Regra conservadora versionada (recomendada)

Aplicar em ordem:

1. override manual auditado por contato/título;
2. título reconciliado com pedido Nuvemshop/Nuvem Pago: `B2C_EXCLUDED_FROM_BLING_FUTURE`;
3. contato com CNPJ válido: `B2B`;
4. canal, categoria ou contato incluído em lista B2B aprovada: `B2B`;
5. demais: `UNCLASSIFIED`, fora do total confirmado e presentes em fila de revisão.

A reconciliação do item 2 só é automática quando houver identificador direto de origem. Sem ID direto, exigir igualdade exata de loja/conexão, número externo do pedido, hash do documento do cliente, moeda, valor bruto e data de emissão. Ausência, divergência ou mais de um candidato resulta em `UNCLASSIFIED`; não haverá matching probabilístico. Override manual exige motivo e auditoria.

**Consequência:** minimiza dupla contagem e falso B2B, mas exige revisão dos casos sem classificação.

### D-001-B — Classificar todo Bling como B2B

**Não recomendada.** É simples, porém pode incluir B2C do Bling e violar a regra de composição do handoff.

### Evidência para fechar

- aprovação da opção;
- fixtures cobrindo CNPJ, CPF, origem Nuvem, documento ausente, override, parcial e cancelamento;
- reconciliação com tolerância de R$ 0,01.

## D-002 — OAuth Nuvemshop no desktop Windows

### Problema

O callback loopback de aplicativo desktop não está comprovado na documentação pública consultada. Embutir client secret ou copiar tokens manualmente é inseguro.

### D-002-A — Relay HTTPS mínimo de autorização (recomendada)

- desktop cria `state`, sessão e segredo de pareamento aleatórios;
- navegador usa o redirect HTTPS registrado;
- relay recebe somente `code/state`;
- desktop lê o código uma única vez dentro de cinco minutos;
- worker local troca o código por token usando o secret protegido por DPAPI;
- relay não recebe client secret, access token nem dados financeiros;
- replay, expiração, cancelamento e exclusão de sessão são testados;
- PKCE será usado se o endpoint oficial aceitar. Se não aceitar, D-002-A permanece bloqueada até existir threat model específico e uma decisão separada `D-002-RISK` aceitando que o fluxo não é equivalente a um cliente nativo protegido por PKCE.

**Consequência:** conexão simples para o usuário, mas cria um componente HTTPS mínimo que precisa ser publicado, monitorado e registrado no provedor.

### D-002-B — Adiar Nuvemshop até callback loopback oficial

**Segura, porém bloqueia B2C e a experiência de conexão.**

### Evidência para fechar

- aplicativo real cadastrado;
- redirect exato aceito;
- confirmação/homologação do provedor;
- teste ponta a ponta e teste adversarial de replay.
- confirmação de PKCE; se ausente, threat model e aceite explícito `D-002-RISK` com mitigação, prazo e responsável.

## D-003 — Fórmula do Valor Máximo para Aplicação

### Problema

A planilha `Calculadora_Aplicacao_Lucro_Liquido_Profissional.xlsx` referenciada pelo handoff não foi fornecida. Não existe oráculo verificável para reproduzir.

### D-003-A — Aprovar fórmula Majucau v1 como novo oráculo (recomendada)

```text
eligible_accumulated_profit = lucro líquido acumulado elegível até M-2
available_profit = max(0, eligible_accumulated_profit - already_invested)
worst_week_cash = menor closing_balance diário encontrado em D0..D+60
financial_limit = max(0, worst_week_cash - minimum_reserve)
maximum_investment = max(0, min(available_profit, financial_limit))
```

Somente recebimentos com fonte oficial e data válida participam. O cálculo usa BRL, decimal exato e `ROUND_HALF_UP` em duas casas apenas na fronteira de apresentação/reconciliação.

**Consequência:** cria uma regra nova e auditável, em vez de alegar equivalência com uma planilha inexistente.

### D-003-B — Aguardar a planilha original

**Consequência:** o card permanece desabilitado até a entrega e os testes de equivalência.

### Evidência para fechar

- aprovação da fórmula ou entrega da planilha;
- casos dourados incluindo limite negativo, lucro insuficiente, pior saldo, reserva e recebimento sem fonte oficial;
- diferença máxima de R$ 0,01.

## D-004 — Fonte financeira do Nuvem Pago

### Problema

A documentação pública localizada não comprova API de ledger do lojista com taxa efetiva, líquido, agenda e data prevista de repasse. Pedidos Nuvemshop não provam esses campos.

### D-004-A — Diferir B2C confirmado e manter adapter oficial (recomendada)

- construir contrato do adapter e painel de integração;
- importar pedidos Nuvemshop apenas como operacionais/provisórios;
- mostrar o módulo financeiro Nuvem Pago como `UNAVAILABLE` ou `PARTIALLY_AVAILABLE`;
- não calcular silenciosamente taxa, líquido ou data prevista;
- aceitar futuramente API privada oficial ou arquivo exportado oficialmente, após contract tests;
- manter G2/G3/G7 bloqueados para qualquer release que prometa B2C confirmado.

**Consequência:** permite avançar fundação, Bling, UX, segurança e operação sem fabricar verdade financeira. O objetivo completo continua aberto.

### D-004-B — Retirar B2C confirmado do release inicial

**Consequência:** permite um release de escopo reduzido, mas altera explicitamente o handoff e não conclui o objetivo integral.

### Evidência para fechar o objetivo integral

- documentação/contrato oficial ou exportação oficial real;
- payload/arquivo sanitizado;
- bruto, taxa, líquido, parcela e data esperada;
- reconciliação com tolerância de R$ 0,01.

## D-005 — Defaults financeiros e operacionais

### D-005-A — Aprovar o pacote de defaults (recomendada)

- moeda única do MVP: BRL;
- timezone de negócio: `America/Sao_Paulo`; instantes técnicos em UTC;
- dinheiro: `numeric(19,4)`/decimal exato, nunca `float`;
- arredondamento: `ROUND_HALF_UP`, duas casas somente na fronteira da regra;
- relatórios de data de negócio: início e fim inclusivos;
- aging mutuamente exclusivo: vencido, hoje, 1–7, 8–15, 16–30, 31–45 e 46–60 dias; usar somente saldo aberto, excluindo cancelados/estornados;
- intervalos técnicos: `[início, fim)`;
- sincronização padrão: 15 minutos;
- `STALE`: duas execuções perdidas ou 30 minutos sem sucesso;
- pagamento parcial preserva saldo/alocação;
- estorno, chargeback e cancelamento geram evento compensatório, sem apagar o fato anterior;
- folha: CSV UTF-8 delimitado por `;`, conforme o dicionário de dados;
- RAW append-only para operação, com exceção exclusiva de redação LGPD autorizada e auditável.

### D-005-B — Fornecer defaults alternativos

Qualquer mudança deve informar moeda, timezone, escala, arredondamento, frequência, limiar de stale, política de eventos compensatórios, folha e tratamento LGPD.

## Matriz de autorização após a decisão

| Decisão | Fundação | Bling | Nuvemshop operacional | B2C financeiro | Aplicação | Produção integral |
|---|---:|---:|---:|---:|---:|---:|
| D-001-A | libera | libera após fixtures | não afeta | reduz dupla contagem | não afeta | necessária |
| D-002-A | libera contrato | não afeta | condicionada ao provedor | condicionada | não afeta | necessária |
| D-003-A | libera contrato | não afeta | não afeta | não afeta | libera após casos dourados | necessária |
| D-004-A | libera | libera | libera provisório | mantém bloqueado | exclui B2C sem fonte | integral continua bloqueada |
| D-005-A | libera | libera implementação | libera implementação | não supre fonte | libera regra técnica | necessária |

## Registro de aprovação

| Campo | Valor |
|---|---|
| Responsável | Gisele |
| Data/hora | 2026-08-16, America/Sao_Paulo |
| D-001 | D-001-A — regra conservadora versionada |
| D-002 | D-002-A — relay HTTPS mínimo, condicionado à homologação e ao tratamento de PKCE |
| D-003 | D-003-A — fórmula Majucau v1 como novo oráculo |
| D-004 | D-004-A — B2C confirmado diferido; adapter oficial mantido |
| D-005 | D-005-A — defaults financeiros e operacionais aprovados |
| Observações | Aprovação explícita recebida no task Codex. Libera G1 e módulos não bloqueados; não libera B2C confirmado nem produção integral sem as evidências externas. |
