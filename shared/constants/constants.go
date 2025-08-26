package constants

// Database driver names
const (
	DriverPostgres = "postgres"
	DriverMySQL    = "mysql"
	DriverSQLite   = "sqlite"
)

// Action names for the single-endpoint router. Keep in sync with templates.
const (
	// System actions
	ActionHome     = "home"
	ActionAssetCSS = "asset_css"
	ActionAssetJS  = "asset_js"
	ActionHealthz  = "healthz"
	ActionReadyz   = "readyz"

	// Authentication
	ActionPageLogin  = "page_login"
	ActionPageLogout = "page_logout"

	// Connection management
	ActionApiConnect    = "api_connect"
	ActionApiDisconnect = "api_disconnect"

	// Profiles
	ActionPageProfiles    = "page_profiles"
	ActionApiProfilesSave = "api_profiles_save"

	// Schema and table operations
	ActionApiListSchemas    = "api_list_schemas"
	ActionApiListTables     = "api_list_tables"
	ActionApiTableInfo      = "api_table_info"
	ActionApiViewDefinition = "api_view_definition"

	// Data operations
	ActionApiBrowseRows = "api_browse_rows"
	ActionApiRowView    = "api_row_view"
	ActionApiInsertRow  = "api_insert_row"
	ActionApiUpdateRow  = "api_update_row"
	ActionApiDeleteRow  = "api_delete_row"

	// SQL operations
	ActionPageSQLExecute = "page_sql_execute"
	ActionApiSQLExecute  = "api_sql_execute"
	ActionApiSQLExplain  = "api_sql_explain"
	ActionApiListSaved   = "api_list_saved_queries"
	ActionApiSaveQuery   = "api_save_query"
)

// Table management constants
const (
	ActionPageTableCreate = "page_table_create"
	ActionPageTableEdit   = "page_table_edit"
	ActionPageTableView   = "page_table_view"

	ActionApiTableCreate = "api_table_create"
	ActionApiTableDrop   = "api_table_drop"
	ActionApiTableEdit   = "api_table_edit"
)
