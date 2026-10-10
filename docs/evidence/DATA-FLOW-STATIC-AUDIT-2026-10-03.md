# Auditoria estática de fluxos de dados — 2026-10-03

## Escopo

Leitura estática dos pontos de entrada e saída de arquivos, CSV, respostas HTTP,
segredos, manifestos, journals, backup e mensagens IPC no código Go e nos
bindings da interface. A auditoria não executa profiling, não mede pico de
memória e não substitui a validação em VM Windows.

## Fluxos com limite ou processamento em fluxo identificado

| Fluxo | Implementação observada | Resultado da revisão |
| --- | --- | --- |
| Resposta de dados do Bling | `json.Decoder` sobre `io.LimitedReader`, limite padrão de 4 MiB e máximo de 100 registros por página | O envelope é lido progressivamente; respostas acima do limite são rejeitadas. A página permanece limitada em memória pelos limites de bytes e registros. |
| Resposta OAuth do Bling | `readLimitedBody` consome no máximo o limite configurado mais um byte; o limite padrão é 4 MiB | Materialização pequena e limitada; excesso é rejeitado antes de decodificar o token. |
| CSV Bling e Nuvem Pago | Leitores CSV encapsulados por `csvlimit.NewReader` e linhas processadas incrementalmente | Mantêm os limites aprovados de 1 MiB por campo, 4 MiB por registro e 256 campos, sem teto total de linhas. |
| Backups V1/V2 | Dumps e payloads V2 cifrados/descriptografados em segmentos; leitura V1 preservada; manifesto e entradas auxiliares têm limites explícitos | Payloads não são carregados inteiros em memória; uma leitura auxiliar de entrada ZIP é limitada pelo tamanho máximo do manifesto/keyset correspondente. A criação e verificação V2 com dumps reais passou na CI `37142318615`. |
| Segredos DPAPI | Arquivo lido com `io.LimitedReader` e limite de tamanho do blob; bytes são limpos em erro | Leitura integral apenas do blob limitado, não de arquivo arbitrário. |
| Journal de instalação e manifesto de release | `json.Decoder` lê via `io.LimitedReader`, valida tamanho e rejeita conteúdo excedente/trailing data | Conteúdo materializado com limite declarado. |
| Entrada IPC | Quadro tem limite de 1 MiB antes da alocação/leitura do payload; request JSON também valida o limite | Entrada remota ao worker fica limitada por mensagem. |

Na busca dos fontes Go, os demais usos produtivos de `io.ReadAll` encontrados
são os leitores acima e permanecem limitados. As ocorrências de `os.ReadFile`
encontradas eram testes; não apareceu leitura produtiva arbitrária de arquivo
inteiro nessa busca. `security.RedactJSON` tinha decodificação JSON sem limite;
agora rejeita entradas acima de 1 MiB com saída totalmente redigida. A busca
encontrou apenas um chamador de teste, portanto nenhum fluxo de produção mudou.

## Varredura complementar — filas de arquivos e respostas de prévia

- A enumeração das pastas CSV lê até 256 entradas por vez. Na importação, os
  nomes entram em uma tabela temporária PostgreSQL e são consultados de volta
  em blocos de até 256; `work_mem` e `maintenance_work_mem` da transação ficam
  em 16 MiB. A tabela é descartada no commit/rollback. A listagem completa não
  é mantida em uma fatia Go.
- Na prévia sem banco, a ordenação externa usa blocos de 256 nomes e intercala
  no máximo 64 arquivos temporários por vez. O índice mantém HMACs dos IDs; os
  nomes/locais temporários da prévia são cifrados. A prévia percorre todos os
  arquivos, mas só devolve até 50 metadados e 50 detalhes de erro, junto dos
  totais completos.
- As duas rotas de prévia montam respostas apenas a partir desses arrays
  limitados. O painel agora consulta recebíveis e pagáveis registro a registro e
  envia o JSON em quadros IPC de até 512 KiB. A mesma resposta final alimenta a
  tela já aprovada. Os limites de 4 KiB por campo e 16 KiB por registro impedem
  que texto sem tamanho máximo na migration gere uma resposta ilimitada.
- A listagem de integrações usa `rows.Next`, mas a tabela de conexões limita
  `provider` a quatro valores únicos; portanto, essa consulta tem teto de
  quatro linhas. Os demais fluxos de cópia de backup e hash de pacote usam
  `io.Copy`, sem materializar o arquivo inteiro.

Esta varredura é estática e não prova limites de disco do PostgreSQL, ACLs ou
remoção de temporários em uma instalação real. A importação guarda nomes de
arquivo nas tabelas temporárias PostgreSQL; a revisão não verificou se a ACL
do diretório de dados protege os arquivos temporários do servidor durante a
execução. Esses pontos continuam dependentes de validação em VM Windows; não
houve alteração de comportamento, tela ou regra de negócio nesta etapa.

## Resposta do snapshot do painel

Em 2026-10-04, a usuária autorizou enviar os dados internos em blocos menores,
mantendo a mesma tela e as mesmas informações. A implementação separa o resumo
financeiro das consultas de detalhe, lê cada linha do PostgreSQL e transmite o
JSON em quadros sequenciais. Uma transação somente leitura `REPEATABLE READ`
mantém resumo e tabelas na mesma visão do banco. Campos acima de 4 KiB ou linhas
acima de 16 KiB falham sem encaminhar o texto excessivo. O cliente verifica a
sequência e só devolve o snapshot quando recebe o quadro final e decodifica o
JSON completo; respostas incompletas não são exibidas.

O protocolo IPC foi elevado para v2 porque o contrato de resposta passou a
suportar quadros de streaming. A validação de execução Windows/PostgreSQL ainda
precisa ocorrer na CI; esta alteração não muda a tela nem as regras financeiras.

### Revalidação complementar — 2026-10-04

A análise inicial não identificou que o método antigo `ReadDashboardSnapshot`
continuava no código apesar de não ter chamadores de produção. Esse método foi
removido junto do contrato de leitura não-streaming; o contrato do worker agora
exige `WriteDashboardSnapshot`. Testes locais com Go 1.26.6 passaram para
`internal/dashboard` e `cmd/worker`, incluindo limites e rejeição de linhas
excedentes. A compilação/CI Windows/PostgreSQL continua pendente.

## Resultado

A maior parte dos caminhos revisados já tem streaming ou limite de tamanho
explícito. O contrato de resposta do painel foi implementado; `DATA-04` continua
`PARTIAL` até a CI Windows/PostgreSQL validar essa mudança, a auditoria dos
demais fluxos terminar e a limpeza/ACL ser comprovada em VM Windows real.

### Revarredura estática de Go, banco e frontend — 2026-10-10

- Busca nos fontes Go de produção (excluindo arquivos `_test.go`) não encontrou
  `os.ReadFile` nem leitura arbitrária de arquivo inteiro. Os usos produtivos
  de `io.ReadAll` identificados estão limitados: resposta HTTP do Bling com
  limite configurável mais um byte, entrada auxiliar do pacote de backup com
  limite do manifesto e arquivo DPAPI com limite do blob.
- O decoder da resposta de dados Bling usa `json.Decoder` sobre
  `io.LimitedReader`; o decoder OAuth usa `readLimitedBody`. Journals e
  manifestos do instalador também usam leitores limitados. Mensagens IPC
  comuns têm teto de 1 MiB e o fluxo do painel usa frames em blocos de 512 KiB.
- `internal/dashboard/stream.go` consulta detalhes linha a linha, limita cada
  tabela a 50 linhas e aplica limites de 4 KiB por campo e 16 KiB por registro.
- `internal/integrations/csvstream/file_queue.go` enumera 256 entradas por
  leitura, armazena a lista em tabela temporária PostgreSQL e lê lotes de até
  256 nomes; não mantém a pasta inteira em uma fatia Go.
- `ListIntegrationStatus` monta uma fatia, mas a migration restringe `provider`
  a quatro valores únicos (`BLING`, `NUVEMSHOP`, `NUVEM_PAGO`, `PAYROLL`), então
  essa consulta tem um teto de quatro linhas.
- A busca no frontend não encontrou `FileReader`, `readAsText`,
  `readAsArrayBuffer`, `File.text()`, `Blob.arrayBuffer()`, `fetch` ou
  `response.json()`. Os fluxos de CSV chamam os métodos Wails do worker com o
  caminho selecionado, sem carregar o arquivo no navegador.
- O resultado é uma revisão estática dos padrões pesquisados; não prova pico de
  memória em execução, ACL do PostgreSQL, limpeza real de temporários ou
  comportamento em máquina Windows limpa. `DATA-04` permanece `PARTIAL` até
  esses gates serem comprovados.
- Nenhum código, tela ou regra de negócio foi alterado nesta revarredura. Testes
  não foram executados.

## Regressão do limite de redação — CI #241

A CI [37144902549](https://github.com/guilhermejacyzin/majucau/actions/runs/37144902549)
passou em Windows e PostgreSQL após `security.RedactJSON` passar a rejeitar
entradas acima de 1 MiB com substituição integral por `[REDACTED]`. O helper
segue sem chamadores de produção; esta prova não substitui limites nos fluxos
produtivos. `DATA-04` permanece `PARTIAL` pelas pendências listadas acima.

### Limites para leituras de texto nos scripts Windows — 2026-10-10

A revarredura encontrou `Get-Content -Raw` nos scripts de validação do instalador.
Esses arquivos eram resultados pequenos gerados pelo próprio helper e o código
NSIS do repositório, mas ainda não tinham um limite explícito de leitura. Para
evitar que uma saída defeituosa ou inesperadamente grande seja carregada sem
limite:

- `scripts/BoundedText.psm1` lê em blocos de 8 KiB, rejeita o byte que ultrapassa
  o teto, valida UTF-8 (com detecção de BOM) e fecha os recursos em `finally`.
- `smoke-windows.ps1` limita JSON/health a 256 KiB e o diagnóstico a 1 MiB.
- `smoke-installer.ps1` limita JSON a 256 KiB, erro a 64 KiB e código de saída a
  4 KiB; o diff local anterior no fim desse script foi mantido fora do commit.
- `test-installer-nsis.ps1` limita a leitura do fonte NSIS a 4 MiB.

Validação local em 2026-10-10: parser PowerShell dos quatro arquivos passou;
leitura abaixo do teto e rejeição acima do teto passaram; o contrato do NSIS
retornou `PASS`; a busca não encontrou mais `Get-Content -Raw` em scripts
PowerShell. A CI `38077877166` passou em Windows e PostgreSQL; no Windows,
também passou pelo smoke completo de instalação/desinstalação, pela geração e
consolidação dos outputs e pela verificação do artefato final. Esta mudança
afeta apenas ferramentas de validação, sem alterar telas, regras financeiras
ou o comportamento de produção. `DATA-04` continua `PARTIAL` até os gates de
VM/ACL/limpeza serem comprovados.
