# Arquivos do planejamento

## Referência vigente: v2

Atualização de escopo em 31/10/2026: a tela independente T05 — Conciliação foi retirada. A documentação v2 foi atualizada para registrar essa decisão; a imagem aprovada original segue apenas no arquivo histórico. Regras internas de correspondência continuam no escopo.

Atualização T02 em 07/11/2026: o saldo acumulado do gráfico segue o cálculo do relatório Bling e é diferente do card de saldo bancário, que usa somente Itaú. A análise, a fórmula conferida na referência, o código preparado, os testes e as pendências estão em [`T02-BLING-WEEKLY-REPORT-AND-ITAU-BALANCE-2026-11-07.md`](../evidence/T02-BLING-WEEKLY-REPORT-AND-ITAU-BALANCE-2026-11-07.md).

Atualização T02 em 08/11/2026: o ponto de partida aprovado do Itaú fica somente no banco local isolado. O card mostra esse último saldo conciliado; movimentos registrados posteriores recebidos pela API ajustam a abertura projetada, e cada semana calcula saldo de abertura + entradas + saídas negativas = saldo de fechamento. O fechamento de cada semana inicia a próxima. A migration está aplicada e o worker tem leitura, sem permissão de gravação, para a tabela do checkpoint. O valor não existe no código ou Git. Os indicadores permanecem parciais até uma carga atualizada e comparação com Caixas e Bancos; falta conferir a T02 visualmente no protótipo. Evidência em [`T02-LAST-ITAU-BALANCE-API-2026-11-08.md`](../evidence/T02-LAST-ITAU-BALANCE-API-2026-11-08.md).

- [Backlog detalhado](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v2.md).
- [Análise e decisões](../../artifacts/planejamento/outputs/MAJUCAU_ANALISE_E_DECISOES_v2.md).
- [Planilha de evolução](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v2.xlsx).
- [Limites da arquitetura Go](../architecture/GO-MODULAR-BOUNDARIES.md).
- Dados estruturados: [tarefas atuais](work/backlog_v2.json), [decisões](work/decisions_v2.json), [cobertura](work/scope_v2.json) e [registro histórico de alterações de 13/09/2026](work/revision_v2.json). O registro histórico não substitui a situação atual das tarefas.
- [Verificações executadas na preparação original](work/verification_v2.json) e [prévias históricas da planilha](work/previews_v2). Novas prévias temporárias são geradas em `../../artifacts/verification/backlog-v2` e removidas depois da conferência.

## Histórico preservado

| Versão | Backlog | Análise | Planilha |
|---|---|---|---|
| v0 — leitura original | [Documento](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v0.md) | [Documento](../../artifacts/planejamento/outputs/MAJUCAU_ANALISE_E_DECISOES_v0.md) | [Excel](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v0.xlsx) |
| v1 — primeira consolidação | [Documento](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_DETALHADO_v1.md) | [Documento](../../artifacts/planejamento/outputs/MAJUCAU_ANALISE_E_DECISOES_v1.md) | [Excel](../../artifacts/planejamento/outputs/MAJUCAU_BACKLOG_E_EVOLUCAO_v1.xlsx) |

As v0/v1 conservam dúvidas, interpretações e regras depois substituídas. Consulte a v2 para trabalho futuro. As auditorias [financeira](work/financeiro.md) e [operacional/UX](work/operacao_ux.md) são leituras históricas das imagens.

## Arquivos de apoio

`work/` conserva dados JSON, extrações das vinte abas, auditorias e scripts de geração/revisão/verificação. Relatórios oficiais versionados ficam em `../../artifacts/planejamento/outputs/`; caches e saídas temporárias ficam sob `../../artifacts/cache/` e `../../artifacts/verification/` e podem ser removidos depois do uso. Esses scripts tratam os artefatos do backlog; não constituem o backend ou frontend do Majucau.

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

O comando sobrescreve a planilha v2 em `../../artifacts/planejamento/outputs/` e cria prévias temporárias em `../../artifacts/verification/backlog-v2/`; execute somente para atualizar conscientemente esses artefatos. Os scripts de revisão v1/v2 também reconstroem documentos a partir das versões anteriores e exigem o texto fonte correspondente. Não os execute sobre alterações manuais sem antes preservar essas alterações.

`prepare_tracker_v1.py` é um registro histórico da transformação inicial e dependia do gerador v0 daquela etapa, posteriormente substituído. Não faz parte do procedimento atual de regeneração. O gerador v1 final foi preservado em `build_tracker_v1.mjs`; `prepare_tracker_v2.py` produz o gerador v2 a partir dele.

A preparação para GitHub ajustou links para caminhos relativos e substituiu caminhos pessoais por `sources/private`. Não alterou os arquivos XLSX nem as regras do backlog.
