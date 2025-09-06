package router

import (
	"L0/internal/router/handlers"
	"L0/internal/router/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	rout    *gin.Engine
	handler *handlers.OrderHandlers
	log     *zap.Logger
}

func NewRouter(handler *handlers.OrderHandlers, log *zap.Logger) *Router {
	router := &Router{
		rout:    gin.Default(),
		handler: handler,
		log:     log,
	}
	router.setupRouter()
	return router
}
func (r *Router) setupRouter() {
	r.rout.Use(middleware.LoggingMiddleware(r.log))
	r.rout.GET("/order/:orderUID", r.handler.GetOrder)
}
func (r *Router) GetHTTPHandler() *gin.Engine {
	return r.rout
}
