package broadcast

import (
	"log"

	"wordwizardry/internal/pkg/websocket"
)

type Client struct {
	Hub       *WebSocketHub
	Conn      *websocket.Conn
	SessionID string
	PlayerID  string
	Send      chan []byte
}

func (c *Client) writePump() {
	defer func() {
		c.Hub.unregister <- c
	}()

	for message := range c.Send {
		err := c.Conn.WriteFrame(websocket.Frame{
			Opcode:  websocket.OpText,
			Payload: message,
		})
		if err != nil {
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		frame, err := c.Conn.ReadFrame()
		if err != nil {
			return
		}

		switch frame.Opcode {
		case websocket.OpClose:
			return
		case websocket.OpPing:
			pongFrame := websocket.Frame{
				Opcode:  websocket.OpPong,
				Payload: frame.Payload,
			}
			if err := c.Conn.WriteFrame(pongFrame); err != nil {
				return
			}
		case websocket.OpText:
			// Handle text messages if needed
			log.Println("Received message:", string(frame.Payload))
		}
	}
}
