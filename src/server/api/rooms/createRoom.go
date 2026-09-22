package rooms

import (
	"net/http"
	"poker_server/helpers"

	"github.com/gin-gonic/gin"
)

const maxBCryptPasswordBytes = 72

func CreateRoom(ctx *gin.Context) {
	var input CreateRoomInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var hashedPassword string
	if input.IsPrivate && input.Password == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Password is required for private rooms"})
		return
	}
	if input.IsPrivate {
		if len([]byte(input.Password)) > maxBCryptPasswordBytes {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Password must be 72 bytes or fewer"})
			return
		}

		var err error
		hashedPassword, err = helpers.HashPassword(input.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
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

	store.mutex.Lock()
	store.rooms[room.ID] = room
	store.mutex.Unlock()

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Room created successfully",
		"room":    toRoomResponse(room),
	})
}
