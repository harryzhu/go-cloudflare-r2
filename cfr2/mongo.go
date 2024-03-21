package cfr2

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

//"go.mongodb.org/mongo-driver/mongo/options"

type Item struct {
	ID       string `json:"_id" bson:"_id"`
	FileSize int64  `json:"filesize" bson:"filesize"`
	Owner    string `json:"owner" bson:"owner"`
}

func NewItem(s string) *Item {
	m := &Item{}
	m.ID = s
	m.FileSize = 0
	m.Owner = ""
	return m
}

func (m *Item) With(k, v string) *Item {

	return m
}

func (m *Item) Save(c string) *Item {

	var mgcoll *mongo.Collection
	mgcoll = mgdb.Collection(c)
	result, err := mgcoll.InsertOne(
		context.TODO(),
		m,
	)
	PrintlnDebug(err)

	PrintlnDebug(result.InsertedID)
	return m
}
