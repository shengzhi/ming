package ming

import (
	"io"
	"log"
	"time"

	"github.com/fatih/color"
)

// Accesslog 访问日志中间件
func Accesslog() HandlerFunc {
	return AccesslogWithWriter(DefaultOutWriter)
}

func AccesslogWithWriter(w io.Writer) HandlerFunc {
	var logger *log.Logger
	if w != nil {
		logger = log.New(w, "", log.LstdFlags)
	}
	return func(c *Context) {
		from := time.Now()
		c.Next()
		d := time.Now().Sub(from)
		var clr func(string, ...interface{}) string
		if c.Response.IsOK() {
			clr = color.New(color.BgGreen).SprintfFunc()
		} else {
			clr = color.New(color.BgRed).SprintfFunc()
		}
		if runMode == DebugMode {
			color.Green("%v", string(c.Request.dump()))
			color.Yellow("%v", string(c.Response.dump()))
		}
		logger.Printf("%s\t%s\t%s\t%s\n", d.String(), clr("%3d", c.Response.ErrCode), c.Request.Header.ClientIP, c.Request.Header.String())

	}
}
