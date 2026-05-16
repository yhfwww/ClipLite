export namespace main {
	
	export class ClipboardRecord {
	    id: number;
	    content: string;
	    // Go type: time
	    created_at: any;
	    is_favorited: boolean;
	    // Go type: time
	    expires_at?: any;
	
	    static createFrom(source: any = {}) {
	        return new ClipboardRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.content = source["content"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.is_favorited = source["is_favorited"];
	        this.expires_at = this.convertValues(source["expires_at"], null);
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
	export class Config {
	    hotkey: string;
	    save_days: number;
	    storage_path: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hotkey = source["hotkey"];
	        this.save_days = source["save_days"];
	        this.storage_path = source["storage_path"];
	    }
	}

}

