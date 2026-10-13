# Ingestão do Google Drive em fluxo — 13/10/2026

## O que foi entregue

A ponte do Majucau agora lê somente os arquivos das três estruturas já aprovadas do Drive. A listagem usa páginas de até 100 arquivos; o conteúdo baixa como fluxo, linha por linha, para CSV/TSV e registro por registro para JSON. O processamento calcula SHA-256 durante a leitura, sem carregar o arquivo inteiro na memória.

Antes e depois do download, o sistema confere os metadados da mesma versão do arquivo: ID, nome, pasta, tamanho, hash informado pelo Drive, data de alteração e versão. Se o arquivo for modificado, movido ou removido durante a leitura, as linhas já preparadas daquele arquivo são apagadas e o lote fica indisponível.

Cada campo pode ter até 1 MiB; cada linha, 4 MiB; cada registro, 256 campos. JSON aceita até 64 níveis e listas de até 256 itens. O banco D1 limita uma linha a 2.000.000 bytes; a aplicação rejeita o arquivo se um registro convertido para JSON passar de 1.500.000 bytes ou se a chave candidata passar de 32 KiB.

As gravações temporárias usam blocos de 40 registros. O lote fica `IMPORTING` enquanto lê. Só depois de terminar a leitura, o hash, as validações estruturais e a auditoria é marcado como pendente de decisão. Uma falha remove o staging parcial; lotes interrompidos por encerramento inesperado são recuperados na próxima importação. Arquivos repetidos são detectados por hash e as contagens temporárias de duplicatas são apagadas ao fechar o lote. Nenhuma linha é promovida a fato financeiro.

Chamadas externas repetíveis são somente GET e usam no máximo três tentativas, com espera aleatória crescente e `Retry-After` limitado a 30 segundos. O serviço de credencial é POST e não é repetido automaticamente. Cancelamentos não são repetidos.

## Publicação e verificação

- Código do Site publicado: versão 15, commit de origem `8edf1b4a57772ea0c1ba3cd1e452ba95806210d3`.
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
- XLS/XLSX continuam sem parser e são recusados; CSV, TSV e JSON são apenas classificados e mantidos pendentes até a decisão de cada layout.

Não houve importação dos dados reais do Drive nesta execução. Nenhuma tela aprovada, fórmula, regra financeira ou regra de negócio foi alterada. `DATA-04` permanece `PARTIAL`.
