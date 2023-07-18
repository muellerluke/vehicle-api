package configs

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var sessions = session.New(session.Config{
	Expiration: 7 * 24 * time.Hour,
})

func GetSession() *session.Store {
	return sessions
}

func ConnectRedis() *redis.Client {
	opt, err := redis.ParseURL(RetrieveEnv("REDIS_URI"))

	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(opt)

	return rdb
}

var Redis *redis.Client = ConnectRedis()

func ConnectDB() *mongo.Client {
	client, err := mongo.NewClient(options.Client().ApplyURI(RetrieveEnv("MONGO_URI")))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer cancel()

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")

	var userCollection *mongo.Collection = GetCollection(client, "users")

	indexName, err := userCollection.Indexes().CreateOne(
		context.Background(),
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)

	log.Println(indexName)

	if err != nil {
		log.Fatal(err)
	}

	return client
}

// Client instance
var DB *mongo.Client = ConnectDB()

// getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("data").Collection(collectionName)
	return collection
}
