package examples_test

import (
	"bytes"
	"strings"
	"testing"

	examples "github.com/jtarchie/comtmpl/examples"
)

func TestPipeFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "pipe functions",
			data: map[string]interface{}{
				"Title": "Hello World",
				"User": map[string]interface{}{
					"Name":        "John Doe",
					"Description": "This is a very long description about the user",
				},
			},
			expected: `<html>
  <head>
    <title>HELLO WORLD</title>
  </head>
  <body>
    <h1>Hello World</h1>
    <p>Name length: 8</p>
    <p>This Is A Very Long Description About The User</p>
  </body>
</html>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := examples.Parsed.ExecuteTemplate(&buf, "pipe.html", tc.data)
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
