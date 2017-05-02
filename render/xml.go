package render

import (
	"encoding/xml"
	"io"
	"net/http"
)

var xmlRender = XML{}

type XML struct{}

var xmlContentType = []string{"text/xml; charset=utf-8"}

func (x XML) Marshal(w http.ResponseWriter, v interface{}) error {
	writeContentType(w, xmlContentType)
	return xml.NewEncoder(w).Encode(v)
}
func (x XML) Unmarshal(r io.Reader, v interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}

// UnmarshalXML XML反序列化
func UnmarshalXML(r io.Reader, v interface{}) error {
	return xmlRender.Unmarshal(r, v)
}
