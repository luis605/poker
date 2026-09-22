package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExitRoom(c *gin.Context) {
	var query JoinRoomQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ?id parameter"})
		return
	}

	var input ExitRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	room, exists := store.rooms[query.ID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if !room.Players[input.Username] {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player is not in the room"})
		return
	}

	delete(room.Players, input.Username)
	room.PlayerCount = len(room.Players)
	store.rooms[query.ID] = room

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully exited the room",
		"player":  input.Username,
		"room":    toRoomResponse(room),
	})
}
