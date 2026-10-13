# Reconciliação dos registros de progresso — 13/10/2026

## Objetivo

Manter o relatório detalhado, o JSON usado pelo projeto e a planilha oficial com a mesma situação das tarefas do backlog.

## O que foi feito

- Conferidos os registros das 127 tarefas e comparados com o detalhamento vigente.
- Atualizados no JSON sete registros cujos percentuais ou status estavam desatualizados: BK-017 (95%), BK-026 (60%), BK-027 (30%), BK-041 (75%), BK-042 (40%), BK-043 (40%) e BK-067 (25%). Os registros se baseiam nas evidências técnicas já existentes; esta ação não reestimou tarefas.
- Estendido `scripts/verify-backlog.ps1` para conferir quantidade, ordem, status e progresso entre o Markdown e o JSON.
- Estendido `scripts/test-verify-backlog.ps1` para provar que resumo incorreto e registro estruturado divergente são rejeitados.
- Acrescentado ao `scripts/SafeCleanup.psm1` um recurso para remover um link simbólico somente quando o caminho e o arquivo de destino esperados coincidem. O teste rejeitou um destino incorreto e confirmou que a remoção do link preserva o arquivo de destino.
- Recriada a planilha a partir do JSON depois de conferir visualmente as abas já existentes. A planilha mantém suas quatro abas e o desenho; apenas números, status e data de atualização refletem a fonte atual.

## Verificações

- `pwsh -NoProfile -File scripts/verify-backlog.ps1` — aprovado; 127 tarefas, 7 concluídas, 18 em andamento e 102 não iniciadas; 11% concluído e 89% restante.
- `pwsh -NoProfile -File scripts/test-verify-backlog.ps1` — aprovado; aceita registros consistentes e rejeita percentuais e registros estruturados divergentes.
- `pwsh -NoProfile -File scripts/test-safe-cleanup.ps1` — aprovado; a limpeza existente continua protegendo arquivos externos e o novo caso trata o link simbólico com caminho e destino exatos.
- `python work/verify_package.py` — aprovado: 127 tarefas, 108 decisões, 80 links locais válidos, zero erros nas fórmulas e média estimada de 11,0%. O manifesto confere os arquivos listados.
- A comparação visual das abas Evolucao e Tarefas confirmou as mesmas áreas, colunas e organização. Mudaram somente os valores reconciliados e a data do status.

## Limites

O percentual é a média estimada do catálogo e não mede prontidão para produção. A reconciliação não valida integração Bling, dados reais, telas, regras financeiras, instalação nem homologação.

Nenhuma tela aprovada ou regra financeira foi alterada.
