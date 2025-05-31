package secondary

import (
	"context"
	
	"github.com/Anthony-Bible/password-exchange/app/internal/domains/notification/ports/contracts"
)

// TemplateRendererPort defines the secondary port for template rendering operations
type TemplateRendererPort interface {
	// RenderTemplate renders a notification template with the provided data
	RenderTemplate(ctx context.Context, templateName string, data contracts.NotificationTemplateData) (string, error)
}