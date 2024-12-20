package broadcast

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"wordwizardry/internal/pkg/models"
	"wordwizardry/internal/pkg/websocket"
)

type Hub interface {
	Run()
	CreateRoom(sessionID string) error
	JoinRoom(sessionID string, playerID string) error
	LeaveRoom(sessionID string, playerID string) error
	BroadcastToRoom(ctx context.Context, sessionID string, message models.WSMessage) error
	SendToPlayer(ctx context.Context, sessionID, playerID string, message models.WSMessage) error
	HandleWebSocket(w http.ResponseWriter, r *http.Request) error
}

type WebSocketHub struct {
	rooms      map[string]*Room // sessionID -> room
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.RLock()
			room, exists := h.rooms[client.SessionID]
			h.mu.RUnlock()

			if !exists {
				log.Printf("attempted to register client to non-existent room: %s", client.SessionID)
				continue
			}

			room.mu.Lock()
			_, ok := room.clients[client.PlayerID]
			if ok {
				// Mark for cleanup without closing channel
				delete(room.clients, client.PlayerID)
			}

			room.clients[client.PlayerID] = client
			room.mu.Unlock()

		case client := <-h.unregister:
			h.mu.RLock()
			room, exists := h.rooms[client.SessionID]
			h.mu.RUnlock()

			if exists {
				room.mu.Lock()
				if currentClient, ok := room.clients[client.PlayerID]; ok {
					// Only remove if it's the current client
					if currentClient == client {
						delete(room.clients, client.PlayerID)
					}
				}
				// Only remove room if truly empty
				if len(room.clients) == 0 {
					h.mu.Lock()
					delete(h.rooms, client.SessionID)
					h.mu.Unlock()
				}
				room.mu.Unlock()
			}
			// Close channel after all locks are released
			select {
			case <-client.Send: // Drain any pending message
			default:
			}
		}
	}
}

func (h *WebSocketHub) HandleWebSocket(w http.ResponseWriter, r *http.Request) error {
	sessionID := r.URL.Query().Get("session_id")
	playerID := r.URL.Query().Get("player_id")

	if !h.IsPlayerRegistered(sessionID, playerID) {
		return fmt.Errorf("player not registered in room")
	}

	conn, err := websocket.Upgrade(w, r)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	client := &Client{
		Hub:       h,
		Conn:      conn,
		SessionID: sessionID,
		PlayerID:  playerID,
		Send:      make(chan []byte, 256),
	}

	h.mu.RLock()
	room := h.rooms[sessionID]
	h.mu.RUnlock()

	room.mu.Lock()
	room.clients[playerID] = client
	room.mu.Unlock()

	h.register <- client

	go client.writePump()
	go client.readPump()

	return nil
}
