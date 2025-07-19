package handlers

import (
	"html/template"
	"net/http"
	"stingray/templates"
)

// RenderMessage renders a message page with the given parameters
func RenderMessage(w http.ResponseWriter, title, header, headerClass, message, buttonURL, buttonText string, status int) {
	data := struct {
		Title       string
		MetaDescription string
		Header      string
		HeaderClass string
		Message     string
		ButtonURL   string
		ButtonText  string
		Footer      string
	}{
		Title: title,
		MetaDescription: header,
		Header: header,
		HeaderClass: headerClass,
		Message: message,
		ButtonURL: buttonURL,
		ButtonText: buttonText,
		Footer: "© 2025 StingRay",
	}
	tmplContent, err := templates.LoadTemplate("message")
	if err != nil {
		http.Error(w, "Message template not found", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.New("message").Parse(tmplContent)
	if err != nil {
		http.Error(w, "Error parsing message template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	tmpl.Execute(w, data)
} 