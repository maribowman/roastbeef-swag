package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/maribowman/roastbeef-swag/app/config"
	"github.com/maribowman/roastbeef-swag/app/controller"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

func InitServer() (*http.Server, error) {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Config.Server.Port),
		Handler: injectRouter(),
	}, nil
}

func injectRouter() *gin.Engine {
	gin.SetMode(config.Config.Server.Mode)
	router := gin.New()
	controller.NewController(&controller.Wiring{
		Router:            router,
		PrometheusHandler: promhttp.Handler(),
	})
	return router
}