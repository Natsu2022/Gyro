package mosquitto

import (
	"encoding/json"
	"fmt"
	"log"

	"GOLANG_SERVER/components/db"
	"GOLANG_SERVER/components/env"
	schema "GOLANG_SERVER/components/schema"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var client mqtt.Client
var subscribed bool = false

// Initialize and connect MQTT client
func InitMQTT() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(env.GetEnv("MQTT_BROKER"))
	opts.SetClientID("go_mqtt_client")
	opts.SetUsername(env.GetEnv("MQTT_USERNAME"))
	opts.SetPassword(env.GetEnv("MQTT_PASSWORD"))

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	fmt.Println("MQTT client connected")
}

// Subscribe to topic and handle incoming messages
func Subscribe() {
	if subscribed {
		log.Println("Already subscribed, skipping...")
		return
	}
	subscribed = true
	var data schema.GyroData
	var topic = "sub_data"
	if token := client.Subscribe(topic, 1, func(client mqtt.Client, msg mqtt.Message) {

		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			log.Println("Error unmarshaling message:", err)
			return
		}
		log.Printf("Received data from topic: %s\n", msg.Topic())

		if _, err := db.StoreGyroData(data); err != nil {
			log.Println("Error storing data in database:", err)
		}
	}); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	fmt.Printf("Subscribed to topic: %s\n", topic)
}

// Start MQTT handling
func HandleMQTT() {
	InitMQTT()
	go Subscribe() // Start subscription in a separate goroutine
}
