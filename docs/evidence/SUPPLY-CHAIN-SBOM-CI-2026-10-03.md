# Inventário SBOM e evidências de licença — 2026-10-03

## Geração e retenção

- O CI Windows instala `cyclonedx-gomod` na versão fixada `v1.12.0` e gera um CycloneDX a partir do grafo Go com `-licenses`; o detector registra licenças como evidências quando consegue identificá-las.
- O npm nativo gera CycloneDX a partir de `frontend/package-lock.json`, sem instalar uma ferramenta npm adicional.
- Os dois arquivos ficam em `artifacts/verification/sbom/`, são incluídos no artefato consolidado existente e têm limite de tamanho de 64 MiB cada.
- CI [37148352628](https://github.com/guilhermejacyzin/majucau/actions/runs/37148352628) passou nos jobs Windows e PostgreSQL. A etapa de geração concluiu, e a verificação confirmou que os arquivos existem, não estão vazios e respeitam o limite.
- CI [37150221735](https://github.com/guilhermejacyzin/majucau/actions/runs/37150221735) passou nos jobs Windows e PostgreSQL e confirmou que os dois SBOMs passam pela validação de schema.

## Validação de schema

- O CI baixa CycloneDX CLI `v0.33.1` para `artifacts/cache/bin/`, confere o SHA-256 `9b360974c41a5d612eb026fb5e3bb1fa28e08865f7a4f915665b8bc94bd3bc31` do asset oficial Windows x64 e valida os arquivos antes de publicar o inventário.
- O SBOM Go é validado como CycloneDX 1.6; o SBOM npm é validado como CycloneDX 1.5. `--fail-on-errors` transforma qualquer erro de schema em falha da CI.
- A validação foi executada com sucesso no CI [37150221735](https://github.com/guilhermejacyzin/majucau/actions/runs/37150221735). O CLI é ferramenta de verificação do runner e não é distribuído com o aplicativo.
- A etapa não gera um relatório tabular das licenças npm. A existência de evidências de licença não é uma aprovação jurídica.

## Lacunas de licença e escopo

- O detector registrou `no licenses detected` para o módulo principal `majucau.local/financial-intelligence` e para `std@go1.26.6`. O repositório não tem arquivo `LICENSE`.
- Nenhuma allow-list ou licença foi escolhida por automação. A licença do código Majucau depende de decisão explícita; o warning, sozinho, não prova qual licença deve ser usada.
- Os inventários cobrem o grafo Go e o lockfile npm. Eles não bastam para demonstrar o inventário de todos os componentes nativos e redistribuídos do instalador, como WebView2, PostgreSQL e NSIS.
- Nenhuma tela aprovada ou regra de negócio foi alterada.

## Artefato de CI

- Instalador unsigned G1, artefato `11283552789`, tamanho 37,497,279 bytes, SHA-256 `ccf93a2b48c375a97d8a707838577b4648b645d62cc09443d962327f4c8712ac`.
- Expira automaticamente em `2026-10-04 20:11:27 UTC`; a ferramenta conectada pode listar e baixar, mas não apagar o artefato. A instrução para exclusão manual está no registro local `artifacts/retained/cleanup-register.md`.

## Fontes primárias

- [CycloneDX CLI v0.33.1](https://github.com/CycloneDX/cyclonedx-cli/releases/tag/v0.33.1), release usada na CI; o README informa a licença Apache 2.0 e documenta o comando `validate` com seleção da versão do schema.
- [Documentação oficial `npm sbom`](https://docs.npmjs.com/cli/v11/commands/npm-sbom/), que exemplifica o formato CycloneDX 1.5.
- [Documentação oficial `cyclonedx-gomod`](https://github.com/CycloneDX/cyclonedx-gomod), que informa suporte até CycloneDX 1.6.
