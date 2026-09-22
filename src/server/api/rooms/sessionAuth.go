package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const roomSessionHeader = "X-Room-Session-Token"

func authenticatedRoomUsername(c *gin.Context, room Room) (string, bool) {
	sessionToken := c.GetHeader(roomSessionHeader)
	if sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing room session token"})
		return "", false
	}

	username, exists := room.Sessions[sessionToken]
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid room session token"})
		return "", false
	}

	return username, true
}
