package rooms

import (
	"net/http"
	"poker_server/helpers"

	"github.com/gin-gonic/gin"
)

// required,gt=0 enforces a present positive room ID and rejects missing/zero/negative values.
type JoinRoomQuery struct {
	ID int64 `form:"id" binding:"required,gt=0"`
}

func JoinRoom(ctx *gin.Context) {
	var query JoinRoomQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ?id parameter"})
		return
	}

	var input JoinRoomInput
	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	store.mutex.RLock()
	room, exists := store.rooms[query.ID]
	store.mutex.RUnlock()
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if room.IsPrivate && input.Password == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "This room is private. A password is required",
		})
		return
	}
	if room.IsPrivate && !helpers.CheckPasswordHash(input.Password, room.PasswordHash) {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Incorrect room password",
		})
		return
	}

	sessionToken, err := generateRoomSessionToken()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to initialize room session",
		})
		return
	}

	store.mutex.Lock()
	room, exists = store.rooms[query.ID]
	if !exists {
		store.mutex.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	if room.Players == nil {
		room.Players = make(map[string]bool)
	}
	if room.Sessions == nil {
		room.Sessions = make(map[string]string)
	}
	if room.Players[input.Username] {
		store.mutex.Unlock()
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "Someone already has this username in this room",
		})
		return
	}
	if room.PlayerCount >= room.PlayerLimit {
		store.mutex.Unlock()
		ctx.JSON(http.StatusConflict, gin.H{
			"error": "Room is already full",
		})
		return
	}
	if room.HostUsername == "" {
		room.HostUsername = input.Username
	}
	room.Players[input.Username] = true
	room.Sessions[sessionToken] = input.Username
	room.PlayerCount = len(room.Players)
	store.rooms[query.ID] = room
	responseRoom := toRoomResponse(room)
	store.mutex.Unlock()

	ctx.JSON(http.StatusOK, gin.H{
		"message":      "Successfully joined the room",
		"player":       input.Username,
		"sessionToken": sessionToken,
		"room":         responseRoom,
	})
}
