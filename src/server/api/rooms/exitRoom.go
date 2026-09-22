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

	store.mu.Lock()
	defer store.mu.Unlock()

	room, exists := store.rooms[query.ID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	username, ok := authenticatedRoomUsername(c, room)
	if !ok {
		return
	}

	if !room.Players[username] {
		c.JSON(http.StatusNotFound, gin.H{"error": "Player is not in the room"})
		return
	}

	sessionToken := c.GetHeader(roomSessionHeader)
	delete(room.Players, username)
	delete(room.Sessions, sessionToken)
	room.PlayerCount = len(room.Players)
	store.rooms[query.ID] = room

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully exited the room",
		"player":  username,
		"room":    toRoomResponse(room),
	})
}
