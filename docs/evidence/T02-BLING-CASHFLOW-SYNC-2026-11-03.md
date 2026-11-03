# T02 — sincronização e falhas encontradas (03/11/2026)

## Pedido

Fazer a tela 02 Fluxo de Caixa apresentar a projeção baseada no mesmo relatório
Fluxo de Caixa do Bling. A decisão registrada é consultar somente APIs oficiais
e reproduzir o relatório com os dados dessas APIs. Telas aprovadas e regras
financeiras permanecem intactas.

## O que foi encontrado e corrigido

1. O Bling pode devolver novamente um conteúdo RAW antigo depois de uma
   alteração. Uma restrição permanente de unicidade tratava essa volta como
   conflito e impedia publicar a atualização diária inteira. A migration 000022
   mantém o histórico imutável, permite que o mesmo hash reapareça como nova
   versão e conserva uma única versão corrente por registro.
2. A consulta de publicação RAW tinha um `ORDER BY` com alias fora do escopo
   após `UNION ALL`. A consulta agora ordena pelos nomes disponíveis no resultado.
   Uma reprodução local dentro de transação confirmou que a gravação funcionou;
   a transação foi revertida e não deixou dado de diagnóstico no banco.

## Sincronização real em andamento

- Contas a Receber: lote normalizado com sucesso, 14.913 registros lidos; os
  registros pessoais e valores não são copiados para este documento.
- Caixas e Bancos: 31.032 movimentos lidos e normalizados, mas a fonte não
  informa a situação de algumas linhas. O lote permanece **Parcial** e a T02 não
  deve tratá-lo como série confirmada.
- Contas a Pagar: em **RAW_ONLY**, etapa de preparação de detalhes. Às 04:15 de
  Brasília havia chegado à página preparada 31 de 102. Enquanto a transação não
  terminar, os dados não são publicados como snapshot utilizável.
- Contas financeiras e condições das formas de pagamento ainda precisam ser
  concluídas nesta mesma atualização diária antes de calcular a projeção com os
  filtros do Bling aprovados (Itaú e Mercado Pago; float e prazos habilitados).

O snapshot consultado às 04:14 informou fluxo de caixa **Indisponível**, série
com 0 dias, 50 linhas de recebíveis e 50 de contas a pagar exibidas no limite
da consulta. Isso confirma que a ligação visual ainda não foi demonstrada com a
execução nova. Não anunciar a T02 como pronta nem como equivalente ao relatório
até uma execução completa e comparação do mesmo intervalo/filtros.

## Limites documentais

A referência pública consultada lista as APIs de contas, caixa e formas de
pagamento; não localizei um endpoint oficial para requisitar diretamente o
relatório pronto de Fluxo de Caixa. A conclusão é limitada à documentação
pública e não prova que nenhuma API privada exista. A opção autorizada é
reproduzir o relatório usando os recursos oficiais publicados e comparar a
saída com o relatório da própria conta.

- [Referência da API Bling](https://developer.bling.com.br/referencia)
- [Como o relatório de futuras entradas e saídas é gerado](https://ajuda.bling.com.br/hc/pt-br/articles/360038306714-Como-gerar-um-relat%C3%B3rio-de-futuras-entradas-e-sa%C3%ADdas-de-valores)
- [Taxas e prazos das formas de pagamento no Bling](https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling)

## Próximo passo obrigatório

Deixar a execução terminar; conferir estado dos cinco recursos e publicação
atômica; abrir T02 no protótipo e comparar o mesmo período e filtros com o
relatório Bling. Se algum campo continuar ausente, conservar o estado Parcial e
registrar o bloqueio específico sem inventar saldo ou status.

## Retomada e correção de token (03/11/2026, 04:45 BRT em diante)

A primeira tentativa terminou com `BLING_API_UNAUTHORIZED` depois de ler 10.198
contas a pagar; ela não chegou a publicar essa fonte como snapshot atual. A
causa técnica encontrada foi que o cliente Bling obtinha o token uma vez no
início da sincronização e não o renovava se uma chamada de leitura recebesse
401 durante um processamento longo.

Foi implementada a renovação reativa: em uma chamada GET segura, um 401 provoca
uma única renovação OAuth protegida, a gravação do token no cofre DPAPI e uma
única repetição daquela leitura. O refresh token é revalidado sob o bloqueio
compartilhado de credenciais para que leituras futuras usem o token novo. Os
testes do conector e da projeção de caixa passaram; `go vet` também passou.
O teste com `-race` não foi executado porque CGO está desabilitado neste
ambiente.

Uma nova leitura da carga diária de 03/11/2026 começou às 04:45 BRT no banco
isolado do protótipo. Ela reutiliza a mesma identificação de carga e as páginas
já preparadas de forma idempotente. O primeiro recurso, Caixas e Bancos,
terminou como parcial pelas linhas sem situação confirmada; Contas a Receber
iniciou em seguida. A carga ainda está em andamento; portanto não há resultado
novo de tela homologado neste registro.

## Como a T02 obtém o resultado do relatório

A referência OpenAPI pública atual do Bling agrupa as APIs de Caixas e Bancos,
Contas a Pagar, Contas a Receber, Contas Financeiras e Formas de Pagamento, mas
não define uma rota de leitura do relatório pronto “Fluxo de Caixa”. A T02 lê
essas fontes oficiais e reproduz as séries para as contas financeiras
selecionadas. Isso é uma composição a partir das APIs; não é uma cópia de um
endpoint de relatório. Campos sem confirmação mantêm o estado Parcial.

Referência conferida: https://developer.bling.com.br/referencia

