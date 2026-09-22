package api

import (
	"log"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	r := gin.Default()

	mapRoutes(r)

	if err := r.Run(); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}

func mapRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Rooms
		api.POST("create-room", createRoom)
		api.GET("active-rooms", getActiveRooms)
	}
}
