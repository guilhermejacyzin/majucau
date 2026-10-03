# IPC local

O transporte Windows usa `\\.\pipe\Majucau-worker`, framing de 4 bytes em
ordem big-endian e payload JSON limitado a `MaxMessageSize` (1 MiB). O servidor
valida a versão, request id e método antes de chamar o handler; respostas nunca
incluem secrets.

O DACL é criado com SDDL explícito para os SID configurados, administradores
locais (`BA`) e sistema (`SY`). A UI recebe apenas `FILE_READ_DATA`,
`FILE_WRITE_DATA` e `SYNCHRONIZE`; o SID do serviço também recebe
`FILE_CREATE_PIPE_INSTANCE` para que o worker crie novas instâncias do pipe.
Isso evita conceder `GENERIC_ALL` ao cliente, pois no named pipe ele inclui o
direito de criar instâncias ([documentação Microsoft](https://learn.microsoft.com/en-us/windows/win32/ipc/named-pipe-security-and-access-rights)).
O instalador deve fornecer o SID exato da UI em
`MAJUCAU_UI_SID` e, quando necessário, o SID do serviço em
`MAJUCAU_SERVICE_SID`. Sem `MAJUCAU_UI_SID`, o fallback vincula o DACL ao SID do
processo do worker; esse caminho de desenvolvimento combina os direitos de
cliente/serviço para a mesma identidade. A instalação final deve configurar os
SIDs de UI e serviço quando forem identidades diferentes. Além do DACL, cada
conexão consulta o SID do processo cliente e rejeita clientes não configurados.

Fora do Windows não há socket substituto: `FakeClient` cobre testes de contrato,
mantendo a propriedade de não abrir rede acidentalmente.
