package examples

import (
	"fmt"
	"github.com/jtarchie/comtmpl/templates"
	"io"
)

var Parsed = templates.NewTemplates(map[string]templates.Template{
	"complex.html": func(t *templates.Templates, writer io.Writer, data any) error {
		var err error

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:1
		_, err = io.WriteString(writer, "<!DOCTYPE html>\n<html>\n<head>\n  <title>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:4
		var result0 any
		result0, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result0)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:4
		_, err = io.WriteString(writer, " - Complex Template Example</title>\n  <style>\n    .highlight { color: blue; }\n    .error { color: red; }\n  </style>\n</head>\n<body>\n  <h1>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:11
		var result1 any
		result1, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result1)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:11
		_, err = io.WriteString(writer, "</h1>\n  \n  ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:13
		_, err = io.WriteString(writer, "\n  \n  <!-- Conditional logic -->\n  <div class=\"user-info\">\n    ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:17
		// If statement
		var cond2 bool
		var ifResult3 any
		ifResult3, err = templates.EvalField(data, []string{"User"})
		if err != nil {
			return err
		}
		cond2, err = templates.IsTrue(ifResult3)
		if err != nil {
			return err
		}
		if cond2 {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:17
			_, err = io.WriteString(writer, "\n      <p>Welcome, <span class=\"highlight\">")
			if err != nil {
				return err
			}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:18
			var result4 any
			result4, err = templates.EvalField(data, []string{"User", "Name"})
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(writer, result4)
			if err != nil {
				return err
			}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:18
			_, err = io.WriteString(writer, "</span>!</p>\n      \n      ")
			if err != nil {
				return err
			}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:20
			// If statement
			var cond5 bool
			var ifResult6 any
			ifResult6, err = templates.EvalField(data, []string{"User", "Admin"})
			if err != nil {
				return err
			}
			cond5, err = templates.IsTrue(ifResult6)
			if err != nil {
				return err
			}
			if cond5 {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:20
				_, err = io.WriteString(writer, "\n        <p class=\"highlight\">You have admin privileges</p>\n      ")
				if err != nil {
					return err
				}
			} else {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:22
				_, err = io.WriteString(writer, "\n        <p>You are a regular user</p>\n      ")
				if err != nil {
					return err
				}
			}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:24
			_, err = io.WriteString(writer, "\n      \n      ")
			if err != nil {
				return err
			}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:26
			// With statement
			var withData7 any
			withData7, err = templates.EvalField(data, []string{"User", "Contact"})
			if err != nil {
				return err
			}
			withCond8, err := templates.IsTrue(withData7)
			if err != nil {
				return err
			}
			if withCond8 {
				// Save old data context and set new one
				oldData := data
				data = withData7

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:26
				_, err = io.WriteString(writer, "\n        <div class=\"contact\">\n          <h3>Contact Information:</h3>\n          <p>Email: ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:29
				var result9 any
				result9, err = templates.EvalField(data, []string{"Email"})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(writer, result9)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:29
				_, err = io.WriteString(writer, "</p>\n          <p>Phone: ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:30
				var result10 any
				result10, err = templates.EvalField(data, []string{"Phone"})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(writer, result10)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:30
				_, err = io.WriteString(writer, "</p>\n        </div>\n      ")
				if err != nil {
					return err
				}
				data = oldData
			} else {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:32
				_, err = io.WriteString(writer, "\n        <p class=\"error\">No contact information available</p>\n      ")
				if err != nil {
					return err
				}
			}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:34
			_, err = io.WriteString(writer, "\n    ")
			if err != nil {
				return err
			}
		} else {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:35
			_, err = io.WriteString(writer, "\n      <p class=\"error\">No user information available</p>\n    ")
			if err != nil {
				return err
			}
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:37
		_, err = io.WriteString(writer, "\n  </div>\n  \n  <!-- Range loop for items -->\n  <div class=\"items\">\n    <h2>Your Items:</h2>\n    ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:43
		// Range statement
		var rangeData11 any
		rangeData11, err = templates.EvalField(data, []string{"Items"})
		if err != nil {
			return err
		}
		iter12, err := templates.GetIterable(rangeData11)
		if err != nil {
			return err
		}
		hasItems := false
		var var_index any // Template variable for index: $index
		var var_item any  // Template variable for value: $item

		if mapData, isMap := iter12.(map[string]any); isMap {
			for k, v := range mapData {
				hasItems = true
				var_index = k
				var_item = v

				// Create range scope
				rangeContext := templates.NewRangeScope(data, k, v)
				oldData := data
				data = rangeContext
				// Range body

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:43
				_, err = io.WriteString(writer, "\n      <div class=\"item\">\n        <h3>")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				var result14 any = var_index // Variable reference
				_, err = fmt.Fprint(writer, result14)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				_, err = io.WriteString(writer, ". ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				var result15 any
				result15, err = templates.EvalField(var_item, []string{"Name"})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(writer, result15)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				_, err = io.WriteString(writer, "</h3>\n        <p>Price: $")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:46
				var result16 any
				result16, err = templates.EvalField(var_item, []string{"Price"})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(writer, result16)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:46
				_, err = io.WriteString(writer, "</p>\n        \n        ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:48
				// If statement
				var cond17 bool
				var ifResult18 any
				ifResult18, err = templates.EvalField(var_item, []string{"OnSale"})
				if err != nil {
					return err
				}
				cond17, err = templates.IsTrue(ifResult18)
				if err != nil {
					return err
				}
				if cond17 {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:48
					_, err = io.WriteString(writer, "\n          <p class=\"highlight\">ON SALE!</p>\n        ")
					if err != nil {
						return err
					}
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:50
				_, err = io.WriteString(writer, "\n        \n        <!-- Nested range for item tags -->\n        ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:53
				// If statement
				var cond19 bool
				var ifResult20 any
				ifResult20, err = templates.EvalField(var_item, []string{"Tags"})
				if err != nil {
					return err
				}
				cond19, err = templates.IsTrue(ifResult20)
				if err != nil {
					return err
				}
				if cond19 {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:53
					_, err = io.WriteString(writer, "\n          <p>Tags:</p>\n          <ul>\n            ")
					if err != nil {
						return err
					}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:56
					// Range statement
					var rangeData21 any
					rangeData21, err = templates.EvalField(var_item, []string{"Tags"})
					if err != nil {
						return err
					}
					iter22, err := templates.GetIterable(rangeData21)
					if err != nil {
						return err
					}

					if mapData, isMap := iter22.(map[string]any); isMap {
						for k, v := range mapData {

							// Create range scope
							rangeContext := templates.NewRangeScope(data, k, v)
							oldData := data
							data = rangeContext
							// Range body

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:56
							_, err = io.WriteString(writer, "\n              <li>")
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							var result24 any = templates.Dot(data)
							_, err = fmt.Fprint(writer, result24)
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							_, err = io.WriteString(writer, "</li>\n            ")
							if err != nil {
								return err
							}
							data = oldData
						}

					} else if sliceData, isSlice := iter22.([]any); isSlice {
						for i, v := range sliceData {

							// Create range scope
							rangeContext := templates.NewRangeScope(data, i, v)
							oldData := data
							data = rangeContext
							// Range body

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:56
							_, err = io.WriteString(writer, "\n              <li>")
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							var result25 any = templates.Dot(data)
							_, err = fmt.Fprint(writer, result25)
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							_, err = io.WriteString(writer, "</li>\n            ")
							if err != nil {
								return err
							}
							data = oldData
						}
					}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:58
					_, err = io.WriteString(writer, "\n          </ul>\n        ")
					if err != nil {
						return err
					}
				} else {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:60
					_, err = io.WriteString(writer, "\n          <p>No tags available</p>\n        ")
					if err != nil {
						return err
					}
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:62
				_, err = io.WriteString(writer, "\n      </div>\n    ")
				if err != nil {
					return err
				}
				data = oldData
			}

		} else if sliceData, isSlice := iter12.([]any); isSlice {
			for i, v := range sliceData {
				hasItems = true
				var_index = i
				var_item = v

				// Create range scope
				rangeContext := templates.NewRangeScope(data, i, v)
				oldData := data
				data = rangeContext
				// Range body

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:43
				_, err = io.WriteString(writer, "\n      <div class=\"item\">\n        <h3>")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				var result26 any = var_index // Variable reference
				_, err = fmt.Fprint(writer, result26)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				_, err = io.WriteString(writer, ". ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				var result27 any
				result27, err = templates.EvalField(var_item, []string{"Name"})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(writer, result27)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:45
				_, err = io.WriteString(writer, "</h3>\n        <p>Price: $")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:46
				var result28 any
				result28, err = templates.EvalField(var_item, []string{"Price"})
				if err != nil {
					return err
				}
				_, err = fmt.Fprint(writer, result28)
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:46
				_, err = io.WriteString(writer, "</p>\n        \n        ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:48
				// If statement
				var cond29 bool
				var ifResult30 any
				ifResult30, err = templates.EvalField(var_item, []string{"OnSale"})
				if err != nil {
					return err
				}
				cond29, err = templates.IsTrue(ifResult30)
				if err != nil {
					return err
				}
				if cond29 {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:48
					_, err = io.WriteString(writer, "\n          <p class=\"highlight\">ON SALE!</p>\n        ")
					if err != nil {
						return err
					}
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:50
				_, err = io.WriteString(writer, "\n        \n        <!-- Nested range for item tags -->\n        ")
				if err != nil {
					return err
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:53
				// If statement
				var cond31 bool
				var ifResult32 any
				ifResult32, err = templates.EvalField(var_item, []string{"Tags"})
				if err != nil {
					return err
				}
				cond31, err = templates.IsTrue(ifResult32)
				if err != nil {
					return err
				}
				if cond31 {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:53
					_, err = io.WriteString(writer, "\n          <p>Tags:</p>\n          <ul>\n            ")
					if err != nil {
						return err
					}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:56
					// Range statement
					var rangeData33 any
					rangeData33, err = templates.EvalField(var_item, []string{"Tags"})
					if err != nil {
						return err
					}
					iter34, err := templates.GetIterable(rangeData33)
					if err != nil {
						return err
					}

					if mapData, isMap := iter34.(map[string]any); isMap {
						for k, v := range mapData {

							// Create range scope
							rangeContext := templates.NewRangeScope(data, k, v)
							oldData := data
							data = rangeContext
							// Range body

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:56
							_, err = io.WriteString(writer, "\n              <li>")
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							var result36 any = templates.Dot(data)
							_, err = fmt.Fprint(writer, result36)
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							_, err = io.WriteString(writer, "</li>\n            ")
							if err != nil {
								return err
							}
							data = oldData
						}

					} else if sliceData, isSlice := iter34.([]any); isSlice {
						for i, v := range sliceData {

							// Create range scope
							rangeContext := templates.NewRangeScope(data, i, v)
							oldData := data
							data = rangeContext
							// Range body

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:56
							_, err = io.WriteString(writer, "\n              <li>")
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							var result37 any = templates.Dot(data)
							_, err = fmt.Fprint(writer, result37)
							if err != nil {
								return err
							}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:57
							_, err = io.WriteString(writer, "</li>\n            ")
							if err != nil {
								return err
							}
							data = oldData
						}
					}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:58
					_, err = io.WriteString(writer, "\n          </ul>\n        ")
					if err != nil {
						return err
					}
				} else {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:60
					_, err = io.WriteString(writer, "\n          <p>No tags available</p>\n        ")
					if err != nil {
						return err
					}
				}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:62
				_, err = io.WriteString(writer, "\n      </div>\n    ")
				if err != nil {
					return err
				}
				data = oldData
			}
		}
		if !hasItems {

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:64
			_, err = io.WriteString(writer, "\n      <p class=\"error\">No items in your cart</p>\n    ")
			if err != nil {
				return err
			}
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:66
		_, err = io.WriteString(writer, "\n  </div>\n  \n  <!-- Function calls and pipes -->\n  <div class=\"footer\">\n    <p>Copyright &copy; ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:71
		var result38 any
		result38, err = templates.EvalField(data, []string{"Year"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result38)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:71
		_, err = io.WriteString(writer, " ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:71
		var result39 any
		result39, err = templates.EvalField(data, []string{"Company"})
		if err != nil {
			return err
		}
		result39, err = t.CallFunc("upper", result39)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result39)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:71
		_, err = io.WriteString(writer, "</p>\n    <p>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:72
		var result40 any
		result40, err = templates.EvalField(data, []string{"Description"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result40)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/complex.html:72
		_, err = io.WriteString(writer, "</p>\n  </div>\n</body>\n</html>")
		if err != nil {
			return err
		}

		return nil
	},
	"index.html": func(t *templates.Templates, writer io.Writer, data any) error {
		var err error

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:1
		_, err = io.WriteString(writer, "<html>\n  <head>\n    <title>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:3
		var result0 any
		result0, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result0)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:3
		_, err = io.WriteString(writer, "</title>\n  </head>\n  <body>\n    <h1>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:6
		var result1 any
		result1, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result1)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:6
		_, err = io.WriteString(writer, "</h1>\n    <p>Welcome, ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:7
		var result2 any
		result2, err = templates.EvalField(data, []string{"User", "Name"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result2)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/index.html:7
		_, err = io.WriteString(writer, "!</p>\n  </body>\n</html>\n")
		if err != nil {
			return err
		}

		return nil
	},
	"pipe.html": func(t *templates.Templates, writer io.Writer, data any) error {
		var err error

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:1
		_, err = io.WriteString(writer, "<html>\n  <head>\n    <title>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:3
		var result0 any
		result0, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		result0, err = t.CallFunc("upper", result0)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result0)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:3
		_, err = io.WriteString(writer, "</title>\n  </head>\n  <body>\n    <h1>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:6
		var result1 any
		result1, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result1)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:6
		_, err = io.WriteString(writer, "</h1>\n    <p>Name length: ")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:7
		var result2 any
		result2, err = templates.EvalField(data, []string{"User", "Name"})
		if err != nil {
			return err
		}
		result2, err = t.CallFunc("len", result2)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result2)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:7
		_, err = io.WriteString(writer, "</p>\n    <p>")
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:8
		var result3 any
		result3, err = templates.EvalField(data, []string{"User", "Description"})
		if err != nil {
			return err
		}
		result3, err = t.CallFunc("title", result3)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, result3)
		if err != nil {
			return err
		}

//line /Users/jtarchie/workspace/comtmpl/examples/pipe.html:8
		_, err = io.WriteString(writer, "</p>\n  </body>\n</html>\n")
		if err != nil {
			return err
		}

		return nil
	},
})
