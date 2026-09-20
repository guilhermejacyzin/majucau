# API Integration Map - Bling, Nuvemshop e Nuvem Pago

- **Status:** Adaptador técnico inicial implementado; validação funcional com credenciais reais ainda pendente
- **Política:** somente leitura no MVP; nenhuma alteração será enviada aos ERPs
- **Criticidade:** Alta
- **Baseline documental:** fontes oficiais consultadas em 2026-08-15; OpenAPI, exemplos sanitizados e respostas reais serão versionados no repositório antes do conector passar pelo gate G2

## 1. Visão geral

| Integração | Papel no produto | Situação técnica |
|---|---|---|
| Bling API v3 | B2B futuro, contas a pagar, realizados, contas financeiras, categorias, contatos, pedidos e NF-e | API pública documentada e suficiente para o primeiro conector |
| Nuvemshop API v1 | Pedidos B2C, clientes, status de pagamento e parcelas | API pública documentada, mas não fornece agenda completa de recebíveis do Nuvem Pago |
| Nuvem Pago | Taxas, líquido, repasses e data esperada de recebimento | API pública para consulta do ledger do lojista não comprovada |
| Folha | Provisões, encargos, pró-labore, funcionários e departamentos | Entrada por arquivo/relatório ainda não especificado |

## 2. Arquitetura do conector

```text
Agendador
  -> rate limiter por provedor
  -> cliente HTTP com timeout
  -> paginação/cursor
  -> lote de sincronização
  -> RAW append-only
  -> normalização idempotente
  -> validação/reconciliação
  -> checkpoint
  -> motor financeiro
```

Cada requisição terá:

- correlation ID interno;
- timeout e cancelamento;
- retry somente para falhas transitórias;
- backoff exponencial com jitter;
- sanitização de logs;
- contagem por rate limit;
- persistência do último cursor somente após commit do lote.

## 3. Bling API v3

### 3.1 Autenticação

- OAuth 2.0 Authorization Code.
- Autorização: `GET https://www.bling.com.br/Api/v3/oauth/authorize` ou URL vigente indicada pelo cadastro do aplicativo.
- Token: `POST https://api.bling.com.br/Api/v3/oauth/token`.
- Token request: `application/x-www-form-urlencoded` com autenticação Basic usando `client_id` e `client_secret`.
- Código de autorização: validade documentada de 1 minuto.
- Access token: validade documentada de 21.600 segundos.
- Refresh token: validade documentada de 30 dias.
- Base da API: `https://api.bling.com.br/Api/v3`.
- A emissão e uso JWT deverão seguir a migração oficial, incluindo `enable-jwt: 1` quando exigido.

O worker renovará o token antes do vencimento, com exclusão mútua para impedir refresh concorrente. Falha permanente muda a conexão para `AUTH_ERROR` e preserva dados anteriores.

Implementação atual: `internal/integrations/bling/oauth.go` executa a troca do
código e a renovação por `refresh_token` somente no worker, usando Basic Auth e
formulário `application/x-www-form-urlencoded`. `internal/integrations/bling/api_client.go`
faz a leitura de `contas/receber`, preserva o JSON bruto e expõe erros
classificados sem devolver corpo de resposta ou segredo em mensagens. Ainda não
há mapeamento de campos financeiros: ele só será ativado depois do BK-040 com
uma resposta real sanitizada da conta autorizada.

`BlingAPIClient.ListPayables` também cobre `GET /contas/pagar`, e o sync RAW
genérico diferencia `bling.contas_receber.page` de `bling.contas_pagar.page`.
Os dois recursos continuam somente leitura e sem normalização financeira.

O início do OAuth desktop está em `internal/integrations/bling/oauth_service.go`:
o worker cria sessão efêmera, valida `state`, escuta somente loopback registrado,
troca o código, executa o teste mínimo de leitura e guarda os tokens no DPAPI.
O botão de sincronização permanece separado e não é acionado por esse fluxo.

### 3.2 Recursos de leitura

| Recurso interno | Endpoint Bling comprovado | Uso e regra |
|---|---|---|
| Contas a receber | `GET /contas/receber` | B2B futuro após classificação Majucau; filtros por situação e datas |
| Detalhe do recebível | `GET /contas/receber/{idContaReceber}` | Drill-down, saldo, vencimento, contato, categoria e origem |
| Recebimentos realizados | `GET /contas/receber` | Filtrar registros pagos/recebidos e data de recebimento; não usar Nuvem para esse card |
| Contas a pagar | `GET /contas/pagar` | Obrigações abertas, vencidas e futuras |
| Detalhe da obrigação | `GET /contas/pagar/{idContaPagar}` | Saldo, vencimento original, categoria, portador e origem |
| Pagamentos realizados | `GET /contas/pagar` | Filtrar pagos e data de pagamento |
| Contas financeiras | `GET /contas-contabeis` | Operação documentada como obtenção de contas financeiras |
| Caixa/banco | `GET /caixas` | Débito/crédito, data, categoria, conta e situação de conciliação |
| Categorias | `GET /categorias/receitas-despesas` | Base para mapeamento financeiro e DRE |
| Clientes e fornecedores | `GET /contatos` | Não há endpoint público separado de fornecedores comprovado |
| Pedidos de venda | `GET /pedidos/vendas` | Origem operacional/faturamento e reconciliação |
| Pedidos de compra | `GET /pedidos/compras` | Compras e fornecedor |
| Notas fiscais | `GET /nfe` | Faturamento fiscal e rastreabilidade |

O conector não utilizará endpoints de baixa ou escrita. O Majucau é consumidor analítico, não sistema de lançamento no MVP.

### 3.3 Paginação e limites

- parâmetros `pagina` e `limite`;
- padrão documentado de 100 registros;
- janelas de filtro superiores a um ano podem retornar HTTP 400;
- limite geral informado: 3 requisições/segundo por conta e 120.000/dia;
- limite adicional de token informado: 20 requisições em 60 segundos;
- HTTP 429: respeitar headers disponíveis, aplicar backoff e não avançar checkpoint.

Usar janelas temporais pequenas e sobrepostas. A sobreposição é segura porque a ingestão é idempotente.

O cliente inicial valida `pagina >= 1`, `1 <= limite <= 100`, envia somente
`GET`, respeita cancelamento do contexto e reconhece `Retry-After` em 429. A
retentativa com backoff e o checkpoint transacional pertencem à camada de
orquestração/sincronização; não são simulados dentro do cliente HTTP.

### 3.4 Lacunas Bling

1. Não existe `business_type = B2B` comprovado. A classificação dependerá de regra configurável e aprovável.
2. Não há endpoint agregado comprovado de saldo financeiro conciliado. A composição deverá usar contas e lançamentos de caixa conciliados.
3. Não há endpoint agregado de faturamento; o indicador será calculado.
4. Webhooks não cobrem todos os objetos financeiros necessários.

### 3.5 Regra B2B proposta para aprovação

A classificação será versionada e conservadora, nesta ordem:

1. override manual auditado por contato/título;
2. título reconciliado com pedido Nuvemshop/Nuvem Pago: `B2C_EXCLUDED_FROM_BLING_FUTURE`;
3. contato com CNPJ válido: `B2B`;
4. canal, categoria ou contato incluído em lista B2B aprovada: `B2B`;
5. qualquer outro caso: `UNCLASSIFIED`, fora do total confirmado e exibido numa fila de revisão.

Correlação Bling ↔ Nuvem: usar primeiro um identificador direto de pedido/origem comprovado no payload. Na ausência, aceitar somente correspondência exata de conexão/loja, número externo, hash do documento do cliente, moeda, valor bruto e data de emissão. Zero, múltiplos candidatos ou qualquer divergência ficam `UNCLASSIFIED`. Matching aproximado/probabilístico é proibido no total confirmado.

Alterar a regra cria nova versão e recalcula snapshots sem sobrescrever os anteriores. Fixtures obrigatórias cobrirão CNPJ, CPF, pedido Nuvemshop reconciliado, ausência de documento, override, recebimento parcial e cancelamento. Essa proposta só se torna regra ativa após aprovação funcional explícita.

## 4. Nuvemshop API

### 4.1 Autenticação

- OAuth 2 restrito, Authorization Code.
- Autorização: `https://www.tiendanube.com/apps/{app_id}/authorize`.
- Token: `POST https://www.tiendanube.com/apps/authorize/token`.
- Entrada: `client_id`, `client_secret` e `code`.
- Saída: `access_token`, `scope` e `user_id`/store ID.
- Código: validade documentada de 5 minutos.
- Token: sem expiração automática documentada; invalida com novo token ou desinstalação.
- Base: `https://api.nuvemshop.com.br/v1/{store_id}`.
- Headers: Bearer token e `User-Agent` identificando aplicativo e contato.
- Escopos propostos: `read_orders`, `read_customers` e `read_payments` quando aprovado.

A documentação não comprova refresh token. Não implementar renovação fictícia.

### 4.2 Recursos de leitura

| Recurso interno | Endpoint comprovado | Uso e limitação |
|---|---|---|
| Pedidos B2C | `GET /orders` | Valor bruto, descontos, total, moeda, gateway, status e datas |
| Pedido | `GET /orders/{order_id}` | Drill-down; usar ID interno, não número visual |
| Clientes | `GET /customers` | Identificação e vínculo com pedido |
| Transações | `GET /orders/{order_id}/transactions` | Documentado como exclusivo para Payment Apps; não assumir disponibilidade |
| Webhooks | `/webhooks` | Eventos de pedido, mas exigem receptor HTTPS público |
| Opções de pagamento | `GET /payment_providers/options` | Identifica provedores/opções, não agenda de repasse |

Campos de pedido úteis:

- `subtotal`;
- `discount`;
- `total`;
- `currency`;
- `gateway`, `gateway_id`, `gateway_name`;
- `payment_status`;
- `payment_details`;
- `paid_at`;
- `total_paid_by_customer`;
- `total_paid_by_customer_including_fees`.

Esses campos não comprovam valor líquido do lojista, taxa descontada ou data de repasse.

### 4.3 Paginação e limites

- `page` inicia em 1;
- `per_page` documentado até 200;
- usar headers `Link` e `x-total-count` quando presentes;
- bucket padrão documentado de 40 requisições;
- reposição documentada de 2 requisições/segundo por loja/aplicativo;
- usar filtros de data e `since_id` para evitar limites de resultados.

## 5. Nuvem Pago

Não foi localizada documentação oficial pública que permita a um aplicativo financeiro externo consultar:

- agenda de recebíveis;
- data prevista de repasse por parcela;
- taxa efetivamente descontada;
- valor líquido do lojista;
- saldo Nuvem Pago;
- antecipações;
- retenções;
- liquidações.

A API de Payment Provider é destinada a aplicativos que implementam um provedor de pagamento. Ela não será tratada como API de extrato do lojista.

### 5.1 Evidência de tarifas recebida

A evidência aprovada está registrada nas capturas [1x D30](docs/source/Nuvem-Pago-Taxas-1x-D30-2026-09-20.png), [2x D30](docs/source/Nuvem-Pago-Taxas-2x-D30-2026-09-20.png), [3x D30](docs/source/Nuvem-Pago-Taxas-3x-D30-2026-09-20.png) e [boleto/PIX](docs/source/Nuvem-Pago-Taxas-Boleto-PIX-2026-09-20.png):

| Meio | Prazo anunciado | Tarifa | TPV |
|---|---:|---:|---|
| Cartão 1x | D+30 | 2,59% + R$ 0,35 | grátis |
| Cartão 2x | D+30 | 4,49% + R$ 0,35 | grátis |
| Cartão 3x | D+30 | 5,44% + R$ 0,35 | grátis |
| Boleto | D+2 | R$ 2,39 | grátis |
| PIX | na hora | 0,99% | grátis |

Todo recebimento de cartão até 3x é configurado em D+30. Boleto e PIX são exceções válidas conforme suas capturas. A evidência resolve a referência de **tarifário configurável**, mas não comprova a existência de um ledger externo consultável. O produto não deve transformar percentual e prazo anunciados em liquidação confirmada sem transação, parcela, status, data efetiva e reconciliação. A tabela anterior com cartão D+2/D+14 foi substituída e não entra no cálculo.

### 5.2 Decisão

- manter conector `nuvempago` separado de `nuvemshop`;
- disponibilizar painel com estado `UNAVAILABLE` ou `PARTIALLY_AVAILABLE`;
- não preencher `fee_amount`, `net_amount` ou `expected_receipt_date` por estimativa silenciosa;
- aceitar futuramente API privada oficial ou arquivo exportado oficialmente;
- implementar contrato de adapter para que a fonte possa ser adicionada sem alterar o domínio financeiro;
- submeter qualquer fallback manual à aprovação funcional.

Enquanto a agenda oficial não existir, pedidos Nuvemshop poderão alimentar apenas informação B2C provisória. O card exigido pelo handoff não poderá ser marcado como confirmado.

## 6. Webhooks em aplicação local

Bling e Nuvemshop documentam webhooks para alguns recursos, porém a aplicação será executada em uma única máquina sem endpoint HTTPS público.

Decisão do MVP:

- não expor porta do computador à internet;
- não usar webhook diretamente;
- usar polling incremental idempotente;
- manter módulo de inbox/webhook preparado, mas desativado;
- avaliar relay hospedado somente se houver justificativa posterior de latência.

Isso reduz superfície de ataque e dependência de infraestrutura externa.

## 7. Tela Configurações > Integrações

### 7.1 Layout funcional

```text
Integrações
├── Bling
│   ├── status e explicação
│   ├── client ID
│   ├── client secret
│   ├── redirect URI
│   ├── Conectar / Testar / Desconectar
│   └── última sincronização e erro
├── Nuvemshop
│   ├── status e explicação
│   ├── app ID
│   ├── client secret
│   ├── redirect URI
│   ├── Conectar / Testar / Desconectar
│   └── store ID, scopes e sincronização
└── Nuvem Pago
    ├── disponibilidade
    ├── explicação da limitação
    ├── fonte autorizada futura
    └── teste somente quando houver endpoint oficial
```

### 7.2 Tooltips mínimos

| Componente | Explicação simples proposta |
|---|---|
| Client ID | “Identifica o aplicativo criado no provedor. Não é sua senha de login.” |
| Client secret | “Chave privada do aplicativo. Ela será protegida nesta máquina e não aparecerá novamente.” |
| Redirect URI | “Endereço para o qual o provedor retorna após você autorizar o acesso.” |
| Scopes | “Permissões de leitura concedidas ao Majucau.” |
| Testar conexão | “Verifica a autorização sem alterar dados no Bling ou na Nuvemshop.” |
| Última sincronização | “Última vez em que os dados foram importados com sucesso.” |
| Dados desatualizados | “A conexão falhou ou ainda não sincronizou. O sistema mantém o último dado válido.” |
| Parcialmente disponível | “A fonte forneceu parte das informações, mas faltam dados necessários para confirmação total.” |

Todos os campos terão label visível; tooltip não substituirá instrução crítica.

### 7.3 Formulário editável e autenticação iniciada pela UI

A tela é a área oficial para preencher e alterar a configuração das APIs. A pessoa usuária poderá digitar os dados públicos e o secret de cada provedor, salvar a configuração, iniciar a conexão, testar, reconectar, desconectar e solicitar uma sincronização. Esse fluxo é editável no frontend e deve parecer uma operação normal de aplicativo desktop.

Isso não significa que o React possa armazenar ou executar a autenticação. O desenho obrigatório é:

```text
form React editável
  -> validação de formato e confirmação
  -> IPC local autenticado
  -> worker valida SID/permissão
  -> DPAPI/ACL para secret e token
  -> navegador externo para OAuth
  -> callback/relay temporário
  -> worker testa e sincroniza
  -> UI recebe estado sanitizado
```

Regras da experiência:

- Client ID/App ID, Redirect URI, Store ID, conta e scopes aparecem como campos editáveis;
- Client Secret aparece mascarado; depois de salvo, a UI mostra apenas “configurado” e permite substituição explícita;
- limpar o campo sem confirmar não apaga o secret existente;
- salvar não dispara sincronização automaticamente;
- “Conectar” abre o navegador externo e não captura login em WebView;
- “Testar conexão” faz uma leitura mínima e explica a causa em caso de falha;
- “Reconectar” permite trocar a conta/loja sem apagar RAW histórico;
- “Desconectar” revoga/remove credenciais e preserva dados já importados;
- nenhuma credencial aparece em URL, localStorage, bundle, log, diagnóstico ou resposta IPC;
- todos os botões, campos, erros e estados têm tooltip simples, label, foco visível e mensagem acessível.

O frontend não cria campos financeiros fictícios para Nuvem Pago. O adapter só habilita seu formulário quando existir contrato oficial para autenticação e ledger; até lá, o card permanece editável quanto à visualização/estado, mas `UNAVAILABLE` para confirmação financeira.

### 7.4 Testes de conexão

- Bling: chamada GET mínima a recurso autorizado, com paginação 1/limite 1.
- Nuvemshop: `GET /orders?page=1&per_page=1` com `read_orders`.
- Nuvem Pago: nenhum teste até existir endpoint e autorização comprovados.

O resultado deverá distinguir:

- DNS/rede;
- timeout;
- credencial inválida;
- escopo ausente;
- rate limit;
- provedor indisponível;
- payload incompatível;
- sucesso com identificação da conta/loja.

## 8. Segurança OAuth no desktop

- abrir autorização no navegador externo;
- usar `state` aleatório e PKCE quando o provedor aceitar;
- nunca usar WebView embutida para capturar login;
- callback em `127.0.0.1`, nunca `localhost`, somente para provedor que o aceite e com redirect registrado;
- validar correspondência exata de provider, state, sessão e redirect URI;
- callback listener temporário e restrito ao loopback;
- para Nuvemshop, usar relay HTTPS mínimo: sessão efêmera, `state`, segredo de pareamento, leitura única e TTL de cinco minutos; o relay transporta somente `code/state`, nunca client secret ou access token;
- exigir confirmação do provedor, redirect exato, teste de replay/cancelamento e PKCE quando o endpoint oficial o suportar; se não suportar, o fluxo permanece bloqueado até threat model e aceite explícito `D-002-RISK`, sem alegar equivalência com PKCE;
- nenhum secret compilado no executável;
- tokens criptografados por DPAPI no worker.

## 9. Critérios de aceite das integrações

### Bling

- autenticação e refresh testados;
- polling respeita limites;
- mesmos dados importados duas vezes não duplicam;
- falha não apaga nem zera o último dado válido;
- contas a pagar/receber reconciliam com tolerância R$ 0,01;
- realizados reconciliam com tolerância R$ 0,01;
- saldo financeiro conciliado D-1 é composto por contas/lançamentos conciliados no mesmo filtro do Bling, fecha com tolerância R$ 0,01 e alimenta o Saldo Inicial D0;
- drill-down alcança o source ID e RAW.

### Nuvemshop

- autorização identifica corretamente a loja;
- pedidos e clientes são idempotentes;
- cancelamentos/reembolsos não são tratados como recebível confirmado;
- valores sem agenda oficial permanecem provisórios;
- drill-down alcança pedido, RAW e loja.
- callback HTTPS intermediário aprovado pelo provedor e testado ponta a ponta; nenhum fluxo manual/clipboard é aceito.

### Nuvem Pago

- não liberar estado confirmado sem fonte oficial;
- qualquer adapter futuro possui teste de contrato, taxas, líquido, parcela e data esperada;
- reconciliação com tolerância R$ 0,01.

## 10. Fontes oficiais

O pipeline arquivará por release: data de consulta, hash do OpenAPI quando disponível, exemplos sanitizados dos recursos usados e fixtures de contrato. Mudança incompatível do provedor coloca a integração em `SCHEMA_MISMATCH`, preserva o último dado válido e bloqueia confirmação de novos números.

### Bling

- [Guia da API](https://developer.bling.com.br/bling-api)
- [Aplicativos e OAuth](https://developer.bling.com.br/aplicativos)
- [Referência OpenAPI](https://developer.bling.com.br/referencia)
- [Limites](https://developer.bling.com.br/limites)
- [Webhooks](https://developer.bling.com.br/webhooks)
- [Migração JWT](https://developer.bling.com.br/migracao-jwt)

### Nuvemshop

- [Autenticação](https://dev.nuvemshop.com.br/docs/applications/authentication)
- [API](https://dev.nuvemshop.com.br/en/docs/developer-tools/nuvemshop-api)
- [Pedidos](https://tiendanube.github.io/api-documentation/v1/resources/order)
- [Clientes](https://tiendanube.github.io/api-documentation/resources/customer)
- [Transações](https://tiendanube.github.io/api-documentation/v1/resources/transaction)
- [Webhooks](https://tiendanube.github.io/api-documentation/v1/resources/webhook)
- [Uso e limites](https://dev.nuvemshop.com.br/docs/developer-tools/erp-guide/api-usage)
- [Payment Provider](https://tiendanube.github.io/api-documentation/guides/payment-provider)
