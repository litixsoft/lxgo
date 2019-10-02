package main

//import (
//	"context"
//	"fmt"
//	"go.mongodb.org/mongo-driver/bson"
//
//	lxDb "github.com/litixsoft/lxgo/db"
//	"log"
//)
//
//func main() {
//	fmt.Println("Hello Playground")
//
//	insertTest(&bson.M{"firstname": "test", "lastname": "test test"})
//}
//
//
//func insertTest(data interface{}) {
//	client, err := lxDb.GetMongoDbClient("mongodb://127.0.0.1")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	insert := bson.M{
//		"Name": "Foo",
//		"data": data,
//	}
//
//	collection := client.Database("test").Collection("users")
//	res, err := collection.InsertOne(context.TODO(), insert)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	log.Println(res)
//}
