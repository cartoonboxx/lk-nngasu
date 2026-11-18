package resources

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefault(t *testing.T) {
	r, err := Load("")
	if err != nil || len(r.Services) == 0 || r.Contacts == nil {
		t.Fatalf("default: %+v, %v", r, err)
	}
}

func TestLoadRejects(t *testing.T) {
	cases := map[string]string{
		"unknown field": `{"services":[],"contacts":[],"extra":1}`,
		"relative url":  `{"services":[{"id":"a","label":"A","description":"","url":"/x","link_text":"go","icon_key":"x"}]}`,
		"duplicate id":  `{"services":[{"id":"a","label":"A","description":"","url":"https://a.ru","link_text":"go","icon_key":"x"},{"id":"a","label":"B","description":"","url":"https://b.ru","link_text":"go","icon_key":"x"}]}`,
	}
	dir := t.TempDir()
	for name, body := range cases {
		p := filepath.Join(dir, "r.json")
		if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(p); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
}
