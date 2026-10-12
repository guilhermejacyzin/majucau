# Atualização do painel após Integrações — 2026-10-12

## Mudança

Quando a pessoa usuária sai de Integrações e retorna ao painel, o frontend consulta novamente o estado do worker e o snapshot persistido no PostgreSQL local. Isso permite que uma sincronização manual já concluída apareça no painel sem reiniciar o protótipo.

A mudança somente relê dados locais. Não inicia autorização OAuth, chamadas à API do Bling, sincronizações automáticas ou alteração de dados. Não muda desenho, textos, fórmulas ou regras financeiras.

## Verificações

- `frontend/src/App.test.tsx`: 16 testes passaram, incluindo a atualização do snapshot ao retornar de Integrações e o uso da data da sincronização Bling confirmada.
- `npm run typecheck`: passou.
- ESLint em `frontend/src/App.tsx` e `frontend/src/App.test.tsx`: passou sem avisos.
- `git diff --check`: passou.
- A checagem visual no protótipo em execução ficou pendente para evitar reiniciar a sessão ativa.

## Limites

Esta atualização não busca dados novos no Bling. A conta continua dependendo de configuração, autorização, teste de conexão e comando manual de sincronização. A atualização automática diária das 09:30 em `America/Sao_Paulo` continua em BK-039; falta definir o tratamento quando algum arquivo externo estiver ausente.
