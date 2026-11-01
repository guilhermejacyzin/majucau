# Validação da sincronização real do Bling — 2026-11-01

## Resultado

- A autorização OAuth está ativa e o teste de conexão somente de leitura passou.
- A listagem de Contas a Receber foi concluída e publicada no banco local como evidência RAW. Ela ainda não alimenta valores financeiros normalizados do painel.
- A listagem de Contas a Pagar foi lida, mas a publicação atômica foi cancelada durante a normalização; nenhum lote parcial foi publicado como atual.
- Uma verificação limitada a um registro, sem exibir ou guardar seus dados, confirmou: detalhe da conta a pagar autorizado, contato do fornecedor autorizado e detalhe da categoria financeira recusado com HTTP 403.

## Causa confirmada

O HTTP 403 significa que a autorização não possui o escopo necessário para consultar aquele recurso. Não significa que a senha expirou. A aplicação precisa de permissão **somente de leitura** para categorias de receitas e despesas; depois de salvar a alteração no cadastro do aplicativo Bling, a autorização anterior será revogada e deverá ser refeita no Majucau.

O navegador bloqueou a leitura automatizada da página de cadastro do aplicativo. Nenhuma tentativa de contornar esse bloqueio foi feita e nenhuma permissão foi alterada.

## Próximo passo e critérios de conclusão

1. Adicionar no cadastro do aplicativo Bling o escopo de leitura de **Categorias — Receitas e Despesas**.
2. Salvar o cadastro e autorizar o aplicativo novamente pelo Majucau.
3. Repetir a sincronização diária completa; validar Contas a Receber, Contas a Pagar e movimentações de caixa antes de considerar o lote completo.
4. Confirmar a publicação atômica e conferir que o painel mostra estados e fontes reais sem transformar dados RAW não homologados em valores contábeis.

Não foram guardados tokens, nomes, IDs, respostas brutas ou valores financeiros neste registro.

## Referências oficiais do Bling

- [Erros comuns: HTTP 403 por escopo insuficiente](https://developer.bling.com.br/erros-comuns)
- [Aplicativos: seleção de escopos e revogação das autorizações ao alterar escopos](https://developer.bling.com.br/aplicativos)

## CI do commit de diagnóstico

Na execução [`CI on: push`](https://github.com/guilhermejacyzin/majucau/actions/runs/41347408143) do commit `a96a634`, a verificação de backlog passou. A etapa `detect-changes` reprovou a mensagem por não conter os dois campos percentuais com os rótulos e formato numérico obrigatórios pelo verificador do repositório; por isso, os trabalhos Windows e PostgreSQL foram pulados. O código e os testes Go locais passaram. O commit `2e1d690` registrou a falha e corrigiu o formato da mensagem. Seu validador local passou, e a [CI 41347755630](https://github.com/guilhermejacyzin/majucau/actions/runs/41347755630) passou em `detect-changes` e `verify-backlog`; Windows e PostgreSQL foram pulados porque a mudança era somente documental. A falha era de governança da mensagem, não uma reprovação dos testes do produto.
