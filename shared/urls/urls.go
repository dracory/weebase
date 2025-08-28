package urls

import (
	neturl "net/url"
	"sort"

	"github.com/dracory/weebase/shared/constants"
	"github.com/samber/lo"
)

const actionParam = "action"

// Login is a convenience wrapper using defaults: basePath "/db" and actionParam "action".
// Signature: Login(params)
func Login(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageLogin, params...)
}

// Connect is a convenience wrapper to construct the connect endpoint URL.
// Signature: Connect(basePath, params)
func Connect(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionApiConnect, params...)
}

// Profiles is a convenience wrapper to construct the profiles endpoint URL.
// Signature: Profiles(basePath, params)
func Profiles(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionApiProfilesList, params...)
}

// Home is a convenience wrapper to construct the Home URL.
// Signature: Home(basePath, params)
func Home(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageHome, params...)
}

// PageTableCreate builds the page action URL for the Create Table page.
// Signature: PageTableCreate(basePath, params)
func PageTableCreate(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageTableCreate, params...)
}

// APITableCreate builds the API action URL for Create Table POSTs.
// Signature: APITableCreate(basePath, params)
func ApiTableCreate(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionApiTableCreate, params...)
}

// URL is a convenience wrapper using defaults: basePath "/db" and actionParam "action".
// Signature: URL(action, parameters)
func URL(basePath, action string, params ...map[string]string) string {
	return Build(basePath, action, params...)
}

// ListTables builds the URL for listing tables
func ListTables(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionApiListTables, params...)
}

// BrowseRows builds the URL for browsing table rows
func BrowseRows(basePath, table string, params ...map[string]string) string {
	p := lo.FirstOr(params, map[string]string{})
	p["table"] = table
	return URL(basePath, constants.ActionApiBrowseRows, p)
}

// TableView builds the URL for table view page
func TableView(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageTableView, params...)
}

// SQLExecute builds the URL for SQL execution
func SQLExecute(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageSQLExecute, params...)
}

// PageSQLExecute builds the URL for the SQL execute page
func PageSQLExecute(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageSQLExecute, params...)
}

// PageImport builds the URL for the import page
func PageImport(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageImport, params...)
}

// PageExport builds the URL for the export page
func PageExport(basePath string, params ...map[string]string) string {
	return URL(basePath, constants.ActionPageExport, params...)
}

// Build constructs a URL like: basePath?actionParam=action&k=v...
// - basePath: mount path, e.g. "/db"
// - actionParam: query key that selects behavior, e.g. "action"
// - action: the action value, e.g. "login"
// - params: optional extra query parameters; nil allowed
// Keys are sorted for stable output. Values are URL-escaped.
func Build(basePath, action string, params ...map[string]string) string {
	p := lo.FirstOr(params, map[string]string{})

	// Ensure basePath starts with '/'
	if basePath == "" || basePath[0] != '/' {
		basePath = "/" + basePath
	}
	q := neturl.Values{}
	q.Set(actionParam, action)
	if len(p) > 0 {
		// stable order
		keys := make([]string, 0, len(p))
		for k := range p {
			if k == "" { // skip empty keys
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			q.Set(k, p[k])
		}
	}
	enc := q.Encode()
	if enc == "" {
		return basePath
	}
	return basePath + "?" + enc
}
