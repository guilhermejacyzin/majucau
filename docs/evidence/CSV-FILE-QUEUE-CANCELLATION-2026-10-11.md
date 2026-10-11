# Evidência — cancelamento da fila de arquivos CSV

**Data:** 11/10/2026  
**Área:** `internal/integrations/csvstream`  
**Rastreabilidade:** BK-024, DATA-04.

## Mudança

`FileQueue.Next` agora confere o cancelamento antes de entregar cada nome de arquivo, mesmo quando ele já está no bloco carregado em memória. Assim, cancelar uma importação interrompe a fila imediatamente e evita que o chamador receba mais nomes pendentes.

## Cobertura

- Teste de integração PostgreSQL percorre 515 arquivos em mais de dois blocos, confirma ordenação determinística e verifica que arquivos não CSV e diretórios são ignorados.
- Outro teste cancela a operação depois de ler um nome e confirma que a fila retorna `context.Canceled` antes de entregar o próximo nome já carregado.
- Os testes usam transação e tabelas temporárias; o rollback encerra os dados de teste.

## Verificação

`go test -count=1 ./internal/integrations/csvstream` passou localmente no Windows. Como `MAJUCAU_TEST_DATABASE_URL` não está configurada nesta máquina, os dois testes PostgreSQL foram pulados localmente; o job `verify-postgres` da CI fornece um PostgreSQL descartável e executa os testes de integração.

Os caches e diretórios temporários do Go foram direcionados a `artifacts/cache` e removidos após a verificação.

## Limites

BK-024 continua em 35%: esta cobertura não substitui a validação de ACL sob o serviço, a instalação em VM limpa, falhas reais de disco/rede ou a homologação de acesso por perfil. Nenhuma tela, regra financeira, fórmula ou dado persistido foi alterado.
