//go:build pprof

package router

import (
    "net/http/pprof"
    "runtime"
    "net/http"

    "github.com/gin-gonic/gin"
)

// registerDebugRoutes adds pprof endpoints.
// Only compiled when built with -tags pprof.
func registerDebugRoutes(api *gin.RouterGroup) {
    debug := api.Group("/debug")
    {
        debug.GET("/goroutines", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "goroutines": runtime.NumGoroutine(),
            })
        })
        debug.GET("/pprof/", gin.WrapF(pprof.Index))
        debug.GET("/pprof/profile", gin.WrapF(pprof.Profile))
        debug.GET("/pprof/heap", gin.WrapH(pprof.Handler("heap")))
        debug.GET("/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
        debug.GET("/pprof/block", gin.WrapH(pprof.Handler("block")))
        debug.GET("/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
    }
}