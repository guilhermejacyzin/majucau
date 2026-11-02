# Status atual do Majucau — 02/11/2026

## Em que fase estamos

Estamos na etapa de **fazer a integração do Bling transformar os dados recebidos em informações que as telas conseguem mostrar**. O protótipo abre e o Bling está conectado, mas ainda não está pronto para homologação.

## O que já foi confirmado

- O login da API do Bling está conectado e o teste de conexão foi aprovado.
- O Bling respondeu com **14.901 registros de Contas a Receber** na última carga. Eles chegaram ao sistema, mas foram guardados como dados brutos; ainda não alimentam a tabela consultada pelo painel.
- O cálculo do fluxo já trata Contas a Pagar como saída negativa; testes automatizados cobrem essa regra. A tela própria de Contas a Pagar mantém o formato aprovado.
- Os movimentos de caixa do Bling agora foram lidos por completo e publicados como lote `SUCCESS`/`NORMALIZED`; o app continua conectado ao worker atualizado.
- Conferi a T01 no protótipo agora: os cards e o gráfico continuam mostrando `—`/“Sem dados confirmados”. Isso é compatível com os estados atuais das fontes, não prova que os dados estejam aparecendo nas demais telas.
- A verificação de permissões do banco passou.

## Por que ainda há campos vazios

Conectar o Bling é como abrir a torneira: primeiro os dados chegam. Depois o Majucau precisa conferir, organizar e registrar cada tipo na tabela usada pela tela. Receber dados pela API não significa que eles já estão prontos para aparecer.

Neste momento:

- Contas a Receber chegou, mas continua em dados brutos.
- O último lote publicado de Contas a Pagar falhou às 18:10 com `BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`. A regra do código já foi corrigida e testada para aceitar saldo devedor negativo; há uma nova transação de sincronização em andamento, mas ainda não publicou lote. Precisa terminar para provar a correção com dados reais.
- O Bling retorna `situacao` vazio nesses movimentos. A tela atual soma somente `R`; os registros foram preservados, mas não entram nos totais até a regra financeira de “realizado” ser confirmada.
- A execução das 19:57 terminou Contas a Receber em `SUCCESS`/`RAW_ONLY`; os dados foram recebidos, mas ainda não passaram para a tabela usada pela tela. O worker iniciou a etapa seguinte, mas Contas a Pagar segue sem lote normalizado confirmado.
- O sistema não tem saldo atual confirmado, separado por banco. Movimento de caixa não pode ser confundido com saldo da conta.
- A leitura do Google Drive está pendente porque o worker informa que as pastas autorizadas ainda não foram configuradas nele.

## O que falta, na ordem

1. Confirmar se movimentos `GET /caixas` com `situacao` vazia são realizados e podem entrar nos totais da tela.
2. Concluir e conferir Contas a Receber e Contas a Pagar na sincronização diária.
3. Transformar Contas a Receber em dados consultáveis, mantendo itens sem classificação como pendentes.
4. Fazer a atualização completa caber no tempo e nos limites da API sem deixar títulos de fora.
5. Ligar e conferir os dados nas telas de origem; depois conferir T01, que deve reunir esses mesmos resultados.
6. Obter uma fonte oficial de saldo atual para cada conta bancária; não inferir o saldo somando movimentos.
7. Configurar as pastas do Drive e validar os arquivos externos.
8. Conferir DRE e CAPEX conforme as regras contábeis já aprovadas. Fluxo de caixa não é igual à DRE.

## Meta e estimativa

- **Meta:** protótipo funcional e pronto para homologação até 30/11/2026.
- **Backlog:** 12,1% concluído; **87,9% falta**. BK-058, BK-059 e BK-062 continuam em 5% cada até os resultados reais aparecerem e serem conferidos.
- **Tempo:** ainda não há uma estimativa confiável. A carga de Contas a Pagar continua em andamento e o resultado não foi publicado. Vou medir o tempo quando essa etapa terminar, sem apresentar estimativa inventada.

Nenhuma tela ou regra financeira aprovada foi alterada nesta atualização.

Relatório técnico completo, com evidências e limitações: [auditoria do caminho dos dados do Bling](./evidence/BLING-DATA-PROJECTION-AUDIT-2026-11-02.md).

