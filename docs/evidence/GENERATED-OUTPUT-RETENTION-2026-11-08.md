# Registro de outputs preservados — 08/11/2026

## Pasta de build do frontend

- Caminho dentro do projeto: `frontend/dist/`.
- Conteúdo: arquivos que o Wails gerou ao compilar o frontend e incorporou ao
  executável de revisão.
- Situação: permanece no checkout como pasta gerada. A tentativa de remover a
  pasta recursivamente foi bloqueada pela proteção automática do ambiente; não
  tentei contornar o bloqueio.
- Revisão manual, se desejada: confirmar que `frontend/dist/` não será usado
  para abrir o app de desenvolvimento e então apagar a pasta. Uma nova
  compilação Wails consegue recriá-la.

O executável atualizado está em
`artifacts/verification/majucau-t02-review-final.exe`. Ele não foi incluído no
Git; os arquivos de código, testes e documentação foram preparados para o
commit.
