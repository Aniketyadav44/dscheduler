package web

import (
	"database/sql"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/handlers"
	"github.com/Aniketyadav44/dscheduler/dashboard/internal/services"
	"github.com/gin-gonic/gin"
)

func registerApiRoutes(router *gin.Engine, db *sql.DB) {
	apiService := services.NewAPIService(db)
	apiHandler := handlers.NewAPIHandler(apiService)

	v1 := router.Group("/api/v1/job")
	{
		v1.POST("/create", apiHandler.CreateJob)
		v1.POST("/delete", apiHandler.DeleteJob)
	}
}
