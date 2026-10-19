# Validação estrutural em fluxo das fontes do Drive — 2026-10-19

## O que foi corrigido

O worker examinava estruturalmente apenas o caminho de `02_nuvem_pago/recebimentos_futuros`. As outras duas raízes aprovadas entravam no inventário de metadados, mas seus arquivos não passavam pelo mesmo staging privado e pelos leitores seguros.

Agora `01_perdas_operacionais` e `03_nuvem_envio` também são verificadas, arquivo por arquivo, usando páginas limitadas do inventário e leitores em fluxo de CSV, TSV, JSON, XLS e XLSX. O conteúdo lido é descartado após a validação estrutural; a alteração não interpreta valores, não cria fatos financeiros e não muda regras ou telas.

O resultado diário de cada fonte é salvo no PostgreSQL por pasta e horário da execução. O registro contém somente contagens, estado, código técnico sanitizado e referência ao lote de inventário; não contém nomes, IDs nem conteúdo dos arquivos. Repetir a mesma execução atualiza o mesmo registro, e a referência ao lote é conferida contra a pasta e a conexão do Drive. O worker recebeu apenas as permissões de leitura, inclusão e atualização dessa nova tabela.

O leitor de CSV reconhece separador por vírgula ou ponto e vírgula olhando apenas o primeiro registro lógico, com tamanho máximo; o limite de colunas também é aplicado a CSV separado por vírgula. Os leitores mantêm limites de tamanho, campo/registro, linhas, colunas, planilhas e dados descompactados existentes. Temporários ficam privados e são removidos; falha de remoção continua sendo reportada.

## Situação e limites

- A validação estrutural não substitui os contratos aprovados para perdas e frete. As duas fontes continuam `PENDING` e não produzem valores para a DRE, CAPEX ou outros módulos.
- Arquivo ausente, atrasado, inválido, sem formato aceito ou que exceda limites permanece pendente; a origem não é alterada.
- Não foi feita sincronização com arquivos reais nesta execução. A sessão humana `financeiro@majucau.com.br` não equivale à identidade técnica de leitura ainda necessária ao worker.
- Nenhuma tela aprovada ou regra financeira foi alterada.

## Verificações

`scripts/verify.ps1` passou localmente, incluindo testes Go, revisão de fronteiras, `go vet`, `govulncheck`, verificação de checksums de migrações, geração SQLC determinística, scan de secrets, checks do instalador, lint, typecheck, 33 testes de interface, build e `npm audit` sem vulnerabilidades. O `govulncheck` encontrou zero vulnerabilidades alcançáveis pelo código; sinalizou uma vulnerabilidade em módulo requerido sem chamada alcançável.

O teste PostgreSQL `TestRecordSourceValidationIsIdempotentAndBounded` foi adicionado para conferir gravação, repetição idempotente, contagens e vínculo correto do lote. Não foi executado localmente porque `MAJUCAU_TEST_DATABASE_URL` não está configurado; é necessário confirmar sua execução na CI PostgreSQL após o push. Portanto, a aprovação local não é apresentada como homologação de banco ou de arquivos reais.

## Próximo passo

Confirmar a CI PostgreSQL e Windows. Depois, provisionar a identidade técnica somente leitura e validar a coleta real. A interpretação dos campos e qualquer lançamento continuam bloqueados até os contratos de cada fonte estarem aprovados.
