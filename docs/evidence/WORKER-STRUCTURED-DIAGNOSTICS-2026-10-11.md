# Registros técnicos estruturados do worker — 11/10/2026

## Estado

Implementação parcial de BK-031. A verificação da CI será registrada depois do
push. Este documento não declara o critério da tarefa completo.

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
- A validação do comportamento Windows/ACL e a suíte da CI ainda estão
  pendentes nesta revisão. Não foram executados testes locais.

## Limites de escopo

Nenhuma tela aprovada, regra financeira ou dado de negócio foi alterado. A
implementação cobre o limite IPC do worker; não registra cada chamada HTTP
individual do provedor.
