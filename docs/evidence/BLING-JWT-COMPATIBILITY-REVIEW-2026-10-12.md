# Revisão de compatibilidade com JWT do Bling

Data da conferência: 12/10/2026.

## Regra publicada pelo Bling

A [página oficial de migração JWT](https://developer.bling.com.br/migracao-jwt)
informa que tokens opacos foram descontinuados, que o header `enable-jwt: 1`
deve ser enviado ao trocar o código OAuth por tokens e ao renovar tokens, e que
ele precisa seguir em todas as chamadas autenticadas junto de
`Authorization: Bearer {token}`. A página também informa que o JWT pode ter de
1.500 a 3.000 caracteres. Na consulta de 12/10/2026, ela ainda dizia que a data
geral de bloqueio dos tokens opacos estava em definição. O aviso mostrado na
conta da Majucau informa 15/10; por prudência, essa data da conta é tratada como
prazo para homologação.

## Conferência no projeto

- `oauth.go` usa uma função comum para a troca do código e a renovação, e nela
  envia `enable-jwt: 1` nas duas operações `POST`.
- `api_client.go` e os dois caminhos de consulta em `api_resources.go` enviam
  `Authorization: Bearer ...` e `enable-jwt: 1` em cada requisição autenticada.
- Os testes já existentes verificam os headers na troca OAuth, renovação,
  consulta financeira e consultas de recursos. Eles foram executados pela CI
  completa `38296818443`, que passou nos jobs Windows e PostgreSQL.
- O token e o segredo ficam no bundle protegido pelo worker via DPAPI. O limite
  da proteção local é 1 MiB, acima dos 3.000 caracteres aproximados descritos
  pelo Bling; não há coluna curta no banco para truncar o token.
- O código não registra tokens ou corpos de resposta em logs.

## Resultado e limite da comprovação

A configuração do cliente está alinhada com os requisitos JWT publicados. Esta
revisão estática não prova que o Bling emitiu um JWT para a conta Majucau nem
substitui uma conexão real. Depois do login manual no Bling, ainda é necessário
executar o teste de conexão e confirmar sucesso sem expor o token. Nenhuma tela
ou regra financeira foi alterada nesta revisão.
