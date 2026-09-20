export type WorkerHealthState = 'OK' | 'DEGRADED' | 'UNAVAILABLE'
export type IntegrationProvider = 'BLING' | 'NUVEMSHOP' | 'NUVEM_PAGO'
export type IntegrationState = 'NOT_CONFIGURED' | 'AUTHORIZING' | 'CONNECTED' | 'TOKEN_EXPIRING' | 'AUTH_ERROR' | 'SCHEMA_MISMATCH' | 'SYNCING' | 'STALE' | 'PARTIALLY_AVAILABLE' | 'UNAVAILABLE'

export type BootstrapIntegration = {
  provider: IntegrationProvider
  status: IntegrationState
  last_success_at?: string
  last_attempt_at?: string
  error_code?: string
}

export type BootstrapState = {
  app_version: string
  worker: { service: string; version: string; state: WorkerHealthState; checked_at: string; dependencies?: Array<{ name: string; state: WorkerHealthState; detail?: string }> }
  integrations: BootstrapIntegration[]
  error_code?: string
  message?: string
  checked_at: string
}

export type BlingReceiptImportPreview = {
  files: Array<{ name: string; sha256: string; receipt_count: number; error_count: number }>
  receipt_count: number
  error_count: number
  ignored_count: number
  issues?: Array<{ file: string; line: number; code: string; message: string }>
  error_code?: string
  message?: string
}

export type BootstrapSource = () => Promise<unknown>
export type BootstrapAdapter = { getState: () => Promise<BootstrapState> }
export type BlingPreviewSource = (folder: string) => Promise<unknown>
export type BlingPreviewAdapter = { preview: (folder: string) => Promise<BlingReceiptImportPreview> }

const unavailableIntegration = (provider: IntegrationProvider, status: IntegrationState = 'NOT_CONFIGURED'): BootstrapIntegration => ({ provider, status })

export function unavailableBootstrap(errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BootstrapState {
  const checkedAt = new Date().toISOString()
  return {
    app_version: 'unknown',
    worker: { service: 'majucau-worker', version: 'unknown', state: 'UNAVAILABLE', checked_at: checkedAt },
    integrations: [unavailableIntegration('BLING'), unavailableIntegration('NUVEMSHOP'), unavailableIntegration('NUVEM_PAGO', 'UNAVAILABLE')],
    error_code: errorCode,
    message,
    checked_at: checkedAt,
  }
}

function isBootstrapState(value: unknown): value is BootstrapState {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BootstrapState>
  const workerStates = ['OK', 'DEGRADED', 'UNAVAILABLE']
  const integrationStates = ['NOT_CONFIGURED', 'AUTHORIZING', 'CONNECTED', 'TOKEN_EXPIRING', 'AUTH_ERROR', 'SCHEMA_MISMATCH', 'SYNCING', 'STALE', 'PARTIALLY_AVAILABLE', 'UNAVAILABLE']
  const providers = ['BLING', 'NUVEMSHOP', 'NUVEM_PAGO']
  return Boolean(candidate.worker && typeof candidate.worker === 'object' && workerStates.includes(candidate.worker.state as string) && Array.isArray(candidate.integrations) && candidate.integrations.every((item) => Boolean(item && providers.includes((item as BootstrapIntegration).provider) && integrationStates.includes((item as BootstrapIntegration).status))))
}

async function readWailsBootstrap(): Promise<unknown> {
  const wailsApp = window.go?.main?.App
  if (typeof wailsApp?.GetBootstrapState !== 'function') return unavailableBootstrap()
  return wailsApp.GetBootstrapState()
}

const unavailableBlingPreview = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingReceiptImportPreview => ({ files: [], receipt_count: 0, error_count: 0, ignored_count: 0, error_code: errorCode, message })

async function readWailsBlingPreview(folder: string): Promise<unknown> {
  const method = window.go?.main?.App?.PreviewBlingReceipts
  if (typeof method !== 'function') return unavailableBlingPreview()
  return method(folder)
}

function isBlingReceiptImportPreview(value: unknown): value is BlingReceiptImportPreview {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingReceiptImportPreview>
  return Array.isArray(candidate.files) && typeof candidate.receipt_count === 'number' && typeof candidate.error_count === 'number' && typeof candidate.ignored_count === 'number'
}

export function createBlingPreviewAdapter(source: BlingPreviewSource = readWailsBlingPreview): BlingPreviewAdapter {
  return {
    preview: async (folder) => {
      try {
        const result = await source(folder)
        return isBlingReceiptImportPreview(result) ? result : unavailableBlingPreview('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingPreview()
      }
    },
  }
}

export function createBootstrapAdapter(source: BootstrapSource = readWailsBootstrap): BootstrapAdapter {
  return {
    getState: async () => {
      try {
        const state = await source()
        return isBootstrapState(state) ? state : unavailableBootstrap('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBootstrap()
      }
    },
  }
}

declare global {
  interface Window {
    go?: { main?: { App?: { GetBootstrapState?: () => Promise<unknown>; PreviewBlingReceipts?: (folder: string) => Promise<unknown> } } }
  }
}
