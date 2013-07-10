package layouts

import (
	"html/template"
	"io"
)

var DefaultGroup = New("")

// SetPath sets a new path for layouts to be loaded from. This should be called
// before any layouts are loaded.
/*
func SetPath(dir string) {
	DefaultGroup.SetPath(dir)
}
*/

// Files loads layouts by individual file names. Each layout's name
// comes from the file name without its extension.
func Files(files ...string) error {
	return DefaultGroup.Files(files...)
}

// Glob loads layouts through pattern matching. Each layout's name
// comes from the file name without its extension.
func Glob(patterns ...string) error {
	return DefaultGroup.Glob(patterns...)
}

// Clear unloads all layouts.
func Clear() {
	DefaultGroup.Clear()
}

// Entries returns info about all loaded layouts.
func Entries() []Entry {
	return DefaultGroup.Entries()
}

// Execute renders the specified template using the named layout,
// passing in data to the layout templates.
func Execute(w io.Writer, layout string, t *template.Template, data interface{}) error {
	return DefaultGroup.Execute(w, layout, t, data)
}

// ExecuteHTML renders the content string using the named layout,
// passing in data to the layout templates. Note that the content
// string is of type template.HTML; it is expected that the content
// string is safe, fully-escaped HTML.
func ExecuteHTML(w io.Writer, layout string, content template.HTML, data interface{}) error {
	return DefaultGroup.ExecuteHTML(w, layout, content, data)
}

// Funcs adds funcs to all layouts that are loaded. See template.Funcs in
// html/template. Note that any templates you pass in to Execute do not have
// these funcs applied.
func Funcs(f template.FuncMap) {
	DefaultGroup.Funcs(f)
}
