package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"GOLANG_SERVER/components/db"
	"GOLANG_SERVER/components/env"
	schema "GOLANG_SERVER/components/schema"

	"go.mongodb.org/mongo-driver/mongo"
)

func HandleRegisterDevice(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON

	// Get the device address from the json request
	deviceAddress := r.URL.Query().Get("deviceAddress")
	if deviceAddress == "" {
		http.Error(w, "Device address not found", http.StatusBadRequest)
		return
	}

	// display device address
	log.Println("Device Address:", deviceAddress)

	// store device address to database
	if _, err := db.RegisterDevice(deviceAddress); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// send device address .json to client
	response := map[string]string{"deviceAddress": deviceAddress}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, `{"message": "Device registered!"}`)
}

func HandleGetDeviceAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON

	// * get device address from database
	deviceAddresses, err := db.GetDeviceAddress()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// * send device addresses .json to client
	response := map[string][]string{"deviceAddresses": deviceAddresses}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HandleGetDeviceAddressByDeviceAddress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON

	// Get the device address from the URL
	deviceAddress := r.URL.Path[len("/checkdeviceaddresses/"):]
	log.Println("Received request for device address:", deviceAddress)

	// Get the data from the database
	deviceAddresses, err := db.GetDeviceAddressByDeviceAddress(deviceAddress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "No documents found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Check if the result is empty
	if len(deviceAddresses) == 0 {
		log.Println("No device addresses found for:", deviceAddress)
		http.Error(w, "No device addresses found", http.StatusNotFound)
		return
	}

	// Encode the data into JSON
	if err := json.NewEncoder(w).Encode(deviceAddresses); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Successfully retrieved device addresses for:", deviceAddress)
}

// Handle a REST API request
func HandleAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Hello from REST API!"}`)
}

// Handle a request for the schema
func HandleGetAllData(w http.ResponseWriter, r *http.Request) {
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
func HandleStore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Decode the request body into a GyroData struct
	var data schema.GyroData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// * print the data
	fmt.Printf("Data: %+v\n", data) // Print the data

	// Store the data in the database
	db.StoreGyroData(data)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Data stored!"}`)
}

// * get data use param
func HandleGetAllDataByDeviceAddress(w http.ResponseWriter, r *http.Request) {
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
func HandleGetLatestData(w http.ResponseWriter, r *http.Request) {
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

func HandleCleanData(w http.ResponseWriter, r *http.Request) {
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
