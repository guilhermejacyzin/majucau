# Leitor XLSX em fluxo — protótipo isolado (18/10/2026)

## O que está implementado

`internal/integrations/xlsxstream` contém um leitor de arquivos `.XLSX` que entrega uma linha por vez ao consumidor. Ele não carrega o arquivo inteiro na memória, não calcula fórmulas e não executa macros. A leitura rejeita partes externas e impõe limites explícitos de tamanho, conteúdo descompactado, quantidade de abas, linhas, colunas, células e texto. O índice temporário de textos compartilhados é criado em área temporária privada e removido ao final.

Foi corrigido um defeito na associação das abas: a validação substituía o identificador da relação numa cópia local, deixando o caminho original sem alteração. O fluxo completo tentava então abrir uma aba inexistente e devolvia erro de limite de segurança. Agora o caminho resolvido é mantido para a leitura de cada aba.

## Verificação

- `go test ./internal/integrations/xlsxstream -count=1` — aprovado em Windows, com arquivos de teste sintéticos.

## Limites e próximos passos

- O leitor ainda é um pacote isolado; não está ligado ao importador do Google Drive nem publica registros financeiros.
- Somente `.XLSX` é coberto. O formato antigo `.XLS` ainda não é suportado.
- Ainda faltam amostras autorizadas, contrato de cada planilha, integração com staging/transação do importador e ensaio de carga/limpeza em ambiente limpo.
- Por isso, arquivos Excel continuam pendentes na importação do produto e o avanço de BK-043 permanece em 40%.

Nenhuma tela, regra financeira ou regra de negócio foi alterada.

## Atualização de escopo — 19/10/2026

O status acima descreve o protótipo isolado em 18/10. Em 19/10, os leitores XLS e XLSX foram ligados à validação estrutural em fluxo das três pastas aprovadas, juntamente com CSV, TSV e JSON. Eles continuam sem interpretar valores financeiros e os arquivos seguem pendentes até a aprovação dos contratos. Consulte `DRIVE-ALL-SOURCE-STRUCTURAL-STREAMING-2026-10-19.md`.
