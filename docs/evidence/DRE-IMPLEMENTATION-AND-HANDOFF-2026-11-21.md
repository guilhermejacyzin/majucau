# DRE: estado, regra tributária e passagem de trabalho

**Data da revisão:** 21/11/2026

**Repositório:** `guilhermejacyzin/majucau`, branch `codex/dre-calculation-memory-20261120`

**Objetivo:** permitir que uma pessoa ou modelo sem contexto anterior entenda a regra, o que está efetivamente implementado, o que falta e como continuar sem criar fatos financeiros.

## 1. Resumo em linguagem simples

A DRE própria deve ser calculada no Intelligence por competência. O Bling fornece registros de origem e serve como referência de conferência; o relatório pronto do Bling não é o resultado final do Majucau. Setembro de 2026 é a competência-base que Gisele disse estar correta e deve ser usada para comparar a reprodução. O seletor da T06 deve permitir qualquer mês; setembro não é fixo nem padrão obrigatório.

Para tributos: até o registro do valor real como pago, a tela pode mostrar uma estimativa calculada com a regra semanal estudada no motor de referência. A palavra **Estimado** tem que ficar visível. Quando a fonte passar a mostrar o pagamento real ligado à competência, o valor real substitui a estimativa daquela competência. Nunca somar estimado + pago. O mês do pagamento não é, sozinho, a competência: a regra operacional já confirmada é que pagamento em novembro normalmente se refere à competência de outubro. Todo valor precisa abrir uma memória que explique fontes, fórmulas, taxa, base, períodos e diferenças.

**Estado real deste checkout:** a base aritmética genérica da DRE, o armazenamento/consulta de passos aritméticos e o seletor de competência estão no código. A leitura de vínculos de origem foi acrescentada à memória. Ainda não há no worker um pipeline DRE completo que importe, classifique, estime tributos, gere execução financeira real e publique os valores. Portanto, este commit melhora a memória e documenta a regra, mas não faz a tela preencher números financeiros reais. Não declarar a T06 homologada.

## 2. Regra tributária aprovada e método de referência

### 2.1 Identificação do valor exibido

Cada competência deve guardar e exibir um único valor corrente de tributo, com tipo e estado:

| Situação na origem da competência | Valor corrente | Identificação na tela | Tratamento |
| --- | --- | --- | --- |
| Há pagamento real ligado à competência | Valor real efetivamente pago | Pago/real | Substitui a estimativa. Mantém fontes e memória. |
| Ainda não há pagamento real, mas as fontes necessárias estão completas | Resultado da estimativa documentada abaixo | Estimado | Provisório, nunca “pago” ou “confirmado”. |
| Não há pagamento e faltam vendas/documentos necessários | Indisponível, com motivo | Sem base suficiente | Não estimar com zero nem esconder fonte faltante. |

Pode haver diferença entre estimado e realizado. Exibir o delta (real − estimado) para conferência, mas não somá-lo novamente na DRE. Se a fonte for corrigida depois, gerar nova execução versionada; não alterar o snapshot anterior.

### 2.2 Sequência de cálculo a reproduzir

A regra seguinte descreve o comportamento da projeção tributária semanal identificado no código de referência do projeto `FinanceiroMajucau/Majucau`, commit `3b67a0418d930a69b8875e58a419ed20398e3a6c`, especialmente `08_6_tributario.py`, `09_dashboard/metrics/dre.py` e `data/ancoras/mapa_contabil.json`. A regra é espelho de um método de projeção do painel; não é uma declaração de que a estimativa equivale ao imposto legalmente apurado.

1. Ler pagamentos DAS e vendas/Notas Fiscais de saída com IDs, competência, datas, valores e estado originais. Não classificar pelo nome do fornecedor ou pelo texto livre se houver ID/categoria/documento disponível.
2. Associar pagamento à competência a que o tributo pertence. O deslocamento pagamento M+1 → competência M é regra operacional já aprovada. Se houver vínculo explícito mais confiável no documento/API, preservar esse vínculo e registrar qual regra escolheu a competência.
3. Encontrar a taxa histórica válida no último período com DAS efetivamente pago (excluindo registros cujo histórico/identificação é de parcelamento da taxa-base) e receita de NF da mesma competência. Taxa efetiva = valor DAS elegível ÷ receita NF elegível. Gravar os IDs dos dois lados do pareamento.
4. Se a competência selecionada já tem DAS pago, usar o total realizado daquela competência no mês. Para distribuir na visão semanal, ratear o total pago proporcionalmente às vendas de NF de cada semana, conforme o motor de referência. Não mostrar esse total como estimativa.
5. Se a competência ainda não tem DAS pago, aplicar a última taxa efetiva válida às vendas de NF reais nas semanas encerradas e às vendas projetadas nas semanas abertas. Se não houver NF real para uma semana, só usar projeção quando a fonte de projeção existir e sua cobertura estiver identificada.
6. A taxa de **8,5%** é o último fallback do motor de referência quando não existe par válido de recolhimento DAS pago e receita NF para derivar a taxa. Exibir explicitamente “fallback 8,5%”, os dados ausentes que levaram ao fallback e o valor que ele alterou. Nunca usar o fallback para disfarçar ausência de vendas ou de cobertura.
7. Arredondar apenas no limite definido pelo código de referência e registrar a política e precisão da taxa. Comparar centavos com o oráculo antes de homologar.

**Parcelamento:** o motor de referência exclui parcelas da amostra que deriva a taxa efetiva; pagamentos da categoria DAS continuam compondo o realizado conforme a regra categorial do próprio motor. Cada parcela precisa permanecer identificada e sua contribuição visível na memória. Não a remover nem contar duas vezes por texto parecido; usar IDs, vínculos e reconciliação.

### 2.3 Limite contábil e de segurança

O usuário autorizou a estimativa como aproximação temporária até existir o valor pago. Isso não libera o sistema para chamar estimativa de realizado, nem substitui a validação da contabilidade. A visão espelho pode reproduzir a referência com status `PROVISIONAL/ESTIMATED`; a visão ajustada não pode herdar a estimativa como fato contábil sem a aprovação exigida. Registros bancários de pagamento não determinam por si sós a competência, a natureza ou o resultado.

## 3. O que foi conferido e o que significa

### Evidência no código do Intelligence

- `internal/financial/dre.go`: cálculo decimal determinístico das linhas já classificadas. Propaga indisponibilidade, não converte campo ausente em zero. Não importa nem classifica fatos.
- `internal/drememory/repository.go`: grava e lê snapshot, linhas, passos, operandos e estados; a alteração desta tarefa acrescenta consulta `dre_line_sources` e vínculos a `dre_source_rows`, eventos de estoque, custos de venda e ajustes.
- `database/privileges/runtime.sql` e `database/privileges/verify_runtime.sql`: esta revisão acrescenta SELECT somente nas colunas necessárias para a memória de origem. Não libera leitura geral das tabelas financeiras nem INSERT do pipeline ainda não implementado. Depois de aplicar migration e ACL, executar o verificador como administrador e guardar o resultado sanitizado.
- `database/migrations/000025_dre_calculation_lineage.up.sql`: estrutura para fontes, custos, eventos, ajustes e contribuições imutáveis. `000026` finaliza execução e `000027` preserva IDs de categoria nos títulos de receber/pagar. As migrations 000025–000027 ainda precisam ser aplicadas e verificadas no banco local da implantação.
- `frontend/src/bootstrap.ts` e `frontend/src/mockupDashboard.tsx`: a T06 aceita e exibe as referências de origem ligadas às linhas. Respostas antigas sem o campo opcional `sources` continuam válidas; nesse caso a memória informa que a execução aritmética ainda não apresenta documentos de origem.
- A memória é leitura de uma execução persistida; ela não busca, importa nem calcula dados faltantes por conta própria.

### Limites comprovados nesta revisão

- Não foi consultado banco de dados privado nem credencial do aplicativo de referência.
- O inventário local `docs/evidence/DRE-INPUT-INVENTORY-2026-11-15.md` registra candidatos para setembro, duplicidades e lacunas. Não declara esses arquivos conciliados com a API.
- O caminho `01_extracao.py` do projeto de referência carrega tabelas normalizadas de origem e grava cache local/GCS para o dashboard; isso não é a API do Intelligence e não prova uma conexão ao banco privado do outro desenvolvedor. A população do dado de origem no projeto de referência fica fora deste repositório.
- A conta real do Bling, OAuth, cobertura completa de APIs, importação dos documentos fiscais/financeiros, classificação DRE, execução com dados reais, e confronto de setembro não foram realizados nesta tarefa.
- O cálculo de tributo estimado descrito acima ainda não foi programado nem conectado ao worker. Não haverá número correto na tela somente por atualizar a documentação ou o leitor de memória.

## 4. Conteúdo deste incremento

1. A memória DRE agora consulta contribuições de origem associadas a cada linha salva e devolve IDs, referência documental, competência, categoria, estado, valor original, inclusão/exclusão, regra e contribuição.
2. A tela de detalhe apresenta a tabela de lançamentos/documentos da linha e informa quando há memória aritmética sem vínculos de origem.
3. O frontend continua compatível com respostas anteriores que ainda não trazem `sources`; ausência desse campo não invalida a tela inteira.
4. A regra tributária de “estimado até constar pago; o real substitui” e o algoritmo de referência foram registrados no manual e no registro de decisões.

Isso melhora auditabilidade do que já tiver sido persistido, mas **não cria as contribuições**. O writer de fontes ainda precisa ser ligado à execução DRE.

## 5. Próximos passos obrigatórios, nesta ordem

O próximo responsável deve seguir esta sequência. Não pule diretamente para marcar a T06 concluída.

### Fase A — reabrir e estabilizar o ambiente

1. Abrir o projeto `C:\Users\Gisele\Documents\Codex\2026-10-02\api-2\majucau`, branch indicada no início deste documento; conferir `git status` antes de editar.
2. Ler primeiro `AGENTS.md`, este handoff, `docs/source/Manual_DRE_Intelligence_Majucau_Dev.md`, `G0-DECISION-REGISTER.md`, `docs/DRE-ACCOUNTING-POLICY-v1.md`, `docs/evidence/DRE-INPUT-INVENTORY-2026-11-15.md` e a auditoria fim a fim de 2026-11-02.
3. Confirmar sem ler/mostrar segredo se o PostgreSQL local e as credenciais OAuth necessárias estão configurados. Se não estiverem, explicar o que está ausente; nunca inventar os dados nem marcar integração como conectada.
4. Aplicar migrations de forma forward-only, seguindo os scripts suportados, e registrar checksum, versão e resultado. Não alterar o banco fora do migrador.

### Fase B — tornar a fonte reproduzível

5. Criar adaptador/importador DRE dedicado usando os recursos oficiais Bling já existentes, relatórios enviados e contratos autorizados. Persistir resposta RAW/lote, hash, período, paginação e totais de controle. Separar dados faltantes de zero confirmado.
6. Reconciliar Contas a Receber, Contas a Pagar, Caixa/Bancos, categorias, vendas, notas de saída/entrada, pagamentos e PGDAS. Não usar “Fluxo de Caixa” como substituto de fonte DRE; não somar título e baixa duas vezes.
7. Carregar o mapa versionado de 76 categorias e os motores sintéticos de custo conforme manual. ID sem classificação, categoria pai/filha duplicada, fonte conflitante ou documento sem competência devem virar pendência rastreável.
8. Implementar inferência de competência tributária por vínculo comprovado e regra M+1, com exceções documentadas quando documento/API indicar outro período. Não inferir a competência só da data do banco.

### Fase C — estimar e substituir sem duplicar

9. Implementar a taxa de referência documentada na seção 2.2 usando `domain.Money`, nunca `float`; declarar versão da regra, precisão e arredondamento.
10. Produzir uma única linha tributária por competência com `value_kind = ESTIMATED | ACTUAL | UNAVAILABLE` (ou estado equivalente). A seleção do pago precisa acontecer antes da composição final da linha: se existir pagamento real elegível para aquela competência, descartar o estimado corrente e usar o real. Preservar o estimado calculado em memória para o delta, sem somá-lo.
11. Gravar como fontes/contribuições a guia/pagamentos usados, NF/receita semanal de base, taxa aplicada, origem da projeção semanal e motivo para o fallback. Documentar parcelas e fazer rateio semanal da mesma forma que o motor de referência.
12. Ligar o writer à mesma execução/snapshot que alimenta tabela, gráfico e detalhe; persistir tudo atomicamente e manter snapshots anteriores imutáveis. Nenhum card deve fazer cálculo paralelo.

### Fase D — conferir com setembro e outros meses

13. Gerar manifesto de setembro de 2026: nomes e hashes dos arquivos/lotes, período, linhas, IDs, totais, vendas NF, tributos pagos, parcelas e totais por semana.
14. Rodar o espelho de setembro e comparar cada categoria/semana com a DRE pronta que Gisele indicou como correta. Mostrar lado a lado: origem, total usado, diferença, fórmula e causa provável. Não alterar a fórmula para forçar igualdade.
15. Conferir tributação: taxas históricas, competência do DAS, parcela excluída da taxa, valor mensal realizado, distribuição por semana, saldo estimado, diferença quando o real substitui a estimativa e qualquer uso do fallback 8,5%.
16. Repetir o mesmo processo para pelo menos uma competência diferente, escolhida pela usuária, para provar que o seletor muda todas as linhas, cards, gráficos, memória e estado sem reaproveitar setembro.
17. Solicitar a aprovação de Gisele depois de apresentar as evidências. Até lá, manter `PROVISIONAL`, `PARTIAL` ou `UNAVAILABLE`, nunca `CONFIRMED` por conveniência.

### Fase E — operação, documentação e entrega

18. Atualizar o manual, G0, backlog, inventário e `docs/META-PROJETO.md` com o que foi realmente executado, evidência, divergências e próximo responsável. Evitar percentual de progresso sem critério verificável.
19. Rodar verificações exigidas no `AGENTS.md` conforme ambiente disponível e registrar comandos, resultado, itens pulados e motivo. Não tratar build/mocks como validação financeira.
20. Abrir a T06 atualizada para aprovação, conferir competência e detalhe com captura/inspeção visual; publicar commit e push somente da branch de trabalho, preservando histórico.

## 6. Critérios mínimos para dizer “corrigido”

- Ao selecionar setembro, a consulta mostra o snapshot de setembro ou explica a indisponibilidade; nunca exibe dados de outra competência.
- Cada total pode ser refeito somando as contribuições incluídas; exclusões mostram motivo; passos aritméticos recalculam a linha.
- O imposto aparece como **Estimado** antes do pagamento real e como **Pago/real** após o pagamento registrado; na transição, uma estimativa sai do valor corrente e não é somada ao pago.
- A competência do pagamento de novembro referente a outubro aparece em outubro, com o elo documental e regra identificáveis.
- Parcelamento não altera taxa histórica de DAS e segue visível quando fizer parte do realizado da categoria.
- Receita-base, taxa, fallback e rateio semanal podem ser inspecionados em centavos.
- O valor de setembro bate com a base de comparação ou cada diferença está listada e explicada. “Bateu o total” não substitui conferência por lançamento.
- Um segundo mês demonstra que seletor, cartões, tabela, gráfico e detalhes acompanham a competência selecionada.
- Nada é marcado confirmado sem cobertura, conciliação, aprovação e execução real.

## 7. Como retomar sem experiência técnica

1. Abra a pasta e a branch indicadas no cabeçalho.
2. Leia a seção 1 para saber a regra em linguagem simples.
3. Use a seção 3 antes de afirmar que um número ou integração já existe.
4. Execute somente o primeiro passo ainda não concluído da seção 5; guarde a saída sem segredos.
5. Se encontrar divergência entre código, tela e este documento, não escolha uma versão por conta própria. Pare a parte financeira, registre os dois comportamentos e peça a decisão de Gisele.
6. Ao terminar qualquer passo, atualize a tabela “estado real” deste documento e os documentos relacionados antes de declarar a tarefa finalizada.

## 8. Verificações executadas neste commit

Foi executado `./scripts/verify.ps1` em 21/11/2026. Resultado: `go test` de todos os pacotes passou; limites de camadas Go passaram; `go vet` passou; `govulncheck` não identificou chamadas vulneráveis alcançáveis, mas registrou uma vulnerabilidade em módulo requerido sem chamada alcançável; checksum de migrations, geração SQLC determinística, busca local de segredos e validação estrutural do instalador passaram; ESLint, TypeScript, 92 testes em 7 arquivos, build Vite e `npm audit` (0 vulnerabilidades) passaram.

A primeira execução da suíte mostrou um teste T03 dependente da data real do computador; o cenário usava data-base 15/11/2026, mas não a fixava no formulário. O teste foi corrigido para selecionar explicitamente 15/11/2026 e a suíte completa passou em seguida. Isso estabiliza o cenário do teste; não altera o cálculo financeiro.

O script de verificação não executou consultas contra um PostgreSQL de teste real porque `MAJUCAU_TEST_DATABASE_URL` não está configurada. Portanto, a sintaxe dos grants SQL foi revisada no diff e a estrutura migração foi checada pelo lint/checksum, mas a consulta nova e `verify_runtime.sql` ainda precisam ser aplicados/executados por administrador em uma instalação local depois da migration. Não houve OAuth, chamada Bling, consulta a dado privado, comparação real de setembro ou inspeção visual de valores financeiros.

`./scripts/verify-backlog.ps1` confirmou o estado atual oficial: 126 tarefas; 7 concluídas, 23 em andamento e 96 não iniciadas; média simples de **12,3% concluído / 87,7% restante**. No escopo DRE, BK-066 está em 5% (**95% restante**) e BK-067 em 25% (**75% restante**). Esses percentuais foram mantidos: consulta de memória e documentação não equivalem ao pipeline financeiro nem ao aceite dos valores.
