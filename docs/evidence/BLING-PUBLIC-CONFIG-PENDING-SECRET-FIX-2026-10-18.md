# Salvamento dos dados públicos do aplicativo Bling — 18/10/2026

## Causa confirmada

O salvamento de Client ID e Redirect URI estava condicionado à existência do Client Secret. Quando o cofre ainda não tinha o segredo, o worker retornava `BLING_SECRET_REQUIRED` antes de gravar qualquer configuração. Por isso, ao abrir Integrações novamente, os dois campos públicos apareciam vazios. O marcador de bolinhas foi corrigido anteriormente; ele não era o segredo real.

## Correção aplicada

- Client ID e Redirect URI agora podem ser salvos no PostgreSQL mesmo quando o Client Secret ainda não foi cadastrado.
- Nesse caso, o sistema não escreve segredo vazio no cofre nem preenche `secret_ref`; a configuração permanece `NOT_CONFIGURED`.
- Se o Client Secret estiver disponível no cofre, salvar sem digitá-lo continua preservando esse segredo e os tokens existentes.
- Se uma gravação do banco falhar após a troca de segredo, o serviço restaura no cofre o conteúdo anterior. Um teste cobre essa recuperação.
- A tela mantém o mesmo desenho e os mesmos controles. A mensagem informa que os dados públicos foram salvos e que o Client Secret do aplicativo ainda é necessário para conectar.
- OAuth, chamadas financeiras, regras de negócio e dados financeiros não foram alterados.

## Verificações

- `go test -count=1 ./internal/integrations/bling ./cmd/worker`: aprovado.
- Frontend: lint e verificação de tipos aprovados; 33 testes aprovados em 3 arquivos.
- Os testes cobrem salvar metadados sem fabricar segredo e restaurar segredo anterior quando o banco falha.
- Nenhuma credencial real foi lida, copiada, registrada ou enviada ao Git. Nenhuma chamada real ao Bling foi feita.

## O que ainda falta

O Bling continuará aparecendo como **Não conectado** até que o Client Secret do aplicativo seja informado diretamente no Majucau e a usuária execute Conectar. Depois ainda será preciso testar a autorização, conferir a paginação com a conta real e sincronizar os dados. A senha usada para entrar no site do Bling não substitui o Client Secret.

BK-041 continua em 78% e a média global continua em 11,6% concluído / 88,4% restante. Esta correção remove o bloqueio para salvar os campos públicos; não comprova conexão nem sincronização reais.
