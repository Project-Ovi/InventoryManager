package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type Part struct {
	ID       primitive.ObjectID `bson:"_id"`
	Name     string             `bson:"part-name"`
	Tags     []string           `bson:"tags"`
	Location string             `bson:"location"`
	Qty      float64            `bson:"qty"`
}
