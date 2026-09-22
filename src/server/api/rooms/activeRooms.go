package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetActiveRooms(c *gin.Context) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	rooms := make([]RoomResponse, 0, len(store.rooms))
	for _, room := range store.rooms {
		rooms = append(rooms, toRoomResponse(room))
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}
