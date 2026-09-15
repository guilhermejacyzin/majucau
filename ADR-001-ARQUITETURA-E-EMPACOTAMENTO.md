# ADR-001 - Arquitetura, segurança e empacotamento Windows

- **Status:** Aprovado no G0 em 2026-08-16; implementação sujeita aos gates G1–G7
- **Data:** 2026-08-15
- **Produto:** Majucau Financial Intelligence
- **Plataforma:** Windows 10/11 x64, single-user, instalação local
- **Criticidade:** Alta

## 1. Contexto

O produto consolidará informações operacionais, financeiras e contábeis provenientes de Bling, Nuvemshop/Nuvem Pago, folha e ajustes autorizados. O sistema deverá manter dados RAW imutáveis, impedir duplicidade, separar caixa de competência e realizado de projetado, preservar rastreabilidade até a origem e produzir cálculos financeiros auditáveis.

O usuário final não utilizará terminal, Docker ou ferramentas de banco de dados. Após executar o instalador e concluir a configuração guiada, deverá abrir o aplicativo por um atalho e utilizá-lo como um programa Windows normal.

## 2. Requisitos bloqueados

1. Frontend em React + TypeScript.
2. Backend e processamento em Go.
3. PostgreSQL local e isolado.
4. Aplicativo desktop single-user para Windows 10/11 x64.
5. Instalador nativo, sem Docker.
6. Sincronizações e cálculos devem continuar quando a janela estiver fechada.
7. Tela de integrações para conectar Bling e Nuvemshop/Nuvem Pago.
8. Tooltips simples e acessíveis em todos os componentes funcionais.
9. Nenhum secret ou token no frontend, código, logs ou backup comum.
10. Nenhum dado calculado pode perder a linhagem até o registro RAW.

## 3. Decisão arquitetural

Adotar uma aplicação desktop composta por três processos locais:

```text
majucau.exe
Wails v2 + React + TypeScript + Vite
        |
        | Named Pipe local autenticado e restrito por ACL
        v
majucau-worker.exe
Serviço Windows em Go
        |
        | pgx + SQL gerado por sqlc
        v
PostgreSQL local dedicado
Serviço, porta e cluster exclusivos do produto
```

### 3.1 Aplicativo desktop

Usar Wails v2 estável e fixar a versão no build. Wails v3 não será adotado enquanto permanecer pré-lançamento. O executável `majucau.exe` conterá:

- janela desktop baseada em WebView2;
- bundle React/TypeScript produzido pelo Vite;
- navegação e apresentação;
- validação de entrada no cliente;
- adaptador Go local para comunicação segura com o worker;
- preferências visuais não sensíveis.

O frontend não acessará PostgreSQL nem APIs externas diretamente.

### 3.2 Worker Windows

`majucau-worker.exe` será registrado como serviço Windows e continuará ativo quando a janela for fechada. Responsabilidades:

- sincronização Bling e Nuvemshop;
- processamento de webhooks quando aplicável;
- normalização e deduplicação;
- motor financeiro e snapshots;
- migrations;
- agendamento, retry e catch-up após suspensão/desligamento;
- backups;
- logs técnicos;
- armazenamento e renovação de tokens;
- health checks.

O serviço deverá iniciar após o PostgreSQL e utilizar uma conta local de baixo privilégio.

### 3.3 Comunicação local

Usar Named Pipe do Windows, não uma porta HTTP pública, para comunicação entre desktop e worker.

Requisitos:

- DACL restrita ao usuário Windows autorizado, à conta do serviço e a administradores locais;
- protocolo versionado;
- limite de tamanho de mensagem;
- timeouts e cancelamento;
- validação de todos os comandos no worker;
- nenhuma credencial em argumentos de processo;
- logs sem payloads sensíveis.

### 3.4 PostgreSQL

Instalar uma instância PostgreSQL própria, sem reutilizar instalações existentes.

- Major proposto: PostgreSQL 17, com o minor suportado mais recente fixado no pacote de release.
- Binários: `%ProgramFiles%\Majucau\PostgreSQL`.
- Dados: `%ProgramData%\Majucau\PostgreSQL\data`.
- Backups: `%ProgramData%\Majucau\Backups`.
- Escuta: somente `127.0.0.1`.
- Porta: exclusiva, detectada e registrada durante a instalação.
- Autenticação: SCRAM.
- Usuário da aplicação: sem superuser, sem criação de roles/databases.
- Conta do serviço PostgreSQL: dedicada e de baixo privilégio.

Valores monetários usarão `NUMERIC`, nunca ponto flutuante. Proposta inicial:

- valores monetários: `numeric(19,4)`;
- taxas e percentuais: `numeric(20,8)`;
- moeda padrão proposta: BRL;
- timezone de negócio proposto: `America/Sao_Paulo`;
- timestamps técnicos: `timestamptz` em UTC;
- datas de negócio: `date`.

Moeda, timezone, escala e arredondamento permanecem sujeitos à aprovação funcional antes do motor financeiro.

### 3.5 Acesso a dados

Usar SQL explícito, `pgx/v5` e `sqlc`.

- schemas e queries versionados;
- geração tipada no CI;
- transactions explícitas;
- constraints de idempotência no banco;
- migrations SQL forward-only, com checksum e advisory lock;
- backup obrigatório antes de migration de release;
- sem ORM amplo como camada principal.

## 4. Organização do repositório

```text
majucau-financial-intelligence/
├── AGENTS.md
├── README.md
├── go.mod
├── wails.json
├── cmd/
│   ├── desktop/
│   └── worker/
├── internal/
│   ├── domain/
│   ├── application/
│   ├── integrations/
│   │   ├── bling/
│   │   ├── nuvemshop/
│   │   └── nuvempago/
│   ├── financial/
│   ├── accounting/
│   ├── security/
│   ├── storage/
│   └── ipc/
├── frontend/
│   ├── src/
│   └── dist/
├── database/
│   ├── migrations/
│   ├── queries/
│   └── generated/
├── docs/
├── reference-models/
├── scripts/
├── tests/
└── build/windows/
```

## 5. Frontend e experiência do usuário

### 5.1 Stack

- React + TypeScript;
- Vite;
- React Router;
- TanStack Query para estado remoto e cache de leitura;
- Radix Primitives como base de componentes acessíveis;
- biblioteca de gráficos escolhida após um spike de acessibilidade e exportação.

O build executará `tsc --noEmit` separadamente, porque Vite transpila TypeScript, mas não substitui o typecheck.

### 5.2 Política de tooltips

Todo componente funcional deverá possuir explicação contextual simples:

- card ou indicador;
- gráfico e série;
- filtro;
- campo de formulário;
- botão ou ação;
- status;
- alerta;
- coluna ou valor financeiro que possa gerar dúvida.

Não aplicar tooltip a divisores e elementos puramente decorativos.

Cada tooltip responderá, quando aplicável:

1. **O que é?**
2. **Como interpretar ou usar?**
3. **Qual é a fonte ou condição importante?**

Regras:

- texto simples e curto;
- nenhuma regra crítica ficará exclusivamente no tooltip;
- abertura por hover e foco;
- fechamento por `Escape`, blur e saída do ponteiro;
- `role="tooltip"` e `aria-describedby`;
- foco permanece no elemento disparador;
- labels visíveis continuam obrigatórios em formulários;
- termos financeiros terão explicação não técnica. Tooltips conterão somente texto não interativo; “Saiba mais” e qualquer ação abrirão um popover/dialog acessível separado, com gerenciamento próprio de foco.

Exemplo:

```text
Saldo Projetado em 60 Dias
Estimativa do dinheiro disponível ao final dos próximos 60 dias,
considerando entradas e saídas previstas no cenário selecionado.
```

## 6. Tela Configurações > Integrações

Criar três painéis independentes: Bling, Nuvemshop e Nuvem Pago.

### 6.1 Estados comuns

- `NOT_CONFIGURED`;
- `AUTHORIZING`;
- `CONNECTED`;
- `TOKEN_EXPIRING`;
- `AUTH_ERROR`;
- `SYNCING`;
- `STALE`;
- `PARTIALLY_AVAILABLE`;
- `UNAVAILABLE`;
- `SCHEMA_MISMATCH`.

Cada painel mostrará:

- explicação simples da integração;
- campos de configuração mascarados;
- botão Conectar/Reconectar;
- botão Testar conexão;
- status;
- escopos autorizados;
- loja/conta identificada;
- última sincronização bem-sucedida;
- última tentativa;
- erro resumido em linguagem simples;
- acesso ao detalhe técnico sanitizado;
- botão Desconectar com confirmação.

### 6.2 Bling

Campos/configurações:

- client ID;
- client secret;
- redirect URI registrada;
- scopes esperados;
- estado da autorização.

Fluxo:

1. usuário informa credenciais do aplicativo Bling;
2. desktop envia os valores ao worker pelo Named Pipe;
3. worker protege os secrets;
4. “Conectar” abre o navegador do sistema;
5. Authorization Code é recebido pelo callback aprovado;
6. worker troca o código por tokens;
7. access e refresh tokens são protegidos;
8. uma chamada somente leitura testa o acesso;
9. primeira sincronização ocorre sob confirmação do usuário.

### 6.3 Nuvemshop

Campos/configurações:

- app ID/client ID;
- client secret;
- redirect URI;
- store ID retornado pela autorização;
- scopes esperados.

O conector de produção adotará um callback HTTPS intermediário de código mínimo, sem banco financeiro e sem webhook público na máquina do usuário. O desktop cria uma sessão efêmera com `state` criptograficamente aleatório e um segredo de pareamento; o relay recebe apenas `code/state`, permite leitura única pelo desktop autenticado e elimina a sessão em até cinco minutos. A troca do código por token ocorre no worker local usando o client secret protegido por DPAPI. O relay não recebe client secret nem access token.

Esse desenho é uma decisão arquitetural, mas permanece bloqueado até: (a) registrar o redirect HTTPS exato em um aplicativo real; (b) obter confirmação do provedor sobre o fluxo para aplicativo desktop; (c) validar `state`, expiração, uso único, replay, cancelamento e teste ponta a ponta; e (d) verificar se PKCE é aceito pelo endpoint do provedor. Se PKCE não for suportado, o aceite de risco deverá ser explícito e o retorno relay -> desktop usará challenge/verifier próprio. Não haverá fluxo manual improvisado, captura de token por clipboard nem client secret embutido no instalador.

Sem PKCE do provedor existe risco residual de interceptação/reuso do authorization code antes da troca, especialmente se sessão, navegador ou relay forem comprometidos. `state`, TTL, uso único, TLS e challenge próprio no retorno relay -> desktop reduzem riscos de CSRF, replay e entrega ao cliente errado, mas não são equivalentes ao binding criptográfico do code verifier no token endpoint. Portanto, ausência de PKCE mantém o conector bloqueado até threat model específico e decisão `D-002-RISK` com responsável, prazo, mitigação e aceite explícito.

### 6.4 Nuvem Pago

A documentação pública consultada não comprova API de saldo, taxas, repasses e agenda de recebíveis para o lojista.

Consequências:

- não reutilizar token Nuvemshop como se autorizasse ledger do Nuvem Pago;
- não afirmar `fee_amount`, `net_amount` ou `expected_receipt_date` sem fonte oficial;
- exibir `PARTIALLY_AVAILABLE` enquanto somente pedidos/status estiverem acessíveis;
- preparar interface de adapter para API/arquivo oficial futuro;
- dados incompletos serão `PROVISIONAL` ou `PARTIALLY_CONFIRMED`, nunca `CONFIRMED`.

## 7. Segurança de credenciais

O worker criptografará tokens e client secrets com Windows DPAPI no contexto da conta dedicada do serviço. O blob protegido ficará em arquivo com ACL restrita.

A identidade será uma conta virtual de serviço `NT SERVICE\MajucauWorker`, com SID de serviço estável e perfil carregado. O instalador será idempotente: cria/atualiza o serviço sem trocar sua identidade, aplica SDDL explícita aos diretórios, ao Named Pipe e aos secrets, e usa DPAPI no escopo da máquina com entropy aleatória por instalação protegida pela mesma ACL. O banco guarda apenas `secret_ref` e a versão do envelope. Reinstalação/upgrade preserva a identidade e o diretório de secrets; mudança de conta exige fluxo administrativo de reproteção ou reconexão.

- secrets nunca serão persistidos no PostgreSQL em texto aberto;
- campos não retornarão o valor completo à UI após salvamento;
- logs registrarão somente identificadores e últimos quatro caracteres quando necessário;
- clipboard não será usado automaticamente;
- tokens não entrarão no backup padrão;
- restauração em outra máquina exigirá reconexão das integrações;
- falha de descriptografia bloqueará sincronização sem destruir o último dado válido.
- desconectar elimina tokens e sessões; desinstalar preserva o cofre somente quando os dados são preservados e oferece exclusão explícita conjunta;
- testes obrigatórios cobrem reboot, upgrade, repair, troca rejeitada de identidade, desinstalação/reinstalação e UX de reconexão.

Administradores locais continuam dentro do trust boundary e podem acessar a máquina. Essa limitação deve constar no modelo de ameaça.

## 8. Sincronização e resiliência

- polling incremental para objetos financeiros;
- webhooks apenas onde houver suporte oficial;
- inbox idempotente para eventos;
- `source_system + source_entity + source_id` como base de unicidade RAW;
- cursor/checkpoint por recurso;
- backoff exponencial com jitter para 429 e falhas transitórias;
- último dado válido preservado em falha;
- dashboard exibe “Dados desatualizados” sem zerar indicadores;
- catch-up após boot, suspensão ou período offline;
- nenhuma execução concorrente da mesma integração/recurso.

## 9. Instalação Windows

Usar NSIS com instalação elevada e pacote x64 assinado.

O instalador deverá:

1. validar Windows 10/11 x64;
2. verificar/instalar WebView2 Evergreen;
3. instalar PostgreSQL dedicado;
4. reservar sob lock de instalação uma porta livre da faixa privada do produto e abortar de forma recuperável se houver corrida;
5. criar cluster, roles e database com senhas aleatórias de 256 bits, sem exibi-las em log/linha de comando, e proteger a credencial do worker via DPAPI;
6. configurar `listen_addresses=127.0.0.1`, SCRAM, `pg_hba.conf` mínimo e ACLs/SDDL;
7. registrar dependência real `MajucauWorker -> MajucauPostgreSQL`, aguardar readiness e executar migrations sob advisory lock antes de liberar o worker;
8. instalar desktop e worker;
9. criar atalhos;
10. executar smoke test;
11. abrir o first-run wizard.

Dados mutáveis não serão gravados em `Program Files`.

### 9.1 Assinatura

Assinar executáveis próprios e instalador com Authenticode, SHA-256 e timestamp. A release deverá falhar se a verificação da assinatura falhar.

### 9.2 Atualização

- aceitar apenas pacote assinado e hash esperado;
- criar backup antes da migration;
- parar desktop/worker de forma ordenada;
- migrations forward-only;
- smoke test após atualização;
- rollback de binário somente dentro da janela de compatibilidade declarada entre binário e schema;
- restauração de banco somente por procedimento controlado.

Cada release declarará `min_schema_version` e `max_schema_version`. Se uma migration romper compatibilidade com o binário anterior, não haverá rollback isolado: o procedimento restaurará banco, roles e binários do mesmo conjunto, a partir do backup pré-update testado.

### 9.3 Desinstalação

Remover aplicativos e serviços. Preservar dados por padrão. Excluir banco e backups somente mediante opção explícita, confirmação reforçada e recomendação de backup.

## 10. Backup e recuperação

- `pg_dump -Fc` agendado;
- `pg_dumpall --globals-only` para roles/metadados necessários à recuperação, sem registrar senhas em logs;
- backup antes de updates e migrations;
- arquivo temporário seguido de rename atômico;
- SHA-256 e manifesto;
- criptografia autenticada do pacote de backup com chave protegida por DPAPI e opção de exportação com senha definida pelo usuário;
- retenção configurável;
- teste periódico de restauração;
- sincronizações pausadas durante restore;
- credenciais excluídas do backup;
- validação de schema, constraints, RAW, snapshots e continuidade após restore.

Nunca copiar diretamente o diretório de dados com PostgreSQL ativo.

Decisão de exceção: RAW é append-only para sincronização, cálculo e correção operacional, mas não é imutável contra obrigação legal de eliminação de PII. A política de produção definirá prazo por categoria, minimização na ingestão, hash/pseudonimização de documento e uma operação exclusiva de redação auditável: o payload original é substituído por envelope de tombstone contendo hash anterior, motivo, autorização e instante, sem manter o dado pessoal apagado. A redação nunca altera valores financeiros não pessoais nem simula nova versão da fonte. Índices e backups vencidos seguem a mesma política; backups ainda retidos tornam-se inacessíveis ao fim de sua chave/retenção. O procedimento LGPD será testado do RAW ao snapshot sem apagar a prova contábil não pessoal.

## 11. Alternativas rejeitadas

| Alternativa | Motivo |
|---|---|
| Node.js/Next.js no backend | Runtime adicional e SSR sem valor para o desktop local |
| Wails v3 pré-release | Risco incompatível com produção |
| Electron | Maior footprint sem benefício funcional necessário |
| Docker Desktop | Expõe complexidade operacional ao usuário final |
| SQLite | Diverge da stack aprovada e reduz recursos úteis para RAW, worker e auditoria |
| ORM amplo | Oculta SQL crítico e dificulta auditoria financeira |
| Redis/Kafka/Kubernetes | Overengineering para single-user local |
| Jobs somente na UI | Param quando a janela é fechada |
| PostgreSQL existente | Risco de conflito, acoplamento e exposição de dados de terceiros |
| Tokens no banco sem proteção | Exposição de credenciais |

## 12. Consequências e riscos aceitos

- instalação exige UAC;
- PostgreSQL aumenta o tamanho e a complexidade do instalador;
- WebView2 pode precisar ser instalado;
- assinatura de código tem custo e validação de identidade;
- administradores locais permanecem altamente privilegiados;
- Nuvem Pago ainda depende de acesso/documentação oficial;
- fórmula completa da aplicação financeira depende de modelo funcional aprovado;
- a matriz Windows 10/11 deverá ser testada em VMs limpas.

## 13. Critérios de aprovação do ADR

- arquitetura desktop/worker/PostgreSQL aceita;
- stack e principais bibliotecas aceitas;
- política de credenciais aceita;
- política de tooltips aceita;
- limitação atual do Nuvem Pago explicitamente aceita;
- estratégia de instalação, backup, atualização e desinstalação aceita;
- moeda, timezone, escala e arredondamento aprovados antes do motor financeiro.

## 14. Fontes primárias

- [Wails v2](https://wails.io/docs/introduction/)
- [Instalador NSIS no Wails](https://wails.io/docs/guides/windows-installer/)
- [Distribuição do WebView2](https://learn.microsoft.com/microsoft-edge/webview2/concepts/distribution)
- [Windows DPAPI](https://learn.microsoft.com/windows/win32/api/dpapi/nf-dpapi-cryptprotectdata)
- [SmartScreen e assinatura](https://learn.microsoft.com/windows/apps/package-and-deploy/smartscreen-reputation)
- [PostgreSQL para Windows](https://www.postgresql.org/download/windows/)
- [PostgreSQL numeric](https://www.postgresql.org/docs/current/datatype-numeric.html)
- [pgx](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [sqlc](https://docs.sqlc.dev/en/stable/)
- [Vite](https://vite.dev/guide/)
- [WAI-ARIA Tooltip](https://www.w3.org/WAI/ARIA/apg/patterns/tooltip/)
- [OAuth para aplicativos nativos](https://www.rfc-editor.org/info/rfc8252/)
