package rooms

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRoomsTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store = NewRoomStore()

	router := gin.New()
	router.POST("/create-room", CreateRoom)
	router.POST("/join-room", JoinRoom)
	router.POST("/exit-room", ExitRoom)
	router.DELETE("/delete-room", DeleteRoom)
	return router
}

func performRequest(router *gin.Engine, method, target, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateRoomRejectsPrivatePasswordOver72Bytes(t *testing.T) {
	router := setupRoomsTestRouter()
	password := strings.Repeat("a", maxBCryptPasswordBytes+1)

	rec := performRequest(router, http.MethodPost, "/create-room", `{"name":"room","isPrivate":true,"password":"`+password+`"}`, nil)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteRoomRequiresHostToken(t *testing.T) {
	router := setupRoomsTestRouter()

	createRec := performRequest(router, http.MethodPost, "/create-room", `{"name":"room","isPrivate":false}`, nil)
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected room creation status %d, got %d", http.StatusCreated, createRec.Code)
	}

	hostJoinRec := performRequest(router, http.MethodPost, "/join-room?id=1", `{"username":"host"}`, nil)
	if hostJoinRec.Code != http.StatusOK {
		t.Fatalf("expected host join status %d, got %d", http.StatusOK, hostJoinRec.Code)
	}
	guestJoinRec := performRequest(router, http.MethodPost, "/join-room?id=1", `{"username":"guest"}`, nil)
	if guestJoinRec.Code != http.StatusOK {
		t.Fatalf("expected guest join status %d, got %d", http.StatusOK, guestJoinRec.Code)
	}

	var hostJoinBody map[string]any
	if err := json.Unmarshal(hostJoinRec.Body.Bytes(), &hostJoinBody); err != nil {
		t.Fatalf("failed to decode host join response: %v", err)
	}
	hostToken, _ := hostJoinBody["sessionToken"].(string)

	var guestJoinBody map[string]any
	if err := json.Unmarshal(guestJoinRec.Body.Bytes(), &guestJoinBody); err != nil {
		t.Fatalf("failed to decode guest join response: %v", err)
	}
	guestToken, _ := guestJoinBody["sessionToken"].(string)

	if rec := performRequest(router, http.MethodDelete, "/delete-room?id=1", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected %d without token, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec := performRequest(router, http.MethodDelete, "/delete-room?id=1", "", map[string]string{roomSessionHeader: guestToken}); rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d for non-host token, got %d", http.StatusForbidden, rec.Code)
	}
	if rec := performRequest(router, http.MethodDelete, "/delete-room?id=1", "", map[string]string{roomSessionHeader: hostToken}); rec.Code != http.StatusOK {
		t.Fatalf("expected %d for host token, got %d", http.StatusOK, rec.Code)
	}
}

func TestExitRoomRemovesAuthenticatedMemberOnly(t *testing.T) {
	router := setupRoomsTestRouter()

	if rec := performRequest(router, http.MethodPost, "/create-room", `{"name":"room","isPrivate":false}`, nil); rec.Code != http.StatusCreated {
		t.Fatalf("expected room creation status %d, got %d", http.StatusCreated, rec.Code)
	}

	hostJoinRec := performRequest(router, http.MethodPost, "/join-room?id=1", `{"username":"host"}`, nil)
	guestJoinRec := performRequest(router, http.MethodPost, "/join-room?id=1", `{"username":"guest"}`, nil)
	if hostJoinRec.Code != http.StatusOK || guestJoinRec.Code != http.StatusOK {
		t.Fatalf("expected join status %d for both players, got host=%d guest=%d", http.StatusOK, hostJoinRec.Code, guestJoinRec.Code)
	}

	var hostJoinBody map[string]any
	if err := json.Unmarshal(hostJoinRec.Body.Bytes(), &hostJoinBody); err != nil {
		t.Fatalf("failed to decode host join response: %v", err)
	}
	hostToken, _ := hostJoinBody["sessionToken"].(string)

	exitRec := performRequest(router, http.MethodPost, "/exit-room?id=1", `{"username":"guest"}`, map[string]string{roomSessionHeader: hostToken})
	if exitRec.Code != http.StatusOK {
		t.Fatalf("expected %d for authenticated host exit, got %d", http.StatusOK, exitRec.Code)
	}

	store.mutex.RLock()
	remainingRoom := store.rooms[1]
	_, hostStillInRoom := remainingRoom.Players["host"]
	_, guestStillInRoom := remainingRoom.Players["guest"]
	store.mutex.RUnlock()

	if hostStillInRoom {
		t.Fatal("expected host to be removed")
	}
	if !guestStillInRoom {
		t.Fatal("expected guest to remain in room")
	}
}
