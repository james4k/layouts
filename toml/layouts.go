package layouts

import (
	"bytes"
	"errors"
	"github.com/james4k/fmatter/toml"
	"html/template"
	"io"
	"io/ioutil"
	"path/filepath"
	"sync"
)

type unexportedEntry struct {
	*template.Template
}

type Entry struct {
	unexportedEntry
	Name   string
	Path   string
	Parent string
}

var tmpls *template.Template
var dict map[string]Entry
var mu sync.Mutex

func init() {
	Clear()
}

func undefinedContent() interface{} {
	panic("content undefined")
}

// makeContentFunc makes a "content" func that is valid for exactly one call
func makeContentFunc(content template.HTML) interface{} {
	var fn func() interface{}
	fn = func() interface{} {
		fn = undefinedContent
		return content
	}
	return func() interface{} {
		return fn()
	}
}

func load(filename string) error {
	front := make(map[string]interface{}, 4)
	content, err := fmatter.ReadFile(filename, front)
	if err != nil {
		return err
	}

	path := filename
	filename = filepath.Base(filename)
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	_, ok := dict[name]
	if ok {
		return nil
	}

	var t *template.Template
	if tmpls == nil {
		t = template.New(name)
		tmpls = t
		t.Funcs(template.FuncMap{
			"content": undefinedContent,
		})
	} else {
		t = tmpls.New(name)
	}

	_, err = t.Parse(string(content))
	if err != nil {
		// hrm.. how do we remove a template..?
		return err
	}

	parent, _ := front["layout"].(string)
	dict[name] = Entry{unexportedEntry{t}, name, path, parent}
	return nil
}

var ErrMissingLayout = errors.New("layouts: missing layout")

// Files loads layouts by individual file names. Each layout's name comes from the file name without its extension.
func Files(files ...string) error {
	mu.Lock()
	defer mu.Unlock()
	for _, f := range files {
		err := load(f)
		if err != nil {
			return err
		}
	}
	return nil
}

// Glob loads layouts through pattern matching. Each layout's name comes from the file name without its extension.
func Glob(patterns ...string) error {
	files := make([]string, 0, 8)
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, m := range matches {
			if filepath.Base(m)[0] == '.' {
				continue
			}
			files = append(files, m)
		}
	}
	return Files(files...)
}

// Clear unloads all layouts.
func Clear() {
	mu.Lock()
	defer mu.Unlock()

	dict = make(map[string]Entry, 8)
	tmpls = nil
}

// Entries returns info about all loaded layouts.
func Entries() []Entry {
	list := make([]Entry, 0, len(dict))
	for _, e := range dict {
		list = append(list, e)
	}
	return list
}

func execute(w io.Writer, layout string, t *template.Template, data interface{}) error {
	l, ok := dict[layout]
	if !ok {
		return ErrMissingLayout
	}

	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	err := t.Execute(buf, data)
	if err != nil {
		return err
	}

	b, err := ioutil.ReadAll(buf)
	if err != nil {
		return err
	}

	l.Funcs(template.FuncMap{
		"content": makeContentFunc(template.HTML(bytes.TrimSpace(b))),
	})

	if l.Parent != "" {
		return execute(w, l.Parent, l.Template, data)
	} else {
		err = l.Execute(w, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func executeHTML(w io.Writer, layout string, content template.HTML, data interface{}) error {
	l, ok := dict[layout]
	if !ok {
		return ErrMissingLayout
	}

	l.Funcs(template.FuncMap{
		"content": makeContentFunc(content),
	})
	if l.Parent != "" {
		return execute(w, l.Parent, l.Template, data)
	} else {
		err := l.Execute(w, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute renders the specified template using the named layout, passing in data to the layout templates.
func Execute(w io.Writer, layout string, t *template.Template, data interface{}) error {
	mu.Lock()
	defer mu.Unlock()

	return execute(w, layout, t, data)
}

// ExecuteHTML renders the content string using the named layout, passing in data to the layout templates. Note that the content string is of type template.HTML; it is expected that the content string is safe, fully-escaped HTML.
func ExecuteHTML(w io.Writer, layout string, content template.HTML, data interface{}) error {
	mu.Lock()
	defer mu.Unlock()

	return executeHTML(w, layout, content, data)
}
