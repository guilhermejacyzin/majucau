# Evidência — resiliência e preflight de instalação — 2026-08-16

## Veredito

O requisito de lidar defensivamente com máquinas Windows de terceiros foi registrado como `OPS-04` e recebeu uma primeira implementação verificável: matriz de 30 cenários e helper read-only de preflight. O preflight agora também bloqueia caminhos UNC/rede, reparse points e diretórios sincronizados por marcadores de OneDrive, Dropbox ou SharePoint, sem devolver o caminho no JSON. O estado permanece `PARTIAL`; esta evidência não substitui instalador NSIS integrado, elevação/UAC real, VMs limpas, repair, upgrade, rollback, backup/restore ou assinatura.

## Validações executadas

- `go test ./internal/installer ./cmd/installer-helper`: aprovado;
- `go vet ./internal/installer ./cmd/installer-helper`: aprovado;
- testes de segurança de localização do preflight (caminho local aceito; UNC/sincronizado bloqueado): aprovados;
- suíte completa `scripts/verify.ps1`: aprovada;
- `govulncheck`: nenhuma vulnerabilidade encontrada;
- frontend: lint, typecheck, 19 testes, build e `npm audit` aprovados;
- build Wails, worker e installer-helper Windows: aprovado;
- execução real `installer-helper.exe preflight`: contrato JSON válido, exit code `2`/`BLOCKED`;
- execução real `installer-helper.exe diagnostics`: produziu ZIP sanitizado mesmo com preflight `BLOCKED`, contendo somente `diagnostic.json` e `README.txt`;
- `JournalStore` e `AcquireInstallLock`: testes de round-trip, validação de fase/hash e segunda instância aprovados;
- execução real `majucau-worker.exe --console`: health `OK`;
- execução real `majucau.exe`: processo permaneceu ativo e abriu a janela `Majucau Financial Intelligence`.

## Resultado do preflight nesta máquina

| Check | Resultado |
|---|---|
| Plataforma | Windows x64 `10.0.19045`, suportado |
| Elevação | bloqueado: `INSTALL_NOT_ELEVATED` |
| Reboot pendente | não detectado |
| Espaço mínimo | aprovado; mais de 4 GiB disponíveis |
| Diretórios | aprovados sem criação/mutação |
| Porta PostgreSQL preferida 54329 | disponível no instante do probe |
| WebView2 | detectado, versão `151.0.4129.86` |

O bloqueio por elevação é esperado nesta sessão e prova comportamento fail-closed. O helper não solicitou UAC nem alterou registro, arquivos, serviços, firewall ou runtime. A verificação de porta é observacional e não reserva a porta; a seleção final ainda exigirá lock no instalador.

## Artefatos locais não versionados

| Arquivo | Bytes | SHA-256 | Authenticode |
|---|---:|---|---|
| `installer-helper.exe` | 4.206.592 | `81795ade359c1a4d19fdd1c9e9e59f458ce517f402125e1dfb00c9c28fb7b259` | `NotSigned` |
| `majucau-worker.exe` | 3.652.608 | `75d9c9974c5f9e026217e1094dfc7777487f344d927139014c8edf6b5279af71` | `NotSigned` |
| `majucau.exe` | 11.644.416 | `055f046eba93a769f95b288bce7ec1c234a631f11c7bca3d23fde210ef1e0fc8` | `NotSigned` |

Os executáveis ficam ignorados pelo Git e precisam ser reconstruídos e assinados no pipeline de release.

## Limites e próximos testes obrigatórios

1. integrar preflight, diagnostics, journal e mutex ao `.onInit` do NSIS e provar a passagem de exit codes;
2. fixar/instalar `makensis` e produzir o primeiro pacote;
3. implementar journal durável, lock global, retomada e rollback;
4. empacotar WebView2/PostgreSQL e aplicar serviços, dependências, ACL/SID e SCRAM;
5. executar a matriz `VM-01`–`VM-30`, preservando evidência por caso;
6. provar repair, reinstall, upgrade interrompido, downgrade bloqueado e uninstall preservando dados;
7. provar backup/restore real e reconexão DPAPI em outra máquina;
8. assinar e verificar todos os binários e o instalador.

Até esses itens passarem, o helper é uma fundação de detecção/diagnóstico, não uma alegação de instalação pronta para terceiros.
