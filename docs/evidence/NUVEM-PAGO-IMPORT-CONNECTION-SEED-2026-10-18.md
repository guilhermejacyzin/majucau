# Registro técnico do Nuvem Pago para importação de arquivos — 18/10/2026

## Problema

O importador aprovado de recebimentos futuros procurava um registro `NUVEM_PAGO` em `integration_connections` para obter o identificador local do lote. As migrações criavam o registro do Google Drive, mas não o do Nuvem Pago. Em uma instalação nova, a importação de um CSV válido terminaria com `NUVEM_PAGO_CONNECTION_MISSING` antes de processar as linhas.

## Correção

- A migration `000008_nuvem_pago_import_connection.up.sql` cria esse registro com estado `UNAVAILABLE`, preservando o aviso existente de que a API do Nuvem Pago não fornece confirmação financeira.
- `ON CONFLICT DO NOTHING` preserva qualquer registro já existente e não altera credenciais nem estado de uma instalação atualizada.
- O teste PostgreSQL da importação verifica o estado inicial vindo das migrations e importa a planilha de teste sem marcar o conector de API como conectado.
- Não houve alteração de colunas, valores calculados, classificação financeira, telas, nem dados reais.

## Validação e limite

- `scripts/verify-migrations.ps1` e a suíte focalizada Go validam a migration e o teste localmente.
- A verificação ponta a ponta exige PostgreSQL descartável com todas as migrations; a CI PostgreSQL aplica esse ambiente e executa o teste. Nenhum banco local com dados financeiros foi usado.
- BK-042 e BK-043 permanecem nas porcentagens atuais: o staging do Drive ainda não está ligado aos importadores e a importação real continua pendente de credenciais, contrato e homologação.
