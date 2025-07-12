package models

type Page struct {
	ID             int
	Slug           string
	Title          string
	MetaDescription string
	Header         string
	Navigation     string
	MainContent    string
	Sidebar        string
	Footer         string
	CSSClass       string
	Scripts        string
	Template       string
} 