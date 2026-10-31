# Status atual do Majucau — 31/10/2026

Este registro resume o estado conferido do protótipo e do backlog. O objetivo continua sendo concluir as tarefas aprovadas, deixar o protótipo pronto para homologação e só então seguir para a finalização.

Relatório completo do processo, com contexto, decisões, evidências, pendências e ordem de retomada: [abrir o documento Word](../artifacts/planejamento/outputs/PROCESSO-REAL-PROJETO-MAJUCAU-2026-10-31.docx).

## FEITO

- O protótipo local abre com um PostgreSQL isolado e um serviço local de integração. Na última conferência, os processos e o banco estavam ativos.
- Há código para conectar ao Bling, guardar o Client Secret no cofre protegido do Windows e usar a API atual com autenticação JWT. A integração também tem controles de paginação, repetição segura e gravação transacional/idempotente. Isso ainda não prova acesso à conta real.
- A leitura de Contas a Pagar já tem conversão técnica para os dados internos e atualização compartilhada das telas Visão Executiva e Contas a Pagar.
- Os cálculos de consulta dos cards EBIT e EBITDA e as validações da calculadora para D0 até D+60 estão implementados. Ainda faltam dados reais e validação contábil da empresa.
- A regra aprovada para CAPEX começa em 01/01/2026 e prevê os cards de valor pago no mês e acumulado. Os valores ainda não estão confirmados.
- Existem leitores em fluxo para CSV, XLS e XLSX. A leitura financeira real dos arquivos do Drive ainda não foi homologada.
- As telas aprovadas e as regras financeiras foram preservadas nesta etapa.

## VALIDADO

- O protótipo e o banco isolado estavam ativos na última verificação. O painel abria sem valores financeiros confirmados.
- Verifiquei apenas se existe o arquivo protegido `artifacts/retained/postgresql-dev/credentials/worker-vault/bling_credentials_v1.dpapi`. Ele não existe. O conteúdo de nenhum cofre foi lido.
- Não houve autorização OAuth, chamada à conta real do Bling nem sincronização de dados reais.
- O commit `d549dff` está enviado ao GitHub. A CI `41244760487` terminou com sucesso nos trabalhos de detecção de mudanças e verificação do backlog; os trabalhos de Windows e PostgreSQL foram pulados porque o commit alterou apenas documentação. A CI anterior `41244616921` também foi documental e pulou esses dois trabalhos. Portanto, esses resultados não devem ser descritos como testes completos do produto.
- O primeiro envio deste resumo foi o commit `92b2875`, na branch `codex/majucau-artifact-consolidation`. A CI [41245308973](https://github.com/guilhermejacyzin/majucau/actions/runs/41245308973) passou na verificação do backlog, mas falhou ao conferir o formato da mensagem: faltou o campo obrigatório `Quanto falta da tarefa: N%`. Os trabalhos de Windows e PostgreSQL foram pulados. A falha é de governança da mensagem do commit, não de código do produto. A correção é enviada em commit posterior, sem reescrever o histórico.
- A correção foi enviada no commit `6bad5ed`. A CI [41245420968](https://github.com/guilhermejacyzin/majucau/actions/runs/41245420968) passou em `detect-changes` e `verify-backlog`; Windows e PostgreSQL foram pulados porque a alteração foi documental. A mensagem corrigida não reescreveu o histórico.
- O checksum local `frontend/package.json.md5` está modificado pelo protótipo em execução e não foi incluído nesses commits. O registro local de limpeza explica quando restaurá-lo.
- O estado geral do backlog é 127 tarefas: 7 concluídas, 20 em andamento e 100 não iniciadas. A estimativa média é **11,8% concluído e 88,2% restante**; ela não mede prontidão para homologação.
- A tarefa da conexão real do Bling, BK-041, está registrada em **80% concluída, com 20% restante**.

## PENDÊNCIA REAL

1. **Conectar o Bling:** falta salvar o Client Secret do aplicativo no campo **Client Secret** em **Importações/Integrações → Bling** no Majucau. Essa credencial é diferente da senha usada para entrar no site Bling; não deve ser enviada pelo chat.
2. Depois de salvar a credencial, concluir a autorização do Bling, testar a conexão, validar a paginação real e fazer a primeira sincronização. Se houver erro, corrigir essa mesma etapa antes de avançar.
3. **Drive:** ainda faltam provisionar o acesso técnico de leitura e validar a ingestão real das três pastas aprovadas no horário diário das 09:30.
4. **Resultados financeiros:** faltam mapear e conferir fontes reais para DRE, EBIT/EBITDA, calculadora e CAPEX. O CAPEX só pode receber valores depois de ligar categoria, pagamento, data e valor de forma rastreável.
5. A homologação do protótipo e as demais tarefas do backlog continuam pendentes; não há prazo confiável enquanto as fontes reais não forem conectadas e conferidas.

## Preservação local

- A credencial de desenvolvimento e os dados do PostgreSQL permanecem em `artifacts/retained/`; caches ficam em `artifacts/cache/` enquanto forem necessários ao protótipo.
- O protótipo alterou localmente o checksum `frontend/package.json.md5`. Ele não foi enviado ao GitHub; restaurar o arquivo versionado após fechar o protótipo pelo ciclo normal.

## Registros relacionados

- [Meta e histórico do projeto](./META-PROJETO.md)
- [Pré-checagem da conexão Bling](./evidence/BLING-CONNECTION-PRECHECK-2026-10-31.md)
- [Política contábil da DRE](./DRE-ACCOUNTING-POLICY-v1.md)
- [Auditoria da fonte de CAPEX](./evidence/CAPEX-REPORT-DATA-SOURCE-AUDIT-2026-10-11.md)
