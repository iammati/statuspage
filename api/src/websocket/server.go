package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"iammati/statuspage/handlers/k8s"
	"iammati/statuspage/utils"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins
}

type Message struct {
	API       string `json:"api"`
	Namespace string `json:"namespace,omitempty"`
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
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break // Exit the loop if there's an error (e.g., client disconnected)
		}

		// Parse the JSON message
		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			utils.SendMessage(ws, `{"error": "Invalid JSON format"}`)
			continue // Skip further processing for this message
		}

		switch msg.API {
		case "k8s/namespaces/list":
			response := k8s.ListNamespaces()
			err := utils.SendMessage(ws, response)
			if err != nil {
				log.Printf("Error sending response: %v", err)
			}
		case "k8s/pods/list":
			if msg.Namespace == "" {
				log.Println("Missing 'namespace' in message payload")
				utils.SendMessage(ws, `{"error": "Missing 'namespace' in message payload"}`)
				continue // Skip further processing for this message
			}

			response := k8s.ListPods(msg.Namespace)
			err := utils.SendMessage(ws, response)
			if err != nil {
				log.Printf("Error sending response: %v", err)
			}
		default:
			log.Printf("Unknown API command: %s", msg.API)
			utils.SendMessage(ws, `{"error": "Unknown API command"}`)
		}
	}

	log.Println("WebSocket client disconnected")
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
