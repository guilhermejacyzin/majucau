# Status atual do Majucau — 02/11/2026

## Em que fase estamos

Estamos na etapa de **fazer a integração do Bling transformar os dados recebidos em informações que as telas conseguem mostrar**. O protótipo abre e o Bling está conectado, mas ainda não está pronto para homologação.

## O que já foi confirmado

- O login da API do Bling está conectado e o teste de conexão foi aprovado.
- O Bling respondeu com **14.896 registros de Contas a Receber**. Eles chegaram ao sistema, mas foram guardados como dados brutos; ainda não alimentam a tabela consultada pelo painel.
- O banco separado do protótipo recebeu as atualizações para movimentos de caixa e para contas a pagar aparecerem negativas, como saídas.
- O worker atual do protótipo foi reconstruído e reiniciado. A carga diária de hoje está em andamento.
- A verificação de permissões do banco passou.

## Por que ainda há campos vazios

Conectar o Bling é como abrir a torneira: primeiro os dados chegam. Depois o Majucau precisa conferir, organizar e registrar cada tipo na tabela usada pela tela. Receber dados pela API não significa que eles já estão prontos para aparecer.

Neste momento:

- Contas a Receber chegou, mas continua em dados brutos.
- Contas a Pagar ainda está sendo organizada; a tentativa anterior parava porque um saldo negativo era rejeitado. O código e o banco foram corrigidos para tratar contas a pagar como saídas negativas.
- A leitura de movimentos do Fluxo de Caixa vem depois de Contas a Pagar e ainda não foi confirmada nesta execução.
- O sistema não tem saldo atual confirmado, separado por banco. Movimento de caixa não pode ser confundido com saldo da conta.
- A leitura do Google Drive está pendente porque o worker informa que as pastas autorizadas ainda não foram configuradas nele.

## O que falta, na ordem

1. Concluir e conferir a sincronização atual de Contas a Pagar e movimentos `/caixas`.
2. Transformar Contas a Receber em dados consultáveis, mantendo itens sem classificação como pendentes.
3. Fazer a atualização completa caber no tempo e nos limites da API sem deixar títulos de fora.
4. Ligar e conferir os dados nas telas de origem; depois conferir T01, que deve reunir esses mesmos resultados.
5. Obter uma fonte oficial de saldo atual para cada conta bancária; não inferir o saldo somando movimentos.
6. Configurar as pastas do Drive e validar os arquivos externos.
7. Conferir DRE e CAPEX conforme as regras contábeis já aprovadas. Fluxo de caixa não é igual à DRE.

## Meta e estimativa

- **Meta:** protótipo funcional e pronto para homologação até 30/11/2026.
- **Backlog:** 12,1% concluído; **87,9% falta**. BK-058, BK-059 e BK-062 continuam em 5% cada até os resultados reais aparecerem e serem conferidos.
- **Tempo:** ainda não há uma estimativa confiável. A sincronização atual de mais de 10 mil títulos não terminou e ainda falta observar o tempo de leitura de caixa. Dar uma data mais exata agora seria inventar.

Nenhuma tela ou regra financeira aprovada foi alterada nesta atualização.

Relatório técnico completo, com evidências e limitações: [auditoria do caminho dos dados do Bling](./evidence/BLING-DATA-PROJECTION-AUDIT-2026-11-02.md).

