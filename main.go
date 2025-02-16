package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"GOLANG_SERVER/components/db"
	"GOLANG_SERVER/components/env"
	schema "GOLANG_SERVER/components/schema"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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
func handleStoreWebSocket(w http.ResponseWriter, r *http.Request) {
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

// Handle a REST API request
func handleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Hello from REST API!"}`)
}

// Handle a request for the schema
func handleGetAllData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the data from the database
	data, err := db.GetGyroData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the data into JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handle a request to store data
func handleStore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Decode the request body into a GyroData struct
	var data schema.GyroData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// * print the data
	fmt.Printf("Data: %+v\n", data)

	// Store the data in the database
	db.StoreGyroData(data)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Data stored!"}`)
}

// * get data use param
func handleGetAllDataByDeviceAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the device address from the URL
	deviceAddress := r.URL.Path[len("/data/"):]
	fmt.Println("Device Address:", deviceAddress)

	// Get the data from the database
	data, err := db.GetGyroDataByDeviceAddress(deviceAddress)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the data into JSON
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// * get latest data
func handleGetLatestData(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		w.Header().Set("Content-Type", "application/json")

		// * get data from request
		var req schema.GyroData
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		// Get the data from the database
		data, err := db.GetGyroDataByDeviceAddressLatest(req.DeviceAddress)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Encode the data into JSON
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleCleanData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the password from client
	if r.Method == "POST" {
		var req schema.PasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// * check password
		if req.Password == req.CFP {
			if req.Password == env.GetEnv("PASSWORD") {
				// * clean data
				if _, err := db.CleanData(); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusOK)
				fmt.Fprintf(w, `{"message": "Data cleaned!"}`)
			} else {
				http.Error(w, "Invalid password", http.StatusUnauthorized)
			}
		} else {
			http.Error(w, "Password doesn't match", http.StatusUnauthorized)
		}

	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handle MQTT connections and messages
func handleMQTT() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(env.GetEnv("MQTT_BROKER"))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(env.GetEnv("MQTT_USERNAME"))
	opts.SetPassword(env.GetEnv("MQTT_PASSWORD"))

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	if token := client.Subscribe("sample", 1, func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
		// Process the message and store it in the database
		var data schema.GyroData
		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			log.Println("Error unmarshaling message:", err)
			return
		}
		if _, err := db.StoreGyroData(data); err != nil {
			log.Println("Error storing data in database:", err)
		}

	}); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	fmt.Println("MQTT client connected and subscribed to topic")
}

// Main function
func main() {
	// TODO: Load environment variables
	if _, err := env.LoadEnv(); err != nil {
		log.Fatal("Error loading environment variables:", err)
		return
	}

	// TODO: Get the port from the environment variables
	port, err := strconv.Atoi(env.GetEnv("PORT"))
	if err != nil {
		log.Fatal("Invalid port number:", err)
		return
	}

	// TODO: Connect to the database
	if _, err := db.Connect(); err == nil {
		// * Welcome message
		fmt.Println("Message:", env.GetEnv("MESSAGE"))

		// TODO: REST API route
		http.HandleFunc("/api", handleAPI)
		http.HandleFunc("/data", handleGetAllData)
		http.HandleFunc("/store", handleStore)
		http.HandleFunc("/latest", handleGetLatestData)

		// * get data use param
		http.HandleFunc("/data/", handleGetAllDataByDeviceAddress)

		// TODO: WebSocket route
		http.HandleFunc("/ws", handleWebSocket)
		http.HandleFunc("/ws/store", handleStoreWebSocket)

		// ! clear data and change device address
		http.HandleFunc("/clean", handleCleanData)

		// TODO: Start the server in a goroutine
		go func() {
			fmt.Println("Server started at Gyro Server")
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
				log.Fatal("Error starting server:", err)
			}
		}()

		// Start MQTT client
		go handleMQTT()

		// TODO: Wait for 'q' or 'Q' to stop the server
		var input string
		for {
			fmt.Scanln(&input)
			if input == "q" || input == "Q" {
				fmt.Println("Server stopping...")
				break
			}
		}
	} else {
		fmt.Println("Error connecting to database something went wrong!!")
		return
	}
}
