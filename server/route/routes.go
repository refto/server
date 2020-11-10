package route

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/refto/server/config"
	"github.com/sirupsen/logrus"
)

func Register(r *gin.Engine) {
	conf := config.Get()

	r.Use(corsConfig())
	r.Use(openAPIHandler)

	// serve static
	r.Use(static.Serve(conf.Server.Static.WebPath, static.LocalFile(conf.Server.Static.LocalPath, false)))

	api := r.Group(conf.Server.ApiBasePath)
	api.Use()

	apply(api,
		entityRoutes,
		topicRoutes,
		webHookRoutes,
	)

	r.NoRoute(func(c *gin.Context) {
		data, err := ioutil.ReadFile(filepath.Join(conf.Server.Static.LocalPath, static.INDEX))
		if err != nil {
			logrus.Error(err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.Data(http.StatusOK, "text/html", data)
	})
}

func apply(rg *gin.RouterGroup, routeFn ...func(*gin.RouterGroup)) {
	for _, fn := range routeFn {
		fn(rg)
	}
}

func corsConfig() gin.HandlerFunc {
	conf := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}

	return cors.New(conf)
}

// openAPIHandler ...
// Checks if uri matches an openAPI file and replaces server addr placeholder with actual addr
// TODO move this mw from here. To separate file or to own package
func openAPIHandler(c *gin.Context) {
	const (
		openAPIURI           = "/openapi/openapi.yaml"
		serverURLPlaceholder = "%SERVER_URL%"
	)

	conf := config.Get()

	if c.Request.RequestURI != path.Join(conf.Server.Static.WebPath, openAPIURI) {
		c.Next()
		return
	}

	data, err := ioutil.ReadFile(filepath.Join(conf.Server.Static.LocalPath, openAPIURI))
	if err != nil {
		logrus.Error(err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	host := conf.Server.Host
	scheme := "http"
	if conf.Server.Port == "443" {
		scheme += "s"
	}
	scheme += "://"
	port := conf.Server.Port
	if port != "" {
		port = ":" + port
	}
	if host == "" {
		scheme = ""
	}

	apiURI := scheme + host + port + path.Join("/", conf.Server.ApiBasePath)
	data = bytes.Replace(data, []byte(serverURLPlaceholder), []byte(apiURI), -1)

	c.Data(http.StatusOK, "text/yaml", data)
	c.Abort()
}
