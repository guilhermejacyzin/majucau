import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import type { DashboardMetric, DashboardSnapshot } from './bootstrap'
import { MockupDashboard } from './mockupDashboard'

const unavailable: DashboardMetric = { state: 'UNAVAILABLE' }

const snapshot: DashboardSnapshot = {
  as_of: '2026-11-15T12:00:00Z',
  data_state: 'PARTIAL',
  receivables: { value: '500.00', count: 2, state: 'CONFIRMED' },
  future_b2c: unavailable,
  receipts_month: { value: '125.00', count: 1, state: 'CONFIRMED' },
  receivables_overdue: { value: '100.00', count: 1, state: 'CONFIRMED' },
  payables: unavailable,
  payables_due_today: unavailable,
  payables_overdue: unavailable,
  payables_due_next_7_days: unavailable,
  payments_month: unavailable,
  receivable_rows: [
    { customer: 'Cliente vencido', origin: 'Bling', due_date: '2026-11-14', gross_value: '100.00', net_value: '95.00', status: 'OPEN' },
    { customer: 'Cliente futuro', origin: 'Nuvem Pago', due_date: '2026-11-20', gross_value: '400.00', net_value: '380.00', status: 'PROJECTED' },
  ],
  cash_movement_rows: [{ date: '2026-11-12', account: 'Itaú', origin: 'Recebimento', value: '125.00' }],
}

describe('T03 detalhes dos recebíveis', () => {
  it('abre a lista pelo card e mantém a tabela fora da tela principal', () => {
    render(<MockupDashboard screen="receivables" snapshot={snapshot} />)

    expect(screen.queryByRole('table')).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Ver detalhes: TOTAL A RECEBER' }))

    const dialog = screen.getByRole('dialog', { name: 'TOTAL A RECEBER' })
    expect(within(dialog).getByText('Cliente vencido')).toBeInTheDocument()
    expect(within(dialog).getByText('Cliente futuro')).toBeInTheDocument()
  })

  it('filtra o detalhe de atraso pela data-base', () => {
    render(<MockupDashboard screen="receivables" snapshot={snapshot} />)
    fireEvent.click(screen.getByRole('button', { name: 'Ver detalhes: EM ATRASO' }))

    const dialog = screen.getByRole('dialog', { name: 'EM ATRASO' })
    expect(within(dialog).getByText('Cliente vencido')).toBeInTheDocument()
    expect(within(dialog).queryByText('Cliente futuro')).not.toBeInTheDocument()
  })

  it('abre Total recebido com a origem bancária e o mesmo valor do card', () => {
    render(<MockupDashboard screen="receivables" snapshot={snapshot} />)
    expect(screen.getByText((_, element) => element?.classList.contains('mocap-empty-value') === true && element.textContent?.includes('125,00') === true)).toBeInTheDocument()
    fireEvent.click(screen.getByRole('button', { name: 'Ver detalhes: TOTAL RECEBIDO' }))

    const dialog = screen.getByRole('dialog', { name: 'TOTAL RECEBIDO' })
    expect(within(dialog).getByText('Itaú')).toBeInTheDocument()
    expect(within(dialog).getByText('Recebimento')).toBeInTheDocument()
    expect(within(dialog).getByText(/R\$\s*125,00/)).toBeInTheDocument()
  })

  it('permite fechar o detalhe', () => {
    render(<MockupDashboard screen="receivables" snapshot={snapshot} />)
    fireEvent.click(screen.getByRole('button', { name: 'Ver detalhes: TOTAL A RECEBER' }))
    fireEvent.click(screen.getByRole('button', { name: 'Fechar detalhes' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('explica o gráfico e não inventa fatias sem dados de status', () => {
    render(<MockupDashboard screen="receivables" snapshot={snapshot} />)

    expect(screen.getByText('Mostra como o total de recebíveis se divide entre situações de confirmação.')).toBeInTheDocument()
    expect(document.querySelector('.mocap-donut')).not.toHaveClass('has-data')
    fireEvent.click(screen.getByRole('button', { name: /O que significa cada situação/ }))
    expect(screen.getByText(/conciliado com a fonte oficial e validado/)).toBeInTheDocument()
    expect(screen.getByText(/Cada fatia representará a parte do total/)).toBeInTheDocument()
  })

  it('abre detalhes sem remover os demais cards da tela', () => {
    render(<MockupDashboard screen="receivables" snapshot={snapshot} />)

    for (const label of ['AMBÍGUOS', 'CHARGEBACKS', 'PIX', 'BOLETOS', 'CARTÃO']) {
      fireEvent.click(screen.getByRole('button', { name: `Ver detalhes: ${label}` }))
      expect(screen.getByRole('dialog', { name: label })).toBeInTheDocument()
      fireEvent.click(screen.getByRole('button', { name: 'Fechar detalhes' }))
    }
  })
})
