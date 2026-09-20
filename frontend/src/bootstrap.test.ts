import { describe, expect, it } from 'vitest'
import { createBlingPreviewAdapter, createBootstrapAdapter, unavailableBootstrap } from './bootstrap'

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

  it('valida o resumo do importador sem aceitar linhas financeiras', async () => {
    const adapter = createBlingPreviewAdapter(async (folder) => ({
      files: [{ name: folder, sha256: 'a'.repeat(64), receipt_count: 2, error_count: 1 }],
      receipt_count: 2,
      error_count: 1,
      ignored_count: 0,
    }))
    await expect(adapter.preview('C:\\imports')).resolves.toMatchObject({ receipt_count: 2, files: [{ receipt_count: 2 }] })
  })

  it('converte erro do worker em estado seguro para a tela', async () => {
    const adapter = createBlingPreviewAdapter(async () => { throw new Error('customer name leaked') })
    const preview = await adapter.preview('C:\\imports')
    expect(preview.error_code).toBe('WORKER_UNAVAILABLE')
    expect(preview.message).not.toContain('customer name leaked')
  })
})
