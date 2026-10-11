# Evidência — publicação condicionada do pacote de CI — 2026-10-11

## Regra e correção

A etapa `Upload unsigned G1 build` do workflow `.github/workflows/ci.yml`
usava `if: always()`. Isso permitia tentar publicar um pacote mesmo depois de
uma verificação anterior falhar. A condição foi alterada para `if: success()`;
como é a última etapa do job Windows, qualquer falha anterior faz a etapa ser
ignorada.

## Verificação executada

- A CI [38115370056](https://github.com/guilhermejacyzin/majucau/actions/runs/38115370056), no commit `c5eac57`, concluiu com sucesso nos jobs `verify-windows` e `verify-postgres`.
- No job Windows, a consolidação, a validação da consolidação e o envio do artefato terminaram com `success`.
- O caso de sucesso está comprovado pela CI. Não foi provocado um caso de falha anterior para verificar dinamicamente o caminho em que o envio é ignorado; esse comportamento se apoia na condição explícita `success()` do GitHub Actions.
- As verificações rodam em runners hospedados pelo GitHub (`windows-2025` e `ubuntu-24.04`). O workflow declara somente `contents: read` e não usa runner da rede da Majucau.
- Nenhum teste foi executado localmente nesta revisão. A CI executou a suíte configurada para as alterações de workflow.

## Artefato e limpeza

A CI gerou o artefato unsigned `majucau-g1-unsigned` (ID `11692559421`,
55.601.214 bytes). Ele expira automaticamente em `2026-10-12 05:44:17 UTC`.
O registro local `artifacts/retained/cleanup-register.md` contém o ID e o prazo,
pois as ferramentas GitHub disponíveis neste ambiente permitem listar e baixar,
mas não excluir o artefato.

## Limites e pendências

- A configuração de proteção da branch e a exigência de status checks antes do
  merge não foram confirmadas; continuam no escopo de BK-025.
- Não houve alteração em telas aprovadas ou regras financeiras.
