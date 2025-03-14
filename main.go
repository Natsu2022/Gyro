package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"GOLANG_SERVER/components/db"
	"GOLANG_SERVER/components/env"
	"GOLANG_SERVER/components/protocal/mosquitto"
	"GOLANG_SERVER/components/protocal/rest"
	"GOLANG_SERVER/components/protocal/ws"
	"GOLANG_SERVER/components/user"
)

// Main function
func main() {
	// Load environment variables
	if err := env.LoadEnv(); err != nil {
		log.Fatal("Error loading environment variables:", err)
		return
	}

	// Get the port from the environment variables
	port, err := strconv.Atoi(env.GetEnv("PORT"))
	if err != nil {
		log.Fatal("Invalid port number:", err)
		return
	}

	// Connect to the database
	if _, err := db.Connect(); err == nil {
		// Welcome message
		fmt.Println("Message:", env.GetEnv("MESSAGE"))

		//TODO REST API route
		http.HandleFunc("/api", rest.HandleAPI)
		http.HandleFunc("/data", rest.HandleGetAllData)
		http.HandleFunc("/store", rest.HandleStore)
		http.HandleFunc("/latest", rest.HandleGetLatestData)
		http.HandleFunc("/clean", rest.HandleCleanData)

		http.HandleFunc("/registerdevice", rest.HandleRegisterDevice)                         //*DONE Register device
		http.HandleFunc("/deviceaddresses", rest.HandleGetDeviceAddress)                      //*DONE Get device address
		http.HandleFunc("/checkdeviceaddresses/", rest.HandleGetDeviceAddressByDeviceAddress) //*DONE Get device address by device address
		http.HandleFunc("/data/", rest.HandleGetAllDataByDeviceAddress)                       //*DONE Get data use param
		http.HandleFunc("/register", user.Register)                                           //*DONE Register user by Enail and Password
		http.HandleFunc("/login", user.Login)                                                 //*DONE login user by Email and Password
		http.HandleFunc("/sendotp", user.SendOTP)                                             //*DONE Send OTP to Email

		// TODO: WebSocket route
		http.HandleFunc("/ws", ws.HandleWebSocket)           //*DONE Handle WebSocket connection
		http.HandleFunc("/storews", ws.HandleStoreWebSocket) //*DONE Store data from websocket

		// TODO: Start the server in a goroutine
		go func() {
			fmt.Println("Server started at Gyro Server.")
			if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
				log.Fatal("Error starting server:", err)
			}
		}()

		//? Start MQTT client
		go mosquitto.HandleMQTT()

		// Wait for 'q' or 'Q' to stop the server
		var input string
		for {
			fmt.Scanln(&input)
			if input == "q" || input == "Q" {
				fmt.Println("Server stopping...")
				break // Stop the server
			}
		}
	} else {
		fmt.Println("Error connecting to database something went wrong!!")
		return
	}
}
