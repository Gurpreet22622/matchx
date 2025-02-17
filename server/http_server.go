package server

import (
	"database/sql"
	"log"
	"matchx/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type HttpServer struct {
	config *viper.Viper
	router *gin.Engine
}

func InitHttpServer(config *viper.Viper, dbHandler *sql.DB) HttpServer {
	// usersRepository := repositories.NewUsersRepository(dbHandler)
	// usersService := services.NewUsersService(usersRepository)
	// usersController := controllers.NewUsersController(usersService)

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://192.168.29.173:3000", "http://localhost:3000"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "Token"},
	}))

	//router.POST("/login", usersController.Login)
	//router.POST("/logout", usersController.Logout)
	router.POST("/login", controller.Login)
	router.GET("/user", controller.GetUser)
	router.POST("/registerNP", controller.RegisterNP)
	router.GET("/nearbyProps", controller.GetNearbyProps)

	hs := HttpServer{
		config: config,
		router: router,
	}

	return hs
}

func (hs HttpServer) Start() {
	//err := hs.router.Run(hs.config.GetString("http.server_address"))
	err := hs.router.Run("0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Error while starting HTTP server: %v", err)
	}
}
