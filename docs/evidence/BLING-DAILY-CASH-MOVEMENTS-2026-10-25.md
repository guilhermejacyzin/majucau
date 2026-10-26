# Leitura diária RAW de movimentações de caixa do Bling

**Data:** 25/10/2026  
**Tarefa:** BK-041 — Implementar o adaptador de leitura do Bling  
**Escopo:** incluir `GET /caixas` na carga diária das 09:30, sem normalizar valores financeiros.

## O que foi alterado

- Adicionado `CashMovementsAPISyncService`, que adapta cada página de `/caixas` ao sincronizador paginado RAW já usado pelos recursos Bling.
- As páginas são persistidas com `source_entity=bling.caixas.page`, hash, versão corrente, checkpoint, cursor e lote de sincronização, dentro da transação do recurso.
- A carga diária passa a exigir sucesso em Contas a Receber, Contas a Pagar e Movimentações de Caixa antes de ser considerada concluída. Se algum recurso falhar, o agendamento pode retomá-lo com a mesma correlação idempotente.
- A sincronização manual existente continua sem chamar `/caixas`; telas e resposta pública não foram alteradas.
- Os logs de sucesso incluem totais de páginas e registros, sem copiar payloads.

## Proteções mantidas

- Somente chamadas `GET` com JWT; o retry compartilhado permanece limitado a falhas transitórias.
- O cliente envia `limite=100` seguindo a orientação geral de paginação do Bling. Como a OpenAPI específica de `/caixas` não foi confirmada para esta conta, o adapter não presume que o parâmetro foi aceito: uma página curta só encerra a lista se a resposta confirmar o tamanho ou trouxer metadados suficientes. Sem confirmação, a carga falha fechada.
- Uma resposta sem prova do fim da paginação falha com `BLING_API_PAGINATION_UNKNOWN`. O rollback preserva o snapshot/cursor anterior; dados incompletos não são exibidos como carga confirmada.
- O conteúdo continua JSON bruto. Não há classificação em CAPEX, conta, categoria contábil, entrada ou saída nesta etapa.

## Verificação

- Na implementação inicial de 25/10, o teste verificava que `/caixas` não enviava `limite`; a atualização de 26/10 passou a verificar que a consulta pede explicitamente 100 itens e mantém comportamento conservador sem confirmação da resposta.
- O teste PostgreSQL `TestBlingDailySyncSkipsACompletedScheduledLoad` foi atualizado para exigir os três recursos e precisa passar na CI com PostgreSQL. Essa evidência ainda está pendente da publicação.
- A sessão local do protótipo verificada antes desta implementação informou `NOT_CONFIGURED` para Bling e Client Secret ausente. Nenhuma chamada autenticada à conta real aconteceu; “testes passaram” não significa que a integração real sincronizou.

## Próximos passos

1. Publicar a alteração e confirmar os jobs PostgreSQL e Windows da CI.
2. Salvar Client ID, Redirect URI e Client Secret do aplicativo na tela Integrações do Majucau; o segredo deve ser digitado diretamente no aplicativo e não enviado por chat.
3. Autorizar o Bling e executar teste de conexão e sincronização em uma conta real.
4. Homologar paginação, campos e categorias da conta com evidência sanitizada antes de mapear qualquer valor para CAPEX, DRE, fluxo de caixa ou indicadores.

Nenhuma tela aprovada, regra financeira, regra de negócio ou credencial foi alterada.

## Atualização do limite de página — 26/10/2026

- A orientação geral oficial de paginação do Bling define `pagina` e `limite`, com 100 registros por padrão. O adaptador envia `limite=100` em `GET /caixas`, mas não deduz que o endpoint aplicou o tamanho solicitado.
- Página curta/vazia só encerra a leitura quando a resposta confirma o limite ou fornece metadados suficientes. Sem confirmação, a sincronização falha e preserva o retrato/cursor anterior; não publica uma carga incompleta como completa.
- Testes sintéticos do pacote Bling e `go vet` passaram; a verificação do backlog também passou. Isso valida montagem da consulta e comportamento conservador quando faltam metadados, não a resposta da conta Majucau.
- A aceitação do parâmetro e o fim da lista continuam pendentes de homologação OAuth com a conta real. O pacote não normaliza os lançamentos nem os classifica como CAPEX.
