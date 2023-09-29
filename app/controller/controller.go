package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Controller struct {
	router            *gin.Engine
	prometheusHandler http.Handler
}

type Wiring struct {
	Router            *gin.Engine
	PrometheusHandler http.Handler
}

func NewController(wiring *Wiring) {
	controller := &Controller{
		router:            wiring.Router,
		prometheusHandler: wiring.PrometheusHandler,
	}
	controller.router.Use(gin.Logger(), gin.Recovery())

	controller.router.GET("/metrics", func(c *gin.Context) {
		controller.prometheusHandler.ServeHTTP(c.Writer, c.Request)
	})
}
