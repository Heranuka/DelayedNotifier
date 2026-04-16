// Package transport provides HTTP routing and middleware setup for the application.
package transport

import (
	"delay/internal/transport/http/handler/notifications"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
)

// Handlers groups all HTTP handlers used by the router.
type Handlers struct {
	Notifications *notifications.Handler
}

// NewRouter creates and configures the HTTP router with middleware and routes.
func NewRouter(handlers Handlers) http.Handler {
	r := ginext.New("")

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:5173",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	config.AllowCredentials = true

	r.POST("/create", handlers.Notifications.Create)
	r.GET("/all", handlers.Notifications.GetAll)
	r.GET("/status/:id", handlers.Notifications.GetStatus)
	r.DELETE("/cancel/:id", handlers.Notifications.Cancel)

	return r
}
