import { describe, expect, it } from 'vitest'
import { createBootstrapAdapter, unavailableBootstrap } from './bootstrap'

describe('bootstrap adapter', () => {
  it('consome contrato tipado sem acoplar ao runtime Wails', async () => {
    const state = unavailableBootstrap('TEST_STATE')
    const adapter = createBootstrapAdapter(async () => state)
    await expect(adapter.getState()).resolves.toMatchObject({ error_code: 'TEST_STATE', worker: { state: 'UNAVAILABLE' } })
  })

  it('converte falha do runtime em WORKER_UNAVAILABLE sem vazar detalhes', async () => {
    const adapter = createBootstrapAdapter(async () => { throw new Error('secret transport detail') })
    const state = await adapter.getState()
    expect(state.error_code).toBe('WORKER_UNAVAILABLE')
    expect(state.message).not.toContain('secret transport detail')
  })

  it('rejeita resposta inválida fail-closed', async () => {
    const adapter = createBootstrapAdapter(async () => ({ worker: { state: 'NOT_A_STATE' } }))
    await expect(adapter.getState()).resolves.toMatchObject({ error_code: 'WORKER_INVALID_RESPONSE' })
  })
})
