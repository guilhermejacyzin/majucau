# Status atual do Majucau — 31/10/2026

Este registro resume o estado conferido do protótipo e do backlog. O objetivo continua sendo concluir as tarefas aprovadas, deixar o protótipo pronto para homologação e só então seguir para a finalização.

Relatório completo do processo, com contexto, decisões, evidências, pendências e ordem de retomada: [abrir o documento Word](../artifacts/planejamento/outputs/PROCESSO-REAL-PROJETO-MAJUCAU-2026-10-31.docx).

## Atualização de 01/11/2026

- **Verificação local no Windows — 01/11/2026:** corrigi o verificador de arquitetura para Windows PowerShell 5.1 e a codificação do teste de instalador. `scripts/verify.ps1` passou: suíte Go, limites de arquitetura, vet, govulncheck, migrations, SQLC, scan de secrets, instalador, lint, TypeScript, 61 testes frontend, build e npm audit (0 vulnerabilidades). A saída `frontend/dist` não pôde ser movida pelo Windows; seu caminho foi registrado para limpeza manual, sem encerrar processos ativos. Nenhuma regra financeira ou tela mudou. BK-025 continua em 70%; projeto continua em 11,9% / 88,1%. Evidência: [verificadores locais](./evidence/WINDOWS-LOCAL-VERIFICATION-2026-11-01.md).
- **Credencial e autorização do Bling — 01/11/2026:** revalidei o estado sem abrir o cofre. O arquivo protegido das credenciais agora existe; portanto, a instrução antiga de salvar o Client Secret novamente está superada. A última captura enviada pela responsável ainda mostra `BLING_NOT_CONNECTED` e “aguardando autorização”, pois a tentativa de OAuth foi fechada antes da conclusão. Uma evidência separada registrou chamada autorizada com HTTP 403 ao ler categorias; são execuções diferentes e precisam ser reconciliadas na próxima tentativa. Não li segredo/token, não alterei permissão no Bling e não contornei bloqueios do navegador/Windows. Evidência: [estado da credencial e autorização](./evidence/BLING-AUTHORIZATION-STATE-2026-11-01.md). BK-041 continua em 80%; a sincronização homologada ainda não está concluída.

- Parâmetros (T16) e Usuários (T17) agora abrem cada uma sua tela, em vez de abrir Integrações. As telas seguem as referências aprovadas e não mostram totais, pessoas ou acessos sem dados confirmados. Na tela de Integrações, removi os campos de conexão direta com a Nuvemshop, conforme a decisão de usar arquivos externos aprovados ou Bling por fonte. `scripts/verify.ps1` passou: 60 testes de frontend, verificações Go, lint, typecheck, build e auditoria npm. Isso corrige navegação e estados; não conclui T16/T17 nem prova sincronização real do Bling. Evidência: [T16/T17 e Integrações](./evidence/T16-T17-ROUTING-AND-INTEGRATIONS-UI-2026-11-01.md).
- A CI do commit `cbca622` passou nos quatro jobs, incluindo build/instalador Windows e testes PostgreSQL ([execução #41348424281](https://github.com/guilhermejacyzin/majucau/actions/runs/41348424281)). O percentual geral segue 11,9% concluído; essa correção não conclui uma tarefa do backlog.
- A captura enviada com os cards `PRIORIDADE`, `FALLBACK`, `REUSO BLOQUEADO` e `SEM ID BLING` foi rastreada até a antiga T05 — Conciliação, já removida. A T03 conserva os cards aprovados; adicionei uma regressão para impedir o retorno dos cards T05. A suíte passou com 61 testes, lint, typecheck e CI nos quatro jobs ([#41349556489](https://github.com/guilhermejacyzin/majucau/actions/runs/41349556489)). Evidência: [retirada T05 e auditoria dos cards](./evidence/T05-RETIREMENT-2026-10-31.md).
- A página independente de Conciliação e o alerta residual foram retirados e enviados ao GitHub. Gisele confirmou que não localizou mais essa página/referência. O card “CONCILIAÇÃO” dentro de Contas a Receber continua porque pertence à tela aprovada.
- Os atalhos “Ver detalhes” da Visão Executiva agora levam às telas que explicam os resultados. O desenho e os cálculos foram preservados e clicar não consulta o Bling. 57 testes, lint e typecheck passaram localmente; a CI [#41274539700](https://github.com/guilhermejacyzin/majucau/actions/runs/41274539700) passou nos quatro jobs, incluindo Windows e PostgreSQL. A confirmação manual no protótipo não foi possível porque a captura da janela expirou.
- O projeto segue estimado em 11,9% concluído (88,1% restante). BK-102 continua em 1%; integração com dados reais e aceites das telas ainda faltam.
- Na auditoria de dados da T01, confirmei que A Receber e A Pagar compartilham seus totais com as respectivas telas. Ainda faltam os demais indicadores do painel, inclusive saldo por banco; Itaú e Mercado Pago não estão confirmados no sistema. Por aprovação de Gisele, valores parciais agora mostram o aviso “Parcial” na linha de apoio existente. Veja [a evidência](./evidence/EXECUTIVE-DETAIL-DATA-FLOW-2026-10-25.md).

## FEITO

- O protótipo local abre com um PostgreSQL isolado e um serviço local de integração. Na última conferência, os processos e o banco estavam ativos.
- Há código para conectar ao Bling, guardar o Client Secret no cofre protegido do Windows e usar a API atual com autenticação JWT. A integração também tem controles de paginação, repetição segura e gravação transacional/idempotente. Isso ainda não prova acesso à conta real.
- A leitura de Contas a Pagar já tem conversão técnica para os dados internos e atualização compartilhada das telas Visão Executiva e Contas a Pagar.
- Em 31/10, o card já aprovado “VENCIMENTOS 7 DIAS” foi ligado ao cálculo do backend: considera saldo aberto com vencimento nos dias 1–7 e exclui situações canceladas/estornadas. O desenho não mudou; essa entrega é parcial e não conclui T04 nem BK-058/BK-059.
- Também em 31/10, os cálculos existentes “Vence hoje” e “Em atraso” passaram a excluir cancelados/estornados, conforme aging aprovado. Os testes locais do pacote passaram; a integração PostgreSQL desta continuação aguarda CI. O total geral e a tabela permaneceram iguais.
- Os cálculos de consulta dos cards EBIT e EBITDA e as validações da calculadora para D0 até D+60 estão implementados. Ainda faltam dados reais e validação contábil da empresa.
- A regra aprovada para CAPEX começa em 01/01/2026 e prevê os cards de valor pago no mês e acumulado. Os valores ainda não estão confirmados.
- Existem leitores em fluxo para CSV, XLS e XLSX. A leitura financeira real dos arquivos do Drive ainda não foi homologada.
- As telas aprovadas e as regras financeiras foram preservadas nesta etapa.

## VALIDADO

- O protótipo e o banco isolado estavam ativos na última verificação. O painel abria sem valores financeiros confirmados.
- O registro de 31/10 dizia que `bling_credentials_v1.dpapi` não existia naquela verificação. Em 01/11 ele passou a existir; somente presença e metadados foram conferidos, sem abrir o cofre. A última captura do Majucau mostra autorização pendente. Outra execução está registrada em `BLING-LIVE-SYNC-SCOPE-2026-11-01.md` e encontrou HTTP 403 na leitura de categorias; consultar `BLING-AUTHORIZATION-STATE-2026-11-01.md` para distinguir os dois estados, sem afirmar que a credencial ainda falta.
- O commit `d549dff` está enviado ao GitHub. A CI `41244760487` terminou com sucesso nos trabalhos de detecção de mudanças e verificação do backlog; os trabalhos de Windows e PostgreSQL foram pulados porque o commit alterou apenas documentação. A CI anterior `41244616921` também foi documental e pulou esses dois trabalhos. Portanto, esses resultados não devem ser descritos como testes completos do produto.
- O primeiro envio deste resumo foi o commit `92b2875`, na branch `codex/majucau-artifact-consolidation`. A CI [41245308973](https://github.com/guilhermejacyzin/majucau/actions/runs/41245308973) passou na verificação do backlog, mas falhou ao conferir o formato da mensagem: faltou o campo obrigatório `Quanto falta da tarefa: N%`. Os trabalhos de Windows e PostgreSQL foram pulados. A falha é de governança da mensagem do commit, não de código do produto. A correção é enviada em commit posterior, sem reescrever o histórico.
- A correção foi enviada no commit `6bad5ed`. A CI [41245420968](https://github.com/guilhermejacyzin/majucau/actions/runs/41245420968) passou em `detect-changes` e `verify-backlog`; Windows e PostgreSQL foram pulados porque a alteração foi documental. A mensagem corrigida não reescreveu o histórico.
- O checksum local `frontend/package.json.md5` está modificado pelo protótipo em execução e não foi incluído nesses commits. O registro local de limpeza explica quando restaurá-lo.
- O estado geral do backlog é 127 tarefas: 7 concluídas, 20 em andamento e 100 não iniciadas. A estimativa média é **11,8% concluído e 88,2% restante**; ela não mede prontidão para homologação.
- A tarefa da conexão real do Bling, BK-041, está registrada em **80% concluída, com 20% restante**.
- Nesta execução `scripts/verify.ps1` terminou com sucesso; também passaram a instalação limpa do frontend a partir do lockfile atualizado e seus lint, typecheck, 39 testes, build e auditoria npm (zero vulnerabilidades). O teste SQL de PostgreSQL foi ignorado localmente por falta de `MAJUCAU_TEST_DATABASE_URL`, mas a CI [41261894619](https://github.com/guilhermejacyzin/majucau/actions/runs/41261894619) executou e aprovou os quatro trabalhos, inclusive PostgreSQL e Windows. Isso não substitui a homologação com a conta real do Bling.

## Atualização da continuação — 31/10/2026

- O commit atual é `ea7e379` na branch
  `codex/majucau-artifact-consolidation`; a alteração local
  `frontend/package.json.md5` continua preservada e fora do Git.
- A falha Windows do commit anterior foi corrigida com repetição limitada das
  consultas somente de leitura ao estado do serviço. Erros persistentes continuam
  resultando em falha.
- A CI [41263767050](https://github.com/guilhermejacyzin/majucau/actions/runs/41263767050)
  passou em `detect-changes`, `verify-backlog`, `verify-postgres` e
  `verify-windows`, incluindo o smoke do ciclo de vida do serviço.
- `scripts/verify.ps1` também passou localmente. O smoke local de instalação do
  serviço foi preservado para a CI, pois o worker do protótipo estava ativo e o
  terminal não estava elevado.
- Evidência e limites: [investigação do smoke Windows](./evidence/WINDOWS-WORKER-SMOKE-RETRY-2026-10-31.md).

## PENDÊNCIA REAL

1. **Concluir/confirmar a autorização do Bling no Majucau:** o Client Secret protegido já está salvo nesta instância; não é necessário enviá-lo novamente. A última captura mostra a tentativa interrompida e o estado `BLING_NOT_CONNECTED`.
2. Depois que a tela confirmar conexão, testar a leitura, resolver a permissão de categorias indicada pelo HTTP 403, validar paginação real e concluir a primeira sincronização. Se houver erro, corrigir essa mesma etapa antes de avançar.
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
- [Evidência parcial T04 — vencimentos nos próximos sete dias](./evidence/T04-PAYABLES-NEXT-7-DAYS-2026-10-31.md)
