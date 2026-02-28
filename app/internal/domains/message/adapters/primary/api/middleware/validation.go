package middleware

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/Anthony-Bible/password-exchange/app/internal/domains/message/adapters/primary/api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function to use json tags
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	validate.RegisterValidation("required_if_notification", requiredIfNotification)
	validate.RegisterValidation("antispam_blue", antiSpamBlue)
}

// ValidationMiddleware creates a middleware that validates request bodies
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set request size limits
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20) // 1MB limit

		c.Next()
	}
}

// ValidateStruct validates a struct and returns formatted validation errors
func ValidateStruct(s interface{}) map[string]interface{} {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errors := make(map[string]interface{})

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := fieldError.Field()
			errors[fieldName] = getValidationErrorMessage(fieldError)
		}
	}

	return errors
}

// ValidateMessageSubmission validates a message submission request with conditional logic
func ValidateMessageSubmission(req *models.MessageSubmissionRequest) map[string]interface{} {
	errors := make(map[string]interface{})

	// Basic struct validation
	if structErrors := ValidateStruct(req); structErrors != nil {
		for k, v := range structErrors {
			errors[k] = v
		}
	}

	// Conditional validation for notifications
	if req.SendNotification {
		if req.Sender == nil {
			errors["sender"] = "Sender information is required when notifications are enabled"
		} else {
			if req.Sender.Name == "" {
				errors["sender.name"] = "Sender name is required when notifications are enabled"
			}
			if req.Sender.Email == "" {
				errors["sender.email"] = "Sender email is required when notifications are enabled"
			} else if senderErrors := ValidateStruct(req.Sender); senderErrors != nil {
				for k, v := range senderErrors {
					errors["sender."+k] = v
				}
			}
		}

		if req.Recipient == nil {
			errors["recipient"] = "Recipient information is required when notifications are enabled"
		} else {
			if req.Recipient.Name == "" {
				errors["recipient.name"] = "Recipient name is required when notifications are enabled"
			}
			if req.Recipient.Email == "" {
				errors["recipient.email"] = "Recipient email is required when notifications are enabled"
			} else if recipientErrors := ValidateStruct(req.Recipient); recipientErrors != nil {
				for k, v := range recipientErrors {
					errors["recipient."+k] = v
				}
			}
		}

		// Anti-spam validation
		if req.AntiSpamAnswer == "" {
			errors["antiSpamAnswer"] = "Anti-spam answer is required when notifications are enabled"
		} else if !IsValidAntiSpamAnswer(req.QuestionID, req.AntiSpamAnswer) {
			errors["antiSpamAnswer"] = "Invalid anti-spam answer"
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return errors
}

// RequestTimeoutMiddleware adds request timeout handling
func RequestTimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a timeout on the request context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context with our timeout context
		c.Request = c.Request.WithContext(ctx)

		// Process the request
		c.Next()

		// Check if the context was cancelled due to timeout
		if ctx.Err() == context.DeadlineExceeded {
			JSONErrorResponse(c, http.StatusRequestTimeout,
				models.ErrorCodeTimeout,
				"Request timeout exceeded", nil)
		}
	}
}

// Custom validator functions

// requiredIfNotification validates that a field is required if SendNotification is true
func requiredIfNotification(fl validator.FieldLevel) bool {
	// Get the parent struct
	parent := fl.Parent()
	if parent.Kind() != reflect.Struct {
		return true
	}

	// Look for SendNotification field
	sendNotificationField := parent.FieldByName("SendNotification")
	if !sendNotificationField.IsValid() {
		return true
	}

	// If SendNotification is true, field must not be empty
	if sendNotificationField.Bool() {
		return fl.Field().String() != ""
	}

	return true
}

// antiSpamBlue validates that the anti-spam answer is "blue"
func antiSpamBlue(fl validator.FieldLevel) bool {
	value := strings.ToLower(strings.TrimSpace(fl.Field().String()))
	return value == "blue"
}

// getValidationErrorMessage returns a human-readable error message for validation errors
func getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fe.Field())
	case "email":
		return "Must be a valid email address"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Must be no more than %s characters", fe.Param())
	case "required_if_notification":
		return "This field is required when notifications are enabled"
	case "antispam_blue":
		return "Invalid anti-spam answer"
	default:
		return fmt.Sprintf("Validation failed for %s", fe.Field())
	}
}

// IsValidAntiSpamAnswer validates the anti-spam answer based on question ID
func IsValidAntiSpamAnswer(questionID *int, answer string) bool {
	if questionID == nil {
		// Default to question 0 if not provided (backward compatibility)
		defaultID := 0
		questionID = &defaultID
	}

	// Define the question-answer mapping
	validAnswers := map[int][]string{
		0: {"blue"},        // What color is the sky?
		1: {"4", "four"},   // What is 2 + 2?
		2: {"7", "seven"},  // How many days are in a week?
		3: {"cat", "cats"}, // What animal says meow?
		4: {"pen"},         // What do you use to write?
		5: {"4", "four"},   // How many legs does a dog have?
	}

	answers, exists := validAnswers[*questionID]
	if !exists {
		return false
	}

	userAnswer := strings.ToLower(strings.TrimSpace(answer))
	for _, validAnswer := range answers {
		if userAnswer == validAnswer {
			return true
		}
	}

	return false
}
