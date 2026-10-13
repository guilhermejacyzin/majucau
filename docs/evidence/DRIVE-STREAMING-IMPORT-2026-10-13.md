# Ingestão do Google Drive em fluxo — 13/10/2026

> Atualização de 13/10/2026: a versão 16 do Site também corrigiu a leitura do
> formulário protegido de configuração do segredo Bling. O corpo agora é lido
> em fluxo, limitado a 8 KiB e cancelado se exceder o teto. A revisão e os
> resultados estão em `DATA-STREAMING-STATIC-AUDIT-2026-10-11.md`.

## O que foi entregue

A ponte do Majucau agora lê somente os arquivos das três estruturas já aprovadas do Drive. A listagem usa páginas de até 100 arquivos; o conteúdo baixa como fluxo, linha por linha, para CSV/TSV e registro por registro para JSON. O processamento calcula SHA-256 durante a leitura, sem carregar o arquivo inteiro na memória.

Antes e depois do download, o sistema confere os metadados da mesma versão do arquivo: ID, nome, pasta, tamanho, hash informado pelo Drive, data de alteração e versão. Se o arquivo for modificado, movido ou removido durante a leitura, as linhas já preparadas daquele arquivo são apagadas e o lote fica indisponível.

Cada campo pode ter até 1 MiB; cada linha, 4 MiB; cada registro, 256 campos. JSON aceita até 64 níveis e listas de até 256 itens. O banco D1 limita uma linha a 2.000.000 bytes; a aplicação rejeita o arquivo se um registro convertido para JSON passar de 1.500.000 bytes ou se a chave candidata passar de 32 KiB.

As gravações temporárias usam blocos de 40 registros. O lote fica `IMPORTING` enquanto lê. Só depois de terminar a leitura, o hash, as validações estruturais e a auditoria é marcado como pendente de decisão. Uma falha remove o staging parcial; lotes interrompidos por encerramento inesperado são recuperados na próxima importação. Arquivos repetidos são detectados por hash e as contagens temporárias de duplicatas são apagadas ao fechar o lote. Nenhuma linha é promovida a fato financeiro.

Chamadas externas repetíveis são somente GET e usam no máximo três tentativas, com espera aleatória crescente e `Retry-After` limitado a 30 segundos. O serviço de credencial é POST e não é repetido automaticamente. Cancelamentos não são repetidos.

## Publicação e verificação

- Código do Site publicado: versão 15, commit de origem `8edf1b4a57772ea0c1ba3cd1e452ba95806210d3`.
- A versão 16, commit `fdce727d8f3d868b5bea418d52208039054e7cf8`, preservou a ponte OAuth e passou a limitar em fluxo o formulário de configuração do segredo a 8 KiB.
- Endereço: <https://majucau-bling-bridge.giselegalha.chatgpt.site>.
- O deploy terminou com sucesso e manteve a configuração pública já existente do Site.
- O banco publicado contém `drive_import_row_hash_counts`, além das tabelas de lote, staging e auditoria. A conferência confirmou que as migrations 0004 e 0005 chegaram ao D1.
- As 27 verificações automatizadas do Site passaram; a compilação e a validação final do pacote também passaram.
- O empacotador oficial foi executado com o Bash do Git para Windows. Não foi necessário instalar o WSL.

## O que falta antes de homologar dados reais

- Configurar no Site a credencial de leitura do Google Drive, o ID da pasta autorizada e o token de proteção. Nenhuma senha ou chave foi colocada no repositório.
- Executar uma importação controlada com arquivos reais e sanitizados, conferir cada layout e aprovar os contratos antes de produzir fatos.
- Confirmar permissão somente de leitura e a rotina diária de 09:30; o agendador ainda não foi criado.
- Medir pico de memória/carga em arquivo real e exercitar falha de rede, retomada e limpeza em VM limpa.
- Em 13/10/2026, Gisele confirmou que os arquivos diários do Drive também incluem Excel XLS e XLSX. Esses formatos continuam sem parser e são recusados; o requisito foi registrado em BK-042/BK-043. Antes de processá-los, implementar leitura em fluxo com limites de extração/tamanho e homologar a aba, os cabeçalhos e o tratamento dos valores de cada estrutura.
- CSV, TSV e JSON são apenas classificados e mantidos pendentes até a decisão de cada layout.

## Investigação técnica de Excel — 13/10/2026

Gisele confirmou que as planilhas `.XLS` e `.XLSX` chegam diariamente. A confirmação registra somente os formatos: não define qual aba usar, nem autoriza interpretar fórmulas, datas, valores ou transformar linhas em dados financeiros.

A ponte atual executa em JavaScript no Cloudflare Workers e recebe o arquivo como `ReadableStream`. Seu leitor não oferece acesso aleatório aos bytes. A documentação do Google Drive permite baixar trechos por `Range`; os Streams do Workers permitem transformar conteúdo em blocos limitados. Isso permite projetar leitura de `.XLSX` por intervalos e análise incremental do XML, mas exige implementar e limitar explicitamente o ZIP, o conteúdo descompactado e as tabelas compartilhadas.

Para `.XLS` (BIFF dentro de OLE/CFB), a busca por intervalo também é necessária para localizar setores do arquivo. Não encontramos um leitor JavaScript maduro que leia `.XLS` em fluxo e seja compatível com o Worker atual. O pacote `go-excel-reader` afirma ler `.XLS` e `.XLSX` linha por linha usando `io.ReadSeeker`, mas está na versão `0.x`, sem importadores conhecidos no pkg.go.dev, e mantém a tabela de textos compartilhados inteira em memória. Não é aprovado como dependência sem auditoria de código, limites e comportamento com arquivos reais. A alternativa Apache POI documenta leitores por eventos e baixo uso de memória para os dois formatos, mas requer um runtime Java, que o Worker atual não possui.

**Recomendação para decisão de arquitetura:** usar o processador Go já existente no Majucau para buscar a fonte autorizada em fluxo, guardá-la temporariamente com proteção local, percorrer as abas e linhas do Excel usando um leitor avaliado, e gravar staging em blocos. A ponte Site continuaria apenas no fluxo de autorização; ela não receberia ou interpretaria dados financeiros. Essa opção evita instalar Java na máquina ou criar um serviço público adicional, mas move a leitura do Drive da ponte Worker para o serviço Go e requer decisões de acesso ao Drive e revisão do parser. Nenhuma implementação foi iniciada antes de confirmar esse limite de arquitetura.

**Ponto para Gisele:** aprovar ou recusar que a leitura das planilhas passe pelo processador Go do Majucau, mantendo a interface e as regras financeiras atuais. Depois disso ainda será necessário decidir, com uma amostra sem dados pessoais, se devem ser percorridas todas as abas ou apenas uma; cada contrato continua pendente até aprovação.

Fontes técnicas consultadas em 13/10/2026:

- [Google Drive: download de arquivos e cabeçalho Range](https://developers.google.com/workspace/drive/api/guides/manage-downloads)
- [Cloudflare Workers: Streams](https://developers.cloudflare.com/workers/runtime-apis/streams/)
- [Apache POI: leitores de eventos para HSSF e XSSF](https://poi.apache.org/components/spreadsheet/)
- [go-excel-reader: documentação e limites declarados](https://pkg.go.dev/github.com/jloor/go-excel-reader@v0.3.0)

Não houve importação dos dados reais do Drive nesta execução. Nenhuma tela aprovada, fórmula, regra financeira ou regra de negócio foi alterada. `DATA-04` permanece `PARTIAL`.
