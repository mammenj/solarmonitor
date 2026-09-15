package web_test

import (
	"testing"

	"solarmonitor/internal/web"
)

func TestEmbeddedAssets(t *testing.T) {
	// 1. Verify template parsing
	tmpl, err := web.InitTemplates()
	if err != nil {
		t.Fatalf("failed to initialize embedded templates: %v", err)
	}

	if tmpl.Lookup("base.html") == nil {
		t.Errorf("expected base.html to be embedded and parsed")
	}

	// 2. Verify static JS assets exist inside embed.FS
	for _, filename := range []string{"static/js/htmx.min.js", "static/js/alpine.min.js"} {
		if _, err := web.StaticFS.Open(filename); err != nil {
			t.Errorf("failed to open embedded asset %s: %v", filename, err)
		}
	}
}
