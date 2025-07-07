package main

import (
	"log"

	"github.com/Aniketyadav44/dscheduler/dashboard/internal/config"
	"github.com/Aniketyadav44/dscheduler/dashboard/web"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err.Error())
	}

	defer cfg.DB.Close()

	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.Static("/static", "./web/static")
	// for handling web interface's session
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	web.RegisterRoutes(router, cfg.DB)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal("error in starting server: ", err.Error())
	}
}
