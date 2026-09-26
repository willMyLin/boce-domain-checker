export namespace keys {
	
	export class Accelerator {
	    Key: string;
	    Modifiers: string[];
	
	    static createFrom(source: any = {}) {
	        return new Accelerator(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Key = source["Key"];
	        this.Modifiers = source["Modifiers"];
	    }
	}

}

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
	export class ValidateTargetsResponse {
	    valid: string[];
	    invalid: string[];
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ValidateTargetsResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.valid = source["valid"];
	        this.invalid = source["invalid"];
	        this.message = source["message"];
	    }
	}

}

export namespace menu {
	
	export class MenuItem {
	    Label: string;
	    Role: number;
	    Accelerator?: keys.Accelerator;
	    Type: string;
	    Disabled: boolean;
	    Hidden: boolean;
	    Checked: boolean;
	    SubMenu?: Menu;
	
	    static createFrom(source: any = {}) {
	        return new MenuItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Label = source["Label"];
	        this.Role = source["Role"];
	        this.Accelerator = this.convertValues(source["Accelerator"], keys.Accelerator);
	        this.Type = source["Type"];
	        this.Disabled = source["Disabled"];
	        this.Hidden = source["Hidden"];
	        this.Checked = source["Checked"];
	        this.SubMenu = this.convertValues(source["SubMenu"], Menu);
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
	export class Menu {
	    Items: MenuItem[];
	
	    static createFrom(source: any = {}) {
	        return new Menu(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Items = this.convertValues(source["Items"], MenuItem);
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

