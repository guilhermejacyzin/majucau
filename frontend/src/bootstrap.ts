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

export type DashboardMetric = { value?: string; count?: number; state: 'CONFIRMED' | 'PROJECTED' | 'PARTIAL' | 'UNAVAILABLE'; source_system?: string }
export type DashboardReceivableRow = { customer?: string; origin: string; due_date?: string; gross_value?: string; net_value?: string; status: string }
export type DashboardPayableRow = { supplier?: string; document?: string; due_date: string; value: string; category?: string; status: string }
export type DashboardSnapshot = {
  as_of: string
  data_state: 'CONFIRMED' | 'PARTIAL' | 'UNAVAILABLE'
  receivables: DashboardMetric
  future_b2c: DashboardMetric
  receipts_month: DashboardMetric
  receivables_overdue: DashboardMetric
  payables: DashboardMetric
  payables_due_today: DashboardMetric
  payables_overdue: DashboardMetric
  payments_month: DashboardMetric
  receivable_rows?: DashboardReceivableRow[]
  payable_rows?: DashboardPayableRow[]
  error_code?: string
  message?: string
}
export type DashboardSnapshotSource = () => Promise<unknown>
export type DashboardSnapshotAdapter = { getSnapshot: () => Promise<DashboardSnapshot> }

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
export type BlingReceiptImportResult = {
  batch_id?: string
  status?: string
  records_read: number
  records_created: number
  records_updated: number
  records_failed: number
  ignored_count: number
  error_code?: string
  message?: string
}
export type BlingImportSource = (folder: string) => Promise<unknown>
export type BlingImportAdapter = { import: (folder: string) => Promise<BlingReceiptImportResult> }

export type NuvemPagoFutureImportPreview = {
  files: Array<{ name: string; sha256: string; receivable_count: number; rejected_row_count: number }>
  receivable_count: number
  error_count: number
  ignored_count: number
  issues?: Array<{ file: string; line: number; code: string; message: string }>
  error_code?: string
  message?: string
}
export type NuvemPagoFuturePreviewSource = (folder: string) => Promise<unknown>
export type NuvemPagoFuturePreviewAdapter = { preview: (folder: string) => Promise<NuvemPagoFutureImportPreview> }
export type NuvemPagoFutureImportResult = {
  batch_id?: string
  status?: string
  records_read: number
  records_created: number
  records_updated: number
  records_failed: number
  ignored_count: number
  error_code?: string
  message?: string
}
export type NuvemPagoFutureImportSource = (folder: string) => Promise<unknown>
export type NuvemPagoFutureImportAdapter = { import: (folder: string) => Promise<NuvemPagoFutureImportResult> }

export type BlingConfigInput = { client_id: string; redirect_uri: string; client_secret: string }
export type BlingConfigResult = { client_id?: string; redirect_uri?: string; secret_configured: boolean; status?: string; error_code?: string; message?: string }
export type BlingConfigSource = (input: BlingConfigInput) => Promise<unknown>
export type BlingConfigAdapter = { save: (input: BlingConfigInput) => Promise<BlingConfigResult> }
export type NuvemshopConfigInput = { app_id: string; redirect_uri: string; client_secret: string }
export type NuvemshopConfigResult = { app_id?: string; redirect_uri?: string; secret_configured: boolean; status?: string; error_code?: string; message?: string }
export type NuvemshopConfigSource = (input: NuvemshopConfigInput) => Promise<unknown>
export type NuvemshopConfigAdapter = { save: (input: NuvemshopConfigInput) => Promise<NuvemshopConfigResult> }
export type BlingOAuthStartResult = { session_id?: string; authorization_url?: string; status?: string; error_code?: string; message?: string }
export type BlingOAuthStatusResult = { session_id?: string; status?: string; error_code?: string; message?: string }
export type BlingOAuthTestResult = { status?: string; page_record_count: number; error_code?: string; message?: string }
export type BlingOAuthDisconnectResult = { status?: string; error_code?: string; message?: string }
export type BlingSyncInput = { page?: number; limit?: number; due_date_from?: string; due_date_to?: string; received_date_from?: string; received_date_to?: string; payment_date_from?: string; payment_date_to?: string; status?: string }
export type BlingSyncResourceResult = { status?: string; batch_id?: string; pages_read: number; records_read: number }
export type BlingSyncResult = { status?: string; receivables: BlingSyncResourceResult; payables: BlingSyncResourceResult; error_code?: string; message?: string }
export type BlingOAuthAdapter = {
  start: () => Promise<BlingOAuthStartResult>
  status: (sessionId: string) => Promise<BlingOAuthStatusResult>
  test: () => Promise<BlingOAuthTestResult>
  disconnect: () => Promise<BlingOAuthDisconnectResult>
  sync: (input: BlingSyncInput) => Promise<BlingSyncResult>
}

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

const unavailableDashboardSnapshot = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): DashboardSnapshot => {
  const asOf = new Date().toISOString()
  const unavailable = (): DashboardMetric => ({ state: 'UNAVAILABLE' })
  return { as_of: asOf, data_state: 'UNAVAILABLE', receivables: unavailable(), future_b2c: unavailable(), receipts_month: unavailable(), receivables_overdue: unavailable(), payables: unavailable(), payables_due_today: unavailable(), payables_overdue: unavailable(), payments_month: unavailable(), error_code: errorCode, message }
}

async function readWailsDashboardSnapshot(): Promise<unknown> {
  const method = window.go?.main?.App?.GetDashboardSnapshot
  if (typeof method !== 'function') return unavailableDashboardSnapshot()
  return method()
}

function isDashboardMetric(value: unknown): value is DashboardMetric {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<DashboardMetric>
  return typeof candidate.state === 'string' && ['CONFIRMED', 'PROJECTED', 'PARTIAL', 'UNAVAILABLE'].includes(candidate.state) && (candidate.value === undefined || typeof candidate.value === 'string') && (candidate.count === undefined || typeof candidate.count === 'number')
}

function isDashboardReceivableRow(value: unknown): value is DashboardReceivableRow {
  if (!value || typeof value !== 'object') return false
  const row = value as Partial<DashboardReceivableRow>
  return typeof row.origin === 'string' && typeof row.status === 'string' && (row.customer === undefined || typeof row.customer === 'string') && (row.due_date === undefined || typeof row.due_date === 'string') && (row.gross_value === undefined || typeof row.gross_value === 'string') && (row.net_value === undefined || typeof row.net_value === 'string')
}

function isDashboardPayableRow(value: unknown): value is DashboardPayableRow {
  if (!value || typeof value !== 'object') return false
  const row = value as Partial<DashboardPayableRow>
  return typeof row.due_date === 'string' && typeof row.value === 'string' && typeof row.status === 'string' && (row.supplier === undefined || typeof row.supplier === 'string') && (row.document === undefined || typeof row.document === 'string') && (row.category === undefined || typeof row.category === 'string')
}

function isDashboardSnapshot(value: unknown): value is DashboardSnapshot {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<DashboardSnapshot>
  return typeof candidate.as_of === 'string' && ['CONFIRMED', 'PARTIAL', 'UNAVAILABLE'].includes(candidate.data_state as string) && isDashboardMetric(candidate.receivables) && isDashboardMetric(candidate.future_b2c) && isDashboardMetric(candidate.receipts_month) && isDashboardMetric(candidate.receivables_overdue) && isDashboardMetric(candidate.payables) && isDashboardMetric(candidate.payables_due_today) && isDashboardMetric(candidate.payables_overdue) && isDashboardMetric(candidate.payments_month) && (candidate.receivable_rows === undefined || (Array.isArray(candidate.receivable_rows) && candidate.receivable_rows.every(isDashboardReceivableRow))) && (candidate.payable_rows === undefined || (Array.isArray(candidate.payable_rows) && candidate.payable_rows.every(isDashboardPayableRow)))
}

export function createDashboardSnapshotAdapter(source: DashboardSnapshotSource = readWailsDashboardSnapshot): DashboardSnapshotAdapter {
  return {
    getSnapshot: async () => {
      try {
        const snapshot = await source()
        return isDashboardSnapshot(snapshot) ? snapshot : unavailableDashboardSnapshot('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableDashboardSnapshot()
      }
    },
  }
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

const unavailableBlingImport = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingReceiptImportResult => ({ records_read: 0, records_created: 0, records_updated: 0, records_failed: 0, ignored_count: 0, error_code: errorCode, message })

async function readWailsBlingImport(folder: string): Promise<unknown> {
  const method = window.go?.main?.App?.ImportBlingReceipts
  if (typeof method !== 'function') return unavailableBlingImport()
  return method(folder)
}

function isBlingReceiptImportResult(value: unknown): value is BlingReceiptImportResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingReceiptImportResult>
  return typeof candidate.records_read === 'number' && typeof candidate.records_created === 'number' && typeof candidate.records_updated === 'number' && typeof candidate.records_failed === 'number' && typeof candidate.ignored_count === 'number'
}

export function createBlingImportAdapter(source: BlingImportSource = readWailsBlingImport): BlingImportAdapter {
  return {
    import: async (folder) => {
      try {
        const result = await source(folder)
        return isBlingReceiptImportResult(result) ? result : unavailableBlingImport('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingImport()
      }
    },
  }
}

const unavailableNuvemPagoFuturePreview = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): NuvemPagoFutureImportPreview => ({ files: [], receivable_count: 0, error_count: 0, ignored_count: 0, error_code: errorCode, message })

async function readWailsNuvemPagoFuturePreview(folder: string): Promise<unknown> {
  const method = window.go?.main?.App?.PreviewNuvemPagoFuture
  if (typeof method !== 'function') return unavailableNuvemPagoFuturePreview()
  return method(folder)
}

function isNuvemPagoFuturePreview(value: unknown): value is NuvemPagoFutureImportPreview {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<NuvemPagoFutureImportPreview>
  return Array.isArray(candidate.files) && typeof candidate.receivable_count === 'number' && typeof candidate.error_count === 'number' && typeof candidate.ignored_count === 'number'
}

export function createNuvemPagoFuturePreviewAdapter(source: NuvemPagoFuturePreviewSource = readWailsNuvemPagoFuturePreview): NuvemPagoFuturePreviewAdapter {
  return {
    preview: async (folder) => {
      try {
        const result = await source(folder)
        return isNuvemPagoFuturePreview(result) ? result : unavailableNuvemPagoFuturePreview('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableNuvemPagoFuturePreview()
      }
    },
  }
}

const unavailableNuvemPagoFutureImport = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): NuvemPagoFutureImportResult => ({ records_read: 0, records_created: 0, records_updated: 0, records_failed: 0, ignored_count: 0, error_code: errorCode, message })

async function readWailsNuvemPagoFutureImport(folder: string): Promise<unknown> {
  const method = window.go?.main?.App?.ImportNuvemPagoFuture
  if (typeof method !== 'function') return unavailableNuvemPagoFutureImport()
  return method(folder)
}

function isNuvemPagoFutureImportResult(value: unknown): value is NuvemPagoFutureImportResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<NuvemPagoFutureImportResult>
  return typeof candidate.records_read === 'number' && typeof candidate.records_created === 'number' && typeof candidate.records_updated === 'number' && typeof candidate.records_failed === 'number' && typeof candidate.ignored_count === 'number'
}

export function createNuvemPagoFutureImportAdapter(source: NuvemPagoFutureImportSource = readWailsNuvemPagoFutureImport): NuvemPagoFutureImportAdapter {
  return {
    import: async (folder) => {
      try {
        const result = await source(folder)
        return isNuvemPagoFutureImportResult(result) ? result : unavailableNuvemPagoFutureImport('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableNuvemPagoFutureImport()
      }
    },
  }
}

const unavailableBlingConfig = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingConfigResult => ({ secret_configured: false, error_code: errorCode, message })

async function readWailsBlingConfig(input: BlingConfigInput): Promise<unknown> {
  const method = window.go?.main?.App?.SaveBlingConfig
  if (typeof method !== 'function') return unavailableBlingConfig()
  return method(input)
}

function isBlingConfigResult(value: unknown): value is BlingConfigResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingConfigResult>
  return typeof candidate.secret_configured === 'boolean'
}

export function createBlingConfigAdapter(source: BlingConfigSource = readWailsBlingConfig): BlingConfigAdapter {
  return {
    save: async (input) => {
      try {
        const result = await source(input)
        return isBlingConfigResult(result) ? result : unavailableBlingConfig('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingConfig()
      }
    },
  }
}

const unavailableNuvemshopConfig = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): NuvemshopConfigResult => ({ secret_configured: false, error_code: errorCode, message })

async function readWailsNuvemshopConfig(input: NuvemshopConfigInput): Promise<unknown> {
  const method = window.go?.main?.App?.SaveNuvemshopConfig
  if (typeof method !== 'function') return unavailableNuvemshopConfig()
  return method(input)
}

function isNuvemshopConfigResult(value: unknown): value is NuvemshopConfigResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<NuvemshopConfigResult>
  return typeof candidate.secret_configured === 'boolean'
}

export function createNuvemshopConfigAdapter(source: NuvemshopConfigSource = readWailsNuvemshopConfig): NuvemshopConfigAdapter {
  return {
    save: async (input) => {
      try {
        const result = await source(input)
        return isNuvemshopConfigResult(result) ? result : unavailableNuvemshopConfig('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableNuvemshopConfig()
      }
    },
  }
}

const unavailableBlingOAuthStart = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingOAuthStartResult => ({ error_code: errorCode, message })
const unavailableBlingOAuthStatus = (sessionId: string, errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingOAuthStatusResult => ({ session_id: sessionId, error_code: errorCode, message })
const unavailableBlingOAuthTest = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingOAuthTestResult => ({ page_record_count: 0, error_code: errorCode, message })

async function readWailsBlingOAuthStart(): Promise<unknown> {
  const method = window.go?.main?.App?.StartBlingOAuth
  if (typeof method !== 'function') return unavailableBlingOAuthStart()
  return method()
}

async function readWailsBlingOAuthStatus(sessionId: string): Promise<unknown> {
  const method = window.go?.main?.App?.GetBlingOAuthStatus
  if (typeof method !== 'function') return unavailableBlingOAuthStatus(sessionId)
  return method(sessionId)
}

async function readWailsBlingOAuthTest(): Promise<unknown> {
  const method = window.go?.main?.App?.TestBlingConnection
  if (typeof method !== 'function') return unavailableBlingOAuthTest()
  return method()
}

async function readWailsBlingSync(input: BlingSyncInput): Promise<unknown> {
  const method = window.go?.main?.App?.SyncBling
  if (typeof method !== 'function') return unavailableBlingSync()
  return method(input)
}

const unavailableBlingSync = (errorCode = 'WORKER_UNAVAILABLE', message = 'O serviço local ainda não está disponível.'): BlingSyncResult => ({ status: 'FAILED', receivables: { pages_read: 0, records_read: 0 }, payables: { pages_read: 0, records_read: 0 }, error_code: errorCode, message })

function isBlingOAuthStartResult(value: unknown): value is BlingOAuthStartResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingOAuthStartResult>
  return (candidate.session_id === undefined || typeof candidate.session_id === 'string') && (candidate.status === undefined || typeof candidate.status === 'string')
}

function isBlingOAuthStatusResult(value: unknown): value is BlingOAuthStatusResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingOAuthStatusResult>
  return (candidate.session_id === undefined || typeof candidate.session_id === 'string') && (candidate.status === undefined || typeof candidate.status === 'string')
}

function isBlingOAuthTestResult(value: unknown): value is BlingOAuthTestResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingOAuthTestResult>
  return typeof candidate.page_record_count === 'number' && (candidate.status === undefined || typeof candidate.status === 'string')
}

async function readWailsBlingOAuthDisconnect(): Promise<unknown> {
  const method = window.go?.main?.App?.DisconnectBlingOAuth
  if (typeof method !== 'function') return unavailableBlingOAuthStatus('', 'WORKER_UNAVAILABLE', 'O serviço local ainda não está disponível.')
  return method()
}

function isBlingSyncResult(value: unknown): value is BlingSyncResult {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<BlingSyncResult>
  return Boolean(candidate.receivables && typeof candidate.receivables === 'object' && typeof candidate.receivables.pages_read === 'number' && typeof candidate.receivables.records_read === 'number' && candidate.payables && typeof candidate.payables === 'object' && typeof candidate.payables.pages_read === 'number' && typeof candidate.payables.records_read === 'number')
}

export function createBlingOAuthAdapter(sources: { start?: () => Promise<unknown>; status?: (sessionId: string) => Promise<unknown>; test?: () => Promise<unknown>; disconnect?: () => Promise<unknown>; sync?: (input: BlingSyncInput) => Promise<unknown> } = {}): BlingOAuthAdapter {
  return {
    start: async () => {
      try {
        const result = await (sources.start ?? readWailsBlingOAuthStart)()
        return isBlingOAuthStartResult(result) ? result : unavailableBlingOAuthStart('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingOAuthStart()
      }
    },
    status: async (sessionId) => {
      try {
        const result = await (sources.status ?? readWailsBlingOAuthStatus)(sessionId)
        return isBlingOAuthStatusResult(result) ? result : unavailableBlingOAuthStatus(sessionId, 'WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingOAuthStatus(sessionId)
      }
    },
    test: async () => {
      try {
        const result = await (sources.test ?? readWailsBlingOAuthTest)()
        return isBlingOAuthTestResult(result) ? result : unavailableBlingOAuthTest('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingOAuthTest()
      }
    },
    disconnect: async () => {
      try {
        const result = await (sources.disconnect ?? readWailsBlingOAuthDisconnect)()
        return isBlingOAuthStatusResult(result) ? result : unavailableBlingOAuthStatus('', 'WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingOAuthStatus('', 'WORKER_UNAVAILABLE', 'O serviço local ainda não está disponível.')
      }
    },
    sync: async (input) => {
      try {
        const result = await (sources.sync ?? readWailsBlingSync)(input)
        return isBlingSyncResult(result) ? result : unavailableBlingSync('WORKER_INVALID_RESPONSE', 'O serviço local respondeu em formato inválido.')
      } catch {
        return unavailableBlingSync()
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
    go?: { main?: { App?: { GetBootstrapState?: () => Promise<unknown>; GetDashboardSnapshot?: () => Promise<unknown>; SaveBlingConfig?: (input: BlingConfigInput) => Promise<unknown>; SaveNuvemshopConfig?: (input: NuvemshopConfigInput) => Promise<unknown>; StartBlingOAuth?: () => Promise<unknown>; GetBlingOAuthStatus?: (sessionId: string) => Promise<unknown>; TestBlingConnection?: () => Promise<unknown>; DisconnectBlingOAuth?: () => Promise<unknown>; SyncBling?: (input: BlingSyncInput) => Promise<unknown>; PreviewBlingReceipts?: (folder: string) => Promise<unknown>; ImportBlingReceipts?: (folder: string) => Promise<unknown>; PreviewNuvemPagoFuture?: (folder: string) => Promise<unknown>; ImportNuvemPagoFuture?: (folder: string) => Promise<unknown> } } }
}
}

