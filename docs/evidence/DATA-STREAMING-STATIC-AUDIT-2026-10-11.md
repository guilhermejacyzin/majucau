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
