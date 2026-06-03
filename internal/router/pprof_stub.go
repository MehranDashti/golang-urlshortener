//go:build !pprof

package router

import "github.com/gin-gonic/gin"

// registerDebugRoutes is a no-op when built without pprof tag.
func registerDebugRoutes(api *gin.RouterGroup) {}