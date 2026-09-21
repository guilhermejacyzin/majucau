# Evidência — realinhamento ao mocap T1–T15/C11

## Fonte visual obrigatória

As referências entregues pelo usuário foram incorporadas em `docs/source` e são a fonte de verdade visual deste incremento:

- `T1-mockup-2026-09-20.png` — Visão Executiva;
- `T2-mockup-2026-09-20.png` — Fluxo de Caixa;
- `T3-mockup-2026-09-20.png` — Contas a Receber;
- `T4-mockup-2026-09-20.png` — Contas a Pagar;
- `T5-mockup-2026-09-20.png` — Conciliação;
- `T6-mockup-2026-09-20.png` — DRE;
- `T7-mockup-2026-09-20.png` — Balancete;
- `T8-mockup-2026-09-20.png` — Calculadora de Aplicação;
- `T9-mockup-2026-09-20.png` — Estoque;
- `T10-mockup-2026-09-20.png` — Produção;
- `T11-mockup-2026-09-20.png` e `C11-mockup-2026-09-20.png` — Compras;
- `T12-mockup-2026-09-20.png` — Projeções;
- `T13-mockup-2026-09-20.png` — Lançamentos & Pendências;
- `T14-mockup-2026-09-20.png` — Alertas;
- `T15-mockup-2026-09-20.png` — Importações / Integrações.

Os números que aparecem nas imagens são exemplos de composição e não foram hardcoded. A interface mostra `—` e `Sem dados confirmados` enquanto não houver fonte oficial conectada.

O handoff técnico-funcional versionado em `docs/source/Handoff-Tecnico-Funcional-Majucau-2026-08-16.md` foi relido nesta revisão. As regras financeiras do handoff continuam prevalecendo sobre o desenho: o card de menor saldo representa o mínimo diário dentro de D0–D+60, e o painel de semana crítica permanece dependente de decisão/regra, sem converter ausência de dados em zero. Os círculos vermelhos da referência T2 foram tratados como indicação desses dois componentes, não como valores a serem copiados.

## Shell compartilhado implementado

- sidebar azul-marinho com a ilustração de referência, seções, ícones e item ativo em teal;
- cabeçalho com número da tela, título em caixa alta, subtítulo, posição, última sincronização e Exportar; o botão `Ajuda` foi removido conforme a regra funcional registrada;
- quatro cards de indicadores na primeira linha;
- painel central de gráfico/tabela e painel de distribuição/status;
- cards inferiores e bloco `ALERTAS IMPORTANTES`;
- responsividade para a janela Windows, preservando a composição desktop do mocap;
- tooltip simples em todos os botões, campos e seletores funcionais;
- navegação para as quinze telas sem abandonar o contrato atual da tela de integrações.

## Mapeamento funcional

| Tela | Rota visual | Estado atual |
| --- | --- | --- |
| T1 | `executive` | shell e composição executiva |
| T2 | `cashflow` | shell de fluxo de caixa |
| T3 | `receivables` | tabela e distribuição de recebíveis |
| T4 | `payables` | tabela e distribuição de obrigações |
| T5 | `conciliation` | matches e status de conciliação |
| T6 | `dre` | DRE consolidada |
| T7 | `trial-balance` | balancete |
| T8 | `investment-calculator` | cálculo de capacidade |
| T9 | `inventory` | estoque e cobertura |
| T10 | `production` | ordens e capacidade |
| T11/C11 | `purchases` | compras, cacau e risco de abastecimento |
| T12 | `forecast` | forecast, cenários e backtest |
| T13 | `pending` | pendências e decisões |
| T14 | `alerts` | central de alertas |
| T15 | `integrations` | integrações funcionais existentes |

## Validação executada

- `npm run lint --prefix frontend` — aprovado;
- `npm run typecheck --prefix frontend` — aprovado;
- `npm run test:run` em `frontend` — 3 arquivos e 19 testes aprovados;
- a navegação do alerta executivo para Integrações permanece funcional;
- controles do shell são verificados estruturalmente com tooltip nos testes;
- o teste de integração continua cobrindo Bling, Nuvemshop, Nuvem Pago e importação da pasta do Bling.

## Limite honesto deste incremento

O realinhamento visual cobre o shell e a composição de todas as telas de referência. A ligação dos números e tabelas aos dados reais de cada módulo, a validação em máquina Windows limpa e o instalador NSIS continuam no backlog de produção. Nenhuma dessas pendências autoriza inserir números fictícios na tela.
