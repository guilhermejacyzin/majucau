import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { axe } from 'vitest-axe'
import { describe, expect, it } from 'vitest'
import { Tooltip } from './components/Tooltip'

describe('Tooltip', () => {
  it('abre em foco, associa o texto por aria-describedby e fecha com Escape', async () => {
    const user = userEvent.setup()
    render(<Tooltip id="treasury.openingBalance"><button type="button">Saldo Inicial do Dia</button></Tooltip>)
    const trigger = screen.getByRole('button', { name: 'Saldo Inicial do Dia' })
    await user.tab()
    expect(screen.getByRole('tooltip')).toHaveTextContent('Dinheiro conciliado disponível')
    expect(trigger).toHaveAttribute('aria-describedby')
    await user.keyboard('{Escape}')
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()
  })

  it('abre com hover e remove a descrição ao sair', async () => {
    const user = userEvent.setup()
    render(<Tooltip id="status.provisional"><span>PROVISIONAL</span></Tooltip>)
    const trigger = screen.getByText('PROVISIONAL')
    await user.hover(trigger)
    expect(screen.getByRole('tooltip')).toHaveTextContent('O valor pode mudar')
    await user.unhover(trigger)
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()
  })

  it('não tem violações axe no uso básico', async () => {
    const { container } = render(<Tooltip id="treasury.scenario"><button type="button">Cenário</button></Tooltip>)
    expect((await axe(container)).violations).toHaveLength(0)
  })
})
