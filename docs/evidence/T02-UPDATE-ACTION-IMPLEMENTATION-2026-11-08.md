# Atualização manual completa da T02 — 08/11/2026

## Problema

O ícone do cabeçalho apenas relia o snapshot local. A usuária esperava uma nova
consulta ao Bling, portanto apertar o ícone não podia corrigir resultados
antigos.

## Correção

O clique agora percorre `frontend/src/mockupDashboard.tsx` → `frontend/src/App.tsx`
→ método Wails `App.SyncBling` → IPC `bling.refresh` → worker →
`BlingOAuthService.SyncNow`. Esse método usa o mesmo conjunto de fontes e o mesmo
bloqueio PostgreSQL da carga diária: recebíveis, pagamentos, movimentos de
Caixa e Bancos, movimentos conciliados, formas de pagamento e contas
financeiras. Cada carga manual recebe um identificador novo e espera a carga
agendada quando necessário; ela não é descartada porque a sincronização das
09:30 já ocorreu. O timeout da chamada manual é de até dez minutos.

O payload IPC não carrega filtros nem dados sensíveis. Os tokens continuam no
cofre protegido usado pelo worker. A interface indica andamento, evita cliques
simultâneos, relê o snapshot depois de uma resposta com fontes concluídas e
mostra erros sem declarar sucesso. Uma carga incompleta atualiza somente o que
concluiu e informa que há consultas pendentes. Nenhum texto de “Parcial” foi
adicionado aos valores financeiros. O desenho e as regras financeiras não
foram alterados.

## Verificações

Testes adicionados para: encaminhar o método IPC dedicado e payload vazio;
aceitar a operação no worker; rejeitar conteúdo inesperado; chamar o refresh
completo no botão, reler o snapshot ao concluir e informar a falha sem reler
dados.

**Resultado local — 08/11/2026:** `scripts/verify.ps1` concluiu com sucesso. Passaram os testes Go, `go vet`, verificação de vulnerabilidades alcançáveis, checksums de migrations, geração determinística de sqlc, busca local de segredos, validação do instalador, lint, TypeScript, os 75 testes de interface, build de produção e `npm audit` (0 vulnerabilidades). A varredura apontou uma vulnerabilidade em módulo requerido, mas não alcançável pelo código importado. `git diff --check` também passou.

## Limites

O código e testes automatizados não provam a execução da nova sincronização na
conta Bling real. A T02 permanece em andamento até a nova versão ser aberta,
Atualizar ser executado e os resultados da mesma semana e filtros serem
comparados com Contas a Receber (valor líquido), Contas a Pagar e o saldo
conciliado do Itaú. A responsável deve aprovar a T02 antes de iniciar T03.
