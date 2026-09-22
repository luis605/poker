package api

import (
	"log"
	"poker_server/api/rooms"

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
		api.POST("create-room", rooms.CreateRoom)
		api.GET("active-rooms", rooms.GetActiveRooms)
		api.POST("join-room", rooms.JoinRoom)
		api.DELETE("delete-room", rooms.DeleteRoom)
		api.POST("exit-room", rooms.ExitRoom)
	}
}
