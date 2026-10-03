# AGENTS.md — Majucau Financial Intelligence

Estas regras valem para todo o futuro repositório. Instruções mais específicas podem existir em subdiretórios, mas não podem enfraquecer invariantes financeiros, segurança ou gates deste arquivo sem aprovação registrada.

## Produto e plataforma

- A interface aprovada é React + TypeScript + Vite dentro do aplicativo Wails v2 para Windows 10/11 x64.
- A implementação atual usa `majucau.exe`, worker local e PostgreSQL dedicado local; preservar esse modo até o gate de migração.
- A direção-alvo aprovada é hospedagem privada centralizada conforme `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md`.
- No alvo centralizado, serviço Go autenticado acessa PostgreSQL privado; clientes não recebem credencial e não conectam ao banco diretamente.
- Provedor, região, identidade, canal dos CSVs, modelo de acesso e segredos no host continuam pendentes; não os escolher por suposição.
- Empacotamento NSIS nativo continua necessário para o cliente Wails; Docker ou runtime Node não serão exigidos na máquina do usuário.

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
- Toda sincronização deve ser idempotente: replay do mesmo evento/página não duplica fatos nem versões RAW sem mudança.
- RAW, dados normalizados, cursor e sucesso do lote são confirmados atomicamente; falha ou cancelamento não avança cursor nem substitui último dado válido.
- Em importação de arquivo com erros de validação por linha, persistir as linhas válidas e fechar o lote como `PARTIAL`; contar em `records_failed` as linhas rejeitadas. Persistência e auditoria ficam na mesma transação, que reverte por inteiro em falha técnica ou cancelamento.
- Toda transformação material deve permitir: valor/card -> regra -> snapshot/linha -> contribuição tipada -> entidade -> RAW -> fonte original.
- FKs polimórficas não são fonte de integridade; use tabelas de contribuição tipadas.
- Snapshots, forecasts e regras aprovadas são imutáveis; correção cria nova versão.

## Streaming e limites de dados

- Ingestão, transformação, consulta e exportação de dados que podem crescer com o histórico devem fluir registro a registro ou em blocos limitados; não carregar arquivo, resposta HTTP, dump, arquivo compactado ou conjunto de linhas inteiro na memória.
- Para envelopes de controle pequenos (por exemplo, estado do instalador), usar decoder em stream e um limite explícito de bytes. Todo caminho precisa limitar tamanho de registro/campo, quantidade de entradas e conteúdo descompactado; exceder o limite deve falhar sem gravar dados parciais.
- Propagar cancelamento e timeout até a leitura e a persistência. Calcular hash enquanto o conteúdo passa pelo stream; validar cada registro antes de persistir e preservar atomicidade/idempotência.
- Entrada externa é não confiável. Não registrar payload, segredos ou PII; arquivos temporários com dados sensíveis ficam em diretório privado, têm limpeza garantida e não entram em caches ou outputs de CI.

## Governança de execução

- `IMPLEMENTATION-PLAN.md` é o plano-raiz; `REQUIREMENTS-TRACEABILITY.md` registra estado e evidência; `G0-DECISION-REGISTER.md` e decisões aprovadas são autoridade para regras de negócio. `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md` registra a direção-alvo aprovada; decisões operacionais pendentes não devem ser presumidas. Artefatos de planejamento v2 conflitantes não substituem esses documentos sem aprovação.
- Telas aprovadas e regras de negócio são protegidas. Dúvida que possa alterar comportamento, apresentação, importação ou contrato persistente exige pergunta antes da mudança.
- Commits pequenos devem informar `O que é`, `O que foi feito`, `Próximo passo`, percentual restante da tarefa e percentual restante do projeto. Cada execução reporta em linguagem simples progresso, pendências e estimativa.
- Caches e outputs gerados ficam em `artifacts/`. Ao fim de cada execução, remover o que puder ser recriado sem risco nem dependência ativa. O que precisar permanecer fica em `artifacts/retained/` e no registro local de limpeza, com motivo, origem e caminho para exclusão manual.

## Integrações

- APIs externas são somente leitura no MVP.
- Frontend nunca chama Bling/Nuvemshop/Nuvem Pago nem PostgreSQL diretamente.
- Retry limitado é o padrão nas integrações HTTP repetíveis e idempotentes: até 3 tentativas no total, para falhas transitórias de transporte, HTTP 408, 429 e 5xx; usar backoff exponencial com jitter e respeitar `Retry-After` até 30 segundos. Não repetir cancelamento do chamador, falha de autenticação/autorização, erro de validação/schema ou resposta inválida. Operação que altera estado externo só pode ser repetida quando o contrato oficial garantir replay seguro ou oferecer chave de idempotência; chamadas OAuth de uso único não herdam retry automaticamente.
- Cada adapter deve testar a classificação de retry, o limite, `Retry-After` e a não repetição de cancelamento/erros permanentes. Aplicar timeout, cancelamento, paginação, rate limit e checkpoint transacional; cursor só avança depois do commit atômico do lote.
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

## Instalação em máquinas de terceiros

- Tratar instalação, repair, upgrade, rollback e uninstall como operações transacionais, idempotentes e retomáveis.
- Antes de mutar a máquina, executar preflight de Windows/arquitetura, UAC, reboot pendente, espaço, paths, portas, WebView2, serviços e instalação anterior.
- Nunca reutilizar nem alterar PostgreSQL, serviço, porta ou diretório de outro produto.
- Falha parcial não pode deixar serviço iniciado contra schema incompleto, credencial exposta, porta liberada sem registro ou dados silenciosamente removidos.
- Preservar dados por padrão; exclusão exige escolha explícita, confirmação reforçada e recomendação de backup.
- Upgrade exige pacote assinado, backup prévio, compatibilidade binário/schema e recuperação conjunta quando rollback isolado for inseguro.
- Todo erro de instalação usa código estável, mensagem simples, ação recomendada e log técnico sanitizado; nunca registrar senha, token, PII ou argumento sensível.
- Suportar diagnóstico offline e support bundle sanitizado. Problema desconhecido deve falhar com segurança e produzir evidência acionável.
- G7 exige execução da matriz de instalação em VMs limpas e cenários adversos; documentação ou mocks não substituem esse teste.

## Frontend e acessibilidade

- Todo componente funcional declara `helpText` ou justificativa de não aplicabilidade.
- Tooltip contém somente texto curto, abre por hover/foco, fecha por Escape/blur e usa `aria-describedby`/`role=tooltip`.
- Links e ações pertencem a popover/dialog acessível, nunca ao tooltip.
- Labels visíveis continuam obrigatórios.
- Informação crítica deve permanecer visível fora do tooltip.
- O botão global `Ajuda` não faz parte do contrato visual vigente; não renderizar esse botão em nenhuma tela. A orientação contextual permanece somente por tooltips simples e acessíveis.
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
