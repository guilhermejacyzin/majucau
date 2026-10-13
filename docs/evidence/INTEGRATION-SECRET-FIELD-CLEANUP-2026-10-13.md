# Limpeza do Client Secret após salvar — 13/10/2026

## Problema encontrado

O campo do Client Secret era enviado diretamente ao worker, sem guardá-lo no estado React, mas continuava preenchido no formulário depois da resposta. Isso deixava o segredo visível à sessão local até a tela ser fechada ou o campo ser substituído.

## Correção

- O formulário agora apaga o campo do Client Secret do Bling ou da Nuvemshop ao terminar cada tentativa de gravação, inclusive quando o worker rejeita a configuração ou a chamada falha.
- Client ID e Redirect URI permanecem preenchidos para facilitar a correção da configuração.
- O desenho, a posição dos campos, os textos e as regras financeiras não foram alterados.

## Verificação

- `npm test -- --config ../artifacts/cache/<execução>/vitest.config.ts src/App.test.tsx`: 18 testes passaram, incluindo Bling e Nuvemshop com sucesso e rejeição; o segredo foi apagado e os campos públicos permaneceram.
- `npm run typecheck`: passou.
- `npm run lint`: passou.
- Caches e arquivos temporários dos comandos foram isolados em `artifacts/cache/` e removidos ao final. Os caches compartilhados existentes não foram tocados.

O teste usa adaptadores locais e não transmite credenciais reais. Nenhum segredo foi lido, salvo ou registrado.
