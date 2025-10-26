package httpserver

import (
	"context"

	"github.com/gin-gonic/gin"
)

// Return the gin engine
func (hs *HttpServer) GetRouter() *gin.Engine {

	return hs.ge
}

// Register a GET route with given callback function
func (hs *HttpServer) RegisterGetRoute(path string, callback gin.HandlerFunc, middlewares ...gin.HandlerFunc) {

	handlers := append(middlewares, callback)
	hs.ge.GET(path, handlers...)
}

// Register a POST route with given callback function
func (hs *HttpServer) RegisterPostRoute(path string, callback gin.HandlerFunc, middlewares ...gin.HandlerFunc) {

	handlers := append(middlewares, callback)
	hs.ge.POST(path, handlers...)
}

// Stop gin server
func (hs *HttpServer) Shutdown(ctx context.Context) error {

	return hs.server.Shutdown(ctx)
}

// Start gin server
func (hs *HttpServer) Serve() error {

	if hs.certLocation != "" && hs.keyLocation != "" {

		return hs.server.ListenAndServeTLS(hs.certLocation, hs.keyLocation)
	}

	return hs.server.ListenAndServe()
}
