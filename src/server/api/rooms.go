package api

import (
	"net/http"
	"poker_server/helpers"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

type Room struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PlayerCount int    `json:"playerCount"`
	PlayerLimit int    `json:"playerLimit"`
	IsPrivate   bool   `json:"isPrivate"`
	Password    string `json:"-"` // Omit from JSON serialization to prevent password leaks
}

type RoomResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	PlayerCount int    `json:"playerCount"`
	PlayerLimit int    `json:"playerLimit"`
	IsPrivate   bool   `json:"isPrivate"`
}

type CreateRoomInput struct {
	Name      string `json:"name" binding:"required,min=1,max=20"`
	IsPrivate bool   `json:"isPrivate"`
	Password  string `json:"password"`
}

type RoomStore struct {
	mu     sync.RWMutex
	rooms  map[int64]Room
	nextID atomic.Int64
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		rooms: make(map[int64]Room),
	}
}

var store = NewRoomStore()

func createRoom(c *gin.Context) {
	var input CreateRoomInput

	// Validates JSON structure and basic field constraints automatically
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

		// Hash the password before saving
		var err error
		hashedPassword, err = helpers.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to secure password"})
			return
		}
	}

	newID := store.nextID.Add(1)

	room := Room{
		ID:          newID,
		Name:        input.Name,
		PlayerCount: 0,
		PlayerLimit: 4,
		IsPrivate:   input.IsPrivate,
		Password:    hashedPassword, // Consider hashing passwords if security is a priority
	}

	// Lock the store for thread-safe writing
	store.mu.Lock()
	store.rooms[room.ID] = room
	store.mu.Unlock()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Room created successfully",
		"room": RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			PlayerCount: room.PlayerCount,
			PlayerLimit: room.PlayerLimit,
			IsPrivate:   room.IsPrivate,
		},
	})
}

func getActiveRooms(c *gin.Context) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Initialize as empty slice (not nil) so JSON renders `[]` instead of `null`
	rooms := make([]RoomResponse, 0, len(store.rooms))
	for _, room := range store.rooms {
		rooms = append(rooms, RoomResponse{
			ID:          room.ID,
			Name:        room.Name,
			PlayerCount: room.PlayerCount,
			PlayerLimit: room.PlayerLimit,
			IsPrivate:   room.IsPrivate,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"rooms": rooms,
	})
}
