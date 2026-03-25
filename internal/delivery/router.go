package delivery

import (
	"DelayedNotifier/internal/delivery/handler"
	"DelayedNotifier/internal/delivery/middleware"

	"github.com/wb-go/wbf/ginext"
)

func NewRouter(h handler.NotificationHandler, enableUI bool) *ginext.Engine {
	router := ginext.New("")

	router.Use(middleware.Logger())
	router.Use(ginext.Recovery())

	if enableUI {
		router.LoadHTMLGlob("templates/*")
		router.GET("/", h.Index)
	}

	router.GET("/notify/:id", h.GetNotificationStatus)
	router.POST("/notify", h.CreateNotification)
	router.DELETE("/notify/:id", h.CancelNotification)

	return router
}
