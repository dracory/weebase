package constants

// Database driver names
const (
	DriverPostgres  = "postgres"
	DriverMySQL     = "mysql"
	DriverSQLite    = "sqlite"
	DriverSQLServer = "sqlserver"
)

// API actions
const (
	// Connection management
	ActionApiConnect    = "api_connect"
	ActionApiDisconnect = "api_disconnect"

	// Data operations
	ActionApiBrowseRows = "api_browse_rows"
	ActionApiRowView    = "api_row_view"
	ActionApiInsertRow  = "api_insert_row"
	ActionApiUpdateRow  = "api_update_row"
	ActionApiDeleteRow  = "api_delete_row"

	// Profile management
	ActionApiProfilesSave = "api_profiles_save"

	// Schema and table operations
	ActionApiListSchemas = "api_list_schemas"
	ActionApiListTables  = "api_list_tables"

	// SQL operations
	ActionApiSQLExecute = "api_sql_execute"
	ActionApiSQLExplain = "api_sql_explain"
	ActionApiListSaved  = "api_list_saved_queries"
	ActionApiSaveQuery  = "api_save_query"

	// Table operations
	ActionApiTableInfo      = "api_table_info"
	ActionApiViewDefinition = "api_view_definition"
	ActionApiTableCreate    = "api_table_create"
	ActionApiTableDrop      = "api_table_drop"
	ActionApiTableEdit      = "api_table_edit"
)

// Page actions
const (
	// System actions
	ActionHome     = "home"
	ActionAssetCSS = "asset_css"
	ActionAssetJS  = "asset_js"

	// Authentication
	ActionPageLogin  = "page_login"
	ActionPageLogout = "page_logout"

	// Profile management
	ActionPageProfiles = "page_profiles"

	// Data operations
	ActionPageSQLExecute = "page_sql_execute"

	// Table operations
	ActionPageTableCreate = "page_table_create"
	ActionPageTableEdit   = "page_table_edit"
	ActionPageTableView   = "page_table_view"
)
