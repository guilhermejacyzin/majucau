# Evidência T06 — seletor de competência

**Data:** 2026-11-20
**Branch:** `codex/majucau-artifact-consolidation`
**Situação:** implementação parcial; T06 não homologada.

## Alteração verificada

- A T06 apresenta seletor de mês (`type="month"`), iniciado em setembro de 2026, mês indicado como base de comparação.
- Ao trocar a competência, os rótulos dos cards e o estado vazio da tabela acompanham o mês escolhido.
- A tabela deixa explícito que não há dados importados/conferidos para o mês selecionado e que não reutiliza valores de outra competência.
- Foi adicionado um teste de interface para a mudança de setembro para outubro de 2026.

## Limite atual

O seletor ainda não consulta um snapshot DRE nem aciona cálculo. Os cards e valores da tabela permanecem sem números; a tabela lista as linhas previstas, mas ainda não exibe lançamentos de origem. Portanto, essa alteração valida o mês e a regra/composição informativa; não entrega resultados financeiros nem a memória da execução da DRE.

## Regra e composição por linha

- A tabela agora lista as linhas do demonstrativo previsto e cada linha abre a regra/fonte que deve formar aquele resultado. A visão identifica as categorias de Receita Bruta e Deduções descritas no manual e explicita que o pagamento feito em novembro se refere à competência outubro.
- Em Tributos sobre o Lucro, o detalhe declara que a estimativa depende de pagamentos já feitos e da regra original validada; a fórmula não é inventada enquanto aguarda o especialista.
- O detalhe agora é uma **memória informativa da regra**, em blocos de fórmula, fontes, resultado da competência e situação. Em tela estreita, os blocos empilham sem cortar o conteúdo. Isso não é ainda a memória de uma execução financeira: não carrega valores, documentos, parcelas ou resultados calculados e não substitui a memória imutável de origem/cálculo que será conectada ao banco e à execução da DRE.

A regra de estimativa de tributos para a competência correta e sua memória de cálculo aguardam a fórmula e orientação do especialista que Gisele consultará. Nenhuma fórmula tributária foi criada nesta entrega.

## Incremento da memória aritmética

- `CalculateDRE` passou a devolver sete passos aritméticos para receita líquida, lucro bruto, resultado operacional, resultado antes de impostos, resultado líquido, EBIT e EBITDA.
- Cada passo registra os códigos e valores dos operandos, a operação, o resultado e a disponibilidade. Ausência de entrada continua propagando indisponibilidade; não vira zero.
- O teste do motor confere operandos e fechamento do resultado líquido e confirma que falta de CMV mantém indisponíveis os resultados dependentes.
- A migration `000025_dre_calculation_lineage.up.sql` permite que linha indisponível tenha `amount = NULL` e cria armazenamento para fontes RAW referenciadas, contribuições incluídas/excluídas, custos de estoque/venda, ajustes aprovados, passos e operandos. Estados indisponíveis exigem motivo e não carregam valor numérico. FKs compostas protegem a associação à mesma execução/snapshot e triggers tornam o histórico append-only. A view de capital de giro foi ajustada para manter os saldos derivados indisponíveis se faltar uma linha necessária.
- A migration ainda não foi aplicada ao banco local. O worker não grava nem consulta essas tabelas; a T06 ainda não apresenta a memória de uma execução real. O banco PostgreSQL local/daemon não estava disponível para aplicação; a CI aplica todas as migrations em PostgreSQL descartável. Não comprova valores reais nem libera confirmação da DRE.
- A regra fiscal e sua memória específica (pagamentos base, competência, fórmula, etapas e arredondamentos) permanecem pendentes da orientação especializada; nenhuma regra foi presumida.

## Validação executada

- `go test ./internal/financial`: passou.
- `go test ./database/...`: passou; a validação PostgreSQL específica da migration foi ignorada localmente porque `MAJUCAU_TEST_DATABASE_URL` não está configurada. O teste opt-in confere que resultado indisponível não armazena número, que a view de capital de giro não transforma ausência em zero e que memória gravada não aceita atualização; a CI aplica as migrations em PostgreSQL descartável.
- `go test ./...` e `go vet ./...`: passaram localmente; integração PostgreSQL opt-in ficou condicionada à CI.
- `scripts/verify-migrations.ps1`: passou para o manifesto com a migration 000025.
- `sqlc generate -f database/sqlc.yaml`: passou; nenhuma query de escrita/leitura da DRE foi adicionada ainda.
- `npm run typecheck`: passou.
- `npm run test -- --run src/App.test.tsx -t "aplica o mês selecionado aos rótulos"`: passou (1 teste).
- `npm run test -- --run src/App.test.tsx`: passou (39 testes). Dois casos existentes da T04 presumiam uma data inicial diferente do padrão e foram corrigidos para selecionar explicitamente as datas que pretendem conferir; a regra da tela não foi alterada.
- `npm run build`: passou.
- Incremento da regra por linha: `npm run typecheck`, `npm run test -- --run src/App.test.tsx -t "abre a memória da linha da DRE"` (1 teste) e `npm run build`: passaram. A suíte completa anterior do mesmo incremento teve 90 testes aprovados; não substitui conferência dos valores reais.
- Apresentação da memória informativa em blocos responsivos: typecheck, teste direcionado, build e suíte completa (90 testes) passaram; a conferência visual mostrou fórmula, fontes, estado e ausência de resultado em tela estreita.

## Próxima implementação

Ligar a competência selecionada ao carregamento/calculador DRE, importar as fontes disponíveis com origem e hash, persistir contribuições e memória de cálculo por linha, e substituir a regra estática pela memória daquela execução nos detalhes. Bloquear CONFIRMADO quando estiverem ausentes fontes exigidas, âncora de CMV, classificação ou apuração fiscal. Revalidar T04 separadamente, preservando as regras já aprovadas.

## Atualização posterior — persistência da memória aritmética

- O pacote `internal/drememory` agora cria uma execução e um snapshot DRE em uma transação, grava linhas, expressões, operandos, valores disponíveis e indisponibilidade explícita; a leitura recupera os mesmos dados por snapshot.
- A finalização da execução usa a função restrita `finish_dre_calculation`; o runtime não recebe permissão geral de atualização de execuções. O acesso às tabelas DRE está explicitado na allowlist de privilégios e no verificador.
- Foram adicionados testes unitários para expressão, rótulos e validação do hash, além de um teste de integração PostgreSQL que cria uma fixture sintética e verifica leitura, operandos, hash e ausência de resultado quando falta tributo. `go test ./...` passou; o teste PostgreSQL foi ignorado localmente porque `MAJUCAU_TEST_DATABASE_URL` não está configurada. A verificação de checksums das migrations e limites de dependência passou.
- Limite atual: ainda não há classificador/importador que monte fontes e contribuições DRE, a migration 000026 não foi aplicada ao banco local, o worker não chama o repositório e a T06 ainda não consulta a memória persistida. Esta entrega é apenas a base transacional da trilha aritmética, não memória de uma apuração real. Tributos continuam sem fórmula até orientação especializada.
