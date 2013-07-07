package layouts

import (
	"html/template"
	"io"
)

var DefaultGroup = New("")

func SetPath(dir string) {
	DefaultGroup.SetPath(dir)
}

func Files(files ...string) error {
	return DefaultGroup.Files(files...)
}

func Glob(patterns ...string) error {
	return DefaultGroup.Glob(patterns...)
}

func Clear() {
	DefaultGroup.Clear()
}

func Entries() []Entry {
	return DefaultGroup.Entries()
}

func Execute(w io.Writer, layout string, t *template.Template, data interface{}) error {
	return DefaultGroup.Execute(w, layout, t, data)
}

func ExecuteHTML(w io.Writer, layout string, content template.HTML, data interface{}) error {
	return DefaultGroup.ExecuteHTML(w, layout, content, data)
}

func Funcs(f template.FuncMap) {
	DefaultGroup.Funcs(f)
}
