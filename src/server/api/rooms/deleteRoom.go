package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeleteRoom(ctx *gin.Context) {
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
	if username != room.HostUsername {
		store.mutex.Unlock()
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Only the room host can delete this room"})
		return
	}

	delete(store.rooms, query.ID)
	store.mutex.Unlock()

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Room deleted successfully",
	})
}
