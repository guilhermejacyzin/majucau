# Cobertura das migrations no pacote NSIS — 2026-10-12

## Problema identificado

O manifesto de release lista todas as migrations SQL em `artifacts/windows/migrations`, mas o passo de verificação do instalador copiava para a pasta temporária apenas `000001_init.up.sql`. Como o verificador confere todos os caminhos e hashes declarados no manifesto, um pacote com migrations posteriores seria rejeitado antes da instalação.

## Correção

O instalador agora copia todos os arquivos `*.sql` da pasta de migrations para a área temporária, preservando os nomes. Assim, o conteúdo entregue ao `verify-package` cobre o inventário completo do manifesto.

Essa correção só afeta a composição e a conferência dos arquivos do pacote. Não executa migrations, não altera o banco, não muda telas e não afeta regras financeiras.

## Limites

A comparação inicial foi estática; não executei testes ou build localmente. A CI `38337400706` passou nos quatro grupos. A etapa Windows compilou o NSIS e passou pelo smoke de instalação e remoção do instalador; também passou a compilação do aplicativo e os verificadores existentes. A verificação com PostgreSQL e o backlog também passaram.

Isso confirma a montagem do pacote e seu smoke automatizado, mas não substitui a instalação em uma máquina limpa Windows 10/11 com PostgreSQL dedicado. O artefato unsigned dessa CI foi mantido temporariamente para validação manual e está registrado no arquivo local `artifacts/retained/cleanup-register.md`.
