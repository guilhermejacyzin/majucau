# Regras contábeis recebidas para a DRE do Majucau — v1

**Estado:** princípios de negócio recebidos de Gisele em 2026-10-04. Este documento organiza as regras compartilhadas para orientar o desenho da DRE própria do Majucau. Não é um parecer contábil completo nem substitui a validação do contador para o regime, as operações e os dados reais da empresa.

**Proteção da interface:** a tela e as linhas aprovadas da DRE permanecem iguais. Este documento detalha o tratamento dos dados por trás dessas linhas; não autoriza criar, remover ou renomear linhas, cards ou telas. EBIT e EBITDA continuam nos cards já aprovados, fora da tabela contábil.

## 1. Estrutura e classificação já informadas

| Linha aprovada da DRE | Entra | Não entra automaticamente |
|---|---|---|
| Receita Bruta de Vendas | Vendas de chocolates, produtos, mercadorias e demais receitas da atividade, conforme o reconhecimento aplicável | PIX sem venda identificada, empréstimos, aporte de sócios e transferências entre contas |
| Deduções da Receita | Devoluções, abatimentos e tributos incidentes conforme o tratamento contábil/fiscal aplicável | Despesas administrativas, frete de compra e salários |
| Receita Líquida | Receita atribuível à operação após as deduções | — |
| CPV/CMV | Custo do produto efetivamente vendido | Toda compra de matéria-prima do mês |
| Lucro Bruto | Receita líquida menos CPV/CMV | — |
| Despesas com Vendas | Comissões, marketing e despesas comerciais/marketplace conforme sua natureza | Ingredientes e embalagens incorporadas ao produto |
| Despesas Administrativas | Contabilidade, administrativo, escritório e sistemas administrativos | Gastos diretamente ligados à fabricação, sem antes avaliar o tratamento de custo aplicável |
| Outras Despesas Operacionais | Despesas operacionais que não se enquadrem adequadamente nas categorias anteriores | Gastos financeiros |
| Resultado Operacional | Resultado das linhas operacionais anteriores | — |
| Resultado Financeiro | Juros, tarifas bancárias financeiras e rendimentos financeiros conforme sua natureza | Compra de máquinas ou pagamento de fornecedores |
| Outras Receitas/Despesas | Eventos classificados conforme sua natureza contábil | Venda normal de chocolates |
| Resultado antes dos tributos sobre o lucro | Resultado anterior aos tributos sobre o lucro | — |
| Tributos sobre o Lucro | Conforme regime e competência aplicáveis | Tributos sobre vendas misturados nesta linha |
| Lucro/Prejuízo Líquido | Resultado econômico final | Saldo bancário |

As linhas são as do modelo funcional já aprovado em `docs/FUNCTIONAL-DECOMPOSITION-v1.md`. Se um evento não puder ser classificado com segurança, deve ficar pendente/indisponível; não se escolhe uma linha por conveniência.

## 2. Estoque, produção e reconhecimento do CPV/CMV

Exemplo simples: se a empresa compra R$ 100.000 de cacau/chocolate e R$ 20.000 de embalagens, mas só transforma e vende parte desse material, não se registra automaticamente R$ 120.000 na DRE. A parte ainda em matéria-prima, produto em processo ou produto acabado fica no estoque. Só a parcela de custo ligada aos produtos vendidos chega ao CPV/CMV.

Fluxo contábil descrito para orientar o sistema:

1. Na compra de matéria-prima: débito em estoque de matéria-prima e crédito em fornecedores/bancos, conforme o fato e o documento. A compra, por si só, não afeta a DRE.
2. Na produção: custos apropriáveis de matéria-prima, mão de obra direta e custos indiretos de fabricação são atribuídos ao produto segundo critérios técnicos aprovados e passam pelo estoque de produtos em processo/acabados.
3. Na venda: o custo correspondente ao produto vendido sai do estoque e é reconhecido no CPV/CMV da DRE.

O CPC 16 prevê custos de aquisição e transformação no estoque e o reconhecimento como despesa quando o estoque é vendido; inclui mão de obra direta e alocação sistemática dos custos indiretos fixos/variáveis. A alocação de custos fixos considera capacidade normal; custo fixo não alocado por baixa produção é despesa do período. Desperdício anormal e custos que não contribuam para trazer o estoque à condição/localização atuais também não devem ser usados para inflar o estoque. Veja [CPC 16 (R1) — Estoques](https://www.cpc.org.br/Arquivos/Documentos/243_CPC_16_R1_rev%2012.pdf) e o [sumário oficial do CPC 16](https://www.cpc.org.br/Arquivos/Documentos/244_CPC_16_Sumario.pdf).

## 3. Máquinas e depreciação

Compra de uma máquina não vira despesa automaticamente no mês da compra. Primeiro é preciso avaliar se ela atende aos critérios de imobilizado e obter o cadastro e os dados necessários de uso e depreciação. A depreciação de máquina usada na produção pode compor o custo dos estoques; a classificação depende de onde e como o ativo é usado e da regra aprovada. Equipamento administrativo deve ser tratado conforme sua função administrativa. Referência: [CPC 27 — Ativo Imobilizado](https://www.cpc.org.br/Arquivos/Documentos/316_CPC_27_rev%2014.pdf), que relaciona depreciação de máquinas produtivas ao custo de produção de estoques conforme CPC 16.

## 4. Tributos recuperáveis

Tributo efetivamente recuperável não compõe o custo de aquisição do estoque; pode representar um crédito/ativo até sua utilização ou baixa. O Majucau não presume que uma empresa tem direito ao crédito. A conclusão depende do regime tributário, tributo, operação, localidade/UF e período. Referências: [CPC 16 (R1), item 11](https://www.cpc.org.br/Arquivos/Documentos/243_CPC_16_R1_rev%2012.pdf) e [pergunta frequente do CFC sobre tributos recuperáveis](https://cfc.org.br/tecnica/perguntas-frequentes/4066-2/).

## 5. Proibições de classificação automática

1. Pagamento não determina despesa; considerar competência e fato econômico.
2. Recebimento não determina receita; identificar o fato gerador/reconhecimento aplicável.
3. Compra de matéria-prima não é automaticamente CPV/CMV; passa pelo estoque e pela produção.
4. Compra de máquina não é automaticamente despesa; avaliar reconhecimento como imobilizado.
5. Saldo bancário não determina lucro; DRE e fluxo de caixa são visões diferentes.
6. Nenhum lançamento relevante entra na DRE sem identificação do documento e do fato econômico que o originaram.

## 6. O que ainda precisa ser definido com a contabilidade e com dados reais

Estas lacunas não serão preenchidas por suposição:

- método de valoração do estoque aplicável à empresa e consistência por natureza/uso;
- critério de capacidade normal e rateio de custos indiretos de fabricação;
- separação de desperdício normal/anormal, ociosidade, armazenamento e demais gastos fabris;
- regime tributário, tributos recuperáveis e tratamento por operação, UF/localidade e período;
- critérios de imobilizado, vida útil, valor residual, início e método de depreciação;
- mapeamento do plano de contas e dos lançamentos reais às linhas já aprovadas, inclusive perdas e outras receitas/despesas;
- competência, documentos e ajustes contábeis que não estejam completos nas fontes atuais.

Para calcular CPV/CMV, extrato bancário ou contas a pagar/receber não bastam. São necessários, no mínimo, saldos e movimentos de matérias-primas, produtos em processo e acabados; consumo e produção por produto; vendas e devoluções; custos de mão de obra e fabricação; documentos de compra e tributos; cadastro/depreciação do imobilizado; plano de contas e balancete. Campo ausente significa “Dado indisponível” no cálculo dependente, nunca zero nem estimativa.

## 7. Próxima etapa

Com a contabilidade, montar uma tabela de conferência com: conta/lançamento, documento, fato econômico, linha da DRE, custo ou despesa, competência, efeito no estoque, tratamento de tributo, fonte e divergência. Depois validar os critérios pendentes e os campos disponíveis nas integrações. Só então fechar casos de referência e implementar o cálculo no Majucau.
