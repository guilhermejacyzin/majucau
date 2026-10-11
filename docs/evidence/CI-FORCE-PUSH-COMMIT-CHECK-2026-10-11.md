# Recuperação da CI após atualização forçada — 2026-10-11

## Ocorrido

Após a reescrita autorizada das duas mensagens de commit, o GitHub Actions executou a CI `38165227608`. O job de validação do backlog passou. O job `detect-changes` encerrou com erro ao conferir mensagens porque o runner, em um checkout limpo, não tinha mais o commit anterior do push forçado (`BEFORE_SHA`). A seleção tentou executar `git rev-list BEFORE_SHA..HEAD`, que falhou antes de ler a lista de commits presente no payload do evento. Os jobs Windows e PostgreSQL foram ignorados por causa da falha de `detect-changes`.

## Correção

- `scripts/verify-commit-messages.py` agora tenta a comparação pelo SHA anterior e, se esse SHA não estiver disponível no checkout, usa os commits informados pelo payload do evento. Se o payload não estiver disponível, mantém a verificação do commit atual como último recurso.
- `scripts/test_verify_commit_messages.py` cobre a situação em que o SHA anterior não existe e confirma a leitura do payload, ignorando commits de merge.
- A CI executa esse teste junto da verificação de mensagens de commit.
- As mensagens dos commits `9d77684` e `caa6ef4` foram conferidas localmente com `scripts/verify-commit-messages.py`; ambos os campos de percentual e os três campos textuais obrigatórios passaram.

## Validação e limite

`python scripts/test_verify_commit_messages.py`, a simulação local do checker com um `BEFORE_SHA` inexistente e o payload de commits, e `git diff --check` passaram. A correção ainda precisa passar na CI acionada pelo próximo push. O run `38165227608` não executou testes de produto, pois parou no job de detecção/validação de mensagens.

Nenhuma tela, regra financeira, fórmula ou código de produto foi alterado.
