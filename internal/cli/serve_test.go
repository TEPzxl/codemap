package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestServeCommandRequiresPath(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"serve"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected serve without path to fail")
	}
	if !strings.Contains(stderr.String(), "usage: codemap serve") {
		t.Fatalf("expected usage in stderr, got %q", stderr.String())
	}
}

func TestServeCommandRejectsInvalidPort(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := Run([]string{"serve", ".", "--port", "-1"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected serve with invalid port to fail")
	}
	if !strings.Contains(stderr.String(), "port must be between") {
		t.Fatalf("expected port error in stderr, got %q", stderr.String())
	}
}
