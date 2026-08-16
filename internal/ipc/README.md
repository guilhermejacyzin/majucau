# IPC local

O transporte Windows usa `\\.\pipe\Majucau-worker`, framing de 4 bytes em
ordem big-endian e payload JSON limitado a `MaxMessageSize` (1 MiB). O servidor
valida a versão, request id e método antes de chamar o handler; respostas nunca
incluem secrets.

O DACL é criado com SDDL explícito para os SID configurados, administradores
locais (`BA`) e sistema (`SY`). O instalador deve fornecer o SID exato da UI em
`MAJUCAU_UI_SID` e, quando necessário, o SID do serviço em
`MAJUCAU_SERVICE_SID`. Sem `MAJUCAU_UI_SID`, o fallback vincula o DACL ao SID do
processo do worker; isso é útil para desenvolvimento, mas não é suficiente
para a instalação final com UI e serviço sob identidades diferentes. Além do
DACL, cada conexão consulta o SID do processo cliente e rejeita clientes não
configurados.

Fora do Windows não há socket substituto: `FakeClient` cobre testes de contrato,
mantendo a propriedade de não abrir rede acidentalmente.
