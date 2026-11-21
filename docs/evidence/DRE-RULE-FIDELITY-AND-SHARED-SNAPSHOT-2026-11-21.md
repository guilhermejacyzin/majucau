# Fidelidade das regras da DRE e resultado compartilhado

**Data:** 21/11/2026
**Branch:** `codex/dre-calculation-memory-20261120`
**Referência consultada:** repositório privado `FinanceiroMajucau/Majucau`, commit `421651ff3ef8ff17b9ce02b7fbb02ff6072a5de3`.

## O que a responsável pediu

A DRE do Intelligence deve seguir as mesmas regras da DRE enviada como referência. A origem pode ser extraída para o banco do computador da Gisele durante o protótipo, e os resultados precisam ser iguais entre todas as telas/acessos. A decisão de arquitetura já registrada exige uma única execução versionada e snapshots compartilhados; múltiplas telas não podem recalcular a DRE por caminhos diferentes.

## Código de referência que foi lido

- [`08_6_tributario.py`](https://github.com/FinanceiroMajucau/Majucau/blob/421651ff3ef8ff17b9ce02b7fbb02ff6072a5de3/projetos/ativos/dashboard-majucau/08_6_tributario.py), blob `421de6428d5289deb643e5ef5f5e0d66a36de297`.
- [`09_dashboard/metrics/dre.py`](https://github.com/FinanceiroMajucau/Majucau/blob/421651ff3ef8ff17b9ce02b7fbb02ff6072a5de3/projetos/ativos/dashboard-majucau/09_dashboard/metrics/dre.py), blob `3feb320013d257d656c05f3569548c10e1e9c2ab`.

Os endereços são links de leitura do repositório conectado. Não foi consultado o banco privado do app de referência nem copiada uma base privada; o código e as regras públicas ao usuário conectado foram lidos no repositório Git.

## Regras do DAS encontradas no código

1. Fonte do pagamento: `Fato_Caixas_Bancos`, categoria DAS `14675501032`, por `data_lancamento`. O mês da competência de DAS é mês do lançamento menos um. O código não usa `competencia` do movimento bancário porque ela pode refletir o mês do pagamento. Assim, pagamento lançado em novembro é da competência de outubro.
2. Taxa histórica: maior competência que tenha DAS positivo pago e faturamento NF positivo. Taxa = DAS pago ÷ faturamento NF da mesma competência. Para esta taxa, exclui pagamentos cuja observação contém “parcelamento”.
3. Faturamento principal: `Fato_NF_Saidas`, situação diferente de `2`, deduplicação por `nf_id`, valor `valor_nota`, mês de `data_emissao`, agrupado pela semana ISO. Se não houver faturamento semanal de NF, o módulo usa `Fato_Vendas` com situação aprovada, deduplicado por `pedido_id`.
4. Guia paga da competência: o total considera a categoria DAS inteira, incluindo pagamentos parcelados, e é alocado entre as semanas proporcionalmente ao faturamento NF semanal. Quando a base semanal total é zero, o código divide igualmente pelas semanas do mês.
5. Sem guia paga: aplica a taxa efetiva mais recente às semanas fechadas usando NF real; se essa semana não tiver NF, usa a projeção semanal de vendas. Semana aberta usa projeção. Semana sem base positiva não gera parcela.
6. Taxa de `8,5%` é fallback do código quando nenhuma competência oferece o par DAS pago + NF correspondente. O Intelligence deve executar a regra somente quando a extração necessária estiver completa; ausência causada por falha de importação permanece pendente e não é tratada como “nenhum par”.
7. As semanas são ISO que intersectam o mês, com a primeira e a última semana truncadas nas fronteiras do mês. O corte é derivado da maior `data_lancamento` em Caixa e Bancos ou, se não houver datas, da data atual. Para orçamento, `modo_orcado=True` usa projeções em todas as semanas e ignora guia paga e NF realizada.
8. O Python da referência calcula em `float` e arredonda cada linha emitida para centavos. O Intelligence usa decimal exato; precisa documentar e testar o limite de arredondamento contra a referência, sem ajustar valores para forçar igualdade.

O método não é uma apuração fiscal oficial. O pagamento real substitui a estimativa na competência correspondente; os dois nunca são somados. O PGDAS da competência serve para reconciliação.

## O que foi implementado nesta alteração

- `internal/financial/dre_tax.go`: resolvedor de DAS com competência M+1 a partir da data de lançamento, amostra histórica sem parcelamento, fallback da taxa, aplicação sobre as bases semanais entregues, rateio proporcional da guia paga e rateio uniforme quando não há base positiva. As entradas exigem cobertura explícita; pagamentos sem competência identificável bloqueiam a estimativa.
- `internal/financial/dre.go`: a DRE recebe uma resolução tributária única por competência e usa esse valor na linha de tributos; estimativa e pago não são somados.
- Migration `000028_dre_tax_resolution_memory.up.sql` e `internal/drememory/repository.go`: persistem o tipo `ACTUAL/ESTIMATED/UNAVAILABLE` e a memória com regra, fontes, taxa, semanas, base, fallback e motivo dentro da própria linha do snapshot. `Load` lê isso do mesmo snapshot imutável.
- Testes unitários focados dos pacotes `internal/financial` e `internal/drememory` passaram em 21/11/2026.

## O que ainda impede dizer que os valores da T06 já ficaram iguais

- A extração e o adaptador que montam pagamentos Caixa e Bancos, NF, fallback de vendas, projeção, IDs e calendário semanal ainda não chamam esse resolvedor.
- As outras linhas DRE ainda não são montadas a partir do conjunto real da referência; não há execução DRE real/snapshot com esses dados.
- A migration `000028` ainda precisa ser aplicada pela rotina oficial ao banco local e validada em PostgreSQL.
- O modo orçamento ainda precisa ser ligado ao perfil correto no cálculo/app.
- Falta conferir os valores semanais e o arredondamento contra o snapshot da referência e comparar setembro/2026, que Gisele informou estar correto.
- Nenhum resultado foi apresentado como corrigido/homologado pela tela. Os acessos só poderão mostrar a mesma DRE depois que todos lerem o mesmo snapshot publicado. O protótipo local sozinho não dá acesso multi-dispositivo.

## Banco local, Drive e várias pessoas

- Durante o protótipo, mantenha PostgreSQL no PC da Gisele e use o migrador/backup oficiais do Majucau.
- Drive pode guardar arquivos de entrada e backups versionados. Não coloque a pasta ativa de dados PostgreSQL em sincronização do Drive: sincronização de arquivo não é coordenação segura de várias escritas no banco.
- Para uso em computadores diferentes, use a hospedagem privada centralizada aprovada em `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md`: um serviço autenticado executa e grava a DRE em PostgreSQL privado; cada tela/usuário lê o snapshot publicado por esse serviço. O PC da Gisele permanece como ponto de administração, não como servidor improvisado que precisa ficar ligado.
- Cada recálculo gera snapshot novo com período, hash de entrada e versão de regras. Cards, tabela, gráfico e drill-down devem ler o mesmo `snapshot_id`; corrigir fonte cria outra execução e preserva a anterior.

## Próximos passos em ordem

1. Aplicar e validar migration 000028 no PostgreSQL descartável/ambiente local permitido.
2. Criar o adapter por fonte com contratos, paginação, lote RAW, hash, contagem e total de controle, preservando IDs originais.
3. Reproduzir o calendário semanal ISO, a data de corte, as semanas fechadas/abertas e o modo orçamento da referência.
4. Integrar a memória de imposto à montagem da DRE e ao snapshot consumido pela T06; conectar detalhe para expor a memória já persistida.
5. Rodar a competência de setembro com os dados originais da referência, comparar linha/semana/documento, corrigir diferenças explicáveis e repetir em outro mês.
6. Validar que cards/tabela/gráfico/detalhes e acesso de diretoria carregam o mesmo `snapshot_id` e valores.
7. Atualizar BK-066/BK-067 somente com evidência do aceite correspondente; abrir a T06 para conferência da responsável antes de declarar homologação.
