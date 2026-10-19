# Base aritmética da DRE própria — 19/10/2026

## O que foi feito

Foi criado em `internal/financial/dre.go` um cálculo puro para as linhas já aprovadas da DRE do Majucau: receita bruta, deduções, receita líquida, CPV/CMV, lucro bruto, despesas operacionais, resultado operacional, resultado financeiro, outras receitas/despesas, resultado antes dos tributos sobre o lucro, tributos sobre o lucro e lucro/prejuízo líquido.

O cálculo recebe valores com sinal já definido por uma etapa anterior de classificação, exige período e versão da regra, usa `domain.Money` sem `float`, propaga erros de overflow sem devolver resultado parcial e distingue valor zero confirmado de campo ausente. Valores ausentes deixam indisponíveis as linhas que dependem deles. Mesmo com todas as entradas presentes, o resultado permanece provisório até que fonte, classificação e conciliação sejam aprovadas.

## O que não foi feito

Não foram criados mapeamentos de contas ou fornecedores, nem integração com dados RAW do Bling ou arquivos do Drive. A compra de matéria-prima, produção, tributos recuperáveis, depreciação capitalizada e demais critérios contábeis continuam sujeitos à validação específica já registrada em `docs/DRE-ACCOUNTING-POLICY-v1.md`. Não calculei os cards EBIT/EBITDA nem CAPEX nesta mudança; eles dependem de valores e classificações conciliados. Nenhuma tela foi modificada.

## Verificação

`go test ./internal/financial` e `scripts/verify.ps1` passaram. A verificação completa cobriu todos os testes Go, `go vet`, vulnerabilidades, migrations, geração SQLC determinística, busca de segredos, instalador, lint, tipos, 33 testes de interface, build e `npm audit` (sem vulnerabilidades no relatório npm). O `govulncheck` não identificou chamadas alcançáveis a vulnerabilidades; registrou uma vulnerabilidade em um módulo requerido sem chamada alcançável. Os casos da DRE verificam a aritmética com valores de referência fornecidos pela usuária, a propagação de indisponibilidade quando falta CMV, a validação de período/versão e o tratamento de overflow.

## Backlog e próximo passo

BK-066 passou de 0% para 5% porque a fundação aritmética está implementada. A tarefa continua em andamento; faltam classificação/mapeamento contábil aprovados, consumo dos lançamentos reais com origem e competência, snapshots rastreáveis e comparação com o relatório do Bling. O backlog global passou para 11,8% concluído e 88,2% restante (127 tarefas); isso não representa aceite, homologação nem prontidão de produção.

O próximo passo dentro desta tarefa é definir o contrato de entrada dos registros financeiros já classificados e ligar a consulta dos lançamentos do Bling sem copiar a DRE pronta. Se o contrato de classificação exigir uma escolha contábil não aprovada, essa parte deve aguardar validação, preservando os totais como indisponíveis.
