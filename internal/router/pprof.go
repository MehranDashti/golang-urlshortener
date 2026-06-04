//go:build pprof

package router

import (
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/gin-gonic/gin"
)

// registerDebugRoutes adds pprof + goroutine endpoints.
// Only compiled when built with: go build -tags pprof
// This is intentional: debug endpoints must not exist in production binaries.
func registerDebugRoutes(api *gin.RouterGroup) {
	debug := api.Group("/debug")
	{
		debug.GET("/goroutines", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"goroutines": runtime.NumGoroutine(),
			})
		})
		debug.GET("/pprof/", gin.WrapF(pprof.Index))
		debug.GET("/pprof/cmdline", gin.WrapF(pprof.Cmdline))
		debug.GET("/pprof/profile", gin.WrapF(pprof.Profile))
		debug.GET("/pprof/symbol", gin.WrapF(pprof.Symbol))
		debug.GET("/pprof/trace", gin.WrapF(pprof.Trace))
		debug.GET("/pprof/heap", gin.WrapH(pprof.Handler("heap")))
		debug.GET("/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debug.GET("/pprof/block", gin.WrapH(pprof.Handler("block")))
		debug.GET("/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	}
}
