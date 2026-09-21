export namespace application {
	
	export class BlingConfigRequest {
	    client_id: string;
	    redirect_uri: string;
	    client_secret: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingConfigRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.client_id = source["client_id"];
	        this.redirect_uri = source["redirect_uri"];
	        this.client_secret = source["client_secret"];
	    }
	}
	export class BlingConfigResponse {
	    client_id?: string;
	    redirect_uri?: string;
	    secret_configured: boolean;
	    status?: string;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.client_id = source["client_id"];
	        this.redirect_uri = source["redirect_uri"];
	        this.secret_configured = source["secret_configured"];
	        this.status = source["status"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class BlingOAuthStartResponse {
	    session_id?: string;
	    authorization_url?: string;
	    status?: string;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingOAuthStartResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session_id = source["session_id"];
	        this.authorization_url = source["authorization_url"];
	        this.status = source["status"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class BlingOAuthStatusResponse {
	    session_id?: string;
	    status?: string;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingOAuthStatusResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.session_id = source["session_id"];
	        this.status = source["status"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class BlingOAuthTestResponse {
	    status?: string;
	    page_record_count: number;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingOAuthTestResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.page_record_count = source["page_record_count"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class BlingOAuthDisconnectResponse {
	    status?: string;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingOAuthDisconnectResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class BlingReceiptImportFile {
	    name: string;
	    sha256: string;
	    receipt_count: number;
	    error_count: number;
	
	    static createFrom(source: any = {}) {
	        return new BlingReceiptImportFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.sha256 = source["sha256"];
	        this.receipt_count = source["receipt_count"];
	        this.error_count = source["error_count"];
	    }
	}
	export class BlingReceiptImportIssue {
	    file: string;
	    line: number;
	    code: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingReceiptImportIssue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.file = source["file"];
	        this.line = source["line"];
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
	export class BlingReceiptImportPreview {
	    files: BlingReceiptImportFile[];
	    receipt_count: number;
	    error_count: number;
	    ignored_count: number;
	    issues?: BlingReceiptImportIssue[];
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingReceiptImportPreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.files = this.convertValues(source["files"], BlingReceiptImportFile);
	        this.receipt_count = source["receipt_count"];
	        this.error_count = source["error_count"];
	        this.ignored_count = source["ignored_count"];
	        this.issues = this.convertValues(source["issues"], BlingReceiptImportIssue);
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BlingReceiptImportResult {
	    batch_id?: string;
	    status?: string;
	    records_read: number;
	    records_created: number;
	    records_updated: number;
	    records_failed: number;
	    ignored_count: number;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingReceiptImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.batch_id = source["batch_id"];
	        this.status = source["status"];
	        this.records_read = source["records_read"];
	        this.records_created = source["records_created"];
	        this.records_updated = source["records_updated"];
	        this.records_failed = source["records_failed"];
	        this.ignored_count = source["ignored_count"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class BlingSyncRequest {
	    page?: number;
	    limit?: number;
	    due_date_from?: string;
	    due_date_to?: string;
	    received_date_from?: string;
	    received_date_to?: string;
	    payment_date_from?: string;
	    payment_date_to?: string;
	    status?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingSyncRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.page = source["page"];
	        this.limit = source["limit"];
	        this.due_date_from = source["due_date_from"];
	        this.due_date_to = source["due_date_to"];
	        this.received_date_from = source["received_date_from"];
	        this.received_date_to = source["received_date_to"];
	        this.payment_date_from = source["payment_date_from"];
	        this.payment_date_to = source["payment_date_to"];
	        this.status = source["status"];
	    }
	}
	export class BlingSyncResourceResult {
	    status?: string;
	    batch_id?: string;
	    pages_read: number;
	    records_read: number;
	
	    static createFrom(source: any = {}) {
	        return new BlingSyncResourceResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.batch_id = source["batch_id"];
	        this.pages_read = source["pages_read"];
	        this.records_read = source["records_read"];
	    }
	}
	export class BlingSyncResponse {
	    status?: string;
	    receivables: BlingSyncResourceResult;
	    payables: BlingSyncResourceResult;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlingSyncResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.receivables = this.convertValues(source["receivables"], BlingSyncResourceResult);
	        this.payables = this.convertValues(source["payables"], BlingSyncResourceResult);
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DependencyHealth {
	    name: string;
	    state: string;
	    detail?: string;
	
	    static createFrom(source: any = {}) {
	        return new DependencyHealth(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.state = source["state"];
	        this.detail = source["detail"];
	    }
	}
	export class HealthResponse {
	    service: string;
	    version: string;
	    state: string;
	    // Go type: time
	    checked_at: any;
	    dependencies?: DependencyHealth[];
	
	    static createFrom(source: any = {}) {
	        return new HealthResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.service = source["service"];
	        this.version = source["version"];
	        this.state = source["state"];
	        this.checked_at = this.convertValues(source["checked_at"], null);
	        this.dependencies = this.convertValues(source["dependencies"], DependencyHealth);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class IntegrationStatus {
	    provider: string;
	    status: string;
	    // Go type: time
	    last_success_at?: any;
	    // Go type: time
	    last_attempt_at?: any;
	    error_code?: string;
	
	    static createFrom(source: any = {}) {
	        return new IntegrationStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.provider = source["provider"];
	        this.status = source["status"];
	        this.last_success_at = this.convertValues(source["last_success_at"], null);
	        this.last_attempt_at = this.convertValues(source["last_attempt_at"], null);
	        this.error_code = source["error_code"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class NuvemshopConfigRequest {
	    app_id: string;
	    redirect_uri: string;
	    client_secret: string;
	
	    static createFrom(source: any = {}) {
	        return new NuvemshopConfigRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_id = source["app_id"];
	        this.redirect_uri = source["redirect_uri"];
	        this.client_secret = source["client_secret"];
	    }
	}
	export class NuvemshopConfigResponse {
	    app_id?: string;
	    redirect_uri?: string;
	    secret_configured: boolean;
	    status?: string;
	    error_code?: string;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new NuvemshopConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_id = source["app_id"];
	        this.redirect_uri = source["redirect_uri"];
	        this.secret_configured = source["secret_configured"];
	        this.status = source["status"];
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }
	}
	export class DashboardMetric {
	    value?: string;
	    count?: number;
	    state: string;
	    source_system?: string;

	    static createFrom(source: any = {}) {
	        return new DashboardMetric(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.count = source["count"];
	        this.state = source["state"];
	        this.source_system = source["source_system"];
	    }
	}
	export class DashboardReceivableRow {
	    customer?: string;
	    origin: string;
	    due_date?: string;
	    gross_value?: string;
	    net_value?: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new DashboardReceivableRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.customer = source["customer"];
	        this.origin = source["origin"];
	        this.due_date = source["due_date"];
	        this.gross_value = source["gross_value"];
	        this.net_value = source["net_value"];
	        this.status = source["status"];
	    }
	}
	export class DashboardPayableRow {
	    supplier?: string;
	    document?: string;
	    due_date: string;
	    value: string;
	    category?: string;
	    status: string;

	    static createFrom(source: any = {}) {
	        return new DashboardPayableRow(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.supplier = source["supplier"];
	        this.document = source["document"];
	        this.due_date = source["due_date"];
	        this.value = source["value"];
	        this.category = source["category"];
	        this.status = source["status"];
	    }
	}
	export class DashboardSnapshot {
	    as_of: any;
	    data_state: string;
	    receivables: DashboardMetric;
	    future_b2c: DashboardMetric;
	    receipts_month: DashboardMetric;
	    receivables_overdue: DashboardMetric;
	    payables: DashboardMetric;
	    payables_due_today: DashboardMetric;
	    payables_overdue: DashboardMetric;
	    payments_month: DashboardMetric;
	    receivable_rows?: DashboardReceivableRow[];
	    payable_rows?: DashboardPayableRow[];
	    error_code?: string;
	    message?: string;

	    static createFrom(source: any = {}) {
	        return new DashboardSnapshot(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.as_of = this.convertValues(source["as_of"], null);
	        this.data_state = source["data_state"];
	        this.receivables = this.convertValues(source["receivables"], DashboardMetric);
	        this.future_b2c = this.convertValues(source["future_b2c"], DashboardMetric);
	        this.receipts_month = this.convertValues(source["receipts_month"], DashboardMetric);
	        this.receivables_overdue = this.convertValues(source["receivables_overdue"], DashboardMetric);
	        this.payables = this.convertValues(source["payables"], DashboardMetric);
	        this.payables_due_today = this.convertValues(source["payables_due_today"], DashboardMetric);
	        this.payables_overdue = this.convertValues(source["payables_overdue"], DashboardMetric);
	        this.payments_month = this.convertValues(source["payments_month"], DashboardMetric);
	        this.receivable_rows = this.convertValues(source["receivable_rows"], DashboardReceivableRow);
	        this.payable_rows = this.convertValues(source["payable_rows"], DashboardPayableRow);
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	    }

		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) return a;
		    if (a.slice && a.map) return (a as any[]).map(elem => this.convertValues(elem, classs));
		    if ("object" === typeof a) return new classs(a);
		    return a;
		}
	}

}

export namespace main {
	
	export class BootstrapState {
	    app_version: string;
	    worker: application.HealthResponse;
	    integrations: application.IntegrationStatus[];
	    error_code?: string;
	    message?: string;
	    // Go type: time
	    checked_at: any;
	
	    static createFrom(source: any = {}) {
	        return new BootstrapState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.app_version = source["app_version"];
	        this.worker = this.convertValues(source["worker"], application.HealthResponse);
	        this.integrations = this.convertValues(source["integrations"], application.IntegrationStatus);
	        this.error_code = source["error_code"];
	        this.message = source["message"];
	        this.checked_at = this.convertValues(source["checked_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

