package handler

import (
	"DelayedNotifier/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
)

type NotificationHandler interface {
	CreateNotification(ctx *ginext.Context)
	GetNotificationStatus(ctx *ginext.Context)
	CancelNotification(ctx *ginext.Context)
}
type notificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler() NotificationHandler {
	return notificationHandler{}
}

type CreateNotificationRequest struct {
	Message string `json:"message" binding:"required"`
	SendAt  string `json:"send_at" binding:"required"`
	Channel string `json:"channel" binding:"required"`
}

// CreateNotification – POST /notify — создание уведомлений с датой и временем отправки
func (h notificationHandler) CreateNotification(ctx *ginext.Context) {
	var req CreateNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notificationId, err := h.service.CreateNotification(
		ctx.Request.Context(),
		req.Message,
		req.SendAt,
		req.Channel,
	)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": notificationId})
}

// GetNotificationStatus – GET /notify/{id} — получение статуса уведомления
func (h notificationHandler) GetNotificationStatus(ctx *ginext.Context) {

}

// CancelNotification – DELETE /notify/{id} — отмена запланированного уведомления
func (h notificationHandler) CancelNotification(ctx *ginext.Context) {

}
