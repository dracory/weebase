package urls

import (
	neturl "net/url"
	"sort"
)

const actionParam = "action"

// Login is a convenience wrapper using defaults: basePath "/db" and actionParam "action".
// Signature: Login(params)
func Login(basePath string, params map[string]string) string {
	return URL(basePath, "login", params)
}

// URL is a convenience wrapper using defaults: basePath "/db" and actionParam "action".
// Signature: URL(action, parameters)
func URL(basePath, action string, parameters map[string]string) string {
	return Build(basePath, action, parameters)
}

// Build constructs a URL like: basePath?actionParam=action&k=v...
// - basePath: mount path, e.g. "/db"
// - actionParam: query key that selects behavior, e.g. "action"
// - action: the action value, e.g. "login"
// - params: optional extra query parameters; nil allowed
// Keys are sorted for stable output. Values are URL-escaped.
func Build(basePath, action string, params map[string]string) string {
	// Ensure basePath starts with '/'
	if basePath == "" || basePath[0] != '/' {
		basePath = "/" + basePath
	}
	q := neturl.Values{}
	q.Set(actionParam, action)
	if len(params) > 0 {
		// stable order
		keys := make([]string, 0, len(params))
		for k := range params {
			if k == "" { // skip empty keys
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			q.Set(k, params[k])
		}
	}
	enc := q.Encode()
	if enc == "" {
		return basePath
	}
	return basePath + "?" + enc
}
