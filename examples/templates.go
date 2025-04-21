package examples

import (
	"fmt"
	"github.com/jtarchie/comtmpl/templates"
	"io"
)

var Parsed = templates.Templates{
	"index.html": func(writer io.Writer, data any) error {
		var err error

		//index.html:1

		_, err = io.WriteString(writer, `<html>
  <head>
    <title>`)
		if err != nil {
			return err
		}

		//index.html:3

		// Handle {{Title}}
		var value0 any
		value0, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, value0)
		if err != nil {
			return err
		}

		//index.html:3

		_, err = io.WriteString(writer, `</title>
  </head>
  <body>
    <h1>`)
		if err != nil {
			return err
		}

		//index.html:6

		// Handle {{Title}}
		var value1 any
		value1, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, value1)
		if err != nil {
			return err
		}

		//index.html:6

		_, err = io.WriteString(writer, `</h1>
    <p>Welcome, `)
		if err != nil {
			return err
		}

		//index.html:7

		// Handle {{User.Name}}
		var value2 any
		value2, err = templates.EvalField(data, []string{"User", "Name"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, value2)
		if err != nil {
			return err
		}

		//index.html:7

		_, err = io.WriteString(writer, `!</p>
  </body>
</html>`)
		if err != nil {
			return err
		}

		return nil
	},
}
