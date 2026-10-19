# Proteção da configuração das pastas do Drive — 19/10/2026

## Problema encontrado

O worker aceitava IDs das três pastas por variáveis de ambiente quando não encontrava os IDs no cofre. Isso permitia habilitar ou redirecionar a leitura sem usar o provisionamento aprovado, que grava a chave de serviço e os IDs juntos em uma configuração protegida por DPAPI.

## Correção implementada

- Removida a leitura dos IDs de variáveis de ambiente.
- O worker agora só habilita a integração quando encontra os três IDs dentro do pacote protegido que também contém a chave da conta de serviço.
- Uma entrada antiga no cofre que contenha somente a chave não habilita o acesso: o Drive permanece pendente até o pacote completo ser provisionado.
- Incluído teste regressivo que define os nomes antigos de variáveis e confirma que elas não habilitam a leitura.
- Documentação histórica foi marcada como superada para evitar instruções conflitantes.

Nenhuma tela aprovada, regra financeira, dado do Drive ou permissão da conta foi alterada. Nenhuma chamada à API do Drive foi feita.

## Verificações

- `scripts/verify.ps1` passou: testes Go, `go vet`, `govulncheck`, checksum das migrations, geração SQLC sem diferenças, scan de secrets, validação do instalador, lint, TypeScript, 33 testes de interface, build e `npm audit` sem vulnerabilidades.
- `govulncheck` encontrou zero vulnerabilidades alcançáveis no código e um alerta em módulo requerido sem chamada alcançável, já acompanhado em `docs/evidence/DEPENDENCY-SECURITY-REMEDIATION-2026-10-10.md`.
- Os testes que dependem de PostgreSQL real não foram executados localmente porque `MAJUCAU_TEST_DATABASE_URL` não está configurado; a cobertura PostgreSQL será confirmada pela CI depois do push.
- O teste regressivo `TestGoogleDriveFolderIDsRequireProtectedBundle` passou junto com a suíte do worker: variáveis antigas não configuram os IDs das pastas.
- Pendente: CI do GitHub para o commit desta correção.
- A homologação real ainda depende da conta técnica Google, da permissão somente de leitura nas três pastas e do provisionamento pelo worker.

## Próximo passo

Após a confirmação de identidade no Google Cloud Console, verificar se já existe a conta técnica aprovada. Provisionar a credencial com acesso somente de leitura nas três raízes, gravar chave e IDs pelo mecanismo DPAPI e validar a primeira sincronização real.
