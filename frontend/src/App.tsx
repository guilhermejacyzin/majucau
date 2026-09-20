import { FormEvent, useEffect, useMemo, useRef, useState } from 'react'
import logo from './assets/images/logo-universal.png'
import { Tooltip } from './components/Tooltip'
import { type TooltipId } from './tooltips'
import { createBlingConfigAdapter, createBlingImportAdapter, createBlingPreviewAdapter, createBootstrapAdapter, type BlingConfigAdapter, type BlingImportAdapter, type BlingPreviewAdapter, type BootstrapAdapter, type BootstrapIntegration, type BootstrapState } from './bootstrap'
import './App.css'

type Screen = 'executive' | 'integrations'
type CardGroup = 'Tesouraria' | 'Recebíveis' | 'Obrigações' | 'Resultado' | 'Planejamento' | 'Gestão'
type Metric = { title: string; tooltip: TooltipId; group: CardGroup; status?: 'UNAVAILABLE' | 'PROVISIONAL' }

const metrics: Metric[] = [
  { title: 'Saldo Inicial do Dia', tooltip: 'treasury.openingBalance', group: 'Tesouraria' },
  { title: 'Saldo Final Projetado Hoje', tooltip: 'treasury.todayClosing', group: 'Tesouraria', status: 'PROVISIONAL' },
  { title: 'Saldo Projetado em 30 Dias', tooltip: 'treasury.balance30', group: 'Tesouraria', status: 'PROVISIONAL' },
  { title: 'Saldo Projetado em 60 Dias', tooltip: 'treasury.balance60', group: 'Tesouraria', status: 'PROVISIONAL' },
  { title: 'Menor Saldo Projetado — 60 Dias', tooltip: 'treasury.minimum60', group: 'Tesouraria', status: 'PROVISIONAL' },
  { title: 'Reserva Mínima', tooltip: 'treasury.minimumReserve', group: 'Tesouraria', status: 'PROVISIONAL' },
  { title: 'Valor Máximo para Aplicação', tooltip: 'treasury.maximumInvestment', group: 'Tesouraria', status: 'UNAVAILABLE' },
  { title: 'A Receber B2C — Nuvem', tooltip: 'receivables.b2cNuvem', group: 'Recebíveis', status: 'UNAVAILABLE' },
  { title: 'A Receber B2B — Bling', tooltip: 'receivables.b2bBling', group: 'Recebíveis', status: 'UNAVAILABLE' },
  { title: 'Total a Receber', tooltip: 'receivables.total', group: 'Recebíveis', status: 'UNAVAILABLE' },
  { title: 'Recebido no Mês', tooltip: 'receivables.receivedMonth', group: 'Recebíveis', status: 'UNAVAILABLE' },
  { title: 'A Pagar', tooltip: 'payables.total', group: 'Obrigações', status: 'UNAVAILABLE' },
  { title: 'Obrigações Vencidas', tooltip: 'payables.overdue', group: 'Obrigações', status: 'UNAVAILABLE' },
  { title: 'Pago no Mês', tooltip: 'payables.paidMonth', group: 'Obrigações', status: 'UNAVAILABLE' },
  { title: 'DRE / P&L', tooltip: 'result.drePnl', group: 'Resultado', status: 'UNAVAILABLE' },
  { title: 'EBITDA', tooltip: 'result.ebitda', group: 'Resultado', status: 'UNAVAILABLE' },
  { title: 'Orçado x Realizado', tooltip: 'planning.budgetActual', group: 'Planejamento', status: 'UNAVAILABLE' },
  { title: 'Forecast', tooltip: 'planning.forecast', group: 'Planejamento', status: 'UNAVAILABLE' },
  { title: 'Capital de Giro', tooltip: 'management.workingCapital', group: 'Gestão', status: 'UNAVAILABLE' },
  { title: 'Alertas', tooltip: 'management.alerts', group: 'Gestão', status: 'UNAVAILABLE' },
]

const navItems: Array<{ id: Screen; label: string; icon: string; tooltip: TooltipId }> = [
  { id: 'executive', label: 'Visão Executiva', icon: '⌂', tooltip: 'treasury.scenario' },
  { id: 'integrations', label: 'Integrações', icon: '↗', tooltip: 'source.lastUpdate' },
]

function integrationStatusLabel(status: BootstrapIntegration['status']): string {
  const labels: Record<BootstrapIntegration['status'], string> = { NOT_CONFIGURED: 'Não conectado', AUTHORIZING: 'Autorizando', CONNECTED: 'Conectado', TOKEN_EXPIRING: 'Token expirando', AUTH_ERROR: 'Erro de autorização', SCHEMA_MISMATCH: 'Formato incompatível', SYNCING: 'Sincronizando', STALE: 'Dados desatualizados', PARTIALLY_AVAILABLE: 'Parcialmente disponível', UNAVAILABLE: 'Indisponível' }
  return labels[status]
}

function integrationStatusKind(status: BootstrapIntegration['status']): Integration['statusKind'] {
  if (status === 'CONNECTED') return 'connected'
  if (status === 'PARTIALLY_AVAILABLE' || status === 'STALE' || status === 'TOKEN_EXPIRING' || status === 'SYNCING') return 'warning'
  if (status === 'UNAVAILABLE' || status === 'AUTH_ERROR' || status === 'SCHEMA_MISMATCH') return 'unavailable'
  return 'neutral'
}

function findIntegration(state: BootstrapState, provider: BootstrapIntegration['provider']): BootstrapIntegration | undefined {
  return state.integrations.find((integration) => integration.provider === provider)
}

function formatBootstrapTime(value?: string): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat('pt-BR', { dateStyle: 'short', timeStyle: 'short' }).format(date)
}

function StatusBadge({ status }: { status: Metric['status'] }) {
  if (!status) return <Tooltip id="status.noSource"><span className="status-badge status-muted">Sem fonte</span></Tooltip>
  return <Tooltip id={status === 'UNAVAILABLE' ? 'nuvempago.availability' : 'status.provisional'}><span className={`status-badge ${status === 'UNAVAILABLE' ? 'status-unavailable' : 'status-provisional'}`}>{status === 'UNAVAILABLE' ? 'UNAVAILABLE' : 'PROVISIONAL'}</span></Tooltip>
}

function MetricCard({ metric }: { metric: Metric }) {
  return <article className="metric-card"><div className="metric-card-header"><Tooltip id={metric.tooltip}><span className="metric-title">{metric.title}</span></Tooltip><Tooltip id="metric.details"><button className="icon-button" aria-label={`Ver detalhes de ${metric.title}`} type="button">↗</button></Tooltip></div><div className="metric-empty"><span className="empty-dash" aria-hidden="true">—</span><span>Sem dados confirmados</span></div><div className="metric-footer"><StatusBadge status={metric.status} /><span className="metric-source">Fonte não conectada</span></div></article>
}

function ExecutiveView({ onNavigate }: { onNavigate: (screen: Screen) => void }) {
  const [scenario, setScenario] = useState('BASE')
  return <div className="page-content"><div className="page-heading"><div><p className="eyebrow">PAINEL FINANCEIRO</p><h1>Visão Executiva</h1><p className="heading-copy">Uma leitura clara da posição financeira e dos próximos 60 dias.</p></div><div className="heading-actions"><label className="select-label" htmlFor="scenario">Cenário <Tooltip id="treasury.scenario"><span className="help-dot">?</span></Tooltip></label><Tooltip id="filter.scenario"><select id="scenario" value={scenario} onChange={(event) => setScenario(event.target.value)}><option>BASE</option><option>CONSERVATIVE</option><option>STRESS</option><option>OPTIMISTIC</option></select></Tooltip><Tooltip id="filter.referenceDate"><button type="button" className="date-control">16 ago 2026 <span aria-hidden="true">⌄</span></button></Tooltip></div></div><section className="notice notice-info" aria-label="Estado dos dados"><span className="notice-icon" aria-hidden="true">i</span><div><strong>Conecte uma fonte para começar</strong><p>A visão será preenchida após a primeira sincronização. Nenhum valor foi estimado.</p></div><Tooltip id="notice.integrations"><button type="button" className="text-link" onClick={() => onNavigate('integrations')}>Ir para Integrações <span aria-hidden="true">→</span></button></Tooltip></section>{(['Tesouraria', 'Recebíveis', 'Obrigações', 'Resultado', 'Planejamento', 'Gestão'] as CardGroup[]).map((group) => <section className="metric-section" key={group}><div className="section-heading"><h2>{group}</h2><span className="section-line" /></div><div className="metric-grid">{metrics.filter((metric) => metric.group === group).map((metric) => <MetricCard metric={metric} key={metric.title} />)}</div></section>)}<section className="projection-panel"><div className="section-heading"><h2>Projeção de liquidez</h2><Tooltip id="chart.projected"><span className="help-dot">?</span></Tooltip></div><div className="chart-empty"><div className="chart-grid-lines" aria-hidden="true"><span /><span /><span /><span /></div><p>Gráfico disponível após conectar e sincronizar uma fonte</p><span>As séries Realizado e Projetado aparecerão aqui.</span></div></section></div>
}

type Integration = { name: string; key: 'bling' | 'nuvemshop' | 'nuvempago'; provider: BootstrapIntegration['provider']; description: string; status: string; statusKind: 'neutral' | 'connected' | 'warning' | 'unavailable'; fields: Array<{ id: string; label: string; type?: string; tooltip: TooltipId; placeholder?: string }> }
const integrations: Integration[] = [
  { name: 'Bling', key: 'bling', provider: 'BLING', description: 'Contas a receber, a pagar e movimentações realizadas.', status: 'Não conectado', statusKind: 'neutral', fields: [{ id: 'bling-client-id', label: 'Client ID', tooltip: 'bling.clientId', placeholder: 'Informe o Client ID' }, { id: 'bling-secret', label: 'Client Secret', type: 'password', tooltip: 'bling.clientSecret', placeholder: '••••••••••••' }, { id: 'bling-redirect', label: 'Redirect URI', tooltip: 'bling.redirectUri', placeholder: 'https://...' }] },
  { name: 'Nuvemshop', key: 'nuvemshop', provider: 'NUVEMSHOP', description: 'Pedidos e status operacionais da loja conectada.', status: 'Não conectado', statusKind: 'neutral', fields: [{ id: 'nuvem-app-id', label: 'App ID', tooltip: 'nuvemshop.appId', placeholder: 'Informe o App ID' }, { id: 'nuvem-secret', label: 'Client Secret', type: 'password', tooltip: 'nuvemshop.clientSecret', placeholder: '••••••••••••' }, { id: 'nuvem-redirect', label: 'Redirect URI', tooltip: 'nuvemshop.redirectUri', placeholder: 'https://...' }] },
  { name: 'Nuvem Pago', key: 'nuvempago', provider: 'NUVEM_PAGO', description: 'Taxas, líquido e agenda de recebimento oficial.', status: 'Indisponível para confirmação financeira', statusKind: 'unavailable', fields: [] },
]

const nuvemPagoPricing = [
  { method: 'Cartão 1x', fee: '2,59% + R$ 0,35', receipt: 'D+30' },
  { method: 'Cartão 2x', fee: '4,49% + R$ 0,35', receipt: 'D+30' },
  { method: 'Cartão 3x', fee: '5,44% + R$ 0,35', receipt: 'D+30' },
  { method: 'Boleto', fee: 'R$ 2,39', receipt: 'D+2' },
  { method: 'PIX', fee: '0,99%', receipt: 'Na hora' },
]

function NuvemPagoPricingPanel() {
  return <section className="pricing-panel" aria-label="Tarifário Nuvem Pago aprovado"><div className="pricing-heading"><strong>Tarifário configurado</strong><Tooltip id="nuvempago.pricing"><span className="help-dot">?</span></Tooltip></div><p>Aplicado somente para estimativas identificadas; a confirmação financeira continua aguardando o ledger oficial.</p><div className="pricing-grid"><span>Forma</span><span>Taxa</span><span>Recebimento</span>{nuvemPagoPricing.map((item) => <div className="pricing-row" key={item.method}><span>{item.method}</span><strong>{item.fee}</strong><span>{item.receipt}</span></div>)}</div></section>
}

function IntegrationCard({ integration, bootstrap, blingConfig }: { integration: Integration; bootstrap: BootstrapState; blingConfig: BlingConfigAdapter }) {
  const [saving, setSaving] = useState(false)
  const [saveResult, setSaveResult] = useState<Awaited<ReturnType<BlingConfigAdapter['save']>> | null>(null)
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (integration.key !== 'bling') return
    const form = new FormData(event.currentTarget)
    setSaving(true)
    try {
      setSaveResult(await blingConfig.save({ client_id: String(form.get('bling-client-id') ?? ''), redirect_uri: String(form.get('bling-redirect') ?? ''), client_secret: String(form.get('bling-secret') ?? '') }))
    } finally {
      setSaving(false)
    }
  }
  const actionTooltip: TooltipId = integration.key === 'bling' ? 'bling.connect' : integration.key === 'nuvemshop' ? 'nuvemshop.connect' : 'nuvempago.availability'
  const liveStatus = findIntegration(bootstrap, integration.provider)
  const displayStatus = liveStatus?.status ?? (integration.key === 'nuvempago' ? 'UNAVAILABLE' : 'NOT_CONFIGURED')
  const displayLabel = integration.key === 'nuvempago' && displayStatus === 'UNAVAILABLE' ? 'Indisponível para confirmação financeira' : integrationStatusLabel(displayStatus)
  const displayKind = integrationStatusKind(displayStatus)
  return <article className={`integration-card integration-${integration.key}`}><div className="integration-header"><div className="integration-mark">{integration.key === 'bling' ? 'B' : integration.key === 'nuvemshop' ? 'N' : 'P'}</div><div><h2>{integration.name}</h2><p>{integration.description}</p></div><Tooltip id={integration.key === 'nuvempago' ? 'nuvempago.availability' : 'integration.connectionStatus'}><span className={`connection-status status-${displayKind}`}><span className="status-dot" />{displayLabel}</span></Tooltip></div>{integration.fields.length > 0 ? <form onSubmit={(event) => void submit(event)} autoComplete="off" className="integration-form"><div className="field-grid">{integration.fields.map((field) => <div className="field" key={field.id}><label htmlFor={field.id}>{field.label} <Tooltip id={field.tooltip}><span className="help-dot">?</span></Tooltip></label><Tooltip id="field.value"><input id={field.id} name={field.id} type={field.type ?? 'text'} autoComplete={field.type === 'password' ? 'new-password' : 'off'} placeholder={field.placeholder} /></Tooltip></div>)}</div><div className="integration-meta"><span>Permissões: <strong>{displayStatus === 'CONNECTED' ? 'leitura autorizada' : 'aguardando autorização'}</strong></span><span>Última tentativa: <strong>{formatBootstrapTime(liveStatus?.last_attempt_at)}</strong></span><span>Último sucesso: <strong>{formatBootstrapTime(liveStatus?.last_success_at)}</strong></span></div><p className="integration-disabled-note">Salvar configuração grava o secret no cofre Windows; autorização e sincronização continuam separadas e não são iniciadas automaticamente.</p>{saveResult && <div className={saveResult.error_code ? 'folder-import-error' : 'folder-import-result'} role="status"><strong>{saveResult.error_code ?? 'Configuração salva'}</strong><span>{saveResult.message ?? 'Os dados públicos foram registrados e o segredo foi protegido no worker.'}</span></div>}<div className="integration-actions">{integration.key === 'bling' && <Tooltip id="bling.save"><button type="submit" className="button primary" disabled={saving}>{saving ? 'Salvando…' : 'Salvar configuração'}</button></Tooltip>}<Tooltip id={actionTooltip}><button type="button" className="button primary" disabled aria-disabled="true">Conectar</button></Tooltip><Tooltip id="action.test"><button type="button" className="button secondary" disabled aria-disabled="true">Testar conexão</button></Tooltip><Tooltip id="action.reconnect"><button type="button" className="button ghost" disabled aria-disabled="true">Reconectar</button></Tooltip><Tooltip id="action.disconnect"><button type="button" className="button ghost" disabled aria-disabled="true">Desconectar</button></Tooltip><Tooltip id="action.syncNow"><button type="button" className="button ghost" disabled aria-disabled="true">Sincronizar agora</button></Tooltip></div></form> : <><NuvemPagoPricingPanel /><div className="unavailable-box"><div className="unavailable-icon" aria-hidden="true">◌</div><div><strong>Fonte oficial não disponível</strong><p>Pedidos e status não comprovam taxa, líquido ou data prevista. O valor financeiro permanece <em>UNAVAILABLE</em> até existir uma fonte oficial.</p><Tooltip id="nuvempago.partial"><span className="text-link muted-link">Por que isso importa?</span></Tooltip></div></div></>}</article>
}

function BlingImportPanel({ adapter, importer }: { adapter: BlingPreviewAdapter; importer: BlingImportAdapter }) {
  const [folder, setFolder] = useState('')
  const [preview, setPreview] = useState<Awaited<ReturnType<BlingPreviewAdapter['preview']>> | null>(null)
  const [importResult, setImportResult] = useState<Awaited<ReturnType<BlingImportAdapter['import']>> | null>(null)
  const [busy, setBusy] = useState(false)
  const validateFolder = async () => {
    if (!folder.trim()) return
    setBusy(true)
    try { setPreview(await adapter.preview(folder.trim())) } finally { setBusy(false) }
  }
  const importFolder = async () => {
    if (!folder.trim() || !preview || preview.receipt_count === 0) return
    setBusy(true)
    try { setImportResult(await importer.import(folder.trim())) } finally { setBusy(false) }
  }
  return <section className="folder-import-panel" aria-label="Importação de recebimentos do Bling"><div className="folder-import-heading"><strong>Importar recebimentos do Bling</strong><Tooltip id="bling.importFolder"><span className="help-dot">?</span></Tooltip></div><p>Coloque os relatórios exportados na pasta e valide antes de gravar os dados.</p><div className="folder-import-controls"><Tooltip id="field.value"><input aria-label="Pasta dos relatórios CSV" value={folder} onChange={(event) => { setFolder(event.target.value); setPreview(null); setImportResult(null) }} placeholder="C:\\Majucau\\imports\\bling" /></Tooltip><Tooltip id="bling.preview"><button type="button" className="button secondary" onClick={() => void validateFolder()} disabled={busy || !folder.trim()}>{busy ? 'Validando…' : 'Validar pasta'}</button></Tooltip>{preview && !preview.error_code && preview.receipt_count > 0 && <Tooltip id="bling.import"><button type="button" className="button primary" onClick={() => void importFolder()} disabled={busy}>{busy ? 'Gravando…' : 'Importar e gravar'}</button></Tooltip>}</div>{preview?.error_code ? <div className="folder-import-error" role="status"><strong>{preview.error_code}</strong><span>{preview.message}</span></div> : preview ? <div className="folder-import-result" role="status"><strong>{preview.receipt_count} recebimento(s) válido(s)</strong><span>{preview.files.length} arquivo(s) · {preview.error_count} linha(s) com atenção · {preview.ignored_count} ignorado(s)</span>{preview.issues && preview.issues.length > 0 && <ul>{preview.issues.slice(0, 3).map((issue, index) => <li key={`${issue.file}-${issue.line}-${index}`}>{issue.file}{issue.line > 0 ? ` · linha ${issue.line}` : ''}: {issue.message}</li>)}</ul>}</div> : null}{importResult && <div className={importResult.error_code ? 'folder-import-error' : 'folder-import-result'} role="status"><strong>{importResult.error_code ?? `Lote ${importResult.status}`}</strong><span>{importResult.message ?? `${importResult.records_created} criado(s), ${importResult.records_updated} atualizado(s), ${importResult.records_failed} com erro.`}</span></div>}</section>
}

function IntegrationsView({ bootstrap, blingPreview, blingImporter, blingConfig }: { bootstrap: BootstrapState; blingPreview: BlingPreviewAdapter; blingImporter: BlingImportAdapter; blingConfig: BlingConfigAdapter }) {
  return <div className="page-content"><div className="page-heading"><div><p className="eyebrow">CONFIGURAÇÕES</p><h1>Integrações</h1><p className="heading-copy">Conecte fontes oficiais para habilitar os dados do painel.</p></div><span className="security-note"><span aria-hidden="true">⌁</span> Credenciais protegidas pelo Windows</span></div><section className="notice notice-warning" aria-label="Aviso de segurança"><span className="notice-icon" aria-hidden="true">!</span><div><strong>Autorização sempre no navegador externo</strong><p>O login não é capturado nesta interface. O Client Secret digitado é transitório e será enviado somente ao worker local; ele não fica em estado, log ou armazenamento do frontend.</p></div></section>{bootstrap.error_code && <section className="notice notice-error" aria-label="Estado do worker"><span className="notice-icon" aria-hidden="true">!</span><div><strong>{bootstrap.error_code}</strong><p>{bootstrap.message}</p></div></section>}<div className="integration-list">{integrations.map((integration) => <div key={integration.key}><IntegrationCard integration={integration} bootstrap={bootstrap} blingConfig={blingConfig} />{integration.key === 'bling' && <BlingImportPanel adapter={blingPreview} importer={blingImporter} />}</div>)}</div></div>
}

function BootstrapLoading() {
  return <div className="page-content bootstrap-loading" aria-live="polite"><div className="loading-orb" aria-hidden="true" /><p className="eyebrow">INICIALIZAÇÃO</p><h1>Carregando estado local</h1><p>Consultando a saúde do worker e o status das integrações…</p></div>
}

function WorkerFooter({ bootstrap }: { bootstrap: BootstrapState }) {
  const workerLabel = bootstrap.worker.state === 'OK' ? 'Worker saudável' : bootstrap.worker.state === 'DEGRADED' ? 'Worker degradado' : 'WORKER_UNAVAILABLE'
  const integrationSummary = bootstrap.integrations.map((integration) => `${integration.provider}: ${integrationStatusLabel(integration.status)}`).join(' · ')
  return <footer className="app-footer"><span>Majucau Financial Intelligence · Fundação G1</span><span className="footer-status"><span className={`online-dot state-${bootstrap.worker.state.toLowerCase()}`} /> {workerLabel} · {integrationSummary}</span></footer>
}

type AppProps = { bootstrapAdapter?: BootstrapAdapter; blingPreviewAdapter?: BlingPreviewAdapter; blingImportAdapter?: BlingImportAdapter; blingConfigAdapter?: BlingConfigAdapter }
function App({ bootstrapAdapter, blingPreviewAdapter, blingImportAdapter, blingConfigAdapter }: AppProps = {}) {
  const [screen, setScreen] = useState<Screen>('executive')
  const adapter = useMemo(() => bootstrapAdapter ?? createBootstrapAdapter(), [bootstrapAdapter])
  const blingPreview = useMemo(() => blingPreviewAdapter ?? createBlingPreviewAdapter(), [blingPreviewAdapter])
  const blingImporter = useMemo(() => blingImportAdapter ?? createBlingImportAdapter(), [blingImportAdapter])
  const blingConfig = useMemo(() => blingConfigAdapter ?? createBlingConfigAdapter(), [blingConfigAdapter])
  const [bootstrap, setBootstrap] = useState<BootstrapState | null>(null)
  const mainRef = useRef<HTMLElement>(null)
  useEffect(() => {
    let active = true
    void adapter.getState().then((state) => { if (active) setBootstrap(state) })
    return () => { active = false }
  }, [adapter])
  const navigate = (next: Screen) => { setScreen(next); window.requestAnimationFrame(() => mainRef.current?.focus()) }
  return <div className="app-shell"><aside className="sidebar"><div className="brand"><img src={logo} alt="Majucau" /><div><strong>MAJUCAU</strong><span>FINANCIAL INTELLIGENCE</span></div></div><div className="workspace-switcher"><span className="workspace-avatar">M</span><div><strong>Minha empresa</strong><span>Workspace principal</span></div><span className="chevron" aria-hidden="true">⌄</span></div><nav aria-label="Navegação principal"><p className="nav-label">MENU PRINCIPAL</p>{navItems.map((item) => <Tooltip id={item.id === 'executive' ? 'nav.executive' : 'nav.integrations'} key={item.id}><button type="button" className={`nav-item ${screen === item.id ? 'active' : ''}`} onClick={() => navigate(item.id)} aria-current={screen === item.id ? 'page' : undefined}><span className="nav-icon" aria-hidden="true">{item.icon}</span>{item.label}</button></Tooltip>)}</nav><div className="sidebar-footer"><Tooltip id="settings.open"><button type="button" className="nav-item" onClick={() => navigate('integrations')}><span className="nav-icon" aria-hidden="true">⚙</span>Configurações</button></Tooltip><div className="user-profile"><span className="user-avatar">G</span><div><strong>Gisele</strong><span>Administradora</span></div><span className="more" aria-hidden="true">•••</span></div></div></aside><main className="main-area" ref={mainRef} tabIndex={-1}>{bootstrap === null ? <BootstrapLoading /> : screen === 'executive' ? <ExecutiveView onNavigate={navigate} /> : <IntegrationsView bootstrap={bootstrap} blingPreview={blingPreview} blingImporter={blingImporter} blingConfig={blingConfig} />} {bootstrap && <WorkerFooter bootstrap={bootstrap} />}</main></div>
}

export { App }
export default App
