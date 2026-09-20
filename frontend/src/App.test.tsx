import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'
import App from './App'
import { tooltipCatalog } from './tooltips'
import type { BlingPreviewAdapter, BootstrapAdapter, BootstrapState } from './bootstrap'

const connectedBootstrap: BootstrapState = { app_version: '0.1.0-g1', worker: { service: 'majucau-worker', version: 'test-worker', state: 'OK', checked_at: '2026-08-16T12:00:00Z' }, integrations: [{ provider: 'BLING', status: 'CONNECTED', last_success_at: '2026-08-16T11:58:00Z', last_attempt_at: '2026-08-16T11:58:00Z' }, { provider: 'NUVEMSHOP', status: 'PARTIALLY_AVAILABLE' }, { provider: 'NUVEM_PAGO', status: 'UNAVAILABLE' }], checked_at: '2026-08-16T12:00:00Z' }
const adapterFor = (state: BootstrapState): BootstrapAdapter => ({ getState: () => Promise.resolve(state) })

describe('shell financeiro', () => {
  it('exibe todos os cards oficiais sem números inventados', async () => {
    render(<App bootstrapAdapter={adapterFor(connectedBootstrap)} />)
    await screen.findByRole('heading', { name: 'Visão Executiva' })
    for (const label of ['Saldo Inicial do Dia', 'Saldo Final Projetado Hoje', 'Saldo Projetado em 30 Dias', 'Saldo Projetado em 60 Dias', 'Menor Saldo Projetado — 60 Dias', 'Reserva Mínima', 'Valor Máximo para Aplicação', 'A Receber B2C — Nuvem', 'A Receber B2B — Bling', 'Total a Receber', 'Recebido no Mês', 'A Pagar', 'Obrigações Vencidas', 'Pago no Mês', 'DRE / P&L', 'EBITDA', 'Orçado x Realizado', 'Forecast', 'Capital de Giro', 'Alertas']) expect(screen.getByText(label)).toBeInTheDocument()
    expect(screen.getAllByText('Sem dados confirmados')).toHaveLength(20)
    expect(screen.getByText(/Nenhum valor foi estimado/)).toBeInTheDocument()
  })

  it('navega por teclado até integrações e mantém campos com labels visíveis', async () => {
    const user = userEvent.setup()
    render(<App bootstrapAdapter={adapterFor(connectedBootstrap)} />)
    await screen.findByRole('heading', { name: 'Visão Executiva' })
    await user.click(screen.getByRole('button', { name: 'Integrações' }))
    expect(screen.getByRole('heading', { name: 'Integrações' })).toBeInTheDocument()
    expect(screen.getByLabelText(/Client ID/)).toBeInTheDocument()
    expect(screen.getAllByLabelText(/Client Secret/)).toHaveLength(2)
    expect(screen.getByRole('heading', { name: 'Nuvem Pago' })).toBeInTheDocument()
    expect(screen.getByText('Indisponível para confirmação financeira')).toBeInTheDocument()
    expect(screen.getByText('Tarifário configurado')).toBeInTheDocument()
    expect(screen.getByText('4,49% + R$ 0,35')).toBeInTheDocument()
    expect(screen.getByText('Na hora')).toBeInTheDocument()
  })

  it('navega pelo aviso e mantém ações sem worker desabilitadas', async () => {
    const user = userEvent.setup()
    render(<App bootstrapAdapter={adapterFor(connectedBootstrap)} />)
    await screen.findByRole('heading', { name: 'Visão Executiva' })
    await user.click(screen.getByRole('button', { name: /Ir para Integrações/ }))
    expect(screen.getByRole('heading', { name: 'Integrações' })).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: 'Conectar' })[0]).toBeDisabled()
    expect(screen.getAllByText(/ações de credencial e sincronização desabilitadas/)).toHaveLength(2)
  })

  it('mantém cobertura estrutural de ajuda em todos os controles funcionais', async () => {
    const { container } = render(<App bootstrapAdapter={adapterFor(connectedBootstrap)} />)
    await screen.findByRole('heading', { name: 'Visão Executiva' })
    const controls = [...container.querySelectorAll('button, input, select')]
    expect(controls.length).toBeGreaterThan(10)
    expect(controls.every((control) => control.closest('.tooltip'))).toBe(true)
    expect(Object.keys(tooltipCatalog).length).toBeGreaterThan(40)
    const user = userEvent.setup()
    await user.click(screen.getByRole('button', { name: 'Integrações' }))
    const integrationControls = [...container.querySelectorAll('button, input, select')]
    expect(integrationControls.every((control) => control.closest('.tooltip'))).toBe(true)
    const secrets = screen.getAllByLabelText(/Client Secret/)
    expect(secrets.every((secret) => secret.getAttribute('autocomplete') === 'new-password')).toBe(true)
  })

  it('mostra loading enquanto o bootstrap ainda não resolve', () => {
    const adapter: BootstrapAdapter = { getState: () => new Promise(() => undefined) }
    render(<App bootstrapAdapter={adapter} />)
    expect(screen.getByRole('heading', { name: 'Carregando estado local' })).toBeInTheDocument()
  })

  it('reflete bootstrap real e falha explicitamente no worker', async () => {
    const unavailable: BootstrapState = { ...connectedBootstrap, worker: { ...connectedBootstrap.worker, state: 'UNAVAILABLE' }, integrations: [{ provider: 'BLING', status: 'NOT_CONFIGURED' }, { provider: 'NUVEMSHOP', status: 'AUTH_ERROR' }, { provider: 'NUVEM_PAGO', status: 'UNAVAILABLE' }], error_code: 'WORKER_UNAVAILABLE', message: 'O serviço local ainda não está disponível.' }
    render(<App bootstrapAdapter={adapterFor(unavailable)} />)
    await screen.findByText(/WORKER_UNAVAILABLE/)
    await userEvent.click(screen.getByRole('button', { name: 'Integrações' }))
    expect(screen.getByText('Erro de autorização')).toBeInTheDocument()
    expect(screen.getByText('Indisponível para confirmação financeira')).toBeInTheDocument()
    expect(screen.getAllByText(/ações de credencial e sincronização desabilitadas/)).toHaveLength(2)
  })

  it('valida a pasta de recebimentos e mostra somente o resumo sanitizado', async () => {
    const user = userEvent.setup()
    const blingPreviewAdapter: BlingPreviewAdapter = { preview: async () => ({ files: [{ name: 'recebidos.csv', sha256: 'a'.repeat(64), receipt_count: 3, error_count: 1 }], receipt_count: 3, error_count: 1, ignored_count: 0, issues: [{ file: 'recebidos.csv', line: 5, code: 'NOT_PAID', message: 'linha não importada' }] }) }
    render(<App bootstrapAdapter={adapterFor(connectedBootstrap)} blingPreviewAdapter={blingPreviewAdapter} />)
    await screen.findByRole('heading', { name: 'Visão Executiva' })
    await user.click(screen.getByRole('button', { name: 'Integrações' }))
    await user.type(screen.getByLabelText('Pasta dos relatórios CSV'), 'C:\\imports\\bling')
    await user.click(screen.getByRole('button', { name: 'Validar pasta' }))
    expect(await screen.findByText('3 recebimento(s) válido(s)')).toBeInTheDocument()
    expect(screen.getByText(/1 linha\(s\) com atenção/)).toBeInTheDocument()
  })
})
