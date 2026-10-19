# Conferência somente de leitura dos arquivos do Drive — 19/10/2026

## Escopo

- A sessão Chrome `majucau.com.br` estava autenticada como `financeiro@majucau.com.br`.
- A pasta confirmada foi `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT`, com `01_perdas_operacionais`, `02_nuvem_pago/Recebimentos_futuros` e `03_nuvem_envio/Frete`.
- A conferência foi somente de leitura. Nenhum arquivo foi baixado, editado, movido, importado ou compartilhado.
- As prévias do Drive também exibiram linhas de dados. Elas continham informações de clientes; nenhum valor, nome, telefone, identificador de pedido, rastreio ou anotação dessas linhas foi copiado para o projeto, para o Git ou para um relatório.

## Estruturas observadas

| Pasta | Formato e aba | Grupos de colunas visíveis | Situação para o Majucau |
|---|---|---|---|
| `01_perdas_operacionais` | Excel `.xlsx`, aba `Problemas frete` | conclusão, datas do registro e da entrega, pedido, motivo, transportadora, atraso, manifestação e andamento. Inclui nome/contato, rastreio e texto livre. | É um controle operacional; não há uma coluna de valor monetário por ocorrência. Não usar o texto livre ou o custo do pedido como perda financeira presumida. O arquivo ainda não passa pelo importador XLSX, hoje ligado somente a Recebimentos Futuros. |
| `02_nuvem_pago/Recebimentos_futuros` | CSV | datas de pagamento e recebimento, movimento/transação, método e parcelas, valores bruto/líquido e taxas/juros/custos. Também há campo de nome. | O caminho atual reconhece CSV, mas os dados permanecem sem uso como fatos confirmados enquanto o contrato financeiro não for homologado. |
| `03_nuvem_envio/Frete` | CSV | envio e rastreio, modalidade/transportadora, peso/dimensões, frete cotado e faturado, taxas, total, valor declarado e opções adicionais. | Ainda não existe contrato/importador financeiro aprovado para esta fonte. Não inferir despesa ou perda a partir de cotação, valor declarado ou anotações operacionais. |

Os itens listados nas três estruturas exibiam modificação em **11/09/2026**, mais de um mês antes desta conferência. Portanto, nenhum deles pode ser apresentado como atualização diária atual; a regra existente exige manter a fonte como pendente até chegar arquivo compatível e recente.

## Consequências técnicas e de segurança

- A validação estrutural XLSX disponível hoje está conectada ao caminho de `02_nuvem_pago/Recebimentos_futuros`; o `.xlsx` real encontrado em `01_perdas_operacionais` não é processado por esse caminho.
- O CSV de Frete também não é interpretado pelo importador de recebimentos futuros. O inventário do Drive é mais amplo que os importadores financeiros hoje ligados.
- Antes de qualquer importação financeira, faltam validadores por fonte/versão, decisão documentada das colunas permitidas e regras de retenção. Nome, telefone, rastreio e texto livre não devem entrar em logs, relatórios ou fatos financeiros sem necessidade e autorização específicas.
- Para `01_perdas_operacionais`, a presença de motivo, reembolso ou manifestação em texto não comprova um valor de perda. O valor precisa vir de uma fonte financeira oficial e de uma regra aprovada, mantendo rastreabilidade.
- A descoberta não altera regras financeiras, telas, permissões nem a agenda diária. Os arquivos continuam pendentes e nenhum valor foi calculado ou publicado.

## Próximo passo seguro

Ampliar a validação estrutural em fluxo para as fontes autorizadas fora de Recebimentos Futuros, sem interpretar linhas nem criar fatos. Depois, aprovar contratos de cada fonte, com minimização de dados pessoais, antes de ligar qualquer coluna aos relatórios ou à DRE.
