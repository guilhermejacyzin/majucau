# Estágio privado e temporário de arquivos do Google Drive — 18/10/2026

## O que foi implementado

`Client.WithPrivateStage` agora compõe o download já verificado do Drive com o diretório temporário protegido do Majucau. O conteúdo é gravado em fluxo, primeiro com o sufixo interno `.partial`; só passa a ser lido pelo processador depois que o adapter confere o tamanho, o checksum disponível e a versão do arquivo antes e depois do download.

- O nome temporário é aleatório e não usa o nome fornecido pelo Drive como caminho.
- O leitor recebe `io.Reader`, metadados limitados, tamanho e SHA-256; o arquivo inteiro não é carregado pelo adapter na memória.
- O callback roda somente depois da validação do download e dentro da duração do estágio. Ao retornar, o workspace é removido.
- Em falha do callback, falha de rede ou retry, os bytes provisórios são descartados. O teste de resposta parcial confirma que a nova tentativa não recebe nem publica bytes repetidos.
- O limite de bytes configurado também é aplicado pelo destino local. A limpeza usa `securetemp`; se falhar, retorna `securetemp.CleanupError`, com o caminho disponível para manutenção autorizada e sem expô-lo na mensagem comum.
- A leitura do arquivo pelo callback é em fluxo. O consumidor futuro ainda deve fazer sua própria gravação de forma atômica e idempotente e terminar de ler antes de retornar.

## Verificação local

- `go test ./... -count=1` — **PASS** em Windows, incluindo os testes do adapter e do estágio privado.
- `go vet ./...` — **PASS**.
- `go mod verify` — **PASS**; todos os módulos foram verificados.
- `scripts/verify-backlog.ps1 -StructuredPath docs/planejamento/work/backlog_v2.json` e `scripts/verify-go-boundaries.ps1` — **PASS**.
- `git diff --check` — **PASS** antes desta atualização de evidência.
- Os testes do estágio cobrem entrega do conteúdo e metadados, remoção do temporário após sucesso e erro do consumidor, limite de tamanho e retry depois de resposta HTTP parcial.
- Nenhuma credencial, pasta real, conexão ao Drive, importação financeira, tela ou regra de negócio foi usada ou alterada.

## Limites e próximo passo

O worker diário continua atualizando apenas o inventário de metadados; ainda não chama o novo componente. Não há contrato de arquivo homologado, parser conectado nem publicação de dados financeiros. BK-010 e BK-043 ainda precisam fornecer amostras e contratos aprovados. XLS/XLSX seguem sem parser seguro. BK-042 foi ajustada de 56% para 58% estimados; faltam 42%, incluindo conexão e homologação reais. Próximo passo: ligar o callback aos importadores somente depois da homologação dos contratos das três estruturas.

Nenhuma tela aprovada ou regra financeira foi alterada.
