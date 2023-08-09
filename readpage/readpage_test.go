package readpage

import (
	"embed"
	"golang.org/x/net/html"
	"testing"
)

//go:embed testdata/*
var testdata embed.FS

func TestTitles(t *testing.T) {
	table := map[string]string{
		"headless-title":       "A title!",
		"no-title-no-heading":  "",
		"no-title-yes-heading": "",
	}
	for name, expectedTitle := range table {
		t.Log("case", name)
		file, _ := testdata.Open("testdata/" + name + ".html")
		doc, _ := html.Parse(file)
		data, err := findData("https://bouncepaw.com", titleWorkers, doc)
		if data.title != expectedTitle {
			t.Errorf("In ‘%s’, expected title ‘%s’, got ‘%s’. Error value is ‘%v’.",
				name, expectedTitle, data.title, err)
		}
	}
}
