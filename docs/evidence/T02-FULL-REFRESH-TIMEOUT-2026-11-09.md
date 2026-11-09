# Auditoria da atualização completa T02 — 09/11/2026

## O que foi observado

Às 02:03 BRT, consultei os lotes do banco isolado do protótipo enquanto a
atualização oficial continuava ativa:

- `contas.receber`: `SUCCESS`, 15.155 registros.
- `caixas.conciliadas`: `SUCCESS`, 3.162 registros.
- `contas.pagar`: `RUNNING`; 104 páginas da lista, 10.301 registros; página
  preparada mais recente: 5 de 104.
- `caixas`: `PARTIAL`, código `BLING_CASH_MOVEMENT_STATUS_MISSING`.

A janela real do aplicativo Wails foi trazida à frente. A árvore da T03 mostrou
traços e “Sem dados confirmados”. Assim, foi possível confirmar que o protótipo
estava aberto e vazio enquanto uma leitura extensa ainda não terminara. A
captura visual da janela falhou por tempo esgotado; não afirmo inspeção visual
de pixel.

## Causa do tempo elevado

A sincronização relê as contas e depois busca detalhes em blocos de 100
registros para normalizar os campos usados nas telas. A API limita o ritmo das
chamadas; a execução observada levou cerca de 40 segundos para preparar cada
página de Contas a Pagar. Com 104 páginas, 30 minutos não são suficientes. O
lote só publica após a preparação completa, portanto a T03 continua sem o novo
resultado enquanto o lote está `RUNNING`.

## Ajuste realizado no código local

Aumentei `BlingRefreshTimeout` para duas horas no contrato IPC compartilhado
por aplicativo e worker. Os comentários e testes usam a mesma constante. O
prazo estendido vale apenas para a leitura completa e somente de leitura; não
altera os prazos dos demais comandos, os relatórios de origem, o desenho ou as
regras financeiras.

## Limites desta evidência

O worker e aplicativo que estão abertos ainda não foram substituídos pela build
com o limite de duas horas. A carga que está em andamento começou antes desse
ajuste. Portanto, este registro não afirma que os dados já apareceram, que a
sincronização terminou ou que a T02 foi homologada. A captura da janela foi
inspecionada por acessibilidade; a captura visual automática falhou.

## Próximos passos

1. Deixar a execução atual chegar a um estado terminal e registrar seu
   resultado, sem iniciar uma sincronização concorrente.
2. Validar a mudança de limite com os testes, reconstruir o worker/aplicativo e
   confirmar que a próxima execução completa termina.
3. Conferir que o snapshot aparece na tela e comparar os períodos iguais com
   Contas a Receber (líquido), Contas a Pagar e o saldo conciliado do Itaú.
4. Abrir a T02 para a aprovação visual; T03 só começa após essa aprovação.
