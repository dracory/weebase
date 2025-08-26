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
