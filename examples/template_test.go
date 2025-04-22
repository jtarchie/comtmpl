package examples_test

import (
	"bytes"
	"strings"
	"testing"

	templates "github.com/jtarchie/comtmpl/examples"
)

func TestDotNotation(t *testing.T) {
	testCases := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "simple field access",
			data: map[string]interface{}{
				"Title": "Hello World",
				"User": map[string]interface{}{
					"Name": "John",
				},
			},
			expected: `<html>
  <head>
    <title>Hello World</title>
  </head>
  <body>
    <h1>Hello World</h1>
    <p>Welcome, John!</p>
  </body>
</html>`,
		},
		{
			name: "missing fields",
			data: map[string]interface{}{
				"Title": "Hello World",
				// User is missing
			},
			expected: `<html>
  <head>
    <title>Hello World</title>
  </head>
  <body>
    <h1>Hello World</h1>
    <p>Welcome, !</p>
  </body>
</html>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := templates.Parsed.ExecuteTemplate(&buf, "index.html", tc.data)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}

			if strings.TrimSpace(buf.String()) != tc.expected {
				t.Errorf("template output does not match expected:\nGot:\n%s\n\nExpected:\n%s",
					buf.String(), tc.expected)
			}
		})
	}
}
