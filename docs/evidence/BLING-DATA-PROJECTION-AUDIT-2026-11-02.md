# Auditoria do caminho dos dados do Bling até as telas

**Data:** 02/11/2026  
**Escopo:** integração diária, dados financeiros normalizados e T01 (Visão Executiva).  
**Limites preservados:** nenhuma tela aprovada ou regra financeira foi alterada nesta auditoria.

## Resposta curta

O Bling está conectado e já respondeu à API. O motivo de os valores ainda não aparecerem é que “recebido da API” e “pronto para a tela” são etapas diferentes. A consulta recente trouxe 14.896 contas a receber, mas essas linhas ficaram guardadas como dados brutos (`RAW_ONLY`). A tabela financeira que o painel consulta continua vazia para esse módulo.

A rotina de hoje começou pelo Bling às 19:29. Contas a Receber terminou às 19:33 com 14.896 linhas brutas. Contas a Pagar está sendo normalizada em uma transação; o banco ainda não mostrou a publicação do lote. A etapa de movimentos do fluxo de caixa vem depois de Contas a Pagar, então ainda não há resultado confirmado de `/caixas` nesta execução. No momento da conferência, as tabelas normalizadas de receber, pagar e movimentos tinham zero linhas.

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
| Contas a Receber | Lote terminou `SUCCESS`, 14.896 linhas lidas, estágio `RAW_ONLY`. | Não alimenta o total nem as linhas da tela porque ainda não foi convertido para a tabela `receivables`. |
| Contas a Pagar | Tentativas anteriores falharam ao rejeitar `saldo` negativo. A regra do parser e a migration 000016 agora aceitam o sinal negativo aprovado. A nova carga ainda estava em andamento durante a conferência. | Nenhum valor foi confirmado na tabela `payables` naquele instante. |
| Movimentos de caixa | O código já consulta `/caixas`, mantém direção separada do valor e grava saídas como saídas. A rotina diária executa receber, pagar e depois caixa. | A nova etapa `/caixas` ainda não tinha um lote finalizado nesta execução. |
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

1. Deixar a carga atual terminar e conferir os lotes de pagar e `/caixas`; comparar somente os totais agregados com o Bling, sem expor linhas pessoais nos logs.
2. Implementar e validar a transformação de Contas a Receber. Registros sem classificação B2B/B2C confirmada devem ficar como não classificados/pendentes e não entrar silenciosamente no total confirmado.
3. Reduzir o custo da carga de Contas a Pagar sem pular títulos, sem alterar sinais nem enfraquecer idempotência, transações, limites ou o retry aprovado. Primeiro é preciso medir quanto tempo a leitura atual realmente leva.
4. Ligar os dados confirmados às telas de origem e então verificar T01 usando o mesmo retrato compartilhado. Preservar desenho, textos e indicadores aprovados.
5. Tratar saldo por banco como fonte distinta. Configurar as pastas Drive antes de contar os arquivos externos. Depois reconciliar a DRE e CAPEX com as regras e documentos aprovados.

## Estado da execução local

- Migrations 000015 e 000016 aplicadas ao PostgreSQL isolado do protótipo.
- Permissões do worker reaplicadas como administrador de provisionamento e verificadas: `runtime database privileges: PASS`.
- Worker recompilado e reiniciado no protótipo local; integração Bling continua `CONNECTED`.
- A tentativa de sincronização diária continua em andamento. O lote bruto de receber chegou; pagar e caixa ainda não foram confirmados como normalizados.
- Nenhum valor de exemplo foi inserido para mascarar a falta de dados.

## Progresso e prazo

O progresso do backlog permanece em **12,1% concluído / 87,9% restante**. BK-058, BK-059 e BK-062 permanecem em **5% cada**: uma alteração local no banco ou uma resposta bruta da API não basta para considerar telas alimentadas e homologadas. Não há prazo confiável para terminar a integração real enquanto a carga de 10.198 títulos ainda não concluiu e o retrato de caixa não foi comparado. A meta registrada no projeto segue em 30/11/2026; essa data é alvo, não uma previsão garantida.

