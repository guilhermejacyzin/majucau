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

O seletor ainda não consulta um snapshot DRE nem aciona cálculo. Os cards permanecem sem valores, a tabela permanece vazia e os detalhes não mostram lançamentos de origem. Portanto, essa alteração valida somente a seleção e o contexto visual do mês; não entrega resultados financeiros, memória de cálculo nem a DRE funcional pedida.

A regra de estimativa de tributos para a competência correta e sua memória de cálculo aguardam a fórmula e orientação do especialista que Gisele consultará. Nenhuma fórmula tributária foi criada nesta entrega.

## Validação executada

- `npm run typecheck`: passou.
- `npm run test -- --run src/App.test.tsx -t "aplica o mês selecionado aos rótulos"`: passou (1 teste).
- `npm run test -- --run src/App.test.tsx`: passou (39 testes). Dois casos existentes da T04 presumiam uma data inicial diferente do padrão e foram corrigidos para selecionar explicitamente as datas que pretendem conferir; a regra da tela não foi alterada.

## Próxima implementação

Ligar a competência selecionada ao carregamento/calculador DRE, importar as fontes disponíveis com origem e hash, persistir contribuições e memória de cálculo por linha, e mostrar seus detalhes. Bloquear CONFIRMADO quando estiverem ausentes fontes exigidas, âncora de CMV, classificação ou apuração fiscal. Revalidar T04 separadamente, preservando as regras já aprovadas.
