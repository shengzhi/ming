// Pakcage render 数据展示模块
package render

import "net/http"
import "io"

var renderMap = map[string]Render{
	"application/json":                jsonRender,
	"application/json; charset=utf-8": jsonRender,
	"text/xml":                        xmlRender,
}

// Render 数据展示接口
type Render interface {
	Marshal(http.ResponseWriter, interface{}) error
	Unmarshal(io.Reader, interface{}) error
}

// Unmarshal 反序列化请求数据
func Unmarshal(r *http.Request, v interface{}) error {
	ren, has := renderMap[r.Header.Get("Content-Type")]
	if !has {
		ren = jsonRender
	}
	return ren.Unmarshal(r.Body, v)
}

// Marshal 序列化响应结果
func Marshal(contentType string, w http.ResponseWriter, v interface{}) error {
	ren, has := renderMap[contentType]
	if !has {
		ren = jsonRender
	}
	return ren.Marshal(w, v)
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
