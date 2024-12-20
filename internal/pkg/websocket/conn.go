package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
)

type Conn struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("webserver doesn't support hijacking")
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, err
	}

	if err := performHandshake(bufrw, r); err != nil {
		conn.Close()
		return nil, err
	}

	return &Conn{
		conn:   conn,
		reader: bufio.NewReader(conn),
		writer: bufrw.Writer,
	}, nil
}

func performHandshake(bufrw *bufio.ReadWriter, r *http.Request) error {
	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return errors.New("missing Sec-WebSocket-Key")
	}

	// Generate accept key
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Send handshake response
	response := fmt.Sprintf(
		"HTTP/1.1 101 Switching Protocols\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Accept: %s\r\n\r\n",
		acceptKey,
	)

	if _, err := bufrw.WriteString(response); err != nil {
		return err
	}
	return bufrw.Flush()
}

func (c *Conn) ReadFrame() (Frame, error) {
	frame := Frame{}

	// Read first byte
	byte1, err := c.reader.ReadByte()
	if err != nil {
		return frame, err
	}

	fin := byte1&0x80 != 0
	if !fin {
		return frame, errors.New("fragmented frames not supported")
	}

	frame.Opcode = byte1 & 0x0F

	// Read second byte
	byte2, err := c.reader.ReadByte()
	if err != nil {
		return frame, err
	}

	frame.Masked = byte2&0x80 != 0
	length := byte2 & 0x7F

	// Read extended payload length
	var payloadLength uint64
	switch length {
	case 126:
		var len16 uint16
		if err := binary.Read(c.reader, binary.BigEndian, &len16); err != nil {
			return frame, err
		}
		payloadLength = uint64(len16)
	case 127:
		if err := binary.Read(c.reader, binary.BigEndian, &payloadLength); err != nil {
			return frame, err
		}
	default:
		payloadLength = uint64(length)
	}

	// Read masking key if present
	var maskKey []byte
	if frame.Masked {
		maskKey = make([]byte, 4)
		if _, err := io.ReadFull(c.reader, maskKey); err != nil {
			return frame, err
		}
	}

	// Read payload
	frame.Payload = make([]byte, payloadLength)
	if _, err := io.ReadFull(c.reader, frame.Payload); err != nil {
		return frame, err
	}

	// Unmask payload if necessary
	if frame.Masked {
		for i := range frame.Payload {
			frame.Payload[i] ^= maskKey[i%4]
		}
	}

	return frame, nil
}

func (c *Conn) WriteFrame(frame Frame) error {
	// Write first byte
	byte1 := uint8(0x80) | frame.Opcode
	if err := c.writer.WriteByte(byte1); err != nil {
		return err
	}

	// Write length
	length := len(frame.Payload)
	var byte2 uint8
	switch {
	case length <= 125:
		byte2 = uint8(length)
	case length <= 65535:
		byte2 = 126
	default:
		byte2 = 127
	}
	if err := c.writer.WriteByte(byte2); err != nil {
		return err
	}

	// Write extended length if necessary
	switch {
	case byte2 == 126:
		if err := binary.Write(c.writer, binary.BigEndian, uint16(length)); err != nil {
			return err
		}
	case byte2 == 127:
		if err := binary.Write(c.writer, binary.BigEndian, uint64(length)); err != nil {
			return err
		}
	}

	// Write payload
	if _, err := c.writer.Write(frame.Payload); err != nil {
		return err
	}

	return c.writer.Flush()
}

func (c *Conn) Close() error {
	closeFrame := Frame{
		Opcode:  OpClose,
		Payload: []byte{},
	}
	if err := c.WriteFrame(closeFrame); err != nil {
		return err
	}
	return c.conn.Close()
}
