package handlers

import (
	"net/http"
	"strconv"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/services"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type WebHandler struct {
	service *services.WebService
}

func NewWebHandler(service *services.WebService) *WebHandler {
	return &WebHandler{
		service: service,
	}
}

func (h *WebHandler) Home(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}

func (h *WebHandler) Create(c *gin.Context) {
	session := sessions.Default(c)
	err := session.Get("error")
	if err != nil {
		session.Delete("error")
		session.Save()
	}
	c.HTML(http.StatusOK, "create-form.html", gin.H{
		"Hours":        24,
		"Minutes":      60,
		"ErrorMessage": err,
	})
}

func (h *WebHandler) ListJobs(c *gin.Context) {
	page := c.Request.URL.Query().Get("page")
	limit := c.Request.URL.Query().Get("limit")
	pageI, oErr := strconv.Atoi(page)
	limitI, lErr := strconv.Atoi(limit)
	if page == "" || oErr != nil {
		c.String(http.StatusBadRequest, "invalid page")
		return
	}
	if limit == "" || lErr != nil {
		c.String(http.StatusBadRequest, "invalid limit")
		return
	}
	if pageI <= 0 {
		c.String(http.StatusBadRequest, "page cannot be 0 or negative")
		return
	}

	jobs, err := h.service.GetAllJobs(pageI, limitI)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	prevPage := pageI - 1
	if prevPage < 0 {
		prevPage = 0
	}
	c.HTML(http.StatusOK, "list-jobs.html", gin.H{
		"Jobs":  jobs,
		"Prev":  prevPage,
		"Next":  pageI + 1,
		"Limit": limitI,
	})
}

func (h *WebHandler) ListJobRuns(c *gin.Context) {
	jobId := c.Request.URL.Query().Get("id")
	status := c.Request.URL.Query().Get("status")
	page := c.Request.URL.Query().Get("page")
	limit := c.Request.URL.Query().Get("limit")
	jobIdI, idErr := strconv.Atoi(jobId)
	pageI, oErr := strconv.Atoi(page)
	limitI, lErr := strconv.Atoi(limit)
	if jobId == "" || idErr != nil {
		c.String(http.StatusBadRequest, "invalid job ID")
		return
	}
	if page == "" || oErr != nil {
		c.String(http.StatusBadRequest, "invalid page")
		return
	}
	if limit == "" || lErr != nil {
		c.String(http.StatusBadRequest, "invalid limit")
		return
	}
	if pageI <= 0 {
		c.String(http.StatusBadRequest, "page cannot be 0 or negative")
		return
	}

	jobEntries, err := h.service.GetJobRuns(jobIdI, pageI, limitI, status)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	prevPage := pageI - 1
	if prevPage < 0 {
		prevPage = 0
	}
	c.HTML(http.StatusOK, "job-entries.html", gin.H{
		"JobId":   jobIdI,
		"Entries": jobEntries,
		"Filter":  status,
		"Prev":    prevPage,
		"Next":    pageI + 1,
		"Limit":   limitI,
	})
}
