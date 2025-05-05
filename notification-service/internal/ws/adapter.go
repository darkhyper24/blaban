package ws

import (
	fiberwsocket "github.com/gofiber/websocket/v2"
)

// WebSocketAdapter adapts a Fiber websocket connection to work with our hub
// that expects Gorilla websocket connections
type WebSocketAdapter struct {
	conn *fiberwsocket.Conn
}

// Close implements the Close method from gorilla websocket.Conn
func (a *WebSocketAdapter) Close() error {
	return a.conn.Close()
}

// WriteMessage implements the WriteMessage method from gorilla websocket.Conn
func (a *WebSocketAdapter) WriteMessage(messageType int, data []byte) error {
	return a.conn.WriteMessage(messageType, data)
}
