# Inventário SBOM e evidências de licença — 2026-10-03

## Geração e retenção

- O CI Windows instala `cyclonedx-gomod` na versão fixada `v1.12.0` e gera um CycloneDX a partir do grafo Go com `-licenses`; o detector registra licenças como evidências quando consegue identificá-las.
- O npm nativo gera CycloneDX a partir de `frontend/package-lock.json`, sem instalar uma ferramenta npm adicional.
- Os dois arquivos ficam em `artifacts/verification/sbom/`, são incluídos no artefato consolidado existente e têm limite de tamanho de 64 MiB cada.
- CI [37148352628](https://github.com/guilhermejacyzin/majucau/actions/runs/37148352628) passou nos jobs Windows e PostgreSQL. A etapa de geração concluiu, e a verificação confirmou que os arquivos existem, não estão vazios e respeitam o limite.
- O pipeline ainda não valida o schema CycloneDX nem apresenta um relatório tabular de licenças npm. Não tratar a existência dos arquivos como revisão jurídica concluída.

## Lacunas de licença e escopo

- O detector registrou `no licenses detected` para o módulo principal `majucau.local/financial-intelligence` e para `std@go1.26.6`. O repositório não tem arquivo `LICENSE`.
- Nenhuma allow-list ou licença foi escolhida por automação. A licença do código Majucau depende de decisão explícita; o warning, sozinho, não prova qual licença deve ser usada.
- Os inventários cobrem o grafo Go e o lockfile npm. Eles não bastam para demonstrar o inventário de todos os componentes nativos e redistribuídos do instalador, como WebView2, PostgreSQL e NSIS.
- Nenhuma tela aprovada ou regra de negócio foi alterada.

## Artefato de CI

- Instalador unsigned G1, artefato `11283034233`, tamanho 37,497,430 bytes, SHA-256 `3ce027a1b86b40312bc21f228d93c822f0982ea364c8dc51ce3a231fc5535b6c`.
- Expira automaticamente em `2026-10-04 19:42:29 UTC`; a ferramenta conectada pode listar e baixar, mas não apagar o artefato. A instrução para exclusão manual está no registro local `artifacts/retained/cleanup-register.md`.
