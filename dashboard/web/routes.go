package web

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	registerApiRoutes(router, db)
	registerWebRoutes(router, db)
}
