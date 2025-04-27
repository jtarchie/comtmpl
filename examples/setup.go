package examples

import (
	"github.com/go-task/slim-sprig/v3"
)

func init() {
	Parsed.Funcs(sprig.FuncMap())
}
