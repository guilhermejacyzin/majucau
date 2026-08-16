# Catálogo inicial de tooltips — Majucau Financial Intelligence

- **Status:** proposto para implementação e revisão de linguagem simples
- **Regra:** todo componente funcional usa um `helpTextId` deste catálogo ou registra justificativa de não aplicabilidade
- **Acessibilidade:** tooltip contém somente texto; ações e links usam popover/dialog separado; informação crítica permanece visível na tela

## Tesouraria

| ID | Componente | Texto simples |
|---|---|---|
| `treasury.openingBalance` | Saldo Inicial do Dia | “Dinheiro conciliado disponível no início do dia. Ele deve ser igual ao saldo final confirmado do dia anterior.” |
| `treasury.todayClosing` | Saldo Final Projetado Hoje | “Estimativa do dinheiro ao final de hoje, considerando o saldo inicial e os movimentos previstos ou realizados do dia.” |
| `treasury.balance30` | Saldo Projetado em 30 Dias | “Estimativa do dinheiro disponível ao final dos próximos 30 dias no cenário selecionado.” |
| `treasury.balance60` | Saldo Projetado em 60 Dias | “Estimativa do dinheiro disponível ao final dos próximos 60 dias no cenário selecionado.” |
| `treasury.minimum60` | Menor Saldo Projetado — 60 Dias | “Menor saldo diário encontrado na projeção de 60 dias. Abra o card para ver a data e o que causou esse valor.” |
| `treasury.minimumReserve` | Reserva Mínima | “Valor mínimo que deve permanecer disponível para proteger a operação. Alterações precisam de aprovação.” |
| `treasury.maximumInvestment` | Valor Máximo para Aplicação | “Cálculo provisório do limite de aplicação segundo a regra proposta. O card permanece bloqueado até a fórmula ou a planilha de referência ser aprovada.” |
| `treasury.scenario` | Cenário | “Conjunto de premissas usado na projeção. Trocar o cenário não altera os dados realizados.” |
| `treasury.referenceDate` | Data de referência | “Data usada como início dos saldos, recebíveis, obrigações e projeções exibidos.” |

## Recebíveis

| ID | Componente | Texto simples |
|---|---|---|
| `receivables.b2cNuvem` | A Receber B2C — Nuvem | “Vendas futuras ao consumidor vindas exclusivamente da fonte Nuvem. Taxa, líquido e data só são confirmados quando a fonte oficial os fornece.” |
| `receivables.b2bBling` | A Receber B2B — Bling | “Títulos futuros classificados como vendas entre empresas no Bling. Casos sem classificação aprovada ficam fora do total confirmado.” |
| `receivables.total` | Total a Receber | “Soma do B2C futuro da Nuvem com o B2B futuro do Bling. O B2C do Bling não é somado.” |
| `receivables.receivedMonth` | Recebido no Mês | “Valores efetivamente recebidos no mês conforme a situação paga ou recebida no Bling.” |
| `receivables.aging` | Faixa de vencimento | “Agrupa o saldo aberto em vencidos, hoje, 7, 15, 30, 45 ou 60 dias, sem contar o mesmo título duas vezes.” |
| `receivables.openBalance` | Saldo em aberto | “Parte do título que ainda não foi recebida. Recebimentos parciais reduzem este valor.” |
| `receivables.classification` | Classificação B2B/B2C | “Regra que define a origem correta do recebível e evita dupla contagem. Alterações ficam registradas na auditoria.” |

## Obrigações

| ID | Componente | Texto simples |
|---|---|---|
| `payables.total` | A Pagar | “Saldo de obrigações ainda abertas no Bling para o período e filtros selecionados.” |
| `payables.overdue` | Obrigações Vencidas | “Obrigações com saldo aberto e vencimento anterior à data de referência.” |
| `payables.paidMonth` | Pago no Mês | “Valores efetivamente pagos no mês conforme os registros do Bling.” |
| `payables.dueDate` | Vencimento | “Data em que a obrigação deve ser paga.” |
| `payables.openBalance` | Saldo em aberto | “Parte da obrigação que ainda precisa ser paga após considerar pagamentos parciais.” |
| `payables.aging` | Faixa de pagamento | “Agrupa obrigações vencidas, de hoje e dos próximos 7, 15, 30, 45 ou 60 dias.” |

## Resultado e planejamento

| ID | Componente | Texto simples |
|---|---|---|
| `result.drePnl` | DRE / P&L | “Resultado de receitas, custos e despesas calculado pelo Majucau. Diferenças exigem ajuste identificado e aprovado.” |
| `result.ebitda` | EBITDA | “Resultado operacional antes de juros, impostos, depreciação e amortização, conforme a regra contábil aprovada.” |
| `result.ebitdaMargin` | Margem EBITDA | “Percentual do EBITDA sobre a receita usada pela regra aprovada.” |
| `planning.budgetActual` | Orçado x Realizado | “Compara o valor planejado com o que realmente ocorreu. Favorável ou desfavorável depende da natureza da linha.” |
| `planning.forecast` | Forecast | “Previsão versionada baseada em premissas e cenário. Uma nova versão não apaga o histórico.” |
| `management.workingCapital` | Capital de Giro | “Diferença entre ativos e obrigações de curto prazo conforme o plano de contas aprovado.” |
| `management.alerts` | Alertas | “Situações financeiras ou operacionais que precisam de atenção. Abra para ver a causa, a fonte e a ação esperada.” |

## Status e origem

| ID | Componente | Texto simples |
|---|---|---|
| `status.confirmed` | Confirmado | “O dado foi conciliado com a fonte oficial e passou pelas regras de validação.” |
| `status.partiallyConfirmed` | Parcialmente confirmado | “Parte do valor foi confirmada, mas ainda existem componentes pendentes.” |
| `status.provisional` | Provisório | “O valor pode mudar porque o período ou a fonte ainda não foi conciliado.” |
| `status.projected` | Projetado | “Estimativa futura calculada com as informações e o cenário disponíveis.” |
| `status.pendingReconciliation` | Pendente de conciliação | “O dado foi importado, mas ainda precisa ser comparado com a fonte oficial.” |
| `status.divergent` | Divergente | “Foi encontrada uma diferença que precisa ser investigada antes da confirmação.” |
| `status.stale` | Dados desatualizados | “A sincronização não terminou no prazo esperado. O último dado válido foi mantido.” |
| `status.schemaMismatch` | Formato incompatível | “A fonte mudou o formato dos dados e o Majucau interrompeu a atualização para evitar números incorretos.” |
| `source.lastUpdate` | Última atualização | “Última vez em que esta fonte terminou a sincronização com sucesso.” |
| `source.raw` | Registro original | “Versão exata do dado recebido da fonte antes das regras e cálculos do Majucau.” |

## Configurações — Bling

| ID | Componente | Texto simples |
|---|---|---|
| `bling.clientId` | Client ID | “Identificador público do aplicativo cadastrado no Bling.” |
| `bling.clientSecret` | Client Secret | “Chave secreta do aplicativo Bling. Ela é protegida no Windows e não volta a ser exibida depois de salva.” |
| `bling.redirectUri` | Redirect URI | “Endereço registrado no Bling para concluir a autorização.” |
| `bling.scopes` | Permissões | “Dados que o Bling autorizou o Majucau a ler.” |
| `bling.connect` | Conectar | “Abre o Bling no navegador para autorizar a leitura dos dados.” |
| `bling.test` | Testar conexão | “Verifica a autorização com uma leitura pequena, sem alterar dados no Bling.” |
| `bling.reconnect` | Reconectar | “Faz uma nova autorização quando a anterior expirou ou foi revogada.” |
| `bling.disconnect` | Desconectar | “Remove os tokens salvos e interrompe novas sincronizações. Os dados já importados são preservados.” |
| `bling.syncNow` | Sincronizar agora | “Solicita uma nova leitura. Registros repetidos não devem ser duplicados.” |

## Configurações — Nuvemshop e Nuvem Pago

| ID | Componente | Texto simples |
|---|---|---|
| `nuvemshop.appId` | App ID | “Identificador do aplicativo cadastrado na Nuvemshop.” |
| `nuvemshop.clientSecret` | Client Secret | “Chave secreta usada para concluir a autorização. Ela fica protegida no Windows.” |
| `nuvemshop.redirectUri` | Redirect URI | “Endereço HTTPS registrado para receber o retorno da autorização.” |
| `nuvemshop.storeId` | Loja conectada | “Identificador da loja devolvido pela Nuvemshop após a autorização.” |
| `nuvemshop.scopes` | Permissões | “Dados da loja que o aplicativo pode ler.” |
| `nuvemshop.connect` | Conectar | “Abre a Nuvemshop no navegador e inicia uma autorização temporária e protegida.” |
| `nuvemshop.test` | Testar conexão | “Confirma que a loja e a permissão de pedidos podem ser lidas sem alterar dados.” |
| `nuvempago.availability` | Disponibilidade do Nuvem Pago | “A confirmação financeira depende de uma API ou arquivo oficial com taxa, líquido e data de repasse.” |
| `nuvempago.partial` | Parcialmente disponível | “Pedidos e status podem estar disponíveis, mas faltam dados oficiais para confirmar o recebimento líquido.” |

## Tabelas, gráficos e filtros

| ID | Componente | Texto simples |
|---|---|---|
| `table.sourceId` | ID da fonte | “Identificador original usado para localizar o registro no sistema de origem.” |
| `table.competenceDate` | Competência | “Período ao qual a receita ou despesa pertence, mesmo que o pagamento ocorra em outra data.” |
| `table.cashDate` | Data de caixa | “Data em que o dinheiro entrou ou saiu de fato.” |
| `table.ruleVersion` | Versão da regra | “Versão exata da regra usada para produzir este valor.” |
| `chart.realized` | Série Realizado | “Valores que já ocorreram e foram conciliados.” |
| `chart.projected` | Série Projetado | “Valores futuros estimados pelo cenário selecionado.” |
| `filter.period` | Período | “Limita os dados pelas datas inicial e final informadas.” |
| `filter.status` | Status | “Mostra somente registros no estado selecionado.” |
| `filter.source` | Fonte | “Mostra somente valores vindos da integração escolhida.” |
| `filter.scenario` | Cenário | “Troca as premissas da projeção sem alterar os valores realizados.” |
| `action.drilldown` | Ver detalhes | “Abre a composição do valor, a regra usada e os registros de origem.” |
| `action.export` | Exportar | “Cria um arquivo com os dados filtrados e registra a ação na auditoria.” |

## Operação

| ID | Componente | Texto simples |
|---|---|---|
| `ops.backup` | Criar backup | “Cria uma cópia protegida do banco para recuperação. Tokens de integração não são incluídos.” |
| `ops.restore` | Restaurar backup | “Substitui os dados atuais por um backup validado. A sincronização fica pausada durante o processo.” |
| `ops.update` | Atualizar aplicativo | “Instala uma versão assinada depois de criar backup e verificar compatibilidade.” |
| `ops.audit` | Auditoria | “Mostra quem realizou uma ação material, quando ocorreu e o que foi alterado.” |
| `ops.syncError` | Erro de sincronização | “A fonte não pôde ser atualizada. Abra para ver uma explicação segura e tentar novamente.” |

## Critério automatizado de cobertura

O frontend manterá um catálogo TypeScript tipado. Testes falham quando:

- um card, campo, filtro, botão, status, alerta, gráfico, série ou coluna financeira funcional não possui `helpTextId` nem justificativa;
- o ID não existe neste catálogo versionado;
- o tooltip contém link, botão ou informação crítica exclusiva;
- hover funciona, mas foco/teclado não;
- `Escape` não fecha;
- `aria-describedby` não liga disparador e conteúdo;
- axe detecta violação ou o texto fica inacessível a 200% de zoom.
