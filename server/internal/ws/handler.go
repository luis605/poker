package ws

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub      *Hub
	upgrader websocket.Upgrader
	logger   *slog.Logger
}

func NewHandler(hub *Hub, allowedOrigins []string, logger *slog.Logger) *Handler {
	// creates the equivalent to C++'s `std::unordered_set<std::string> originSet`.
	originSet := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = struct{}{}
	}

	return &Handler{
		hub:    hub,
		logger: logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}
				_, ok := originSet[origin]
				return ok
			},
		},
	}
}

func (h *Handler) Serve(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Warn("failed to upgrade websocket connection", "error", err, "remote_addr", c.Request.RemoteAddr)
		return
	}

	client := NewClient(h.hub, conn)
	h.hub.register <- client

	go client.writePump()
	go client.readPump()
}