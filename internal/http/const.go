// Package main provides a CLI application to process Apache log files.
package http

var HttpMethods = map[string]string{
	"GET":        "Retrieve a resource",
	"POST":       "Create a new resource",
	"PUT":        "Update an existing resource",
	"PATCH":      "Partially update an existing resource",
	"DELETE":     "Delete a resource",
	"HEAD":       "Retrieve metadata about a resource",
	"OPTIONS":    "Describe the HTTP methods supported by a resource",
	"CONNECT":    "Establish a tunnel to the server",
	"TRACE":      "Perform a message loop-back test",
	"LINK":       "Create a relationship between two existing resources",
	"UNLINK":     "Remove a relationship between two existing resources",
	"SEARCH":     "Query resources",
	"MKCALENDAR": "Create a new calendar",
	"MKCOL":      "Create a new collection",
	"COPY":       "Copy a resource",
	"MOVE":       "Move a resource",
	"LOCK":       "Lock a resource",
	"UNLOCK":     "Unlock a resource",
	"PROPFIND":   "Retrieve properties of a resource",
	"PROPPATCH":  "Update properties of a resource",
}

var HttpStatusCodes = map[uint16]string{
	// 1xx - Informational
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",
	103: "Early Hints",

	// 2xx - Success
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	208: "Already Reported",
	226: "IM Used",

	// 3xx - Redirection
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	305: "Use Proxy",
	306: "Switch Proxy",
	307: "Temporary Redirect",
	308: "Permanent Redirect",

	// 4xx - Client Error
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Entity",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",

	// Non-standard
	444: "No Response",
	495: "SSL Certificate Error",
	496: "SSL Certificate Required",
	497: "HTTP Request Sent to HTTPS Port",
	499: "Client Closed Request",

	// 5xx - Server Error
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",

	// Non-standard
	520: "Unknown Error",
	521: "Web Server Is Down",
	522: "Connection Timed Out",
	523: "Origin Is Unreachable",
	524: "A Timeout Occurred",
	598: "Network Read Timeout Error",
}

var ContentTypes = map[string]string{
	// Text Content Types
	"text/plain":      "Plain text",
	"text/html":       "HTML documents",
	"text/css":        "CSS stylesheets",
	"text/javascript": "JavaScript files",
	"text/xml":        "XML documents",
	"text/markdown":   "Markdown documents",

	// Data Interchange Content Types
	"application/javascript": "JavaScript files",
	"application/json":       "JSON data",
	"application/xml":        "XML data",
	"application/x-yaml":     "YAML data",
	"application/x-ndjson":   "Newline-delimited JSON",

	// Form and Multipart Content Types
	"application/x-www-form-urlencoded": "Form data (often used for HTML form submissions)",
	"multipart/form-data":               "Form data with file uploads",
	"multipart/byteranges":              "Multipart responses with byte ranges",

	// Binary Content Types
	"application/octet-stream":     "Binary data (often used for file downloads)",
	"application/pdf":              "PDF documents",
	"application/zip":              "ZIP archives",
	"application/gzip":             "GZIP archives",
	"application/x-tar":            "TAR archives",
	"application/x-rar-compressed": "RAR archives",
	"application/x-7z-compressed":  "7-Zip archives",

	// Other Content Types
	"application/atom+xml":          "Atom feeds",
	"application/rss+xml":           "RSS feeds",
	"application/x-shockwave-flash": "Flash files",
	"application/msword":            "Microsoft Word documents",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": "Microsoft Word (.docx) documents",
	"application/vnd.ms-excel": "Microsoft Excel documents",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         "Microsoft Excel (.xlsx) documents",
	"application/vnd.ms-powerpoint":                                             "Microsoft PowerPoint documents",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": "Microsoft PowerPoint (.pptx) documents",

	// Image Content Types
	"image/jpeg":               "JPEG images",
	"image/png":                "PNG images",
	"image/gif":                "GIF images",
	"image/svg+xml":            "SVG images",
	"image/bmp":                "BMP images",
	"image/tiff":               "TIFF images",
	"image/webp":               "WebP images",
	"image/x-icon":             "ICO images",
	"image/vnd.microsoft.icon": "ICO images",
	"image/ico":                "ICO images",
	"image/icon":               "ICO images",

	// Audio Content Types
	"audio/mpeg": "MP3 audio files",
	"audio/wav":  "WAV audio files",
	"audio/aac":  "AAC audio files",
	"audio/ogg":  "OGG audio files",
	"audio/flac": "FLAC audio files",

	// Video Content Types
	"video/mp4":       "MP4 video files",
	"video/webm":      "WebM video files",
	"video/ogg":       "OGG video files",
	"video/quicktime": "QuickTime video files",
	"video/x-msvideo": "AVI video files",

	// Font Content Types
	"font/ttf":   "TrueType fonts",
	"font/woff":  "Web Open Font Format (WOFF) fonts",
	"font/woff2": "Web Open Font Format 2 (WOFF2) fonts",
	"font/otf":   "OpenType fonts",
}
