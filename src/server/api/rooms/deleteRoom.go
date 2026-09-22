package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeleteRoom(c *gin.Context) {
	var query JoinRoomQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ?id parameter"})
		return
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if _, exists := store.rooms[query.ID]; !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	delete(store.rooms, query.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Room deleted successfully",
	})
}
