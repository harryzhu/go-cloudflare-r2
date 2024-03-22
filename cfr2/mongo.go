package cfr2

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	//"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Item struct {
	ID       string            `json:"_id" bson:"_id"`
	Size     int64             `json:"size" bson:"size"`
	Owner    string            `json:"owner" bson:"owner"`
	IsPublic bool              `json:"is_public" bson:"is_public"`
	Meta     map[string]string `json:"meta" bson:"meta"`
}

func NewItem(k string) *Item {
	return &Item{
		ID:       k,
		IsPublic: false,
	}
}

func (m *Item) SaveTo(coll string) error {
	if m.Owner == "" || coll == "" {
		err := errors.New("owner/collection cannot be empty")
		PrintlnError(err)
		return err
	}
	var mgcoll *mongo.Collection
	mgcoll = mgdb.Collection(coll)

	opts := options.Replace().SetUpsert(true)
	filter := bson.D{{"_id", m.ID}}
	replacement := m
	result, err := mgcoll.ReplaceOne(context.TODO(), filter, replacement, opts)

	// result, err := mgcoll.InsertOne(
	// 	context.TODO(),
	// 	m,
	// )
	if err != nil {
		PrintlnError(err)
		return err
	}
	if result.MatchedCount != 0 {
		PrintlnDebug("matched and replaced an existing document")
	}
	if result.UpsertedCount != 0 {
		PrintlnDebug("inserted a new document with ID %v\n", result.UpsertedID)
	}
	//insertID := fmt.Sprintf("%v", result.InsertedID)
	//PrintlnDebug(insertID)
	return nil
}

func (m *Item) DeleteFrom(coll string) string {
	if m.ID == "" || coll == "" {
		PrintlnDebug("ID/collection cannot be empty")
		return ""
	}
	var mgcoll *mongo.Collection
	mgcoll = mgdb.Collection(coll)
	result, err := mgcoll.DeleteOne(
		context.TODO(),
		bson.D{{"_id", m.ID}},
	)
	PrintlnDebug(err)
	deleteCount := fmt.Sprintf("%d", result.DeletedCount)
	PrintlnDebug(deleteCount)
	return deleteCount
}

func (m *Item) GetFrom(coll string) {
	var mgcoll *mongo.Collection
	var result bson.M
	opts := options.FindOne().SetSort(bson.D{{"owner", 1}})

	mgcoll = mgdb.Collection(coll)

	err := mgcoll.FindOne(
		context.TODO(),
		bson.D{{"_id", m.ID}},
		opts,
	).Decode(&result)

	if err != nil {
		PrintlnError(err)
	}

	fmt.Printf("found document %v", result)
}
