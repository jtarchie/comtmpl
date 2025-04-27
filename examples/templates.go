package examples

import (
	"fmt"
	"github.com/jtarchie/comtmpl/templates"
	"io"
)

var Parsed = templates.NewTemplates(map[string]templates.Template{
	"index.html": func(t *templates.Templates, writer io.Writer, data any) error {
		var err error

		//index.html:1

		_, err = io.WriteString(writer, `<html>
  <head>
    <title>`)
		if err != nil {
			return err
		}

		//index.html:3

		// Handle {{.Title}}
		var pipeValue0 any
		pipeValue0, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue0)
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

		// Handle {{.Title}}
		var pipeValue1 any
		pipeValue1, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue1)
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

		// Handle {{.User.Name}}
		var pipeValue2 any
		pipeValue2, err = templates.EvalField(data, []string{"User", "Name"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue2)
		if err != nil {
			return err
		}

		//index.html:7

		_, err = io.WriteString(writer, `!</p>
  </body>
</html>
`)
		if err != nil {
			return err
		}

		return nil
	},
	"pipe.html": func(t *templates.Templates, writer io.Writer, data any) error {
		var err error

		//pipe.html:1

		_, err = io.WriteString(writer, `<html>
  <head>
    <title>`)
		if err != nil {
			return err
		}

		//pipe.html:3

		// Handle {{.Title | upper}}
		var pipeValue0 any
		pipeValue0, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		// Pipe to function upper
		pipeValue0, err = t.CallFunc("upper", pipeValue0)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue0)
		if err != nil {
			return err
		}

		//pipe.html:3

		_, err = io.WriteString(writer, `</title>
  </head>
  <body>
    <h1>`)
		if err != nil {
			return err
		}

		//pipe.html:6

		// Handle {{.Title}}
		var pipeValue1 any
		pipeValue1, err = templates.EvalField(data, []string{"Title"})
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue1)
		if err != nil {
			return err
		}

		//pipe.html:6

		_, err = io.WriteString(writer, `</h1>
    <p>Name length: `)
		if err != nil {
			return err
		}

		//pipe.html:7

		// Handle {{.User.Name | len}}
		var pipeValue2 any
		pipeValue2, err = templates.EvalField(data, []string{"User", "Name"})
		if err != nil {
			return err
		}
		// Pipe to function len
		pipeValue2, err = t.CallFunc("len", pipeValue2)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue2)
		if err != nil {
			return err
		}

		//pipe.html:7

		_, err = io.WriteString(writer, `</p>
    <p>`)
		if err != nil {
			return err
		}

		//pipe.html:8

		// Handle {{.User.Description | title}}
		var pipeValue3 any
		pipeValue3, err = templates.EvalField(data, []string{"User", "Description"})
		if err != nil {
			return err
		}
		// Pipe to function title
		pipeValue3, err = t.CallFunc("title", pipeValue3)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(writer, pipeValue3)
		if err != nil {
			return err
		}

		//pipe.html:8

		_, err = io.WriteString(writer, `</p>
  </body>
</html>
`)
		if err != nil {
			return err
		}

		return nil
	},
})
