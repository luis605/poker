package rooms

import (
	"net/http"
	"poker_server/helpers"

	"github.com/gin-gonic/gin"
)

const maxBCryptPasswordBytes = 72

func CreateRoom(c *gin.Context) {
	var input CreateRoomInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var hashedPassword string
	if input.IsPrivate {
		if input.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required for private rooms"})
			return
		}
		if len([]byte(input.Password)) > maxBCryptPasswordBytes {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be 72 bytes or fewer"})
			return
		}

		if len(input.Password) > 72 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password is too long"})
			return
		}

		var err error
		hashedPassword, err = helpers.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
			return
		}
	}

	newID := store.nextID.Add(1)
	room := Room{
		ID:           newID,
		Name:         input.Name,
		PlayerCount:  0,
		PlayerLimit:  4,
		IsPrivate:    input.IsPrivate,
		PasswordHash: hashedPassword,
		Players:      make(map[string]bool),
		Sessions:     make(map[string]string),
	}

	store.mu.Lock()
	store.rooms[room.ID] = room
	store.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Room created successfully",
		"room":    toRoomResponse(room),
	})
}
