export namespace main {
	
	export class BlobData {
	    index: number;
	    size: number;
	    mime: string;
	    isText: boolean;
	    text: string;
	    hex: string;
	    base64: string;
	    truncated: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BlobData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.size = source["size"];
	        this.mime = source["mime"];
	        this.isText = source["isText"];
	        this.text = source["text"];
	        this.hex = source["hex"];
	        this.base64 = source["base64"];
	        this.truncated = source["truncated"];
	    }
	}
	export class BlobInfo {
	    index: number;
	    size: number;
	    name?: string;
	
	    static createFrom(source: any = {}) {
	        return new BlobInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.size = source["size"];
	        this.name = source["name"];
	    }
	}
	export class CellDTO {
	    column: number;
	    name: string;
	    type: string;
	    length: number;
	    value: string;
	    rawHex: string;
	    empty: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CellDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.column = source["column"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.length = source["length"];
	        this.value = source["value"];
	        this.rawHex = source["rawHex"];
	        this.empty = source["empty"];
	    }
	}
	export class ColumnDTO {
	    index: number;
	    name: string;
	    type: string;
	    length: number;
	    startPos: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.name = source["name"];
	        this.type = source["type"];
	        this.length = source["length"];
	        this.startPos = source["startPos"];
	    }
	}
	export class ColumnSpec {
	    name: string;
	    type: string;
	    length: number;
	
	    static createFrom(source: any = {}) {
	        return new ColumnSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.type = source["type"];
	        this.length = source["length"];
	    }
	}
	export class CreateListRequest {
	    path: string;
	    maxEntries: number;
	    overwrite: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateListRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.maxEntries = source["maxEntries"];
	        this.overwrite = source["overwrite"];
	    }
	}
	export class CreateTableRequest {
	    path: string;
	    rows: number;
	    columns: ColumnSpec[];
	    overwrite: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateTableRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.rows = source["rows"];
	        this.columns = this.convertValues(source["columns"], ColumnSpec);
	        this.overwrite = source["overwrite"];
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
	export class DeflateInfo {
	    path: string;
	    fileSize: number;
	    blockSize: number;
	    headerSize: number;
	    maxBlocks: number;
	    blocks: number;
	    originalLength: number;
	    outputSize: number;
	    ratio: number;
	
	    static createFrom(source: any = {}) {
	        return new DeflateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.fileSize = source["fileSize"];
	        this.blockSize = source["blockSize"];
	        this.headerSize = source["headerSize"];
	        this.maxBlocks = source["maxBlocks"];
	        this.blocks = source["blocks"];
	        this.originalLength = source["originalLength"];
	        this.outputSize = source["outputSize"];
	        this.ratio = source["ratio"];
	    }
	}
	export class FileFilter {
	    displayName: string;
	    pattern: string;
	
	    static createFrom(source: any = {}) {
	        return new FileFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.displayName = source["displayName"];
	        this.pattern = source["pattern"];
	    }
	}
	export class ListInfo {
	    path: string;
	    readOnly: boolean;
	    compressed: boolean;
	    entries: number;
	    maxEntries: number;
	    schemaSize: number;
	    indexSize: number;
	    fileSize: number;
	    blobs: BlobInfo[];
	
	    static createFrom(source: any = {}) {
	        return new ListInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.readOnly = source["readOnly"];
	        this.compressed = source["compressed"];
	        this.entries = source["entries"];
	        this.maxEntries = source["maxEntries"];
	        this.schemaSize = source["schemaSize"];
	        this.indexSize = source["indexSize"];
	        this.fileSize = source["fileSize"];
	        this.blobs = this.convertValues(source["blobs"], BlobInfo);
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
	export class RowDTO {
	    index: number;
	    cells: CellDTO[];
	
	    static createFrom(source: any = {}) {
	        return new RowDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.cells = this.convertValues(source["cells"], CellDTO);
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
	export class TableInfo {
	    path: string;
	    readOnly: boolean;
	    compressed: boolean;
	    rows: number;
	    rowSize: number;
	    schemaSize: number;
	    capacity: number;
	    fileSize: number;
	    schema: string;
	    columns: ColumnDTO[];
	
	    static createFrom(source: any = {}) {
	        return new TableInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.readOnly = source["readOnly"];
	        this.compressed = source["compressed"];
	        this.rows = source["rows"];
	        this.rowSize = source["rowSize"];
	        this.schemaSize = source["schemaSize"];
	        this.capacity = source["capacity"];
	        this.fileSize = source["fileSize"];
	        this.schema = source["schema"];
	        this.columns = this.convertValues(source["columns"], ColumnDTO);
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
	export class TablePage {
	    start: number;
	    total: number;
	    limit: number;
	    rows: RowDTO[];
	
	    static createFrom(source: any = {}) {
	        return new TablePage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.start = source["start"];
	        this.total = source["total"];
	        this.limit = source["limit"];
	        this.rows = this.convertValues(source["rows"], RowDTO);
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

