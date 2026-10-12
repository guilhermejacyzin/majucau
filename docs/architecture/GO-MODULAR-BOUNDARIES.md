# Limites da arquitetura Go

Revisado em 12/10/2026 durante o fechamento de BK-017.

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
| Entrada IPC e composição | `cmd/worker` | Receber pedidos locais, validar o envelope, ligar adaptadores aos contratos e administrar o ciclo de vida | Pode conhecer os adaptadores para montá-los; o restante do projeto não deve depender de `cmd/worker` |
| Outros pontos de entrada | `cmd/installer-helper` | Compor os serviços da instalação | Adaptadores de instalação e biblioteca padrão |

O `frontend` é uma camada de apresentação separada. Ele fala com o worker pelo IPC; não importa pacotes Go nem acessa o banco diretamente.

## Caminho permitido das dependências

```text
frontend → IPC → cmd/worker (composição) → contratos → adaptadores → PostgreSQL/Bling/arquivos/cofre
                                  ↓
                         aplicação e domínio
```

Regras para mudanças futuras:

1. `internal/domain` não importa HTTP, React, PostgreSQL/ORM, Bling nem pacotes de `cmd`.
2. `internal/financial` pode usar o domínio; não pode chamar rede, interface visual ou banco.
3. A aplicação não importa adaptadores concretos. As interfaces são definidas pelo consumidor do serviço sempre que isso preserva a dependência para dentro.
4. Adaptadores podem implementar contratos da aplicação e usar drivers externos. Um adaptador não deve chamar a interface visual.
5. Só os pontos de composição (`cmd/*`) ligam implementações concretas aos consumidores.
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

## Verificação e limite conhecido

Em 12/10/2026, a lista de imports de Go foi conferida com `go list`: domínio e cálculos financeiros não dependem de integrações/HTTP/React/ORM, e não há ciclo entre os módulos listados. Os testes existentes de `cmd/worker` usam implementações falsas das interfaces consumidoras.

O repositório ainda não tem uma regra automática de CI que fiscalize toda a direção dos imports. Mudanças futuras devem respeitar este mapa e registrar desvios antes de introduzir acoplamento. A existência desse limite não muda as interfaces ou os dados atuais.
