export namespace application {
	
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

