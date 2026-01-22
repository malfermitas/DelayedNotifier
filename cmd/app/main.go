package main

import (
	"DelayedNotifier/internal/delivery"
	"DelayedNotifier/internal/repository"
)
import "DelayedNotifier/internal/delivery/handler"

func main() {
	_ = repository.NewNotificationRepository(nil, nil)
	_ = handler.NewNotificationHandler()
	_ = delivery.NewRouter(nil)

}
