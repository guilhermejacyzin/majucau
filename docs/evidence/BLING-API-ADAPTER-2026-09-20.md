# Evidência — adaptador Bling/API inicial

Data: 20/09/2026  
Commit: `953137b feat: add read-only Bling API OAuth adapter`

## Entregue

- cliente Bling v3 somente leitura para `GET /contas/receber`;
- paginação explícita por `pagina`/`limite`, com limite de 100 registros;
- filtros enviados apenas quando informados pelo chamador;
- cancelamento por `context.Context` e timeout no cliente padrão;
- classificação sanitizada de autorização negada, limite 429, indisponibilidade,
  resposta inválida e schema não homologado;
- leitura limitada do corpo para impedir payload inesperado;
- OAuth Authorization Code e refresh token com Basic Auth no worker;
- testes `httptest` sem credenciais reais, incluindo verificação de que token,
  client secret e corpo de erro não aparecem nas mensagens;
- sincronização transacional das páginas para `raw_records`, com hash,
  deduplicação e versão corrente, sem normalização prematura;
- documentação atualizada para separar API oficial de CSV de apoio.

## Limite deliberado

O cliente devolve cada item como JSON bruto. Nenhum campo como valor, situação,
data de liquidação, cliente ou categoria foi presumido. O mapeamento financeiro
continua bloqueado até BK-040 receber uma resposta real sanitizada da conta
Bling autorizada e registrar campo, significado, ausência e limites.

## Verificações

```text
go test ./...
go vet ./...
```

Ambas passaram localmente em 20/09/2026. A evidência não representa conexão
com a conta de produção nem homologação funcional dos campos.

## Próximo gate

Com a conta Bling autorizada pelo formulário seguro do aplicativo, executar o
teste de conexão mínimo e guardar somente uma amostra sanitizada no repositório
de evidências. Em seguida, implementar o adaptador de normalização e a carga
RAW/idempotente usando exclusivamente os campos comprovados.
