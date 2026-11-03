# Status atual do Majucau — 02/11/2026

## Em que fase estamos

Estamos na etapa de **fazer a integração do Bling transformar os dados recebidos em informações que as telas conseguem mostrar**. O protótipo abre e o Bling está conectado, mas ainda não está pronto para homologação.

## O que já foi confirmado

- O login da API do Bling está conectado e o teste de conexão foi aprovado.
- O Bling respondeu com **14.901 registros de Contas a Receber** na última carga. Eles chegaram ao sistema, mas foram guardados como dados brutos; ainda não alimentam a tabela consultada pelo painel.
- O cálculo do fluxo já trata Contas a Pagar como saída negativa; testes automatizados cobrem essa regra. A tela própria de Contas a Pagar mantém o formato aprovado.
- O Bling trouxe **31.032 movimentos de caixa** entre 01/01/2026 e 30/10/2026, em nove contas. O código agora inclui entradas com situação `R` ou vazia nos indicadores mensais e marca o resultado como **Parcial** quando a situação falta. A tela ainda precisa receber essa nova versão e ser conferida.
- Conferi a T01 no protótipo agora: os cards e o gráfico continuam mostrando `—`/“Sem dados confirmados”. Isso é compatível com os estados atuais das fontes, não prova que os dados estejam aparecendo nas demais telas.
- A verificação de permissões do banco passou.

## Por que ainda há campos vazios

Conectar o Bling é como abrir a torneira: primeiro os dados chegam. Depois o Majucau precisa conferir, organizar e registrar cada tipo na tabela usada pela tela. Receber dados pela API não significa que eles já estão prontos para aparecer.

Neste momento:

- Contas a Receber chegou, mas continua em dados brutos.
- O erro antigo de Contas a Pagar (`BLING_PAYABLE_OPEN_BALANCE_NEGATIVE`) era da validação do saldo em aberto do título; não significava que o saldo de Caixa e Bancos estava negativo. Conta a pagar é saída e deve ser negativa no fluxo. Às 20:39, a tentativa iniciada às 19:57 seguia esperando dentro do banco havia aproximadamente 42 minutos e ainda não tinha publicado um novo lote.
- Preparei a correção para buscar páginas e detalhes do Bling sem manter aberta a transação que publica os dados. As páginas são guardadas em checkpoints limitados; no final, RAW, valores normalizados, snapshot, cursor e auditoria são publicados juntos em uma transação. A integração PostgreSQL descartável confirmou saída negativa, falha sem publicação parcial, preservação do retrato anterior e ausência de transação aberta durante as chamadas ao Bling. Esse código ainda não foi enviado ao GitHub nem aplicado ao worker que está aberto.
- Todas as 31.032 linhas de caixa vieram com `situacao` vazia. A orientação da usuária foi usar o fluxo de caixa do Bling como fonte dos movimentos registrados; passei a mostrar esses valores com aviso **Parcial**, sem mudar o sinal: entrada positiva e saída negativa.
- A execução das 19:57 terminou Contas a Receber em `SUCCESS`/`RAW_ONLY`; os dados foram recebidos, mas ainda não passaram para a tabela usada pela tela. O worker iniciou a etapa seguinte, mas Contas a Pagar segue sem lote normalizado confirmado.
- O sistema não tem saldo atual confirmado, separado por banco. Movimento de caixa não pode ser confundido com saldo da conta.
- A leitura do Google Drive está pendente porque o worker informa que as pastas autorizadas ainda não foram configuradas nele.

## O que falta, na ordem

1. Aplicar a nova versão ao protótipo, repetir a sincronização de `/caixas` e conferir os indicadores com aviso Parcial.
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

Atualização técnica e evidências de 02/11: [staging paginado e publicação atômica de Contas a Pagar](./evidence/BLING-API-BOUNDED-READ-ATOMIC-PUBLISH-2026-11-02.md).

Nenhuma tela ou regra financeira aprovada foi alterada nesta atualização.

## Correção após a conferência da tela aprovada

A responsável confirmou que a T02 deve refletir o resultado do relatório
**Fluxo de Caixa** do Bling; **somente os saldos por instituição** vêm de
**Caixas e Bancos**. Removi da T01 os dois avisos “Obrigações vencidas
aguardando fonte Bling” e “Recebimentos em atraso aguardando conciliação”, que
foram apontados como desnecessários. Também retirei a ligação incorreta dos
cards T02 “Entradas 30 dias” e “Saídas 30 dias” com os totais das telas de
Contas a Receber e Contas a Pagar.

Da captura do relatório foi possível registrar o período 02/11/2026–09/12/2026,
duas contas selecionadas, float e prazo de recebimento ativados, saldo inicial
R$ 74.463,36, receitas R$ 76.600,30, despesas -R$ 281.052,64, ajuste R$ 0,00 e
saldo final projetado -R$ 129.988,98. A tabela tem agrupamento semanal; não
permite reconstruir com precisão todos os dias da janela móvel de 30 dias. Os
valores foram registrados como evidência, não como dados vivos do protótipo.

**Ainda falta:** ligar o resultado real do relatório à T02 e a T01. A referência
pública atual da [API do Bling](https://developer.bling.com.br/referencia)
documenta os recursos de Caixas e Bancos, Contas a Pagar e Contas a Receber; não
lista um endpoint de relatório de Fluxo de Caixa. Portanto, não marquei a
integração do relatório como concluída nem usei os números da imagem como se
fossem atualização automática. Evidência: [leitura visual do relatório](./evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md).

Relatório técnico completo, com evidências e limitações: [auditoria do caminho dos dados do Bling](./evidence/BLING-DATA-PROJECTION-AUDIT-2026-11-02.md).

## Ajustes adicionais de conteúdo solicitados

Na T01, foram retirados os avisos genéricos de fonte/conciliação; na T02,
“Saídas previstas sem categoria”, “Recebimentos em atraso” e “Conciliação
pendente” foram removidos, mantendo o aviso de saldo projetado abaixo da
reserva. Na T10, foram retirados “Lotes com perda acima do esperado” e
“Separar plano realizado e obrigatório”. Em T11, os quatro avisos de cacau e
fornecedores permanecem porque a responsável pediu explicação antes de decidir
se os mantém. São textos-modelo sem confirmação de dados, não falhas ativas; a
grade foi empilhada para leitura.

Teste de interface e build frontend foram executados numa cópia isolada; a
atualização ainda não foi iniciada no executável Wails que já estava aberto.
Detalhes e explicações dos quatro avisos de Compras: [registro de conteúdo das
telas](./evidence/SCREEN-CONTENT-UPDATES-2026-11-02.md).

## Ajuste pedido na tela Estoque

Na tela **09 — Estoque**, foram removidos os cards `ENVELHECENDO` e
`VALIDADE` e o alerta `Lotes com validade próxima`, conforme as duas capturas
enviadas. Os demais itens ficaram inalterados. Foi acrescentado um teste para
proteger essa composição. Evidência: [remoção dos indicadores de validade](./evidence/INVENTORY-EXPIRATION-INDICATORS-REMOVAL-2026-11-02.md).


## Período dos cards da Visão Executiva — 02/11/2026

O período dos cards existentes **A RECEBER** e **A PAGAR** foi confirmado: vencimentos nos próximos 90 dias. A T01 agora liga cada card a uma métrica específica de 90 dias, em vez do total geral. Contas a Pagar mantém sinal negativo de saída. O card **MENOR SALDO 30 / 60 DIAS** e o desenho aprovado da tela continuam inalterados.

A captura anterior da prévia `127.0.0.1:4179` mostrava um valor geral sem período e não confirma o resultado de 90 dias. Essa prévia não acessa o worker. O app conectado está em `http://localhost:34115/`. Para conferir o novo valor de Contas a Pagar nessa aplicação é necessário carregar a versão atual do worker; Contas a Receber continuará sem número até a normalização dos dados Bling RAW-only.

**Verificações:** `go test ./internal/dashboard` passou; `npm run build` passou em `artifacts/verification/frontend-live`. A prévia estática recebeu o frontend atualizado. O progresso geral permanece **12,1% concluído / 87,9% restante**.

**Próximo passo:** reiniciar o protótipo conectado com o worker atualizado, conferir Contas a Pagar em 90 dias e implementar/validar a normalização de Contas a Receber.

### Continuação — preparação da integração de Contas a Receber

O código local agora prepara a normalização diária das contas recebidas do Bling. Cada conta mantém seu vínculo com o dado original, não é classificada automaticamente como B2B/B2C e só passa a valer no painel depois da publicação de um snapshot completo. A T01 e a tela de origem usam o mesmo conjunto publicado. A composição Bling + Nuvem segue a regra aprovada de comparar nome do cliente e valor bruto, ignorando a data; se o arquivo diário da Nuvem não estiver atualizado, o total combinado fica marcado como **Parcial**.

**Ainda não está rodando no protótipo:** falta aplicar a migration 000018 e as permissões, validar no PostgreSQL/CI, iniciar o worker novo e conferir os números reais nas telas. Gisele avisou que o arquivo externo de recebimentos futuros foi atualizado. A tentativa de abrir a pasta nesta execução foi interrompida pelo controle do navegador antes de ler o conteúdo; portanto, ainda não confirmei o formato, as colunas nem a importação.

O projeto continua em **12,1% concluído / 87,9% restante** e não está homologado. Evidência da preparação e dos passos pendentes: [normalização de Contas a Receber](./evidence/BLING-RECEIVABLE-NORMALIZATION-2026-11-02.md).

### Verificação após atualização do worker — 02/11/2026

Reiniciei o app Wails e o worker com o snapshot de 90 dias. A consulta sanitizada do worker ativo confirmou 418 títulos de Contas a Pagar na janela, com estado **Parcial**. Contas a Receber está **Indisponível** porque ainda não há linhas normalizadas; não exibi zero como valor confirmado. O total geral que aparecia na imagem da prévia não corresponde à métrica de 90 dias.

A aba estática `127.0.0.1:4179` continua sem conexão com o banco. O app local conectado permanece em `http://localhost:34115/`. Falta conferir visualmente no app e normalizar Contas a Receber. O projeto não está pronto para homologação. Meta geral: 12,1% concluído / 87,9% restante.
