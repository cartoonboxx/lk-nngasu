package validate

import "testing"

func TestHTTPURL(t *testing.T) {
	good := []string{"https://example.com/a?b=1", "http://localhost:8080/x"}
	bad := []string{"", "javascript:alert(1)", "ftp://example.com", "//example.com", "example.com", "data:text/html,x"}
	for _, s := range good {
		if !HTTPURL(s) {
			t.Errorf("%q must be valid", s)
		}
	}
	for _, s := range bad {
		if HTTPURL(s) {
			t.Errorf("%q must be rejected", s)
		}
	}
}
