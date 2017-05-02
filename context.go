package ming

import (
	"math"

	"bytes"

	"github.com/shengzhi/ming/render"
)

const (
	abortIndex int8 = math.MaxInt8 / 2
)

// Error API错误
type Error struct {
	ErrCode int
	Err     error
}

// NewError 创建错误
func NewError(code int, err error) Error {
	return Error{ErrCode: code, Err: err}
}

func (e Error) Error() string {
	return e.Err.Error()
}

type (
	// HandlerFunc 请求处理器
	HandlerFunc func(c *Context)

	handlerFuncChain []HandlerFunc

	// Context 请求上下文
	Context struct {
		render_  render.Render
		Request  APIRequest
		Response APIResponse
		Err      error
		chain    handlerFuncChain
		User     interface{}
		index    int8
	}
)

func (c *Context) reset() {
	c.Response = APIResponse{}
	c.Err = nil
	c.index = -1
	c.chain = nil
	c.render_ = nil
	c.User = nil
}

// Next 执行下一个处理程序
func (c *Context) Next() {
	c.index++
	for ; c.index < int8(len(c.chain)); c.index++ {
		c.chain[c.index](c)
	}
}

// Abort 终止处理
func (c *Context) Abort() {
	c.index = abortIndex
}

// AbortWithError 终端处理并返回错误
func (c *Context) AbortWithError(err error) {
	c.index = abortIndex
	c.ServerError(err)
}

// IsAborted 是否已终止程序
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// Data 设置返回结果
func (c *Context) Data(v interface{}) {
	c.Response.Result = v
}

// Error 设置返回错误结果
func (c *Context) Error(code int, err error) {
	c.Response.ErrCode, c.Response.ErrMsg = code, err.Error()
}

// ServerError 返回500服务端错误
func (c *Context) ServerError(err error) {
	if e, ok := err.(Error); ok {
		c.Response.ErrCode, c.Response.ErrMsg = e.ErrCode, e.Error()
	} else {
		c.Response.ErrCode, c.Response.ErrMsg = 500, err.Error()
	}
}

// Bind 绑定数据
func (c *Context) Bind(v interface{}) error {
	return c.render_.Unmarshal(bytes.NewReader(c.Request.Data), v)
}
