# Limites da arquitetura Go

Revisado em 13/10/2026 durante o avanço de BK-017.

Mapa técnico inicial dos sete contextos de negócio: [`DOMAIN-CONTEXT-MAP-v1.md`](DOMAIN-CONTEXT-MAP-v1.md). O mapa registra lacunas sem aprovar novos cálculos, regras ou mudanças de interface.

## Objetivo

Este documento registra como o monólito Go separa regras, casos de uso, integrações e transporte. O objetivo é permitir trocar uma fonte externa ou um meio de armazenamento sem levar dependências de infraestrutura para as regras financeiras.

Este registro não aprova mudanças nas telas nem nas regras financeiras. Ele descreve o código existente e uma correção de propriedade do recurso do banco.

## Mapa dos módulos

| Parte | Local | Responsabilidade | Dependências permitidas |
| --- | --- | --- | --- |
| Domínio | `internal/domain` | Valores monetários exatos, identidade externa, estados e data de negócio | Biblioteca padrão do Go e dados de fuso embutidos |
| Cálculos financeiros | `internal/financial` | Cálculos puros que recebem dados e devolvem resultados | `internal/domain` e biblioteca padrão |
| Contratos da aplicação | `internal/application` | Formatos de entrada/saída e contratos usados entre worker, painel e adaptadores | `internal/domain` e biblioteca padrão |
| Adaptadores | `internal/integrations/*`, `internal/dashboard`, `internal/security`, `internal/backup` | Bling, arquivos, PostgreSQL, cofre de segredos, exportação e leitura do painel | Contratos internos, domínio e bibliotecas/serviços externos necessários ao adaptador |
| Aplicativo desktop e limite Wails | `main.go`, `app.go` | Abrir a janela, publicar os métodos da interface e encaminhar pedidos locais ao worker por IPC | Pode usar Wails, contratos de `application`/`domain` e cliente IPC; não guarda credenciais de provedor nem compõe adaptadores de dados |
| Entrada IPC e composição do worker | `cmd/worker` | Receber pedidos locais, validar o envelope, ligar adaptadores aos contratos e administrar o ciclo de vida | Pode conhecer os adaptadores para montá-los; nenhum pacote depende dos pontos de entrada em `cmd/` |
| Outros pontos de entrada | `cmd/installer-helper` | Compor os serviços da instalação | Adaptadores de instalação e biblioteca padrão |

O `frontend` é uma camada de apresentação separada. Ele fala com o worker pelo IPC; não importa pacotes Go nem acessa o banco diretamente.

## Caminho permitido das dependências

```text
main.go/app.go (Wails) → frontend → IPC → cmd/worker (composição)
                                             ↓
                   contratos → adaptadores → PostgreSQL/Bling/arquivos/cofre
                       ↓
                aplicação → domínio ← cálculos financeiros
```

Regras para mudanças futuras:

1. `internal/domain` não importa HTTP, React, PostgreSQL/ORM, Bling nem pacotes de `cmd`.
2. `internal/financial` pode usar o domínio; não pode chamar rede, interface visual ou banco.
3. A aplicação não importa adaptadores concretos. As interfaces são definidas pelo consumidor do serviço sempre que isso preserva a dependência para dentro.
4. Adaptadores podem implementar contratos da aplicação e usar drivers externos. Um adaptador não deve chamar a interface visual.
5. Só os pontos de composição — aplicativo Wails na raiz, `cmd/worker` e `cmd/installer-helper` — ligam seus serviços locais. Nenhum pacote importa um pacote de `cmd/`.
6. Payloads de arquivos, rede e IPC devem continuar em fluxo e com limites definidos nos próprios adaptadores; os contratos de controle permanecem pequenos.

## Portas e adaptadores já existentes

| Necessidade | Contrato atual | Implementação/adaptador |
| --- | --- | --- |
| Sincronizar ou importar lançamentos | Interfaces consumidoras no `cmd/worker/handler.go`: importação Bling, arquivo futuro, configuração, OAuth e sincronização | Serviços de `internal/integrations/bling`, `nuvempago` e `nuvemshop` |
| Consultar painel em fluxo | `application.DashboardSnapshotStreamer` | `dashboard.Reader`, com leitura PostgreSQL em fluxo |
| Consultar estado das integrações | `application.IntegrationStatusReader` | Leitor PostgreSQL montado em `cmd/worker/database.go` |
| Verificar saúde | `application.HealthChecker` | `application.StaticHealth` |
| Persistir estado OAuth Bling | `blingOAuthRepository` | `postgresBlingOAuthRepository` |
| Proteger segredos | `security.SecretStore` | Armazenamento protegido pelo Windows DPAPI |
| Chamar serviço HTTP | `bling.HTTPDoer` | Cliente HTTP configurado para a API Bling |
| Ler/gravar conteúdo em fluxo | `io.Reader` e `io.Writer` | IPC, API, arquivos CSV, painel e exportação |
| Enfileirar arquivos e deduplicar importação | `CSVFileQueue` e `SourceIDIndex` | Implementações em disco/temporárias de `internal/integrations/csvstream` |
| Obter a hora para a consulta do painel | função `now func() time.Time` injetável no `dashboard.Reader` | Relógio UTC do processo |

As interfaces estão próximas de seus consumidores. Algumas interfaces de banco ainda usam tipos do driver PostgreSQL; isso é um limite do adaptador atual e não permite que o domínio dependa do banco.

## Composição e ciclo de vida

`newWorkerHandler()` é a raiz de composição do worker. Ela cria um único pool PostgreSQL e, quando disponível, o cofre de segredos e os serviços Bling/Nuvemshop/Nuvem Pago. O handler guarda os serviços por interfaces para que possam ser substituídos em verificações locais.

O worker agora também é dono explícito do pool compartilhado e o fecha ao encerrar. Os serviços de importação não fecham mais, cada um por conta própria, um pool que não criaram. Isso mantém a responsabilidade de abertura e fechamento junto ao ponto que criou o recurso.

## Fronteiras de transação observadas

| Fluxo | Fronteira e resultado |
| --- | --- |
| Sincronização de páginas da API Bling | Uma transação por recurso e janela. Páginas RAW, cursor e conclusão bem-sucedida são gravados juntos. Um savepoint permite desfazer os dados/cursor em caso de falha e ainda registrar o lote como falho. |
| Importação CSV de recebimentos Bling | Uma transação por importação. Linhas válidas são persistidas; linhas inválidas são contadas e o lote fica parcial, conforme regra já aprovada. Erro estrutural ou de banco aborta a transação. |
| Importação CSV de recebimentos futuros Nuvem Pago | Uma transação por importação, com as mesmas fronteiras de linhas válidas e erros de linha parciais. |
| Leitura do painel | Uma transação somente leitura, `REPEATABLE READ`, para manter resumo e detalhes na mesma visão; os detalhes são emitidos um por vez. |

As transações pertencem aos serviços que persistem os dados. A raiz de composição cria e fecha conexões/pools; ela não decide o que é um registro válido nem altera os resultados financeiros.

## Verificação automática e limite conhecido

Em 12/10/2026, a lista de imports de Go foi conferida com `go list`: domínio e cálculos financeiros não dependiam de integrações/HTTP/React/ORM, e não havia ciclo entre os módulos listados. Os testes existentes de `cmd/worker` usam implementações falsas das interfaces consumidoras.

Em 13/10/2026, `scripts/verify-go-boundaries.ps1` passou a conferir todos os pacotes do módulo pelo grafo real obtido por `go list`, usando a lista oficial de pacotes padrão da versão ativa do Go. A verificação exige biblioteca padrão em `internal/domain`, permite apenas biblioteca padrão e `internal/domain` em `internal/financial`/`internal/application`, rejeita pacotes de rede, banco, sistema operacional, caminhos de arquivo, chamadas de sistema, plugins e `unsafe` nessas três camadas e impede dependências de qualquer pacote `cmd/`. O script roda na verificação local e no job Windows da CI.

A verificação confirma direção de imports, não propriedade final de todos os fatos do negócio. A definição dos contextos, agregados, eventos e contratos funcionais segue em BK-016, e a implementação de casos de uso ainda depende do restante do backlog. A regra não altera interfaces, dados, telas ou regras financeiras.
