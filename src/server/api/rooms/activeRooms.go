package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetActiveRooms(ctx *gin.Context) {
	store.mutex.RLock()
	rooms := make([]RoomResponse, 0, len(store.rooms))
	for _, room := range store.rooms {
		rooms = append(rooms, toRoomResponse(room))
	}
	store.mutex.RUnlock()

	ctx.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}
