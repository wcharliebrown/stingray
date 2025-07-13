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
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	// Get all table metadata
	tableMetadata, err := h.db.GetAllTableMetadata()
	if err != nil {
		database.LogSQLError(err)
		http.Error(w, "Error fetching table metadata", http.StatusInternalServerError)
		return
	}

	// Check if JSON response is requested
	if r.URL.Query().Get("response_format") == "json" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tableMetadata)
		return
	}

	// HTML response
	tmpl := `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Database Tables - Sting Ray</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; background: #f5f5f5; margin: 0; padding: 2rem; }
			.container { max-width: 1200px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
			h1 { color: #2c3e50; margin-bottom: 1rem; }
			.table-list { list-style: none; padding: 0; }
			.table-item { padding: 1rem; border-bottom: 1px solid #e9ecef; display: flex; justify-content: space-between; align-items: center; }
			.table-item:last-child { border-bottom: none; }
			.table-info { flex: 1; }
			.table-name { font-weight: 600; color: #667eea; margin-bottom: 0.25rem; }
			.table-description { color: #6c757d; font-size: 0.9rem; }
			.table-actions { display: flex; gap: 0.5rem; }
			.btn { padding: 0.5rem 1rem; border: none; border-radius: 4px; text-decoration: none; font-size: 0.9rem; cursor: pointer; }
			.btn-primary { background: #667eea; color: white; }
			.btn-secondary { background: #6c757d; color: white; }
			.btn:hover { opacity: 0.8; }
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Database Tables</h1>
			<ul class="table-list">
				{{range .}}
				<li class="table-item">
					<div class="table-info">
						<div class="table-name">{{.DisplayName}}</div>
						<div class="table-description">{{.Description}}</div>
					</div>
					<div class="table-actions">
						<a href="/metadata/table/{{.TableName}}" class="btn btn-primary">View Data</a>
						<a href="/metadata/edit/{{.TableName}}" class="btn btn-secondary">Edit</a>
					</div>
				</li>
				{{end}}
			</ul>
		</div>
	</body>
	</html>`

	t, err := template.New("tables").Parse(tmpl)
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.Execute(w, tableMetadata)
}

// HandleTableData handles viewing table data
func (h *MetadataHandler) HandleTableData(w http.ResponseWriter, r *http.Request) {
	// Check if user is authenticated
	if !h.sm.IsAuthenticated(r) {
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

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

	// Check user permissions
	session, err := h.sm.GetSessionFromRequest(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	userID := session.UserID

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
			if inGroup, _ := h.db.IsUserInGroup(userID, group); inGroup {
				canEdit = true
				canDelete = true
				canCreate = true
				break
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
			.container { max-width: 1400px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
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