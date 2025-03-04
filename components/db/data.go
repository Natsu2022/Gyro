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
	"golang.org/x/crypto/bcrypt"
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

func GetGyroDataByDeviceAddressLatest(DeviceAddress string) ([]schema.GyroData, error) {
	if len(DeviceAddress) == 0 {
		return nil, errors.New("device address is empty")
	}
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_COLLECTION"))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var gyroData []schema.GyroData
	cursor, err := collection.Find(ctx, bson.M{strings.ToLower("deviceaddress"): DeviceAddress}, options.Find().SetSort(bson.D{{Key: strings.ToLower("timestamp"), Value: -1}}).SetLimit(50))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &gyroData); err != nil {
		return nil, err
	}

	if len(gyroData) == 0 {
		return nil, errors.New("no data found")
	}
	return gyroData, nil
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

	// Check if device already exists
	filter := bson.M{"deviceaddress": DeviceAddress}

	log.Println("Querying database with filter:", filter)

	// Check if device already exists
	var result struct {
		DeviceAddress string `bson:"deviceaddress"`
	}
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == nil { // Device already exists
		return false, errors.New("device already exists")
	} else if err != mongo.ErrNoDocuments {
		return false, err
	} else {
		_, err := collection.InsertOne(ctx, bson.M{"deviceaddress": DeviceAddress})
		if err != nil {
			return false, err
		}

		log.Println("Device registered successfully.")
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

// get data from collection data in mongoDB by device address
func GetDataByDeviceAddress(deviceAddress string) ([]schema.GyroData, error) {
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_COLLECTION")) // Get collection data
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)                        // Create a context with timeout
	defer cancel()                                                                                  // Defer cancel the context
	cursor, err := collection.Find(ctx, bson.M{"deviceaddress": deviceAddress})                     // Find data by device address
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

// Store Email and Password to mongoDB collection user
func StoreUser(user schema.User) (bool, error) {
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_USERCOLLECTION")) // Get collection user
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)                            // Create a context with timeout
	defer cancel()                                                                                      // Defer cancel the context

	// Check if user already exists
	filter := bson.M{"email": user.Email}
	var result schema.User
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err == nil {
		return false, errors.New("user already exists")
	} else if err != mongo.ErrNoDocuments {
		return false, err
	} else {
		_, err := collection.InsertOne(ctx, user) // Insert user to collection
		if err != nil {
			return false, err
		}
	}

	return true, nil
}

// Login checks if the user exists and returns the user object and an error
func Login(email string, password string) (schema.User, error) {
	collection = client.Database(env.GetEnv("MONGO_DB")).Collection(env.GetEnv("MONGO_USERCOLLECTION")) // Get collection user
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)                            // Create a context with timeout
	defer cancel()                                                                                      // Defer cancel the context

	// Check if user exists
	filter := bson.M{"email": email}
	var result schema.User
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return schema.User{}, errors.New("user not found")
		}
		return schema.User{}, err
	}

	log.Println("User found:", email)

	// In Database Password is hashed
	// Compare the stored password with the input password
	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(password))
	if err != nil {
		return schema.User{}, errors.New("invalid password")
	}

	return result, nil
}
