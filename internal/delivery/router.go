package delivery

import (
	"DelayedNotifier/internal/delivery/handler"

	"github.com/wb-go/wbf/config"
	"github.com/wb-go/wbf/ginext"
)

func NewRouter(cfg *config.Config) *ginext.Engine {
	router := ginext.New("")
	notificationHandler := handler.NewNotificationHandler()

	router.GET("/notify/:id", notificationHandler.GetNotificationStatus)
	router.POST("/notify", notificationHandler.CreateNotification)
	router.DELETE("/notify/:id", notificationHandler.CancelNotification)

	return router
}
