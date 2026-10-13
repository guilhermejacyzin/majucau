# Revisão das leituras de dados em fluxo — 13/10/2026

## Escopo

Revisão estática de código Go de produção e de `frontend/src`, procurando leituras que pudessem carregar arquivos ou respostas grandes inteiras na memória. A consulta não alterou código do produto, telas, regras financeiras ou dados.

## Resultado

- Não foram encontrados `os.ReadFile`, `ioutil.ReadAll`, `csv.ReadAll`, `FileReader`, `readAsText`, `readAsArrayBuffer`, `Blob.text()`, `response.text()` ou `response.json()` no escopo pesquisado.
- Os usos diretos de `io.ReadAll` encontrados são limitados: `internal/backup/backup_stream.go` lê uma entrada ZIP somente após validar seu tamanho e usa `io.LimitReader`; `internal/security/dpapi_windows.go` limita blobs do cofre a 1 MiB mais 64 KiB de margem.
- O cliente Bling decodifica as respostas JSON do endpoint em fluxo, com limite padrão de 4 MiB por resposta. A aplicação mantém a página atual e rejeita resposta que exceda o limite.
- O transporte IPC restringe cada mensagem a 1 MiB antes de decodificar JSON. A interface envia caminhos e pedidos pequenos ao worker, não o conteúdo completo dos arquivos importados.
- `internal/security.RedactJSON` também limita a entrada a 1 MiB antes de decodificar.
- Os importadores CSV do Bling e Nuvem Pago já leem linha a linha, limitam cada campo a 1 MiB, cada registro a 4 MiB e cada registro a 256 campos. O total de linhas permanece sem teto, conforme decisão registrada; a transação por pasta e o uso de WAL podem crescer mesmo com a memória em fluxo.

## Limites e próximos controles

Esta busca confirma apenas os padrões e caminhos citados. Não mede consumo real de memória, tempo de uma importação completa, crescimento de tabelas temporárias/WAL, impacto de falha de disco, permissões em instalação ou comportamento numa VM limpa. Essas validações continuam necessárias antes de fechar BK-024 e o gate de instalação. BK-024 permanece em 35%; esta evidência não equivale a homologação com dados reais.

Não foram executados testes nem builds nesta revisão; o resultado é uma inspeção estática. Nenhuma alteração de tela ou regra de negócio foi feita.
