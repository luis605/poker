package rooms

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const roomSessionHeader = "X-Room-Session-Token"

func roomSessionToken(ctx *gin.Context) (string, bool) {
	sessionToken := ctx.GetHeader(roomSessionHeader)
	if sessionToken == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Missing room session token"})
		return "", false
	}

	return sessionToken, true
}

func authenticatedRoomUsername(room Room, sessionToken string) (string, bool) {
	username, exists := room.Sessions[sessionToken]
	if !exists {
		return "", false
	}

	return username, true
}
