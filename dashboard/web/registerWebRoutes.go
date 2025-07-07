package web

import (
	"database/sql"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/handlers"
	"github.com/Aniketyadav44/dscheduler/dashboard/internal/services"
	"github.com/gin-gonic/gin"
)

func registerWebRoutes(router *gin.Engine, db *sql.DB) {
	webService := services.NewWebServcie(db)
	webHandler := handlers.NewWebHandler(webService)

	d := router.Group("/")
	{
		d.GET("", webHandler.Home)
		d.GET("/jobs", webHandler.ListJobs)
		d.GET("/jobs/new", webHandler.Create)
		d.GET("/jobs/runs", webHandler.ListJobRuns)
	}
}
