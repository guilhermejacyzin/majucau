# Cobertura das migrations no pacote NSIS — 2026-10-12

## Problema identificado

O manifesto de release lista todas as migrations SQL em `artifacts/windows/migrations`, mas o passo de verificação do instalador copiava para a pasta temporária apenas `000001_init.up.sql`. Como o verificador confere todos os caminhos e hashes declarados no manifesto, um pacote com migrations posteriores seria rejeitado antes da instalação.

## Correção

O instalador agora copia todos os arquivos `*.sql` da pasta de migrations para a área temporária, preservando os nomes. Assim, o conteúdo entregue ao `verify-package` cobre o inventário completo do manifesto.

Essa correção só afeta a composição e a conferência dos arquivos do pacote. Não executa migrations, não altera o banco, não muda telas e não afeta regras financeiras.

## Limites

A conferência foi estática: comparei o padrão de arquivos gerado pelo manifesto com o arquivo que o NSIS copiava antes. Não executei testes ou build localmente. A CI acionada pelo push deve confirmar a compilação e os verificadores existentes; a instalação em uma máquina limpa continua pendente.
