# IPC local

O transporte Windows usa `\\.\pipe\Majucau-worker`, framing de 4 bytes em
ordem big-endian e payload JSON limitado a `MaxMessageSize` (1 MiB). O servidor
valida a versão, request id e método antes de chamar o handler; respostas nunca
incluem secrets.

O contrato v2 envia `dashboard.snapshot` em quadros sequenciais de até 512 KiB
de dados, com número de sequência e quadro final. O Named Pipe continua
limitado a 1 MiB por quadro; o cliente valida versão, request id, ordem e
tamanho antes de encaminhar os bytes para o decoder JSON. A consulta lê os
detalhes em fluxo, até 50 linhas por tabela, dentro de uma transação somente
leitura `REPEATABLE READ`; cada campo textual tem limite de 1 MiB e cada linha,
4 MiB. Se uma linha exceder esses limites, o painel recebe falha e não mostra
um snapshot incompleto. O JSON final mantém o contrato usado pela tela atual.

O servidor abre o pipe com `FILE_FLAG_OVERLAPPED` e conecta por uma operação
cancelável. Isso permite que a leitura que detecta a desconexão do cliente
aconteça ao mesmo tempo que a escrita da resposta. A UI recebe somente
`FILE_READ_DATA`, `FILE_WRITE_DATA`, `FILE_READ_ATTRIBUTES` e `SYNCHRONIZE`; o
SID do serviço também recebe `FILE_CREATE_PIPE_INSTANCE` para que o worker crie
novas instâncias do pipe. Isso evita conceder `GENERIC_ALL` ao cliente, pois no
named pipe ele inclui o direito de criar instâncias ([documentação Microsoft](https://learn.microsoft.com/en-us/windows/win32/ipc/named-pipe-security-and-access-rights)).
O instalador deve fornecer o SID exato da UI em
`MAJUCAU_UI_SID` e, quando necessário, o SID do serviço em
`MAJUCAU_SERVICE_SID`. Sem `MAJUCAU_UI_SID`, o fallback vincula o DACL ao SID do
processo do worker; esse caminho de desenvolvimento combina os direitos de
cliente/serviço para a mesma identidade. A instalação final deve configurar os
SIDs de UI e serviço quando forem identidades diferentes. Além do DACL, cada
conexão consulta o SID do processo cliente e rejeita clientes não configurados.

Fora do Windows não há socket substituto: `FakeClient` cobre testes de contrato,
mantendo a propriedade de não abrir rede acidentalmente.
