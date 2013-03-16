package layouts

import (
	"bytes"
	"github.com/james4k/fmatter/toml"
	"html/template"
	"io/ioutil"
	"testing"
)

func TestTemplate(t *testing.T) {
	Clear()
	err := Glob("testdata/layouts/*.html")
	if err != nil {
		t.Fatal(err)
	}

	matter := make(map[string]interface{})
	content, err := fmatter.ReadFile("testdata/index.html", matter)
	if err != nil {
		t.Fatal(err)
	}

	tmpl, err := template.New("index").Parse(string(content))
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 256))
	data := map[string]interface{}{
		"page": matter,
	}
	err = Execute(buf, matter["layout"].(string), tmpl, data)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := ioutil.ReadFile("testdata/indexresult.html")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := ioutil.ReadAll(buf)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(expected, actual) != 0 {
		t.Fatal("result does not match!")
	}
}

func TestUserHTML(t *testing.T) {
	Clear()
	err := Glob("testdata/layouts/*.html")
	if err != nil {
		t.Fatal(err)
	}

	matter := make(map[string]interface{})
	content, err := fmatter.ReadFile("testdata/index.html", matter)
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0, 256))
	data := map[string]interface{}{
		"page": matter,
	}
	err = ExecuteHTML(buf, matter["layout"].(string), template.HTML(content), data)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := ioutil.ReadFile("testdata/indexresult.html")
	if err != nil {
		t.Fatal(err)
	}
	actual, err := ioutil.ReadAll(buf)
	if err != nil {
		t.Fatal(err)
	}

	if bytes.Compare(expected, actual) != 0 {
		t.Fatal("result does not match!")
	}
}

func TestMakeContentFunc(t *testing.T) {
	fn := makeContentFunc(template.HTML("blerg")).(func() interface{})
	str := fn()
	if str.(template.HTML) != template.HTML("blerg") {
		t.Fatal("unexpected content string")
	}

	paniced := false
	func() {
		defer func() {
			if x := recover(); x == "content undefined" {
				paniced = true
			}
		}()

		str = fn()
	}()

	if !paniced {
		t.Fatal("did not get expected panic")
	}
}