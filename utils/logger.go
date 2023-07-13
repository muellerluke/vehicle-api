package utils

import (
	"context"
	"time"
	"vehicle-api/configs"
	"vehicle-api/models"

	//"github.com/stripe/stripe-go/v74"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var callCollection *mongo.Collection = configs.GetCollection(configs.DB, "calls")
var userCollection *mongo.Collection = configs.GetCollection(configs.DB, "users")

func LogCall(key models.Key, originalURL string, routeName string) {
	// first log call in the db
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// first find the user that called the api
	newCall := models.Call{
		User:       key.User,
		Key:        key.ID,
		RequestURL: originalURL,
		CreatedAt:  time.Now().Unix(),
	}

	_, err := callCollection.InsertOne(ctx, newCall)

	if err != nil {
		panic(err)
	}

	// then update the user's call count
	if routeName != "years" && routeName != "makes" && routeName != "models" && routeName != "trims" {
		//add usage record for subscription
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user models.User

		err := userCollection.FindOne(ctx, bson.M{"_id": key.User}).Decode(&user)

		if err != nil {
			if err == mongo.ErrNoDocuments {
				//no user found
				panic(err)
			}
		}

		/*stripe.Key = configs.RetrieveEnv("STRIPE_SECRET_KEY")

		params := &stripe.UsageRecordParams{
			Quantity: stripe.Int64(1),
			SubscriptionItem: stripe.String(user.SubscriptionItemID),
			Timestamp: stripe.Int64(time.Now().Unix()),
		}
		ur, _ := usagerecord.New(params)*/
	}
}
