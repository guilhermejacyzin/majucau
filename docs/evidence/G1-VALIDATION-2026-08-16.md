# Evidência de validação G1 — 2026-08-16

## Escopo e veredito

Esta evidência cobre a fundação executável do G1: domínio financeiro inicial, schema PostgreSQL, acesso SQL gerado, cofre DPAPI, worker/Named Pipe, bootstrap Wails, interface React, verificações automatizadas e geração dos dois executáveis Windows.

**Veredito:** G1 permanece `PARTIAL`. A fundação compila e suas verificações locais passam, mas o instalador não foi produzido porque `makensis` não está instalado; PostgreSQL ainda não está embutido/registrado; o worker ainda não implementa os métodos de credenciais/OAuth/sincronização; não há assinatura Authenticode nem smoke em VM limpa. Nenhum release de produção foi aprovado.

## Ambiente reproduzido

| Componente | Versão |
|---|---:|
| Windows | x64 |
| Go | 1.26.6 |
| Node.js | 24.14.0 |
| npm | 11.9.0 |
| Wails | 2.13.0 |
| sqlc | 1.31.1 |
| govulncheck | 1.7.0; base atualizada em 2026-08-14 |
| PostgreSQL usado no teste descartável | 18.6 x64 |

O ZIP oficial do Go 1.26.6 usado localmente teve SHA-256 `5b6c5b556525810463b5c897b50dc7a82d6a3dc0bfaf55d990a7e9f31d6b2318`. O ZIP EDB/PostgreSQL 18.6 teve SHA-256 `fbe23da234ee31547bf8a36d29dfd81e82b849df2d2b78d2eecb43d360252f8c`. Os binários individuais do ZIP PostgreSQL não apresentaram assinatura Authenticode; isso é risco de cadeia de suprimentos a resolver antes do empacotamento final.

## Pipeline local executado

`scripts/verify.ps1` foi executado com sucesso após tornar falhas de comandos nativos fatais. Resultado:

- `go test . ./cmd/... ./database/... ./internal/...`: aprovado;
- `go vet . ./cmd/... ./database/... ./internal/...`: aprovado;
- `govulncheck . ./cmd/... ./database/... ./internal/...`: **No vulnerabilities found**;
- checksums de migrations: aprovado;
- `sqlc generate`: determinístico após atualização do baseline gerado;
- scan local de secrets: aprovado;
- ESLint sem warnings: aprovado;
- TypeScript `--noEmit`: aprovado;
- Vitest: 3 arquivos e 12 testes aprovados;
- Vite production build: aprovado;
- `npm audit`: zero vulnerabilidades.

Durante a preparação, a toolchain Go 1.26.5 acusou `GO-2026-6090` e `GO-2026-5972`; a atualização para Go 1.26.6 eliminou os achados alcançáveis. `golang.org/x/text` também foi atualizado para 0.39.0 depois de um achado de pacote sem símbolo chamado.

## PostgreSQL real e migration

A migration `000001_init.up.sql` foi aplicada com `ON_ERROR_STOP` em cluster PostgreSQL 18.6 efêmero, restrito a loopback e encerrado após o teste.

- 50 tabelas, 7 views e 103 foreign keys criadas;
- duplicidade RAW por hash exato rejeitada;
- alteração direta de payload RAW rejeitada por trigger;
- rollover RAW manteve duas versões e exatamente uma corrente;
- redação LGPD gerou tombstone, preservou hash anterior e acrescentou auditoria;
- role `majucau_runtime` não conseguiu atualizar RAW diretamente;
- a mesma role conseguiu realizar rollover pela função autorizada;
- grants mínimos estão versionados em `database/privileges/runtime.sql`;
- checksum SHA-256 da migration: `22fed94f72c3128eb8d7ac0c162ba99f19e1bb8222dc3531e2a0ac0252fbbd01`.

## Segurança local e IPC

- teste DPAPI real comprovou cifragem em repouso, reabertura, rotação atômica e remoção;
- o teste revelou e levou à correção de um use-after-free no buffer devolvido por `CryptProtectData`/`CryptUnprotectData`;
- criação concorrente da entropia passou a usar `O_EXCL`, evitando cofres incompatíveis;
- escrita do segredo passou a usar arquivo temporário sincronizado e `MoveFileExW` com replace/write-through;
- Named Pipe usa framing de 4 bytes, limite de 1 MiB, protocolo versionado, DACL e autorização de SID;
- testes Windows reais cobriram health e rejeição de cliente não autorizado;
- o instalador ainda precisa materializar a conta de serviço e as ACLs esperadas; portanto, a fronteira não está homologada em instalação limpa.

## Frontend e bootstrap desktop

- 20 cards oficiais são exibidos sem inventar números;
- todos os componentes funcionais têm tooltip simples, acessível por foco/hover e fechável por Escape/blur;
- Bling, Nuvemshop e Nuvem Pago têm cards independentes;
- o frontend chama `window.go.main.App.GetBootstrapState()` por adapter testável;
- loading, worker indisponível, resposta inválida e falha sanitizada têm testes;
- status reais do worker e das integrações alimentam rodapé/cards;
- entradas de credenciais não persistem no estado do frontend;
- ações de conectar/testar/sincronizar permanecem desabilitadas até o protocolo seguro do worker existir;
- inspeção visual em viewport 1440×900 confirmou hierarquia, estados fail-closed e ausência de valores fictícios.

## Artefatos Windows produzidos

| Artefato | Bytes | SHA-256 | Authenticode |
|---|---:|---|---|
| `majucau.exe` | 11.644.416 | `055f046eba93a769f95b288bce7ec1c234a631f11c7bca3d23fde210ef1e0fc8` | `NotSigned` |
| `majucau-worker.exe` | 3.652.096 | `23a659ebbab09f8639042b53986a0d232e145759af41630faf2f6765f3c643ab` | `NotSigned` |

Os executáveis são evidência local e ficam ignorados pelo Git. O pipeline do GitHub deve reconstruí-los a partir do código. A tentativa NSIS foi encerrada sem pacote porque `makensis` não estava disponível; o script agora detecta ausência do artefato e falha, em vez de produzir falso sucesso.

## Riscos e próximos critérios obrigatórios

1. instalar/fixar NSIS e construir o instalador reproduzível;
2. embutir PostgreSQL, criar cluster/roles/SCRAM/ACLs e registrar os serviços;
3. implementar métodos IPC de credencial, OAuth externo, teste e sync;
4. validar contratos reais e idempotência com fixtures sanitizadas;
5. implementar backup/restore real e testar perda/upgrade/rollback;
6. gerar SBOM e relatório de licenças;
7. assinar UI, worker e instalador com Authenticode;
8. executar instalação, reboot, upgrade e uninstall em Windows 10/11 limpos;
9. obter aceite funcional dos resultados dependentes das fontes oficiais.

Até essas evidências existirem, a nomenclatura correta é **fundação G1 parcial**, não “aplicativo pronto para produção”.
