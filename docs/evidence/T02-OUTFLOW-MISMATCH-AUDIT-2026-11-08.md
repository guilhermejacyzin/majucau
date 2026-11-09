# T02 — auditoria da divergência de despesas por semana

**Data:** 08/11/2026  
**Escopo:** abrir o protótipo, localizar o valor mencionado e rastrear a composição
no código. Nenhuma regra financeira ou tela foi alterada nesta auditoria.

## O que aparece nas telas

O protótipo web em `http://127.0.0.1:5173/` está na T02, com o período
08/11/2026–07/01/2027. Essa prévia não tem snapshot do banco local: saldo,
entradas, saídas, tabela e gráfico aparecem sem dados confirmados. Ela não é a
janela Wails que exibe a série sincronizada.

Na janela Wails, a Visão Executiva mostrou estes pontos:

| Semana do ponto | Despesas no Majucau |
|---|---:|
| 22/11–28/11 | -R$ 27.012,27 |
| 29/11–05/12 | -R$ 101.721,05 |

Portanto, o valor de aproximadamente R$ 101 mil não está no ponto de
22/11–28/11; está no ponto seguinte, 29/11–05/12.

## Comparação que os registros existentes permitem

A evidência do relatório Bling para 08/11–30/11 registra:

- 22/11–28/11: despesas de **-R$ 27.012,27** — igual ao Majucau nessa faixa;
- 29/11–30/11: despesas de **-R$ 9.180,00** e receitas de **R$ 1.068,87**.

Os R$ 1.068,87 são receitas, não despesas. E o trecho 29–30/11 cobre apenas
dois dias, enquanto a faixa do Majucau 29/11–05/12 cobre sete dias; essa
comparação não demonstra sozinha que as despesas dos sete dias estão corretas.

Existe também uma captura anterior do Bling com a faixa inteira 29/11–05/12 em
**-R$ 57.445,14**. Ela foi gerada com **duas contas financeiras selecionadas**
e outras opções registradas na evidência de 02/11. Como não está comprovado que
os filtros são iguais aos da carga Wails atual, não uso esse número como
comparação final. A diferença aritmética para o Majucau é R$ 44.275,91, mas ela
não identifica a causa.

## Auditoria do caminho no código

### Correção do escopo das contas — esclarecimento da responsável em 08/11/2026

Gisele confirmou que o relatório de referência foi consultado com **Itaú e
Nuvem Pago** selecionados. Essa é a seleção que deve ser reproduzida nas
entradas e saídas do Fluxo de Caixa. O card de saldo bancário continua sendo um
resultado separado, limitado ao Itaú.

Encontrei no código a causa comprovada de uma parte da divergência: tanto
`readLatestItauBalance` quanto `readBlingCashFlow` chamam
`selectedCashFlowAccountIDs`; essa função descarta todas as contas cujo nome não
começa com “Itaú”. Portanto, apesar da regra do relatório de usar Itaú + Nuvem
Pago, o gráfico e os cards de fluxo calculados pelo serviço atualmente ignoram
a conta Nuvem Pago. Isso pode alterar receitas e despesas, conforme os títulos
e movimentos existentes nela. É um erro de implementação, diferente da soma
das telas T03/T04.

A correção precisa separar dois conjuntos: contas do relatório (Itaú + Nuvem
Pago) para entradas/saídas do fluxo; Itaú somente para o card de saldo atual.
Também não se pode usar o saldo inicial Itaú como abertura da série agregada
das duas contas. A linha de saldo do gráfico só pode ser publicada quando a
abertura combinada das duas contas estiver comprovada. Até essa validação, não
se deve apresentar um saldo combinado inventado ou reaproveitar o saldo do
card Itaú como se fosse o do relatório.

`internal/dashboard/cash_flow.go` monta a série com três grupos de origem:

1. movimentos de `Caixas e Bancos`;
2. títulos em aberto de Contas a Receber;
3. títulos em aberto de Contas a Pagar.

Depois, filtra as contas selecionadas, aplica as regras de data e forma de
pagamento e soma os valores por dia; em seguida, agrupa os dias em semanas. A
consulta combina movimentos e títulos com `UNION ALL`, sem um identificador de
origem comum que permita eliminar com segurança um movimento e um título que
eventualmente representem a mesma obrigação. Isso é um risco estrutural de
duplicidade, **não prova** que a semana de R$ 101.721,05 contenha uma
duplicidade.

Também há um deslocamento de um dia para previsões, baseado em comparação
anterior, que permanece marcado como parcial enquanto não for validado com os
mesmos filtros do relatório. A tela atual usa snapshots normalizados por
recurso e seleciona somente lançamentos que passam pelos filtros financeiros
conhecidos; registros sem dados necessários podem ser excluídos e deixar a
série parcial.

## Conclusão e próxima verificação

- A afirmação de que a faixa 22/11–28/11 mostra R$ 101 mil não corresponde ao
  que aparece na janela Wails: essa faixa mostra R$ 27.012,27.
- A faixa de R$ 101.721,05 é 29/11–05/12. O dado salvo do Bling para 29–30/11 é
  parcial e não deve ser comparado como se cobrisse a semana inteira.
- Os registros salvos não provam a causa da diferença para 29/11–05/12. A
  hipótese de dupla contagem entre movimento e título continua sem prova de
  linha a linha.
- A divergência de escopo de contas está comprovada no código: o relatório
  escolhido pela responsável usa Itaú + Nuvem Pago, enquanto o serviço reduz
  o fluxo a Itaú. Isso explica por que a composição local pode divergir, mas
  ainda não quantifica quanto dos R$ 101.721,05 vem da conta Nuvem Pago.
- A próxima auditoria numérica precisa comparar exatamente 29/11–05/12 com as
  mesmas opções e contas, e decompor os R$ 101.721,05 por conta, tipo de origem,
  data, identificador e saldo do título. Corrigir o filtro do fluxo para usar
  Itaú + Nuvem Pago; manter o card de saldo limitado ao Itaú. Só depois validar
  a abertura combinada da linha de saldo, preservando as regras aprovadas.

**Situação:** auditoria de tela e código concluída; causa numérica da semana
29/11–05/12 ainda não comprovada. Nenhum número foi ajustado. T02 permanece em
validação; gráfico visual já aprovado. BK-062: 30% concluído (70% restante);
projeto: 12,3% concluído (87,7% restante).

**Referências internas:**

- `docs/evidence/T02-BLING-WEEKLY-REPORT-AND-ITAU-BALANCE-2026-11-07.md`
- `docs/evidence/BLING-CASH-FLOW-REPORT-READING-2026-11-02.md`
- `internal/dashboard/cash_flow.go`
