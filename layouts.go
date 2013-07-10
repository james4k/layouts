package layouts

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"j4k.co/fmatter"
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

// Group is a collection of template layouts.
type Group struct {
	dir   string
	tmpls *template.Template
	funcs template.FuncMap
	dict  map[string]Entry
	mu    sync.Mutex
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

// New returns a group with layouts relative to dir
func New(dir string) *Group {
	g := &Group{
		dir: dir,
	}
	g.Clear()
	return g
}

// SetPath sets a new path for layouts to be loaded from. This should be called
// before any layouts are loaded.
/*
func (g *Group) SetPath(dir string) {
	g.dir = dir
}
*/

func (g *Group) load(filename string) error {
	front := make(map[string]interface{}, 4)
	content, err := fmatter.ReadFile(filename, front)
	if err != nil {
		return err
	}

	path := filename
	filename = filepath.Base(filename)
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	_, ok := g.dict[name]
	if ok {
		return nil
	}

	var t *template.Template
	if g.tmpls == nil {
		t = template.New(name)
		g.tmpls = t
		t.Funcs(template.FuncMap{
			"content": undefinedContent,
		})
		t.Funcs(g.funcs)
	} else {
		t = g.tmpls.New(name)
	}

	_, err = t.Parse(string(content))
	if err != nil {
		// hrm.. how do we remove a template..?
		return err
	}

	parent, _ := front["layout"].(string)
	g.dict[name] = Entry{unexportedEntry{t}, name, path, parent}
	return nil
}

// Files loads layouts by individual file names. Each layout's name
// comes from the file name without its extension.
func (g *Group) Files(files ...string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, f := range files {
		f = filepath.Join(g.dir, f)
		err := g.load(f)
		if err != nil {
			return err
		}
	}
	return nil
}

// Glob loads layouts through pattern matching. Each layout's name
// comes from the file name without its extension.
func (g *Group) Glob(patterns ...string) error {
	files := make([]string, 0, 8)
	for _, pattern := range patterns {
		pattern = filepath.Join(g.dir, pattern)
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		for _, m := range matches {
			m, err = filepath.Rel(g.dir, m)
			if err != nil {
				return err
			}
			if filepath.Base(m)[0] == '.' {
				continue
			}
			files = append(files, m)
		}
	}
	return g.Files(files...)
}

// Clear unloads all layouts.
func (g *Group) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.dict = make(map[string]Entry)
	g.tmpls = nil
}

// Entries returns info about all loaded layouts.
func (g *Group) Entries() []Entry {
	list := make([]Entry, 0, len(g.dict))
	for _, e := range g.dict {
		list = append(list, e)
	}
	return list
}

func (g *Group) execute(w io.Writer, layout string, t *template.Template, data interface{}) error {
	l, ok := g.dict[layout]
	if !ok {
		return fmt.Errorf(`layouts: missing layout "%s"`, layout)
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
		return g.execute(w, l.Parent, l.Template, data)
	} else {
		err = l.Execute(w, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *Group) executeHTML(w io.Writer, layout string, content template.HTML, data interface{}) error {
	l, ok := g.dict[layout]
	if !ok {
		return fmt.Errorf(`layouts: missing layout "%s"`, layout)
	}

	l.Funcs(template.FuncMap{
		"content": makeContentFunc(content),
	})
	if l.Parent != "" {
		return g.execute(w, l.Parent, l.Template, data)
	} else {
		err := l.Execute(w, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute renders the specified template using the named layout,
// passing in data to the layout templates.
func (g *Group) Execute(w io.Writer, layout string, t *template.Template, data interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.execute(w, layout, t, data)
}

// ExecuteHTML renders the content string using the named layout,
// passing in data to the layout templates. Note that the content
// string is of type template.HTML; it is expected that the content
// string is safe, fully-escaped HTML.
func (g *Group) ExecuteHTML(w io.Writer, layout string, content template.HTML, data interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.executeHTML(w, layout, content, data)
}

// Funcs adds funcs to all templates that are executed. See
// template.Funcs in html/template
func (g *Group) Funcs(f template.FuncMap) {
	if g.tmpls != nil {
		for _, t := range g.tmpls.Templates() {
			t.Funcs(f)
		}
	}
	if g.funcs == nil {
		g.funcs = template.FuncMap{}
	}
	for k, v := range f {
		g.funcs[k] = v
	}
}
