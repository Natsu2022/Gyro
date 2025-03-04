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

// Handle MQTT connections and messages
func HandleMQTT() {
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
		// fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
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

	fmt.Println("MQTT client connected and subscribed to topic.")
}
