// Package ming is an API gateway implementaion
package ming

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)
import "bytes"

import "errors"

// 运行模式定义
const (
	ReleaseMode = "release"
	DebugMode   = "debug"
)

// 系统异常定义
var (
	ErrServerError = errors.New("Server Error")
)

// 全局变量定义
var (
	DefaultErrorWriter io.Writer = os.Stderr
	DefaultOutWriter   io.Writer = os.Stdout
)

// M 字典对象
type M map[string]interface{}

type (
	// Config 服务配置
	Config struct {
		EnableFastCgi bool
	}
	// IServer 服务器接口
	IServer interface {
		// Run 启动server，开始监听端口
		Run(addr string)
		// Group 归组服务
		Group(ver, service, module string, handlers ...HandlerFunc) *Route
		// Use
		Use(handlers ...HandlerFunc)
		// Config 配置服务
		Config(c Config)
	}

	// APIHeader API请求头部
	APIHeader struct {
		Version    string `json:",omitempty"` // 接口版本
		Service    string `json:",omitempty"` // API服务
		Module     string `json:",omitempty"` // API服务下的模块
		Controller string `json:",omitempty"` // 模块下的控制器
		Action     string `json:",omitempty"` // 控制器下的方法
		Timestamp  int64  `json:",omitempty"` // 时间戳
		Noncestr   string `json:",omitempty"` // 随机字符串
		Token      string `json:",omitempty"` // 令牌环
		SignType   string `json:",omitempty"` // 签名方式
		Sign       string `json:",omitempty"` // 签名
		ClientIP   string `json:",omitempty"` // 客户端IP地址
		Channel    string `json:",omitempty"` // 平台渠道
	}
	// APIRequest API 请求参数
	APIRequest struct {
		Header APIHeader // 请求头部
		Data   []byte    // 请求数据
	}
	// APIResponse API 返回数据
	APIResponse struct {
		ErrCode int
		ErrMsg  string
		Result  interface{}
	}
)

// New 创建指定协议的服务器监听程序 可指定protocol 为tcp ，http，https
func New(protocol string) IServer {
	switch protocol {
	case "http":
		return newHTTPServer()
	default:
		panic(fmt.Sprintf("Not support protocol %s", protocol))
	}
}

// Default 创建默认服务器
func Default(protocol string) IServer {
	s := New(protocol)
	s.Use(Recovery())
	return s
}

var runMode = DebugMode

// SetMode 设置程序运行模式
func SetMode(mode string) {
	runMode = mode
}

func (h APIHeader) String() string {
	ver := "v1"
	if h.Version != "" {
		ver = h.Version
	}
	return fmt.Sprintf("%s_%s_%s_%s_%s", ver, h.Service, h.Module, h.Controller, h.Action)
}

func (req APIRequest) dump() []byte {
	var buf bytes.Buffer
	buf.WriteString("Request Header:\r\n")
	b, _ := json.Marshal(req.Header)
	buf.Write(b)
	buf.WriteString("\r\nRequest Body:\r\n")
	buf.Write(req.Data)
	return buf.Bytes()
}

// IsFromApp 请求是否来自移动终端
func (req APIRequest) IsFromApp() bool {
	switch strings.ToLower(req.Header.Channel) {
	case "ios", "android":
		return true
	default:
		return false
	}
}

// IsOK 是否返回成功结果
func (res APIResponse) IsOK() bool {
	return res.ErrCode == 0
}

// dump dump resnponse
func (res APIResponse) dump() []byte {
	var buf bytes.Buffer
	buf.WriteString("\r\nResponse Body:\r\n")
	b, _ := json.Marshal(res)
	buf.Write(b)
	return buf.Bytes()
}
