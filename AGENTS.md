# AGENTS.md — Majucau Financial Intelligence

Estas regras valem para todo o futuro repositório. Instruções mais específicas podem existir em subdiretórios, mas não podem enfraquecer invariantes financeiros, segurança ou gates deste arquivo sem aprovação registrada.

## Produto e plataforma

- Aplicativo desktop single-user para Windows 10/11 x64.
- Frontend React + TypeScript + Vite dentro de Wails v2 estável.
- Backend, worker e integrações em Go.
- PostgreSQL local dedicado; acesso por `pgx/v5` e SQL gerado por `sqlc`.
- `majucau.exe` é a UI; `majucau-worker.exe` é o serviço Windows responsável por sincronização, cálculo, migrations, credenciais, backup e jobs.
- Empacotamento NSIS nativo, sem Docker e sem runtime Node em produção.

## Gates

- Não implementar código de produção antes da aprovação do G0.
- Não marcar G2 sem OAuth, RAW, idempotência e contratos reais validados.
- Não marcar G3/G5 sem regras financeiras aprovadas e casos dourados.
- Não marcar G7 sem instalador assinado, testes em VMs limpas, backup/restore real e aceite funcional.
- Um P0 diferido mantém seus módulos e resultados dependentes desabilitados; não converter ausência de fonte em estimativa.

## Invariantes financeiros protegidos

- B2C futuro vem exclusivamente de fonte Nuvem/Nuvem Pago oficial.
- B2B futuro vem exclusivamente do Bling após classificação aprovada.
- Nunca somar B2C Nuvem + B2C Bling.
- Total a Receber = B2C Nuvem + B2B Bling.
- Recebimentos e pagamentos realizados usam a fonte oficial Bling.
- Saldo Inicial D = Saldo Final Conciliado D-1.
- Saldo Final = Saldo Inicial + Entradas - Saídas +/- Ajustes Autorizados.
- D0 é provisório/em andamento; D+1 a D+60 é projetado.
- Cada saldo final alimenta o saldo inicial do dia seguinte.
- DRE, P&L, EBITDA, forecast, capital de giro e aplicação são calculados pelo motor Majucau; demonstrativos prontos do Bling não são verdade contábil.
- Débitos = créditos; Ativo = Passivo + Patrimônio Líquido. Diferença diferente de zero é `DIVERGENT`.
- Nenhuma dessas regras pode ser alterada sem nova versão, casos de regressão e aprovação funcional registrada.

## Dados, idempotência e lineage

- Valores monetários usam decimal exato; nunca `float`.
- Chave normalizada externa: `(connection_id, source_system, source_entity, source_id)`.
- Versão RAW: acrescentar `payload_hash`; sincronização nunca sobrescreve payload anterior.
- Redação LGPD é a única exceção de mutação RAW e exige tombstone, hash anterior, motivo, ator, timestamp e auditoria.
- Cursor só avança após commit integral do lote.
- Toda transformação material deve permitir: valor/card -> regra -> snapshot/linha -> contribuição tipada -> entidade -> RAW -> fonte original.
- FKs polimórficas não são fonte de integridade; use tabelas de contribuição tipadas.
- Snapshots, forecasts e regras aprovadas são imutáveis; correção cria nova versão.

## Integrações

- APIs externas são somente leitura no MVP.
- Frontend nunca chama Bling/Nuvemshop/Nuvem Pago nem PostgreSQL diretamente.
- Aplicar timeout, cancelamento, paginação, rate limit, backoff com jitter, retry apenas transitório e checkpoint transacional.
- Preservar último dado válido em falha; nunca substituir falha por zero.
- Nuvem Pago permanece `UNAVAILABLE`/`PARTIALLY_AVAILABLE` até existir fonte oficial de taxa, líquido, agenda e data prevista.
- OAuth abre navegador externo. Não capturar login em WebView, não usar clipboard para token e não embutir client secret.
- Contract tests usam OpenAPI versionado e fixtures sanitizadas; credenciais e payloads pessoais reais não entram no repositório.

## Segurança

- Secrets e tokens nunca no frontend, banco em texto aberto, Git, prompt, log, crash dump, linha de comando ou backup comum.
- Worker protege secrets com DPAPI e ACL/SDDL da conta `NT SERVICE\MajucauWorker`.
- IPC via Named Pipe autenticado pelo SID, DACL mínima, protocolo versionado e autorização no worker.
- Toda ação material verifica permissão no worker e gera `audit_events` sanitizado.
- Não logar documento, folha, token ou payload completo sem necessidade legal e proteção explícita.
- Rodar scan de secrets, dependências, vulnerabilidades, SBOM e licenças no pipeline.

## Frontend e acessibilidade

- Todo componente funcional declara `helpText` ou justificativa de não aplicabilidade.
- Tooltip contém somente texto curto, abre por hover/foco, fecha por Escape/blur e usa `aria-describedby`/`role=tooltip`.
- Links e ações pertencem a popover/dialog acessível, nunca ao tooltip.
- Labels visíveis continuam obrigatórios.
- Informação crítica deve permanecer visível fora do tooltip.
- Testar teclado, foco, leitor de tela, axe, contraste e zoom de 200%.
- Usar nomenclaturas oficiais do handoff; não criar card principal genérico chamado “Caixa”.

## Código e estrutura

- Separar domínio, casos de uso, adapters, infraestrutura e apresentação.
- `internal/domain` não importa UI, banco ou clientes HTTP.
- `internal/integrations` traduz payload externo para contratos do domínio, sem espalhar schemas de provedor.
- `internal/financial` e `internal/accounting` usam regras versionadas e funções determinísticas.
- SQL é explícito, versionado e revisável; migrations são forward-only, checksumadas e executadas sob advisory lock.
- Erros públicos usam códigos estáveis e mensagem simples; detalhes técnicos são sanitizados.
- Não adicionar dependência sem justificar maturidade, licença, CVEs, footprint e necessidade.

## Definition of Done por funcionalidade

Uma funcionalidade só está pronta quando possui:

1. regra e fonte aprovadas;
2. código e migrations necessários;
3. testes unitários, integração/contrato e casos negativos pertinentes;
4. idempotência, erros e concorrência testados quando aplicáveis;
5. lineage e auditoria demonstráveis;
6. UI com estados loading/empty/error/stale/partial/confirmed e tooltips acessíveis;
7. logs sanitizados e nenhuma exposição de secrets/PII;
8. critério financeiro executado contra oráculo com tolerância documentada;
9. documentação operacional e rollback/recuperação quando houver impacto;
10. evidência vinculada ao gate correspondente.

## Comandos mínimos de validação

```text
go test . ./cmd/... ./database/... ./internal/...
go vet . ./cmd/... ./database/... ./internal/...
govulncheck . ./cmd/... ./database/... ./internal/...
sqlc generate + verificação de diff
lint/checksum de migrations
eslint
tsc --noEmit
vite build
testes frontend + axe
scan de secrets
SBOM/licenças
installer smoke test
signtool verify
teste real de backup/restore
```

Não declarar um comando como aprovado sem executá-lo e preservar a evidência da execução.
