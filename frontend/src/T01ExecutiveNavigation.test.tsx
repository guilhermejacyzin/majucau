import { fireEvent, render, screen, within } from '@testing-library/react'
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
    expect(screen.getByText('ENTRADAS').closest('article')).not.toHaveTextContent('1.250,00')
    expect(screen.getByText('SAÍDAS').closest('article')).not.toHaveTextContent('-800,00')
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

    expect(screen.getByText('ENTRADAS').closest('article')).toHaveTextContent('30 dias')
    expect(screen.getByText('ENTRADAS').closest('article')).toHaveTextContent('60 dias')
    expect(screen.getByText('SAÍDAS').closest('article')).toHaveTextContent('30 dias')
    expect(screen.getByText('SAÍDAS').closest('article')).toHaveTextContent('60 dias')
    expect(screen.getByText('ENTRADAS').closest('article')).not.toHaveTextContent('125,00')
    expect(screen.getByText('SAÍDAS').closest('article')).not.toHaveTextContent('80,00')
  })

  it('mostra o relatório semanal em tabela e preserva o gráfico agregado da visão executiva', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const partial = { state: 'PARTIAL' as const, source_system: 'BLING' }
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-07T12:00:00Z',
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
        inflows_next_60_days: { ...partial, value: '3250.0000' },
        outflows_next_60_days: { ...partial, value: '-1800.0000' },
        days: [
          ...['2026-11-03', '2026-11-04', '2026-11-05', '2026-11-06'].map((date) => ({ date, inflows: { ...partial, value: '0.0000' }, outflows: { ...partial, value: '0.0000' } })),
          { date: '2026-11-07', inflows: { ...partial, value: '1250.0000' }, outflows: { ...partial, value: '-800.0000' } },
          { date: '2026-11-08', inflows: { ...partial, value: '1500.0000' }, outflows: { ...partial, value: '-600.0000' } },
          ...['2026-11-09', '2026-11-10', '2026-11-11', '2026-11-12', '2026-11-13', '2026-11-14'].map((date) => ({ date, inflows: { ...partial, value: '0.0000' }, outflows: { ...partial, value: '0.0000' } })),
        ],
        periods: [
          { from: '2026-11-07', through: '2026-11-07', inflows: { ...partial, value: '1250.0000' }, outflows: { ...partial, value: '-800.0000' }, closing_balance: { ...partial, value: '-83776.1500' } },
          { from: '2026-11-08', through: '2026-11-14', inflows: { ...partial, value: '1500.0000' }, outflows: { ...partial, value: '-600.0000' }, closing_balance: { ...partial, value: '-76683.2000' } },
        ],
        critical_week: { from: '2026-11-07', through: '2026-11-07', inflows: { ...partial, value: '1250.0000' }, outflows: { ...partial, value: '-800.0000' } },
      },
    }

    const { rerender } = render(<MockupDashboard screen="cashflow" snapshot={snapshot} />)
    expect(screen.getByText('ENTRADAS').closest('article')).toHaveTextContent('30 dias')
    expect(screen.getByText('ENTRADAS').closest('article')).toHaveTextContent('R$ 1.250,00 · Parcial')
    expect(screen.getByText('ENTRADAS').closest('article')).toHaveTextContent('60 dias')
    expect(screen.getByText('ENTRADAS').closest('article')).toHaveTextContent('R$ 3.250,00 · Parcial')
    expect(screen.getByText('SAÍDAS').closest('article')).toHaveTextContent('-R$ 800,00 · Parcial')
    expect(screen.getByText('SAÍDAS').closest('article')).toHaveTextContent('-R$ 1.800,00 · Parcial')
    const weeklyTable = screen.getByRole('region', { name: 'Tabela semanal do fluxo de caixa' })
    expect(weeklyTable).toHaveTextContent(/08\/11\/2026 a\s*14\/11\/2026/)
    expect(weeklyTable).not.toHaveTextContent(/07\/11\/2026/)
    expect(weeklyTable).not.toHaveTextContent('R$ 1.250,00 · Parcial')
    expect(weeklyTable).toHaveTextContent('R$ 1.500,00 · Parcial')
    expect(weeklyTable).not.toHaveTextContent('−25,0%')
    expect(weeklyTable).not.toHaveTextContent('+20,0%')
    const cashFlowChart = screen.getByRole('group', { name: 'Gráfico semanal de receitas, despesas e saldo resultante do fluxo de caixa Bling' })
    const cashFlowStartDate = cashFlowChart.querySelector('text[data-axis-date="start"]')
    const cashFlowEndDate = cashFlowChart.querySelector('text[data-axis-date="end"]')
    expect(cashFlowStartDate).toHaveTextContent('08/11')
    expect(cashFlowEndDate).toHaveTextContent('14/11')
    expect(cashFlowStartDate).toHaveAttribute('y', '153')
    expect(cashFlowEndDate).toHaveAttribute('y', '165')
    const chartViewBoxHeight = Number(cashFlowChart.getAttribute('viewBox')?.split(' ')[3])
    expect(Number(cashFlowEndDate?.getAttribute('y')) + 4).toBeLessThan(chartViewBoxHeight)
    const criticalWeek = screen.getByRole('heading', { name: 'SEMANA CRÍTICA' }).closest('section')
    expect(criticalWeek).toHaveTextContent('08/11–14/11')
    expect(criticalWeek).toHaveTextContent('-R$ 600,00 · Parcial')
    expect(criticalWeek).toHaveTextContent('R$ 1.500,00 · Parcial')
    expect(screen.getByText('Reserva mínima').parentElement).toHaveTextContent('R$ 50.000,00')
    for (const label of ['SALDO ATUAL', 'ENTRADAS', 'SAÍDAS']) {
      expect(within(screen.getByText(label).closest('article') as HTMLElement).queryByRole('button', { name: /Ver detalhes/ })).not.toBeInTheDocument()
    }
    expect(criticalWeek).not.toHaveTextContent('Posição de caixa da semana: SEM DADOS')

    rerender(<MockupDashboard screen="executive" snapshot={snapshot} />)
    const executiveChart = screen.getByRole('group', { name: 'Gráfico semanal de receitas, despesas e saldo resultante do fluxo de caixa Bling' })
    expect(executiveChart).toBeInTheDocument()
    expect(executiveChart.querySelector('text[data-axis-date="start"]')).toHaveTextContent('07/11')
    expect(executiveChart.querySelector('text[data-axis-date="end"]')).toHaveTextContent('07/11')
    expect(executiveChart.querySelectorAll('text[data-axis-date="start"]')).toHaveLength(2)
    expect(executiveChart.querySelectorAll('text[data-axis-date="end"]')).toHaveLength(2)
  })

  it('filtra a tabela e a semana crítica pelo período escolhido sem recalcular com ponto flutuante', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const partial = { state: 'PARTIAL' as const, source_system: 'BLING' }
    const days = Array.from({ length: 8 }, (_, index) => {
      const date = new Date(Date.UTC(2026, 10, 7 + index)).toISOString().slice(0, 10)
      const inflows = date === '2026-11-08' ? '1.2000' : date === '2026-11-13' ? '2.3000' : '0.0000'
      const outflows = date === '2026-11-10' ? '-0.0500' : date === '2026-11-08' ? '-0.2000' : '0.0000'
      return { date, inflows: { ...partial, value: inflows }, outflows: { ...partial, value: outflows } }
    })
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-07T12:00:00Z',
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
        inflows_next_30_days: unavailable,
        outflows_next_30_days: unavailable,
        inflows_next_60_days: unavailable,
        outflows_next_60_days: unavailable,
        days,
      },
    }
    render(<MockupDashboard screen="cashflow" snapshot={snapshot} />)

    expect(screen.getByRole('heading', { name: 'FLUXO DE CAIXA POR SEMANA' })).toBeInTheDocument()
    expect(screen.queryByText('RECEBIDO NO MÊS')).not.toBeInTheDocument()
    expect(screen.queryByText('SALDO TRANSFERÍVEL')).not.toBeInTheDocument()
    const from = screen.getByLabelText('Data inicial do fluxo de caixa')
    const through = screen.getByLabelText('Data final do fluxo de caixa')
    expect(from).toHaveAttribute('min', '2026-11-08')
    fireEvent.change(through, { target: { value: '2026-11-14' } })
    expect(from).toHaveValue('2026-11-08')
    expect(through).toHaveValue('2026-11-14')
    fireEvent.change(from, { target: { value: '2026-11-08' } })
    expect(through).toHaveValue('2026-11-14')

    const weekly = screen.getByRole('region', { name: 'Tabela semanal do fluxo de caixa' })
    expect(weekly).toHaveTextContent(/08\/11\/2026 a\s*14\/11\/2026/)
    expect(weekly).toHaveTextContent('R$ 3,50 · Parcial')
    expect(weekly).toHaveTextContent('-R$ 0,25 · Parcial')
    expect(weekly).toHaveTextContent('Saldo inicial')
    expect(weekly).not.toHaveTextContent('Ajuste de saldo')
    expect(weekly).toHaveTextContent('Saldo acumulado')
    expect(screen.getByRole('heading', { name: 'SEMANA CRÍTICA' }).closest('section')).toHaveTextContent('08/11–14/11')
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

  it('mantém o gráfico sem o botão removido e a distribuição abre Fluxo de Caixa', async () => {
    const user = userEvent.setup()
    const onNavigate = vi.fn()
    render(<MockupDashboard screen="executive" onNavigate={onNavigate} />)

    const chart = screen.getByRole('heading', { name: 'FLUXO DE CAIXA PROJETADO' }).closest('section')
    expect(chart).not.toBeNull()
    expect(within(chart as HTMLElement).queryByRole('button', { name: /Ver fluxo completo/ })).not.toBeInTheDocument()

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
