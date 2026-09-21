import { FormEvent, useEffect, useMemo, useRef, useState } from 'react'
import { Tooltip } from './components/Tooltip'
import { type TooltipId } from './tooltips'
import { createBlingConfigAdapter, createBlingImportAdapter, createBlingOAuthAdapter, createBlingPreviewAdapter, createBootstrapAdapter, createNuvemshopConfigAdapter, type BlingConfigAdapter, type BlingConfigResult, type BlingImportAdapter, type BlingOAuthAdapter, type BlingPreviewAdapter, type BlingSyncInput, type BootstrapAdapter, type BootstrapIntegration, type BootstrapState, type NuvemshopConfigAdapter, type NuvemshopConfigResult } from './bootstrap'
import { MockupDashboard, MockupSidebar, type MockupScreenKey } from './mockupDashboard'
import './App.css'
import './mockup.css'

type Screen = MockupScreenKey
type CardGroup = 'Tesouraria' | 'Recebíveis' | 'Obrigações' | 'Resultado' | 'Planejamento' | 'Gestão'
type MetricIcon = 'wallet' | 'trend' | 'calendar' | 'clock' | 'cart' | 'document' | 'database' | 'warning' | 'chart' | 'building' | 'settings' | 'home' | 'flow' | 'receivable' | 'payable' | 'reconcile' | 'result' | 'balance' | 'investment' | 'calculator' | 'inventory' | 'production' | 'shopping' | 'control' | 'import' | 'user' | 'export' | 'pie' | 'refresh' | 'help' | 'bank' | 'money' | 'target'
type Metric = { title: string; tooltip: TooltipId; group: CardGroup; icon: MetricIcon; status?: 'UNAVAILABLE' | 'PROVISIONAL' }

const metrics: Metric[] = [
  { title: 'Saldo Inicial do Dia', tooltip: 'treasury.openingBalance', group: 'Tesouraria', icon: 'wallet', status: 'UNAVAILABLE' },
  { title: 'Saldo Final Projetado Hoje', tooltip: 'treasury.todayClosing', group: 'Tesouraria', icon: 'trend', status: 'PROVISIONAL' },
  { title: 'Saldo Projetado em 30 Dias', tooltip: 'treasury.balance30', group: 'Tesouraria', icon: 'calendar', status: 'PROVISIONAL' },
  { title: 'Saldo Projetado em 60 Dias', tooltip: 'treasury.balance60', group: 'Tesouraria', icon: 'calendar', status: 'PROVISIONAL' },
  { title: 'Menor Saldo Projetado — 60 Dias', tooltip: 'treasury.minimum60', group: 'Tesouraria', icon: 'clock', status: 'PROVISIONAL' },
  { title: 'Reserva Mínima', tooltip: 'treasury.minimumReserve', group: 'Tesouraria', icon: 'wallet', status: 'PROVISIONAL' },
  { title: 'Valor Máximo para Aplicação', tooltip: 'treasury.maximumInvestment', group: 'Tesouraria', icon: 'trend', status: 'UNAVAILABLE' },
  { title: 'A Receber B2C — Nuvem', tooltip: 'receivables.b2cNuvem', group: 'Recebíveis', icon: 'cart', status: 'UNAVAILABLE' },
  { title: 'A Receber B2B — Bling', tooltip: 'receivables.b2bBling', group: 'Recebíveis', icon: 'document', status: 'UNAVAILABLE' },
  { title: 'Total a Receber', tooltip: 'receivables.total', group: 'Recebíveis', icon: 'database', status: 'UNAVAILABLE' },
  { title: 'Recebido no Mês', tooltip: 'receivables.receivedMonth', group: 'Recebíveis', icon: 'database', status: 'UNAVAILABLE' },
  { title: 'A Pagar', tooltip: 'payables.total', group: 'Obrigações', icon: 'wallet', status: 'UNAVAILABLE' },
  { title: 'Obrigações Vencidas', tooltip: 'payables.overdue', group: 'Obrigações', icon: 'warning', status: 'UNAVAILABLE' },
  { title: 'Pago no Mês', tooltip: 'payables.paidMonth', group: 'Obrigações', icon: 'wallet', status: 'UNAVAILABLE' },
  { title: 'DRE / P&L', tooltip: 'result.drePnl', group: 'Resultado', icon: 'chart', status: 'UNAVAILABLE' },
  { title: 'EBITDA', tooltip: 'result.ebitda', group: 'Resultado', icon: 'chart', status: 'UNAVAILABLE' },
  { title: 'Orçado x Realizado', tooltip: 'planning.budgetActual', group: 'Planejamento', icon: 'chart', status: 'UNAVAILABLE' },
  { title: 'Forecast', tooltip: 'planning.forecast', group: 'Planejamento', icon: 'trend', status: 'UNAVAILABLE' },
  { title: 'Capital de Giro', tooltip: 'management.workingCapital', group: 'Gestão', icon: 'building', status: 'UNAVAILABLE' },
  { title: 'Alertas', tooltip: 'management.alerts', group: 'Gestão', icon: 'warning', status: 'UNAVAILABLE' },
]

type NavItem = { id?: Screen; label: string; ariaLabel?: string; icon: MetricIcon; tooltip: TooltipId; section?: string }
const navItems: NavItem[] = [
  { id: 'executive', label: 'Visão Executiva', icon: 'home', tooltip: 'nav.executive', section: 'VISÃO EXECUTIVA' },
  { label: 'Fluxo de Caixa', icon: 'flow', tooltip: 'nav.treasury', section: 'TESOURARIA' },
  { label: 'Contas a Receber', icon: 'receivable', tooltip: 'nav.receivables' },
  { label: 'Contas a Pagar', icon: 'payable', tooltip: 'nav.payables' },
  { label: 'Conciliação', icon: 'reconcile', tooltip: 'nav.conciliations' },
  { label: 'DRE', icon: 'result', tooltip: 'nav.result', section: 'RESULTADOS' },
  { label: 'Balancete', icon: 'balance', tooltip: 'nav.accounting' },
  { label: 'Balanço Patrimonial', icon: 'building', tooltip: 'nav.accounting' },
  { label: 'Aplicações', icon: 'investment', tooltip: 'nav.investments', section: 'INVESTIMENTOS' },
  { label: 'Calculadora de Aplicação', icon: 'calculator', tooltip: 'nav.investments' },
  { label: 'Estoque', icon: 'inventory', tooltip: 'nav.treasury', section: 'OPERAÇÃO' },
  { label: 'Produção', icon: 'production', tooltip: 'nav.treasury' },
  { label: 'Compras', icon: 'shopping', tooltip: 'nav.treasury' },
  { label: 'Lançamentos & Pendências', icon: 'control', tooltip: 'nav.result', section: 'CONTROLE' },
  { label: 'Alertas', icon: 'warning', tooltip: 'management.alerts' },
  { id: 'integrations', label: 'Importações / Integrações', ariaLabel: 'Integrações', icon: 'import', tooltip: 'nav.integrations' },
  { label: 'Parâmetros', icon: 'settings', tooltip: 'settings.open', section: 'CONFIGURAÇÕES' },
  { label: 'Usuários', icon: 'user', tooltip: 'settings.open' },
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

function MetricGlyph({ icon }: { icon: MetricIcon }) {
  const paths: Record<MetricIcon, string> = {
    wallet: 'M3 7h18v12H3z M7 7V5h10v2 M16 13h3',
    trend: 'M3 17l6-6 4 4 8-9 M16 6h5v5',
    calendar: 'M5 4v3 M19 4v3 M4 8h16v12H4z M8 12h3 M13 12h3 M8 16h3',
    clock: 'M12 7v5l3 2 M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18',
    cart: 'M3 4h2l2 11h10l3-8H6 M9 19a1 1 0 1 0 0 .01 M17 19a1 1 0 1 0 0 .01',
    document: 'M6 3h9l4 4v14H6z M15 3v5h4 M9 12h6 M9 16h6',
    database: 'M4 6c0-2 16-2 16 0v12c0 2-16 2-16 0z M4 6c0 2 16 2 16 0 M4 12c0 2 16 2 16 0',
    warning: 'M12 3l9 17H3z M12 9v5 M12 17v.01',
    chart: 'M4 19V5 M4 19h17 M8 16v-4 M12 16V8 M16 16v-6',
    building: 'M4 20V6l8-3 8 3v14 M9 20v-5h6v5 M8 9h1 M12 9h1 M16 9h1',
    settings: 'M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8 M4 12h2 M18 12h2 M12 4v2 M12 18v2',
    home: 'M3 11 12 4l9 7 M5 10v10h14V10 M9 20v-6h6v6',
    flow: 'M4 7h16 M4 12h10 M4 17h16 M16 9l4 3-4 3',
    receivable: 'M12 3 15 8h4l-3 4 1 8H7l1-8-3-4h4z M10 12h4',
    payable: 'M5 3h10l4 4v14H5z M15 3v5h4 M8 12h6 M8 16h4',
    reconcile: 'M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18 M8 12l3 3 5-6',
    result: 'M5 3h11l3 3v15H5z M8 10h8 M8 14h8 M8 18h5',
    balance: 'M4 20h16 M6 20V9h12v11 M4 9l8-5 8 5 M9 13h2 M13 13h2 M9 16h2 M13 16h2',
    investment: 'M4 17l4-5 3 3 5-7 M14 8h5v5',
    calculator: 'M5 3h14v18H5z M8 7h8 M8 12h2 M14 12h2 M8 16h2 M14 16h2',
    inventory: 'M4 7l8-4 8 4v12l-8 4-8-4z M4 7l8 4 8-4 M12 11v12',
    production: 'M4 19V8l8-4 8 4v11 M8 19v-5h8v5 M7 10h2 M15 10h2',
    shopping: 'M3 5h2l2 10h11l3-8H6 M9 20a1 1 0 1 0 0 .01 M17 20a1 1 0 1 0 0 .01',
    control: 'M5 4h14v16H5z M8 8h8 M8 12h8 M8 16h5',
    import: 'M4 5h16v14H4z M12 8v8 M9 13l3 3 3-3',
    user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8 M5 21a7 7 0 0 1 14 0',
    export: 'M5 4h9 M14 4v5 M14 4l5 5 M19 9v11H5V9',
    pie: 'M12 3a9 9 0 1 0 9 9h-9z M12 3v9h9',
    refresh: 'M20 11a8 8 0 0 0-14-4L4 9 M4 5v4h4 M4 13a8 8 0 0 0 14 4l2-2 M20 19v-4h-4',
    help: 'M9.5 9a2.5 2.5 0 1 1 4.5 1.5c-1 1-2 1.2-2 3 M12 17v.01 M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18',
    bank: 'M3 9l9-5 9 5 M5 10v8 M9 10v8 M15 10v8 M19 10v8 M3 21h18',
    money: 'M4 6h16v12H4z M8 12h8 M7 9h.01 M17 15h.01',
    target: 'M12 3a9 9 0 1 0 9 9 M12 7a5 5 0 1 0 5 5 M12 11a1 1 0 1 0 1 1',
  }
  return <svg className="metric-glyph" viewBox="0 0 24 24" aria-hidden="true"><path d={paths[icon]} fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" /></svg>
}

function StatusBadge({ status }: { status: Metric['status'] }) {
  const label = status === 'PROVISIONAL' ? 'Provisório / em andamento' : status === 'UNAVAILABLE' ? 'Sem fonte confirmada' : 'Sem dados confirmados'
  const tooltip = status === 'PROVISIONAL' ? 'status.provisional' : status === 'UNAVAILABLE' ? 'nuvempago.availability' : 'status.noSource'
  return <Tooltip id={tooltip}><span className={`status-badge ${status === 'PROVISIONAL' ? 'status-provisional' : status === 'UNAVAILABLE' ? 'status-unavailable' : 'status-muted'}`}><span className="status-dot" />{label}</span></Tooltip>
}

function MetricCard({ metric }: { metric: Metric }) {
  return <article className={`metric-card metric-card-${metric.status?.toLowerCase() ?? 'empty'}`}><div className="metric-card-header"><div className="metric-heading"><MetricGlyph icon={metric.icon} /><Tooltip id={metric.tooltip}><span className="metric-title">{metric.title}</span></Tooltip></div><Tooltip id="metric.details"><button className="icon-button" aria-label={`Ver detalhes de ${metric.title}`} type="button">ⓘ</button></Tooltip></div><div className="metric-value">—</div><div className="metric-subtitle">Sem dados confirmados</div><div className="metric-footer"><StatusBadge status={metric.status} /></div></article>
}

function RealizedPanel({ title, tooltip, icon }: { title: string; tooltip: TooltipId; icon: MetricIcon }) {
    return <section className="realized-panel"><div className="panel-title"><div className="metric-heading"><MetricGlyph icon={icon} /><Tooltip id={tooltip}><h2>{title}</h2></Tooltip></div><Tooltip id="metric.details"><button type="button" className="icon-button" aria-label={`Ver detalhes de ${title}`}>ⓘ</button></Tooltip></div><div className="realized-value">—</div><div className="realized-caption">Sem dados confirmados do Bling</div><div className="realized-buckets">{['Hoje (D0)', '7 dias (D+7)', '15 dias (D+15)', '30 dias (D+30)', '45 dias (D+45)', '60 dias (D+60)'].map((bucket) => <span key={bucket}><strong>—</strong><small>{bucket}</small></span>)}</div></section>
}

function EmptyChart({ title, tooltip, kind }: { title: string; tooltip: TooltipId; kind: 'balance' | 'flow' }) {
  return <section className={`visual-panel chart-panel chart-${kind}`}><div className="panel-title"><Tooltip id={tooltip}><h2>{title}</h2></Tooltip><Tooltip id="metric.details"><button type="button" className="icon-button" aria-label={`Ver detalhes de ${title}`}>ⓘ</button></Tooltip></div><div className="chart-legend"><span className="legend-muted">Sem série confirmada</span></div><div className="chart-area" aria-label={`${title}: sem série confirmada`}><div className="chart-grid-lines" aria-hidden="true"><span /><span /><span /><span /></div><span className="chart-empty-label">Sem dados confirmados</span></div></section>
}

function ResultPanel({ type }: { type: 'dre' | 'ebitda' }) {
  const title = type === 'dre' ? 'DRE / P&L' : 'EBITDA'
  const tooltip: TooltipId = type === 'dre' ? 'result.drePnl' : 'result.ebitda'
  return <section className={`visual-panel result-panel result-${type}`}><div className="panel-title"><Tooltip id={tooltip}><h2>{title}</h2></Tooltip><Tooltip id="metric.details"><button type="button" className="icon-button" aria-label={`Ver detalhes de ${title}`}>ⓘ</button></Tooltip></div><div className="result-empty">—</div><p>Sem dados confirmados</p>{type === 'ebitda' && <div className="result-subrows"><span>Margem EBITDA <strong>—</strong></span><span>Realizado / Orçado <strong>—</strong></span></div>}</section>
}

  function SmallMetric({ metric }: { metric: Metric }) {
    return <div className="small-metric" data-label={metric.title} aria-label={metric.title}><div className="metric-heading"><MetricGlyph icon={metric.icon} /><Tooltip id={metric.tooltip}><span className="small-metric-label" aria-hidden="true">Dados financeiros</span></Tooltip></div>{metric.title !== 'Capital de Giro' && <span className="sr-only">{metric.title}</span>}<span>—</span><span className="sr-only">Sem dados confirmados</span><StatusBadge status={metric.status} /></div>
  }

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function ExecutiveView({ onNavigate }: { onNavigate: (screen: Screen) => void }) {
  const [scenario, setScenario] = useState('BASE')
  const treasury = metrics.filter((metric) => metric.group === 'Tesouraria')
  const receivables = metrics.filter((metric) => metric.group === 'Recebíveis').slice(0, 3)
  const payables = metrics.filter((metric) => metric.group === 'Obrigações').slice(0, 2)
  const planning = metrics.filter((metric) => metric.group === 'Planejamento')
  const management = metrics.filter((metric) => metric.group === 'Gestão')
  return <div className="page-content executive-page"><header className="dashboard-header"><div><h1>Visão Executiva</h1><p>Panorama financeiro e econômico da empresa</p></div><div className="dashboard-header-actions"><Tooltip id="filter.referenceDate"><button type="button" className="dashboard-control"><span className="control-icon">▣</span>Data-base: — <span aria-hidden="true">⌄</span></button></Tooltip><Tooltip id="source.lastUpdate"><button type="button" className="dashboard-control"><span className="control-icon">⟳</span>Última atualização: —</button></Tooltip></div></header><div className="dashboard-context"><span>Dados financeiros exibidos conforme fontes oficiais e regras aprovadas.</span><Tooltip id="filter.scenario"><label htmlFor="scenario">Cenário <select id="scenario" value={scenario} onChange={(event) => setScenario(event.target.value)}><option>BASE</option><option>CONSERVATIVE</option><option>STRESS</option><option>OPTIMISTIC</option></select></label></Tooltip></div><section className="dashboard-row treasury-row" aria-label="Tesouraria">{treasury.map((metric) => <MetricCard metric={metric} key={metric.title} />)}</section><section className="dashboard-row five-column-row" aria-label="Recebíveis e obrigações">{[...receivables, ...payables].map((metric) => <MetricCard metric={metric} key={metric.title} />)}</section><section className="dashboard-row realized-row" aria-label="Movimentações realizadas"><RealizedPanel title="Recebido no Mês" tooltip="receivables.receivedMonth" icon="database" /><RealizedPanel title="Pago no Mês" tooltip="payables.paidMonth" icon="wallet" /></section><section className="dashboard-lower-grid"><EmptyChart title="EVOLUÇÃO DO SALDO PROJETADO (60 DIAS)" tooltip="treasury.balance60" kind="balance" /><EmptyChart title="FLUXO PROJETADO (ENTRADAS X SAÍDAS)" tooltip="chart.projected" kind="flow" /><ResultPanel type="dre" /><ResultPanel type="ebitda" /></section><section className="dashboard-bottom-grid"><section className="visual-panel alerts-panel"><div className="panel-title"><div className="metric-heading"><MetricGlyph icon="warning" /><Tooltip id="management.alerts"><h2>ALERTAS E PONTOS DE ATENÇÃO</h2></Tooltip></div><Tooltip id="metric.details"><button type="button" className="icon-button" aria-label="Ver detalhes dos alertas">ⓘ</button></Tooltip></div><div className="alert-list"><div className="alert-row alert-info"><span className="alert-symbol">i</span><span>Nenhum valor foi estimado. Conecte as fontes oficiais para preencher o painel.</span><Tooltip id="notice.integrations"><button type="button" className="text-link" onClick={() => onNavigate('integrations')}>Ir para Integrações →</button></Tooltip></div><div className="alert-row"><span className="alert-symbol">!</span><span>Obrigações vencidas: sem dados confirmados.</span><Tooltip id="payables.overdue"><button type="button" className="text-link" disabled>Ver contas</button></Tooltip></div><div className="alert-row"><span className="alert-symbol">!</span><span>Recebíveis B2B: classificação aguardando fonte Bling.</span><Tooltip id="receivables.b2bBling"><button type="button" className="text-link" disabled>Ver recebíveis</button></Tooltip></div></div><div className="compact-metrics">{planning.map((metric) => <SmallMetric metric={metric} key={metric.title} />)}</div></section><section className="visual-panel updates-panel"><div className="panel-title"><div className="metric-heading"><MetricGlyph icon="clock" /><Tooltip id="source.lastUpdate"><h2>ÚLTIMAS ATUALIZAÇÕES</h2></Tooltip></div><Tooltip id="metric.details"><button type="button" className="icon-button" aria-label="Ver histórico de atualizações">ⓘ</button></Tooltip></div><div className="updates-empty">Sem atualizações confirmadas.</div><div className="compact-metrics">{management.map((metric) => <SmallMetric metric={metric} key={metric.title} />)}</div></section></section></div>
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

function BlingSyncPanel({ adapter, open, connected }: { adapter: BlingOAuthAdapter; open: boolean; connected: boolean }) {
  const [mode, setMode] = useState<'due' | 'received' | 'payment'>('due')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [status, setStatus] = useState('')
  const [busy, setBusy] = useState(false)
  const [validation, setValidation] = useState('')
  const [result, setResult] = useState<Awaited<ReturnType<BlingOAuthAdapter['sync']>> | null>(null)

  if (!open) return null

  const buildInput = (): BlingSyncInput => {
    const input: BlingSyncInput = { status: status.trim() || undefined }
    if (mode === 'due') { input.due_date_from = from || undefined; input.due_date_to = to || undefined }
    if (mode === 'received') { input.received_date_from = from || undefined; input.received_date_to = to || undefined }
    if (mode === 'payment') { input.payment_date_from = from || undefined; input.payment_date_to = to || undefined }
    return input
  }
  const run = async () => {
    if (!from && !to && !status.trim()) { setValidation('Informe pelo menos uma data ou uma situação para limitar a leitura.'); return }
    if (from && to && from > to) { setValidation('A data inicial não pode ser posterior à data final.'); return }
    setValidation('')
    setBusy(true)
    try { setResult(await adapter.sync(buildInput())) } finally { setBusy(false) }
  }
  const resultClass = result?.error_code ? 'folder-import-error' : 'folder-import-result'
  return <div id="bling-sync-panel" className="sync-panel" role="region" aria-label="Sincronização filtrada do Bling"><div className="sync-panel-heading"><strong>Sincronizar dados do Bling</strong><Tooltip id="bling.sync.info"><span className="help-dot">?</span></Tooltip></div><p>Escolha um filtro explícito. A leitura consulta recebíveis e obrigações do Bling; nada é sincronizado automaticamente.</p><div className="sync-field-grid"><div className="field"><label htmlFor="bling-sync-mode">Tipo de data <Tooltip id="bling.sync.mode"><span className="help-dot">?</span></Tooltip></label><Tooltip id="bling.sync.mode"><select id="bling-sync-mode" value={mode} onChange={(event) => setMode(event.target.value as typeof mode)}><option value="due">Vencimento</option><option value="received">Recebimento</option><option value="payment">Pagamento</option></select></Tooltip></div><div className="field"><label htmlFor="bling-sync-from">Data inicial <Tooltip id="bling.sync.period"><span className="help-dot">?</span></Tooltip></label><Tooltip id="bling.sync.period"><input id="bling-sync-from" type="date" value={from} onChange={(event) => setFrom(event.target.value)} /></Tooltip></div><div className="field"><label htmlFor="bling-sync-to">Data final <Tooltip id="bling.sync.period"><span className="help-dot">?</span></Tooltip></label><Tooltip id="bling.sync.period"><input id="bling-sync-to" type="date" value={to} onChange={(event) => setTo(event.target.value)} /></Tooltip></div><div className="field"><label htmlFor="bling-sync-status">Situação (opcional) <Tooltip id="bling.sync.status"><span className="help-dot">?</span></Tooltip></label><Tooltip id="bling.sync.status"><input id="bling-sync-status" value={status} onChange={(event) => setStatus(event.target.value)} placeholder="Ex.: em aberto" /></Tooltip></div></div>{!connected && <div className="sync-panel-warning" role="status">Conecte e autorize o Bling antes de iniciar a sincronização.</div>}{validation && <div className="folder-import-error" role="alert"><strong>Filtro inválido</strong><span>{validation}</span></div>}<div className="sync-panel-actions"><Tooltip id="bling.sync.submit"><button type="button" className="button primary" onClick={() => void run()} disabled={busy || !connected}>{busy ? 'Sincronizando…' : 'Executar sincronização'}</button></Tooltip></div>{result && <div className={resultClass} role="status"><strong>{result.error_code ?? `Lote ${result.status ?? 'concluído'}`}</strong><span>{result.message ?? `Recebíveis: ${result.receivables.records_read} registro(s) em ${result.receivables.pages_read} página(s). Obrigações: ${result.payables.records_read} registro(s) em ${result.payables.pages_read} página(s).`}</span></div>}</div>
}

function IntegrationCard({ integration, bootstrap, blingConfig, nuvemshopConfig, blingOAuth }: { integration: Integration; bootstrap: BootstrapState; blingConfig: BlingConfigAdapter; nuvemshopConfig: NuvemshopConfigAdapter; blingOAuth: BlingOAuthAdapter }) {
  const [saving, setSaving] = useState(false)
  const [saveResult, setSaveResult] = useState<(BlingConfigResult | NuvemshopConfigResult) | null>(null)
  const [oauthBusy, setOauthBusy] = useState(false)
  const [oauthSessionId, setOauthSessionId] = useState('')
  const [oauthResult, setOauthResult] = useState<Awaited<ReturnType<BlingOAuthAdapter['status']>> | null>(null)
  const [testBusy, setTestBusy] = useState(false)
  const [testResult, setTestResult] = useState<Awaited<ReturnType<BlingOAuthAdapter['test']>> | null>(null)
  const [syncOpen, setSyncOpen] = useState(false)
  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (integration.key !== 'bling' && integration.key !== 'nuvemshop') return
    const form = new FormData(event.currentTarget)
    setSaving(true)
    try {
      if (integration.key === 'bling') {
        setSaveResult(await blingConfig.save({ client_id: String(form.get('bling-client-id') ?? ''), redirect_uri: String(form.get('bling-redirect') ?? ''), client_secret: String(form.get('bling-secret') ?? '') }))
      } else {
        setSaveResult(await nuvemshopConfig.save({ app_id: String(form.get('nuvem-app-id') ?? ''), redirect_uri: String(form.get('nuvem-redirect') ?? ''), client_secret: String(form.get('nuvem-secret') ?? '') }))
      }
    } finally { setSaving(false) }
  }
  useEffect(() => {
    if (integration.key !== 'bling' || !oauthSessionId) return undefined
    let active = true
    const poll = async () => {
      const result = await blingOAuth.status(oauthSessionId)
      if (!active) return
      setOauthResult(result)
      if (result.status !== 'AUTHORIZING') setOauthSessionId('')
    }
    void poll()
    const timer = window.setInterval(() => void poll(), 1000)
    return () => { active = false; window.clearInterval(timer) }
  }, [blingOAuth, integration.key, oauthSessionId])
  const connect = async () => {
    setOauthBusy(true)
    setOauthResult(null)
    try {
      const result = await blingOAuth.start()
      setOauthResult(result)
      if (result.session_id) setOauthSessionId(result.session_id)
    } finally { setOauthBusy(false) }
  }
  const testConnection = async () => {
    setTestBusy(true)
    try { setTestResult(await blingOAuth.test()) } finally { setTestBusy(false) }
  }
  const actionTooltip: TooltipId = integration.key === 'bling' ? 'bling.connect' : integration.key === 'nuvemshop' ? 'nuvemshop.connect' : 'nuvempago.availability'
  const liveStatus = findIntegration(bootstrap, integration.provider)
  const displayStatus = oauthResult?.status === 'CONNECTED' ? 'CONNECTED' : oauthResult?.status === 'AUTHORIZING' ? 'AUTHORIZING' : (liveStatus?.status ?? (integration.key === 'nuvempago' ? 'UNAVAILABLE' : 'NOT_CONFIGURED'))
  const displayLabel = integration.key === 'nuvempago' && displayStatus === 'UNAVAILABLE' ? 'Indisponível para confirmação financeira' : integrationStatusLabel(displayStatus)
  const displayKind = integrationStatusKind(displayStatus)
  return <article className={`integration-card integration-${integration.key}`}>
    <div className="integration-header"><div className="integration-mark">{integration.key === 'bling' ? 'B' : integration.key === 'nuvemshop' ? 'N' : 'P'}</div><div><h2>{integration.name}</h2><p>{integration.description}</p></div><Tooltip id={integration.key === 'nuvempago' ? 'nuvempago.availability' : 'integration.connectionStatus'}><span className={`connection-status status-${displayKind}`}><span className="status-dot" />{displayLabel}</span></Tooltip></div>
    {integration.fields.length > 0 ? <form onSubmit={(event) => void submit(event)} autoComplete="off" className="integration-form">
      <div className="field-grid">{integration.fields.map((field) => <div className="field" key={field.id}><label htmlFor={field.id}>{field.label} <Tooltip id={field.tooltip}><span className="help-dot">?</span></Tooltip></label><Tooltip id="field.value"><input id={field.id} name={field.id} type={field.type ?? 'text'} autoComplete={field.type === 'password' ? 'new-password' : 'off'} placeholder={field.placeholder} /></Tooltip></div>)}</div>
      <div className="integration-meta"><span>Permissões: <strong>{displayStatus === 'CONNECTED' ? 'leitura autorizada' : 'aguardando autorização'}</strong></span><span>Última tentativa: <strong>{formatBootstrapTime(liveStatus?.last_attempt_at)}</strong></span><span>Último sucesso: <strong>{formatBootstrapTime(liveStatus?.last_success_at)}</strong></span></div>
      <p className="integration-disabled-note">Salvar configuração grava o secret no cofre Windows; autorização e sincronização continuam separadas e não são iniciadas automaticamente.</p>
      {saveResult && <div className={saveResult.error_code ? 'folder-import-error' : 'folder-import-result'} role="status"><strong>{saveResult.error_code ?? 'Configuração salva'}</strong><span>{saveResult.message ?? 'Os dados públicos foram registrados e o segredo foi protegido no worker.'}</span></div>}
      {oauthResult?.message && <Tooltip id="bling.oauth.status"><div className={oauthResult.error_code ? 'folder-import-error' : 'folder-import-result'} role="status"><strong>{oauthResult.error_code ?? `OAuth ${oauthResult.status ?? 'em andamento'}`}</strong><span>{oauthResult.message}</span></div></Tooltip>}
      {testResult && <div className={testResult.error_code ? 'folder-import-error' : 'folder-import-result'} role="status"><strong>{testResult.error_code ?? 'Conexão testada'}</strong><span>{testResult.message ?? `${testResult.page_record_count} registro(s) retornado(s) na leitura mínima.`}</span></div>}
      <BlingSyncPanel adapter={blingOAuth} open={integration.key === 'bling' && syncOpen} connected={displayStatus === 'CONNECTED'} />
      <div className="integration-actions">{(integration.key === 'bling' || integration.key === 'nuvemshop') && <Tooltip id={integration.key === 'bling' ? 'bling.save' : 'nuvemshop.save'}><button type="submit" className="button primary" disabled={saving} aria-label={integration.key === 'bling' ? 'Salvar configuração' : 'Salvar configuração da Nuvemshop'}>{saving ? 'Salvando…' : 'Salvar configuração'}</button></Tooltip>}<Tooltip id={actionTooltip}><button type="button" className="button primary" onClick={() => void connect()} disabled={integration.key !== 'bling' || oauthBusy || displayStatus === 'AUTHORIZING'} aria-disabled={integration.key !== 'bling' || undefined}>{oauthBusy ? 'Abrindo…' : 'Conectar'}</button></Tooltip><Tooltip id="action.test"><button type="button" className="button secondary" onClick={() => void testConnection()} disabled={integration.key !== 'bling' || testBusy} aria-disabled={integration.key !== 'bling' || undefined}>{testBusy ? 'Testando…' : 'Testar conexão'}</button></Tooltip><Tooltip id="action.reconnect"><button type="button" className="button ghost" disabled aria-disabled="true">Reconectar</button></Tooltip><Tooltip id="action.disconnect"><button type="button" className="button ghost" disabled aria-disabled="true">Desconectar</button></Tooltip><Tooltip id="action.syncNow"><button type="button" className="button ghost" disabled={integration.key !== 'bling'} aria-expanded={integration.key === 'bling' ? syncOpen : undefined} aria-controls={integration.key === 'bling' ? 'bling-sync-panel' : undefined} onClick={() => setSyncOpen((open) => !open)}>{syncOpen ? 'Fechar sincronização' : 'Sincronizar agora'}</button></Tooltip></div>
    </form> : <><NuvemPagoPricingPanel /><div className="unavailable-box"><div className="unavailable-icon" aria-hidden="true">◌</div><div><strong>Fonte oficial não disponível</strong><p>Pedidos e status não comprovam taxa, líquido ou data prevista. O valor financeiro permanece <em>UNAVAILABLE</em> até existir uma fonte oficial.</p><Tooltip id="nuvempago.partial"><span className="text-link muted-link">Por que isso importa?</span></Tooltip></div></div></>}
  </article>
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

function IntegrationsView({ bootstrap, blingPreview, blingImporter, blingConfig, nuvemshopConfig, blingOAuth }: { bootstrap: BootstrapState; blingPreview: BlingPreviewAdapter; blingImporter: BlingImportAdapter; blingConfig: BlingConfigAdapter; nuvemshopConfig: NuvemshopConfigAdapter; blingOAuth: BlingOAuthAdapter }) {
  return <div className="page-content"><div className="page-heading"><div><p className="eyebrow">CONFIGURAÇÕES</p><h1>Integrações</h1><p className="heading-copy">Conecte fontes oficiais para habilitar os dados do painel.</p></div><span className="security-note"><span aria-hidden="true">⌁</span> Credenciais protegidas pelo Windows</span></div><section className="notice notice-warning" aria-label="Aviso de segurança"><span className="notice-icon" aria-hidden="true">!</span><div><strong>Autorização sempre no navegador externo</strong><p>O login não é capturado nesta interface. O Client Secret digitado é transitório e será enviado somente ao worker local; ele não fica em estado, log ou armazenamento do frontend.</p></div></section>{bootstrap.error_code && <section className="notice notice-error" aria-label="Estado do worker"><span className="notice-icon" aria-hidden="true">!</span><div><strong>{bootstrap.error_code}</strong><p>{bootstrap.message}</p></div></section>}<div className="integration-list">{integrations.map((integration) => <div key={integration.key}><IntegrationCard integration={integration} bootstrap={bootstrap} blingConfig={blingConfig} nuvemshopConfig={nuvemshopConfig} blingOAuth={blingOAuth} />{integration.key === 'bling' && <BlingImportPanel adapter={blingPreview} importer={blingImporter} />}</div>)}</div></div>
}

function BootstrapLoading() {
  return <div className="page-content bootstrap-loading" aria-live="polite"><div className="loading-orb" aria-hidden="true" /><p className="eyebrow">INICIALIZAÇÃO</p><h1>Carregando estado local</h1><p>Consultando a saúde do worker e o status das integrações…</p></div>
}

function WorkerFooter({ bootstrap }: { bootstrap: BootstrapState }) {
  const workerLabel = bootstrap.worker.state === 'OK' ? 'Worker saudável' : bootstrap.worker.state === 'DEGRADED' ? 'Worker degradado' : 'WORKER_UNAVAILABLE'
  const integrationSummary = bootstrap.integrations.map((integration) => `${integration.provider}: ${integrationStatusLabel(integration.status)}`).join(' · ')
  return <footer className="app-footer"><span>Majucau Financial Intelligence · Fundação G1</span><span className="footer-status"><span className={`online-dot state-${bootstrap.worker.state.toLowerCase()}`} /> {workerLabel} · {integrationSummary}</span></footer>
}

function CocoaMark() {
  return <svg className="cocoa-mark" viewBox="0 0 32 42" aria-hidden="true"><path d="M16 1C8 7 4 15 5 25c1 9 6 15 11 16 5-1 10-7 11-16C28 15 24 7 16 1Z" fill="none" stroke="currentColor" strokeWidth="1.4" /><path d="M16 3v35M16 14c-4-3-7-6-8-10M16 21c4-3 7-6 9-10M16 29c-4-2-7-4-9-7M16 34c4-2 7-4 9-7" fill="none" stroke="currentColor" strokeWidth="1" strokeLinecap="round" /></svg>
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
function Sidebar({ screen, onNavigate }: { screen: Screen; onNavigate: (screen: Screen) => void }) {
  return <aside className="sidebar"><div className="brand"><CocoaMark /><div><strong>MAJUCAU</strong><span>CHOCOLATE BRASILEIRO</span></div><Tooltip id="settings.open"><button type="button" className="sidebar-collapse" aria-label="Recolher menu">‹</button></Tooltip></div><nav aria-label="Navegação principal"><p className="nav-label">MENU PRINCIPAL</p>{navItems.map((item) => <Tooltip id={item.tooltip} key={item.label}><button type="button" className={`nav-item ${item.id && screen === item.id ? 'active' : ''}`} onClick={() => item.id && onNavigate(item.id)} disabled={!item.id} aria-disabled={!item.id || undefined} aria-current={item.id && screen === item.id ? 'page' : undefined} aria-label={item.ariaLabel}><MetricGlyph icon={item.icon} /><span>{item.label}</span></button></Tooltip>)}</nav><section className="global-filters" aria-label="Filtros globais"><h2>FILTROS GLOBAIS</h2><label htmlFor="global-base-date">Período base</label><Tooltip id="filter.referenceDate"><input id="global-base-date" type="date" aria-label="Período base" /></Tooltip><label htmlFor="global-company">Empresa</label><Tooltip id="field.value"><select id="global-company" defaultValue=""><option value="">Selecione a empresa</option></select></Tooltip><label htmlFor="global-scenario">Cenário</label><Tooltip id="treasury.scenario"><select id="global-scenario" defaultValue="BASE"><option>BASE</option><option>CONSERVATIVE</option><option>STRESS</option><option>OPTIMISTIC</option></select></Tooltip><Tooltip id="filter.referenceDate"><button type="button" className="clear-filters">›&nbsp; Limpar filtros</button></Tooltip></section><div className="sidebar-footer"><Tooltip id="settings.open"><button type="button" className="nav-item sidebar-settings" onClick={() => onNavigate('integrations')}><MetricGlyph icon="settings" /><span>Configurações</span></button></Tooltip><div className="user-profile"><span className="user-avatar">G</span><div><strong>Gisele</strong><span>Administradora</span></div><span className="more" aria-hidden="true">•••</span></div></div></aside>
}

type AppProps = { bootstrapAdapter?: BootstrapAdapter; blingPreviewAdapter?: BlingPreviewAdapter; blingImportAdapter?: BlingImportAdapter; blingConfigAdapter?: BlingConfigAdapter; nuvemshopConfigAdapter?: NuvemshopConfigAdapter; blingOAuthAdapter?: BlingOAuthAdapter }
function App({ bootstrapAdapter, blingPreviewAdapter, blingImportAdapter, blingConfigAdapter, nuvemshopConfigAdapter, blingOAuthAdapter }: AppProps = {}) {
  const [screen, setScreen] = useState<Screen>('executive')
  const adapter = useMemo(() => bootstrapAdapter ?? createBootstrapAdapter(), [bootstrapAdapter])
  const blingPreview = useMemo(() => blingPreviewAdapter ?? createBlingPreviewAdapter(), [blingPreviewAdapter])
  const blingImporter = useMemo(() => blingImportAdapter ?? createBlingImportAdapter(), [blingImportAdapter])
  const blingConfig = useMemo(() => blingConfigAdapter ?? createBlingConfigAdapter(), [blingConfigAdapter])
  const nuvemshopConfig = useMemo(() => nuvemshopConfigAdapter ?? createNuvemshopConfigAdapter(), [nuvemshopConfigAdapter])
  const blingOAuth = useMemo(() => blingOAuthAdapter ?? createBlingOAuthAdapter(), [blingOAuthAdapter])
  const [bootstrap, setBootstrap] = useState<BootstrapState | null>(null)
  const mainRef = useRef<HTMLElement>(null)
  useEffect(() => {
    let active = true
    void adapter.getState().then((state) => { if (active) setBootstrap(state) })
    return () => { active = false }
  }, [adapter])
  const navigate = (next: Screen) => { setScreen(next); window.requestAnimationFrame(() => mainRef.current?.focus()) }
  return <div className="app-shell"><MockupSidebar screen={screen} onNavigate={navigate} /><main className="main-area" ref={mainRef} tabIndex={-1}>{bootstrap === null ? <BootstrapLoading /> : screen === 'integrations' ? <IntegrationsView bootstrap={bootstrap} blingPreview={blingPreview} blingImporter={blingImporter} blingConfig={blingConfig} nuvemshopConfig={nuvemshopConfig} blingOAuth={blingOAuth} /> : <MockupDashboard screen={screen} onNavigate={navigate} />} {bootstrap && <WorkerFooter bootstrap={bootstrap} />}</main></div>
}

export { App }
export default App
