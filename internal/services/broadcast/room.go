package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"wordwizardry/internal/pkg/models"
)

type Room struct {
	clients map[string]*Client // playerID -> client
	mu      sync.RWMutex
	players map[string]bool // playerID -> true
}

func (h *WebSocketHub) CreateRoom(sessionID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.rooms[sessionID]; exists {
		return fmt.Errorf("room already exists")
	}

	h.rooms[sessionID] = &Room{
		clients: make(map[string]*Client),
		players: make(map[string]bool),
	}
	return nil
}

func (h *WebSocketHub) JoinRoom(sessionID string, playerID string) error {
	h.mu.RLock()
	room, exists := h.rooms[sessionID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if _, exists := room.clients[playerID]; exists {
		return fmt.Errorf("player already in room")
	}

	// Just register the player, Client will be created when WebSocket connects
	room.players[playerID] = true

	return nil
}

func (h *WebSocketHub) LeaveRoom(sessionID string, playerID string) error {
	h.mu.RLock()
	room, exists := h.rooms[sessionID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if _, exists := room.clients[playerID]; !exists {
		return fmt.Errorf("player not in room")
	}

	delete(room.clients, playerID)
	return nil
}

func (h *WebSocketHub) BroadcastToRoom(ctx context.Context, sessionID string, message models.WSMessage) error {
	h.mu.RLock()
	room, exists := h.rooms[sessionID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, client := range room.clients {
		select {
		case client.Send <- data:
		default:
			go func(c *Client) {
				h.unregister <- c
			}(client)
		}
	}

	return nil
}
