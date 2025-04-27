package examples_test

import (
	"bytes"
	"html/template"
	"testing"

	sprig "github.com/go-task/slim-sprig/v3"
	examples "github.com/jtarchie/comtmpl/examples"
)

func BenchmarkStandardTemplate(b *testing.B) {
	parsedTemplates, err := template.New("").Funcs(sprig.FuncMap()).ParseGlob("*.html")
	if err != nil {
		b.Fatalf("failed to parse templates: %v", err)
	}

	data := map[string]interface{}{
		"Title": "Hello, World!",
		"User": map[string]interface{}{
			"Name":        "John Doe",
			"Description": "This is a very long description about the user",
		},
	}

	templates := []string{"index.html", "pipe.html"}
	for _, tmpl := range templates {
		b.Run(tmpl, func(b *testing.B) {
			var buf bytes.Buffer
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				err = parsedTemplates.ExecuteTemplate(&buf, tmpl, data)
				if err != nil {
					b.Fatalf("failed to execute template %s: %v", tmpl, err)
				}
			}
		})
	}
}

func BenchmarkCustomTemplate(b *testing.B) {
	data := map[string]interface{}{
		"Title": "Hello, World!",
		"User": map[string]interface{}{
			"Name":        "John Doe",
			"Description": "This is a very long description about the user",
		},
	}

	templates := []string{"index.html", "pipe.html"}
	for _, tmpl := range templates {
		b.Run(tmpl, func(b *testing.B) {
			var buf bytes.Buffer
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				buf.Reset()
				err := examples.Parsed.ExecuteTemplate(&buf, tmpl, data)
				if err != nil {
					b.Fatalf("failed to execute template %s: %v", tmpl, err)
				}
			}
		})
	}
}
