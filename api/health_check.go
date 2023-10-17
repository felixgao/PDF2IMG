package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterHealthCheckHandlers(handler *gin.Engine) {
	handler.GET("/healthcheck", HealthCheck)
	handler.GET("/health", HealthCheck)
	handler.GET("/ping", HealthCheck)
	handler.GET("/api/healthcheck", HealthCheck)
	handler.GET("/api/health", HealthCheck)
	handler.GET("/api/ping", HealthCheck)
}

// @Summary Healthcheck for the service
// @Tags Healthcheck
// @Produce application/json
// @Success 200 {object} object{message=string} "Resposta de Successo"
// @Router /healthcheck [get]
func HealthCheck(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}
