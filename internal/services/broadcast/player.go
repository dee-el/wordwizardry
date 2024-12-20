package broadcast

import (
	"context"
	"encoding/json"
	"fmt"

	"wordwizardry/internal/pkg/models"
)

func (h *WebSocketHub) SendToPlayer(ctx context.Context, sessionID, playerID string, message models.WSMessage) error {
	h.mu.RLock()
	room, exists := h.rooms[sessionID]
	h.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	client, exists := room.clients[playerID]
	if !exists {
		return fmt.Errorf("player not in room")
	}

	select {
	case client.Send <- data:
		return nil
	default:
		go func(c *Client) {
			h.unregister <- c
		}(client)
		return nil
	}
}

func (h *WebSocketHub) IsPlayerRegistered(sessionID string, playerID string) bool {
	h.mu.RLock()
	room, exists := h.rooms[sessionID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.players[playerID]
}
