import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { MockupDashboard, MockupSidebar } from './mockupDashboard'

describe('retirada da tela T05', () => {
  it('não oferece Conciliação na navegação principal', () => {
    render(<MockupSidebar screen="executive" onNavigate={() => undefined} />)

    expect(screen.queryByRole('button', { name: 'Conciliação' })).not.toBeInTheDocument()
  })

  it('não mostra o antigo alerta da tela T05 na Visão Executiva', () => {
    render(<MockupDashboard screen="executive" />)

    expect(screen.queryByText('Conciliação pendente aguardando fonte oficial')).not.toBeInTheDocument()
  })

  it('mantém os cards de recebíveis aprovados na T03', () => {
    render(<MockupDashboard screen="receivables" />)

    for (const label of ['TOTAL A RECEBER', 'EM ATRASO', 'TOTAL RECEBIDO', 'PIX', 'BOLETOS', 'CARTÃO', 'AMBÍGUOS', 'CHARGEBACKS', 'DISTRIBUIÇÃO DOS RECEBÍVEIS']) {
      expect(screen.getByText(label)).toBeInTheDocument()
    }
    expect(screen.queryByText('CONCILIAÇÃO')).not.toBeInTheDocument()
    expect(screen.queryByText('RECEBÍVEIS')).not.toBeInTheDocument()
    for (const retiredLabel of ['PRIORIDADE', 'FALLBACK', 'REUSO BLOQUEADO', 'SEM ID BLING']) {
      expect(screen.queryByText(retiredLabel)).not.toBeInTheDocument()
    }
  })
})
