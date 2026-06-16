export namespace main {
	
	export class DetectRequest {
	    apiKey: string;
	    enablePollution: boolean;
	    enableHijack: boolean;
	    enableWechat: boolean;
	    enableIcp: boolean;
	    enableBlacklist: boolean;
	    enableWall: boolean;
	    concurrency: number;
	    timeoutSeconds: number;
	    importedTargetCount: number;
	    targets: string[];
	
	    static createFrom(source: any = {}) {
	        return new DetectRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.apiKey = source["apiKey"];
	        this.enablePollution = source["enablePollution"];
	        this.enableHijack = source["enableHijack"];
	        this.enableWechat = source["enableWechat"];
	        this.enableIcp = source["enableIcp"];
	        this.enableBlacklist = source["enableBlacklist"];
	        this.enableWall = source["enableWall"];
	        this.concurrency = source["concurrency"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.importedTargetCount = source["importedTargetCount"];
	        this.targets = source["targets"];
	    }
	}
	export class DetectSummary {
	    total: number;
	    checked: number;
	    pollution: number;
	    normal: number;
	    unregistered: number;
	    failed: number;
	
	    static createFrom(source: any = {}) {
	        return new DetectSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.checked = source["checked"];
	        this.pollution = source["pollution"];
	        this.normal = source["normal"];
	        this.unregistered = source["unregistered"];
	        this.failed = source["failed"];
	    }
	}
	export class DetectRow {
	    id: number;
	    type: string;
	    target: string;
	    status: string;
	    checkedAt: string;
	    errorRemark: string;
	    domain: string;
	    beianCode: string;
	    siteName: string;
	
	    static createFrom(source: any = {}) {
	        return new DetectRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.target = source["target"];
	        this.status = source["status"];
	        this.checkedAt = source["checkedAt"];
	        this.errorRemark = source["errorRemark"];
	        this.domain = source["domain"];
	        this.beianCode = source["beianCode"];
	        this.siteName = source["siteName"];
	    }
	}
	export class DetectResponse {
	    rows: DetectRow[];
	    summary: DetectSummary;
	    progress: number;
	    exportPath: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new DetectResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.rows = this.convertValues(source["rows"], DetectRow);
	        this.summary = this.convertValues(source["summary"], DetectSummary);
	        this.progress = source["progress"];
	        this.exportPath = source["exportPath"];
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
	
	
	export class ExportResponse {
	    path: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.message = source["message"];
	    }
	}
	export class ImportResponse {
	    targets: string[];
	    message: string;
	    canceled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ImportResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.targets = source["targets"];
	        this.message = source["message"];
	        this.canceled = source["canceled"];
	    }
	}

}

