package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

type Message struct {
	Message string `json:"message"`
}

var clients = make(map[*websocket.Conn]bool) // Connected clients
var broadcast = make(chan Message)           // Broadcast channel
var mutex = &sync.Mutex{}

func Handle(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP request to a WebSocket connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}
	defer ws.Close()

	log.Println("New WebSocket client connected")

	// Continuous message handling loop
	for {
		// Read message from client
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break // Exit the loop if there's an error (e.g., client disconnected)
		}

		log.Printf("Received message: %s", message)

		// Handle "ping" message by responding with "pong"
		if string(message) == "ping" {
			err = ws.WriteMessage(messageType, []byte("pong"))
			log.Println("Sent pong")
			if err != nil {
				log.Printf("Error sending pong message: %v", err)
				break // Exit the loop if unable to send a message
			}
		}
	}

	log.Println("WebSocket client disconnected")
}

func readMessages(ws *websocket.Conn) {
	for {
		// Read message from client
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			mutex.Lock()
			delete(clients, ws)
			mutex.Unlock()
			break
		}

		log.Printf("Received message: %s", message)

		// If "ping" received, reply with "pong"
		if string(message) == "ping" {
			err = ws.WriteMessage(messageType, []byte("pong"))
			log.Println("Sent pong")
			if err != nil {
				log.Printf("Error sending pong message: %v", err)
				break
			}
		}

		// Broadcast the received message to all connected clients
		broadcast <- Message{Message: string(message)}
	}
}

func BroadcastMessages() {
	for {
		// Grab next message from broadcast channel
		msg := <-broadcast

		// Send to every client connected
		mutex.Lock()
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error broadcasting message: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}
