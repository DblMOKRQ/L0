package router

import (
	_ "L0/docs"
	"L0/internal/router/handlers"
	"L0/internal/router/middleware"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
)

type Router struct {
	rout    *gin.Engine
	handler *handlers.OrderHandlers
	log     *zap.Logger
}

func NewRouter(handler *handlers.OrderHandlers, mode string, log *zap.Logger) *Router {
	switch mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}
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
	r.rout.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.rout.GET("/order/:orderUID", r.handler.GetOrder)
	r.rout.LoadHTMLGlob("static/*")
	r.rout.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
}
func (r *Router) GetHTTPHandler() *gin.Engine {
	return r.rout
}
