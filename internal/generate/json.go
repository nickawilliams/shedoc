package generate

import (
	"encoding/json"
	"io"

	"github.com/nickawilliams/shedoc"
)

func init() {
	shedoc.RegisterFormatter("json", &JSONFormatter{})
}

// JSONFormatter outputs a Document as JSON.
type JSONFormatter struct{}

func (f *JSONFormatter) Format(w io.Writer, doc *shedoc.Document) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(doc)
}
