package templates_test

import (
	"bytes"
	"html/template"
	"testing"

	templates "github.com/jtarchie/comtmpl/examples"
)

func BenchmarkStandardTemplate(b *testing.B) {
	parsedTemplates, err := template.ParseGlob("*.html")
	if err != nil {
		b.Fatalf("failed to parse templates: %v", err)
	}

	var buf bytes.Buffer
	data := map[string]interface{}{
		"Title": "Hello, World!",
		"User": map[string]interface{}{
			"Name": "John Doe",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		err = parsedTemplates.ExecuteTemplate(&buf, "index.html", data)
		if err != nil {
			b.Fatalf("failed to execute template: %v", err)
		}
	}
}

func BenchmarkCustomTemplate(b *testing.B) {
	var buf bytes.Buffer
	data := map[string]interface{}{
		"Title": "Hello, World!",
		"User": map[string]interface{}{
			"Name": "John Doe",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := templates.Parsed.ExecuteTemplate(&buf, "index.html", data)
		if err != nil {
			b.Fatalf("failed to execute template: %v", err)
		}
	}
}
