package render

import (
	"encoding/json"
	"io"
	"net/http"
)

var jsonRender = JSON{}

type (
	JSON struct{}

	IndentedJSON struct{}
)

var jsonContentType = []string{"application/json; charset=utf-8"}

func (j JSON) Marshal(w http.ResponseWriter, v interface{}) error {
	writeContentType(w, jsonContentType)
	return json.NewEncoder(w).Encode(v)
}

func (j JSON) Unmarshal(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func (j IndentedJSON) Marshal(w http.ResponseWriter, v interface{}) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func (j IndentedJSON) Unmarshal(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// UnmarshalJSON JSON反序列化
func UnmarshalJSON(r io.Reader, v interface{}) error {
	return jsonRender.Unmarshal(r, v)
}
