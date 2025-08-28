package constants

// Database driver names
const (
	DriverPostgres  = "postgres"
	DriverMySQL     = "mysql"
	DriverSQLite    = "sqlite"
	DriverSQLServer = "sqlserver"
)

// Cookie names
const (
	CookieProfiles = "weebase_profiles"
)

// API actions
const (
	// Connection management
	ActionApiConnect    = "api_connect"
	ActionApiDisconnect = "api_disconnect"

	// Database operations
	ActionApiDatabasesList = "api_databases_list"

	// Data operations
	ActionApiBrowseRows = "api_browse_rows"
	ActionApiRowView    = "api_row_view"
	ActionApiInsertRow  = "api_insert_row"
	ActionApiUpdateRow  = "api_update_row"
	ActionApiDeleteRow  = "api_delete_row"

	// Profile management
	ActionApiProfilesList = "api_profiles_list"
	ActionApiProfilesSave = "api_profiles_save"

	// Schema operations
	ActionApiSchemasList = "api_schemas_list"

	// Table operations
	ActionApiTablesList = "api_tables_list"

	// SQL operations
	ActionApiSQLExecute = "api_sql_execute"
	ActionApiSQLExplain = "api_sql_explain"

	// Table operations
	ActionApiTableCreate = "api_table_create"
	ActionApiTableList   = "api_table_list"
)

// Page actions
const (
	// Assets
	ActionAssetCSS = "asset_css"
	ActionAssetJS  = "asset_js"

	// Pages
	ActionPageHome        = "page_home"
	ActionPageExport      = "page_export"
	ActionPageImport      = "page_import"
	ActionPageLogin       = "page_login"
	ActionPageLogout      = "page_logout"
	ActionPageProfiles    = "page_profiles"
	ActionPageSQLExecute  = "page_sql_execute"
	ActionPageServer      = "page_server"
	ActionPageDatabase    = "page_database"
	ActionPageTable       = "page_table"
	ActionPageTableCreate = "page_table_create"
)
