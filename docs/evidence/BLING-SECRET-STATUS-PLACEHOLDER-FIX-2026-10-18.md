# Indicação correta do Client Secret do Bling — 18/10/2026

## Problema encontrado

O campo Client Secret da integração Bling usava bolinhas como texto fixo de exemplo. As bolinhas apareciam mesmo sem segredo no cofre do Windows, parecendo que o aplicativo estava configurado. O serviço local já fornecia o indicador seguro `secret_configured`, mas a tela não o consultava.

O diagnóstico local registrado em 13/10/2026 retornou `BLING_SECRET_REQUIRED`. Isso significa que o Client Secret do aplicativo ainda precisava ser cadastrado no Majucau. A senha usada para entrar no site Bling é uma credencial diferente.

## Correção

- A tela agora usa o resultado verdadeiro de `secret_configured`, tanto ao abrir Integrações quanto após salvar.
- Sem segredo cadastrado, o campo existente mostra “Informe o segredo do app”. Com segredo confirmado no cofre, mostra o marcador visual anterior, sem preencher nem revelar o valor secreto.
- A ajuda do campo explica que o Client Secret é do aplicativo cadastrado no Bling e não é a senha de acesso ao site.
- O Client Secret continua fora do estado React, da resposta pública e dos registros; o campo é limpo ao concluir cada tentativa de salvamento.
- Nenhuma posição, dimensão, estrutura da tela, regra financeira ou regra de negócio foi alterada.

## Validação

- Testes do frontend: 32 aprovados em 3 arquivos, incluindo os estados de segredo ausente, salvo e salvo durante a sessão atual.
- A primeira execução apontou apenas uma seleção ambígua no teste novo, porque existem dois campos chamados Client Secret; o seletor foi corrigido e toda a suíte passou na execução seguinte.
- A compilação completa do Windows foi concluída; testes Go, `go vet`, verificação de vulnerabilidades, lint, checagem de tipos e 32 testes do frontend passaram. O pacote portátil atualizado foi gerado em `artifacts/windows/`.
- `git diff --check` passou.

## O que ainda falta

Esta correção impede que o exemplo visual esconda a configuração ausente; ela não cadastra credenciais e não conecta a conta. O próximo passo é digitar o Client Secret do aplicativo diretamente no campo do Majucau e salvar. Depois, executar **Conectar**, **Testar conexão** e, após confirmar o fim da paginação, sincronizar. Não enviar o segredo pelo chat.

BK-041 permanece em 78%: o adaptador ainda depende da configuração e homologação com a conta real, e os dados RAW ainda precisam alimentar os relatórios. A estimativa global segue em 11,6% concluído e 88,4% restante (127 tarefas); esta correção não altera o progresso medido.

## Correção adicional do salvamento — 18/10/2026

Depois de remover o marcador enganoso, foi confirmado por que Client ID e Redirect URI voltavam vazios: o worker recusava gravar os metadados públicos quando o Client Secret ainda não existia. Agora os dados públicos podem ser salvos sozinhos, sem criar segredo nem habilitar OAuth. A correção está detalhada em `BLING-PUBLIC-CONFIG-PENDING-SECRET-FIX-2026-10-18.md`. A conexão real ainda exige o Client Secret do aplicativo, digitado diretamente no Majucau.
