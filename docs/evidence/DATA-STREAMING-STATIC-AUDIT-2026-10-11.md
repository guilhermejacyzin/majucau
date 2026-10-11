# Auditoria estática de leitura em fluxo — 2026-10-11

## Objetivo e limite

Revisar a versão de código no commit `6b69298` em busca de leituras que tragam
arquivos, respostas ou coleções inteiras para a memória. Esta é uma conferência
estática por busca e leitura dos trechos identificados; não substitui medição de
memória, testes de carga nem validação de permissões em uma máquina Windows
limpa. Nenhuma tela ou regra financeira foi alterada.

## Resultado da busca

- Não foi encontrado `os.ReadFile` no código Go de produção.
- A revisão atual encontrou duas ocorrências de `io.ReadAll` no código Go de
  produção. Ambas usam `io.LimitReader` com limite explícito e leem somente um
  registro pequeno do pacote de backup ou o arquivo cifrado do DPAPI. A resposta
  OAuth do Bling não está entre elas: ela é decodificada diretamente do corpo
  HTTP com limite de bytes.
- A listagem de pastas que podem crescer usa `os.File.ReadDir(256)` e grava os
  nomes em uma tabela temporária do PostgreSQL, sem manter a pasta inteira numa
  lista Go.
- As linhas de `DATA-04` têm limites e streaming próprios descritos abaixo. A
  busca é um sinal favorável, não prova que todo fluxo do produto está concluído.

## Controles encontrados

| Fluxo | Limites e comportamento observados |
|---|---|
| CSV Bling/Nuvem Pago e prévia | Leitura contínua; campo até 1 MiB, registro até 4 MiB e 256 campos. Prévia usa índice temporário cifrado em disco; a fila de nomes de arquivo é lida em blocos de até 256 e ordenada em tabela temporária com memória do PostgreSQL limitada a 16 MiB. |
| API Bling | Corpo JSON até 4 MiB; página até 100 registros; decoder percorre a resposta e mantém um bloco limitado. A resposta OAuth também é decodificada diretamente do fluxo, sem criar uma cópia integral do corpo; os campos de token necessários continuam em memória. Retry fica restrito às falhas transitórias definidas para chamadas repetíveis. |
| Snapshot do painel | Consulta PostgreSQL em fluxo, linha por linha; máximo de 50 detalhes por tabela, campo até 4 KiB e registro até 16 KiB. IPC transmite blocos de até 512 KiB, abaixo do teto de mensagem de 1 MiB. |
| Backup | Dumps e criptografia V2 usam cópia/criptografia em fluxo. O leitor V1 antigo também descriptografa em fluxo. Somente manifesto (até 1 MiB) e chave do pacote (até 64 KiB) são lidos como pequenos registros limitados. |
| Segredos DPAPI | Valor até 1 MiB; arquivo cifrado é lido com limite adicional de 64 KiB e o buffer é zerado quando a leitura falha ou excede o teto. |
| Manifesto do instalador | Decoder em fluxo com teto de 4 MiB, limite para quantidade de arquivos e limite por caminho. Hashes de arquivos usam `io.Copy` para o hash, sem carregar o arquivo inteiro. |
| Logs técnicos | Arquivos giratórios com tamanho e quantidade máximos; cada registro tem limite próprio. |

## Pendências para comprovar o requisito inteiro

- Executar a matriz de carga com arquivos grandes, falhas e cancelamento, observando
  memória máxima do processo.
- Confirmar em VM Windows limpa as ACLs e a remoção dos temporários privados em
  encerramento normal, erro e cancelamento.
- Continuar a inspeção de fluxos menos óbvios e de componentes fora das buscas
  estáticas, incluindo exportações futuras.

Por isso, `DATA-04` continua `PARTIAL`; esta auditoria não promove o requisito a
verificado.

## Complemento — resposta OAuth do Bling — 2026-10-11

A conferência detalhada encontrou uma exceção pequena: o cliente OAuth lia até
4 MiB da resposta para um `[]byte`, mantendo o corpo inteiro junto com os
campos decodificados. O código agora usa `json.Decoder` diretamente no corpo
HTTP, atrás de um limite de 4 MiB.
Ele exige um único objeto JSON, rejeita conteúdo adicional e mantém a mesma
classificação de erro para resposta grande/ilegível e JSON inválido. Os tokens
continuam como campos tipados porque o serviço precisa usá-los; o corpo bruto
não fica duplicado em memória.

Validação local em 2026-10-11: `go test ./internal/integrations/bling` e
`go vet ./internal/integrations/bling` passaram. Isso não substitui medição de
memória nem homologação com a conta real do Bling.

## Revisão complementar dos leitores e da interface — 2026-10-11

No commit `6b69298`, a busca estática foi ampliada para código de produção Go,
`frontend/src` e scripts PowerShell. Na interface não foram encontrados usos de
`fetch`, Axios, `arrayBuffer` ou `blob` para carregar dados diretamente; os
fluxos passam pelos contratos Wails já descritos acima. Nos scripts, as leituras
de JSON de preflight/health usam `Read-BoundedTextFile`; a busca não encontrou
`ReadAllBytes`, `ReadAllText` ou `Get-Content -Raw`. As enumerações recursivas
identificadas servem para verificar links antes de limpeza segura ou localizar
um arquivo pequeno de preflight, e não para carregar o conteúdo dos arquivos.

Essa revisão é estática: não mede o pico de memória e não prova o comportamento
de cada dependência ou fluxo em execução. O diretório `artifacts/cache` estava
ausente no momento da conferência. `DATA-04` permanece `PARTIAL` até executar
carga com observação de memória e confirmar ACL/limpeza em VM Windows 10/11
limpa.

## Revalidação do estado atual — 11/10/2026

- Revalidei no commit `5d1769875a6177710a8eef93cfe6d2aeca49d6d2` os limites já descritos: respostas do Bling de até 4 MiB e 100 registros por página; CSV processado por linha com limites de campo/registro; listagem de pastas em blocos de 256 nomes; linhas do painel em fluxo, com limite de 50 por tabela, campos de 4 KiB, registros de 16 KiB e resposta Wails de até 16 MiB; quadros IPC de até 512 KiB. Os caminhos de leitura completa continuam limitados por `io.LimitReader`.
- A revisão confirmou que o painel decodifica os dados recebidos em fluxo e rejeita respostas acima de 16 MiB. Isso mantém um teto explícito de memória, embora o objeto final do painel seja materializado para a interface.
- A CI [38182492594](https://github.com/guilhermejacyzin/majucau/actions/runs/38182492594) concluiu com sucesso nos quatro jobs: `detect-changes`, `verify-backlog`, `verify-windows` e `verify-postgres`. O Windows executou testes, análise estática, build do aplicativo/worker/instalador, geração de SBOM e limpeza/consolidação de saídas; PostgreSQL executou as verificações de integração e backup/restore descartável.
- A CI não mede pico de memória sob carga real nem comprova ACL/limpeza em VM limpa. Também não homologa o cursor oficial do Bling. `DATA-04` permanece `PARTIAL`; a aprovação da CI não muda esse status.
- Nenhuma tela, regra financeira ou comportamento do produto foi alterado nesta revalidação. O relatório unsigned gerado pela CI está identificado no registro local de limpeza, pois a conexão disponível não oferece exclusão de artefatos remotos.

## Revalidação da leitura de movimentações de caixa — 11/10/2026

- No commit `1926856b8cf6f09b406f0d1901194c655c721988`, `ListCashMovements` lê uma página de `GET /caixas` por chamada, com limite local de 100 registros e filtros de página/período. O leitor usa o limite compartilhado de 4 MiB e o decoder JSON do adapter; `GetCashMovement` também usa a leitura de detalhe limitada. Cada objeto permanece como JSON bruto, sem classificar seu significado financeiro.
- A ausência de marcador de continuação permanece desconhecida: página curta não encerra a leitura. A rota ainda não está ligada à persistência; não há estimativa de CAPEX a partir desses dados.
- Os testes focais do adapter e `go vet` passaram localmente; a CI [38184670780](https://github.com/guilhermejacyzin/majucau/actions/runs/38184670780) passou nos jobs Windows e PostgreSQL. Isso valida os limites e contratos sintéticos, não mede pico de memória com dados reais nem homologa a paginação da conta.
- A inspeção da documentação da conta do Bling continua pendente por bloqueio de navegação no Chrome. Para terminar a homologação, é necessário confirmar os campos reais da resposta e a forma oficial de avançar de página. Nenhuma tela ou regra financeira foi alterada.
