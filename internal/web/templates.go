package web

import (
	"embed"
	"encoding/json"
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
		"subf": func(a, b float64) float64 {
			return a - b
		},

		"add": func(a, b int) int {
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
		// Maps a value against a maximum scale to output SVG pixel heights
		"scaleHeight": func(value, max, maxHeight float64) float64 {
			if max == 0 {
				return 0
			}
			return (value / max) * maxHeight
		},
		"mul":  func(a int, b int) int { return a * b },
		"mulf": func(a, b float64) float64 { return a * b },
		"marshalJSON": func(v any) (template.JS, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			return template.JS(b), nil
		},
	}

	return template.New("base").Funcs(funcMap).ParseFS(templateFS, "templates/*.html")
}
