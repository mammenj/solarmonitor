package web

import (
	"embed"
	"html/template"
	"time"
)

//go:embed static/*
var StaticFS embed.FS

//go:embed templates/*.html
var templateFS embed.FS

func InitTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b float64) float64 {
			return a + b
		},
		"formatDate": func(t time.Time) string {
			return t.Format("02-Jan-2006")
		},
		"formatShortDate": func(t time.Time) string {
			return t.Format("02/01/06")
		},
		"formatFloat": func(f float64) string {
			return template.HTMLEscapeString(template.HTMLEscapeString(""))
		},
		"todayDate": func() string {
			return time.Now().Format("2006-01-02")
		},
	}

	return template.New("base").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
}
