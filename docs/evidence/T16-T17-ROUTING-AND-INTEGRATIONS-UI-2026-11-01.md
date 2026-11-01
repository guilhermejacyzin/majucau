# Correção da navegação T16/T17 e da tela de Integrações — 2026-11-01

## O que foi corrigido

- Os itens **Parâmetros (T16)** e **Usuários (T17)** apontavam para a página de Integrações. Cada item agora abre sua própria tela.
- Os conteúdos de T16 e T17 foram comparados com as imagens de referência preservadas em `sources/private/material-original/`. A tela mantém a estrutura dos cards, tabela, distribuição e regras importantes; números de usuários e resultados sem fonte aparecem como indisponíveis, sem dados fictícios.
- T17 reflete a decisão vigente da versão 1: apenas os perfis **ADMIN** e **DIRETORES**. Nomes, contas, quantidades e acessos individuais não são presumidos.
- A tela de Integrações deixou de solicitar App ID, segredo e redirect da Nuvemshop e informa que os dados correspondentes vêm de arquivo externo aprovado ou do Bling, conforme a fonte de cada informação.
- O estado exibido após o fluxo OAuth do Bling agora é atualizado ao concluir ou cancelar. “Testar conexão” permanece desabilitado enquanto o Bling não estiver conectado.

## Proteções

- Nenhuma fórmula financeira ou regra de negócio foi alterada.
- Nenhum nome, e-mail, total, saldo ou atividade real foi inventado.
- T16 e T17 não foram marcadas como concluídas no backlog. Esta fatia corrige navegação e estados de consulta; autenticação, autorização no worker, dados reais e aceites funcionais continuam pendentes.
- A sincronização real do Bling continua pendente: a consulta de categoria recebeu HTTP 403 por falta de escopo. O acesso automatizado ao cadastro do aplicativo permanece bloqueado, e nenhuma permissão foi alterada nesta execução.

## Verificações executadas

`scripts/verify.ps1` terminou com sucesso em 2026-11-01:

- testes Go, `go vet` e limites de arquitetura;
- `govulncheck`: nenhuma vulnerabilidade alcançável no código; encontrou uma vulnerabilidade em módulo requerido sem chamada alcançável;
- checksums de migrations, geração determinística do `sqlc`, scan local de segredos e checagens estruturais do instalador;
- frontend: lint, typecheck, **60 testes em 6 arquivos**, build e `npm audit` (zero vulnerabilidades reportadas pelo audit).

Esta execução não validou a conta Bling real, nem rodou a matriz de instalação em VM limpa. A confirmação manual visual da janela Wails também não foi feita; a validação desta fatia foi por comparação com as imagens de referência e pelos testes automatizados.

A CI do commit `cbca622` terminou com sucesso em todos os quatro jobs: `detect-changes`, `verify-backlog`, `verify-windows` e `verify-postgres` ([execução #41348424281](https://github.com/guilhermejacyzin/majucau/actions/runs/41348424281)). Isso confirma compilação e verificações automatizadas no Windows e PostgreSQL, mas não substitui a homologação com a conta real do Bling nem o teste visual manual.

## Limpeza

Os bundles temporários de verificação foram removidos com `scripts/SafeCleanup.psm1`. Foram preservados os caches centralizados necessários ao desenvolvimento e os artefatos oficiais do backlog em `artifacts/verification/backlog-v2/`.
