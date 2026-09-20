# Plano de implementação - Majucau Financial Intelligence

- **Status:** G0 aprovado; execução G1 em andamento
- **Gate G0:** aprovado em 2026-08-16 com D-001-A, D-002-A, D-003-A, D-004-A e D-005-A
- **Objetivo final:** instalador Windows assinado, testado e pronto para produção
- **Método:** entregas verticais com gates; nenhum gate é aprovado apenas porque “aparece na tela”

## 1. Escala relativa

| Tamanho | Significado |
|---|---|
| S | tarefa delimitada, baixo acoplamento |
| M | módulo com integração ou testes relevantes |
| L | fluxo completo com persistência, erros e UI |
| XL | fase sistêmica, múltiplos módulos e validação de produção |

Estimativas são relativas. Datas dependem de acesso às APIs, aprovação financeira, certificado de assinatura e ambiente Windows de teste.

### Fora do escopo deste MVP

Não implementar nesta entrega: previsão comercial por SKU, previsão de demanda, planejamento de produção, otimização de estoque, recomendação de compras, pricing comercial avançado, CRM, gestão de marketing e integrações diretas com Appmax, Mercado Pago ou Itaú. Esses itens exigem nova decisão de escopo e não podem entrar por dependência incidental.

## 2. Gates obrigatórios

| Gate | Condição |
|---|---|
| G0 - Arquitetura | ADR, ERD, mapa de APIs, dicionário e plano aprovados; disposição de cada P0 registrada; autoriza somente a fundação e módulos não bloqueados |
| G1 - Fundação | build reproduzível, banco, worker, desktop e CI funcionando |
| G2 - Integrações | credenciais seguras, OAuth, RAW e idempotência validados |
| G3 - Tesouraria | projeção e reconciliações fechando com casos aprovados |
| G4 - Executivo | dashboard, tooltips e drill-down completos |
| G5 - Resultado | DRE/P&L/EBITDA com regras aprovadas e rastreáveis |
| G6 - Operação | backup, restore, update, logs e segurança validados |
| G7 - Produção | instalador assinado aprovado em VMs limpas e aceite funcional concluído |

G0 não será aprovado apenas por existir documentação. Cada P0 abaixo precisa de uma disposição explícita: resolvido por evidência, retirado do release por decisão de escopo ou diferido mantendo todos os módulos/resultados dependentes desabilitados. O diferimento permite construir a fundação, mas continua bloqueando G2/G3/G5/G7 conforme a dependência. Até G0, não há código de produção; somente pesquisa, protótipos descartáveis e artefatos de decisão.

| P0 | Resolução exigida |
|---|---|
| Fonte B2C/Nuvem Pago | Tarifário aprovado: cartão 1x/2x/3x em D+30 (2,59% + R$ 0,35; 4,49% + R$ 0,35; 5,44% + R$ 0,35), boleto D+2 (R$ 2,39) e PIX na hora (0,99%); ainda falta API/arquivo oficial com bruto, taxa efetiva, líquido, status e data de liquidação, ou retirada explícita do B2C confirmado do release |
| OAuth Nuvemshop desktop | redirect HTTPS do relay registrado e aceito, state/replay/TTL testados e posição do provedor sobre PKCE documentada |
| Classificação B2B | aprovação da regra conservadora e de seus casos ambíguos/overrides |
| Aplicação financeira | planilha-oráculo ou aprovação formal da fórmula v1 proposta, com casos dourados |

## 3. Fase 0 - Governança e decisões

**Tamanho:** M  
**Gate:** G0

Entregáveis:

- ADR de arquitetura/empacotamento;
- ERD v1;
- API Integration Map;
- Data Dictionary v1;
- este plano;
- registro explícito de decisões pendentes;
- política “não inventar regra financeira”.

Critérios:

- nenhuma contradição material entre documentos;
- fontes oficiais e lacunas identificadas;
- Nuvem Pago não tratado como API comprovada;
- decisões de moeda, timezone, escala e arredondamento encaminhadas;
- usuário aprova o início do código e escolhe como tratar cada P0.

## 4. Fase 1 - Fundação do repositório e CI

**Tamanho:** L  
**Dependência:** G0

### 4.1 Estrutura

- criar monorepo;
- adicionar `AGENTS.md` com regras do handoff;
- inicializar Go modules;
- inicializar Wails v2 estável;
- criar React + TypeScript + Vite;
- configurar formatting, lint e typecheck;
- adicionar SQL migrations e sqlc;
- criar contratos de domínio sem integrações reais.

### 4.2 Pipeline

Executar:

```text
go test ./...
go vet ./...
govulncheck ./...
sqlc generate e verificação de diff
migration lint/checksum
eslint
tsc --noEmit
vite build
testes frontend
scan de secrets
SBOM/licenças
```

### 4.3 Primeiro vertical slice

- `majucau.exe` abre janela Wails;
- `majucau-worker.exe` responde health check via Named Pipe;
- PostgreSQL efêmero de teste aplica migrations;
- UI mostra estado real do worker e banco;
- todos os componentes do slice possuem tooltip acessível.

## 5. Fase 2 - Fundação Windows e instalador inicial

**Tamanho:** XL  
**Dependência:** Fase 1

- criar worker como serviço Windows;
- configurar conta virtual `NT SERVICE\MajucauWorker`, SID estável, perfil e SDDL idempotente;
- empacotar PostgreSQL x64 dedicado;
- reservar porta sob lock, configurar loopback/SCRAM e gerar credenciais aleatórias protegidas por DPAPI;
- registrar dependência do worker no serviço PostgreSQL e bloquear execução antes de readiness/migrations;
- definir `%ProgramData%` e ACLs;
- instalar WebView2 quando ausente;
- gerar NSIS x64;
- criar atalhos e desinstalador;
- preservar dados em upgrade/uninstall;
- validar reboot e recuperação do serviço;
- produzir installer smoke test em VM limpa.
- verificar hash/pacote offline do WebView2 e rollback recuperável de instalação interrompida.

Gate parcial: instalação não exige terminal, Docker, Node ou configuração manual de PostgreSQL.

## 6. Fase 3 - Banco, RAW, auditoria e segurança

**Tamanho:** XL  
**Dependência:** Fases 1-2

- migrations do núcleo;
- `integration_connections`;
- `integration_sync_batches`;
- view/relatório `integration_sync_logs` sobre os lotes, preservando o nome e os campos requeridos pelo handoff;
- `sync_cursors`;
- `raw_records` append-only;
- tabelas de contribuição/linhagem tipadas com FKs reais e view agregada `record_lineage`;
- `audit_events`;
- constraints de idempotência;
- cálculo de payload hash;
- transações por lote;
- DPAPI sob a conta do worker;
- Named Pipe com DACL e protocolo versionado;
- sanitização de logs;
- testes de duplicidade, concorrência e falha no meio do lote.
- testes de percurso card -> snapshot -> contribuição -> entidade -> RAW.

Critérios:

- repetir o mesmo lote não duplica fatos;
- falha antes do commit não avança cursor;
- nova versão externa preserva versão RAW anterior;
- secrets não aparecem no banco, logs ou frontend.

## 7. Fase 4 - Design system, tooltips e tela de integrações

**Tamanho:** L  
**Dependência:** Fases 1 e 3

### 7.1 Design system

- tokens de cor, tipografia e espaçamento;
- componentes de card, status, alerta, campo, botão, tabela, modal e gráfico;
- wrapper único de tooltip baseado em primitive acessível;
- catálogo central de textos explicativos;
- contraste e navegação por teclado;
- estados loading/empty/error/stale/partial/confirmed.

### 7.2 Regra de cobertura de tooltips

- todo componente funcional declara `helpText` ou justificativa explícita de não aplicabilidade;
- teste de componente verifica tooltip em hover e foco;
- `Escape` fecha tooltip;
- tooltip contém somente texto; links e ações usam popover/dialog acessível;
- conteúdo essencial também aparece de forma persistente;
- textos financeiros passam por revisão de linguagem simples.
- axe automatizado, leitor de tela, zoom 200%, ordem de foco e teclado entram no aceite de cada tela.

### 7.3 Screen Configurações > Integrações

- cards independentes Bling, Nuvemshop e Nuvem Pago;
- campos mascarados;
- salvar no worker/DPAPI;
- Conectar, Testar, Reconectar e Desconectar;
- scopes, conta/loja e status;
- última tentativa e último sucesso;
- erros simples + detalhe sanitizado;
- first-run wizard direciona para esta screen;
- nenhuma sincronização inicial sem confirmação.

## 8. Fase 5 - Conector Bling

**Tamanho:** XL  
**Dependência:** Fases 3-4 e aplicativo Bling registrado

### 8.1 OAuth

- Authorization Code no navegador externo;
- callback aprovado;
- state e PKCE quando suportado;
- troca e refresh serializados;
- armazenamento DPAPI;
- teste de conexão read-only;
- tratamento de revogação/expiração.

### 8.2 Recursos

- contas a receber;
- contas a pagar;
- realizados;
- contas financeiras;
- caixa/banco;
- categorias;
- contatos;
- pedidos de venda/compra;
- NF-e.

### 8.3 Resiliência

- polling incremental;
- janelas sobrepostas;
- paginação;
- rate limiting;
- 429/backoff;
- retries controlados;
- checkpoint transacional;
- contrato e fixtures sanitizados.

Critérios:

- contas a pagar fecham com Bling no mesmo filtro com tolerância R$ 0,01;
- saldo financeiro conciliado D-1 fecha com a visão equivalente do Bling, no mesmo conjunto de contas e filtros, com tolerância R$ 0,01; somente então alimenta o Saldo Inicial D0;
- recebíveis B2B fecham após regra de classificação aprovada;
- pagamentos/recebimentos realizados fecham com tolerância R$ 0,01;
- todo valor possui drill-down até o RAW.
- após aprovação D-005, aging de recebíveis e obrigações separa vencidos, hoje, 7, 15, 30, 45 e 60 dias, com início/fim inclusivos e sem dupla contagem entre faixas;
- testes de aging cobrem vencimento no limite, feriado/fim de semana, parcial, cancelado e timezone de negócio.

## 9. Fase 6 - Conector Nuvemshop e estratégia Nuvem Pago

**Tamanho:** L para Nuvemshop; indeterminado para Nuvem Pago  
**Dependência:** aplicativo Nuvemshop registrado e callback HTTPS relay aprovado

### 9.1 Nuvemshop

- OAuth Authorization Code;
- relay HTTPS mínimo para `code/state`, sessão de uso único e TTL de cinco minutos; secrets/tokens permanecem locais;
- pedidos;
- clientes;
- status de pagamento;
- parcelas e gateway quando disponíveis;
- polling por datas/`since_id`;
- cancelamento, reembolso e void;
- loja e scopes exibidos na screen.

### 9.2 Nuvem Pago

- manter interface de adapter;
- não liberar valor líquido/agenda sem fonte oficial;
- solicitar documentação privada ou habilitação oficial;
- avaliar importação de arquivo oficial como fallback, somente após obter exemplo real;
- marcar dados incompletos como provisórios/parciais.

Gate G2 somente será aprovado quando o B2C futuro possuir fonte oficial para bruto, taxa, líquido e data prevista, ou quando o escopo do release retirar explicitamente o B2C confirmado. Não existe aprovação parcial de G2 para um release que prometa o card B2C.

## 10. Fase 7 - Tesouraria e projeção financeira

**Tamanho:** XL  
**Dependência:** Bling validado, regras temporais aprovadas

- saldo final conciliado D-1;
- saldo inicial D0;
- D0 provisório;
- projeção diária D+1 a D+60;
- continuidade diária;
- entradas e saídas por data;
- ajustes autorizados;
- cenários BASE/CONSERVATIVE/STRESS/OPTIMISTIC;
- menor saldo e respectiva composição;
- reserva mínima;
- alertas de liquidez;
- snapshots e lineage.

Testes:

- diferença diária R$ 0,00;
- continuidade R$ 0,00;
- datas em virada de mês/ano e horário de verão histórico;
- valores negativos, cancelamentos e estornos;
- repetição de cálculo com mesmos inputs produz mesmo hash/resultado;
- alteração de regra cria nova versão, sem reescrever snapshot.

## 11. Fase 8 - Dashboard executivo e drill-down

**Tamanho:** XL  
**Dependência:** Fase 7

- cards obrigatórios do handoff;
- gráficos de projeção;
- filtros de cenário e período;
- status e origem do dado;
- última atualização por fonte;
- alertas;
- drill-down: card -> módulo -> detalhe -> registro -> RAW;
- tooltips simples em todos os componentes funcionais;
- estados de dado desatualizado sem zerar valor;
- exportação controlada quando aprovada.

Cards obrigatórios, com estes nomes oficiais:

- Tesouraria: Saldo Inicial do Dia; Saldo Final Projetado Hoje; Saldo Projetado em 30 Dias; Saldo Projetado em 60 Dias; Menor Saldo Projetado — 60 Dias; Reserva Mínima; Valor Máximo para Aplicação;
- Recebíveis: A Receber B2C — Nuvem; A Receber B2B — Bling; Total a Receber; Recebido no Mês;
- Obrigações: A Pagar; Obrigações Vencidas; Pago no Mês;
- Resultado: DRE / P&L; EBITDA;
- Planejamento: Orçado x Realizado; Forecast;
- Gestão: Capital de Giro; Alertas.

O menu usará “Tesouraria e Projeção Financeira”. Não haverá card genérico chamado “Caixa”. O wireframe aprovado do handoff é referência visual obrigatória para sidebar, cards clicáveis, estados, gráficos, alertas, drill-down, filtros e atualização por fonte; não é especificação matemática. O aceite visual terá comparação lado a lado e checklist desses elementos.

Tipos iniciais obrigatórios de alerta:

- saldo projetado abaixo da reserva mínima;
- obrigações vencidas;
- recebíveis vencidos;
- grandes pagamentos próximos;
- queda relevante de EBITDA;
- erro de sincronização;
- divergência contábil;
- dado desatualizado.

Critérios:

- cada card responde onde estamos, para onde vamos, origem e risco;
- navegação completa por teclado;
- nenhum tooltip é a única fonte de informação crítica;
- nenhuma tela apresenta valor sem fonte/status.

## 12. Fase 9 - DRE, P&L e EBITDA

**Tamanho:** XL  
**Dependência:** categorias, plano de contas e regras funcionais aprovados

- regras de classificação versionadas;
- DRE parametrizável;
- P&L reconciliável;
- ajustes com motivo, usuário, data e aprovação;
- EBITDA e margem;
- realizado, orçado, forecast e variação;
- lineage por linha;
- divergências explícitas.
- importador de folha CSV versionado, hash de arquivo/linha, idempotência, validação de PII e total por competência;
- template/exemplo sanitizado de folha e fila de erros sem ingestão parcial silenciosa.
- separar Produção, Comercial e Administrativo;
- separar empregados CLT, sócios/pró-labore e desligados, preservando histórico;
- tratar adiantamento salarial como compensação contra folha, nunca despesa adicional;
- manter PLR como `NOT_APPLICABLE` até nova aprovação funcional.

Não usar demonstrações prontas do Bling como verdade contábil.

## 13. Fase 10 - Planejamento e gestão

**Tamanho:** XL  
**Dependência:** Fases 7 e 9

- Orçado x Realizado;
- favorabilidade por natureza da linha;
- Forecast versionado;
- Capital de Giro;
- parâmetros de aplicação;
- modelo de referência proposto;
- cálculo do Valor Máximo para Aplicação;
- testes de equivalência com casos aprovados.

O cálculo de aplicação permanece provisório até aprovação do modelo de referência. Diferença permitida após aprovação: R$ 0,01.

## 14. Fase 11 - Contabilidade

**Tamanho:** XL  
**Dependência:** revisão contábil especializada

- plano de contas;
- períodos;
- lançamentos e partidas;
- ajustes;
- balancete;
- balanço patrimonial;
- validação Ativo = Passivo + Patrimônio Líquido;
- status DIVERGENT para diferença não zero;
- trilha de auditoria.

## 15. Fase 12 - Operação, segurança e produção

**Tamanho:** XL  
**Dependência:** módulos do release concluídos

### 15.1 Backup e restore

- `pg_dump -Fc`;
- backup pré-update;
- manifesto e SHA-256;
- criptografia autenticada, globals/roles necessários e chave protegida por DPAPI;
- retenção;
- restauração testada;
- validação pós-restore;
- UI simples para backup manual e restore controlado.

### 15.2 Atualização

- pacote assinado;
- hash e assinatura verificados;
- backup;
- migrations;
- smoke test;
- rollback de binário somente quando compatível com o schema; caso contrário, restauração atômica de banco/roles/binários;
- dados preservados.

### 15.3 Segurança

- threat model;
- menor privilégio;
- scan de dependências e secrets;
- LGPD;
- política de retenção, pseudonimização e redação auditável de RAW/backups;
- testes de Named Pipe/ACL;
- matriz de autorização no worker, vínculo SID -> usuário e testes negativos por ação material;
- logs sanitizados;
- assinatura Authenticode;
- SBOM e third-party notices.

### 15.4 Matriz Windows

Testar em VMs limpas:

- Windows 10 x64;
- Windows 11 x64;
- WebView2 ausente;
- PostgreSQL preexistente;
- porta ocupada;
- instalação online/offline;
- reboot/suspensão;
- máquina sem internet;
- UAC negado;
- upgrade e uninstall;
- antivírus/SmartScreen;
- falha de migration;
- backup e restore.

Além do caminho limpo, executar a matriz versionada em `docs/WINDOWS-INSTALLATION-RESILIENCE.md`, cobrindo preflight, instalação interrompida, repair, reinstall, upgrade, rollback conjunto, arquivos bloqueados, disco cheio, reboot pendente, ACL/SID/DPAPI, serviço que não inicia, corrupção de dados, proxy/TLS e diagnóstico offline. Cada caso deve registrar entrada, fase da falha, código esperado, estado final da máquina, preservação de dados e evidência.

### 15.5 Garantias de falha e suporte

- journal de instalação sem secrets, com fase confirmada somente após commit;
- rollback em ordem reversa para artefatos criados pela tentativa atual;
- dados existentes e backups nunca entram no rollback automático destrutivo;
- locks contra duas instalações/upgrades concorrentes;
- códigos estáveis separados de texto traduzido;
- coleta de diagnóstico local, sanitizada e revisável antes do envio;
- botão de retry somente para fases idempotentes e condições já corrigidas;
- repair reconcilia estado desejado versus observado sem recriar identidade DPAPI;
- falha desconhecida termina bloqueada, preserva evidência e orienta suporte.

## 16. Critério de pronto para produção

O produto só será classificado como pronto quando:

- instalador assinado executar em Windows limpo;
- aplicação abrir por atalho sem terminal;
- first-run configurar as integrações suportadas;
- dados persistirem após reboot e atualização;
- cálculos aprovados possuírem testes e lineage;
- critérios financeiros do handoff forem atendidos;
- backup e restauração forem realmente executados;
- falha de API preservar último dado válido;
- testes automáticos, funcionais, negativos e de instalação passarem;
- vulnerabilidades críticas/altas conhecidas estiverem corrigidas ou formalmente aceitas;
- documentação de instalação, operação, atualização, backup e recuperação estiver concluída;
- riscos externos remanescentes estiverem explicitamente aceitos.
- matriz de casos executáveis cobrir timezone, limites inclusivos, parcial, cancelamento, estorno, chargeback, atraso e atualização tardia, com payload/oráculo e saída esperada.
- a suíte obrigatória terá casos nominados de duplicidade, idempotência, datas, valores negativos, cancelamentos, estornos, recebível vencido, obrigação vencida, ausência de sincronização, mudança de status, projeção diária, mudança de cenário, cálculo de aplicação e arredondamento.

## 17. Caminho crítico e bloqueios externos

1. Aprovação dos cinco artefatos.
2. Cadastro de aplicativos Bling e Nuvemshop e redirect URIs aceitas.
3. Credenciais reais de teste fornecidas pelo usuário por meio seguro, nunca no chat/código.
4. Confirmação da fonte oficial Nuvem Pago ou fallback aprovado.
5. Aprovação de moeda, timezone e arredondamento.
6. Regra B2B e mapeamento DRE/P&L.
7. Modelo de referência da aplicação financeira.
8. Certificado/serviço de assinatura de código.
9. Ambiente de teste Windows 10/11.

Esses itens não impedem o desenvolvimento da fundação e dos módulos independentes, mas condicionam o aceite final de produção.
