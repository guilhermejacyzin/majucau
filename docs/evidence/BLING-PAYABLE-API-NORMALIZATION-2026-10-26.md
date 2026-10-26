# Normalização diária de Contas a Pagar do Bling — 26/10/2026

## Objetivo

Fazer a informação de Contas a Pagar que o Majucau já busca pela API chegar às
tabelas que alimentam a Visão Executiva e T4. A mudança cobre somente os dados
de obrigações e mantém as telas e regras financeiras aprovadas.

## O que foi implementado

- A carga diária completa consulta a lista de títulos e busca, por identificador,
  o detalhe da obrigação, o contato fornecedor e a categoria associada.
- Cada resposta continua preservada em RAW. Os campos tipados incluem valor
  original, saldo em aberto, vencimento, competência, documento, situação,
  fornecedor e categoria. Valores são lidos como decimais exatos; não se usa
  `float`.
- O contato tipado guarda somente nome e identidade de origem. Dados adicionais
  do contato permanecem apenas na evidência RAW protegida.
- Lineage liga a obrigação à sua resposta RAW e às respostas RAW de fornecedor e
  categoria. A tabela de associação registra quais obrigações pertencem à carga
  completa.
- Só uma carga diária completa com todas as páginas e referências válidas é
  marcada `NORMALIZED`. A Visão Executiva e T4 exibem obrigações da mesma carga
  completa; registros que não vieram nela não são misturados ao snapshot atual.
- O agendador também só considera a carga do dia concluída quando Contas a
  Pagar terminou em `NORMALIZED`; um lote antigo ou apenas `RAW_ONLY` é repetido
  com segurança em vez de ser pulado.
- A publicação de RAW, dados tipados, associações, cursor e estado de sucesso é
  atômica. Uma falha deixa o snapshot anterior disponível e audita a tentativa
  como falha.
- A nova função que escreve contatos valida a conexão Bling e a correspondência
  do registro RAW; a role do worker recebe somente a permissão necessária.

## Verificações feitas

- Testes unitários cobrem status desconhecido, ID divergente, campos
  obrigatórios, datas inválidas, precisão decimal e referências incoerentes.
- A suíte Go completa (`go test ./...`) e `go vet ./...` passaram localmente.
- `sqlc generate` repetido produziu os mesmos arquivos gerados; o teste de
  contas a pagar e o teste do agendador compilaram. Os dois cenários PostgreSQL
  foram pulados localmente porque a URL do banco dedicado não está configurada.
- Também passaram a verificação de limites entre pacotes Go, a conferência de
  checksums das migrations, o scan local de segredos e o validador do backlog.
- Teste PostgreSQL foi criado para conferir normalização, lineage, estado
  `NORMALIZED` e preservação do snapshot anterior em falha de categoria. Ele só
  executa no banco dedicado `majucau_test`.
- Ainda não há `MAJUCAU_TEST_DATABASE_URL` configurada nesta máquina; por isso o
  teste de banco precisa ser confirmado pela CI após o push.
- A conta Bling real ainda não foi homologada; esta mudança não declara os dados
  reais confirmados.

O contrato de endpoints utilizado foi conferido na [referência oficial da API
Bling](https://developer.bling.com.br/referencia). A resposta autenticada da
conta Majucau ainda precisa validar a compatibilidade prática dos campos.

## Limites atuais

Contas a Receber, recebimentos/pagamentos realizados, movimentos de Caixa,
conciliação, CAPEX e as telas operacionais restantes ainda não possuem a mesma
normalização da API. Importação manual/filtrada de Contas a Pagar permanece
`RAW_ONLY`. A implementação não calcula DRE ou CAPEX e não altera apresentação,
design ou regra financeira.

## Regra de CAPEX reafirmada

O início do histórico do card `CAPEX ACUMULADO` é `01/01/2026`, conforme
aprovação da Gisele. A regra permanece informativa e separada do lucro da DRE;
esta integração de Contas a Pagar não fornece sozinha o vínculo necessário para
calcular o CAPEX pago ou reconhecido.
