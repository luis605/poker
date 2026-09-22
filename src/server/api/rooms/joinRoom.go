package rooms

import (
	"net/http"
	"poker_server/helpers"

	"github.com/gin-gonic/gin"
)

// Bind query parameters (?id=2)
type JoinRoomQuery struct {
	ID int64 `form:"id" binding:"required,gt=0"`
}

func JoinRoom(c *gin.Context) {
	// 1. Read query param (?id=...)
	var query JoinRoomQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ?id parameter"})
		return
	}

	// 2. Read body
	var input JoinRoomInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// 3. Look up by query.ID
	room, exists := store.rooms[query.ID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// --- SAFETY CHECKS ---

	// Defensive check: ensure the map is initialized
	if room.Players == nil {
		room.Players = make(map[string]bool)
	}

	// 4. Check if player is already inside the room (PREVENT JOINING TWICE)
	if room.Players[input.Username] {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Someone already has this username in this room",
		})
		return
	}

	// 5. Check room capacity
	if room.PlayerCount >= room.PlayerLimit {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Room is already full",
		})
		return
	}

	// 6. Check privacy & password verification
	if room.IsPrivate {
		if input.Password == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "This room is private. A password is required",
			})
			return
		}

		if !helpers.CheckPasswordHash(input.Password, room.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Incorrect room password",
			})
			return
		}
	}

	// 7. Update room state: register user and sync count
	room.Players[input.Username] = true
	room.PlayerCount = len(room.Players)
	store.rooms[query.ID] = room

	// 8. Respond with sanitized room details
	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully joined the room",
		"player":  input.Username,
		"room":    toRoomResponse(room),
	})
}
