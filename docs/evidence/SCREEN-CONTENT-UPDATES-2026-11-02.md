# Atualizações aprovadas de conteúdo e legibilidade das telas — 02/11/2026

## Alterações aplicadas no protótipo

| Tela | Alteração | Itens mantidos |
|---|---|---|
| T01 — Visão Executiva | Retirados “Obrigações vencidas aguardando fonte Bling” e “Recebimentos em atraso aguardando conciliação”. | Cards, gráfico, distribuição, outros avisos e navegação. |
| T02 — Fluxo de Caixa | Retirados “Saídas previstas sem categoria”, “Recebimentos em atraso” e “Conciliação pendente”. | Cards e alerta “Saldo projetado abaixo da reserva mínima”. |
| T09 — Estoque | Retirados os cards “ENVELHECENDO” e “VALIDADE” e o alerta “Lotes com validade próxima”; subtítulo “Saldo, cobertura e validade” encurtado para “Saldo e cobertura”. | Estoque físico, utilizável, cobertura, risco de ruptura, segurança, ponto de pedido, giro e outros avisos. |
| T10 — Produção | Retirados “Lotes com perda acima do esperado” e “Separar plano realizado e obrigatório”. | “Ordens em atraso” e “BOM versionada impacta a necessidade de insumos”, demais cards e tabela. |
| T11 — Compras | Os quatro avisos foram mantidos e empilhados em uma coluna para evitar sobreposição e permitir leitura. Nenhum aviso foi removido sem decisão da responsável. | Cards, tabela somente de consulta, restante do layout. |

## O que significam os avisos de Compras

No protótipo atual, estes quatro textos são **mensagens de exemplo**, não erros
detectados na empresa. O subtítulo “Sem dados confirmados” em cada cartão
indica que a fonte e o cálculo ainda não foram ligados. Portanto, a tela não
está afirmando que esses riscos estão ocorrendo agora.

1. **Preço do cacau em queda:** comparação entre cotações de cacau em datas
   diferentes. Para confirmar, precisa de fonte e período de cotação.
2. **Risco climático elevado:** possibilidade de clima afetar o fornecimento.
   Precisa de fonte de risco ligada às regiões dos fornecedores.
3. **Cobertura de manteiga de cacau inferior a 20 dias:** o estoque utilizável
   de manteiga, dividido pelo consumo médio diário, seria menor que 20 dias.
   Precisa de estoque e consumo validados.
4. **Fornecedores com lead time acima da média:** alguns fornecedores levariam
   mais tempo que a média de entrega de referência. Precisa do histórico de
   pedidos e entregas.

Estes avisos foram preservados para a responsável decidir depois de entender o
significado. Nenhum risco real foi confirmado por esta atualização.

## Referências de telas

As imagens em `docs/source/` são as referências aprovadas originais e não foram
substituídas. As diferenças aprovadas nesta data estão registradas aqui e no
código atual em `frontend/src/mockupDashboard.tsx`; a regra de coluna única dos
avisos de Compras está em `frontend/src/mockup.css`. Não se deve tratar os
mockups originais como capturas do protótipo atualizado.

As cinco capturas da versão atual estão em
`docs/evidence/current-screens-2026-11-02/` e são listadas no README dessa pasta.

## Validação

Teste frontend `T01ExecutiveNavigation.test.tsx`: 18 casos aprovados. Build
TypeScript/Vite aprovado usando cópia isolada em `artifacts/verification`, sem
modificar as dependências do protótipo aberto. A validação da tela Wails
instalada/aberta ainda não foi feita nesta atualização.

Atualização posterior: o app Wails e worker foram reiniciados com o snapshot de
90 dias. A consulta sanitizada ao worker confirmou 418 títulos a pagar no
período, em estado parcial. A leitura visual automatizada não foi feita nesta
execução; a janela de Computer Use foi encerrada pelo usuário.
