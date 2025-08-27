# MVP Implementation Plan

## UI Framework
- Use Vue.js 3 for simple page-level interactivity (via CDN)
- No build step - use Vue in simple script tag mode
- No single-file components - keep JavaScript with each page
- Use Bootstrap 5 for all UI components
- Include Bootstrap Icons for visual elements
- Ensure responsive design works on all screen sizes
- Use Bootstrap's utility classes for layout and styling
- Use Bootstrap's form components for all input elements
- Implement Bootstrap's grid system for page layouts

## Current State Analysis

The codebase already has most of the required functionality for the MVP,
but needs some refactoring and simplification to focus on the core features.

## Code to Keep

1. **Login**
   - Directory: `pages/page_login`
   - Route: `{basePath}?action=page_login`
   - Allows user to enter database connection parameters (e.g. host, port, user, password, database)
   - No actual authentication logic
   - Simply stores connection parameters in session if connection is successful
   - If SQLite, only database path is needed
   - If SQLite and database path does not exist, create it

2. **Logout**
   - Directory: `pages/page_logout`
   - Route: `{basePath}?action=page_logout`
   - Clears session and redirects to login page

3. **Server**
   - Directory: `pages/page_server`
   - Route: `{basePath}?action=page_server`
   - Displays a list of databases on the server
   - If SQLite, only shows the database path
   - If MySQL or PostgreSQL, shows database name
   - Allows selecting a database, to view its tables

4. **Database**
   - Directory: `pages/page_database`
   - Route: `{basePath}?action=page_database`
   - Shows tables in the selected database
   - Allows selecting a table, to browse its rows

5. **Table**
   - Directory: `pages/page_table`
   - Route: `{basePath}?action=page_table`
   - Shows rows in the selected table
   - Allows pagination and filtering

6. **Database Connection**
   - Directory: `api/api_connect`
   - Route: `{basePath}/?action=api_connect`
   - Handles connecting to MySQL, PostgreSQL, and SQLite
   - Implements proper DSN building for each database type
   - Manages connection state in session

7. **Database Listing**
   - Directory: `api/api_databases_list`
   - Route: `{basePath}/?action=api_databases_list`
   - Lists databases on the server
   - Handles different database dialects

8. **Table Listing**
   - Directory: `api/api_tables_list`
   - Route: `{basePath}/?action=api_tables_list`
   - Lists tables with database-specific queries
   - Handles different database dialects

9. **Table Rows**
   - Directory: `api/api_table_rows`
   - Route: `{basePath}/?action=api_table_rows`
   - Returns rows from a table
   - Handles different database dialects

## Code to Remove/Refactor

1. **Remove Non-MVP API Handlers**
   - `api_row_delete/`
   - `api_row_insert/`
   - `api_row_update/`
   - `api_sql_execute/`
   - `api_sql_explain/`
   - `api_table_create/`
   - `api_table_info/`
   - `api_view_definition/`

2. **Simplify UI Pages**
   - Keep only the essential pages:
     - Home/Connection page
     - Database browser
   - Remove login/logout functionality for MVP
   - Remove table creation/editing UI

3. **Simplify Configuration**
   - Remove complex configuration options
   - Keep only essential database connection parameters

## New Code Needed

1. **Database Listing**
   - Add handler to list databases on the server
   - Implement database-specific queries for each supported database
   - Add API endpoint: `GET /api/databases`

2. **Simplified UI**
   - Create a clean, minimal UI with:
     - Connection form
     - Database selector
     - Table list
     - Row browser
   - Remove unnecessary UI components

3. **Error Handling**
   - Add consistent error handling
   - Improve error messages for database operations

## Implementation Steps

1. **Phase 1: Core Functionality**
   - [ ] Implement database listing functionality
   - [ ] Clean up connection handling
   - [ ] Remove non-essential API endpoints

2. **Phase 2: UI Implementation**
   - [ ] Set up Vue.js 3 (via CDN) in each page's template
   - [ ] Include Bootstrap 5 and Bootstrap Icons in base template
   - [ ] Implement responsive UI templates using Bootstrap 5
   - [ ] Style all forms with Bootstrap form controls
   - [ ] Implement Bootstrap tables for data display
   - [ ] Add Bootstrap navigation components
   - [ ] Ensure mobile responsiveness with Bootstrap's grid system
   - [ ] Add page-specific Vue instances for interactivity

3. **Phase 3: Testing**
   - [ ] Test with all supported databases
   - [ ] Verify error handling
   - [ ] Test with different database versions

## Technical Decisions

1. **Session Management**
   - Keep simple in-memory session for MVP
   - Store only active connection info

2. **API Design**
   - All API endpoints use POST method for security
   - Always return HTTP 200 status code
   - Standardized response format:
     ```json
     {
       "status": "success/error",
       "message": "Message to display",
       "data": {}
     }
     ```
   - Keep endpoints minimal and action-based
   - Use action parameter in query string to identify intent: `?action=api_some_action`
   - Never use GET method for API calls, always use POST
   - Always use github.com/dracory/api library to generate consistent responses

3. **Frontend**
   - Use Vue.js 3 for simple page-level reactivity (via CDN)
   - No build step or component system
   - Each page manages its own Vue instance
   - Use Bootstrap 5 for all UI components
   - Include Bootstrap Icons for visual elements
   - Implement responsive design with Bootstrap's grid system

## Success Criteria

- [ ] Can connect to MySQL, PostgreSQL, and SQLite
- [ ] Can list databases on the server
- [ ] Can list tables in a database
- [ ] Can browse rows in a table
- [ ] No authentication required
- [ ] Clean, minimal UI
