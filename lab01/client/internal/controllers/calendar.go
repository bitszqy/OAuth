package controllers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/cyllective/oauth-labs/lab01/client/internal/client"
	"github.com/cyllective/oauth-labs/lab01/client/internal/services"
	"github.com/cyllective/oauth-labs/lab01/client/internal/session"
)

type CalendarController struct {
	tokenService *services.TokenService
}

func NewCalendarController(tokenService *services.TokenService) *CalendarController {
	return &CalendarController{tokenService: tokenService}
}

func (cc *CalendarController) Get(c *gin.Context) {
	s := sessions.Default(c)
	taskID := c.Query("task")
	if taskID == "" {
		if stored, ok := session.GetString(s, "task_id"); ok {
			taskID = stored
		}
	}
	if taskID != "calendar.freebusy" && taskID != "calendar.reminder" {
		taskID = "calendar.freebusy"
	}

	tokens, err := cc.tokenService.Get(s)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "calendar.tmpl", gin.H{
			"IsAuthenticated": true,
			"Error":           "Unable to load OAuth tokens.",
			"TaskID":          taskID,
		})
		return
	}

	view, err := client.NewAPIClient(c.Request.Context(), tokens).GetCalendarTask(taskID)
	if err != nil {
		c.HTML(http.StatusForbidden, "calendar.tmpl", gin.H{
			"IsAuthenticated": true,
			"Error":           err.Error(),
			"TaskID":          taskID,
		})
		return
	}

	c.HTML(http.StatusOK, "calendar.tmpl", gin.H{
		"IsAuthenticated": true,
		"TaskID":          taskID,
		"View":            view,
	})
}
