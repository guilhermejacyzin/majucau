# Auditoria do caminho dos dados do Bling até as telas

**Data:** 02/11/2026  
**Escopo:** integração diária, dados financeiros normalizados e T01 (Visão Executiva).  
**Limites preservados:** nenhuma tela aprovada ou regra financeira foi alterada nesta auditoria.

## Resposta curta

Na verificação ao vivo das 20:06, o Bling continua conectado. `/caixas` terminou `SUCCESS`/`NORMALIZED`; Contas a Receber terminou `SUCCESS`/`RAW_ONLY` com 14.901 registros lidos; Contas a Pagar ainda não tem lote normalizado publicado. O último lote de pagar publicado falhou por rejeitar um saldo negativo, mas o normalizador atual já aceita esse sinal e os testes focados passaram. A tentativa atual permanece dentro de uma transação sem publicação visível; só um lote concluído provará a correção com dados reais.

A T01 foi conferida no protótipo e continua exibindo `—` e “Sem dados confirmados”. Isso ocorre porque receber ainda está em dados brutos, pagar ainda não publicou retrato normalizado e os movimentos de caixa têm `situacao` vazia. Essa situação não pode ser interpretada como realizado sem aprovação; portanto esses movimentos não entram nos totais realizados. Não existe saldo atual oficial por banco confirmado, e a leitura do Drive falhou porque as pastas ainda não estão configuradas no worker.

O primeiro diagnóstico, feito às 19:29 antes da execução completar, está preservado abaixo como histórico. Ele dizia que `/caixas` vinha depois de Contas a Pagar; essa ordem foi alterada para publicar primeiro a fonte de caixa e foi confirmada no lote real acima.

## Caminho esperado

1. O worker guarda as páginas recebidas do Bling em staging protegido.
2. Valida paginação e campos; atualizações incompletas não substituem o último retrato válido.
3. Converte registros aceitos para fatos financeiros tipados, com origem e lote rastreáveis.
4. O banco publica o lote completo ou parcial conforme a regra aprovada.
5. As telas de origem e T01 leem os mesmos fatos locais. T01 não faz uma consulta financeira independente como alternativa.

## O que foi confirmado

| Fonte/etapa | Evidência observada | Efeito na tela |
|---|---|---|
| Autorização Bling | Integração `CONNECTED`; último teste de conexão `SUCCESS`. | A autenticação está funcionando; isso, sozinho, não confirma carga financeira. |
| Contas a Receber | Lote mais recente terminou `SUCCESS`, 14.901 linhas lidas, estágio `RAW_ONLY`. | Não alimenta o total nem as linhas da tela porque ainda não foi convertido para a tabela `receivables`. |
| Contas a Pagar | O último lote publicado falhou com `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`. O parser atual converte saldo em aberto para saída negativa e os testes passaram. A tentativa posterior ainda não publicou lote e mantém uma transação aberta. | Sem lote novo confirmado na tabela `payables`; a tela permanece sem total atual. |
| Movimentos de caixa | `/caixas` terminou `SUCCESS`/`NORMALIZED`; a paginação e o sinal foram confirmados tecnicamente. `situacao` veio vazia e foi preservada sem interpretação. | Os movimentos não entram no realizado até a responsável confirmar se os itens dessa lista podem ser considerados concluídos. |
| Google Drive | O worker registrou `GOOGLE_DRIVE_FOLDERS_NOT_CONFIGURED`. | Os arquivos externos não podem ser lidos até a configuração das três pastas autorizadas; o sistema deixa a fonte pendente. |
| Saldo por banco | A tabela `bling_cash_movements` guarda movimentos com data e direção, não um saldo atual validado de cada conta. | Não é seguro somar movimentos e chamar o resultado de saldo do Itaú ou Mercado Pago. |

## Por que a API pode estar conectada e a tela vazia

O painel não lê linhas brutas: ele lê tabelas normalizadas e snapshots completos. Na leitura feita para esta auditoria, a consulta de receber chegou ao Bling, mas nenhum registro foi publicado em `receivables`. Assim, o painel não tem uma base confirmada para mostrar.

Em Contas a Pagar, a API lista 10.198 títulos. O registro de lista observado contém identificadores, situação, valor original, vencimento, contato e forma de pagamento; o normalizador solicita o detalhe de cada título para obter campos que não estão nessa lista, incluindo o saldo aberto. Ele também pode consultar contato e categoria. A etapa é sequencial e leva mais tempo que ler as páginas. A nova tentativa ainda não havia concluído quando esta auditoria terminou; não há estimativa confiável de término até vermos o primeiro lote finalizado.

## O que cada fonte pode sustentar

- **Contas a Pagar:** títulos e saldos em aberto. No Majucau, o saldo em aberto é negativo por representar saída. O saldo do banco não muda de sinal por causa disso.
- **Fluxo de caixa (`/caixas`):** movimentações efetivamente registradas, com entrada/saída, data e conta informada. Ajuda nos realizados mensais; não substitui a lista de obrigações futuras.
- **Saldo atual por banco:** exige um saldo oficial identificado por conta e data. Movimento histórico, sozinho, não prova o saldo atual.
- **DRE:** continua sendo montada pelo Majucau conforme a regra contábil aprovada. Fluxo de caixa mostra quando o dinheiro entrou/saiu; a DRE depende também de competência, classificação, estoque, custos de produção e imobilizado. Não se pode transformar todo pagamento em despesa nem toda compra em CPV/CMV.
- **CAPEX:** pagamentos e investimentos devem permanecer separados do lucro. Para calcular os cards aprovados, é necessário que o lançamento seja reconhecido como CAPEX com categoria confirmada e que o período de pagamento seja conhecido.
- **Arquivos externos:** não entram até as pastas do Drive estarem configuradas e a fonte chegar. Falta de arquivo deve continuar marcada como pendente, sem reutilizar um valor antigo como se fosse atual.

## Próxima ordem de trabalho

1. Aguardar a nova leitura de Contas a Pagar terminar e conferir se o sinal negativo foi aceito; comparar somente totais agregados com o Bling, sem expor linhas pessoais nos logs.
2. Aplicar a decisão da responsável para movimentos `/caixas` com situação vazia; manter pendente até a resposta, sem adivinhar se foram concluídos.
3. Implementar e validar a transformação de Contas a Receber. Registros sem classificação B2B/B2C confirmada devem ficar como não classificados/pendentes e não entrar silenciosamente no total confirmado.
4. Medir e, se necessário, reduzir o tempo de Contas a Pagar sem pular títulos, sem alterar sinais nem enfraquecer idempotência, transações, limites ou o retry aprovado.
5. Ligar os dados confirmados às telas de origem e então verificar T01 usando o mesmo retrato compartilhado. Preservar desenho, textos e indicadores aprovados.
6. Tratar saldo por banco como fonte distinta. Configurar as pastas Drive antes de contar os arquivos externos. Depois reconciliar a DRE e CAPEX com as regras e documentos aprovados.

## Estado da execução local

- Migrations 000015, 000016 e 000017 aplicadas ao PostgreSQL isolado do protótipo.
- Permissões do worker reaplicadas como administrador de provisionamento e verificadas: `runtime database privileges: PASS`.
- Worker recompilado e reiniciado no protótipo local; integração Bling continua `CONNECTED`.
- A leitura de `/caixas` foi publicada normalizada e receber foi recebida em RAW. A tentativa de Contas a Pagar está em transação sem lote publicado. Na observação das 20:04, o banco mostrava uma transação aberta há 376 segundos, aguardando resposta do cliente; o worker continuava ativo. Não é possível afirmar ainda que a importação de pagar terminou.
- Nenhum valor de exemplo foi inserido para mascarar a falta de dados.

## Progresso e prazo

O progresso do backlog permanece em **12,1% concluído / 87,9% restante**. BK-058, BK-059 e BK-062 permanecem em **5% cada**: uma alteração local no banco ou uma resposta bruta da API não basta para considerar telas alimentadas e homologadas. Não há prazo confiável para terminar a integração real enquanto a carga de 10.198 títulos ainda não concluiu e o retrato de caixa não foi comparado. A meta registrada no projeto segue em 30/11/2026; essa data é alvo, não uma previsão garantida.

