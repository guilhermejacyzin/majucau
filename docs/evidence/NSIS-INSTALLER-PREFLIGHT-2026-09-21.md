# Evidência — preflight no instalador NSIS (2026-09-21)

## O que é

Integração do `installer-helper.exe` ao fluxo do instalador Windows x64.

## O que foi feito

- `build/windows/installer/project.nsi` agora extrai o helper para o diretório
  temporário do NSIS antes da primeira escrita no computador;
- `.onInit` executa `diagnostics`, que aplica o mesmo preflight usado no smoke
  portátil e grava um ZIP sanitizado em `%TEMP%\Majucau\installer-preflight.zip`;
- qualquer retorno diferente de zero interrompe a instalação; retorno `2`
  (`BLOCKED`) recebe mensagem operacional específica;
- o helper também é instalado junto ao aplicativo para diagnóstico posterior;
- `scripts/test-installer-nsis.ps1` verifica o contrato textual e a ordem
  `.onInit -> Section`;
- `scripts/verify.ps1` passa a executar essa verificação junto dos gates locais.

## Verificação executada

```powershell
./scripts/test-installer-nsis.ps1
```

Resultado: `PASS` (8 invariantes do instalador).

O binário NSIS final ainda não foi compilado neste ambiente porque `makensis`
não está disponível. Portanto esta evidência comprova a integração no fonte e
o guard automatizado, mas não substitui a compilação do instalador e o smoke em
VM Windows limpa.

## Compilação independente

Após a publicação do guard, o GitHub Actions executou o run **60** com sucesso:

- Go tests, vet e `govulncheck` passaram;
- bundle React, lint, typecheck, testes e audit passaram;
- Wails produziu o executável desktop;
- worker e `installer-helper.exe` foram compilados;
- `test-installer-nsis.ps1` passou;
- `makensis` produziu o instalador NSIS x64;
- o artefato `majucau-g1-unsigned` foi publicado pelo workflow.

Run verificável: https://github.com/guilhermejacyzin/majucau/actions/runs/35560415145

O artefato ainda é **unsigned** e o run não substitui a instalação em VM limpa
com reboot, upgrade, rollback e desinstalação.

## Próximo passo

Disponibilizar `makensis`/Wails no pipeline de build, produzir o instalador
assinado e executar instalação, reboot, upgrade, rollback e desinstalação em
Windows 10/11 limpo.

