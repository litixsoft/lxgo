package lxDb_test

import (
	"context"
	"encoding/json"
	"fmt"
	lxDb "github.com/litixsoft/lxgo/db"
	lxErrors "github.com/litixsoft/lxgo/errors"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"testing"
	"time"
)

//
//import (
//	"encoding/json"
//	"github.com/globalsign/mgo"
//	"github.com/globalsign/mgo/bson"
//	"github.com/litixsoft/lxgo/db"
//	"github.com/stretchr/testify/assert"
//	"io/ioutil"
//	"log"
//	"os"
//	"reflect"
//	"sort"
//	"testing"
//	"time"
//)
//
// TestUser, struct for test users
//type TestUser struct {
//	Id       bson.ObjectId `json:"id" bson:"_id"`
//	Name     string        `json:"name" bson:"name"`
//	Gender   string        `json:"gender" bson:"gender"`
//	Email    string        `json:"email" bson:"email"`
//	IsActive bool          `json:"is_active" bson:"is_active"`
//}

type TestUser struct {
	Id       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Gender   string             `json:"gender,omitempty" bson:"gender,omitempty"`
	Email    string             `json:"email" bson:"email"`
	IsActive bool               `json:"is_active,omitempty" bson:"is_active"`
}

const (
	TestDbName     = "lxgo_test"
	TestCollection = "users"
	FixturesPath   = "fixtures/users.test.json"
)

var (
	dbHost string
)

func init() {
	// Check DbHost environment
	dbHost = os.Getenv("DB_HOST")

	// When not defined set default host
	if dbHost == "" {
		dbHost = "mongodb://127.0.0.1:27017"
	}

	log.Println("DB_HOST:", dbHost)
}

//// setupData, create the test data and prepare the database
//func setupData(conn *mgo.Session) []TestUser {
//	// Delete collection if exists
//	err := conn.DB(TestDbName).C(TestCollection).DropCollection()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Setup indexes
//	indexes := []mgo.Index{
//		{Key: []string{"name"}},
//		{Key: []string{"email"}, Unique: true},
//	}
//
//	// Ensure indexes
//	col := conn.DB(TestDbName).C(TestCollection)
//	for _, i := range indexes {
//		if err := col.EnsureIndex(i); err != nil {
//			log.Fatal(err)
//		}
//	}
//
//	// Load test data from json file
//	raw, err := ioutil.ReadFile(FixturesPath)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Convert
//	var users []TestUser
//	if err := json.Unmarshal(raw, &users); err != nil {
//		log.Fatal(err)
//	}
//
//	// Make Test users map and insert test data in db
//	for i := 0; i < len(users); i++ {
//		users[i].Id = bson.NewObjectId()
//
//		// Insert user
//		if err := conn.DB(TestDbName).C(TestCollection).Insert(users[i]); err != nil {
//			log.Fatal(err)
//		}
//	}
//
//	// Sort test users with id for compare
//	sort.Slice(users[:], func(i, j int) bool {
//		return users[i].Id < users[j].Id
//	})
//
//	// Return mongo connection
//	return users
//}
//
// setupData, create the test data and prepare the database
func setupDataNew(db *mongo.Database) []TestUser {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Delete collection if exists
	collection := db.Collection(TestCollection)
	if err := collection.Drop(ctx); err != nil {
		log.Fatal(err)
	}

	// Load test data from json file
	raw, err := ioutil.ReadFile(FixturesPath)
	if err != nil {
		log.Fatal(err)
	}

	// Convert
	var users []TestUser
	if err := json.Unmarshal(raw, &users); err != nil {
		log.Fatal(err)
	}

	// Make Test users map and insert test data in db
	for i := 0; i < len(users); i++ {
		users[i].Id = primitive.NewObjectID()

		// Insert user
		_, err := collection.InsertOne(ctx, users[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	// Sort test users with id for compare
	sort.Slice(users[:], func(i, j int) bool {
		return users[i].Name < users[j].Name
	})

	// Return mongo connection
	return users
}

//
//func TestGetMongoDbClient(t *testing.T) {
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	assert.NoError(t, err)
//
//	t.Run("correct type", func(t *testing.T) {
//		db := client.Database(TestDbName)
//		chkT := reflect.TypeOf(db)
//		assert.Equal(t, "*mongo.Database", chkT.String())
//	})
//	//t.Run("test query", func(t *testing.T) {
//	//	db := client.Database(TestDbName)
//	//
//	//	// setup data
//	//	testUsers := setupData(conn)
//	//
//	//
//	//})
//}
//
//func TestNewMongoDb(t *testing.T) {
//	conn := lxDb.GetMongoDbConnection(dbHost)
//	defer conn.Close()
//
//	t.Run("correct type", func(t *testing.T) {
//		db := lxDb.NewMongoDb(conn, TestDbName, TestCollection)
//		chkT := reflect.TypeOf(db)
//		assert.Equal(t, "*lxDb.MongoDb", chkT.String())
//	})
//
//	t.Run("test query", func(t *testing.T) {
//		db := lxDb.NewMongoDb(conn, TestDbName, TestCollection)
//
//		// setup data
//		testUsers := setupData(conn)
//
//		var result []TestUser
//		assert.NoError(t, db.Conn.DB(db.Name).C(db.Collection).Find(nil).All(&result))
//
//		// Sort result for compare
//		sort.Slice(result[:], func(i, j int) bool {
//			return result[i].Id < result[j].Id
//		})
//
//		assert.Equal(t, testUsers, result)
//	})
//}
//
//func TestMongoDb_Setup(t *testing.T) {
//	conn := lxDb.GetMongoDbConnection(dbHost)
//	defer conn.Close()
//
//	// Delete collection if exists
//	err := conn.DB(TestDbName).C(TestCollection).DropCollection()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	db := lxDb.NewMongoDb(conn, TestDbName, TestCollection)
//
//	// setup indexes
//	assert.NoError(t, db.Setup([]mgo.Index{
//		{Key: []string{"name"}},
//		{Key: []string{"email"}, Unique: true},
//	}))
//
//	// index should be correct set
//	idx, err := db.Conn.DB(db.Name).C(db.Collection).Indexes()
//	assert.NoError(t, err)
//
//	assert.Equal(t, "email_1", idx[1].Name)
//	assert.True(t, idx[1].Unique)
//	assert.Equal(t, "name_1", idx[2].Name)
//}
//
/////// ############# ############# /////
// TestLogger
type TestAuditLogger struct{}

func (l *TestAuditLogger) Insert(authUser interface{}, collection string, data interface{}) error {
	log.Printf("authUser: %v\n", authUser)
	log.Printf("collection: %s\n", collection)
	log.Printf("data: %v\n", data)

	return nil
}

func TestGetMongoDbClient(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)
	its.IsType(&mongo.Client{}, client)
}

func TestMongoDbBaseRepo_InsertOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	lxErrors.HandlePanicErr(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Drop for test
	its.NoError(collection.Drop(context.Background()))

	testUser := TestUser{
		Name:  "TestName",
		Email: "test@test.de",
	}

	// Test the base repo
	auditBaseRepo := &TestAuditLogger{}
	base := lxDb.NewMongoBaseRepo(collection, auditBaseRepo)
	au := lxDb.AuthAudit{User: bson.M{"name": "Timo Liebetrau"}}

	// Channel for close
	done := make(chan bool)
	res, err := base.InsertOne(testUser, &au, done)

	// Wait for close channel
	<-done

	its.NoError(err)

	// Check insert result id
	var checkUser TestUser
	filter := bson.D{{"_id", res.(primitive.ObjectID)}}
	its.NoError(base.FindOne(filter, &checkUser))

	its.Equal(testUser.Name, checkUser.Name)
	its.Equal(testUser.Email, checkUser.Email)
	its.Equal(testUser.IsActive, checkUser.IsActive)
}

func TestMongoDbBaseRepo_InsertMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	lxErrors.HandlePanicErr(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Drop for test
	its.NoError(collection.Drop(context.Background()))

	testUsers := make([]interface{}, 0)
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("TestName%d", i)
		email := fmt.Sprintf("test%d@test.de", i)
		testUsers = append(testUsers, TestUser{Name: name, Email: email})
	}

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(collection)
	res, err := base.InsertMany(testUsers)
	its.NoError(err)
	its.Equal(5, len(res))

	// Find and compare
	var checkUsers []TestUser
	filter := bson.D{}
	err = base.Find(filter, &checkUsers)
	its.NoError(err)

	// Check users
	for i, v := range checkUsers {
		its.Equal(testUsers[i].(TestUser).Name, v.Name)
		its.Equal(testUsers[i].(TestUser).Email, v.Email)
	}

	// Check ids
	for _, id := range res {
		var res TestUser
		its.NoError(base.FindOne(bson.D{{"_id", id}}, &res))
		its.Equal(id, res.Id)
	}
}

func TestMongoDbBaseRepo_CountDocuments(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	setupDataNew(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
	// is 13
	filter := bson.D{{"gender", "Female"}}

	// Find options in other format
	// is 9
	fo := lxDb.FindOptions{
		Skip:  int64(4),
		Limit: int64(10),
	}

	// Get count
	res, err := base.CountDocuments(filter, fo.ToMongoCountOptions())
	its.NoError(err)
	its.Equal(int64(9), res)
}

func TestMongoDbBaseRepo_EstimatedDocumentCount(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupDataNew(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	// Get count
	res, err := base.EstimatedDocumentCount()
	its.NoError(err)
	its.Equal(int64(len(testUsers)), res)
}

func TestMongoDbBaseRepo_Find(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupDataNew(db)

	// Create expectUsers for compare result
	skip := 5
	limit := 5
	y := 0
	var (
		expectUsers  []TestUser
		expectFemale []TestUser
	)
	for i := range testUsers {
		if testUsers[i].Gender == "Female" {
			expectFemale = append(expectFemale, testUsers[i])
		}
		if i > (skip-1) && y < limit {
			expectUsers = append(expectUsers, testUsers[i])
			y++
		}
	}

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	t.Run("with options", func(t *testing.T) {
		// Find options in other format
		fo := lxDb.FindOptions{
			Sort:  map[string]int{"name": 1},
			Skip:  int64(5),
			Limit: int64(5),
		}

		// Find and compare with converted find options
		filter := bson.D{}
		var result []TestUser
		err = base.Find(filter, &result, fo.ToMongoFindOptions())
		its.NoError(err)
		its.Equal(expectUsers, result)
	})
	t.Run("with filter and options", func(t *testing.T) {
		// Find options in other format
		fo := lxDb.FindOptions{
			Sort: map[string]int{"name": 1},
		}
		filter := bson.D{{"gender", "Female"}}
		var result []TestUser
		err = base.Find(filter, &result, fo.ToMongoFindOptions())
		its.NoError(err)
		its.Equal(expectFemale, result)
	})
}

func TestMongoDbBaseRepo_FindOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupDataNew(db)
	testSkip := int64(5)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	t.Run("with options", func(t *testing.T) {
		// Find options in other format
		fo := lxDb.FindOptions{
			Sort: map[string]int{"name": 1},
			Skip: testSkip,
		}

		// Find and compare with converted find options
		filter := bson.D{}
		var result TestUser
		err = base.FindOne(filter, &result, fo.ToMongoFindOneOptions())
		its.NoError(err)
		its.Equal(testUsers[testSkip], result)
	})
	t.Run("with filter", func(t *testing.T) {
		filter := bson.D{{"email", testUsers[testSkip].Email}}
		var result TestUser
		err = base.FindOne(filter, &result)
		its.NoError(err)
		its.Equal(testUsers[testSkip], result)
	})
	t.Run("not found error", func(t *testing.T) {
		filter := bson.D{{"email", "unknown@email"}}
		var result TestUser
		err = base.FindOne(filter, &result)
		its.Error(err)
		its.IsType(&lxErrors.NotFoundError{}, err)
	})
}

func TestMongoDbBaseRepo_UpdateOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupDataNew(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	t.Run("PUT", func(t *testing.T) {
		// Update testUser
		tu := testUsers[4]
		tu.Name = "Is Updated"

		filter := bson.D{{"_id", tu.Id}}
		update := bson.D{{"$set", tu}}
		res, err := base.UpdateOne(filter, update)
		its.NoError(err)
		its.Equal(int64(1), res.MatchedCount)
		its.Equal(int64(1), res.ModifiedCount)
		its.Equal(int64(0), res.UpsertedCount)
		its.Nil(res.UpsertedID)

		// Check with Find
		var check TestUser
		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
		its.Equal(tu.Name, check.Name)
	})
	t.Run("PATCH", func(t *testing.T) {
		// Update testUser
		tu := testUsers[5]
		newName := "Is Updated"

		filter := bson.D{{"_id", tu.Id}}
		update := bson.D{{"$set", bson.D{{"name", newName}}}}
		res, err := base.UpdateOne(filter, update)
		its.NoError(err)
		its.Equal(int64(1), res.MatchedCount)
		its.Equal(int64(1), res.ModifiedCount)
		its.Equal(int64(0), res.UpsertedCount)
		its.Nil(res.UpsertedID)

		// Check with Find
		var check TestUser
		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
		its.Equal(newName, check.Name)
	})
}

func TestMongoDbBaseRepo_UpdateMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	setupDataNew(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	// All females to active, should be 13
	filter := bson.D{{"gender", "Female"}}
	update := bson.D{{"$set",
		bson.D{{"is_active", true}},
	}}
	res, err := base.UpdateMany(filter, update)
	its.NoError(err)
	// 13 female in db
	its.Equal(int64(13), res.MatchedCount)
	// 8 inactive female
	its.Equal(int64(8), res.ModifiedCount)
	its.Equal(int64(0), res.UpsertedCount)
	its.Nil(res.UpsertedID)

	// Check with Count
	filter = bson.D{{"gender", "Female"}, {"is_active", true}}
	count, err := base.CountDocuments(filter)
	its.NoError(err)
	its.Equal(int64(13), count)
}

func TestMongoDbBaseRepo_DeleteOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupDataNew(db)
	user := testUsers[10]

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	filter := bson.D{{"_id", user.Id}}
	res, err := base.DeleteOne(filter)
	its.NoError(err)
	its.Equal(int64(1), res)

	// Check with Find
	var check TestUser
	err = base.FindOne(bson.D{{"_id", user.Id}}, &check)
	its.Error(err)
	its.IsType(&lxErrors.NotFoundError{}, err)
}

func TestMongoDbBaseRepo_DeleteMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	setupDataNew(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	// All Males, should be 12
	filter := bson.D{{"gender", "Male"}}
	res, err := base.DeleteMany(filter)
	its.NoError(err)
	its.Equal(int64(12), res)

	// Check with Count
	filter = bson.D{{"gender", "Male"}}
	count, err := base.CountDocuments(filter)
	its.NoError(err)
	its.Equal(int64(0), count)
}
