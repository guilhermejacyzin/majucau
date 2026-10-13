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

Esta atualização não busca dados novos no Bling. A conta continua dependendo de configuração, autorização, teste de conexão e comando manual de sincronização. A rotina automática diária às 09:30 em `America/Sao_Paulo` continua em BK-039. Gisele aprovou processar as fontes disponíveis independentemente, manter arquivos ausentes pendentes até às 09:30 do dia seguinte e reler diariamente todas as Contas a Receber e Contas a Pagar. A agenda diária substitui a proposta anterior de sincronização automática a cada 15 minutos; o limite `STALE` ainda precisa ser compatibilizado com a agenda. Ainda faltam implementar o agendamento, comprovar como termina a paginação e definir acesso persistente ao Drive pelo serviço.
