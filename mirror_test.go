package mirror

import (
	"io/ioutil"
	"testing"
)

func TestGetLinksFromHTML(t *testing.T) {
	client := NewClient(param{})
	filePath := "./testdata/test.html"
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	links, err := client.getLinksFromHTML(string(bytes), "src")
	if err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		path string
	}{
		{path: "/apple.jpg"},
		{path: "/jquery.js"},
	}
	for i, link := range links {
		if link != expected[i].path {
			t.Errorf("path do not match expected:%s real:%s\n", expected[i].path, link)
		}
	}

	links, err = client.getLinksFromHTML(string(bytes), "href")
	if err != nil {
		t.Fatal(err)
	}
	expected = []struct {
		path string
	}{
		{path: "/favicon.ico"},
	}
	for i, link := range links {
		if link != expected[i].path {
			t.Errorf("path do not match expected:%s real:%s\n", expected[i].path, link)
		}
	}
}

func TestGetLinksFromCSS(t *testing.T) {
	client := NewClient(param{})

	filePath := "./testdata/style.css"
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	links, err := client.getLinksFromCSS(string(bytes))
	if err != nil {
		t.Fatal(err)
	}
	expected := []struct {
		path string
	}{
		{path: "/assets/common/parts/fonts/icon.ttf"},
		{path: "/assets/common/parts/fonts/icon.woff"},
		{path: "/assets/common/parts/fonts/icon.svg"},
	}
	for i, link := range links {
		if link != expected[i].path {
			t.Errorf("path do not match expected:%s real:%s\n", expected[i].path, link)
		}
	}
}
