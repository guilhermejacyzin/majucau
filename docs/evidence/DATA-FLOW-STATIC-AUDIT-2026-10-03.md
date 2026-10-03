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

## Pendência: resposta do snapshot do painel

O SQL do dashboard limita a 50 linhas de contas a receber e 50 de contas a
pagar, mas serializa os dois conjuntos como JSON. Alguns campos selecionados,
como nomes, documentos e categorias, são `text` sem limite de comprimento na
migration. O IPC limita a resposta a 1 MiB somente quando escreve o quadro:
nesse momento o handler já materializou os arrays e serializou a resposta.
Assim, o limite evita enviar um quadro maior, mas não impede o pico de memória
anterior nem garante ao cliente uma mensagem de erro clara.

A regra para conteúdo acima de 1 MiB — erro claro usando o limite atual ou
paginação/streaming com contrato aprovado — foi perguntada à usuária e aguarda
resposta. Nenhuma tela ou regra de negócio foi alterada nesta auditoria.

## Resultado

A maior parte dos caminhos revisados já tem streaming ou limite de tamanho
explícito. `DATA-04` continua `PARTIAL` até resolver e validar o contrato de
resposta IPC do dashboard, terminar a auditoria dos demais fluxos fora deste
recorte e comprovar a limpeza/ACL em VM Windows real.
