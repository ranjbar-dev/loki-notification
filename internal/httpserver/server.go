package httpserver

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HttpServer represent all data needed for running gin
type HttpServer struct {
	host         string
	port         string
	certLocation string
	keyLocation  string
	server       *http.Server
	ge           *gin.Engine
}

// Return a HttpServer instance with given data
func NewHttpServer(host string, port string, certLocation string, keyLocation string) *HttpServer {

	gin.DefaultWriter = io.Discard
	ge := gin.New()

	ge.Use(gin.Recovery())

	ge.SetTrustedProxies(nil)

	return &HttpServer{
		host:         host,
		port:         port,
		certLocation: certLocation,
		keyLocation:  keyLocation,
		server: &http.Server{
			Addr:    host + ":" + port,
			Handler: ge,
		},
		ge: ge,
	}
}
