import { FormEvent, useEffect, useMemo, useRef, useState } from 'react'
import logo from './assets/images/logo-universal.png'
import { Tooltip } from './components/Tooltip'
import { type TooltipId } from './tooltips'
import { createBootstrapAdapter, type BootstrapAdapter, type BootstrapIntegration, type BootstrapState } from './bootstrap'
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

function IntegrationCard({ integration, bootstrap }: { integration: Integration; bootstrap: BootstrapState }) {
  const submit = (event: FormEvent) => event.preventDefault()
  const actionTooltip: TooltipId = integration.key === 'bling' ? 'bling.connect' : integration.key === 'nuvemshop' ? 'nuvemshop.connect' : 'nuvempago.availability'
  const liveStatus = findIntegration(bootstrap, integration.provider)
  const displayStatus = liveStatus?.status ?? (integration.key === 'nuvempago' ? 'UNAVAILABLE' : 'NOT_CONFIGURED')
  const displayLabel = integration.key === 'nuvempago' && displayStatus === 'UNAVAILABLE' ? 'Indisponível para confirmação financeira' : integrationStatusLabel(displayStatus)
  const displayKind = integrationStatusKind(displayStatus)
  return <article className={`integration-card integration-${integration.key}`}><div className="integration-header"><div className="integration-mark">{integration.key === 'bling' ? 'B' : integration.key === 'nuvemshop' ? 'N' : 'P'}</div><div><h2>{integration.name}</h2><p>{integration.description}</p></div><Tooltip id={integration.key === 'nuvempago' ? 'nuvempago.availability' : 'integration.connectionStatus'}><span className={`connection-status status-${displayKind}`}><span className="status-dot" />{displayLabel}</span></Tooltip></div>{integration.fields.length > 0 ? <form onSubmit={submit} autoComplete="off" className="integration-form"><div className="field-grid">{integration.fields.map((field) => <div className="field" key={field.id}><label htmlFor={field.id}>{field.label} <Tooltip id={field.tooltip}><span className="help-dot">?</span></Tooltip></label><Tooltip id="field.value"><input id={field.id} name={field.id} type={field.type ?? 'text'} autoComplete={field.type === 'password' ? 'new-password' : 'off'} placeholder={field.placeholder} /></Tooltip></div>)}</div><div className="integration-meta"><span>Permissões: <strong>{displayStatus === 'CONNECTED' ? 'leitura autorizada' : 'aguardando autorização'}</strong></span><span>Última tentativa: <strong>{formatBootstrapTime(liveStatus?.last_attempt_at)}</strong></span><span>Último sucesso: <strong>{formatBootstrapTime(liveStatus?.last_success_at)}</strong></span></div><p className="integration-disabled-note">Fundação G1: ações de credencial e sincronização desabilitadas até existir método seguro no protocolo do worker.</p><div className="integration-actions"><Tooltip id={actionTooltip}><button type="button" className="button primary" disabled aria-disabled="true">Conectar</button></Tooltip><Tooltip id="action.test"><button type="button" className="button secondary" disabled aria-disabled="true">Testar conexão</button></Tooltip><Tooltip id="action.reconnect"><button type="button" className="button ghost" disabled aria-disabled="true">Reconectar</button></Tooltip><Tooltip id="action.disconnect"><button type="button" className="button ghost" disabled aria-disabled="true">Desconectar</button></Tooltip><Tooltip id="action.syncNow"><button type="button" className="button ghost" disabled aria-disabled="true">Sincronizar agora</button></Tooltip></div></form> : <div className="unavailable-box"><div className="unavailable-icon" aria-hidden="true">◌</div><div><strong>Fonte oficial não disponível</strong><p>Pedidos e status não comprovam taxa, líquido ou data prevista. O valor financeiro permanece <em>UNAVAILABLE</em> até existir uma fonte oficial.</p><Tooltip id="nuvempago.partial"><span className="text-link muted-link">Por que isso importa?</span></Tooltip></div></div>}</article>
}

function IntegrationsView({ bootstrap }: { bootstrap: BootstrapState }) {
  return <div className="page-content"><div className="page-heading"><div><p className="eyebrow">CONFIGURAÇÕES</p><h1>Integrações</h1><p className="heading-copy">Conecte fontes oficiais para habilitar os dados do painel.</p></div><span className="security-note"><span aria-hidden="true">⌁</span> Credenciais protegidas pelo Windows</span></div><section className="notice notice-warning" aria-label="Aviso de segurança"><span className="notice-icon" aria-hidden="true">!</span><div><strong>Autorização sempre no navegador externo</strong><p>O login não é capturado nesta interface. O Client Secret digitado é transitório e será enviado somente ao worker local quando essa conexão estiver disponível; ele não fica em estado, log ou armazenamento do frontend.</p></div></section>{bootstrap.error_code && <section className="notice notice-error" aria-label="Estado do worker"><span className="notice-icon" aria-hidden="true">!</span><div><strong>{bootstrap.error_code}</strong><p>{bootstrap.message}</p></div></section>}<div className="integration-list">{integrations.map((integration) => <IntegrationCard integration={integration} bootstrap={bootstrap} key={integration.key} />)}</div></div>
}

function BootstrapLoading() {
  return <div className="page-content bootstrap-loading" aria-live="polite"><div className="loading-orb" aria-hidden="true" /><p className="eyebrow">INICIALIZAÇÃO</p><h1>Carregando estado local</h1><p>Consultando a saúde do worker e o status das integrações…</p></div>
}

function WorkerFooter({ bootstrap }: { bootstrap: BootstrapState }) {
  const workerLabel = bootstrap.worker.state === 'OK' ? 'Worker saudável' : bootstrap.worker.state === 'DEGRADED' ? 'Worker degradado' : 'WORKER_UNAVAILABLE'
  const integrationSummary = bootstrap.integrations.map((integration) => `${integration.provider}: ${integrationStatusLabel(integration.status)}`).join(' · ')
  return <footer className="app-footer"><span>Majucau Financial Intelligence · Fundação G1</span><span className="footer-status"><span className={`online-dot state-${bootstrap.worker.state.toLowerCase()}`} /> {workerLabel} · {integrationSummary}</span></footer>
}

type AppProps = { bootstrapAdapter?: BootstrapAdapter }
function App({ bootstrapAdapter }: AppProps = {}) {
  const [screen, setScreen] = useState<Screen>('executive')
  const adapter = useMemo(() => bootstrapAdapter ?? createBootstrapAdapter(), [bootstrapAdapter])
  const [bootstrap, setBootstrap] = useState<BootstrapState | null>(null)
  const mainRef = useRef<HTMLElement>(null)
  useEffect(() => {
    let active = true
    void adapter.getState().then((state) => { if (active) setBootstrap(state) })
    return () => { active = false }
  }, [adapter])
  const navigate = (next: Screen) => { setScreen(next); window.requestAnimationFrame(() => mainRef.current?.focus()) }
  return <div className="app-shell"><aside className="sidebar"><div className="brand"><img src={logo} alt="Majucau" /><div><strong>MAJUCAU</strong><span>FINANCIAL INTELLIGENCE</span></div></div><div className="workspace-switcher"><span className="workspace-avatar">M</span><div><strong>Minha empresa</strong><span>Workspace principal</span></div><span className="chevron" aria-hidden="true">⌄</span></div><nav aria-label="Navegação principal"><p className="nav-label">MENU PRINCIPAL</p>{navItems.map((item) => <Tooltip id={item.id === 'executive' ? 'nav.executive' : 'nav.integrations'} key={item.id}><button type="button" className={`nav-item ${screen === item.id ? 'active' : ''}`} onClick={() => navigate(item.id)} aria-current={screen === item.id ? 'page' : undefined}><span className="nav-icon" aria-hidden="true">{item.icon}</span>{item.label}</button></Tooltip>)}</nav><div className="sidebar-footer"><Tooltip id="settings.open"><button type="button" className="nav-item" onClick={() => navigate('integrations')}><span className="nav-icon" aria-hidden="true">⚙</span>Configurações</button></Tooltip><div className="user-profile"><span className="user-avatar">G</span><div><strong>Gisele</strong><span>Administradora</span></div><span className="more" aria-hidden="true">•••</span></div></div></aside><main className="main-area" ref={mainRef} tabIndex={-1}>{bootstrap === null ? <BootstrapLoading /> : screen === 'executive' ? <ExecutiveView onNavigate={navigate} /> : <IntegrationsView bootstrap={bootstrap} />} {bootstrap && <WorkerFooter bootstrap={bootstrap} />}</main></div>
}

export { App }
export default App
