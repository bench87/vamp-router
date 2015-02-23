package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/magneticio/vamp-loadbalancer/haproxy"
	"github.com/magneticio/vamp-loadbalancer/metrics"
	gologger "github.com/op/go-logging"
)

func CreateApi(port int, haConfig *haproxy.Config, haRuntime *haproxy.Runtime, log *gologger.Logger, SSEBroker *metrics.SSEBroker) *gin.Engine {

	gin.SetMode("release")

	r := gin.New()
	r.Use(HaproxyMiddleware(haConfig, haRuntime))
	r.Use(LoggerMiddleware(log))
	r.Use(gin.Recovery())
	r.Static("/www", "./www")
	v1 := r.Group("/v1")

	{
		r.GET("/", func(c *gin.Context) {
			c.Redirect(301, "www/index.html")
		})

		/*
		   Frontend
		*/
		v1.GET("/frontends", GetFrontends)
		v1.POST("/frontends", PostFrontend)
		v1.POST("frontends/:name/filters", PostFrontendFilter)
		v1.GET("/frontends/:name/filters", GetFrontendFilters)
		v1.DELETE("/frontends/:name/filters/:filter_name", DeleteFrontendFilter)
		v1.GET("/frontends/:name", GetFrontend)
		v1.DELETE("/frontends/:name", DeleteFrontend)

		/*
		   Backend
		*/
		v1.GET("/backends", GetBackends)
		v1.POST("/backends", PostBackend)
		v1.GET("/backends/:name", GetBackend)
		v1.DELETE("/backends/:name", DeleteBackend)
		v1.GET("/backends/:name/servers", GetServers)
		v1.GET("/backends/:name/servers/:server", GetServer)
		v1.PUT("/backends/:name/servers/:server", PutServerWeight)
		v1.POST("/backends/:name/servers", PostServer)
		v1.DELETE("/backends/:name/servers/:server", DeleteServer)

		/*
		   Stats
		*/
		v1.GET("/stats", GetAllStats)
		v1.GET("/stats/backends", GetBackendStats)
		v1.GET("/stats/frontends", GetFrontendStats)
		v1.GET("/stats/servers", GetServerStats)
		v1.GET("/stats/stream", SSEMiddleware(SSEBroker), GetSSEStream)

		/*
		   Config
		*/
		v1.GET("/config", GetConfig)
		v1.POST("/config", PostConfig)

		/*
			Routes
		*/
		v1.GET("/routes", GetRoutes)
		v1.POST("/routes", PostRoute)

		v1.GET("/routes/:route", GetRoute)
		v1.PUT("/routes/:route", PutRoute)
		v1.DELETE("/routes/:route", DeleteRoute)

		v1.GET("/routes/:route/groups", GetRouteGroups)
		v1.POST("/routes/:route/groups", PostRouteGroup)
		v1.GET("/routes/:route/groups/:group", GetRouteGroup)
		v1.PUT("/routes/:route/groups/:group", PutRouteGroup)
		v1.DELETE("/routes/:route/groups/:group", DeleteRouteGroup)

		v1.GET("/routes/:route/groups/:group/servers", GetGroupServers)
		v1.GET("/routes/:route/groups/:group/servers/:server", GetGroupServer)
		v1.PUT("/routes/:route/groups/:group/servers/:server", PutGroupServer)
		v1.POST("/routes/:route/groups/:group/servers", PostGroupServer)
		v1.DELETE("/routes/:route/groups/:group/servers/:server", DeleteGroupServer)
		/*
		   Info
		*/
		v1.GET("/info", GetInfo)
	}

	return r
}

// Handles the reloading and persisting of the Haproxy config after a successful mutation of the
// config object.
func HandleReload(c *gin.Context, config *haproxy.Config, status int, message string) {

	runtime := c.MustGet("haRuntime").(*haproxy.Runtime)

	err := config.RenderAndPersist()
	if err != nil {
		HandleError(c, &haproxy.Error{500, errors.New("Error rendering config file")})
		return
	}

	err = runtime.Reload(config)
	if err != nil {
		HandleError(c, &haproxy.Error{500, errors.New("Error reloading the HAproxy configuration")})
		return
	}

	HandleSucces(c, status, message)
}

// Handles the simple successful return status
func HandleSucces(c *gin.Context, status int, message string) {
	c.String(status, message)
}

// Handles the return of an error from the Haproxy object
func HandleError(c *gin.Context, err *haproxy.Error) {
	c.String(err.Code, err.Error())
}

// helper methods to grab the injected Config from the Http context
func Config(c *gin.Context) *haproxy.Config {
	return c.MustGet("haConfig").(*haproxy.Config)
}

// helper methods to grab the injected Runtime from the Http context
func Runtime(c *gin.Context) *haproxy.Runtime {
	return c.MustGet("haRuntime").(*haproxy.Runtime)
}
