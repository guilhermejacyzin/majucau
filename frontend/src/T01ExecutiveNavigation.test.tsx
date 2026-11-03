import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import type { DashboardSnapshot } from './bootstrap'
import { MockupDashboard, type MockupScreenKey } from './mockupDashboard'

const cardRoutes = [
  ['SALDO ATUAL', 'cashflow'],
  ['A RECEBER', 'receivables'],
  ['A PAGAR', 'payables'],
  ['MENOR SALDO 30 / 60 DIAS', 'cashflow'],
  ['VENDAS DO MÊS', 'dre'],
  ['MARGEM BRUTA', 'dre'],
  ['ESTOQUE / COBERTURA', 'inventory'],
  ['PRODUÇÃO PREVISTA', 'production'],
  ['COMPRAS PREVISTAS', 'purchases'],
] as const satisfies ReadonlyArray<readonly [string, MockupScreenKey]>

describe('atalhos da Visão Executiva para as telas de origem', () => {
  it('não exibe na tela Estoque os cards ou alerta de envelhecimento/validade removidos', () => {
    render(<MockupDashboard screen="inventory" />)

    expect(screen.getByText('ESTOQUE FÍSICO')).toBeInTheDocument()
    expect(screen.getByText('GIRO')).toBeInTheDocument()
    expect(screen.getByText('Saldo e cobertura')).toBeInTheDocument()
    expect(screen.queryByText('Saldo, cobertura e validade')).not.toBeInTheDocument()
    expect(screen.queryByText('ENVELHECENDO')).not.toBeInTheDocument()
    expect(screen.queryByText('VALIDADE', { exact: true })).not.toBeInTheDocument()
    expect(screen.queryByText('Lotes com validade próxima')).not.toBeInTheDocument()
  })

  it('retira da tela Produção os dois avisos apontados e mantém os demais', () => {
    render(<MockupDashboard screen="production" />)

    expect(screen.getByText('Ordens em atraso')).toBeInTheDocument()
    expect(screen.getByText('BOM versionada impacta a necessidade de insumos')).toBeInTheDocument()
    expect(screen.queryByText('Lotes com perda acima do esperado')).not.toBeInTheDocument()
    expect(screen.queryByText('Separar plano realizado e obrigatório')).not.toBeInTheDocument()
  })

  it('organiza os alertas de Compras em uma coluna para não sobrepor textos', () => {
    render(<MockupDashboard screen="purchases" />)

    const alerts = screen.getByRole('heading', { name: 'ALERTAS IMPORTANTES' }).closest('section')
    const grid = alerts?.querySelector('.mocap-alert-grid-purchases')
    expect(grid).not.toBeNull()
    expect(within(alerts as HTMLElement).getByText('Preço do cacau em queda')).toBeInTheDocument()
    expect(within(alerts as HTMLElement).getByText('Risco climático elevado')).toBeInTheDocument()
    expect(within(alerts as HTMLElement).getByText('Cobertura de manteiga de cacau inferior a 20 dias')).toBeInTheDocument()
    expect(within(alerts as HTMLElement).getByText('Fornecedores com lead time acima da média')).toBeInTheDocument()
  })

  it('remove toda a seção de alertas da tela Contas a Receber', () => {
    render(<MockupDashboard screen="receivables" />)

    expect(screen.getByText('TOTAL A RECEBER')).toBeInTheDocument()
    expect(screen.getByText('RECEBIDO NO MÊS')).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: 'RECEBÍVEIS' })).toBeInTheDocument()
    expect(screen.queryByRole('heading', { name: 'ALERTAS IMPORTANTES' })).not.toBeInTheDocument()
    expect(screen.queryByText('Recebimentos em atraso')).not.toBeInTheDocument()
    expect(screen.queryByText('Matches ambíguos aguardando análise')).not.toBeInTheDocument()
    expect(screen.queryByText('Transações sem ID Bling')).not.toBeInTheDocument()
    expect(screen.queryByText('Regras de match ativas')).not.toBeInTheDocument()
  })

  it('remove os três alertas genéricos da tela Fluxo de Caixa e mantém o da reserva', () => {
    render(<MockupDashboard screen="cashflow" />)

    expect(screen.queryByText('Saídas previstas sem categoria')).not.toBeInTheDocument()
    expect(screen.queryByText('Recebimentos em atraso')).not.toBeInTheDocument()
    expect(screen.queryByText('Conciliação pendente')).not.toBeInTheDocument()
    expect(screen.getByText('Saldo projetado abaixo da reserva mínima')).toBeInTheDocument()
  })

  it('liga o saldo atual e não reaproveita contas a receber/pagar no fluxo de caixa', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-02T12:00:00Z',
      data_state: 'PARTIAL',
      receivables: unavailable,
      future_b2c: unavailable,
      receipts_month: unavailable,
      receivables_overdue: unavailable,
      receivables_due_next_90_days: { value: '1250.00', count: 4, state: 'CONFIRMED', source_system: 'BLING' },
      payables: unavailable,
      payables_due_today: unavailable,
      payables_overdue: unavailable,
      payables_due_next_7_days: unavailable,
      payables_due_next_90_days: { value: '-800.00', count: 3, state: 'PARTIAL', source_system: 'BLING' },
      cash_balance: { value: '4500.00', count: 2, state: 'CONFIRMED', source_system: 'BLING' },
      payments_month: unavailable,
    }
    const { rerender } = render(<MockupDashboard screen="executive" snapshot={snapshot} />)
    expect(screen.getByText('SALDO ATUAL').closest('article')).toHaveTextContent('4.500,00')
    rerender(<MockupDashboard screen="cashflow" snapshot={snapshot} />)
    expect(screen.getByText('ENTRADAS 30 DIAS').closest('article')).not.toHaveTextContent('1.250,00')
    expect(screen.getByText('SAÍDAS 30 DIAS').closest('article')).not.toHaveTextContent('-800,00')
    expect(screen.getByText('SALDO ATUAL').closest('article')).toHaveTextContent('4.500,00')
  })

  it('não substitui os valores do Fluxo de Caixa por totais de Contas a Receber/Pagar', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-02T12:00:00Z',
      data_state: 'PARTIAL',
      receivables: unavailable,
      future_b2c: unavailable,
      receipts_month: unavailable,
      receivables_overdue: unavailable,
      receivables_due_next_90_days: { value: '999.00', state: 'CONFIRMED', source_system: 'BLING' },
      receivables_due_next_30_days: { value: '125.00', state: 'CONFIRMED', source_system: 'BLING + NUVEM_PAGO' },
      payables: unavailable,
      payables_due_today: unavailable,
      payables_overdue: unavailable,
      payables_due_next_7_days: unavailable,
      payables_due_next_90_days: { value: '-800.00', state: 'CONFIRMED', source_system: 'BLING' },
      payables_due_next_30_days: { value: '-80.00', state: 'PARTIAL', source_system: 'BLING' },
      payments_month: unavailable,
    }
    render(<MockupDashboard screen="cashflow" snapshot={snapshot} />)

    expect(screen.getByText('ENTRADAS 30 DIAS').closest('article')).toHaveTextContent('—')
    expect(screen.getByText('SAÍDAS 30 DIAS').closest('article')).toHaveTextContent('—')
    expect(screen.getByText('ENTRADAS 30 DIAS').closest('article')).toHaveTextContent('Realizado + previsto')
    expect(screen.getByText('SAÍDAS 30 DIAS').closest('article')).toHaveTextContent('Compromissos e pagamentos')
    expect(screen.getByText('ENTRADAS 30 DIAS').closest('article')).not.toHaveTextContent('125,00')
    expect(screen.getByText('SAÍDAS 30 DIAS').closest('article')).not.toHaveTextContent('80,00')
  })

  it('usa a composição da API do Fluxo de Caixa na T02 e na T01 com aviso parcial', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const partial = { state: 'PARTIAL' as const, source_system: 'BLING' }
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-03T12:00:00Z',
      data_state: 'PARTIAL',
      receivables: unavailable,
      future_b2c: unavailable,
      receipts_month: unavailable,
      receivables_overdue: unavailable,
      payables: unavailable,
      payables_due_today: unavailable,
      payables_overdue: unavailable,
      payables_due_next_7_days: unavailable,
      payments_month: unavailable,
      cash_flow: {
        inflows_next_30_days: { ...partial, value: '1250.0000' },
        outflows_next_30_days: { ...partial, value: '-800.0000' },
        days: [
          { date: '2026-11-03', inflows: { ...partial, value: '500.0000' }, outflows: { ...partial, value: '-200.0000' } },
          { date: '2026-11-04', inflows: { ...partial, value: '750.0000' }, outflows: { ...partial, value: '-600.0000' } },
        ],
      },
    }

    const { rerender } = render(<MockupDashboard screen="cashflow" snapshot={snapshot} />)
    expect(screen.getByText('ENTRADAS 30 DIAS').closest('article')).toHaveTextContent('1.250,00')
    expect(screen.getByText('SAÍDAS 30 DIAS').closest('article')).toHaveTextContent('-R$ 800,00')
    expect(screen.getByRole('img', { name: 'Entradas e saídas diárias previstas no fluxo de caixa Bling' })).toBeInTheDocument()

    rerender(<MockupDashboard screen="executive" snapshot={snapshot} />)
    expect(screen.getByRole('img', { name: 'Entradas e saídas diárias previstas no fluxo de caixa Bling' })).toBeInTheDocument()
  })

  it('identifica como parcial o valor incompleto sem mudar a identificação dos demais cards', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-01T12:00:00Z',
      data_state: 'PARTIAL',
      receivables: { value: '900.00', state: 'PARTIAL', source_system: 'BLING + NUVEM_PAGO' },
      receivables_due_next_90_days: { value: '125.00', state: 'PARTIAL', source_system: 'BLING + NUVEM_PAGO' },
      future_b2c: unavailable,
      receipts_month: unavailable,
      receivables_overdue: unavailable,
      payables: { value: '-500.00', state: 'CONFIRMED', source_system: 'BLING' },
      payables_due_next_90_days: { value: '-80.00', state: 'CONFIRMED', source_system: 'BLING' },
      payables_due_today: unavailable,
      payables_overdue: unavailable,
      payables_due_next_7_days: unavailable,
      payments_month: unavailable,
    }
    render(<MockupDashboard screen="executive" snapshot={snapshot} />)

    const receivables = screen.getByText('A RECEBER').closest('article')
    const payables = screen.getByText('A PAGAR').closest('article')
    expect(receivables).toHaveTextContent('125,00')
    expect(receivables).not.toHaveTextContent('900,00')
    expect(within(receivables as HTMLElement).getByText('Parcial · B2C + B2B')).toBeInTheDocument()
    expect(within(payables as HTMLElement).getByText('Bling')).toBeInTheDocument()
    expect(payables?.textContent).toContain('-')
    expect(payables).toHaveTextContent('80,00')
    expect(payables).not.toHaveTextContent('500,00')
    expect(within(payables as HTMLElement).queryByText(/Parcial/)).not.toBeInTheDocument()
  })

  it.each(cardRoutes)('o card %s abre sua tela de origem', async (label, target) => {
    const user = userEvent.setup()
    const onNavigate = vi.fn()
    render(<MockupDashboard screen="executive" onNavigate={onNavigate} />)

    const card = screen.getByText(label).closest('article')
    expect(card).not.toBeNull()
    await user.click(within(card as HTMLElement).getByRole('button', { name: `Ver detalhes: ${label}` }))

    expect(onNavigate).toHaveBeenCalledWith(target)
  })

  it('o gráfico e a distribuição abrem Fluxo de Caixa', async () => {
    const user = userEvent.setup()
    const onNavigate = vi.fn()
    render(<MockupDashboard screen="executive" onNavigate={onNavigate} />)

    const chart = screen.getByRole('heading', { name: 'FLUXO DE CAIXA PROJETADO' }).closest('section')
    expect(chart).not.toBeNull()
    await user.click(within(chart as HTMLElement).getByRole('button', { name: /Ver fluxo completo/ }))
    expect(onNavigate).toHaveBeenLastCalledWith('cashflow')

    const distribution = screen.getByRole('heading', { name: 'DISTRIBUIÇÃO DO SALDO' }).closest('section')
    expect(distribution).not.toBeNull()
    await user.click(within(distribution as HTMLElement).getByRole('button', { name: /Ver detalhes/ }))
    expect(onNavigate).toHaveBeenLastCalledWith('cashflow')
  })

  it('mantém o alerta aprovado e não mostra avisos de fonte/conciliação removidos', async () => {
    const user = userEvent.setup()
    const onNavigate = vi.fn()
    render(<MockupDashboard screen="executive" onNavigate={onNavigate} />)

    const alerts = [
      ['Saldo projetado abaixo da reserva mínima', /Ver fluxo/, 'cashflow'],
    ] as const

    expect(screen.queryByText('Obrigações vencidas aguardando fonte Bling')).not.toBeInTheDocument()
    expect(screen.queryByText('Recebimentos em atraso aguardando conciliação')).not.toBeInTheDocument()

    for (const [message, buttonName, target] of alerts) {
      const alert = screen.getByText(message).closest('article')
      expect(alert).not.toBeNull()
      await user.click(within(alert as HTMLElement).getByRole('button', { name: buttonName }))
      expect(onNavigate).toHaveBeenLastCalledWith(target)
    }
  })
})
