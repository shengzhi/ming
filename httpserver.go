// http server 实现

package ming

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"errors"
	"io"
	"os"

	"net/http/fcgi"

	"github.com/shengzhi/ming/render"
)

// 错误定义
var (
	ErrNoAction = errors.New("Not found action")
)

// 类型定义
type (
	// HTTPServer Http serve 实现
	HTTPServer struct {
		pool *sync.Pool
		*routeGroup
		globalChain handlerFuncChain
		conf        Config
	}
)

// newHTTPServer 创建HTTP Server实例
func newHTTPServer() *HTTPServer {
	return &HTTPServer{
		pool: &sync.Pool{
			New: func() interface{} {
				return &Context{}
			},
		},
		routeGroup: newRouteGroup(),
	}
}

// Config 配置服务
func (hs *HTTPServer) Config(conf Config) {
	hs.conf = conf
}

// Use 配置全局中间件
func (hs *HTTPServer) Use(handlers ...HandlerFunc) {
	hs.globalChain = append(hs.globalChain, handlers...)
}

// Group group
func (hs *HTTPServer) Group(ver, service, module string, handlers ...HandlerFunc) *Route {
	r := &Route{version: ver, service: service, module: module}
	r.chain = make(handlerFuncChain, 0, 20)
	r.chain = append(hs.globalChain, handlers...)
	hs.routeGroup.add(r)
	return r
}
func (hs *HTTPServer) runCgi(addr string) {
	stopchan := make(chan os.Signal)
	signal.Notify(stopchan, os.Interrupt)
	srv := http.Server{Addr: addr, Handler: hs}
	go func() {
		err := srv.ListenAndServe()
		log.Printf("Listen: %v\r\n", err)
	}()

	<-stopchan
	log.Println("Shutting down server......")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	srv.Shutdown(ctx)
	log.Println("Server gracefully stopped")
}

// Run 启动服务，开始监听端口
func (hs *HTTPServer) Run(addr string) {
	hs.parseActions()
	if addr == "" {
		addr = ":9090"
	}
	fmt.Printf("Listen Port: %s \t", addr)
	fmt.Println("All Actions:")
	fmt.Println(hs.actions())

	if hs.conf.EnableFastCgi {
		l, err := net.Listen("tcp", addr)
		if err != nil {
			log.Printf("Listen: %v\r\n", err)
		}
		if err = fcgi.Serve(l, hs); err != nil {
			panic(err)
		}
	} else {
		hs.runCgi(addr)
	}
}

// ServeHTTP 实现http请求处理
func (hs *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := hs.pool.Get().(*Context)
	defer hs.pool.Put(ctx)
	accept := req.Header.Get("Accept")
	ctx.reset()
	apireq, ren, err := hs.makeAPIRequest(req)
	if err != nil {
		ctx.Error(400, err)
		render.Marshal(accept, w, ctx.Response)
		return
	}
	ctx.Request, ctx.render_ = apireq, ren
	ctx.Request.Header.ClientIP = hs.clientIP(req)
	handlers, has := hs.getHandler(ctx.Request.Header)
	if has {
		ctx.chain = handlers
	} else {
		ctx.chain = hs.globalChain
		ctx.Error(404, ErrNoAction)
	}
	ctx.Next()
	render.Marshal(accept, w, ctx.Response)
}

// makeAPIRequest 构造API请求参数
func (hs *HTTPServer) makeAPIRequest(r *http.Request) (APIRequest, render.Render, error) {
	defer r.Body.Close()
	ct := r.Header.Get("Content-Type")
	switch ct {
	case "application/xml", "text/xml":
		return hs.makeAPIRequestXML(r.Body)
	case "application/json", "text/json", "application/json; charset=utf-8", "text/json; charset=utf-8":
		return hs.makeAPIRequestJSON(r.Body)
	default:
		return hs.makeAPIRequestForm(r)
	}
}

func (hs *HTTPServer) makeAPIRequestForm(r *http.Request) (APIRequest, render.Render, error) {
	// body, _ := ioutil.ReadAll(r.Body)
	// bodyStr, _ := url.QueryUnescape(string(body))
	// fmt.Println("POST body:", bodyStr)
	// val, _ := url.ParseQuery(bodyStr)
	r.ParseForm()
	val := r.Form
	apireq := APIRequest{
		Header: APIHeader{
			Version:    val.Get("Version"),
			Service:    val.Get("Service"),
			Module:     val.Get("Module"),
			Controller: val.Get("Controller"),
			Action:     val.Get("Action"),
			Token:      val.Get("Token"),
			Noncestr:   val.Get("Nonce"),
			Channel:    val.Get("Channel"),
		},
		Data: []byte(val.Get("Request")),
	}
	apireq.Header.Timestamp, _ = strconv.ParseInt(r.FormValue("Timestamp"), 10, 64)
	return apireq, render.JSON{}, nil
}

func (hs *HTTPServer) makeAPIRequestJSON(r io.Reader) (APIRequest, render.Render, error) {
	var req struct {
		APIHeader
		Body json.RawMessage `json:"request"`
	}
	err := render.UnmarshalJSON(r, &req)
	if err != nil {
		return APIRequest{}, render.JSON{}, err
	}
	return APIRequest{Header: req.APIHeader, Data: req.Body}, render.JSON{}, nil
}

func (hs *HTTPServer) makeAPIRequestXML(r io.Reader) (APIRequest, render.Render, error) {
	var req struct {
		XMLName xml.Name `xml:"APIRequest"`
		APIHeader
		Body struct {
			Request []byte `xml:",innerxml"`
		} `xml:"Request"`
	}
	err := render.UnmarshalXML(r, &req)
	if err != nil {
		return APIRequest{}, render.XML{}, err
	}
	var buf bytes.Buffer
	buf.WriteString("<Req>")
	buf.Write(req.Body.Request)
	buf.WriteString("</Req>")
	return APIRequest{Header: req.APIHeader, Data: buf.Bytes()}, render.XML{}, nil
}

func (hs *HTTPServer) clientIP(r *http.Request) string {
	clientIP := strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = r.Header.Get("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}
