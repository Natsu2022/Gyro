package db

import (
	"context"
	"errors"
	"fmt"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Defer cancel the context
	currentTime := time.Now()
	data.TimeStamp = currentTime.Format(time.RFC3339)
	_, err := collection.InsertOne(ctx, data)
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
	var Data []schema.GyroData
	if err = cursor.All(ctx, &Data); err != nil {
		return nil, err
	}
	return Data, nil
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

	var Data []schema.GyroData
	if err = cursor.All(ctx, &Data); err != nil {
		return nil, err
	}

	if len(Data) == 0 {
		return nil, errors.New("no data found")
	}

	return Data, nil
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
