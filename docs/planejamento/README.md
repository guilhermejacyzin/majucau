# Arquivos do planejamento

## Referência vigente: v2

- [Backlog detalhado](outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md).
- [Análise e decisões](outputs/MAJUCAU_ANALISE_E_DECISOES_v2.md).
- [Planilha de evolução](outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx).
- Dados estruturados: [tarefas](work/backlog_v2.json), [decisões](work/decisions_v2.json), [cobertura](work/scope_v2.json) e [registro de alterações](work/revision_v2.json).
- [Verificações executadas na preparação original](work/verification_v2.json) e [prévias da planilha](work/previews_v2).

## Histórico preservado

| Versão | Backlog | Análise | Planilha |
|---|---|---|---|
| v0 — leitura original | [Documento](outputs/MAJUCAU_BACKLOG_DETALHADO_v0.md) | [Documento](outputs/MAJUCAU_ANALISE_E_DECISOES_v0.md) | [Excel](outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v0.xlsx) |
| v1 — primeira consolidação | [Documento](outputs/MAJUCAU_BACKLOG_DETALHADO_v1.md) | [Documento](outputs/MAJUCAU_ANALISE_E_DECISOES_v1.md) | [Excel](outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v1.xlsx) |

As v0/v1 conservam dúvidas, interpretações e regras depois substituídas. Consulte a v2 para trabalho futuro. As auditorias [financeira](work/financeiro.md) e [operacional/UX](work/operacao_ux.md) são leituras históricas das imagens.

## Arquivos de apoio

`work/` conserva dados JSON, extrações das vinte abas, auditorias, scripts de geração/revisão/verificação e prévias visuais. Esses scripts tratam os artefatos do backlog; não constituem o backend ou frontend do Majucau.

O [manifesto](MANIFESTO.json) relaciona arquivos e hashes SHA-256 deste pacote. Arquivos originais recebidos, dependências instaladas e saídas temporárias de inspeção não estão incluídos. O conteúdo criado a partir das fontes está preservado; os [originais são identificados separadamente](FONTES.md).

## Verificar e manter os artefatos

Execute a partir de `docs/planejamento`. A leitura e verificação de XLSX requer Python e `openpyxl`. A geração visual de Excel usa Node.js e `@oai/artifact-tool`, disponível no ambiente original de autoria; a biblioteca não é distribuída neste repositório e sua disponibilidade em outro ambiente deve ser verificada. As planilhas prontas podem ser abertas diretamente no Excel.

```sh
python work/verify_package.py
```

O comando acima verifica os artefatos incluídos, sem acessar Bling, Drive ou originais externos. Os scripts `verify_deliverables.py`, `verify_v1.py` e `verify_v2.py` preservam a verificação completa feita durante a autoria e também exigem os originais em `sources/private`, conforme [FONTES.md](FONTES.md).

Para recriar a planilha v2 a partir dos JSON existentes, em ambiente com a biblioteca de autoria disponível:

```sh
node work/build_tracker.mjs
```

O comando sobrescreve a planilha v2 e suas prévias; execute somente para atualizar conscientemente esses artefatos. Os scripts de revisão v1/v2 também reconstroem documentos a partir das versões anteriores e exigem o texto fonte correspondente. Não os execute sobre alterações manuais sem antes preservar essas alterações.

`prepare_tracker_v1.py` é um registro histórico da transformação inicial e dependia do gerador v0 daquela etapa, posteriormente substituído. Não faz parte do procedimento atual de regeneração. O gerador v1 final foi preservado em `build_tracker_v1.mjs`; `prepare_tracker_v2.py` produz o gerador v2 a partir dele.

A preparação para GitHub ajustou links para caminhos relativos e substituiu caminhos pessoais por `sources/private`. Não alterou os arquivos XLSX nem as regras do backlog.
