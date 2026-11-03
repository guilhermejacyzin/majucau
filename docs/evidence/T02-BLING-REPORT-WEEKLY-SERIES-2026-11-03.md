# T02 — série semanal do Fluxo de Caixa do Bling — 03/11/2026

## Decisão atendida

A T02 deve apresentar o resultado do relatório **Fluxo de Caixa** do Bling. A
fonte de produção continua sendo somente a API oficial direta. Como a
especificação pública não oferece uma rota documentada para ler o relatório
pronto, o Majucau compõe o resultado com os dados oficiais e mantém o estado
**Parcial** até a comparação com o relatório real ficar completa.

## Mudança desta entrega

Antes, o gráfico recebia valores diários. O relatório do Bling organiza o
período em faixas semanais. Agora o serviço cria, a partir da mesma série diária
da API, os totais semanais para a tela:

- a primeira faixa começa na data consultada e termina no sábado;
- as faixas seguintes vão de domingo a sábado;
- a última faixa termina no último dia disponível;
- entradas permanecem positivas e saídas permanecem negativas;
- as somas usam valores decimais exatos;
- nenhuma abertura, fechamento ou saldo projetado é calculado por estimativa.

O novo campo `cash_flow.periods` atravessa o worker, a camada Wails e o React.
T02 e T01 recebem a mesma composição, sem duplicar a consulta ao Bling. O
gráfico existente mostra as semanas e suas datas, sem mudar o desenho aprovado.

## Verificações

- `go test ./internal/dashboard ./ -count=1`: passou.
- `npm run typecheck`: passou na cópia isolada de verificação.
- `npm test`: 70 testes passaram na cópia isolada de verificação.
- `npm run build`: passou.
- Build Wails: passou usando a CLI Wails, com os arquivos de interface já
  compilados.
- O worker atualizado foi iniciado com o PostgreSQL 18 isolado do Majucau na
  porta `55438`; a nova janela do protótipo iniciou e exibiu o painel usando o
  snapshot local. O painel indicou a última sincronização real como
  `03/11/2026 05:25`.

## Limites que permanecem

O resultado segue **Parcial**. A janela de hoje ainda não foi atualizada no
horário diário das 09:30. A comparação registrada contra a captura do Bling
mantém uma diferença de R$ 0,01 em uma faixa de entrada; a primeira faixa
começa em 02/11, antes do início da carga disponível. O saldo inicial/float e
os indicadores que dependem dele continuam sem comprovação. Nenhum número foi
marcado como confirmado e nenhuma regra financeira foi alterada.

## Próximo passo

Executar a carga oficial das 09:30, comparar as mesmas contas, período e opções
do relatório do Bling, investigar o centavo e conferir se a primeira faixa
ficou completa. Depois, tratar a fonte aprovada para saldo por conta e só então
liberar saldo projetado, menor saldo e semana crítica.

## Referências oficiais

- [Referência pública da API Bling](https://developer.bling.com.br/referencia)
- [Relatório de Fluxo de Caixa do Bling](https://ajuda.bling.com.br/hc/pt-br/articles/360038306714-Como-gerar-um-relat%C3%B3rio-de-futuras-entradas-e-sa%C3%ADdas-de-valores)
- [Taxas e prazo de recebimento no Fluxo de Caixa](https://ajuda.bling.com.br/hc/pt-br/articles/360035872974-Como-controlar-as-taxas-que-pago-sobre-algumas-formas-de-pagamento-no-Bling)
