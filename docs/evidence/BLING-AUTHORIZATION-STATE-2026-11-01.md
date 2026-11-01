# Estado da credencial e autorização Bling — 2026-11-01

## Resultado desta conferência

- O arquivo protegido `bling_credentials_v1.dpapi` existe no cofre do worker do
  protótipo. A conferência verificou somente existência e metadados; nenhum
  segredo ou token foi aberto, copiado ou registrado.
- Na última captura da tela de Integrações enviada por Gisele, o Bling aparecia
  como `BLING_NOT_CONNECTED`, com permissões aguardando autorização e sem último
  sucesso. Gisele informou que iniciou a autorização e fechou a página antes de
  concluir.
- Portanto, a etapa de salvar as credenciais avançou. A captura não comprova
  que o fluxo OAuth foi concluído nessa instância do protótipo.
- Há também uma evidência separada de 01/11, em
  `BLING-LIVE-SYNC-SCOPE-2026-11-01.md`, que registrou uma chamada autorizada e
  HTTP 403 ao consultar categorias financeiras. Como os registros descrevem
  estados/execuções diferentes, não os tratei como uma única sessão atual. A
  autorização e a permissão de categorias precisam ser confirmadas na próxima
  tentativa feita pelo Majucau.

## Limites e segurança

- Não foi lida a página de cadastro do aplicativo Bling. A tentativa de acesso
  pela ferramenta de navegador foi encerrada porque ela não conseguiu
  identificar o endereço atual com confiança suficiente. Não houve tentativa
  de contornar essa proteção por outra ferramenta.
- Uma consulta ao Named Pipe do worker feita por um processo PowerShell também
  foi recusada pelas permissões do Windows. Ela não alterou dados. Não tentei
  contornar a ACL.
- O arquivo protegido permanece no cofre local, fora do Git. O protótipo e o
  PostgreSQL ativos foram preservados.

## Próxima etapa

Confirmar no próprio Majucau se o botão `Conectar` conclui a autorização. Se a
conta mostrar conexão ativa, testar primeiro a leitura autorizada e conferir a
permissão de consulta das categorias. Se a autorização voltar a ser interrompida,
registrar a mensagem sanitizada que a própria tela mostrar e corrigir essa mesma
etapa antes de seguir para a homologação da sincronização.

Nenhuma tela, regra financeira, credencial ou permissão externa foi alterada.
