package usecase

import (
	"context"
	"log"
	"time"

	"github.com/auto-devs/auto-devs/internal/entity"
	"github.com/auto-devs/auto-devs/internal/service/webhook"
)

type WebhookNotificationHandler struct {
	client *webhook.Client
}

func NewWebhookNotificationHandler(client *webhook.Client) *WebhookNotificationHandler {
	return &WebhookNotificationHandler{client: client}
}

func (h *WebhookNotificationHandler) HandleNotification(event entity.NotificationEvent) error {
	payload := map[string]any{
		"project_id": event.ProjectID,
		"task_id":    event.TaskID,
		"new_status": event.Data["to_status"],
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := h.client.Send(ctx, payload); err != nil {
			log.Printf("webhook notification failed: %v", err)
		}
	}()

	return nil
}
