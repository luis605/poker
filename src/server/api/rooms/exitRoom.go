package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ExitRoom(ctx *gin.Context) {
	var query JoinRoomQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ?id parameter"})
		return
	}

	store.mutex.RLock()
	room, exists := store.rooms[query.ID]
	store.mutex.RUnlock()
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	sessionToken, ok := roomSessionToken(ctx)
	if !ok {
		return
	}

	store.mutex.Lock()
	room, exists = store.rooms[query.ID]
	if !exists {
		store.mutex.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	username, ok := authenticatedRoomUsername(room, sessionToken)
	if !ok {
		store.mutex.Unlock()
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid room session token"})
		return
	}
	if !room.Players[username] {
		store.mutex.Unlock()
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Player is not in the room"})
		return
	}
	delete(room.Players, username)
	delete(room.Sessions, sessionToken)
	room.PlayerCount = len(room.Players)
	store.rooms[query.ID] = room
	responseRoom := toRoomResponse(room)
	store.mutex.Unlock()

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Successfully exited the room",
		"player":  username,
		"room":    responseRoom,
	})
}
