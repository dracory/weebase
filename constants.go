package weebase

import (
    shconst "github.com/dracory/weebase/shared/constants"
)

// Action names for the single-endpoint router. Keep in sync with templates.
const (
    ActionHome            = shconst.ActionHome
    ActionAssetCSS        = shconst.ActionAssetCSS
    ActionAssetJS         = shconst.ActionAssetJS
    ActionHealthz         = shconst.ActionHealthz
    ActionReadyz          = shconst.ActionReadyz

    ActionLogin           = shconst.ActionLogin
    ActionLogout          = shconst.ActionLogout
    ActionLoginJS         = shconst.ActionLoginJS
    ActionLoginCSS        = shconst.ActionLoginCSS

    ActionConnect         = shconst.ActionConnect
    ActionDisconnect      = shconst.ActionDisconnect

    ActionProfiles        = shconst.ActionProfiles
    ActionProfilesSave    = shconst.ActionProfilesSave

    ActionListSchemas     = shconst.ActionListSchemas
    ActionListTables      = shconst.ActionListTables
    ActionTableInfo       = shconst.ActionTableInfo
    ActionViewDefinition  = shconst.ActionViewDefinition

    ActionBrowseRows      = shconst.ActionBrowseRows
    ActionRowView         = shconst.ActionRowView
    ActionInsertRow       = shconst.ActionInsertRow
    ActionUpdateRow       = shconst.ActionUpdateRow
    ActionDeleteRow       = shconst.ActionDeleteRow

    ActionSQLExecute      = shconst.ActionSQLExecute
    ActionSQLExplain      = shconst.ActionSQLExplain
    ActionListSaved       = shconst.ActionListSaved
    ActionSaveQuery       = shconst.ActionSaveQuery

    ActionDDLCreateTable  = shconst.ActionDDLCreateTable
    ActionDDLAlterTable   = shconst.ActionDDLAlterTable
    ActionDDLDropTable    = shconst.ActionDDLDropTable

    ActionExport          = shconst.ActionExport
    ActionImport          = shconst.ActionImport
)

// Asset paths and content types used by the embedded server.
const (
    AssetPathCSS  = "assets/style.css"
    AssetPathJS   = "assets/app.js"
    ContentTypeCSS = "text/css; charset=utf-8"
    ContentTypeJS  = "application/javascript; charset=utf-8"

    LoginAssetPathJS  = "pages/login/script.js"
    LoginAssetPathCSS = "pages/login/styles.css"
)
