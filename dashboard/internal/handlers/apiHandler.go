package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/models"
	"github.com/Aniketyadav44/dscheduler/dashboard/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	service *services.APIService
}

func NewAPIHandler(service *services.APIService) *APIHandler {
	return &APIHandler{
		service: service,
	}
}

func (h *APIHandler) CreateJob(c *gin.Context) {
	hour := c.PostForm("hour")
	minute := c.PostForm("minute")
	payload := make(map[string]any)
	payload["hour"] = hour
	payload["minute"] = minute
	taskType := c.PostForm("type")
	timezone := c.PostForm("timezone")
	switch taskType {
	case "ping":
		payload["url"] = c.PostForm("url")
	case "email":
		payload["email"] = c.PostForm("email")
		payload["subject"] = c.PostForm("subject")
		payload["body"] = c.PostForm("body")
	case "slack":
		payload["url"] = c.PostForm("url")
		payload["msg"] = c.PostForm("msg")
	case "webhook":
		payload["url"] = c.PostForm("url")
		payload["body"] = c.PostForm("body")
	}

	hourInt, err := strconv.Atoi(hour)
	if err != nil {
		log.Println("invalid hour: ", err.Error())
		session := sessions.Default(c) // using sessions only for handling & showing errors
		session.Set("error", err.Error())
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/jobs/new")
		return
	}

	minuteInt, err := strconv.Atoi(minute)
	if err != nil {
		log.Println("invalid minute: ", err.Error())
		session := sessions.Default(c) // using sessions only for handling & showing errors
		session.Set("error", err.Error())
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/jobs/new")
		return
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Println("invalid timezone: ", err.Error())
		session := sessions.Default(c) // using sessions only for handling & showing errors
		session.Set("error", err.Error())
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/jobs/new")
		return
	}

	log.Println("got timezone: ", loc)

	jobRunTimeUTC := time.Date(2000, time.January, 1, hourInt, minuteInt, 0, 0, loc).UTC()

	log.Println("utc time: ", jobRunTimeUTC)

	job := &models.Job{
		Hour:    jobRunTimeUTC.Hour(),
		Minute:  jobRunTimeUTC.Minute(),
		Type:    taskType,
		Payload: payload,
	}

	if err := h.service.CreateNewJob(job); err != nil {
		log.Println("error in creating cron job: ", err.Error())
		session := sessions.Default(c) // using sessions only for handling & showing errors
		session.Set("error", err.Error())
		session.Save()
		c.Redirect(http.StatusMovedPermanently, "/jobs/new")
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/jobs?page=1&limit=10")
}

func (h *APIHandler) DeleteJob(c *gin.Context) {
	id := c.Request.URL.Query().Get("id")
	if id == "" {
		c.String(http.StatusBadRequest, "invalid id query")
		return
	}

	idInt, _ := strconv.Atoi(id)
	err := h.service.DeleteJob(idInt)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		return
	}

	c.Redirect(http.StatusMovedPermanently, "/jobs?page=1&limit=10")
}
