package main

import (
	"log"

	"Golang_Server/components/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var clients = make(map[*websocket.Conn]bool)      // Store connected WebSocket clients
var clientsStore = make(map[*websocket.Conn]bool) // Store clients that send data to be saved
var DeviceAdd = "test"                            // Variable to store the Device Address sent by the client

func handleWebSocket(c *websocket.Conn) {
	defer c.Close() // Close the connection when done

	// Register new client
	clients[c] = true

	for {
		// Read message from the client
		_, msg, err := c.ReadMessage()
		if err != nil {
			delete(clients, c)
			log.Println(err)
			return
		}

		// Check the data sent by the client
		if string(msg) == "Save" {
			clientsStore[c] = true
			// Save data to the database
			err := db.SaveData(map[string]string{"device_address": DeviceAdd})
			if err != nil {
				log.Println("Failed to save data:", err)
			}
		} else {
			DeviceAdd = string(msg)
		}

		// Log the received data
		log.Println("Device Address: ", DeviceAdd)

		// Send data back to the client
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(DeviceAdd))
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func main() {
	app := fiber.New()

	// Connect to the database
	db.ConnectDB()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// WebSocket endpoint
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		handleWebSocket(c)
	}))

	// Get new timestamp
	app.Get("/timestamp", func(c *fiber.Ctx) error {
		var timestamp string
		// Mock implementation: replace with actual database logic
		// Assume the date is saved successfully
		timestamp = "2021-10-17 12:00:01"
		return c.SendString(timestamp)
	})

	log.Fatal(app.Listen(":3000"))
}
