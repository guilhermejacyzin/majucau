import { useState } from 'react'
import { Tooltip } from './components/Tooltip'
import type { TooltipId } from './tooltips'
import type { DashboardBankBalance, DashboardCashFlowPeriod, DashboardMetric, DashboardPayableRow, DashboardReceivableRow, DashboardSnapshot } from './bootstrap'
import { formatBRL, formatPercentVariation } from './money'
import sidebarReference from './assets/images/mocap-sidebar.png'

export type MockupScreenKey = 'executive' | 'cashflow' | 'receivables' | 'payables' | 'dre' | 'trial-balance' | 'investment-calculator' | 'inventory' | 'production' | 'purchases' | 'forecast' | 'pending' | 'alerts' | 'integrations' | 'parameters' | 'users'

type MockupIcon = 'home' | 'flow' | 'receivable' | 'payable' | 'reconcile' | 'result' | 'balance' | 'investment' | 'calculator' | 'inventory' | 'production' | 'shopping' | 'control' | 'import' | 'user' | 'wallet' | 'trend' | 'calendar' | 'clock' | 'cart' | 'document' | 'database' | 'warning' | 'chart' | 'building' | 'settings' | 'pie' | 'refresh' | 'export' | 'help' | 'bank' | 'money' | 'target' | 'box' | 'flag' | 'leaf'

const iconPaths: Record<MockupIcon, string> = {
  home: 'M3 11 12 4l9 7 M5 10v10h14V10 M9 20v-6h6', flow: 'M4 7h16 M4 12h10 M4 17h16 M16 9l4 3-4 3', receivable: 'M12 3 15 8h4l-3 4 1 8H7l1-8-3-4h4z M10 12h4', payable: 'M5 3h10l4 4v14H5z M15 3v5h4 M8 12h6 M8 16h4', reconcile: 'M12 3a9 9 0 1 0 0 18 9 9 0 0 0-18 M8 12l3 3 5-6', result: 'M5 3h11l3 3v15H5z M8 10h8 M8 14h8 M8 18h5', balance: 'M4 20h16 M6 20V9h12v11 M4 9l8-5 8 5 M9 13h2 M13 13h2', investment: 'M4 17l4-5 3 3 5-7 M14 8h5v5', calculator: 'M5 3h14v18H5z M8 7h8 M8 12h2 M14 12h2 M8 16h2 M14 16h2', inventory: 'M4 7l8-4 8 4v12l-8 4-8-4z M4 7l8 4 8-4 M12 11v12', production: 'M4 19V8l8-4 8 4v11 M8 19v-5h8v5 M7 10h2 M15 10h2', shopping: 'M3 5h2l2 10h11l3-8H6 M9 20a1 1 0 1 0 0 .01 M17 20a1 1 0 1 0 0 .01', control: 'M5 4h14v16H5z M8 8h8 M8 12h8 M8 16h5', import: 'M4 5h16v14H4z M12 8v8 M9 13l3 3 3-3', user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8 M5 21a7 7 0 0 1 14 0', wallet: 'M3 7h18v12H3z M7 7V5h10v2 M16 13h3', trend: 'M3 17l6-6 4 4 8-9 M16 6h5v5', calendar: 'M5 4v3 M19 4v3 M4 8h16v12H4z M8 12h3 M13 12h3 M8 16h3', clock: 'M12 7v5l3 2 M12 3a9 9 0 1 0 0 18 9 9 0 0 0-18', cart: 'M3 4h2l2 11h10l3-8H6 M9 19a1 1 0 1 0 0 .01 M17 19a1 1 0 1 0 0 .01', document: 'M6 3h9l4 4v14H6z M15 3v5h4 M9 12h6 M9 16h6', database: 'M4 6c0-2 16-2 16 0v12c0 2-16 2-16 0z M4 6c0 2 16 2 16 0 M4 12c0 2 16 2 16 0', warning: 'M12 3l9 17H3z M12 9v5 M12 17v.01', chart: 'M4 19V5 M4 19h17 M8 16v-4 M12 16V8 M16 16v-6', building: 'M4 20V6l8-3 8 3v14 M9 20v-5h6v5 M8 9h1 M12 9h1 M16 9h1', settings: 'M12 8a4 4 0 1 0 0 8 4 4 0 0 0-8 M4 12h2 M18 12h2 M12 4v2 M12 18v2', pie: 'M12 3a9 9 0 1 0 9 9h-9z M12 3v9h9', refresh: 'M20 11a8 8 0 0 0-14-4L4 9 M4 5v4h4 M4 13a8 8 0 0 0 14 4l2-2 M20 19v-4h-4', export: 'M5 4h9 M14 4v5 M14 4l5 5 M19 9v11H5V9', help: 'M9.5 9a2.5 2.5 0 1 1 4.5 1.5c-1 1-2 1.2-2 3 M12 17v.01 M12 3a9 9 0 1 0 0 18 9 9 0 0 0-18', bank: 'M3 9l9-5 9 5 M5 10v8 M9 10v8 M15 10v8 M19 10v8 M3 21h18', money: 'M4 6h16v12H4z M8 12h8 M7 9h.01 M17 15h.01', target: 'M12 3a9 9 0 1 0 9 9 M12 7a5 5 0 1 0 5 5 M12 11a1 1 0 1 0 1 1', box: 'M4 7l8-4 8 4v12l-8 4-8-4z M4 7l8 4 8-4 M12 11v12', flag: 'M5 21V4 M5 5h11l-2 4 2 4H5', leaf: 'M20 4C10 4 4 10 4 20c8-1 15-6 16-16z M4 20l8-8',
}

function Icon({ icon, className = '' }: { icon: MockupIcon; className?: string }) {
  return <svg className={`mocap-icon ${className}`} viewBox="0 0 24 24" aria-hidden="true"><path d={iconPaths[icon]} fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" /></svg>
}

type NavItem = { key: MockupScreenKey; label: string; icon: MockupIcon; section?: string; tooltip: TooltipId }
const navItems: NavItem[] = [
  { key: 'executive', label: 'Visão Executiva', icon: 'home', section: 'VISÃO EXECUTIVA', tooltip: 'nav.executive' },
  { key: 'cashflow', label: 'Fluxo de Caixa', icon: 'flow', section: 'TESOURARIA', tooltip: 'nav.treasury' },
  { key: 'receivables', label: 'Contas a Receber', icon: 'receivable', tooltip: 'nav.receivables' },
  { key: 'payables', label: 'Contas a Pagar', icon: 'payable', tooltip: 'nav.payables' },
  { key: 'dre', label: 'DRE', icon: 'result', section: 'RESULTADOS', tooltip: 'nav.result' },
  { key: 'trial-balance', label: 'Balancete', icon: 'balance', tooltip: 'nav.accounting' },
  { key: 'investment-calculator', label: 'Calculadora de Aplicação', icon: 'calculator', tooltip: 'nav.investments' },
  { key: 'inventory', label: 'Estoque', icon: 'inventory', section: 'OPERAÇÃO', tooltip: 'nav.treasury' },
  { key: 'production', label: 'Produção', icon: 'production', tooltip: 'nav.treasury' },
  { key: 'purchases', label: 'Compras', icon: 'shopping', tooltip: 'nav.treasury' },
  { key: 'pending', label: 'Lançamentos & Pendências', icon: 'control', section: 'CONTROLE', tooltip: 'nav.result' },
  { key: 'alerts', label: 'Alertas', icon: 'warning', tooltip: 'management.alerts' },
  { key: 'integrations', label: 'Importações / Integrações', icon: 'import', tooltip: 'nav.integrations' },
  { key: 'parameters', label: 'Parâmetros', icon: 'settings', section: 'CONFIGURAÇÕES', tooltip: 'settings.open' },
  { key: 'users', label: 'Usuários', icon: 'user', tooltip: 'settings.open' },
]

export function MockupSidebar({ screen, onNavigate }: { screen: string; onNavigate: (screen: MockupScreenKey) => void }) {
  return <aside className="mocap-sidebar"><div className="mocap-people" style={{ backgroundImage: `url(${sidebarReference})` }} aria-hidden="true" /><nav aria-label="Navegação principal" className="mocap-nav">{navItems.map((item, index) => <span key={`${item.label}-${index}`} className="mocap-nav-wrap">{item.section && <span className="mocap-nav-section">{item.section}</span>}<Tooltip id={item.tooltip}><button type="button" className={`mocap-nav-item ${screen === item.key ? 'active' : ''}`} onClick={() => onNavigate(item.key)} aria-current={screen === item.key ? 'page' : undefined} aria-label={item.label === 'Importações / Integrações' ? 'Integrações' : undefined}><Icon icon={item.icon} /><span>{item.label}</span></button></Tooltip></span>)}</nav></aside>
}

type Card = { label: string; icon: MockupIcon; tone: 'green' | 'blue' | 'orange' | 'red' | 'purple'; supporting?: string; action?: string; variant?: 'minimum-balance' | 'period-comparison'; value?: string; displayValue?: string; periodValues?: Array<{ label: string; metric?: DashboardMetric }>; minimumValues?: Array<{ label: string; metric?: DashboardMetric; date?: string; headroom?: DashboardMetric }> }

// Approved by D-003-A and recorded in DATA-DICTIONARY-v1.md.
const APPROVED_MINIMUM_RESERVE = '50000.00'
type ScreenLayout = 'standard' | 'critical' | 'calculation' | 'market' | 'forecast'
type ScreenConfig = { number: string; title: string; subtitle: string; cards: Card[]; middleTitle: string; middleSubtitle: string; rightTitle: string; rightSubtitle: string; lower: Card[]; alerts: string[]; table?: string[]; tableRows?: string[][]; tableEmptyMessage?: string; distributionEmptyMessage?: string; distributionItems?: string[]; alertSubtitles?: string[]; alertsTitle?: string; hideCardActions?: boolean; chart?: boolean; layout?: ScreenLayout }

const configs: Record<Exclude<MockupScreenKey, 'integrations'>, ScreenConfig> = {
  executive: { number: '01', title: 'VISÃO EXECUTIVA', subtitle: 'Panorama financeiro e operacional do negócio', cards: [{ label: 'SALDO ATUAL', icon: 'bank', tone: 'green', supporting: 'Banco conciliado' }, { label: 'A RECEBER', icon: 'receivable', tone: 'blue', supporting: 'B2C + B2B' }, { label: 'A PAGAR', icon: 'payable', tone: 'orange', supporting: 'Bling' }, { label: 'MENOR SALDO 30 / 60 DIAS', icon: 'trend', tone: 'blue', variant: 'minimum-balance' }], middleTitle: 'FLUXO DE CAIXA PROJETADO', middleSubtitle: 'Próximos 30 dias', rightTitle: 'DISTRIBUIÇÃO DO SALDO', rightSubtitle: 'Por conta / meio', lower: [{ label: 'VENDAS DO MÊS', icon: 'cart', tone: 'blue', supporting: 'Meta' }, { label: 'MARGEM BRUTA', icon: 'chart', tone: 'green', supporting: 'Meta' }, { label: 'ESTOQUE / COBERTURA', icon: 'box', tone: 'orange', supporting: 'Cacau, embalagens e acabados' }, { label: 'PRODUÇÃO PREVISTA', icon: 'production', tone: 'blue', supporting: 'Próximas 2 semanas' }, { label: 'COMPRAS PREVISTAS', icon: 'shopping', tone: 'purple', supporting: 'Próximos 15 dias' }], alerts: ['Saldo projetado abaixo da reserva mínima'], chart: true },
  cashflow: { number: '02', title: 'FLUXO DE CAIXA', subtitle: 'Realizado, previsto e projetado do caixa', cards: [{ label: 'SALDO ATUAL', icon: 'bank', tone: 'green', supporting: 'Banco conciliado' }, { label: 'ENTRADAS', icon: 'trend', tone: 'green', supporting: 'Realizado + previsto', variant: 'period-comparison' }, { label: 'SAÍDAS', icon: 'trend', tone: 'orange', supporting: 'Compromissos e pagamentos', variant: 'period-comparison' }, { label: 'MENOR SALDO 30 / 60 DIAS', icon: 'trend', tone: 'blue', variant: 'minimum-balance' }], middleTitle: 'FLUXO DE CAIXA CONSOLIDADO', middleSubtitle: 'Próximos 30 dias', rightTitle: 'SEMANA CRÍTICA', rightSubtitle: 'Maior total de despesas / compromissos', lower: [{ label: 'RECEBIDO NO MÊS', icon: 'cart', tone: 'blue', supporting: 'Até a data-base' }, { label: 'PAGO NO MÊS', icon: 'document', tone: 'red', supporting: 'Até a data-base' }, { label: 'RESERVA MÍNIMA', icon: 'wallet', tone: 'orange', supporting: 'Parametrizada' }, { label: 'SALDO TRANSFERÍVEL', icon: 'flow', tone: 'blue', supporting: 'Aplicação possível' }, { label: 'NÍVEL DE CAIXA', icon: 'chart', tone: 'blue', supporting: 'Acima da reserva' }], alerts: ['Saldo projetado abaixo da reserva mínima'], chart: true, layout: 'critical' },
  receivables: { number: '03', title: 'CONTAS A RECEBER', subtitle: 'Recebíveis, atrasos e conciliação com Nuvem Pago', cards: [{ label: 'TOTAL A RECEBER', icon: 'bank', tone: 'green', supporting: 'Clientes e boletos' }, { label: 'RECEBIDO NO MÊS', icon: 'trend', tone: 'green', supporting: 'Até a data-base' }, { label: 'EM ATRASO', icon: 'clock', tone: 'orange', supporting: 'Recebimentos' }, { label: 'CONCILIAÇÃO', icon: 'refresh', tone: 'blue', supporting: 'Bling × Nuvem Pago' }], middleTitle: 'RECEBÍVEIS', middleSubtitle: 'Títulos em aberto e previstos', rightTitle: 'DISTRIBUIÇÃO DOS RECEBÍVEIS', rightSubtitle: 'Por status', lower: [{ label: 'PIX', icon: 'money', tone: 'blue', supporting: 'Liquidação rápida' }, { label: 'BOLETOS', icon: 'document', tone: 'green', supporting: 'Títulos' }, { label: 'CARTÃO', icon: 'money', tone: 'blue', supporting: 'Nuvem Pago' }, { label: 'AMBÍGUOS', icon: 'warning', tone: 'orange', supporting: 'Não conciliar automaticamente' }, { label: 'CHARGEBACKS', icon: 'refresh', tone: 'purple', supporting: 'Vinculados à origem' }], alerts: [], table: ['Cliente', 'Origem', 'Vencimento', 'Valor bruto', 'Valor líquido', 'Status'] },
  payables: { number: '04', title: 'CONTAS A PAGAR', subtitle: 'Obrigações, vencimentos e divergências', cards: [{ label: 'TOTAL A PAGAR', icon: 'document', tone: 'orange', supporting: 'Bling' }, { label: 'VENCE HOJE', icon: 'calendar', tone: 'blue', supporting: 'Contas' }, { label: 'EM ATRASO', icon: 'clock', tone: 'orange', supporting: 'Obrigações' }, { label: 'DIVERGÊNCIAS', icon: 'refresh', tone: 'purple', supporting: 'Diferença ≠ pagamento parcial' }], middleTitle: 'OBRIGAÇÕES', middleSubtitle: 'Títulos a pagar e vencimentos', rightTitle: 'DISTRIBUIÇÃO A PAGAR', rightSubtitle: 'Por status', lower: [{ label: 'FORNECEDORES', icon: 'user', tone: 'blue', supporting: 'Ativos' }, { label: 'FRETES', icon: 'shopping', tone: 'green', supporting: 'Etiquetas e NF' }, { label: 'IMPOSTOS', icon: 'document', tone: 'blue', supporting: 'Competência' }, { label: 'SEM PAGTO PARCIAL', icon: 'warning', tone: 'orange', supporting: 'Regra' }, { label: 'VENCIMENTOS 7 DIAS', icon: 'calendar', tone: 'purple', supporting: 'Prioridade' }], alerts: ['Contas a pagar vencendo hoje', 'Obrigações em atraso', 'Títulos com divergência de valor', 'Pagamentos parciais não permitidos'], table: ['Fornecedor', 'Documento', 'Vencimento', 'Valor', 'Categoria', 'Status'] },
  dre: { number: '06', title: 'DRE', subtitle: 'Demonstrativo do resultado do exercício', cards: [{ label: 'RECEITA BRUTA', icon: 'chart', tone: 'green', supporting: 'Mês atual' }, { label: 'RECEITA LÍQUIDA', icon: 'trend', tone: 'blue', supporting: 'Após deduções' }, { label: 'LUCRO BRUTO', icon: 'box', tone: 'orange', supporting: 'Margem' }, { label: 'RESULTADO LÍQUIDO', icon: 'database', tone: 'green', supporting: 'DRE Majucau' }], middleTitle: 'DRE CONSOLIDADA', middleSubtitle: 'Competência selecionada', rightTitle: 'COMPOSIÇÃO DO RESULTADO', rightSubtitle: 'Principais blocos', lower: [{ label: 'CMV / CPV', icon: 'box', tone: 'green', supporting: 'Base Bling' }, { label: 'DESPESAS', icon: 'chart', tone: 'green', supporting: 'Operacionais' }, { label: 'PERDAS', icon: 'warning', tone: 'orange', supporting: 'Sem dupla contagem' }, { label: 'TRIBUTOS', icon: 'document', tone: 'blue', supporting: 'Competência' }, { label: 'UNMAPPED', icon: 'document', tone: 'purple', supporting: 'Lançamentos pendentes' }], alerts: ['Lançamentos UNMAPPED na DRE', 'Perdas com competência pendente', 'Resultado financeiro em revisão', 'Sem conta automática de fechamento'], table: ['Descrição', 'Valor (R$)'] },
  'trial-balance': { number: '07', title: 'BALANCETE', subtitle: 'Débitos, créditos e saldos com rastreabilidade', cards: [{ label: 'DÉBITOS', icon: 'bank', tone: 'green', supporting: 'Período' }, { label: 'CRÉDITOS', icon: 'trend', tone: 'blue', supporting: 'Período' }, { label: 'SALDO', icon: 'document', tone: 'orange', supporting: 'Sem conta de ajuste' }, { label: 'DIVERGÊNCIAS', icon: 'warning', tone: 'red', supporting: 'Revisão necessária' }], middleTitle: 'LANÇAMENTOS DO BALANCETE', middleSubtitle: 'Competência selecionada', rightTitle: 'COMPOSIÇÃO DOS SALDOS', rightSubtitle: 'Por grupo contábil', lower: [{ label: 'LANÇAMENTOS', icon: 'cart', tone: 'blue', supporting: 'No período' }, { label: 'CONTRAPARTIDAS', icon: 'chart', tone: 'green', supporting: 'Rastreáveis' }, { label: 'SALDOS INICIAIS', icon: 'box', tone: 'orange', supporting: 'Pendente' }, { label: 'SEM AJUSTE', icon: 'reconcile', tone: 'blue', supporting: 'Sem conta automática' }, { label: 'ORIGEM', icon: 'control', tone: 'purple', supporting: 'Bling + Majucau' }], alerts: ['Divergências de saldo / contrapartida', 'Saldos iniciais permanecem abertos', 'Ausência de saldo inicial não vira zero', 'Sem conta automática para fechar diferença'], table: ['Conta', 'Descrição', 'Débito', 'Crédito', 'Saldo', 'Status'] },
  'investment-calculator': { number: '08', title: 'CALCULADORA DE APLICAÇÃO', subtitle: 'Capacidade de investimento e regras de segurança', cards: [{ label: 'LUCRO ELEGÍVEL', icon: 'chart', tone: 'green', supporting: 'DRE Majucau' }, { label: 'CAIXA CONSOLIDADO', icon: 'bank', tone: 'blue', supporting: 'Posição atual' }, { label: 'RESERVA MÍNIMA', icon: 'wallet', tone: 'orange', supporting: 'Parametrizada' }, { label: 'CAPACIDADE DE APLICAÇÃO', icon: 'trend', tone: 'blue', supporting: 'Saldo transferível' }], middleTitle: 'CÁLCULO DE CAPACIDADE', middleSubtitle: 'Fórmula auditável', rightTitle: 'REGRAS DA DECISÃO', rightSubtitle: 'Critérios vigentes', lower: [{ label: 'CDI PLUS APLICADO', icon: 'chart', tone: 'green', supporting: 'Histórico' }, { label: 'SALDO TRANSFERÍVEL', icon: 'flow', tone: 'blue', supporting: 'Disponível' }, { label: 'SEMANA CRÍTICA', icon: 'clock', tone: 'orange', supporting: 'Diagnóstico' }, { label: 'PMR', icon: 'chart', tone: 'purple', supporting: 'Diagnóstico' }, { label: 'CICLO FINANCEIRO', icon: 'calendar', tone: 'blue', supporting: 'Indicador' }], alerts: ['Posição de caixa da semana crítica permanece aberta', 'Reserva mínima ativa', 'PMR / PMP / PME não alteram automaticamente a aplicação', 'Validação final depende das fontes reais'], table: ['Componente', 'Valor'], layout: 'calculation' },
  inventory: { number: '09', title: 'ESTOQUE', subtitle: 'Estoque realizado, utilizável e cobertura', cards: [{ label: 'ESTOQUE FÍSICO', icon: 'box', tone: 'green', supporting: 'Matérias-primas + acabados' }, { label: 'UTILIZÁVEL', icon: 'chart', tone: 'blue', supporting: 'Descontando reservas e bloqueios' }, { label: 'COBERTURA', icon: 'clock', tone: 'orange', supporting: 'Média atual' }, { label: 'RISCO DE RUPTURA', icon: 'warning', tone: 'red', supporting: 'Abaixo do mínimo' }], middleTitle: 'ESTOQUE POR ITEM', middleSubtitle: 'Saldo e cobertura', rightTitle: 'DISTRIBUIÇÃO DO ESTOQUE', rightSubtitle: 'Por grupo', lower: [{ label: 'SEGURANÇA', icon: 'control', tone: 'blue', supporting: 'Estoque mínimo' }, { label: 'PONTO DE PEDIDO', icon: 'cart', tone: 'green', supporting: 'Monitorados' }, { label: 'GIRO', icon: 'chart', tone: 'purple', supporting: 'Últimos 12 meses' }], alerts: ['SKUs abaixo do estoque mínimo', 'Itens com cobertura inferior a 7 dias', 'Estoque projetado depende do forecast'], table: ['SKU', 'Descrição', 'Físico', 'Reservado', 'Utilizável', 'Cobertura', 'Status'] },
  production: { number: '10', title: 'PRODUÇÃO', subtitle: 'Planejado, em andamento e realizado', cards: [{ label: 'PLANEJADA', icon: 'production', tone: 'blue', supporting: 'Próximas 2 semanas' }, { label: 'EM ANDAMENTO', icon: 'settings', tone: 'green', supporting: 'Lotes liberados' }, { label: 'CONCLUÍDA', icon: 'flag', tone: 'orange', supporting: 'Período atual' } as Card, { label: 'YIELD MÉDIO', icon: 'trend', tone: 'purple', supporting: 'Com perdas controladas' }], middleTitle: 'ORDENS DE PRODUÇÃO', middleSubtitle: 'Planejado x realizado', rightTitle: 'CAPACIDADE E STATUS', rightSubtitle: 'Distribuição das ordens', lower: [{ label: 'CAPACIDADE', icon: 'box', tone: 'blue', supporting: '2 semanas' }, { label: 'LOTES', icon: 'document', tone: 'green', supporting: 'Ativos' }, { label: 'PERDAS', icon: 'warning', tone: 'orange', supporting: 'No período' }, { label: 'RENDIMENTO', icon: 'chart', tone: 'green', supporting: 'Yield médio' }, { label: 'SKU PRODUZIDO', icon: 'box', tone: 'purple', supporting: 'No mês' }], alerts: ['Ordens em atraso', 'BOM versionada impacta a necessidade de insumos'], table: ['Ordem', 'SKU', 'Planejado', 'Em andamento', 'Realizado', 'Status'] },
  purchases: { number: '11', title: 'COMPRAS', subtitle: 'Compras sugeridas, insumos, cacau e fornecedores', cards: [{ label: 'COMPRAS INDICADAS — HOJE', icon: 'shopping', tone: 'purple', supporting: 'Com base na demanda, estoque e lead time' }, { label: 'NECESSIDADE LÍQUIDA TOTAL', icon: 'chart', tone: 'green', supporting: 'Próximos 60 dias' }, { label: 'PEDIDOS CONFIRMADOS', icon: 'document', tone: 'orange', supporting: 'Em trânsito / a receber' }, { label: 'LEAD TIME MÉDIO', icon: 'clock', tone: 'blue', supporting: 'Fornecedores principais' }], middleTitle: 'CACAU — MERCADO E RISCO DE ABASTECIMENTO', middleSubtitle: 'Cotações, clima, safra e câmbio', rightTitle: 'RISCO DE ABASTECIMENTO (CACAU)', rightSubtitle: 'Alto, médio e baixo', lower: [{ label: 'NECESSIDADE DE COMPRA (TONELADAS)', icon: 'chart', tone: 'green', supporting: 'Necessidade líquida' }, { label: 'ORIGEM DA NECESSIDADE', icon: 'leaf', tone: 'green', supporting: 'Produção e estoque' }, { label: 'ALERTAS IMPORTANTES', icon: 'warning', tone: 'orange', supporting: 'Risco climático e cobertura' } as Card], alerts: ['Preço do cacau em queda', 'Risco climático elevado', 'Cobertura de manteiga de cacau inferior a 20 dias', 'Fornecedores com lead time acima da média'], table: ['Insumo', 'Categoria', 'Estoque atual', 'Nec. líquida', 'Qtd. sugerida', 'Preço referência', 'Risco'], layout: 'market' },
  forecast: { number: '12', title: 'PROJEÇÕES', subtitle: 'Forecast, cenários e backtest da demanda', cards: [{ label: 'FORECAST BASE', icon: 'chart', tone: 'green', supporting: 'Próximos 60 dias' }, { label: 'CENÁRIO PESSIMISTA', icon: 'trend', tone: 'red', supporting: 'Variação' }, { label: 'CENÁRIO OTIMISTA', icon: 'trend', tone: 'green', supporting: 'Variação' }, { label: 'WAPE', icon: 'target', tone: 'blue', supporting: 'Acurácia do backtest' }], middleTitle: 'FORECAST x REALIZADO', middleSubtitle: 'Demanda, receita e cenários', rightTitle: 'VERSÃO E BACKTEST', rightSubtitle: 'Metadados da versão ativa', lower: [{ label: 'DEMANDA PREVISTA', icon: 'cart', tone: 'blue', supporting: 'Próximo período' }, { label: 'COBERTURA PROJETADA', icon: 'chart', tone: 'green', supporting: 'Dias' }, { label: 'PRODUÇÃO SUGERIDA', icon: 'box', tone: 'orange', supporting: 'Toneladas' }, { label: 'COMPRA DERIVADA', icon: 'shopping', tone: 'purple', supporting: 'Valor estimado' }, { label: 'SAZONALIDADE', icon: 'trend', tone: 'blue', supporting: 'Páscoa / Natal' }], alerts: ['Forecast preserva a versão original', 'SKUs com backtest acima do limite', 'Promoções afetam a leitura de demanda', 'Ruptura não vira baixa automática'], table: ['SKU', 'Canal', 'Forecast', 'Realizado', 'Desvio', 'Cobertura projetada', 'Status'], chart: true, layout: 'forecast' },
  pending: { number: '13', title: 'LANÇAMENTOS & PENDÊNCIAS', subtitle: 'Exceções, divergências e decisões em aberto', cards: [{ label: 'PENDÊNCIAS ABERTAS', icon: 'warning', tone: 'red', supporting: 'Itens aguardando decisão' }, { label: 'DIVERGÊNCIAS', icon: 'document', tone: 'orange', supporting: 'Registros com divergência' }, { label: 'OPEN / BLOCKED', icon: 'control', tone: 'purple', supporting: 'Itens bloqueados no fluxo' }, { label: 'UNMAPPED', icon: 'import', tone: 'blue', supporting: 'Lançamentos sem mapeamento' }], middleTitle: 'ITENS PENDENTES', middleSubtitle: 'Backlog operacional e financeiro', rightTitle: 'DISTRIBUIÇÃO DAS PENDÊNCIAS', rightSubtitle: 'Por status', lower: [{ label: 'SEMANA CRÍTICA OPEN', icon: 'calendar', tone: 'red', supporting: 'Semana sem definição' }, { label: 'SALDOS INICIAIS OPEN', icon: 'chart', tone: 'green', supporting: 'Contas sem saldo inicial' }, { label: 'PERDAS SEM COMPETÊNCIA', icon: 'document', tone: 'orange', supporting: 'Lançamentos em análise' }, { label: 'TAXA / JUROS NUVEM', icon: 'money', tone: 'blue', supporting: 'Parâmetros pendentes' }, { label: 'MATCH AMBÍGUO', icon: 'refresh', tone: 'purple', supporting: 'Aguardando análise' }], alerts: ['DEV não decide sozinho os itens OPEN/BLOCKED', 'Ausência de dado não vira zero automaticamente', 'Divergência não significa pagamento parcial', 'Regras pendentes permanecem pendentes de decisão'], table: ['Módulo', 'Origem', 'Descrição', 'Status', 'Responsável', 'Ação'] },
  alerts: { number: '14', title: 'ALERTAS', subtitle: 'Central consolidada de exceções e monitoramento', cards: [{ label: 'CRÍTICOS', icon: 'warning', tone: 'red', supporting: 'Abertos' }, { label: 'ATENÇÃO', icon: 'warning', tone: 'orange', supporting: 'Abertos' }, { label: 'INFORMATIVOS', icon: 'help', tone: 'blue', supporting: 'Abertos' }, { label: 'RESOLVIDOS 7 DIAS', icon: 'reconcile', tone: 'green', supporting: 'Histórico' }], middleTitle: 'ALERTAS CONSOLIDADOS', middleSubtitle: 'Todos os módulos', rightTitle: 'DISTRIBUIÇÃO DOS ALERTAS', rightSubtitle: 'Por módulo', lower: [{ label: 'CAIXA', icon: 'bank', tone: 'green', supporting: 'Alertas' }, { label: 'RECEBÍVEIS', icon: 'receivable', tone: 'blue', supporting: 'Alertas' }, { label: 'ESTOQUE', icon: 'box', tone: 'orange', supporting: 'Alertas' }, { label: 'PRODUÇÃO', icon: 'production', tone: 'blue', supporting: 'Alertas' }, { label: 'COMPRAS', icon: 'shopping', tone: 'purple', supporting: 'Alertas' }], alerts: ['Caixa previsto rompe reserva mínima', 'Estoque crítico afeta forecast e produção', 'Alertas sem responsável definido', 'Atualização das fontes requerida para remover STALE'], table: ['Módulo', 'Alerta', 'Impacto', 'Prioridade', 'Atualizado', 'Status'] },
  parameters: { number: '16', title: 'PARÂMETROS', subtitle: 'Regras aprovadas, fontes oficiais e parâmetros do sistema', cards: [{ label: 'PROJEÇÕES', icon: 'chart', tone: 'red', displayValue: '24 meses', supporting: 'Histórico aprovado' }, { label: 'HORIZONTE', icon: 'calendar', tone: 'orange', displayValue: '12 meses', supporting: 'Forecast aprovado' }, { label: 'INVESTIMENTOS', icon: 'database', tone: 'blue', supporting: 'Dados não confirmados' }, { label: 'INTEGRAÇÕES', icon: 'flow', tone: 'green', supporting: 'Estado em Integrações' }], middleTitle: 'PARÂMETROS VIGENTES', middleSubtitle: 'Regras aprovadas e estado das fontes', rightTitle: 'MAPA DE PARÂMETROS', rightSubtitle: 'Distribuição por grupo', distributionItems: ['Projeções', 'Investimentos', 'Estoque', 'Compras', 'Integrações'], lower: [{ label: 'FONTE OFICIAL', icon: 'flow', tone: 'green', displayValue: 'Bling', supporting: 'Produtos, vendas e estoque' }, { label: 'PROJEÇÃO', icon: 'box', tone: 'blue', displayValue: 'SKU', supporting: 'Sem família' }, { label: 'META DE PRODUÇÃO', icon: 'settings', tone: 'orange', displayValue: 'Por SKU', supporting: 'Tela 10' }, { label: 'CENÁRIO META', icon: 'chart', tone: 'purple', displayValue: 'Faturamento / SKU', supporting: 'Tela 12' }, { label: 'EXTRATOS', icon: 'document', tone: 'blue', supporting: 'Fonte externa pendente' }], alerts: ['Design aprovado permanece imutável', 'Estoque pronto não é descontado da produção projetada', 'T08 é consulta e não oferece exportação', 'Forecast considera sazonalidade do chocolate'], alertSubtitles: ['Somente dados e conteúdos aprovados podem mudar.', 'Usar apenas a regra da Tela 10.', 'A memória de cálculo permanece disponível na Tela 08.', 'Páscoa, Natal e outros períodos críticos.'], alertsTitle: 'REGRAS IMPORTANTES', table: ['Grupo', 'Parâmetro', 'Regra aprovada', 'Origem', 'Estado'], tableRows: [['Projeções', 'Histórico do forecast', '24 meses', 'Majucau', 'Aprovado'], ['Projeções', 'Horizonte projetado', '12 meses', 'Tela 12', 'Aprovado'], ['Projeções', 'Métricas de qualidade', 'WAPE / Bias / MASE', 'Tela 12', 'Limiares pendentes'], ['Estoque', 'Estoque mínimo', 'Por SKU', 'Bling', 'Fonte pendente'], ['Compras', 'Recomendação', 'Saldo ≤ mínimo Bling e sem pedido aberto', 'Telas 09 e 11', 'Fonte pendente'], ['Investimentos', 'Categoria de aporte', 'Transferência', 'Bling', 'Fonte pendente'], ['Investimentos', 'Categoria de rendimento', 'Rendimentos', 'Bling', 'Fonte pendente'], ['Integrações', 'Recebimentos futuros', 'Arquivo aprovado', 'Drive', 'Fonte pendente'], ['Integrações', 'Perdas operacionais', 'Arquivo aprovado', 'Drive', 'Fonte pendente']], hideCardActions: true },
  users: { number: '17', title: 'USUÁRIOS', subtitle: 'Consulta de acessos e perfis do sistema', cards: [{ label: 'USUÁRIOS ATIVOS', icon: 'user', tone: 'green', supporting: 'Aguardando consulta segura' }, { label: 'PERFIS', icon: 'target', tone: 'orange', displayValue: '2 tipos', supporting: 'Admin / Diretor' }, { label: 'BLOQUEADOS', icon: 'warning', tone: 'red', supporting: 'Sem dados confirmados' }, { label: 'PENDENTES', icon: 'clock', tone: 'blue', supporting: 'Sem dados confirmados' }], middleTitle: 'USUÁRIOS E ACESSOS', middleSubtitle: 'Consulta administrativa', rightTitle: 'DISTRIBUIÇÃO DE PERFIS', rightSubtitle: 'Por tipo de acesso', distributionEmptyMessage: 'Sem dados de usuários confirmados', distributionItems: ['ADMIN', 'DIRETORES'], lower: [{ label: 'ADMIN', icon: 'user', tone: 'green', supporting: 'Conta financeira aprovada' }, { label: 'DIRETORES', icon: 'document', tone: 'blue', supporting: 'Conta compartilhada aprovada' }, { label: 'BLOQUEADOS', icon: 'warning', tone: 'orange', supporting: 'Estado não consultável' }, { label: 'PENDENTES', icon: 'clock', tone: 'purple', supporting: 'Estado não consultável' }, { label: 'ALTERAÇÕES', icon: 'refresh', tone: 'red', supporting: 'Bloqueadas até política aprovada' }], alerts: ['Permissões seguem o perfil', 'Ações administrativas precisam de auditoria', 'Conta bloqueada deve ter acesso negado', 'Convites e mudanças de conta permanecem bloqueados'], alertSubtitles: ['A autorização deve ser aplicada no backend.', 'A trilha precisa identificar ator, data e ação.', 'A tela não substitui a verificação do servidor.', 'A política de identidade ainda não foi aprovada.'], alertsTitle: 'REGRAS IMPORTANTES', table: ['Usuário', 'Perfil', 'Login', 'Último acesso', 'Status', 'Ação'], tableEmptyMessage: 'Consulta segura de usuários ainda não está disponível. Nenhuma conta será presumida.', hideCardActions: true },
}

function EmptyValue() { return <span className="mocap-empty-value">—</span> }
function formatSnapshotDate(value?: string): string {
  if (!value) return '—'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '—'
  return new Intl.DateTimeFormat('pt-BR', { timeZone: 'America/Sao_Paulo', day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false }).format(date)
}

function MocapHeader({ config, snapshot }: { config: ScreenConfig; snapshot?: DashboardSnapshot }) {
  return <header className="mocap-header"><div className="mocap-heading"><span className="mocap-number">{config.number}</span><div><h1 aria-label={config.title === 'VISÃO EXECUTIVA' ? 'Visão Executiva' : config.title}>{config.title}</h1><p>{config.subtitle}</p></div></div><div className="mocap-header-actions"><div className="mocap-sync"><span>Posição: <strong>—</strong></span><span>Última sincronização: <strong>{formatSnapshotDate(snapshot?.last_bling_sync_at)}</strong> <Icon icon="refresh" /></span></div></div></header>
}

function MocapCard({ card, detailScreen, onNavigate, hideAction = false }: { card: Card; detailScreen?: MockupScreenKey; onNavigate?: (screen: MockupScreenKey) => void; hideAction?: boolean }) {
  const formatPeriodMetric = (metric?: DashboardMetric) => !metric?.value || metric.state === 'UNAVAILABLE' ? '—' : `${formatBRL(metric.value)}${metric.state === 'PARTIAL' ? ' · Parcial' : ''}`
  const negative = (value?: string) => /^\s*[−-]/.test(value ?? '')
  const formatBalanceDate = (value?: string) => value && /^\d{4}-\d{2}-\d{2}$/.test(value) ? new Intl.DateTimeFormat('pt-BR', { timeZone: 'UTC', day: '2-digit', month: '2-digit' }).format(new Date(`${value}T12:00:00.000Z`)) : '—'
  return <article className={`mocap-card tone-${card.tone} ${card.variant ? `variant-${card.variant}` : ''}`}><div className="mocap-card-icon"><Icon icon={card.icon} /></div><div className="mocap-card-body"><Tooltip id="metric.details"><span className="mocap-card-label">{card.label}</span></Tooltip>{card.variant === 'minimum-balance' ? <div className="mocap-minimum-values">{(card.minimumValues ?? [{ label: '30 dias' }, { label: '60 dias' }]).map((item) => <div key={item.label}><strong>{item.label}</strong><b className={negative(item.metric?.value) ? 'mocap-negative-value' : undefined}>{formatPeriodMetric(item.metric)}</b><span>Em: {formatBalanceDate(item.date)}</span><small className={negative(item.headroom?.value) ? 'mocap-negative-value' : undefined}>Folga: {formatPeriodMetric(item.headroom)}</small></div>)}</div> : card.variant === 'period-comparison' ? <div className="mocap-period-values">{card.periodValues?.map((period) => <div key={period.label}><span>{period.label}</span><b className={negative(period.metric?.value) ? 'mocap-negative-value' : undefined}>{formatPeriodMetric(period.metric)}</b></div>)}</div> : <><span className={`mocap-empty-value${negative(card.value) || negative(card.displayValue) ? ' mocap-negative-value' : ''}`}>{card.displayValue ?? (card.value ? formatBRL(card.value) : '—')}</span>{card.supporting && <span className="mocap-card-support">{card.supporting}</span>}</>}{!hideAction && <Tooltip id="metric.details"><button type="button" className="mocap-card-action" aria-label={detailScreen ? `Ver detalhes: ${card.label}` : undefined} disabled={Boolean(detailScreen) && !onNavigate} onClick={detailScreen && onNavigate ? () => onNavigate(detailScreen) : undefined}>Ver detalhes <span>⌄</span></button></Tooltip>}</div></article>
}

function displayMoney(value?: string): string {
  return value && value.trim() !== '' ? formatBRL(value) : '—'
}

function receivableTableRow(row: DashboardReceivableRow): string[] {
  return [row.customer || '—', row.origin, row.due_date || '—', displayMoney(row.gross_value), displayMoney(row.net_value), row.status]
}

function payableTableRow(row: DashboardPayableRow): string[] {
  return [row.supplier || '—', row.document || '—', row.due_date, displayMoney(row.value), row.category || '—', row.status]
}

function cashFlowDayDate(value: string): Date {
  return new Date(`${value}T00:00:00.000Z`)
}

function sumCashFlowMetrics(metrics: DashboardMetric[]): DashboardMetric {
  if (metrics.length === 0 || metrics.some((metric) => !metric.value || metric.state === 'UNAVAILABLE')) {
    return { state: 'UNAVAILABLE', source_system: 'BLING' }
  }
  let units = 0n
  for (const metric of metrics) {
    const match = /^(-?)(\d{1,15})(?:\.(\d{1,4}))?$/.exec(metric.value ?? '')
    if (!match) return { state: 'UNAVAILABLE', source_system: 'BLING' }
    const amount = BigInt(match[2]) * 10_000n + BigInt((match[3] ?? '').padEnd(4, '0'))
    units += match[1] ? -amount : amount
  }
  const negative = units < 0n
  const absolute = negative ? -units : units
  const value = `${negative ? '-' : ''}${absolute / 10_000n}.${(absolute % 10_000n).toString().padStart(4, '0')}`
  return { value, state: metrics.some((metric) => metric.state === 'PARTIAL') ? 'PARTIAL' : 'CONFIRMED', source_system: 'BLING' }
}

function calculateClosingBalance(opening: DashboardMetric | undefined, inflows: DashboardMetric, outflows: DashboardMetric): DashboardMetric {
  const parseUnits = (value?: string) => {
    const match = /^(-?)(\d{1,15})(?:\.(\d{1,4}))?$/.exec(value ?? '')
    if (!match) return undefined
    const amount = BigInt(match[2]) * 10_000n + BigInt((match[3] ?? '').padEnd(4, '0'))
    return match[1] ? -amount : amount
  }
  const openingUnits = parseUnits(opening?.value)
  const inflowUnits = parseUnits(inflows.value)
  const outflowUnits = parseUnits(outflows.value)
  if (openingUnits === undefined || inflowUnits === undefined || outflowUnits === undefined || opening?.state === 'UNAVAILABLE' || inflows.state === 'UNAVAILABLE' || outflows.state === 'UNAVAILABLE') {
    return { state: 'UNAVAILABLE', source_system: 'BLING' }
  }
  const expenseUnits = outflowUnits < 0n ? -outflowUnits : outflowUnits
  const closingUnits = openingUnits + inflowUnits - expenseUnits
  const absolute = closingUnits < 0n ? -closingUnits : closingUnits
  return {
    value: `${closingUnits < 0n ? '-' : ''}${absolute / 10_000n}.${(absolute % 10_000n).toString().padStart(4, '0')}`,
    state: [opening, inflows, outflows].some((metric) => metric?.state === 'PARTIAL') ? 'PARTIAL' : 'CONFIRMED',
    source_system: 'BLING',
  }
}

function groupSelectedCashFlowDays(cashFlow: DashboardSnapshot['cash_flow'], from: string, through: string): DashboardCashFlowPeriod[] {
  if (!cashFlow?.days || !from || !through || from > through) return []
  const selected = cashFlow.days.filter((day) => day.date >= from && day.date <= through)
  const result: DashboardCashFlowPeriod[] = []
  let days: typeof selected = []
  let previousClosing: DashboardMetric | undefined
  const publishWeek = () => {
    if (days.length === 0) return
    const start = days[0].date
    const end = days[days.length - 1].date
    const sourcePeriod = cashFlow.periods?.find((period) => period.from === start && period.through === end)
    const inflows = sumCashFlowMetrics(days.map((day) => day.inflows))
    const outflows = sumCashFlowMetrics(days.map((day) => day.outflows))
    const opening = result.length > 0 ? previousClosing : sourcePeriod?.opening_balance
    const closing = calculateClosingBalance(opening, inflows, outflows)
    result.push({
      from: start,
      through: end,
      inflows,
      outflows,
      ...(opening ? { opening_balance: opening } : {}),
      closing_balance: closing,
    })
    previousClosing = closing.state === 'UNAVAILABLE' ? undefined : closing
    days = []
  }
  for (let index = 0; index < selected.length; index += 1) {
    const day = selected[index]
    days.push(day)
    const weekday = cashFlowDayDate(day.date).getUTCDay()
    if (weekday === 6 || index === selected.length - 1) publishWeek()
  }
  return result
}

function availableCashFlowDate(cashFlow: DashboardSnapshot['cash_flow'], snapshot?: DashboardSnapshot, offsetDays = 0): string {
  const base = cashFlow?.days?.[0]?.date ?? snapshot?.as_of.slice(0, 10)
  if (!base) return ''
  const date = cashFlowDayDate(base)
  date.setUTCDate(date.getUTCDate() + offsetDays)
  return date.toISOString().slice(0, 10)
}

function saoPauloToday(): string {
  const parts = new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Sao_Paulo', year: 'numeric', month: '2-digit', day: '2-digit' }).formatToParts(new Date())
  const part = (type: Intl.DateTimeFormatPartTypes) => parts.find((value) => value.type === type)?.value ?? ''
  return `${part('year')}-${part('month')}-${part('day')}`
}

function CashFlowDateControls({ from, through, min, max, onFromChange, onThroughChange }: { from: string; through: string; min: string; max: string; onFromChange: (value: string) => void; onThroughChange: (value: string) => void }) {
  return <section className="mocap-period-controls" aria-label="Período do fluxo de caixa"><h2>Período do relatório</h2><label>De<input aria-label="Data inicial do fluxo de caixa" type="date" min={min} max={max} value={from} onChange={(event) => onFromChange(event.currentTarget.value)} /></label><label>Até<input aria-label="Data final do fluxo de caixa" type="date" min={min} max={max} value={through} onChange={(event) => onThroughChange(event.currentTarget.value)} /></label><span>A tabela, o gráfico e a semana crítica acompanham o período selecionado.</span></section>
}

function CashFlowWeeklyTable({ periods }: { periods: DashboardCashFlowPeriod[] }) {
  const notAvailable: DashboardMetric = { state: 'UNAVAILABLE', source_system: 'BLING' }
  const rows: Array<{ label: string; metrics: Array<DashboardMetric | string>; total?: DashboardMetric | string }> = [
    { label: 'Saldo inicial', metrics: periods.map((period) => period.opening_balance ?? notAvailable), total: periods[0]?.opening_balance ?? notAvailable },
    { label: 'Receitas', metrics: periods.map((period) => period.inflows), total: sumCashFlowMetrics(periods.map((period) => period.inflows)) },
    { label: 'Despesas', metrics: periods.map((period) => period.outflows), total: sumCashFlowMetrics(periods.map((period) => period.outflows)) },
    { label: 'Saldo acumulado', metrics: periods.map((period) => period.closing_balance ?? notAvailable), total: periods[periods.length - 1]?.closing_balance ?? notAvailable },
  ]
  const metricText = (metric?: DashboardMetric | string) => typeof metric === 'string' ? metric : !metric?.value || metric.state === 'UNAVAILABLE' ? '—' : `${formatBRL(metric.value)}${metric.state === 'PARTIAL' ? ' · Parcial' : ''}`
  const isNegative = (metric?: DashboardMetric | string) => {
    const value = typeof metric === 'string' ? metric : metric?.value
    return /^\s*[−-]/.test(value ?? '')
  }
  const dateLabel = (value: string) => {
    const [year, month, day] = value.split('-')
    return `${day}/${month}/${year}`
  }
  return <section className="mocap-panel mocap-cashflow-table" aria-label="Tabela semanal do fluxo de caixa"><div className="mocap-panel-heading"><div><Icon icon="calendar" /><div><h2>FLUXO DE CAIXA POR SEMANA</h2><p>Entradas e saídas da API do Bling, agrupadas no período selecionado</p></div></div></div><div className="mocap-table-scroll"><table><thead><tr><th>Período</th>{periods.map((period) => <th key={`${period.from}-${period.through}`}>{dateLabel(period.from)} a<br />{dateLabel(period.through)}</th>)}<th>Total</th></tr></thead><tbody>{rows.map((row) => <tr key={row.label}><th scope="row">{row.label}</th>{row.metrics.map((metric, index) => <td key={`${row.label}-${periods[index]?.from ?? index}`} className={isNegative(metric) ? 'mocap-negative-value' : undefined}>{metricText(metric)}</td>)}<td className={isNegative(row.total) ? 'mocap-negative-value' : undefined}>{metricText(row.total)}</td></tr>)}</tbody></table></div><p className="mocap-cashflow-table-note">Se houver mais semanas, use a barra inferior para vê-las. Saldos sem confirmação oficial permanecem como “—”. O saldo final é calculado por saldo inicial + receitas − despesas.</p></section>
}

function metricForCard(screen: MockupScreenKey, label: string, snapshot?: DashboardSnapshot): DashboardMetric | undefined {
  if (!snapshot) return undefined
  if (screen === 'cashflow' && label === 'RESERVA MÍNIMA') return { value: APPROVED_MINIMUM_RESERVE, state: 'CONFIRMED', source_system: 'MAJUCAU_RULE' }
  if ((screen === 'executive' || screen === 'cashflow') && label === 'SALDO ATUAL') return snapshot.cash_balance
  if (screen === 'executive' && label === 'A RECEBER') return snapshot.receivables_due_next_90_days
  if (screen === 'executive' && label === 'A PAGAR') return snapshot.payables_due_next_90_days
  if (screen === 'receivables' && label === 'TOTAL A RECEBER') return snapshot.receivables
  if (screen === 'receivables' && label === 'RECEBIDO NO MÊS') return snapshot.receipts_month
  if (screen === 'receivables' && label === 'EM ATRASO') return snapshot.receivables_overdue
  if (screen === 'payables' && label === 'TOTAL A PAGAR') return snapshot.payables
  if (screen === 'payables' && label === 'VENCE HOJE') return snapshot.payables_due_today
  if (screen === 'payables' && label === 'EM ATRASO') return snapshot.payables_overdue
  if (screen === 'payables' && label === 'VENCIMENTOS 7 DIAS') return snapshot.payables_due_next_7_days
  if (screen === 'cashflow' && label === 'RECEBIDO NO MÊS') return snapshot.receipts_month
  if (screen === 'cashflow' && label === 'PAGO NO MÊS') return snapshot.payments_month
  return undefined
}

function withSnapshot(screen: MockupScreenKey, card: Card, snapshot?: DashboardSnapshot): Card {
  if ((screen === 'cashflow' || screen === 'executive') && card.variant === 'minimum-balance') {
    const cashFlow = snapshot?.cash_flow
    return { ...card, minimumValues: [
      { label: '30 dias', metric: cashFlow?.minimum_balance_30_days?.balance, date: cashFlow?.minimum_balance_30_days?.date, headroom: cashFlow?.minimum_balance_30_days?.headroom },
      { label: '60 dias', metric: cashFlow?.minimum_balance_60_days?.balance, date: cashFlow?.minimum_balance_60_days?.date, headroom: cashFlow?.minimum_balance_60_days?.headroom },
    ] }
  }
  if (screen === 'cashflow' && card.variant === 'period-comparison') {
    const metrics = card.label === 'ENTRADAS'
      ? [snapshot?.cash_flow?.inflows_next_30_days, snapshot?.cash_flow?.inflows_next_60_days]
      : [snapshot?.cash_flow?.outflows_next_30_days, snapshot?.cash_flow?.outflows_next_60_days]
    return { ...card, periodValues: [{ label: '30 dias', metric: metrics[0] }, { label: '60 dias', metric: metrics[1] }] }
  }
  const metric = metricForCard(screen, card.label, snapshot)
  if (metric?.state === 'UNAVAILABLE' || !metric?.value) return card
  return {
    ...card,
    value: metric.value,
    supporting: metric.state === 'PARTIAL'
      ? `Parcial${card.label === 'SALDO ATUAL' ? '' : card.supporting ? ` · ${card.supporting}` : ''}`
      : card.supporting,
  }
}

function EmptyChart({ title, cashFlow, rangeLabel }: { title: string; cashFlow?: DashboardSnapshot['cash_flow']; rangeLabel?: string }) {
  const [hoveredPeriod, setHoveredPeriod] = useState<number | undefined>()
  const [pinnedPeriod, setPinnedPeriod] = useState<number | undefined>()
  const periods = cashFlow?.periods ?? []
  const value = (amount?: string) => {
    if (!amount) return undefined
    const parsed = Number(amount)
    return Number.isFinite(parsed) ? parsed : undefined
  }
  const inflows = periods.map((period) => value(period.inflows.value))
  const outflows = periods.map((period) => value(period.outflows.value))
  const plottedOutflows = outflows.map((amount) => amount === undefined ? undefined : Math.abs(amount))
  const balances = periods.map((period) => value(period.closing_balance?.value))
  const points = [...inflows, ...plottedOutflows, ...balances].filter((amount): amount is number => amount !== undefined)
  const scale = Math.max(1, ...points.map((amount) => Math.abs(amount)))
  const plot = { left: 76, right: 640, top: 12, bottom: 125 }
  const xFor = (index: number) => periods.length < 2 ? (plot.left + plot.right) / 2 : plot.left + index * ((plot.right - plot.left) / (periods.length - 1))
  const yFor = (amount: number) => plot.top + ((scale - amount) / (scale * 2)) * (plot.bottom - plot.top)
  const line = (series: Array<number | undefined>) => series.map((amount, index) => {
    const x = xFor(index)
    const y = yFor(amount ?? 0)
    return `${x},${y}`
  }).join(' ')
  const axisLabel = (amount: number) => {
    const magnitude = Math.abs(amount)
    const sign = amount < 0 ? '-' : ''
    if (magnitude >= 1_000_000) return `${sign}R$ ${new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 1 }).format(magnitude / 1_000_000)} mi`
    if (magnitude >= 1_000) return `${sign}R$ ${new Intl.NumberFormat('pt-BR', { maximumFractionDigits: 0 }).format(magnitude / 1_000)} mil`
    return formatBRL(amount.toFixed(2))
  }
  const fullDateLabel = (date: string) => {
    const [year, month, day] = date.split('-')
    return `${day}/${month}/${year}`
  }
  const moneyLabel = (amount?: string) => amount ? formatBRL(amount) : '—'
  const absoluteMoneyLabel = (amount?: string) => moneyLabel(amount?.startsWith('-') ? amount.slice(1) : amount)
  const shortVariation = (current?: string, previous?: string, magnitude = false) => current && previous
    ? formatPercentVariation(current, previous, magnitude)?.replace(' vs. semana anterior', '')
    : undefined
  const hasFlow = periods.length > 0 && points.length > 0
  const hasConfirmed = cashFlow?.inflows_next_30_days.state === 'CONFIRMED' && cashFlow?.outflows_next_30_days.state === 'CONFIRMED'
  const yTicks = [scale, scale / 2, 0, -scale / 2, -scale]
  const hasBalanceSeries = balances.every((amount) => amount !== undefined)
  const activePeriod = hoveredPeriod ?? pinnedPeriod
  const shortDateRange = (period: DashboardCashFlowPeriod) => `${fullDateLabel(period.from)} a ${fullDateLabel(period.through)}`
  const popupPosition = (index: number) => {
    const x = xFor(index)
    const yValues = [inflows[index], plottedOutflows[index], balances[index]].filter((amount): amount is number => amount !== undefined)
    const y = yValues.length > 0 ? yFor(yValues[0]) : plot.top
    return { x: Math.max(plot.left, Math.min(plot.right - 202, x + 10)), y: Math.max(plot.top + 2, Math.min(plot.bottom - 84, y - 14)) }
  }
  return (
    <section className="mocap-panel mocap-chart-panel">
      <div className="mocap-panel-heading"><div><Icon icon="calendar" /><div><h2>{title}</h2><p>{hasFlow ? (hasConfirmed ? 'Fonte: Bling' : 'Parcial · fonte: Bling') : 'Sem série confirmada'}{rangeLabel ? ` · ${rangeLabel}` : ''}</p></div></div></div>
      <div className="mocap-chart-legend"><span className="legend-blue">●</span> Entradas <span className="legend-red">●</span> Saídas <span className="legend-orange">━●</span> Saldo <span>Variação semanal</span></div>
      <div className="mocap-chart-empty">
        {hasFlow ? (
          <svg role="group" aria-label="Gráfico semanal de receitas, despesas e saldo resultante do fluxo de caixa Bling" viewBox="0 0 680 220" preserveAspectRatio="none" style={{ position: 'relative', zIndex: 1, width: '100%', height: '100%', overflow: 'visible' }}>
            {yTicks.map((tick) => {
              const y = yFor(tick)
              return <g key={tick}><line x1={plot.left} x2={plot.right} y1={y} y2={y} stroke={tick === 0 ? '#d8e5ef' : '#e7eff6'} strokeWidth={tick === 0 ? 2 : 1} /><text x={plot.left - 8} y={y + 4} textAnchor="end" fill={tick < 0 ? '#d93025' : '#4564b0'} fontSize="10">{axisLabel(tick)}</text></g>
            })}
            <polyline points={line(inflows)} fill="none" stroke="#2877eb" strokeWidth="2" vectorEffect="non-scaling-stroke" />
            <polyline points={line(plottedOutflows)} fill="none" stroke="#fa3038" strokeWidth="2" vectorEffect="non-scaling-stroke" />
            {hasBalanceSeries && <polyline points={line(balances)} fill="none" stroke="#ff9f1a" strokeWidth="2" vectorEffect="non-scaling-stroke" />}
            {periods.map((period, index) => {
              const x = xFor(index)
              const inflow = inflows[index]
              const outflow = plottedOutflows[index]
              const balance = balances[index]
              const inflowVariation = shortVariation(period.inflows.value, periods[index - 1]?.inflows.value)
              const outflowVariation = shortVariation(period.outflows.value, periods[index - 1]?.outflows.value, true)
              const selected = activePeriod === index
              const expenseLabel = absoluteMoneyLabel(period.outflows.value)
              const pointLabel = `${shortDateRange(period)}: Despesas ${expenseLabel}${outflowVariation ? `, ${outflowVariation}` : ''}; Receitas ${moneyLabel(period.inflows.value)}${inflowVariation ? `, ${inflowVariation}` : ''}; Saldo ${moneyLabel(period.closing_balance?.value)}`
              return (
                <g key={`${period.from}-${period.through}`} role="button" tabIndex={0} aria-label={pointLabel} onMouseEnter={() => setHoveredPeriod(index)} onMouseLeave={() => setHoveredPeriod(undefined)} onFocus={() => setHoveredPeriod(index)} onBlur={() => setHoveredPeriod(undefined)} onClick={() => setPinnedPeriod((current) => current === index ? undefined : index)} onKeyDown={(event) => { if (event.key === 'Escape') { setPinnedPeriod(undefined); setHoveredPeriod(undefined) } else if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); setPinnedPeriod((current) => current === index ? undefined : index) } }} style={{ cursor: 'pointer', outline: 'none' }}>
                  <title>{`${shortDateRange(period)} · Despesas ${expenseLabel} · Receitas ${moneyLabel(period.inflows.value)} · Saldo ${moneyLabel(period.closing_balance?.value)}`}</title>
                  <rect x={x - 13} y={plot.top} width="26" height={plot.bottom - plot.top} fill="transparent" />
                  <text x={x} y="164" textAnchor="middle" fill="#17305b" fontSize="9" aria-hidden="true" data-axis-date="start">{fullDateLabel(period.from)} a</text>
                  <text x={x} y="183" textAnchor="middle" fill="#17305b" fontSize="9" aria-hidden="true" data-axis-date="end">{fullDateLabel(period.through)}</text>
                  {inflow !== undefined && <circle cx={x} cy={yFor(inflow)} r={selected ? 4.5 : 3.5} fill="#2877eb" />}
                  {outflow !== undefined && <circle cx={x} cy={yFor(outflow)} r={selected ? 4.5 : 3.5} fill="#fa3038" />}
                  {balance !== undefined && <circle cx={x} cy={yFor(balance)} r={selected ? 4.5 : 3.5} fill={balance < 0 ? '#d93025' : '#ff9f1a'} />}
                </g>
              )
            })}
            {activePeriod !== undefined && periods[activePeriod] && (() => {
              const period = periods[activePeriod]
              const position = popupPosition(activePeriod)
              const expenseText = absoluteMoneyLabel(period.outflows.value)
              const incomeVariation = shortVariation(period.inflows.value, periods[activePeriod - 1]?.inflows.value)
              const expenseVariation = shortVariation(period.outflows.value, periods[activePeriod - 1]?.outflows.value, true)
              const closingBalanceValue = value(period.closing_balance?.value)
              return (
                <g className="cash-flow-chart-tooltip" pointerEvents="none" aria-hidden="true" transform={`translate(${position.x} ${position.y})`}>
                  <rect width="202" height="82" rx="2" fill="#fff" stroke="#c8c8c8" /><rect width="202" height="21" fill="#ededed" />
                  <text x="7" y="14" fill="#555" fontSize="10" fontWeight="700">{shortDateRange(period)}</text><line x1="0" x2="202" y1="42" y2="42" stroke="#ddd" /><line x1="0" x2="202" y1="62" y2="62" stroke="#ddd" />
                  <rect x="8" y="28" width="9" height="9" fill="#fa3038" /><text x="23" y="37" fill="#222" fontSize="10">Despesas</text><text x="194" y="37" textAnchor="end" fill="#d93025" fontSize="10">{expenseText}{expenseVariation ? ` · ${expenseVariation}` : ''}</text>
                  <rect x="8" y="48" width="9" height="9" fill="#2877eb" /><text x="23" y="57" fill="#222" fontSize="10">Receitas</text><text x="194" y="57" textAnchor="end" fill={value(period.inflows.value) !== undefined && value(period.inflows.value)! < 0 ? '#d93025' : '#222'} fontSize="10">{moneyLabel(period.inflows.value)}{incomeVariation ? ` · ${incomeVariation}` : ''}</text>
                  <rect x="8" y="68" width="9" height="9" fill={closingBalanceValue !== undefined && closingBalanceValue < 0 ? '#d93025' : '#ff9f1a'} /><text x="23" y="77" fill="#222" fontSize="10">Saldo</text><text x="194" y="77" textAnchor="end" fill={closingBalanceValue !== undefined && closingBalanceValue < 0 ? '#d93025' : '#222'} fontSize="10">{moneyLabel(period.closing_balance?.value)}</text>
                </g>
              )
            })()}
          </svg>
        ) : <span>Sem dados confirmados</span>}
      </div>
    </section>
  )
}

function EmptyDistribution({ title, subtitle, onNavigate, emptyMessage = 'Sem dados', items, bankBalances = [] }: { title: string; subtitle: string; onNavigate?: (screen: MockupScreenKey) => void; emptyMessage?: string; items?: string[]; bankBalances?: DashboardBankBalance[] }) {
  const legendItems = bankBalances.length > 0 ? bankBalances.map((balance) => balance.account) : items ?? ['Confirmado', 'Provisório', 'Pendente', 'Divergente']
  const colors = ['dot-teal', 'dot-blue', 'dot-orange', 'dot-purple']
  return <section className="mocap-panel mocap-distribution"><div className="mocap-panel-heading"><div><Icon icon="pie" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-donut-wrap"><div className="mocap-donut"><span>{bankBalances.length > 0 ? 'Soma dos lançamentos' : emptyMessage}</span></div><div className="mocap-legend-list">{legendItems.map((item, index) => <span key={item}><i className={`tone-dot ${colors[index % colors.length]}`} />{item}<b>{bankBalances[index] ? formatBRL(bankBalances[index].value) : '—'}</b></span>)}</div></div><Tooltip id="metric.details"><button type="button" className="mocap-link" disabled={!onNavigate} onClick={() => onNavigate?.('cashflow')}>Ver detalhes <span>→</span></button></Tooltip></section>
}

function EmptyTable({ title, subtitle, columns, rows = [], emptyMessage = 'Sem dados confirmados' }: { title: string; subtitle: string; columns: string[]; rows?: string[][]; emptyMessage?: string }) {
  return <section className="mocap-panel mocap-table-panel"><div className="mocap-panel-heading"><div><Icon icon="calendar" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-table-scroll"><table><thead><tr>{columns.map((column) => <th key={column}>{column}</th>)}</tr></thead><tbody>{rows.length === 0 ? <tr><td colSpan={columns.length}><span>{emptyMessage}</span></td></tr> : rows.map((row, rowIndex) => <tr key={`${rowIndex}-${row[0]}`}>{columns.map((column, columnIndex) => <td key={`${column}-${columnIndex}`}>{row[columnIndex] ?? '—'}</td>)}</tr>)}</tbody></table></div><Tooltip id="metric.details"><button type="button" className="mocap-link" disabled>Ver detalhes <span>→</span></button></Tooltip></section>
}

function EmptyRightPanel({ title, subtitle }: { title: string; subtitle: string }) {
  return <section className="mocap-panel mocap-right-empty"><div className="mocap-panel-heading"><div><Icon icon="pie" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-right-grid"><div><Icon icon="calendar" /><strong>Período</strong><EmptyValue /></div><div><Icon icon="warning" /><strong>Status</strong><span>Sem dados confirmados</span></div><div><Icon icon="chart" /><strong>Regra</strong><span>Aguardando fonte oficial</span></div><div><Icon icon="refresh" /><strong>Atualização</strong><EmptyValue /></div></div><Tooltip id="metric.details"><button type="button" className="mocap-link" disabled>Ver detalhes <span>→</span></button></Tooltip></section>
}

function formatCashFlowPeriod(period?: DashboardCashFlowPeriod): string {
  if (!period) return '—'
  const label = (date: string) => {
    const [, month, day] = date.split('-')
    return `${day}/${month}`
  }
  return `${label(period.from)}–${label(period.through)}`
}

function criticalWeekMetric(metric?: DashboardMetric): string {
  if (!metric?.value || metric.state === 'UNAVAILABLE') return '—'
  return `${formatBRL(metric.value)}${metric.state === 'PARTIAL' ? ' · Parcial' : ''}`
}

function EmptyCriticalPanel({ title, subtitle, cashFlow }: { title: string; subtitle: string; cashFlow?: DashboardSnapshot['cash_flow'] }) {
  const week = cashFlow?.critical_week
  return <section className="mocap-panel mocap-critical-panel"><div className="mocap-panel-heading"><div><Icon icon="clock" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-critical-grid"><div><Icon icon="calendar" /><span>Período</span><strong>{formatCashFlowPeriod(week)}</strong></div><div><Icon icon="trend" /><span>Saídas</span><strong className={week?.outflows.value && /^\s*[−-]/.test(week.outflows.value) ? 'mocap-negative-value' : undefined}>{criticalWeekMetric(week?.outflows)}</strong></div><div><Icon icon="trend" /><span>Entradas</span><strong className={week?.inflows.value && /^\s*[−-]/.test(week.inflows.value) ? 'mocap-negative-value' : undefined}>{criticalWeekMetric(week?.inflows)}</strong></div><div><Icon icon="money" /><span>Reserva mínima</span><strong>{formatBRL(APPROVED_MINIMUM_RESERVE)}</strong></div></div></section>
}

function EmptyCalculationPanel({ title, subtitle }: { title: string; subtitle: string }) {
  const rows = ['Resultado Líquido DRE', '(-) Valor já aplicado CDI Plus', 'Caixa consolidado', 'Reserva mínima', 'Saldo transferível', 'Capacidade de investimento']
  return <section className="mocap-panel mocap-calculation-panel"><div className="mocap-panel-heading"><div><Icon icon="calculator" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-calc-rows">{rows.map((row, index) => <div key={row} className={index === rows.length - 1 ? 'emphasis' : ''}><span>{row}</span><strong>—</strong></div>)}</div><Tooltip id="metric.details"><button type="button" className="mocap-link" disabled>Ver memória de cálculo <span>→</span></button></Tooltip></section>
}

function EmptyRulesPanel({ title, subtitle }: { title: string; subtitle: string }) {
  return <section className="mocap-panel mocap-rules-panel"><div className="mocap-panel-heading"><div><Icon icon="document" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-rules-grid"><div><Icon icon="chart" /><span>Fonte do lucro<strong>DRE Majucau</strong></span></div><div><Icon icon="wallet" /><span>Reserva mínima<strong>—</strong></span></div><div><Icon icon="settings" /><span>PMR / PMP / PME<strong>Diagnóstico</strong></span></div><div><Icon icon="clock" /><span>Semana crítica<strong>OPEN</strong></span></div></div><div className="mocap-rules-warning"><Icon icon="warning" />O DEV não define sozinho a posição do caixa da semana crítica</div><Tooltip id="settings.open"><button type="button" className="mocap-link" disabled>Ver parâmetros <span>→</span></button></Tooltip></section>
}

function EmptyMarketPanel({ title, subtitle }: { title: string; subtitle: string }) {
  const cards: Array<[MockupIcon, string, string]> = [['trend', 'COTAÇÃO INTERNACIONAL', 'Sem cotação confirmada'], ['warning', 'ALERTA CLIMÁTICO', 'Aguardando fonte'], ['leaf', 'SAFRA 2026/2027', 'Aguardando fonte'], ['flow', 'CÂMBIO (USD/BRL)', 'Aguardando fonte']]
  return <section className="mocap-panel mocap-market-panel"><div className="mocap-panel-heading"><div><Icon icon="leaf" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-market-grid">{cards.map(([icon, label, value]) => <div key={label} className="mocap-market-card"><Icon icon={icon} /><strong>{label}</strong><b>{value}</b><span>Fonte oficial pendente</span></div>)}</div></section>
}

function EmptyRiskPanel({ title, subtitle }: { title: string; subtitle: string }) {
  return <section className="mocap-panel mocap-risk-panel"><div className="mocap-panel-heading"><div><Icon icon="pie" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-risk-gauge"><div className="mocap-risk-arc"><span>Risco<strong>SEM DADOS</strong></span></div><div className="mocap-risk-legend"><span><i className="dot-red" />Alto</span><span><i className="dot-yellow" />Médio</span><span><i className="dot-green" />Baixo</span></div></div></section>
}

function EmptyMetadataPanel({ title, subtitle }: { title: string; subtitle: string }) {
  const rows = ['Versão ativa', 'Gerado em', 'Janela de treino', 'Bias', 'MAE', 'Modelo', 'Status']
  return <section className="mocap-panel mocap-metadata-panel"><div className="mocap-panel-heading"><div><Icon icon="document" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-metadata-table">{rows.map((row) => <div key={row}><span>{row}</span><strong>—</strong></div>)}</div><span className="mocap-metadata-note">ⓘ Venda ≠ Demanda</span></section>
}

function EmptyImpactTable({ title, subtitle, columns }: { title: string; subtitle: string; columns: string[] }) {
  return <section className="mocap-panel mocap-impact-panel"><div className="mocap-panel-heading"><div><Icon icon="box" /><div><h2>{title}</h2><p>{subtitle}</p></div></div></div><div className="mocap-table-scroll"><table><thead><tr>{columns.map((column) => <th key={column}>{column}</th>)}</tr></thead><tbody><tr><td colSpan={columns.length}><span>Sem dados confirmados</span></td></tr></tbody></table></div></section>
}

function EmptyPurchaseTable({ columns }: { columns: string[] }) {
  return <section className="mocap-panel mocap-purchase-panel"><div className="mocap-panel-heading"><div><Icon icon="shopping" /><div><h2>COMPRAS RECOMENDADAS</h2><p>Com base em forecast, produção, estoque, lead time e mercado</p></div></div><div className="mocap-table-actions"><button type="button" disabled>Próximos 60 dias⌄</button><button type="button" disabled>Todos os insumos⌄</button></div></div><div className="mocap-table-scroll"><table><thead><tr>{columns.map((column) => <th key={column}>{column}</th>)}</tr></thead><tbody><tr><td colSpan={columns.length}><span>Sem dados confirmados</span></td></tr></tbody></table></div></section>
}

function DreDerivedIndicators() {
  return (
    <>
      <article className="mocap-card mocap-dre-indicator-card tone-blue" aria-labelledby="mocap-dre-ebit-title">
        <div className="mocap-card-icon"><Icon icon="chart" /></div>
        <div className="mocap-card-body">
          <Tooltip id="result.ebit"><span id="mocap-dre-ebit-title" className="mocap-card-label">EBIT <span className="mocap-dre-indicator-alias">LAJIR</span></span></Tooltip>
          <strong className="mocap-empty-value">—</strong>
          <span className="mocap-card-support">Sem dados confirmados</span>
          <p><b>Em termos simples:</b> resultado depois de descontar a depreciação.</p>
        </div>
      </article>
      <article className="mocap-card mocap-dre-indicator-card tone-green" aria-labelledby="mocap-dre-ebitda-title">
        <div className="mocap-card-icon"><Icon icon="trend" /></div>
        <div className="mocap-card-body">
          <Tooltip id="result.ebitda"><span id="mocap-dre-ebitda-title" className="mocap-card-label">EBITDA <span className="mocap-dre-indicator-alias">LAJIDA</span></span></Tooltip>
          <strong className="mocap-empty-value">—</strong>
          <span className="mocap-card-support">Sem dados confirmados</span>
          <p><b>Em termos simples:</b> resultado antes de descontar a depreciação.</p>
        </div>
      </article>
    </>
  )
}

function capexValue(metric?: DashboardMetric): string {
  if (!metric || metric.state === 'UNAVAILABLE' || metric.state === 'PROJECTED' || !metric.value) return '—'
  return formatBRL(metric.value)
}

function capexSupport(metric?: DashboardMetric): string {
  if (metric?.state === 'CONFIRMED' && metric.value) return 'Valor confirmado no relatório'
  if (metric?.state === 'PARTIAL' && metric.value) return 'Relatório parcialmente conferido'
  return 'Aguardando integração do relatório'
}

function CapexIndicators({ snapshot }: { snapshot?: DashboardSnapshot }) {
  const paidMonth = snapshot?.capex_paid_month
  const accumulated = snapshot?.capex_accumulated
  return (
    <>
      <article className="mocap-card mocap-dre-indicator-card mocap-dre-capex-card tone-purple" aria-labelledby="mocap-dre-capex-paid-title">
        <div className="mocap-card-icon"><Icon icon="building" /></div>
        <div className="mocap-card-body">
          <Tooltip id="result.capexPaid"><span id="mocap-dre-capex-paid-title" className="mocap-card-label">CAPEX PAGO NO MÊS</span></Tooltip>
          <strong className="mocap-empty-value">{capexValue(paidMonth)}</strong>
          <span className="mocap-card-support">{capexSupport(paidMonth)}</span>
          <p>Informativo; não entra no lucro da DRE.</p>
        </div>
      </article>
      <article className="mocap-card mocap-dre-indicator-card mocap-dre-capex-card tone-blue" aria-labelledby="mocap-dre-capex-accumulated-title">
        <div className="mocap-card-icon"><Icon icon="building" /></div>
        <div className="mocap-card-body">
          <Tooltip id="result.capexAccumulated"><span id="mocap-dre-capex-accumulated-title" className="mocap-card-label">CAPEX ACUMULADO</span></Tooltip>
          <strong className="mocap-empty-value">{capexValue(accumulated)}</strong>
          <span className="mocap-card-support">{capexSupport(accumulated)}</span>
          <p>Informativo; não entra no lucro da DRE.</p>
        </div>
      </article>
    </>
  )
}

function executiveAlertAction(message: string): { screen: MockupScreenKey; label: string } | undefined {
  if (message === 'Saldo projetado abaixo da reserva mínima') return { screen: 'cashflow', label: 'Ver fluxo' }
  return undefined
}

function executiveCardTarget(label: string): MockupScreenKey | undefined {
  if (label === 'SALDO ATUAL' || label === 'MENOR SALDO 30 / 60 DIAS') return 'cashflow'
  if (label === 'A RECEBER') return 'receivables'
  if (label === 'A PAGAR') return 'payables'
  if (label === 'VENDAS DO MÊS' || label === 'MARGEM BRUTA') return 'dre'
  if (label === 'ESTOQUE / COBERTURA') return 'inventory'
  if (label === 'PRODUÇÃO PREVISTA') return 'production'
  if (label === 'COMPRAS PREVISTAS') return 'purchases'
  return undefined
}

function AlertPanel({ messages, screen, onNavigate, title = 'ALERTAS IMPORTANTES', subtitles = [], hideActions = false }: { messages: string[]; screen: MockupScreenKey; onNavigate?: (screen: MockupScreenKey) => void; title?: string; subtitles?: string[]; hideActions?: boolean }) {
  return <section className={`mocap-alerts${screen === 'cashflow' ? ' mocap-alerts-cashflow' : ''}`}><div className="mocap-alerts-heading"><Icon icon="warning" /><h2>{title}</h2></div><div className={`mocap-alert-grid ${screen === 'purchases' ? 'mocap-alert-grid-purchases' : ''}`}>{messages.map((message, index) => {
    const action = screen === 'executive' ? executiveAlertAction(message) : index === 0 ? { screen: 'integrations' as const, label: 'Ir para Integrações' } : undefined
    return <article key={message} className={`mocap-alert alert-${index % 4}`}><div className="mocap-alert-icon"><Icon icon={index === 0 ? 'warning' : index === 1 ? 'document' : index === 2 ? 'clock' : 'refresh'} /></div><div><strong>{message}</strong><span>{subtitles[index] ?? 'Sem dados confirmados'}</span>{!hideActions && <Tooltip id={screen === 'executive' ? 'metric.details' : 'notice.integrations'}><button type="button" disabled={!onNavigate || !action} aria-label={action ? `${action.label}: ${message}` : undefined} onClick={action && onNavigate ? () => onNavigate(action.screen) : undefined}>{action?.label ?? 'Ver detalhes'} <span>→</span></button></Tooltip>}</div></article>
  })}</div></section>
}

export function MockupDashboard({ screen, snapshot, onNavigate }: { screen: MockupScreenKey; snapshot?: DashboardSnapshot; onNavigate?: (screen: MockupScreenKey) => void }) {
  const [periodRange, setPeriodRange] = useState({ from: '', through: '' })
  const config = configs[screen === 'integrations' ? 'executive' : screen]
  const rawCashFlow = screen === 'cashflow' ? snapshot?.cash_flow : undefined
  const sourceStartDate = rawCashFlow?.days?.[0]?.date ?? snapshot?.as_of.slice(0, 10) ?? ''
  const currentDate = saoPauloToday()
  const minDate = sourceStartDate && sourceStartDate > currentDate ? sourceStartDate : currentDate
  const maxDate = rawCashFlow?.days?.[rawCashFlow.days.length - 1]?.date ?? availableCashFlowDate(rawCashFlow, snapshot, 60)
  const selectedFrom = periodRange.from >= minDate && periodRange.from <= maxDate ? periodRange.from : minDate
  const selectedThrough = periodRange.through >= minDate && periodRange.through <= maxDate ? periodRange.through : maxDate
  const selectedPeriods = groupSelectedCashFlowDays(rawCashFlow, selectedFrom, selectedThrough)
  const selectedCriticalWeek = selectedPeriods.reduce<DashboardCashFlowPeriod | undefined>((selected, period) => {
    const parseUnits = (value?: string) => {
      const match = /^(-?)(\d{1,15})(?:\.(\d{1,4}))?$/.exec(value ?? '')
      if (!match) return undefined
      const amount = BigInt(match[2]) * 10_000n + BigInt((match[3] ?? '').padEnd(4, '0'))
      return match[1] ? -amount : amount
    }
    const current = parseUnits(period.outflows.value)
    const best = parseUnits(selected?.outflows.value)
    if (current === undefined) return selected
    return best === undefined || (current < 0n ? -current : current) > (best < 0n ? -best : best) ? period : selected
  }, undefined)
  const cashFlowView = rawCashFlow ? { ...rawCashFlow, periods: selectedPeriods, critical_week: selectedCriticalWeek } : undefined
  const cards = config.cards.map((card) => withSnapshot(screen, card, snapshot))
  const lowerCards = config.lower
    .filter((card) => screen !== 'dre' || card.label !== 'UNMAPPED')
    .map((card) => withSnapshot(screen, card, snapshot))
  const receivableRows = screen === 'receivables' ? (snapshot?.receivable_rows ?? []).map(receivableTableRow) : []
  const payableRows = screen === 'payables' ? (snapshot?.payable_rows ?? []).map(payableTableRow) : []
  const tableRows = screen === 'receivables' ? receivableRows : screen === 'payables' ? payableRows : []
  const middleLeft = screen === 'cashflow' ? <div className="mocap-cashflow-visuals"><CashFlowWeeklyTable periods={selectedPeriods} /><EmptyChart title="GRÁFICO DO FLUXO DE CAIXA" cashFlow={cashFlowView} rangeLabel={`${selectedFrom} a ${selectedThrough}`} /></div> : config.layout === 'market' ? <EmptyMarketPanel title={config.middleTitle} subtitle={config.middleSubtitle} /> : config.layout === 'calculation' ? <EmptyCalculationPanel title={config.middleTitle} subtitle={config.middleSubtitle} /> : config.table ? <EmptyTable title={config.middleTitle} subtitle={config.middleSubtitle} columns={config.table} rows={config.tableRows ?? tableRows} emptyMessage={config.tableEmptyMessage} /> : config.chart ? <EmptyChart title={config.middleTitle} cashFlow={snapshot?.cash_flow} /> : <EmptyRightPanel title={config.middleTitle} subtitle={config.middleSubtitle} />
  const middleRight = config.layout === 'market' ? <EmptyRiskPanel title={config.rightTitle} subtitle={config.rightSubtitle} /> : config.layout === 'critical' ? <EmptyCriticalPanel title={config.rightTitle} subtitle={config.rightSubtitle} cashFlow={cashFlowView} /> : config.layout === 'calculation' ? <EmptyRulesPanel title={config.rightTitle} subtitle={config.rightSubtitle} /> : config.layout === 'forecast' ? <EmptyMetadataPanel title={config.rightTitle} subtitle={config.rightSubtitle} /> : <EmptyDistribution title={config.rightTitle} subtitle={config.rightSubtitle} emptyMessage={config.distributionEmptyMessage} items={config.distributionItems} bankBalances={screen === 'executive' ? snapshot?.bank_balances : undefined} />
  const lower = screen === 'cashflow' ? null : screen === 'dre' ? <section className="mocap-lower-cards mocap-dre-lower-cards" aria-label="Indicadores e composição da DRE">{lowerCards.map((card) => <MocapCard key={card.label} card={card} />)}<DreDerivedIndicators /><CapexIndicators snapshot={snapshot} /></section> : config.layout === 'market' ? <><EmptyPurchaseTable columns={config.table ?? []} /><section className="mocap-market-lower"><EmptyChart title="NECESSIDADE DE COMPRA (TONELADAS)" /><EmptyDistribution title="ORIGEM DA NECESSIDADE" subtitle="Produção e estoque" /><AlertPanel messages={config.alerts} screen={screen} onNavigate={onNavigate} /></section></> : config.layout === 'forecast' ? <><EmptyImpactTable title="ITENS DE MAIOR IMPACTO" subtitle="SKUs com maior desvio entre forecast e realizado" columns={config.table ?? []} /><section className="mocap-lower-cards">{lowerCards.map((card) => <MocapCard key={card.label} card={card} />)}</section></> : <section className="mocap-lower-cards">{lowerCards.map((card) => <MocapCard key={card.label} card={card} detailScreen={screen === 'executive' ? executiveCardTarget(card.label) : undefined} onNavigate={screen === 'executive' ? onNavigate : undefined} />)}</section>
  return <div className={`mocap-page${screen === 'cashflow' ? ' mocap-page-cashflow' : ''}`}><MocapHeader config={config} snapshot={snapshot} /><section className="mocap-top-cards">{cards.map((card) => <MocapCard key={card.label} card={card} hideAction={config.hideCardActions || (screen === 'cashflow' && ['SALDO ATUAL', 'ENTRADAS', 'SAÍDAS'].includes(card.label))} detailScreen={screen === 'executive' ? executiveCardTarget(card.label) : undefined} onNavigate={screen === 'executive' ? onNavigate : undefined} />)}</section>{screen === 'cashflow' && <CashFlowDateControls from={selectedFrom} through={selectedThrough} min={minDate} max={maxDate} onFromChange={(value) => setPeriodRange((current) => { const through = current.through || maxDate; return { from: value, through: through >= value ? through : value } })} onThroughChange={(value) => setPeriodRange((current) => { const from = current.from || minDate; return { from: from <= value ? from : value, through: value } })} />}<section className={`mocap-middle ${config.layout ? `layout-${config.layout}` : ''}`}>{screen === 'executive' && config.chart ? <EmptyChart title={config.middleTitle} cashFlow={snapshot?.cash_flow} /> : middleLeft}{screen === 'executive' && !config.layout ? <EmptyDistribution title={config.rightTitle} subtitle={config.rightSubtitle} onNavigate={onNavigate} /> : middleRight}</section>{lower}{config.layout !== 'market' && config.alerts.length > 0 && <AlertPanel messages={config.alerts} screen={screen} onNavigate={onNavigate} title={config.alertsTitle} subtitles={config.alertSubtitles} hideActions={config.hideCardActions} />}</div>
}
