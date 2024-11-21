package utils

import (
	"log"

	"github.com/gorilla/websocket"
)

func SendMessage(ws *websocket.Conn, message string) error {
	err := ws.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		log.Printf("Error sending message to WebSocket client: %v", err)
		return err
	}
	return nil
}
