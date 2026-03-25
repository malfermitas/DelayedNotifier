package handler

import (
	"DelayedNotifier/internal/service"
	"DelayedNotifier/internal/shared"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
)

type NotificationHandler interface {
	CreateNotification(ctx *ginext.Context)
	GetNotificationStatus(ctx *ginext.Context)
	CancelNotification(ctx *ginext.Context)
	Index(ctx *ginext.Context)
}
type notificationHandler struct {
	service     shared.NotificationService
	botUsername string
}

func NewNotificationHandler(service shared.NotificationService, botUsername string) NotificationHandler {
	return notificationHandler{service: service, botUsername: botUsername}
}

type CreateNotificationRequest struct {
	Message   string `json:"message" binding:"required"`
	SendAt    string `json:"send_at" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Email     string `json:"email"`
	Recipient struct {
		Email  string `json:"email"`
		ChatID string `json:"chat_id"`
		UserID string `json:"user_id"`
	} `json:"recipient"`
}

// CreateNotification – POST /notify — создание уведомлений с датой и временем отправки
func (h notificationHandler) CreateNotification(ctx *ginext.Context) {
	var req CreateNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requestEmail := req.Recipient.Email
	if requestEmail == "" {
		requestEmail = req.Email
	}

	notificationId, err := h.service.CreateNotification(
		ctx.Request.Context(),
		req.Message,
		req.SendAt,
		req.Channel,
		requestEmail,
		req.Recipient.ChatID,
		req.Recipient.UserID,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidNotificationChannel) ||
			errors.Is(err, service.ErrEmailRecipientRequired) ||
			errors.Is(err, service.ErrTelegramRecipientRequired) ||
			errors.Is(err, service.ErrTelegramRecipientNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"id": notificationId, "message": notificationId, "status": "scheduled"})
}

// GetNotificationStatus – GET /notify/{id} — получение статуса уведомления
func (h notificationHandler) GetNotificationStatus(ctx *ginext.Context) {
	notificationID := ctx.Param("id")
	if notificationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "notification id is required"})
		return
	}

	notification, err := h.service.GetNotificationById(ctx, notificationID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, notification)
}

// CancelNotification – DELETE /notify/{id} — отмена запланированного уведомления
func (h notificationHandler) CancelNotification(ctx *ginext.Context) {
	notificationID := ctx.Param("id")
	if notificationID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "notification id is required"})
		return
	}

	err := h.service.MarkNotificationAsCancelled(ctx, notificationID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

func (h notificationHandler) Index(ctx *ginext.Context) {
	notifications, err := h.service.GetAllNotifications(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"notifications": notifications,
		"botUsername":   h.botUsername,
	})
}
