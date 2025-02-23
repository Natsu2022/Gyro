package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	env "GOLANG_SERVER/components/env"
	schema "GOLANG_SERVER/components/schema"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// * mongo db connection
var client *mongo.Client
var collection *mongo.Collection

// * Connect to mongo db
func Connect() (bool, error) {
	clientOptions := options.Client().ApplyURI(env.GetEnv("MONGO_URI"))
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		fmt.Println("Can't connect to mongo db:", err)
		return false, err
	}

	// Check the connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		fmt.Println("Can't ping mongo db:", err)
		return false, err
	}

	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_COLLECTION"))
	return true, nil
}

// * store data to mongo db and use upper camel case for function name
func StoreGyroData(data schema.GyroData) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Create a context with timeout
	defer cancel()                                                           // Defer cancel the context

	// load Bangkok timezone
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return false, err
	}

	currentTime := time.Now().In(loc) // Get current time

	data.DateTime = currentTime.Format(time.RFC3339) // Set timestamp to current time
	data.TimeStamp = currentTime.UnixMilli()         // Set timestamp to current time
	_, err = collection.InsertOne(ctx, data)
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetGyroData() ([]schema.GyroData, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var gyroData []schema.GyroData
	if err = cursor.All(ctx, &gyroData); err != nil {
		return nil, err
	}
	return gyroData, nil
}

func GetGyroDataByDeviceAddress(DeviceAddress string) ([]schema.GyroData, error) {
	if len(DeviceAddress) == 0 {
		return []schema.GyroData{}, errors.New("device address is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{strings.ToLower("DeviceAddress"): DeviceAddress})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var gyroData []schema.GyroData
	if err = cursor.All(ctx, &gyroData); err != nil {
		return nil, err
	}
	return gyroData, nil
}

func GetGyroDataByDeviceAddressLatest(DeviceAddress string) (schema.GyroData, error) {
	if len(DeviceAddress) == 0 {
		return schema.GyroData{}, errors.New("device address is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var Data schema.GyroData
	err := collection.FindOne(ctx, bson.M{strings.ToLower("deviceaddress"): DeviceAddress}, options.FindOne().SetSort(bson.D{{Key: strings.ToLower("timestamp"), Value: -1}})).Decode(&Data)
	if err != nil {
		return Data, err
	}

	if len(Data.DeviceAddress) == 0 {
		return Data, errors.New("no data found")
	}
	return Data, nil
}

func CleanData() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		return false, err
	}
	return true, nil
}

func RegisterDevice(DeviceAddress string) (bool, error) {
	if len(DeviceAddress) == 0 {
		return false, errors.New("device address is empty")
	}
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_DEVICECOLLECTION"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := collection.InsertOne(ctx, bson.M{"deviceaddress": DeviceAddress})
	if err != nil {
		return false, err
	}
	return true, nil
}

func GetDeviceAddress() ([]string, error) {
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_DEVICECOLLECTION"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Create a context with timeout
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var deviceAddresses []string
	for cursor.Next(ctx) {
		var result struct {
			DeviceAddress string `bson:"deviceaddress"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		deviceAddresses = append(deviceAddresses, result.DeviceAddress)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return deviceAddresses, nil
}

func GetDeviceAddressByDeviceAddress(deviceAddress string) ([]string, error) {
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_DEVICECOLLECTION"))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Create a context with timeout
	defer cancel()

	filter := bson.M{"deviceaddress": deviceAddress}
	log.Println("Querying database with filter:", filter)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var deviceAddresses []string
	for cursor.Next(ctx) {
		var result struct {
			DeviceAddress string `bson:"deviceaddress"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		deviceAddresses = append(deviceAddresses, result.DeviceAddress)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	log.Println("Found device addresses:", deviceAddresses)
	return deviceAddresses, nil
}
