package gateway

import "github.com/gin-gonic/gin"

type HTTPApiExtension interface {
	ConfigureRoutes(router *gin.Engine, cfg GatewayRuntimeConfig)
}
