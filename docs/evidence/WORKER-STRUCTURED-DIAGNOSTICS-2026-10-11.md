# Registros técnicos estruturados do worker — 11/10/2026

## Estado

Implementação parcial de BK-031. A CI `38116971454`, no commit
`ca7237e8b3e337a16d0ef50aec9643d778f3d24d`, terminou com sucesso em Windows e
PostgreSQL. Depois, a CI `38122175947`, no commit `1f58ffc3c360580d545049fbe07af6518fe5af74`,
passou nos jobs Windows e PostgreSQL/Linux. Foram executados testes próprios e
`go vet` no pacote de logs no Linux, a suíte Go completa e `go vet` no Windows,
além das verificações gerais de segurança, frontend, build, instalador e
consolidação de outputs. O teste da ACL em uma instalação real/VM ainda não foi
feito. Este documento não declara o critério da tarefa completo.

## O que foi preparado

- O worker escreve eventos JSON em
  `%ProgramData%\Majucau\Logs\majucau-worker.jsonl`.
- Cada evento pode conter operação, origem conhecida (Bling, Nuvemshop,
  Nuvem Pago ou local), identificador da requisição, resultado, duração,
  código de erro estável, ID de lote e contagens de processamento.
- O registro não inclui o payload da chamada, mensagem técnica bruta, valores
  financeiros, token, segredo, nome de arquivo ou caminho da pasta importada.
- O identificador da requisição é validado antes de entrar no log; IDs de lote
  são aceitos somente no formato UUID. Status e códigos passam por validação de
  formato.
- No Windows, a pasta só é usada se a pasta de dados tiver ACL privada. A pasta
  `Logs` é criada com DACL protegida para Owner Rights, LocalSystem e
  Administradores. Links/reparse points ou arquivos com permissões inesperadas
  são recusados.
- Cada arquivo é limitado a 10 MiB, são mantidos até cinco arquivos e cada
  evento tem limite de 32 KiB. O uso normal fica limitado a cerca de 50 MiB.
  A rotação substitui a cópia mais antiga.
- A criação do gravador rejeita limites inválidos. Um evento que ultrapasse o
  tamanho máximo do arquivo é recusado antes de gravar, mesmo que esteja abaixo
  do limite geral de 32 KiB.
- Testes próprios cobrem limites inválidos, rejeição de evento grande, rotação
  e retenção dos arquivos mais recentes, JSON estruturado e permissões privadas
  no ambiente Linux. A CI Windows executa o teste de limites inválidos; a suíte
  Go e `go vet` também passaram nesse runner.
- Se o log em arquivo não puder ser aberto, a aplicação continua sem mudar o
  resultado financeiro e emite somente um aviso genérico no `stderr`; não
  registra o caminho nem o erro bruto de sistema.

## O que ainda falta

- A tela atual mostra códigos de erro, mas descarta o identificador único da
  chamada. Por isso, ainda não é possível filtrar com precisão um erro exibido
  pelo usuário. A tela está protegida e a mudança mínima para exibir uma
  referência está aguardando decisão da responsável.
- Os eventos permitem contar falhas e calcular a duração observada por
  operação. Ainda não há painel/endpoint de métricas nem agregação persistente.
- Não existe fila de processamento de lotes no worker síncrono atual. A métrica
  de profundidade da fila deve ser definida quando uma fila real for introduzida;
  não se deve reportar uma fila inexistente como tamanho zero.
- A validação de runtime da ACL na conta instalada e o teste em uma VM limpa
  continuam pendentes. No Windows local, `go test ./internal/diagnosticslog
  -count=1` e `go vet ./internal/diagnosticslog` passaram; a validação dos
  arquivos em diretório temporário é feita no Linux porque a proteção Windows
  corretamente recusa diretórios temporários sem a ACL privada exigida.

## Limites de escopo

Nenhuma tela aprovada, regra financeira ou dado de negócio foi alterado. A
implementação cobre o limite IPC do worker; não registra cada chamada HTTP
individual do provedor.

O run gerou o artefato unsigned `majucau-g1-unsigned` (ID `11694326007`,
55.971.795 bytes, expiração em 2026-10-12 06:13:32 UTC). A conexão atual não
permite apagá-lo; ele foi incluído no registro local
`artifacts/retained/cleanup-register.md` para expiração automática ou exclusão
manual no GitHub Actions.

A CI `38122175947` também gerou o artefato unsigned `majucau-g1-unsigned`
(ID `11696115469`, 55.987.769 bytes, SHA-256
`4025c2781e022de5cfeed2af0d6b5a42a05b38a2c56a3f6b2786f6445e01d2e9`), com
expiração automática em 2026-10-12 07:47:26 UTC. Ele está registrado no
controle local de limpeza. Nenhum artefato unsigned é uma versão de produção.

## Complemento — referência técnica junto às mensagens de erro — 11/10/2026

- A responsável autorizou acrescentar uma referência curta ao texto de erro,
  preservando posição, estrutura e desenho da tela aprovada.
- `app.go` agora inclui o `request_id` já usado pelo worker nos logs nas
  mensagens de falhas recebidas nas operações de saúde, painel, integrações,
  sincronização e importação. A mensagem informa “Referência técnica” seguida
  pelo mesmo ID que aparece no evento `worker_request_completed`; o código de
  erro continua separado e não muda.
- O ID é acrescentado somente quando existe um código de erro e uma chamada
  identificada. Falhas anteriores ao envio ao worker não recebem referência,
  pois não há evento correspondente para localizar. A mensagem não inclui
  segredos, conteúdo financeiro ou detalhes brutos do erro.
- Revisão feita com `gofmt` e `git diff --check`. Nenhum teste ou build local foi
  executado nesta etapa. A validação automatizada da CI e a conferência visual
  da referência no protótipo continuam pendentes.
- A mudança altera somente o texto complementar autorizado. Nenhum layout,
  regra financeira, regra de negócio ou dado persistido foi alterado.
