# AGENTS.md — Majucau Financial Intelligence

Estas regras valem para todo o futuro repositório. Instruções mais específicas podem existir em subdiretórios, mas não podem enfraquecer invariantes financeiros, segurança ou gates deste arquivo sem aprovação registrada.

## Produto e plataforma

- A interface aprovada é React + TypeScript + Vite dentro do aplicativo Wails v2 para Windows 10/11 x64.
- A implementação atual usa `majucau.exe`, worker local e PostgreSQL dedicado local; preservar esse modo até o gate de migração.
- A direção-alvo aprovada é hospedagem privada centralizada conforme `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md`.
- No alvo centralizado, serviço Go autenticado acessa PostgreSQL privado; clientes não recebem credencial e não conectam ao banco diretamente.
- A atualização diária foi aprovada para 09:30 no fuso `America/Sao_Paulo` (horário de Brasília). O worker Windows relê todas as contas do Bling sem filtro de data e, em 18/10/2026, passou a persistir a identidade da carga diária, impedir concorrência entre instâncias por trava PostgreSQL e retomar a partir da última página confirmada em checkpoint após reinício. A paginação completa ainda não foi homologada com a conta real. Não limitar a consulta automática a alterações recentes nem aos próximos 60 dias. Na tela de sincronização manual, vencimento consulta Contas a Receber e Contas a Pagar, recebimento consulta somente Contas a Receber e pagamento consulta somente Contas a Pagar. Processar fontes externas disponíveis de forma independente; se um arquivo não estiver na pasta às 09:30, deixá-lo pendente até às 09:30 do dia seguinte. Se o worker iniciar depois das 09:30, recuperar a execução do dia com a mesma identidade idempotente; arquivo modificado depois do corte das 09:30 continua pendente até o próximo horário aprovado. Preservar o último dado válido e nunca apresentá-lo como atual quando uma fonte estiver pendente. Para listas Bling com paginação efetiva por `pagina`/`limite` (Contas a Receber/Pagar e categorias), página cheia continua e página curta ou vazia encerra; resposta acima do limite falha com segurança. `GET /caixas` envia `limite=100` conforme a orientação geral oficial do Bling, mas não presume que o endpoint o aplicou: página curta/vazia só é terminal se a resposta confirmar o limite ou trouxer metadados suficientes. Se o endpoint rejeitar o parâmetro ou não permitir provar o fim, preservar o retrato e cursor anteriores. Em endpoints sem limite efetivo nem metadados, página vazia permanece ambígua e a sincronização preserva o retrato e cursor anteriores. A homologação real ainda é obrigatória antes de liberar dados como confirmados. A publicação das páginas RAW, aposentadoria do retrato anterior, cursor e sucesso é atômica; falha preserva a versão anterior. A CI passou para o checkpoint por página no commit `951d4a6` em 18/10/2026. A autenticação Bling usa o host atual `https://api.bling.com.br/Api/v3` e `enable-jwt: 1` na troca/renovação de token e nas consultas autenticadas; isso não comprova autorização ou sincronização da conta real. A arquitetura de leitura do Drive foi aprovada em 2026-10-18: o serviço Go usa uma identidade técnica com permissão Leitor somente nas três raízes autorizadas. O caminho aprovado é `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT/01_perdas_operacionais`, `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT/02_nuvem_pago/recebimentos_futuros` e `Meu Drive/MAJUCAU_INTELLIGENCE_INPUT/03_nuvem_envio`; subpastas internas, incluindo `03_nuvem_envio/Frete`, são percorridas apenas quando descobertas na listagem da raiz autorizada. Os IDs reais foram identificados, mas ainda não configurados localmente. O inventário do worker percorre subpastas com checkpoints por pasta, retoma depois de falha e publica tudo somente após a árvore completa. O leitor de arquivos e o staging privado estão ligados ao worker; somente CSV está ligado ao importador financeiro de recebimentos futuros, e o CSV visto no Drive na revisão de 18/10 estava desatualizado e não foi importado. A identidade de serviço e credencial DPAPI ainda não foram provisionadas. XLS e XLSX têm leitores estruturais em fluxo aplicados às três raízes autorizadas do Drive, com staging privado e limites; essa validação não importa valores financeiros. Arquivos XLS/XLSX permanecem pendentes para importação financeira até existir mapeamento aprovado, contratos, amostras e homologação real. O Site permanece apenas no fluxo OAuth do Bling. Provedor e região da hospedagem, MFA, segredos no host e demais detalhes operacionais também continuam pendentes. Para identidade, Gisele escolheu login próprio do Majucau, informou `financeiro@majucau.com.br` para `ADMIN` e aprovou `contato@majucau.com.br` como conta compartilhada dos Diretores, ciente de que os registros não identificam qual diretor fez cada ação. Essa decisão, registrada no `G0-DECISION-REGISTER.md` em 2026-10-11, substitui a exigência de identidade individual da ADR-002 somente para o perfil `DIRECTOR`. MFA, sessão, recuperação e controles de autorização continuam pendentes; não provisionar antes de implementá-los e validá-los.
- Empacotamento NSIS nativo continua necessário para o cliente Wails; Docker ou runtime Node não serão exigidos na máquina do usuário.
- A usuária confirmou que a conta Google humana de referência para o Drive é `financeiro@majucau.com.br`. Essa conta do Chrome serve para consulta autorizada; o worker usa identidade técnica separada, com acesso somente de leitura. Nunca transportar senha, cookie ou sessão do navegador para o worker.

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
- Sucesso ao baixar uma fonte não significa que os fatos financeiros foram montados. Marque `integration_sync_batches.data_stage` como `RAW_ONLY` até a conversão e persistência normalizadas; somente um lote `NORMALIZED` do recurso certo pode confirmar zero ou frescor de métrica. Lotes antigos permanecem `UNVERIFIED`.
- Em importação de arquivo com erros de validação por linha, persistir as linhas válidas e fechar o lote como `PARTIAL`; contar em `records_failed` as linhas rejeitadas. Persistência e auditoria ficam na mesma transação, que reverte por inteiro em falha técnica ou cancelamento.
- Toda transformação material deve permitir: valor/card -> regra -> snapshot/linha -> contribuição tipada -> entidade -> RAW -> fonte original.
- FKs polimórficas não são fonte de integridade; use tabelas de contribuição tipadas.
- Snapshots, forecasts e regras aprovadas são imutáveis; correção cria nova versão.

## Streaming e limites de dados

- Ingestão, transformação, consulta e exportação de dados que podem crescer com o histórico devem fluir registro a registro ou em blocos limitados; não carregar arquivo, resposta HTTP, dump, arquivo compactado ou conjunto de linhas inteiro na memória.
- Para envelopes de controle pequenos (por exemplo, estado do instalador), usar decoder em stream e um limite explícito de bytes. Todo caminho precisa limitar tamanho de registro/campo, quantidade de entradas e conteúdo descompactado; exceder o limite deve falhar sem gravar dados parciais.
- Propagar cancelamento e timeout até a leitura e a persistência. Calcular hash enquanto o conteúdo passa pelo stream; validar cada registro antes de persistir e preservar atomicidade/idempotência.
- Entrada externa é não confiável. Não registrar payload, segredos ou PII; arquivos temporários com dados sensíveis ficam em diretório privado, têm limpeza garantida e não entram em caches ou outputs de CI. A falha de limpeza deve ser propagada como `securetemp.CleanupError`, com o caminho acessível para operador autorizado e omitido da mensagem comum/log. Não descartar erros de remoção de temporários; se um workspace contiver o arquivo, a remoção do workspace deve reportar qualquer resíduo.
- Scripts PowerShell que apagam diretórios recursivamente devem usar `scripts/SafeCleanup.psm1`, restringir o destino ao caminho relativo previsto dentro do projeto, rejeitar links/reparse points nos pais e no conteúdo e propagar qualquer falha de limpeza.
- Inventários de artefatos devem ser escritos linha a linha, sem ordenar ou acumular todos os caminhos em memória. Liste apenas entregáveis e evidências de verificação; não percorra caches nem conteúdo retido.

## Governança de execução

- `IMPLEMENTATION-PLAN.md` é o plano-raiz; `REQUIREMENTS-TRACEABILITY.md` registra estado e evidência; `G0-DECISION-REGISTER.md` e decisões aprovadas são autoridade para regras de negócio. `ADR-002-HOSPEDAGEM-PRIVADA-CENTRALIZADA.md` registra a direção-alvo aprovada; decisões operacionais pendentes não devem ser presumidas. Artefatos de planejamento v2 conflitantes não substituem esses documentos sem aprovação.
- Telas aprovadas e regras de negócio são protegidas. Dúvida que possa alterar comportamento, apresentação, importação ou contrato persistente exige pergunta antes da mudança.
- Commits pequenos devem informar `O que é`, `O que foi feito`, `Próximo passo`, percentual restante da tarefa e percentual restante do projeto. Cada execução reporta em linguagem simples progresso, pendências e estimativa.
- Caches e outputs gerados ficam em `artifacts/`: os relatórios oficiais versionados do planejamento em `artifacts/planejamento/outputs/`, caches em `artifacts/cache/` e saídas temporárias de conferência em `artifacts/verification/`. Ao fim de cada execução, remover o que puder ser recriado sem risco nem dependência ativa. O que precisar permanecer fica em `artifacts/retained/` e no registro local de limpeza, com motivo, origem e caminho para exclusão manual.
- Builds locais devem preservar `artifacts/windows` e `artifacts/MANIFEST.txt` anteriores até a compilação e o empacotamento terminarem. Se houver falha depois de publicar saídas novas, guardar as saídas parciais em `artifacts/retained/manual-cleanup/` e registrar os caminhos localmente antes de restaurar a última versão válida.
- Scripts que executam Go devem direcionar `GOCACHE`, `GOMODCACHE` e `GOTMPDIR` para subpastas de `artifacts/cache` e criar os diretórios necessários antes do comando. Não executar `go` manualmente sem esses caminhos definidos; isso evita deixar downloads ou temporários no perfil global do computador.

## Integrações

- APIs externas são somente leitura no MVP.
- Frontend nunca chama Bling/Nuvemshop/Nuvem Pago nem PostgreSQL diretamente.
- Retry limitado é o padrão nas integrações HTTP repetíveis e idempotentes e deve usar `internal/integrations/httpretry`: até 3 tentativas no total, para falhas transitórias de transporte, HTTP 408, 429 e 5xx; usar backoff exponencial com jitter e respeitar `Retry-After` até 30 segundos. No transporte, repetir apenas timeouts, EOF/EOF inesperado, DNS temporário e falhas de conexão como reset/recusa; certificado TLS inválido, DNS permanente e erro de configuração não são transitórios. Preservar a mensagem externa sem expor o erro bruto de rede. Não repetir cancelamento do chamador, falha de autenticação/autorização, erro de validação/schema ou resposta inválida. A implementação comum só aceita `GET`; operação que altera estado externo só pode ganhar suporte de retry separado quando o contrato oficial garantir replay seguro ou oferecer chave de idempotência. Chamadas OAuth de uso único não herdam retry automaticamente. O teste `TestProductionHTTPCallsUseReviewedPaths` verifica o inventário de chamadas HTTP de produção e exige revisão explícita para novos caminhos diretos.
- Cada adapter deve testar a classificação de retry, o limite, `Retry-After` e a não repetição de cancelamento/erros permanentes. Aplicar timeout, cancelamento, paginação, rate limit e checkpoint transacional; cursor só avança depois do commit atômico do lote.
- Preservar último dado válido em falha; nunca substituir falha por zero.
- Nuvem Pago permanece `UNAVAILABLE`/`PARTIALLY_AVAILABLE` até existir fonte oficial de taxa, líquido, agenda e data prevista.
- OAuth abre navegador externo. Não capturar login em WebView, não usar clipboard para token e não embutir client secret.
- Contract tests usam OpenAPI versionado e fixtures sanitizadas; credenciais e payloads pessoais reais não entram no repositório.

## Segurança

- Secrets e tokens nunca no frontend, banco em texto aberto, Git, prompt, log, crash dump, linha de comando ou backup comum.
- No protótipo em console, o worker mantém o DPAPI do usuário atual. Quando iniciado pelo Service Control Manager, usa DPAPI no escopo da máquina e exige ACL/SDDL restrita ao SID do serviço `NT SERVICE\MajucauWorker`, SYSTEM e Administradores; o escopo da máquina nunca substitui a ACL. O instalador prepara e valida as pastas do cofre e dos logs antes de iniciar o serviço, usa a conta virtual com SID restrito, grava o SID da interface em uma chave de registro com leitura para o serviço e administração somente para SYSTEM/Administradores, e preserva os dados ao remover o serviço. O smoke automatizado de instalação/remoção em Windows e a validação em VM limpa ainda precisam passar antes de declarar o ciclo homologado.
- Credencial técnica do Google Drive e IDs das três raízes aprovadas são provisionados juntos pelo worker, por Named Pipe local autenticado. Confirmar o SID Windows do cliente; registrar somente metadados sanitizados por `audit_google_drive_credential_configured`; nunca guardar chave ou IDs no Git, argumentos de processo, logs ou PostgreSQL.
- IPC via Named Pipe autenticado pelo SID, DACL mínima, protocolo versionado e autorização no worker.
- Toda ação material verifica permissão no worker e gera `audit_events` sanitizado.
- `majucau_runtime` usa uma allowlist explícita de leitura, escrita e execução; nunca conceder acesso amplo `ALL TABLES/FUNCTIONS` ao runtime. RAW e auditoria continuam sem mutação direta, e cada nova relação/função requer uma permissão deliberada coberta por `database/privileges/verify_runtime.sql`.
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
- `scripts/verify-go-boundaries.ps1` confere automaticamente as camadas Go na verificação local e na CI; desvios exigem revisão de arquitetura.
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
