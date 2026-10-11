# Como contribuir com o Majucau

Este guia ajuda a fazer mudanças pequenas e seguras. Antes de começar, consulte
`AGENTS.md`, `G0-DECISION-REGISTER.md` e os documentos da tarefa.

## Proteções obrigatórias

- Telas aprovadas e regras de negócio/financeiras não podem ser alteradas sem
  aprovação explícita registrada. Se uma tarefa ou arquivo deixar dúvida sobre
  apresentação, cálculo, fonte de dados ou comportamento, pare e pergunte.
- Use as decisões aprovadas como fonte de verdade. Um mockup, planilha, exemplo
  ou texto antigo não substitui uma decisão posterior registrada.
- Dados externos são não confiáveis. Não coloque tokens, credenciais, dados
  pessoais ou payloads reais no Git, em logs ou em evidências.
- Processamento de histórico deve usar fluxo ou blocos limitados, com
  cancelamento, limites explícitos e limpeza segura de temporários.
- Persistência de lotes deve respeitar atomicidade, idempotência e lineage.
  Só repita chamadas HTTP quando forem seguras para replay, seguindo o retry
  limitado descrito em `AGENTS.md`.
- Caches e saídas geradas ficam em `artifacts/`. Remova saídas recriáveis ao
  terminar; preserve caches necessários e registre resíduos que precisem de
  remoção manual no registro local de limpeza.

## Antes de abrir uma mudança

1. Confirme o ID e o critério de aceite no backlog e confira as dependências.
2. Leia as decisões e evidências relacionadas à tarefa.
3. Mantenha o diff limitado à tarefa; não misture refatoração sem relação.
4. Se uma entrada necessária estiver ausente, represente-a como indisponível;
   não invente valor ou regra.

## Verificação e evidências

- Para mudanças de código, informe quais verificações foram executadas e seus
  resultados. Se não forem executadas, registre isso claramente; não declare
  aprovação sem evidência.
- Não use testes sintéticos para afirmar que um contrato externo real foi
  homologado. Fixtures devem ser sintéticas ou sanitizadas.
- Registre limites, riscos e pendências junto à evidência da tarefa, sem incluir
  credenciais ou dados pessoais.

## Commits e publicação

- Faça um commit pequeno por tarefa ou unidade coerente. Não inclua alterações
  locais de outra pessoa.
- Use um título curto no formato `tipo(área): resumo`, por exemplo
  `docs(governanca): registrar convencoes de contribuicao`.
- O corpo do commit deve informar:

  ```text
  O que é: tarefa que está subindo
  O que foi feito: resumo técnico detalhado
  Próximo passo: detalhe do próximo passo
  Quanto falta da tarefa: percentual
  Quanto falta do projeto: percentual
  ```

- Publique somente na branch de trabalho autorizada. Não faça force push,
  merge em `main` ou publicação em produção como parte de uma mudança comum.
- A CI existente roda em `push` e `pull_request`; uma mudança documental pode
  não executar as verificações de produto. A configuração de proteção de
  branches e de revisores obrigatórios do GitHub precisa ser conferida nas
  configurações do repositório; este arquivo não afirma que ela já está ativa.

