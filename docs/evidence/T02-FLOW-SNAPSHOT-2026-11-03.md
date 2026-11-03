# T02 — valores da API do Bling ligados à série — 03/11/2026

## Atualização às 05:38 BRT

A tela agora recebe números do snapshot diário mais recente do Bling. A leitura
do serviço local, após atualizar o worker, retornou para os próximos 30 dias:

- Entradas: **R$ 69.994,12 — Parcial**.
- Saídas: **R$ 263.930,47 — Parcial** (valor interno negativo, conforme a regra
  aprovada para saídas).
- Série disponível: **61 datas** (hoje até D+60).

O worker também transmite os mesmos campos à interface T02 e ao painel T01;
`App.tsx` atualiza o snapshot a cada 60 segundos e ao trocar de tela. A captura
visual automatizada desta sessão expirou; a confirmação é da resposta real do
serviço local, que já estava conectado ao banco isolado e às 50 linhas de
recebíveis e 50 de pagamentos disponibilizadas no snapshot.

## Correções que fizeram os valores aparecerem

- A tela buscava apenas o lote associado ao horário diário anterior. Agora usa
  o conjunto mais recente que publicou, na mesma execução, Caixas, Contas a
  Pagar e Contas a Receber. Se uma nova execução ainda estiver incompleta,
  preserva o conjunto anterior completo.
- `GET /contas-contabeis` continua respondendo 403 nesta conta. Para manter as
  duas contas já aprovadas, Itaú e Mercado Pago, a T02 identifica seus IDs pelo
  nome que acompanha cada lançamento do `GET /caixas`. A alternativa só é usada
  quando não há catálogo de contas normalizado, e exige um único ID para cada
  nome; qualquer ambiguidade deixa a projeção indisponível.
- A API retornou 31.032 movimentos com o campo `situacao` vazio. A orientação
  registrada é mostrar os lançamentos que o Bling listou, com aviso Parcial.
  Assim, `R` e situação vazia entram nos movimentos; situações não vazias e
  desconhecidas não entram e mantêm o resultado parcial.
- O `saldo` de Contas a Pagar já chega como valor restante e é a origem do
  compromisso futuro. A normalização guarda a situação do Bling como número;
  a consulta antiga procurava os nomes textuais `OPEN` e `PARTIALLY_PAID`,
  excluindo todos os títulos e zerando as saídas. Agora usa saldo restante não
  zero e preserva bloqueios textuais explícitos de cancelamento, sempre com
  estado Parcial enquanto os códigos numéricos não estiverem descritos.
- As taxas de pagamento vêm com oito casas numéricas, muitas vezes preenchidas
  por zeros sem significado. A leitura agora remove somente esses zeros finais
  antes da conversão monetária; um valor com precisão significativa acima do
  suportado deixa só aquele recebimento fora e marca a série Parcial, sem
  derrubar todo o painel.

Não houve mudança no desenho aprovado nem no sentido contábil: entradas são
positivas, saídas são negativas. O resultado continua Parcial, não homologado e
sem equivalência comprovada com o relatório pronto, porque a referência da API
não publica uma rota que devolva esse relatório diretamente. A aplicação
reproduz seus componentes usando as APIs oficiais e as contas selecionadas.

## Verificação desta atualização

- `go build -trimpath -o artifacts/cache/dev/majucau-worker-refreshed.exe ./cmd/worker` — passou.
- Leitura real do serviço IPC após reiniciar somente o worker — devolveu os
  valores acima, `PARTIAL`, 61 datas e as tabelas do snapshot.
- `git diff --check` — passou.
- Não executei suíte de testes nesta atualização.

## O que falta para concluir BK-062

1. Comparar a mesma janela, contas e opções do relatório Bling com os valores
   obtidos pela API.
2. Confirmar equivalência de taxa/prazo, float e filtros do relatório.
3. Obter o saldo atual oficial por conta. Não calcular saldo final projetado,
   menor saldo ou semana crítica sem a fonte necessária.
4. Só depois considerar a T02 pronta e promover sua contribuição à T01.

## Decisão atendida

A responsável confirmou que a T02 deve usar o resultado do relatório **Fluxo de
Caixa** do Bling. A integração deve consultar apenas as APIs oficiais do Bling;
relatórios CSV/XLS não são fonte de produção. O desenho aprovado da T02 e as
regras financeiras permanecem protegidos.

## O que foi ligado

- O backend lê, por cursor de linhas, candidatos das execuções normalizadas do
  mesmo lote diário: `/caixas`, Contas a Receber, Contas a Pagar, identidade de
  contas e termos de formas de pagamento. A série é limitada a hoje até D+60;
  os cards usam os primeiros 30 dias.
- Só Itaú e Mercado Pago, aprovados como contas do relatório de referência,
  entram na composição. Títulos e movimentos sem conta correspondente deixam
  o resultado parcial e não são somados às cegas.
- Movimentos de caixa com situação `R` ou vazia entram; os vazios ficam
  identificados pelo estado Parcial. Situações desconhecidas não são somadas.
  Contas a Pagar usam saldo restante não zero; recebimentos exigem destino,
  taxa, prazo e configuração de dias conhecidos. Saídas mantêm sinal negativo.
- `DashboardSnapshot.cash_flow` carrega os totais de 30 dias e a série diária.
  A T02 liga esses dados aos cards e ao gráfico aprovado; a T01 lê o mesmo
  snapshot compartilhado.

## Limite que continua visível

Esta etapa é uma **composição parcial pela API**, não prova que o resultado já
seja igual ao relatório Bling. A referência pública consultada não documenta
um endpoint pronto para o relatório. Ainda faltam provar a equivalência dos
filtros e do float, tratar prazos em dias úteis com calendário aprovado,
obter o saldo inicial oficial por conta e comparar valores/datas com uma
execução atual do relatório.

Por isso os valores da série são marcados `PARTIAL`. Não se calcula saldo
projetado, menor saldo, saldo transferível ou semana crítica sem o saldo
inicial e sem a regra/entrada oficial necessária. O saldo atual continua
indisponível até vir da fonte aprovada de Caixa e Bancos; não é inferido dos
movimentos.

## Verificações desta etapa

- `go test ./internal/...` — passou localmente.
- `go vet ./internal/dashboard ./internal/financial` — passou.
- `git diff --check` — passou.
- Build TypeScript/Vite e ESLint — passaram numa cópia isolada em
  `artifacts/verification`.
- Testes frontend — 70 testes passaram.
- A cópia de verificação frontend foi mantida em `artifacts/verification/`
  porque a revisão automática bloqueou a exclusão solicitada após a validação.
  O local e o motivo estão listados em
  `docs/evidence/OUTPUT-RETIDO-MANUAL-2026-11-03.md`.
- A captura visual automatizada desta sessão expirou; a resposta do worker foi
  confirmada pelo cliente IPC local, sem imprimir dados de clientes.
- CI 41587012467 — passou em todos os trabalhos: governança, testes Go,
  integração PostgreSQL, frontend, build desktop e smoke Windows do serviço.
- A CI anterior 41586134294 havia encontrado o identificador do serviço de teste
  ainda aberto durante a limpeza; o smoke foi corrigido e a nova CI confirmou a
  remoção segura.
- O executável Wails aberto permaneceu intacto. Apenas o worker de dados local
  foi reiniciado com a compilação atualizada.

## Próxima sequência

1. Conferir os valores com o relatório Bling usando a mesma janela, contas
   Itaú/Mercado Pago, float, prazo e opção de incluir pagamentos/recebimentos.
2. Reexecutar a sincronização diária depois da conferência e verificar se a
   composição por API mantém os mesmos valores.
3. Só depois completar os indicadores que dependem do saldo inicial e liberar
   a T02/T01 como homologadas.
