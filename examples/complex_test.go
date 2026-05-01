package examples_test

import (
	"bytes"
	"strings"
	"testing"

	examples "github.com/jtarchie/comtmpl/examples"
)

func TestComplexTemplate(t *testing.T) {
	testCases := []struct {
		name     string
		data     map[string]interface{}
		expected string
	}{
		{
			name: "complete data",
			data: map[string]interface{}{
				"Title":       "Test Page",
				"Year":        2025,
				"Company":     "Acme Corp",
				"Description": "This is a description of the page.",
				"User": map[string]interface{}{
					"Name":  "John Doe",
					"Admin": true,
					"Contact": map[string]interface{}{
						"Email": "john@example.com",
						"Phone": "123-456-7890",
					},
				},
				"Items": []map[string]interface{}{
					{
						"Name":   "Product 1",
						"Price":  19.99,
						"OnSale": true,
						"Tags":   []string{"electronics", "gadget"},
					},
					{
						"Name":   "Product 2",
						"Price":  29.99,
						"OnSale": false,
						"Tags":   []string{"home", "kitchen"},
					},
					{
						"Name":  "Product 3",
						"Price": 39.99,
						"Tags":  []string{},
					},
				},
			},
			expected: `<!DOCTYPE html>
<html>
<head>
  <title>Test Page - Complex Template Example</title>
  <style>
    .highlight { color: blue; }
    .error { color: red; }
  </style>
</head>
<body>
  <h1>Test Page</h1>
  
  
  
  <!-- Conditional logic -->
  <div class="user-info">
    
      <p>Welcome, <span class="highlight">John Doe</span>!</p>
      
      
        <p class="highlight">You have admin privileges</p>
      
      
      
        <div class="contact">
          <h3>Contact Information:</h3>
          <p>Email: john@example.com</p>
          <p>Phone: 123-456-7890</p>
        </div>
      
    
  </div>
  
  <!-- Range loop for items -->
  <div class="items">
    <h2>Your Items:</h2>
    
      <div class="item">
        <h3>0. Product 1</h3>
        <p>Price: $19.99</p>
        
        
          <p class="highlight">ON SALE!</p>
        
        
        <!-- Nested range for item tags -->
        
          <p>Tags:</p>
          <ul>
            
              <li>electronics</li>
            
              <li>gadget</li>
            
          </ul>
        
      </div>
    
      <div class="item">
        <h3>1. Product 2</h3>
        <p>Price: $29.99</p>
        
        
        
        <!-- Nested range for item tags -->
        
          <p>Tags:</p>
          <ul>
            
              <li>home</li>
            
              <li>kitchen</li>
            
          </ul>
        
      </div>
    
      <div class="item">
        <h3>2. Product 3</h3>
        <p>Price: $39.99</p>
        
        
        
        <!-- Nested range for item tags -->
        
          <p>No tags available</p>
        
      </div>
    
  </div>
  
  <!-- Function calls and pipes -->
  <div class="footer">
    <p>Copyright &copy; 2025 ACME CORP</p>
    <p>This is a description of the page.</p>
  </div>
</body>
</html>`,
		},
		{
			name: "missing user and items",
			data: map[string]interface{}{
				"Title":       "Test Page",
				"Year":        2025,
				"Company":     "Acme Corp",
				"Description": "Short description.",
			},
			expected: `<!DOCTYPE html>
<html>
<head>
  <title>Test Page - Complex Template Example</title>
  <style>
    .highlight { color: blue; }
    .error { color: red; }
  </style>
</head>
<body>
  <h1>Test Page</h1>
  
  
  
  <!-- Conditional logic -->
  <div class="user-info">
    
      <p class="error">No user information available</p>
    
  </div>
  
  <!-- Range loop for items -->
  <div class="items">
    <h2>Your Items:</h2>
    
      <p class="error">No items in your cart</p>
    
  </div>
  
  <!-- Function calls and pipes -->
  <div class="footer">
    <p>Copyright &copy; 2025 ACME CORP</p>
    <p>Short description.</p>
  </div>
</body>
</html>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := examples.Parsed.ExecuteTemplate(&buf, "complex.html", tc.data)
			if err != nil {
				t.Fatalf("failed to execute template: %v", err)
			}

			got := strings.TrimSpace(buf.String())
			expected := strings.TrimSpace(tc.expected)

			if got != expected {
				t.Errorf("template output does not match expected.\n\nGot:\n%s\n\nExpected:\n%s",
					got, expected)

				// Find where they differ for easier debugging
				minLen := len(got)
				if len(expected) < minLen {
					minLen = len(expected)
				}

				for i := 0; i < minLen; i++ {
					if got[i] != expected[i] {
						t.Errorf("First difference at position %d: got '%c', expected '%c'",
							i, got[i], expected[i])

						start := i - 20
						if start < 0 {
							start = 0
						}
						end := i + 20
						if end > len(got) {
							end = len(got)
						}
						endExp := i + 20
						if endExp > len(expected) {
							endExp = len(expected)
						}

						t.Errorf("Context (got): %s", got[start:end])
						t.Errorf("Context (exp): %s", expected[start:endExp])
						break
					}
				}
			}
		})
	}
}
