package constants

// Action names for the single-endpoint router. Keep in sync with templates.
const (
	ActionHome            = "home"
	ActionAssetCSS        = "asset_css"
	ActionAssetJS         = "asset_js"
	ActionHealthz         = "healthz"
	ActionReadyz          = "readyz"

	ActionLogin           = "login"
	ActionLogout          = "logout"
	ActionLoginJS         = "login_js"
	ActionLoginCSS        = "login_css"

	ActionConnect         = "connect"
	ActionDisconnect      = "disconnect"

	ActionProfiles        = "profiles"
	ActionProfilesSave    = "profiles_save"

	ActionListSchemas     = "list_schemas"
	ActionListTables      = "list_tables"
	ActionTableInfo       = "table_info"
	ActionViewDefinition  = "view_definition"

	// Dedicated landing page when a table is selected from the sidebar or grid
	ActionTableView       = "table_view"

	ActionBrowseRows      = "browse_rows"
	ActionRowView         = "row_view"
	ActionInsertRow       = "insert_row"
	ActionUpdateRow       = "update_row"
	ActionDeleteRow       = "delete_row"

	ActionSQLExecute      = "sql_execute"
	ActionSQLExplain      = "sql_explain"
	ActionListSaved       = "list_saved_queries"
	ActionSaveQuery       = "save_query"

	ActionDDLCreateTable  = "ddl_create_table"
	ActionDDLAlterTable   = "ddl_alter_table"
	ActionDDLDropTable    = "ddl_drop_table"

	ActionExport          = "export"
	ActionImport          = "import"
)

// Noun_Verb alias actions for consistency; keep old names for backward compatibility.
const (
	ActionTablesList   = "tables_list"   // alias of list_tables
	ActionSchemasList  = "schemas_list"  // alias of list_schemas

	ActionTableCreate  = "table_create"  // alias of ddl_create_table
	ActionTableEdit    = "table_edit"    // alias of ddl_alter_table
	ActionTableDrop    = "table_drop"    // alias of ddl_drop_table

	ActionRowsBrowse   = "rows_browse"   // alias of browse_rows
	ActionRowInsert    = "row_insert"    // alias of insert_row
	ActionRowUpdate    = "row_update"    // alias of update_row
	ActionRowDelete    = "row_delete"    // alias of delete_row

	ActionProfilesList = "profiles_list" // alias of profiles
	ActionProfileSave  = "profile_save"  // alias of profiles_save
)
