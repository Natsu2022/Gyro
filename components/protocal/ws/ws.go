package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"GOLANG_SERVER/components/db"
	schema "GOLANG_SERVER/components/schema"

	"github.com/gorilla/websocket"
)

// Handle WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow all connections
}

// Store all connected clients
var clients = make(map[*websocket.Conn]bool)

// Store all connected clients for storing data
var clientsStore = make(map[*websocket.Conn]bool)

var DeviceAdd = "test"

// Handle a WebSocket connection
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection to WebSocket:", err)
		return
	}

	// Register the client
	clients[conn] = true

	// * get message from client
	_, message, err := conn.ReadMessage()
	if err != nil {
		log.Println("Error reading message from client:", err)
		return
	}

	var req schema.GyroData
	if err := json.Unmarshal(message, &req); err != nil {
		log.Println("Error unmarshaling message:", err)
		return
	}
	// ! Not add checking data yet : waiting for next time.
	fmt.Println("Message from client:", string(req.DeviceAddress))

	// * change device address
	DeviceAdd = req.DeviceAddress

	// Log the number of clients
	fmt.Println("Number of clients:", len(clients))

	// Wait for a message from the client
	for {
		// Send the message to all clients
		for client := range clients {
			// * get data from database
			data, err := db.GetGyroDataByDeviceAddressLatest(DeviceAdd)
			if err != nil {
				// log.Println("Error getting data from database:", err)
				continue
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling data to JSON:", err)
				continue
			}
			if err := client.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Println("Error writing message to client:", err)
				log.Println("Closing client connection...")
				client.Close()
				delete(clients, client)
			}

			// * delay 1 second
			time.Sleep(1 * time.Second)
		}
	}
}

// Handle a WebSocket connection for storing data
func HandleStoreWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection to WebSocket:", err)
		return
	}

	// * disconnect client if there is already a client connected
	if len(clientsStore) > 0 {
		log.Println("Client already connected!")
		conn.Close()
		return
	}

	// Register the client
	clientsStore[conn] = true

	if len(clientsStore) == 1 {
		// Wait for a message from the client
		for {
			// Read the message from the client
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Disconnected from store client")
				delete(clientsStore, conn)
				break
			}

			// Store the data in the database
			var data schema.GyroData
			if err := json.Unmarshal(message, &data); err != nil {
				log.Println("Error unmarshaling message:", err)
				continue
			}
			if _, err := db.StoreGyroData(data); err != nil {
				log.Println("Error storing data in database:", err)
				continue
			}

			resmes := []byte(`{"message": "Data stored!"}`)

			// Send the message to all clients
			for client := range clientsStore {
				if err := client.WriteMessage(websocket.TextMessage, resmes); err != nil {
					log.Println("Error writing message to client:", err)
					log.Println("Closing client connection...")
					client.Close()
					delete(clientsStore, client)
				}
			}
		}
	}
}

// Handle a WebSocket connection for getting data
func handleWebSocketGetData(w http.ResponseWriter, r *http.Request) {
	// Upgrade the connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection to WebSocket:", err)
		return
	}

	// Register the client
	clients[conn] = true

	// Log the number of clients
	fmt.Println("Number of clients:", len(clients))

	// Wait for a message from the client
	for {
		// Send the message to all clients
		for client := range clients {
			// * get data from database
			data, err := db.GetGyroDataByDeviceAddressLatest(DeviceAdd)
			if err != nil {
				// log.Println("Error getting data from database:", err)
				continue
			}

			jsonData, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling data to JSON:", err)
				continue
			}
			if err := client.WriteMessage(websocket.TextMessage, jsonData); err != nil {
				log.Println("Error writing message to client:", err)
				log.Println("Closing client connection...")
				client.Close()
				delete(clients, client)
			}

			// * delay 1 second
			time.Sleep(1 * time.Second)
		}
	}
}
