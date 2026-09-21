# Evidência do smoke NSIS — 2026-09-21

## Resultado

O pipeline Windows conseguiu executar toda a cadeia de qualidade e gerar o
instalador NSIS x64. O ciclo install/uninstall ainda não pode ser aceito como
verde: o instalador encerrou com código `2`, que é o contrato de `BLOCKED` do
preflight, antes de copiar os binários para o diretório de instalação.

Isso é uma evidência de bloqueio controlado, não de instalação bem-sucedida.
Não há base para declarar produção pronta a partir desse run.

## Runs observados

| Run | Commit | Build/testes anteriores | Smoke NSIS |
|---|---|---|---|
| 76 | `b4f40bc` | PASS: Go, frontend, Wails, worker, helper, NSIS e guard | `BLOCKED`/código 2; smoke inicial não localizou o bundle sanitizado |
| 77 | `6ec1ed0` | PASS: Go, frontend, Wails, worker, helper, NSIS e guard | `BLOCKED`/código 2; script registrou `PREFLIGHT_DIAGNOSTIC_NOT_FOUND` |
| 80 | `d1791aa` | PASS: Go, frontend, Wails, worker, helper, NSIS e guard | `BLOCKED`/código 2; diagnóstico determinístico `OS_UNSUPPORTED` |
| 81–83 | `685785a`–`2586c63` | PASS: cadeia completa de build, testes, NSIS e guard | PASS no CI; smoke desktop classificado como não elegível por `OS_UNSUPPORTED` |

O log público do job 77 confirma que as etapas 1–26 passaram e somente a
etapa “Smoke install and uninstall unsigned installer” falhou. A falha não
foi mascarada como sucesso.

## Interpretação técnica

- o executável desktop, worker, helper e instalador são compiláveis no Windows
  x64;
- o contrato NSIS continua válido e o helper de preflight é executado antes da
  instalação;
- a execução elevada via tarefa agendada não produziu evidência suficiente de
  que o token recebido no runner satisfaz o preflight;
- sem evidência em Windows limpo com elevação real não se aceita instalação,
  primeiro uso e desinstalação como gates de produção;
- o smoke foi extraído para `scripts/smoke-installer.ps1` e agora retorna os
  códigos de bloqueio sanitizados quando o diagnóstico é localizado;
- o run 80 confirmou que o bloqueio não é perda do ZIP: o runner hospedado é
  Windows Server (`ProductType != workstation`) e a política do produto o
  rejeita corretamente com `OS_UNSUPPORTED`;
- o CI passou a registrar esse caso como ambiente não elegível para o smoke
  desktop, sem enfraquecer a política nem declarar o gate de produção fechado.
- os runs 81–83 terminaram `success` porque o CI comprovou a cadeia de build e
  tratou o runner Server como inelegível; eles não são evidência de instalação,
  primeiro uso ou desinstalação em workstation.

## Próximo gate

Executar o mesmo instalador em uma VM Windows 10/11 limpa, com usuário
administrador interativo e WebView2 presente, capturando:

1. preflight `READY`;
2. instalação silenciosa concluída;
3. existência do desktop, worker, helper e uninstaller;
4. preflight pós-instalação com schema `1.0`;
5. primeiro uso e health do worker;
6. desinstalação sem diretório residual.

Até esse gate, G1/G6/G7 permanecem parciais.

