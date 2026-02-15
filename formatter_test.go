package shedoc

import (
	"bytes"
	"io"
	"sort"
	"testing"
)

// stubFormatter is a trivial formatter for registry tests.
type stubFormatter struct{}

func (f *stubFormatter) Format(w io.Writer, doc *Document) error {
	_, err := w.Write([]byte("stub"))
	return err
}

func TestFormatterRegistry(t *testing.T) {
	// Save and restore the global registry.
	saved := formatters
	formatters = map[string]Formatter{}
	defer func() { formatters = saved }()

	// GetFormatter returns nil for unknown name.
	if got := GetFormatter("nope"); got != nil {
		t.Errorf("GetFormatter(unknown) = %v, want nil", got)
	}

	// RegisterFormatter adds entries.
	RegisterFormatter("a", &stubFormatter{})
	RegisterFormatter("b", &stubFormatter{})

	if got := GetFormatter("a"); got == nil {
		t.Error("GetFormatter(a) = nil after register")
	}

	names := RegisteredFormats()
	sort.Strings(names)
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("RegisteredFormats() = %v, want [a b]", names)
	}

	// Verify the stub formatter works.
	var buf bytes.Buffer
	if err := GetFormatter("a").Format(&buf, &Document{}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "stub" {
		t.Errorf("stub output = %q", buf.String())
	}
}
