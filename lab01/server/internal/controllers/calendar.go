package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab01/server/internal/dto"
	"github.com/cyllective/oauth-labs/lab01/server/internal/services"
	"github.com/cyllective/oauth-labs/lab01/server/internal/taskpolicy"
	"github.com/cyllective/oauth-labs/lab01/server/internal/variant"
)

type CalendarController struct {
	crypto   *services.AttributeCryptoService
	taskKeys *services.TaskKeyService
}

func NewCalendarController(crypto *services.AttributeCryptoService, taskKeys *services.TaskKeyService) *CalendarController {
	return &CalendarController{crypto: crypto, taskKeys: taskKeys}
}

// GET /api/calendar/events
func (cc *CalendarController) Events(c *gin.Context) {
	token := c.MustGet("access_token").(*dto.AccessToken)
	events := sampleEvents(token.UserID)
	if variant.IsVulnerable() {
		c.JSON(http.StatusOK, dto.CalendarResponse{Events: events})
		return
	}

	encrypted := make([]dto.EncryptedCalendarEvent, 0, len(events))
	for _, event := range events {
		fields := map[string]string{
			taskpolicy.Start:       event.Start,
			taskpolicy.End:         event.End,
			taskpolicy.Title:       event.Title,
			taskpolicy.Location:    event.Location,
			taskpolicy.Attendees:   strings.Join(event.Attendees, ", "),
			taskpolicy.Description: event.Description,
		}
		encryptedFields := make(map[string]dto.EncryptedField, len(fields))
		for attribute, value := range fields {
			field, err := cc.crypto.Encrypt(token.UserID, event.ResourceID, attribute, value)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to protect calendar field"})
				return
			}
			encryptedFields[attribute] = field
		}
		encrypted = append(encrypted, dto.EncryptedCalendarEvent{
			ID:         event.ID,
			ResourceID: event.ResourceID,
			Fields:     encryptedFields,
		})
	}
	c.JSON(http.StatusOK, dto.CalendarResponse{Encrypted: encrypted})
}

// POST /api/task-keys
func (cc *CalendarController) TaskKeys(c *gin.Context) {
	token := c.MustGet("access_token").(*dto.AccessToken)
	var request dto.TaskKeyRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task key request"})
		return
	}
	expectedResourceID := fmt.Sprintf("calendar:%s", token.UserID)
	if request.ResourceID != expectedResourceID {
		c.JSON(http.StatusForbidden, gin.H{"error": "resource_is_not_authorized"})
		return
	}

	response, err := cc.taskKeys.Release(c.Request.Context(), token.UserID, token.ClientID, &request)
	if err != nil {
		if errors.Is(err, services.ErrTaskNotAuthorized) {
			c.JSON(http.StatusForbidden, gin.H{"error": "task_is_not_authorized"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to release task keys"})
		return
	}
	c.JSON(http.StatusOK, response)
}

func sampleEvents(userID string) []dto.PlainCalendarEvent {
	resourceID := fmt.Sprintf("calendar:%s", userID)
	return []dto.PlainCalendarEvent{
		{
			ID:          "evt-001",
			ResourceID:  resourceID,
			Start:       "2026-11-13T09:00:00+08:00",
			End:         "2026-11-13T09:30:00+08:00",
			Title:       "Security project review",
			Location:    "Room 4A",
			Attendees:   []string{"alice@example.test", "bob@example.test"},
			Description: "Review the threat model and discuss the prototype.",
		},
		{
			ID:          "evt-002",
			ResourceID:  resourceID,
			Start:       "2026-11-13T14:00:00+08:00",
			End:         "2026-11-13T15:00:00+08:00",
			Title:       "Doctor appointment",
			Location:    "City Clinic",
			Attendees:   []string{"user@example.test"},
			Description: "Private appointment details.",
		},
	}
}
