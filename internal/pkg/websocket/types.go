package websocket

const (
	// Frame opcodes
	OpContinuation = 0x0
	OpText         = 0x1
	OpBinary       = 0x2
	OpClose        = 0x8
	OpPing         = 0x9
	OpPong         = 0xA
)

type Frame struct {
	Opcode  byte
	Payload []byte
	Masked  bool
}
