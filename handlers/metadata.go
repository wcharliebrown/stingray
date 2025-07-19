package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"stingray/database"
	"stingray/models"
	"stingray/templates"
)

// MetadataHandler handles metadata-related requests
type MetadataHandler struct {
	db *database.Database
	sm *SessionMiddleware
}

// NewMetadataHandler creates a new metadata handler
func NewMetadataHandler(db *database.Database) *MetadataHandler {
	return &MetadataHandler{
		db: db,
		sm: NewSessionMiddleware(db),
	}
}

// HandleTableList handles the table listing page
func (h *MetadataHandler) HandleTableList(w http.ResponseWriter, r *http.Request) {
	// Get all table metadata
	tableMetadata, err := h.db.GetAllTableMetadata()
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error fetching table metadata", http.StatusInternalServerError)
		return
	}

	// Check if user is authenticated
	isAuthenticated := h.sm.IsAuthenticated(r)
	var userID int
	var isEngineer, isAdmin bool
	if isAuthenticated {
		session, err := h.sm.GetSessionFromRequest(r)
		if err == nil {
			userID = session.UserID
			// Check if user is in engineer or admin group
			isEngineer, _ = h.db.IsUserInGroup(userID, "engineer")
			isAdmin, _ = h.db.IsUserInGroup(userID, "admin")
		}
	}

	// Check if engineer mode is requested
	engineerMode := r.URL.Query().Get("engineer") == "true" && isEngineer

	// Filter tables based on user access
	var accessibleTables []models.TableMetadata
	if engineerMode {
		// In engineer mode, show all tables
		accessibleTables = tableMetadata
	} else {
		// Normal mode - filter based on permissions
		for _, table := range tableMetadata {
			hasAccess := false
			
			// Parse read groups
			var readGroups []string
			if table.ReadGroups != "" {
				if err := json.Unmarshal([]byte(table.ReadGroups), &readGroups); err != nil {
					continue
				}
			}

			// Check access
			if len(readGroups) == 0 {
				hasAccess = true // No restrictions
			} else {
				for _, group := range readGroups {
					// Everyone is automatically in the 'everyone' group
					if group == "everyone" {
						hasAccess = true
						break
					}
					// For authenticated users, check their groups
					if isAuthenticated {
						if inGroup, _ := h.db.IsUserInGroup(userID, group); inGroup {
							hasAccess = true
							break
						}
					}
				}
			}

			if hasAccess {
				accessibleTables = append(accessibleTables, table)
			}
		}
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		response := map[string]interface{}{
			"tables":       accessibleTables,
			"is_engineer":  isEngineer,
			"engineer_mode": engineerMode,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// HTML response - use template system
	mainContentTemplate := `<h1>Database Tables</h1>
			
			{{if .IsEngineer}}
			<div class="toggle-container">
				<label class="toggle-label">View Mode:</label>
				<div class="toggle-buttons">
					<a href="?" class="toggle-btn {{if not .EngineerMode}}active{{end}}">Admin View</a>
					<a href="?engineer=true" class="toggle-btn {{if .EngineerMode}}active{{end}}">Engineer View</a>
				</div>
			</div>
			{{else}}
			<div class="toggle-container">
				<label class="toggle-label">View Mode:</label>
				<div class="toggle-buttons">
					<button class="toggle-btn active" disabled>Admin View</button>
					<button class="toggle-btn" disabled>Engineer View</button>
				</div>
				<small style="color: #6c757d; margin-top: 0.5rem; display: block;">Engineer view is only available to users in the Engineer group.</small>
			</div>
			{{end}}
			
			{{if .EngineerMode}}
			<div class="engineer-mode-notice">
				<strong>Engineer Mode:</strong> Showing all tables in the database, including system tables.
			</div>
			{{end}}
			
			{{if or .IsEngineer .IsAdmin .EngineerMode}}
			<div class="table-actions" style="margin-bottom: 2rem;">
				<a href="/metadata/create-table" class="btn btn-success">Create Table</a>
			</div>
			{{end}}
			
			<ul class="table-list">
				{{range .Tables}}
				<li class="table-item">
					<div class="table-info">
						<div class="table-name">{{.DisplayName}}</div>
						<div class="table-description">{{.Description}}</div>
					</div>
					<div class="table-actions">
						<a href="/metadata/table/{{.TableName}}" class="btn btn-primary">View Data</a>
						{{if or $.IsEngineer $.IsAdmin $.EngineerMode}}
						<a href="/metadata/edit-table/{{.TableName}}" class="btn btn-secondary">Edit Metadata</a>
						<a href="/metadata/delete-table/{{.TableName}}" class="btn btn-danger" onclick="return confirm('Are you sure you want to delete this table? This will permanently remove the table, its metadata, and all field metadata. This action cannot be undone.')">Delete</a>
						{{end}}
					</div>
				</li>
				{{end}}
			</ul>`

	// Process the main content template
	contentTmpl, err := template.New("content").Parse(mainContentTemplate)
	if err != nil {
		http.Error(w, "Error parsing content template", http.StatusInternalServerError)
		return
	}

	var contentBuffer strings.Builder
	contentData := map[string]interface{}{
		"Tables":       accessibleTables,
		"IsEngineer":   isEngineer,
		"IsAdmin":      isAdmin,
		"EngineerMode": engineerMode,
	}
	err = contentTmpl.Execute(&contentBuffer, contentData)
	if err != nil {
		http.Error(w, "Error executing content template", http.StatusInternalServerError)
		return
	}

	mainContent := contentBuffer.String()

	// Create template data
	data := map[string]interface{}{
		"Title":          "Database Tables - Sting Ray",
		"MetaDescription": "View and manage database tables",
		"Header":         "Database Tables",
		"Navigation":     template.HTML(`<a href="/">Home</a> | <a href="/page/about">About</a> | <a href="/metadata/tables">Tables</a>`),
		"MainContent":    template.HTML(mainContent),
		"Sidebar":        "",
		"Footer":         "© 2025 StingRay",
		"CSSClass":       "metadata",
		"Scripts":        "",
		"Tables":         accessibleTables,
		"IsEngineer":     isEngineer,
		"IsAdmin":        isAdmin,
		"EngineerMode":   engineerMode,
	}

	// Render using template system
	html, err := templates.RenderMetadataPage(data)
	if err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// HandleTableData handles viewing table data
func (h *MetadataHandler) HandleTableData(w http.ResponseWriter, r *http.Request) {
	// Extract table name from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/metadata/table/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}
	tableName := pathParts[0]

	// Get table metadata
	tableMetadata, err := h.db.GetTableMetadata(tableName)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Check if user is authenticated
	isAuthenticated := h.sm.IsAuthenticated(r)
	var userID int
	if isAuthenticated {
		session, err := h.sm.GetSessionFromRequest(r)
		if err != nil {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			return
		}
		userID = session.UserID
	}

	// Parse read groups
	var readGroups []string
	if tableMetadata.ReadGroups != "" {
		if err := json.Unmarshal([]byte(tableMetadata.ReadGroups), &readGroups); err != nil {
			http.Error(w, "Error parsing read groups", http.StatusInternalServerError)
			return
		}
	}

	// Check if user has read access
	hasAccess := false
	if len(readGroups) == 0 {
		hasAccess = true // No restrictions
	} else {
		for _, group := range readGroups {
			// Everyone is automatically in the 'everyone' group
			if group == "everyone" {
				hasAccess = true
				break
			}
			// For authenticated users, check their groups
			if isAuthenticated {
				if inGroup, _ := h.db.IsUserInGroup(userID, group); inGroup {
					hasAccess = true
					break
				}
			}
		}
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	// Get table data
	rows, total, err := h.db.GetTableRows(tableName, page, pageSize)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error fetching table data", http.StatusInternalServerError)
		return
	}

	// Get field metadata
	fieldMetadata, err := h.db.GetFieldMetadata(tableName)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error fetching field metadata", http.StatusInternalServerError)
		return
	}

	// Check write permissions
	var writeGroups []string
	if tableMetadata.WriteGroups != "" {
		if err := json.Unmarshal([]byte(tableMetadata.WriteGroups), &writeGroups); err != nil {
			http.Error(w, "Error parsing write groups", http.StatusInternalServerError)
			return
		}
	}

	canEdit := false
	canDelete := false
	canCreate := false
	if len(writeGroups) == 0 {
		canEdit = true
		canDelete = true
		canCreate = true
	} else {
		for _, group := range writeGroups {
			// Everyone is automatically in the 'everyone' group
			if group == "everyone" {
				canEdit = true
				canDelete = true
				canCreate = true
				break
			}
			// For authenticated users, check their groups
			if isAuthenticated {
				if inGroup, _ := h.db.IsUserInGroup(userID, group); inGroup {
					canEdit = true
					canDelete = true
					canCreate = true
					break
				}
			}
		}
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		response := map[string]interface{}{
			"table_name":    tableName,
			"display_name":  tableMetadata.DisplayName,
			"fields":        fieldMetadata,
			"rows":          rows,
			"total_rows":    total,
			"current_page":  page,
			"page_size":     pageSize,
			"can_edit":      canEdit,
			"can_delete":    canDelete,
			"can_create":    canCreate,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>{{.DisplayName}} - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 100%; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.table-actions { margin-bottom: 1rem; }
			.btn { padding: 0.5rem 1rem; border: none; border-radius: 4px; text-decoration: none; font-size: 0.9rem; cursor: pointer; margin-right: 0.5rem; }
			.btn-primary { background: #667eea; color: white; }
			.btn-secondary { background: #6c757d; color: white; }
			.btn-danger { background: #dc3545; color: white; }
			.btn:hover { opacity: 0.8; }
			table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
			th, td { padding: 0.75rem; text-align: left; border-bottom: 1px solid #e9ecef; }
			th { background: #f8f9fa; font-weight: 600; }
			tr:hover { background: #f8f9fa; }
			.pagination { margin-top: 1rem; text-align: center; }
			.pagination a { padding: 0.5rem 1rem; margin: 0 0.25rem; text-decoration: none; border: 1px solid #dee2e6; border-radius: 4px; }
			.pagination a:hover { background: #e9ecef; }
			.pagination .current { background: #667eea; color: white; border-color: #667eea; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>{{.DisplayName}}</h1>
			<div class="table-actions">
				{{if .CanCreate}}
				<a href="/metadata/edit/{{.TableName}}/new" class="btn btn-primary">Create New</a>
				{{end}}
				<a href="/metadata/tables" class="btn btn-secondary">Back to Tables</a>
			</div>
			<table>
				<thead>
					<tr>
						{{range .Fields}}
						{{if ge .ListPosition 0}}
						<th>{{.DisplayName}}</th>
						{{end}}
						{{end}}
						{{if or .CanEdit .CanDelete}}
						<th>Actions</th>
						{{end}}
					</tr>
				</thead>
				<tbody>
					{{range .Rows}}
					    {{ $row := . }}
					<tr>
						{{range $.Fields}}
						{{if ge .ListPosition 0}}
						<td>{{index $row.Data .FieldName}}</td>
						{{end}}
						{{end}}
						{{if or $.CanEdit $.CanDelete}}
						<td>
							{{if $.CanEdit}}
							<a href="/metadata/edit/{{$.TableName}}/{{$row.ID}}" class="btn btn-secondary">Edit</a>
							{{end}}
							{{if $.CanDelete}}
							<a href="/metadata/delete/{{$.TableName}}/{{$row.ID}}" class="btn btn-danger" onclick="return confirm('Are you sure?')">Delete</a>
							{{end}}
						</td>
						{{end}}
					</tr>
					{{end}}
				</tbody>
			</table>
			<div class="pagination">
				{{if gt .CurrentPage 1}}
				<a href="?page={{subtract .CurrentPage 1}}&page_size={{.PageSize}}">Previous</a>
				{{end}}
				<span class="current">Page {{.CurrentPage}} of {{divide .TotalRows .PageSize}}</span>
				{{if lt .CurrentPage (divide .TotalRows .PageSize)}}
				<a href="?page={{add .CurrentPage 1}}&page_size={{.PageSize}}">Next</a>
				{{end}}
			</div>
		</div>
	</body>
	</html>`

	// Create template functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
		"divide": func(a, b int) int { 
			if b == 0 { return 1 }
			result := a / b
			if a%b > 0 { result++ }
			return result
		},
	}

	t, err := template.New("table_data").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	data := models.TableData{
		TableName:    tableName,
		DisplayName:  tableMetadata.DisplayName,
		Fields:       fieldMetadata,
		Rows:         rows,
		TotalRows:    total,
		CurrentPage:  page,
		PageSize:     pageSize,
		CanEdit:      canEdit,
		CanDelete:    canDelete,
		CanCreate:    canCreate,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// HandleEditRow handles editing or creating table rows
func (h *MetadataHandler) HandleEditRow(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Extract table name and row ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/metadata/edit/"), "/")
	if len(pathParts) < 1 {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}
	tableName := pathParts[0]
	rowID := "new"
	if len(pathParts) > 1 {
		rowID = pathParts[1]
	}

	// Get table metadata
	tableMetadata, err := h.db.GetTableMetadata(tableName)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Check user permissions
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

	// Parse write groups
	var writeGroups []string
	if tableMetadata.WriteGroups != "" {
		if err := json.Unmarshal([]byte(tableMetadata.WriteGroups), &writeGroups); err != nil {
			http.Error(w, "Error parsing write groups", http.StatusInternalServerError)
			return
		}
	}

	// Check if user has write access
	hasAccess := false
	if len(writeGroups) == 0 {
		hasAccess = true // No restrictions
	} else {
		for _, group := range writeGroups {
			if inGroup, _ := h.db.IsUserInGroup(userID, group); inGroup {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Get field metadata
	fieldMetadata, err := h.db.GetFieldMetadata(tableName)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error fetching field metadata", http.StatusInternalServerError)
		return
	}

	// Check if engineer mode is requested
	engineerMode := r.URL.Query().Get("engineer") == "true"

	// Handle form submission
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Build data map from form
		data := make(map[string]interface{})
		for key, values := range r.Form {
			if len(values) > 0 {
				data[key] = values[0]
			}
		}

		// Remove non-field keys
		delete(data, "engineer")

		if rowID == "new" {
			// Create new row
			delete(data, "id")
			if err := h.db.CreateTableRow(tableName, data); err != nil {
				database.LogSQLError(err)
				http.Error(w, "Error creating row", http.StatusInternalServerError)
				return
			}
		} else {
			// Update existing row
			id, err := strconv.Atoi(rowID)
			if err != nil {
				http.Error(w, "Invalid row ID", http.StatusBadRequest)
				return
			}
			if err := h.db.UpdateTableRow(tableName, id, data); err != nil {
				database.LogSQLError(err)
				http.Error(w, "Error updating row", http.StatusInternalServerError)
				return
			}
		}

		// Redirect back to table view
		http.Redirect(w, r, fmt.Sprintf("/metadata/table/%s", tableName), http.StatusSeeOther)
		return
	}

	// Get existing row data if editing
	var row models.TableRow
	isNew := rowID == "new"
	if !isNew {
		id, err := strconv.Atoi(rowID)
		if err != nil {
			http.Error(w, "Invalid row ID", http.StatusBadRequest)
			return
		}
		existingRow, err := h.db.GetTableRow(tableName, id)
		if err != nil {
			database.LogSQLError(err)
			http.Error(w, "Row not found", http.StatusNotFound)
			return
		}
		row = *existingRow
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		response := map[string]interface{}{
			"table_name":    tableName,
			"display_name":  tableMetadata.DisplayName,
			"fields":        fieldMetadata,
			"row":           row,
			"is_new":        isNew,
			"engineer_mode": engineerMode,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>{{if .IsNew}}Create{{else}}Edit{{end}} {{.DisplayName}} - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 800px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.form-group { margin-bottom: 1rem; }
			label { display: block; margin-bottom: 0.5rem; font-weight: 600; }
			input, textarea, select { width: 100%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; }
			textarea { min-height: 100px; }
			.btn { padding: 0.75rem 1.5rem; border: none; border-radius: 4px; text-decoration: none; font-size: 1rem; cursor: pointer; margin-right: 0.5rem; }
			.btn-primary { background: #667eea; color: white; }
			.btn-secondary { background: #6c757d; color: white; }
			.btn:hover { opacity: 0.8; }
			.engineer-mode { background: #fff3cd; border: 1px solid #ffeaa7; padding: 1rem; border-radius: 4px; margin-bottom: 1rem; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>{{if .IsNew}}Create{{else}}Edit{{end}} {{.DisplayName}}</h1>
			{{if .EngineerMode}}
			<div class="engineer-mode">
				<strong>Engineer Mode:</strong> Showing raw field names and values
			</div>
			{{end}}
			<form method="POST">
				{{if .EngineerMode}}
				<input type="hidden" name="engineer" value="true">
				{{end}}
				{{range .Fields}}
				{{if ge .FormPosition 0}}
				<div class="form-group">
					<label for="{{.FieldName}}">{{if $.EngineerMode}}{{.FieldName}}{{else}}{{.DisplayName}}{{end}}</label>
					{{if eq .HTMLInputType "textarea"}}
					<textarea name="{{.FieldName}}" id="{{.FieldName}}" {{if .IsRequired}}required{{end}} {{if .IsReadOnly}}readonly{{end}}>{{if not $.IsNew}}{{index $.Row.Data .FieldName}}{{end}}</textarea>
					{{else if eq .HTMLInputType "select"}}
					<select name="{{.FieldName}}" id="{{.FieldName}}" {{if .IsRequired}}required{{end}} {{if .IsReadOnly}}disabled{{end}}>
						<option value="">Select...</option>
						<!-- Add options here based on validation rules -->
					</select>
					{{else}}
					<input type="{{.HTMLInputType}}" name="{{.FieldName}}" id="{{.FieldName}}" value="{{if not $.IsNew}}{{index $.Row.Data .FieldName}}{{end}}" {{if .IsRequired}}required{{end}} {{if .IsReadOnly}}readonly{{end}}>
					{{end}}
					{{if not $.IsNew}}
					<small>DB Type: {{.DBType}} | Form Position: {{.FormPosition}}</small>
					{{end}}
				</div>
				{{end}}
				{{end}}
				<div class="form-group">
					<button type="submit" class="btn btn-primary">{{if .IsNew}}Create{{else}}Update{{end}}</button>
					<a href="/metadata/table/{{.TableName}}" class="btn btn-secondary">Cancel</a>
					{{if not .EngineerMode}}
					<a href="?engineer=true" class="btn btn-secondary">Engineer Mode</a>
					{{else}}
					<a href="?" class="btn btn-secondary">Normal Mode</a>
					{{end}}
				</div>
			</form>
		</div>
	</body>
	</html>`

	t, err := template.New("edit_row").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	data := models.FormData{
		TableName:    tableName,
		DisplayName:  tableMetadata.DisplayName,
		Fields:       fieldMetadata,
		Row:          row,
		IsNew:        isNew,
		EngineerMode: engineerMode,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// HandleDeleteRow handles deleting table rows
func (h *MetadataHandler) HandleDeleteRow(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Extract table name and row ID from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/metadata/delete/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Table name and row ID required", http.StatusBadRequest)
		return
	}
	tableName := pathParts[0]
	rowID := pathParts[1]

	// Get table metadata
	tableMetadata, err := h.db.GetTableMetadata(tableName)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Check user permissions
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

	// Parse write groups
	var writeGroups []string
	if tableMetadata.WriteGroups != "" {
		if err := json.Unmarshal([]byte(tableMetadata.WriteGroups), &writeGroups); err != nil {
			http.Error(w, "Error parsing write groups", http.StatusInternalServerError)
			return
		}
	}

	// Check if user has write access
	hasAccess := false
	if len(writeGroups) == 0 {
		hasAccess = true // No restrictions
	} else {
		for _, group := range writeGroups {
			if inGroup, _ := h.db.IsUserInGroup(userID, group); inGroup {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete the row
	id, err := strconv.Atoi(rowID)
	if err != nil {
		http.Error(w, "Invalid row ID", http.StatusBadRequest)
		return
	}

	if err := h.db.DeleteTableRow(tableName, id); err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error deleting row", http.StatusInternalServerError)
		return
	}

	// Redirect back to table view
	http.Redirect(w, r, fmt.Sprintf("/metadata/table/%s", tableName), http.StatusSeeOther)
} 

// HandleEditTableMetadata handles editing table metadata
func (h *MetadataHandler) HandleEditTableMetadata(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Extract table name from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/metadata/edit-table/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}
	tableName := pathParts[0]

	// Get table metadata
	tableMetadata, err := h.db.GetTableMetadata(tableName)
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Check user permissions - only admin and engineer can edit table metadata
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

	isAdmin, _ := h.db.IsUserInGroup(userID, "admin")
	isEngineer, _ := h.db.IsUserInGroup(userID, "engineer")

	if !isAdmin && !isEngineer {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Handle form submission
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Update table metadata
		tableMetadata.DisplayName = r.FormValue("display_name")
		tableMetadata.Description = r.FormValue("description")
		tableMetadata.ReadGroups = r.FormValue("read_groups")
		tableMetadata.WriteGroups = r.FormValue("write_groups")

		if err := h.db.UpdateTableMetadata(tableMetadata); err != nil {
			database.LogSQLError(err)
			http.Error(w, "Error updating table metadata", http.StatusInternalServerError)
			return
		}

		// Redirect back to tables list
		http.Redirect(w, r, "/metadata/tables", http.StatusSeeOther)
		return
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		response := map[string]interface{}{
			"table_name":   tableName,
			"metadata":     tableMetadata,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Edit {{.DisplayName}} - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 800px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.form-group { margin-bottom: 1rem; }
			label { display: block; margin-bottom: 0.5rem; font-weight: 600; }
			input, textarea { width: 100%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; }
			textarea { min-height: 100px; }
			.btn { padding: 0.75rem 1.5rem; border: none; border-radius: 4px; text-decoration: none; font-size: 1rem; cursor: pointer; margin-right: 0.5rem; }
			.btn-primary { background: #667eea; color: white; }
			.btn-secondary { background: #6c757d; color: white; }
			.btn-danger { background: #dc3545; color: white; }
			.btn:hover { opacity: 0.8; }
			.help-text { font-size: 0.9rem; color: #6c757d; margin-top: 0.25rem; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Edit {{.DisplayName}}</h1>
			<form method="POST">
				<div class="form-group">
					<label for="table_name">Table Name</label>
					<input type="text" id="table_name" value="{{.TableName}}" readonly>
					<div class="help-text">Table name cannot be changed</div>
				</div>
				<div class="form-group">
					<label for="display_name">Display Name</label>
					<input type="text" name="display_name" id="display_name" value="{{.DisplayName}}" required>
				</div>
				<div class="form-group">
					<label for="description">Description</label>
					<textarea name="description" id="description">{{.Description}}</textarea>
				</div>
				<div class="form-group">
					<label for="read_groups">Read Groups (JSON array)</label>
					<textarea name="read_groups" id="read_groups">{{.ReadGroups}}</textarea>
					<div class="help-text">Example: ["admin", "engineer", "everyone"]</div>
				</div>
				<div class="form-group">
					<label for="write_groups">Write Groups (JSON array)</label>
					<textarea name="write_groups" id="write_groups">{{.WriteGroups}}</textarea>
					<div class="help-text">Example: ["admin", "engineer"]</div>
				</div>
				<div class="form-group">
					<button type="submit" class="btn btn-primary">Update Metadata</button>
					<a href="/metadata/tables" class="btn btn-secondary">Cancel</a>
					<a href="/metadata/delete-table/{{.TableName}}" class="btn btn-danger" onclick="return confirm('Are you sure you want to delete this table? This will permanently remove the table, its metadata, and all field metadata. This action cannot be undone.')">Delete Table</a>
				</div>
			</form>
		</div>
	</body>
	</html>`

	t, err := template.New("edit_table_metadata").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"TableName":   tableName,
		"DisplayName": tableMetadata.DisplayName,
		"Description": tableMetadata.Description,
		"ReadGroups":  tableMetadata.ReadGroups,
		"WriteGroups": tableMetadata.WriteGroups,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, data)
}

// HandleDeleteTable handles deleting a table and all its metadata
func (h *MetadataHandler) HandleDeleteTable(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Extract table name from URL
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/metadata/delete-table/"), "/")
	if len(pathParts) == 0 {
		http.Error(w, "Table name required", http.StatusBadRequest)
		return
	}
	tableName := pathParts[0]

	// Check user permissions - only admin and engineer can delete tables
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

	isAdmin, _ := h.db.IsUserInGroup(userID, "admin")
	isEngineer, _ := h.db.IsUserInGroup(userID, "engineer")

	if !isAdmin && !isEngineer {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Prevent deletion of system tables
	systemTables := []string{"_user", "_group", "_user_and_group", "_session", "_table_metadata", "_field_metadata", "_page"}
	for _, sysTable := range systemTables {
		if tableName == sysTable {
			http.Error(w, "Cannot delete system tables", http.StatusForbidden)
			return
		}
	}

	// Delete the table and all its metadata
	if err := h.db.DeleteTableMetadata(tableName); err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error deleting table", http.StatusInternalServerError)
		return
	}

	// Redirect back to tables list
	http.Redirect(w, r, "/metadata/tables", http.StatusSeeOther)
} 

// HandleCreateTable handles creating new tables with metadata
func (h *MetadataHandler) HandleCreateTable(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Check user permissions - only admin and engineer can create tables
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

	isAdmin, _ := h.db.IsUserInGroup(userID, "admin")
	isEngineer, _ := h.db.IsUserInGroup(userID, "engineer")

	if !isAdmin && !isEngineer {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Handle form submission
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Get table information
		tableName := r.FormValue("table_name")
		displayName := r.FormValue("display_name")
		description := r.FormValue("description")
		readGroups := r.FormValue("read_groups")
		writeGroups := r.FormValue("write_groups")

		// Validate table name
		if tableName == "" {
			http.Error(w, "Table name is required", http.StatusBadRequest)
			return
		}

		// Check if table already exists
		existingTable, err := h.db.GetTableMetadata(tableName)
		if err == nil && existingTable != nil {
			http.Error(w, "Table already exists", http.StatusBadRequest)
			return
		}

		// Parse fields from form
		var fields []models.FieldMetadata
		fieldCount := 0
		for {
			fieldName := r.FormValue(fmt.Sprintf("fields[%d][field_name]", fieldCount))
			if fieldName == "" {
				break
			}

			// Skip management fields
			if fieldName == "id" || fieldName == "created" || fieldName == "modified" || 
			   fieldName == "read_groups" || fieldName == "write_groups" {
				fieldCount++
				continue
			}

			displayName := r.FormValue(fmt.Sprintf("fields[%d][display_name]", fieldCount))
			fieldDescription := r.FormValue(fmt.Sprintf("fields[%d][description]", fieldCount))
			dbType := r.FormValue(fmt.Sprintf("fields[%d][db_type]", fieldCount))
			htmlInputType := r.FormValue(fmt.Sprintf("fields[%d][html_input_type]", fieldCount))
			formPositionStr := r.FormValue(fmt.Sprintf("fields[%d][form_position]", fieldCount))
			listPositionStr := r.FormValue(fmt.Sprintf("fields[%d][list_position]", fieldCount))
			isRequiredStr := r.FormValue(fmt.Sprintf("fields[%d][is_required]", fieldCount))
			isReadOnlyStr := r.FormValue(fmt.Sprintf("fields[%d][is_read_only]", fieldCount))
			defaultValue := r.FormValue(fmt.Sprintf("fields[%d][default_value]", fieldCount))
			validationRules := r.FormValue(fmt.Sprintf("fields[%d][validation_rules]", fieldCount))

			formPosition, _ := strconv.Atoi(formPositionStr)
			listPosition, _ := strconv.Atoi(listPositionStr)
			isRequired := isRequiredStr == "on"
			isReadOnly := isReadOnlyStr == "on"

			field := models.FieldMetadata{
				TableName:       tableName,
				FieldName:       fieldName,
				DisplayName:     displayName,
				Description:     fieldDescription,
				DBType:          dbType,
				HTMLInputType:   htmlInputType,
				FormPosition:    formPosition,
				ListPosition:    listPosition,
				IsRequired:      isRequired,
				IsReadOnly:      isReadOnly,
				DefaultValue:    defaultValue,
				ValidationRules: validationRules,
			}

			fields = append(fields, field)
			fieldCount++
		}

		// Create the table with metadata
		if err := h.db.CreateTableWithMetadata(tableName, displayName, description, readGroups, writeGroups, fields); err != nil {
			database.LogSQLError(err)
			http.Error(w, "Error creating table", http.StatusInternalServerError)
			return
		}

		// Redirect back to tables list
		http.Redirect(w, r, "/metadata/tables", http.StatusSeeOther)
		return
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		response := map[string]interface{}{
			"message": "Create table form",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Create New Table - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 1000px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.form-group { margin-bottom: 1rem; }
			label { display: block; margin-bottom: 0.5rem; font-weight: 600; }
			input, textarea, select { width: 100%; padding: 0.75rem; border: 1px solid #ddd; border-radius: 4px; font-size: 1rem; }
			textarea { min-height: 100px; }
			.btn { padding: 0.75rem 1.5rem; border: none; border-radius: 4px; text-decoration: none; font-size: 1rem; cursor: pointer; margin-right: 0.5rem; }
			.btn-primary { background: #667eea; color: white; }
			.btn-secondary { background: #6c757d; color: white; }
			.btn-success { background: #28a745; color: white; }
			.btn:hover { opacity: 0.8; }
			.help-text { font-size: 0.9rem; color: #6c757d; margin-top: 0.25rem; }
			.field-section { border: 1px solid #e9ecef; border-radius: 4px; padding: 1rem; margin-bottom: 1rem; }
			.field-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
			.field-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
			.remove-field { background: #dc3545; color: white; border: none; padding: 0.25rem 0.5rem; border-radius: 4px; cursor: pointer; }
			.add-field { background: #28a745; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; margin-top: 1rem; }
			.checkbox-group { display: flex; gap: 1rem; }
			.checkbox-group label { display: flex; align-items: center; font-weight: normal; }
			.checkbox-group input { width: auto; margin-right: 0.5rem; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Create New Table</h1>
			<form method="POST" id="createTableForm">
				<div class="form-group">
					<label for="table_name">Table Name</label>
					<input type="text" name="table_name" id="table_name" required>
					<div class="help-text">Database table name (e.g., products, orders, customers)</div>
				</div>
				<div class="form-group">
					<label for="display_name">Display Name</label>
					<input type="text" name="display_name" id="display_name" required>
					<div class="help-text">Human-readable name for the table</div>
				</div>
				<div class="form-group">
					<label for="description">Description</label>
					<textarea name="description" id="description"></textarea>
					<div class="help-text">Description of what this table contains</div>
				</div>
				<div class="form-group">
					<label for="read_groups">Read Groups (JSON array)</label>
					<textarea name="read_groups" id="read_groups">["admin", "engineer"]</textarea>
					<div class="help-text">Example: ["admin", "engineer", "everyone"]</div>
				</div>
				<div class="form-group">
					<label for="write_groups">Write Groups (JSON array)</label>
					<textarea name="write_groups" id="write_groups">["admin", "engineer"]</textarea>
					<div class="help-text">Example: ["admin", "engineer"]</div>
				</div>

				<h2>Fields</h2>
				<div id="fieldsContainer">
					<div class="field-section" data-field-index="0">
						<div class="field-header">
							<h3>Field 1</h3>
							<button type="button" class="remove-field" onclick="removeField(this)">Remove</button>
						</div>
						<div class="field-grid">
							<div class="form-group">
								<label>Field Name</label>
								<input type="text" name="fields[0][field_name]" required>
								<div class="help-text">Database field name (e.g., name, email, price)</div>
							</div>
							<div class="form-group">
								<label>Display Name</label>
								<input type="text" name="fields[0][display_name]" required>
								<div class="help-text">Human-readable field name</div>
							</div>
							<div class="form-group">
								<label>Description</label>
								<input type="text" name="fields[0][description]">
								<div class="help-text">Field description</div>
							</div>
							<div class="form-group">
								<label>DB Type</label>
								<select name="fields[0][db_type]" required>
									<option value="VARCHAR(255)">VARCHAR(255)</option>
									<option value="TEXT">TEXT</option>
									<option value="INT">INT</option>
									<option value="DECIMAL(10,2)">DECIMAL(10,2)</option>
									<option value="BOOLEAN">BOOLEAN</option>
									<option value="DATE">DATE</option>
									<option value="DATETIME">DATETIME</option>
									<option value="TIMESTAMP">TIMESTAMP</option>
								</select>
							</div>
							<div class="form-group">
								<label>HTML Input Type</label>
								<select name="fields[0][html_input_type]" required>
									<option value="text">text</option>
									<option value="email">email</option>
									<option value="password">password</option>
									<option value="number">number</option>
									<option value="textarea">textarea</option>
									<option value="select">select</option>
									<option value="checkbox">checkbox</option>
									<option value="datetime-local">datetime-local</option>
									<option value="date">date</option>
								</select>
							</div>
							<div class="form-group">
								<label>Form Position</label>
								<input type="number" name="fields[0][form_position]" value="1" min="0">
								<div class="help-text">Position in edit form (0-based)</div>
							</div>
							<div class="form-group">
								<label>List Position</label>
								<input type="number" name="fields[0][list_position]" value="1" min="-1">
								<div class="help-text">Position in table listing (-1 to hide)</div>
							</div>
							<div class="form-group">
								<label>Default Value</label>
								<input type="text" name="fields[0][default_value]">
								<div class="help-text">Default value for the field</div>
							</div>
							<div class="form-group">
								<label>Validation Rules</label>
								<textarea name="fields[0][validation_rules]"></textarea>
								<div class="help-text">JSON validation rules</div>
							</div>
						</div>
						<div class="checkbox-group">
							<label>
								<input type="checkbox" name="fields[0][is_required]">
								Required
							</label>
							<label>
								<input type="checkbox" name="fields[0][is_read_only]">
								Read Only
							</label>
						</div>
					</div>
				</div>

				<button type="button" class="add-field" onclick="addField()">Add Field</button>

				<div class="form-group" style="margin-top: 2rem;">
					<button type="submit" class="btn btn-primary">Create Table</button>
					<a href="/metadata/tables" class="btn btn-secondary">Cancel</a>
				</div>
			</form>
		</div>

		<script>
			let fieldIndex = 1;

			function addField() {
				const container = document.getElementById('fieldsContainer');
				const newField = document.createElement('div');
				newField.className = 'field-section';
				newField.setAttribute('data-field-index', fieldIndex);
				
				newField.innerHTML = 
					'<div class="field-header">' +
						'<h3>Field ' + (fieldIndex + 1) + '</h3>' +
						'<button type="button" class="remove-field" onclick="removeField(this)">Remove</button>' +
					'</div>' +
					'<div class="field-grid">' +
						'<div class="form-group">' +
							'<label>Field Name</label>' +
							'<input type="text" name="fields[' + fieldIndex + '][field_name]" required>' +
							'<div class="help-text">Database field name (e.g., name, email, price)</div>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>Display Name</label>' +
							'<input type="text" name="fields[' + fieldIndex + '][display_name]" required>' +
							'<div class="help-text">Human-readable field name</div>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>Description</label>' +
							'<input type="text" name="fields[' + fieldIndex + '][description]">' +
							'<div class="help-text">Field description</div>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>DB Type</label>' +
							'<select name="fields[' + fieldIndex + '][db_type]" required>' +
								'<option value="VARCHAR(255)">VARCHAR(255)</option>' +
								'<option value="TEXT">TEXT</option>' +
								'<option value="INT">INT</option>' +
								'<option value="DECIMAL(10,2)">DECIMAL(10,2)</option>' +
								'<option value="BOOLEAN">BOOLEAN</option>' +
								'<option value="DATE">DATE</option>' +
								'<option value="DATETIME">DATETIME</option>' +
								'<option value="TIMESTAMP">TIMESTAMP</option>' +
							'</select>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>HTML Input Type</label>' +
							'<select name="fields[' + fieldIndex + '][html_input_type]" required>' +
								'<option value="text">text</option>' +
								'<option value="email">email</option>' +
								'<option value="password">password</option>' +
								'<option value="number">number</option>' +
								'<option value="textarea">textarea</option>' +
								'<option value="select">select</option>' +
								'<option value="checkbox">checkbox</option>' +
								'<option value="datetime-local">datetime-local</option>' +
								'<option value="date">date</option>' +
							'</select>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>Form Position</label>' +
							'<input type="number" name="fields[' + fieldIndex + '][form_position]" value="' + (fieldIndex + 1) + '" min="0">' +
							'<div class="help-text">Position in edit form (0-based)</div>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>List Position</label>' +
							'<input type="number" name="fields[' + fieldIndex + '][list_position]" value="' + (fieldIndex + 1) + '" min="-1">' +
							'<div class="help-text">Position in table listing (-1 to hide)</div>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>Default Value</label>' +
							'<input type="text" name="fields[' + fieldIndex + '][default_value]">' +
							'<div class="help-text">Default value for the field</div>' +
						'</div>' +
						'<div class="form-group">' +
							'<label>Validation Rules</label>' +
							'<textarea name="fields[' + fieldIndex + '][validation_rules]"></textarea>' +
							'<div class="help-text">JSON validation rules</div>' +
						'</div>' +
					'</div>' +
					'<div class="checkbox-group">' +
						'<label>' +
							'<input type="checkbox" name="fields[' + fieldIndex + '][is_required]">' +
							'Required' +
						'</label>' +
						'<label>' +
							'<input type="checkbox" name="fields[' + fieldIndex + '][is_read_only]">' +
							'Read Only' +
						'</label>' +
					'</div>';
				
				container.appendChild(newField);
				fieldIndex++;
			}

			function removeField(button) {
				const fieldSection = button.closest('.field-section');
				if (document.querySelectorAll('.field-section').length > 1) {
					fieldSection.remove();
				}
			}
		</script>
	</body>
	</html>`

	t, err := template.New("create_table").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, nil)
} 