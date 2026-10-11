# Fontes do planejamento

Os originais foram fornecidos pelo usuário para análise. Este pacote publica os artefatos produzidos e a rastreabilidade; não duplica os arquivos originais recebidos. As extrações preservadas são referência histórica, sujeita às substituições expressas da v2.

## Material original

A pasta continha uma planilha e dezesseis PNGs. O PNG com nome ChatGPT repete a tela T10; são quinze imagens únicas. A planilha contém vinte abas, extraídas em `work/source_extract`.

| Arquivo recebido | SHA-256 |
|---|---|
| ChatGPT Image 12 de set. de 2026, 20_48_11.png | `cf04afe7be720d359fc783c4380d04efc785ba830e36925d93626f7a27b30f56` |
| MAJUCAU_CONSOLIDACAO_DEV_GATE_RC3_CEI_CONFIGURACOES_T16_T17.xlsx | `41e35e0460aab54430a8ca2ced518f06dd571ecde53287d52bc4b6910a5c6777` |
| T02_Fluxo_de_Caixa.png | `28e98d1be53a319cb8b4a83fae5103b57fabc65d0308a2b3eac175e6a4ad0d0a` |
| T03_Contas_a_Receber.png | `1844863e9a7cfb271ac9c5077c2dce29099093fcc7b885e703bea6180d0efde9` |
| T04_Contas_a_Pagar.png | `e0beba0257e027e925cd644e170903a9e7b8a463d3edf8d1d3d02d0a1c22126b` |
| T05_Conciliacao.png | `50b75ffb1ca4c95f1983acef64b4c4a8d5b47fbb7954da63743dba7de092b734` |
| T06_DRE.png | `94a5ac357405720e19cc2e92375dc0816aee1c1578d174bea4e03360d8497c02` |
| T07_Balancete.png | `20b3ef6b31af48983c8b18416010ba5023115b939374595b1c33600735252192` |
| T08_Calculadora_de_Aplicacao.png | `7987ae139fa1b0590635660578b73ba2ba859178d0517a46884f895ec2f61ac1` |
| T09_Estoque.png | `70887af8f35d3ab10a6af4f03859fb34c8e87ccf308d63316e1f20000cbc540f` |
| T10_Producao.png | `cf04afe7be720d359fc783c4380d04efc785ba830e36925d93626f7a27b30f56` |
| T12_Projecoes.png | `a99cd3b7fe9e8befaae36221f8cebc994d725e983cba61729cc5b752af58af51` |
| T13_Lancamentos_e_Pendencias.png | `5787c545ed0c29674fa41f7243bb2f52d9fc332996cbc65c0d6b5ef0076ea39a` |
| T14_Alertas.png | `b55f597b2fb945680a76d96a688a51f5c029623ed34b7ed585a4b312053ef731` |
| T15_Importacoes_e_Integracoes.png | `91626d3c6fb92b523113deeeb75502091f83105e1111c4db30134613ce361b49` |
| T16_Parametros.png | `ae91e304c0185c5908e3ac40497d98c6d3e63936ab9a35dc9a9ac8729329753c` |
| T17_Usuarios.png | `fcd45dde50ea33a1ed039931a4f39ff4b85aacc232c1f8b4cca64749f4261406` |

### Rechecagem preliminar do pacote inicial — 11/10/2026 (superada)

Na primeira busca daquele dia, os 17 arquivos ainda não estavam na pasta privada usada para a conferência. Essa observação ficou desatualizada quando o pacote original foi localizado. A extração continua preservada em `docs/planejamento/work/source_extract/`: `workbook.json` representa 20 abas (286.067 bytes; SHA-256 `1b8cee72563b88fd359ebc2a984d9f5cbd41c7cbe12f63b521924c043f6abe7c`) e `manifest.json` registra 17 itens históricos (2.596 bytes; SHA-256 `86d1bcd91d404925ee6d180c09c7c5c231b835d6e5b99f73d4ee72bf16f71cb6`). A constatação preliminar de ausência foi substituída pela conferência conclusiva abaixo.

### Conferência conclusiva do pacote inicial — 11/10/2026

- O pacote original contém exatamente 17 arquivos. O SHA-256 do ZIP é `323d1f246f0eeaf66e00cc3bfa5ee7a23a592d95fbd9968f9d45b6181378d95f`.
- Os 17 arquivos dentro do ZIP conferem com o inventário histórico por nome, tamanho e SHA-256. As 17 cópias em `sources/private/material-original/` também conferem com o mesmo inventário. O ZIP original foi preservado sem alteração.
- `sources/private/` é ignorada pelo Git. Os originais, portanto, ficam disponíveis localmente para conferência e não são incluídos nos commits nem enviados ao GitHub.
- A extração do livro contém 20 abas e seus arquivos mantêm os hashes registrados acima. A planilha binária extraída do pacote foi mantida junto dos demais originais na área privada. O PNG chamado T10 repete a imagem ChatGPT Image; a igualdade dos dois hashes confirma a duplicata registrada.
- Os 19 arquivos listados como baseline e mockups em `docs/source/README.md` conferem por SHA-256; os três nomes alternativos T1/C11 também são cópias exatas de suas referências.
- `revision_v1.json` e `revision_v2.json` estão preservados na área local de trabalho ignorada pelo Git. Os textos completos de consolidação e fechamento continuam ausentes; seus hashes e caminhos esperados permanecem registrados nesta documentação. As ausências não foram preenchidas por suposição.
- Esta conferência fecha o inventário, a preservação e a rastreabilidade das versões. Ela não é uma nova validação contábil nem uma revisão semântica de todas as células da planilha; esse trabalho permanece na tarefa técnica correspondente.

## Material complementar recebido em outubro de 2026

Em 11/10/2026, quatro arquivos de apoio à calculadora e relatórios financeiros foram encontrados nos Downloads e copiados para `sources/private/material-original/`. Os originais nos Downloads foram preservados. A pasta privada é ignorada pelo Git; o repositório não contém as planilhas nem os relatórios. Nome, tamanho e SHA-256 estão registrados apenas em `artifacts/retained/cleanup-register.md`. As duas cópias da calculadora têm conteúdo idêntico.

Na conferência de 11/10/2026, os quatro arquivos continuam presentes nessa pasta privada e seus SHA-256 correspondem ao registro local de limpeza. Eles são materiais complementares recebidos em outubro e não substituem o pacote histórico de 17 arquivos nem a planilha de vinte abas descrita acima. Os arquivos privados e o registro local seguem fora do Git.

## Consolidacao das respostas

Texto recebido e incorporado à versão 1. SHA-256: `ed8f04f469577e8772a91412edb6f49df65e5788beafd7f87920e226aa6aae42`.

Para repetir a verificação integral local, disponibilize o original em `sources/private/consolidacao-respostas.txt`.

## Fechamento das questoes

Texto recebido e incorporado à versão 2. SHA-256: `0f413addc1e0ab7e3b1d294b86c99c458a3439fa6661e73d6cef01d70ba10341`.

Para repetir a verificação integral local, disponibilize o original em `sources/private/fechamento-questoes.txt`.

## Reprodução da conferência dos originais

Os dezessete arquivos originais já estão preservados em `sources/private/material-original/`, com os nomes do inventário. A pasta é ignorada pelo Git. Para repetir a conferência em outro clone, copie os arquivos para esse caminho. Os textos completos de consolidação e fechamento continuam ausentes; se forem localizados, devem ser preservados nos caminhos indicados acima. Os originais não precisam ser publicados no Git para consultar os documentos ou verificar a estrutura do pacote.

A confirmação posterior “Somente atualizar o backlog” determina o escopo desta entrega. O pedido seguinte autoriza incluir no repositório o material produzido, sem iniciar implementação.
