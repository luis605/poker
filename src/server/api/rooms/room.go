package rooms

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"sync/atomic"
)

type Room struct {
	ID           int64           `json:"id"`
	Name         string          `json:"name"`
	PlayerCount  int             `json:"playerCount"`
	PlayerLimit  int             `json:"playerLimit"`
	IsPrivate    bool            `json:"isPrivate"`
	PasswordHash string          `json:"-"`
	HostUsername string          `json:"-"`
	Players      map[string]bool `json:"-"`
	Sessions     map[string]string
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

type JoinRoomURI struct {
	ID int64 `uri:"id" binding:"required,gt=0"`
}

type JoinRoomInput struct {
	Username string `json:"username" binding:"required,min=2,max=20"`
	Password string `json:"password"`
}

type ExitRoomInput struct {
	Username string `json:"username" binding:"required,min=2,max=20"`
}

type RoomStore struct {
	mutex  sync.RWMutex
	rooms  map[int64]Room
	nextID atomic.Int64
}

func NewRoomStore() *RoomStore {
	return &RoomStore{
		rooms: make(map[int64]Room),
	}
}

var store = NewRoomStore()

func toRoomResponse(room Room) RoomResponse {
	return RoomResponse{
		ID:          room.ID,
		Name:        room.Name,
		PlayerCount: room.PlayerCount,
		PlayerLimit: room.PlayerLimit,
		IsPrivate:   room.IsPrivate,
	}
}

func generateRoomSessionToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(tokenBytes), nil
}
