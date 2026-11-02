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
  it('identifica como parcial o valor incompleto sem mudar a identificação dos demais cards', () => {
    const unavailable = { state: 'UNAVAILABLE' as const }
    const snapshot: DashboardSnapshot = {
      as_of: '2026-11-01T12:00:00Z',
      data_state: 'PARTIAL',
      receivables: { value: '125.00', state: 'PARTIAL', source_system: 'BLING + NUVEM_PAGO' },
      future_b2c: unavailable,
      receipts_month: unavailable,
      receivables_overdue: unavailable,
      payables: { value: '-80.00', state: 'CONFIRMED', source_system: 'BLING' },
      payables_due_today: unavailable,
      payables_overdue: unavailable,
      payables_due_next_7_days: unavailable,
      payments_month: unavailable,
    }
    render(<MockupDashboard screen="executive" snapshot={snapshot} />)

    const receivables = screen.getByText('A RECEBER').closest('article')
    const payables = screen.getByText('A PAGAR').closest('article')
    expect(receivables).toHaveTextContent('125,00')
    expect(within(receivables as HTMLElement).getByText('Parcial · B2C + B2B')).toBeInTheDocument()
    expect(within(payables as HTMLElement).getByText('Bling')).toBeInTheDocument()
    expect(payables?.textContent).toContain('-')
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

  it('cada alerta abre a tela que já apresenta aquele resultado', async () => {
    const user = userEvent.setup()
    const onNavigate = vi.fn()
    render(<MockupDashboard screen="executive" onNavigate={onNavigate} />)

    const alerts = [
      ['Saldo projetado abaixo da reserva mínima', /Ver fluxo/, 'cashflow'],
      ['Obrigações vencidas aguardando fonte Bling', /Ver pagar/, 'payables'],
      ['Recebimentos em atraso aguardando conciliação', /Ver receber/, 'receivables'],
    ] as const

    for (const [message, buttonName, target] of alerts) {
      const alert = screen.getByText(message).closest('article')
      expect(alert).not.toBeNull()
      await user.click(within(alert as HTMLElement).getByRole('button', { name: buttonName }))
      expect(onNavigate).toHaveBeenLastCalledWith(target)
    }
  })
})
