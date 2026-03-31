package utils

import (
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/common"

	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure"
)

// NewHTTPServer 一键创建HTTP服务器
func NewHTTPServer(svr common.ServerInfo, opf map[string][]map[string]func(c *gin.Context)) error {
	r := gin.New()
	r.Use(gin.Recovery())
	for op, v := range opf {
		switch op {
		case "GET":
			for _, pf := range v {
				for p, f := range pf {
					r.GET(p, f)
				}
			}
		case "POST":
			for _, pf := range v {
				for p, f := range pf {
					r.POST(p, f)
				}
			}
		case "PUT":
			for _, pf := range v {
				for p, f := range pf {
					r.PUT(p, f)
				}
			}
		case "DELETE":
			for _, pf := range v {
				for p, f := range pf {
					r.DELETE(p, f)
				}
			}
		}
	}

	if svr.SSLOn {
		r.Use(func() gin.HandlerFunc {
			return func(c *gin.Context) {
				secreMiddleware := secure.New(secure.Options{
					SSLRedirect: true,
					SSLHost:     fmt.Sprintf(":%v", svr.Port),
				})
				err := secreMiddleware.Process(c.Writer, c.Request)
				if nil != err {
					return
				}
				c.Next()
			}
		}())
		return r.RunTLS(fmt.Sprintf(":%v", svr.Port), svr.CertFile, svr.KeyFile)
	}
	return r.Run(fmt.Sprintf(":%v", svr.Port))
}
