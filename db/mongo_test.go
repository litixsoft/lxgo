package lxDb_test

import (
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	lxDb "github.com/litixsoft/lxgo/db"
	lxDbMocks "github.com/litixsoft/lxgo/db/mocks"
	lxHelper "github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"testing"
	"time"
)

//
//import (
//	"context"
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/golang/mock/gomock"
//	lxDb "github.com/litixsoft/lxgo/db"
//	lxDbMocks "github.com/litixsoft/lxgo/db/mocks"
//	"github.com/litixsoft/lxgo/helper"
//	"github.com/stretchr/testify/assert"
//	"go.mongodb.org/mongo-driver/bson"
//	"go.mongodb.org/mongo-driver/bson/primitive"
//	"go.mongodb.org/mongo-driver/mongo"
//	"go.mongodb.org/mongo-driver/mongo/options"
//	"io/ioutil"
//	"log"
//	"os"
//	"sort"
//	"testing"
//	"time"
//)

type TestUser struct {
	Id       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Gender   string             `json:"gender,omitempty" bson:"gender,omitempty"`
	Email    string             `json:"email" bson:"email"`
	IsActive bool               `json:"is_active,omitempty" bson:"is_active"`
}

type TestUserLocale struct {
	Index    int    `json:"index" bson:"index"`
	Name     string `json:"name" bson:"name"`
	Gender   string `json:"gender" bson:"gender"`
	Email    string `json:"email" bson:"email"`
	IsActive bool   `json:"isActive,omitempty" bson:"isActive"`
	Username string `json:"username" bson:"username"`
}

const (
	TestDbName           = "lxgo_test"
	TestCollection       = "users"
	TestCollectionLocale = "locale"
	FixturesPath         = "fixtures/users.test.json"
	FixturesPathLocale   = "fixtures/locale.test.json"
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

// setupData, create the test data and prepare the database
func setupData(db *mongo.Database) []TestUser {
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

func setupDataLocale(db *mongo.Database) []TestUserLocale {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// Locale
	collection := db.Collection(TestCollectionLocale)
	if err := collection.Drop(ctx); err != nil {
		log.Fatal(err)
	}

	// Load test data from json file
	raw, err := ioutil.ReadFile(FixturesPathLocale)
	if err != nil {
		log.Fatal(err)
	}

	// Convert
	var users []TestUserLocale
	if err := json.Unmarshal(raw, &users); err != nil {
		log.Fatal(err)
	}

	// Make Test users map and insert test data in db
	for i := 0; i < len(users); i++ {
		// Insert user
		_, err := collection.InsertOne(ctx, users[i])
		if err != nil {
			log.Fatal(err)
		}
	}

	// Return mongo connection
	return users
}

func TestGetMongoDbClient(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)
	its.IsType(&mongo.Client{}, client)
}

func TestMongoBaseRepo_CreateIndexes(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	lxHelper.HandlePanicErr(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Get indexes sub function
	getIndexes := func() []bson.D {
		// Check indexes
		idx := collection.Indexes()
		cur, err := idx.List(context.Background())
		assert.NoError(t, err)

		// Indexes
		var indexes []bson.D

		for cur.Next(context.Background()) {
			index := bson.D{}
			assert.NoError(t, cur.Decode(&index))
			indexes = append(indexes, index)
		}

		return indexes
	}

	// Create test data
	setupData(db)

	// Update many with expireAt
	update := bson.D{{"$set",
		bson.D{{"expireAt", time.Now().Add(time.Hour * 24)}},
	}}
	_, err = collection.UpdateMany(context.Background(), bson.D{}, update)
	lxHelper.HandlePanicErr(err)

	// Check indexes before test setup
	t.Run("before setup", func(t *testing.T) {
		indexes := getIndexes()
		its.Equal(1, len(indexes))

		expected := primitive.D{
			{"v", int32(2)},
			{"key", primitive.D{{"_id", int32(1)}}},
			{"name", "_id_"},
			{"ns", "lxgo_test.users"},
		}

		its.Equal(expected, indexes[0])
	})
	// Check indexes after test setup
	t.Run("after setup", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

		// Models for test
		indexModels := []mongo.IndexModel{
			{
				Keys: primitive.D{{"email", int32(1)}},
			},
			{
				Keys:    primitive.D{{"expireAt", int32(1)}},
				Options: options.Index().SetExpireAfterSeconds(int32(1)),
			},
		}

		// Check return from CreateMany
		retIdx, err := base.CreateIndexes(indexModels)
		its.NoError(err)

		// Should be return only created indexes
		its.Len(retIdx, 2)

		// Check indexes from helper sub function
		indexes := getIndexes()

		// Should be return all indexes
		its.Equal(3, len(indexes))

		// indexes[1] should be email
		expectedUserId := primitive.D{
			{"v", int32(2)},
			{"key", primitive.D{{"email", int32(1)}}},
			{"name", "email_1"},
			{"ns", "lxgo_test.users"},
		}

		its.Equal(expectedUserId, indexes[1])

		// indexes[2] should be expireAt
		expectedExpire := primitive.D{
			{"v", int32(2)},
			{"key", primitive.D{{"expireAt", int32(1)}}},
			{"name", "expireAt_1"},
			{"ns", "lxgo_test.users"},
			{"expireAfterSeconds", int32(1)},
		}

		its.Equal(expectedExpire, indexes[2])
	})
}

func TestMongoDbBaseRepo_InsertOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	lxHelper.HandlePanicErr(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Mock for base repo
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

	// Test the base repo with mock
	base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

	// AuditAuth user
	au := &bson.M{"name": "Timo Liebetrau"}

	testUser := TestUser{Name: "TestName", Email: "test@test.de"}

	t.Run("without_audit", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))

		// Channel for close
		res, err := base.InsertOne(&testUser, time.Second*10, options.InsertOne())
		its.NoError(err)

		// Check insert result id
		var checkUser TestUser
		filter := bson.D{{"_id", res.(primitive.ObjectID)}}
		its.NoError(base.FindOne(filter, &checkUser))

		its.Equal(testUser.Name, checkUser.Name)
		its.Equal(testUser.Email, checkUser.Email)
		its.Equal(testUser.IsActive, checkUser.IsActive)
	})
	t.Run("with_audit", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))

		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Repo with audit mocks
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// Check mock params
		//var chkInsertedId interface{}
		doAction := func(elem interface{}) {
			switch val := elem.(type) {
			case bson.M:
				//	if val["action"] == "insert" {
				//		bm, err := lxDb.ToBsonMap(val["data"])
				//		its.NoError(err)
				//		for k, v := range val["InsertedID"].(bson.M) {
				//			if _, ok := bm[k]; !ok {
				//				bm[k] = v
				//				chkInsertedId = v
				//			}
				//		}
				//		val["data"] = bm
				//

				//chkInsertedId := val["data"].(bson.M)[""]

				// Check
				its.Equal(lxDb.Insert, val["action"].(string))
				its.Equal(au, val["user"])
				its.Equal(testUser.Email, val["data"].(bson.M)["email"])
			//	}
			default:
				t.Fail()
			}
		}

		// Configure mock
		//mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
		mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
		mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

		// Channel for close and run test
		//done := make(chan bool)
		sidName := &lxDb.SubIdName{Name: "_id"}
		res, err := base.InsertOne(&testUser, lxDb.SetAuditAuth(au), sidName)

		// Check id
		//its.Equal(chkInsertedId, res.(primitive.ObjectID))

		//// Wait for close channel and check err
		//<-done
		its.NoError(err)

		// Check insert result id
		var checkUser TestUser
		filter := bson.D{{"_id", res.(primitive.ObjectID)}}
		its.NoError(base.FindOne(filter, &checkUser))

		//its.Equal(ident, checkUser.Id)
		its.Equal(testUser.Name, checkUser.Name)
		its.Equal(testUser.Email, checkUser.Email)
		its.Equal(testUser.IsActive, checkUser.IsActive)
	})
	//	t.Run("with audit error", func(t *testing.T) {
	//		// Drop for test
	//		its.NoError(collection.Drop(context.Background()))
	//
	//		// Test the base repo with mock
	//		mockCtrl := gomock.NewController(t)
	//		defer mockCtrl.Finish()
	//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
	//
	//		// Repo with audit mocks
	//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
	//
	//		// AuthUser
	//		au := &bson.M{"name": "Timo Liebetrau"}
	//
	//		// Check mock params
	//		doAction := func(act string, usr, data interface{}, elem ...interface{}) {
	//			its.Equal(lxDb.Insert, act)
	//			its.Equal(au, usr)
	//			its.Equal(testUser.Name, data.(bson.M)["name"])
	//			its.Equal(testUser.Email, data.(bson.M)["email"])
	//		}
	//
	//		// Configure mock
	//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test-error")).Do(doAction).Times(1)
	//
	//		// Channel for close and run test
	//		done := make(chan bool)
	//		chanErr := make(chan error)
	//		_, err := base.InsertOne(&testUser, lxDb.SetAuditAuth(au), done, chanErr)
	//
	//		// Wait for close and error channel from audit thread
	//		its.Error(<-chanErr)
	//		its.True(<-done)
	//
	//		// Check update return
	//		its.NoError(err)
	//	})
}

//func TestMongoDbBaseRepo_InsertMany(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Mock for base repo
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//	// AuditAuth user
//	au := &bson.M{"name": "Timo Liebetrau"}
//
//	t.Run("insert many", func(t *testing.T) {
//		// Drop for test
//		its.NoError(collection.Drop(context.Background()))
//
//		testUsers := make([]interface{}, 0)
//		for i := 0; i < 5; i++ {
//			name := fmt.Sprintf("TestName%d", i)
//			email := fmt.Sprintf("test%d@test.de", i)
//			testUsers = append(testUsers, TestUser{Name: name, Email: email})
//		}
//
//		// Test the base repo
//		res, err := base.InsertMany(testUsers, time.Second*10, options.InsertMany())
//		its.NoError(err)
//		its.Equal(5, len(res.InsertedIDs))
//		its.Equal(int64(0), res.FailedCount)
//
//		// Find and compare
//		var checkUsers []TestUser
//		err = base.Find(bson.D{}, &checkUsers)
//		its.NoError(err)
//
//		// Check users
//		for i, v := range checkUsers {
//			its.Equal(testUsers[i].(TestUser).Name, v.Name)
//			its.Equal(testUsers[i].(TestUser).Email, v.Email)
//		}
//
//		// Check ids
//		for _, id := range res.InsertedIDs {
//			var res TestUser
//			its.NoError(base.FindOne(bson.D{{"_id", id}}, &res))
//			its.Equal(id, res.Id)
//		}
//	})
//	t.Run("with audit", func(t *testing.T) {
//		// Drop for test
//		its.NoError(collection.Drop(context.Background()))
//
//		testUsers := make([]interface{}, 0)
//		for i := 0; i < 5; i++ {
//			name := fmt.Sprintf("TestName%d", i)
//			email := fmt.Sprintf("test%d@test.de", i)
//			testUsers = append(testUsers, TestUser{Name: name, Email: email})
//		}
//
//		// Check mock params
//		doAction := func(entries []interface{}, elem ...interface{}) {
//			// Log entries for audit should be 5
//			its.Len(entries, 5)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//		// Test the base repo
//		done := make(chan bool)
//		sidName := &lxDb.SubIdName{Name: "_id"}
//		res, err := base.InsertMany(testUsers, lxDb.SetAuditAuth(au), done, sidName)
//
//		// Wait for close channel and check err
//		<-done
//
//		its.NoError(err)
//		its.Equal(5, len(res.InsertedIDs))
//		its.Equal(int64(0), res.FailedCount)
//
//		// Find and compare
//		var checkUsers []TestUser
//		err = base.Find(bson.D{}, &checkUsers)
//		its.NoError(err)
//
//		// Check users
//		for i, v := range checkUsers {
//			its.Equal(testUsers[i].(TestUser).Name, v.Name)
//			its.Equal(testUsers[i].(TestUser).Email, v.Email)
//		}
//
//		// Check ids
//		for _, id := range res.InsertedIDs {
//			var res TestUser
//			its.NoError(base.FindOne(bson.D{{"_id", id}}, &res))
//			its.Equal(id, res.Id)
//		}
//	})
//	t.Run("with audit error", func(t *testing.T) {
//		// Drop for test
//		its.NoError(collection.Drop(context.Background()))
//
//		testUsers := make([]interface{}, 0)
//		for i := 0; i < 5; i++ {
//			name := fmt.Sprintf("TestName%d", i)
//			email := fmt.Sprintf("test%d@test.de", i)
//			testUsers = append(testUsers, TestUser{Name: name, Email: email})
//		}
//
//		// Check mock params
//		doAction := func(entries []interface{}, elem ...interface{}) {
//			// Log entries for audit should be 5
//			its.Len(entries, 5)
//		}
//
//		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(
//			errors.New("test-error")).Do(doAction).Times(1)
//
//		// Test the base repo
//		done := make(chan bool)
//		chanErr := make(chan error)
//		res, err := base.InsertMany(testUsers, lxDb.SetAuditAuth(au), done, chanErr)
//
//		// Wait for close and error channel from audit thread
//		its.Error(<-chanErr)
//		its.True(<-done)
//
//		its.NoError(err)
//		its.Equal(5, len(res.InsertedIDs))
//		its.Equal(int64(0), res.FailedCount)
//
//		// Find and compare
//		var checkUsers []TestUser
//		err = base.Find(bson.D{}, &checkUsers)
//		its.NoError(err)
//
//		// Check users
//		for i, v := range checkUsers {
//			its.Equal(testUsers[i].(TestUser).Name, v.Name)
//			its.Equal(testUsers[i].(TestUser).Email, v.Email)
//		}
//
//		// Check ids
//		for _, id := range res.InsertedIDs {
//			var res TestUser
//			its.NoError(base.FindOne(bson.D{{"_id", id}}, &res))
//			its.Equal(id, res.Id)
//		}
//	})
//	t.Run("with audit empty inserts", func(t *testing.T) {
//		// Drop for test
//		its.NoError(collection.Drop(context.Background()))
//
//		testUsers := make([]interface{}, 0)
//
//		// Test the base repo
//		res, err := base.InsertMany(testUsers, lxDb.SetAuditAuth(au))
//
//		its.NoError(err)
//		its.Equal(0, len(res.InsertedIDs))
//		its.Equal(int64(0), res.FailedCount)
//	})
//}
//
//func TestMongoDbBaseRepo_CountDocuments(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	setupData(db)
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//	// is 13
//	filter := bson.D{{"gender", "Female"}}
//
//	// Find options in other format
//	// is 9
//	fo := lxHelper.FindOptions{
//		Skip:  int64(4),
//		Limit: int64(10),
//	}
//
//	// Get count
//	res, err := base.CountDocuments(filter, fo.ToMongoCountOptions(), time.Second*10)
//	its.NoError(err)
//	its.Equal(int64(9), res)
//}
//
//func TestMongoDbBaseRepo_EstimatedDocumentCount(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	testUsers := setupData(db)
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//	// Get count
//	res, err := base.EstimatedDocumentCount(time.Second*10, options.EstimatedDocumentCount())
//	its.NoError(err)
//	its.Equal(int64(len(testUsers)), res)
//}
//
//func TestMongoDbBaseRepo_Find(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	testUsers := setupData(db)
//
//	// Create expectUsers for compare result
//	skip := 5
//	limit := 5
//	y := 0
//	var (
//		expectUsers  []TestUser
//		expectFemale []TestUser
//	)
//	for i := range testUsers {
//		if testUsers[i].Gender == "Female" {
//			expectFemale = append(expectFemale, testUsers[i])
//		}
//		if i > (skip-1) && y < limit {
//			expectUsers = append(expectUsers, testUsers[i])
//			y++
//		}
//	}
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//	t.Run("with options", func(t *testing.T) {
//		// Find options in other format
//		fo := lxHelper.FindOptions{
//			Sort:  map[string]int{"name": 1},
//			Skip:  int64(5),
//			Limit: int64(5),
//		}
//
//		// Find and compare with converted find options
//		filter := bson.D{}
//		var result []TestUser
//		err = base.Find(filter, &result, fo.ToMongoFindOptions(), time.Second*10)
//		its.NoError(err)
//		its.Equal(expectUsers, result)
//	})
//	t.Run("with filter and options", func(t *testing.T) {
//		// Find options in other format
//		fo := lxHelper.FindOptions{
//			Sort: map[string]int{"name": 1},
//		}
//		filter := bson.D{{"gender", "Female"}}
//		var result []TestUser
//		err = base.Find(filter, &result, fo.ToMongoFindOptions())
//		its.NoError(err)
//		its.Equal(expectFemale, result)
//	})
//}
//
//func TestMongoDbBaseRepo_FindOne(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	testUsers := setupData(db)
//	testSkip := int64(5)
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//	t.Run("with options", func(t *testing.T) {
//		// Find options in other format
//		fo := lxHelper.FindOptions{
//			Sort: map[string]int{"name": 1},
//			Skip: testSkip,
//		}
//
//		// Find and compare with converted find options
//		filter := bson.D{}
//		var result TestUser
//		err = base.FindOne(filter, &result, fo.ToMongoFindOneOptions(), time.Second*10)
//		its.NoError(err)
//		its.Equal(testUsers[testSkip], result)
//	})
//	t.Run("with filter", func(t *testing.T) {
//		filter := bson.D{{"email", testUsers[testSkip].Email}}
//		var result TestUser
//		err = base.FindOne(filter, &result)
//		its.NoError(err)
//		its.Equal(testUsers[testSkip], result)
//	})
//	t.Run("not found error", func(t *testing.T) {
//		filter := bson.D{{"email", "unknown@email"}}
//		var result TestUser
//		err = base.FindOne(filter, &result)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//}
//
//func TestMongoBaseRepo_FindOneAndDelete(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Sorted by name
//	testUsers := setupData(db)
//
//	t.Run("with options", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		// Find options in other format
//		// Sort by name as option
//		fo := lxHelper.FindOptions{
//			Sort: map[string]int{"name": 1},
//		}
//
//		// Empty filter for find first entry
//		filter := bson.D{}
//		var result TestUser
//
//		// Find first entry sort by name and delete
//		err = base.FindOneAndDelete(filter, &result, fo.ToMongoFindOneAndDeleteOptions(), time.Second*15)
//		its.NoError(err)
//		its.Equal(testUsers[0], result)
//	})
//	t.Run("with filter", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		filter := bson.D{{"email", testUsers[5].Email}}
//		var result TestUser
//		err = base.FindOneAndDelete(filter, &result)
//		its.NoError(err)
//		its.Equal(testUsers[5], result)
//	})
//	t.Run("not found error", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		filter := bson.D{{"email", "unknown@email"}}
//		var result TestUser
//		err = base.FindOneAndDelete(filter, &result)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//	t.Run("with audit", func(t *testing.T) {
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// Test user
//		tu := testUsers[6]
//
//		// AuditAuth
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check mock params
//		doAction := func(action string, authUser, data interface{}, elem ...interface{}) {
//			its.Equal(lxDb.Delete, action)
//			its.Equal(au, authUser)
//			its.Equal(&tu, data)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//		filter := bson.D{{"_id", tu.Id}}
//		done := make(chan bool)
//		var result TestUser
//		err = base.FindOneAndDelete(filter, &result, lxDb.SetAuditAuth(au), done)
//
//		// Wait for close channel and check err
//		<-done
//		its.NoError(err)
//
//		// Check result
//		its.Equal(tu, result)
//
//		// Check with find, user should be deleted
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//	t.Run("with audit error", func(t *testing.T) {
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// Test user
//		tu := testUsers[7]
//
//		// AuditAuth
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check mock params
//		doAction := func(action string, authUser, data interface{}, elem ...interface{}) {
//			its.Equal(lxDb.Delete, action)
//			its.Equal(au, authUser)
//			its.Equal(&tu, data)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(
//			errors.New("test-error")).Do(doAction).Times(1)
//
//		filter := bson.D{{"_id", tu.Id}}
//		done := make(chan bool)
//		chanErr := make(chan error)
//		var result TestUser
//		err = base.FindOneAndDelete(filter, &result, lxDb.SetAuditAuth(au), done, chanErr)
//
//		// Wait for close and error channel from audit thread
//		its.Error(<-chanErr)
//		its.True(<-done)
//
//		its.NoError(err)
//
//		// Check result
//		its.Equal(tu, result)
//
//		// Check with find, user should be deleted
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//}
//
//func TestMongoBaseRepo_FindOneAndReplace(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	t.Run("without audits", func(t *testing.T) {
//		t.Run("with options after", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and replace it
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, time.Second*10)
//			its.NoError(err)
//			its.Equal(replaceUser, result)
//		})
//		t.Run("with options before", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts)
//			its.NoError(err)
//			its.Equal(tu, result)
//		})
//		t.Run("without options", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			//opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and replace
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts)
//			its.NoError(err)
//			its.Equal(tu, result)
//		})
//		t.Run("not found error", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{{"_id", primitive.NewObjectID()}}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts)
//			its.Error(err)
//			its.True(errors.Is(err, lxDb.ErrNotFound))
//		})
//	})
//	t.Run("with audits", func(t *testing.T) {
//		t.Run("SetReturnDocument before", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				chkUser, err := lxDb.ToBsonMap(replaceUser)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and replace
//			// Empty filter for find first entry
//			filter := bson.D{}
//			done := make(chan bool)
//			var result TestUser
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(au), done)
//
//			// Wait for close channel and check err
//			<-done
//			its.NoError(err)
//			its.Equal(tu, result)
//		})
//		t.Run("SetReturnDocument before audit error", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				chkUser, err := lxDb.ToBsonMap(replaceUser)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(errors.New("test-error")).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and replace
//			// Empty filter for find first entry
//			filter := bson.D{}
//			done := make(chan bool)
//			chanErr := make(chan error)
//			var result TestUser
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(au), done, chanErr)
//
//			// Wait for close and error channel from audit thread
//			its.Error(<-chanErr)
//			its.True(<-done)
//
//			its.NoError(err)
//			its.Equal(tu, result)
//		})
//		t.Run("SetReturnDocument before not found error", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			fakeId := primitive.NewObjectID()
//			replaceUser := TestUser{
//				Id:     fakeId,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and replace
//			// Empty filter for find first entry
//			filter := bson.D{{"_id", fakeId}}
//			//done := make(chan bool)
//			var result TestUser
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(au))
//
//			// Wait for close channel and check err
//			//<-done
//			its.Error(err)
//			its.True(errors.Is(err, lxDb.ErrNotFound))
//		})
//		// Should be use options.After by default
//		t.Run("SetReturnDocument without options", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				chkUser, err := lxDb.ToBsonMap(replaceUser)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			// Should be use options.After by default
//			//opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and replace
//			// Empty filter for find first entry
//			filter := bson.D{}
//			done := make(chan bool)
//			var result TestUser
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(au), done)
//
//			// Wait for close channel and check err
//			<-done
//			its.NoError(err)
//			its.Equal(replaceUser, result)
//		})
//		t.Run("SetReturnDocument after audit error", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			replaceUser := TestUser{
//				Id:     tu.Id,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				chkUser, err := lxDb.ToBsonMap(replaceUser)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(errors.New("test-error")).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and replace
//			// Empty filter for find first entry
//			filter := bson.D{}
//			done := make(chan bool)
//			chanErr := make(chan error)
//			var result TestUser
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(au), done, chanErr)
//
//			// Wait for close and error channel from audit thread
//			its.Error(<-chanErr)
//			its.True(<-done)
//
//			its.NoError(err)
//			its.Equal(replaceUser, result)
//		})
//		t.Run("SetReturnDocument after not found error", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//			tu := testUsers[0]
//
//			// without is_active
//			fakeId := primitive.NewObjectID()
//			replaceUser := TestUser{
//				Id:     fakeId,
//				Name:   "NewReplacementUser",
//				Gender: tu.Gender,
//				Email:  tu.Email,
//			}
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndReplaceOptions()
//			opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and replace
//			// Empty filter for find first entry
//			filter := bson.D{{"_id", fakeId}}
//			//done := make(chan bool)
//			var result TestUser
//			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(au))
//
//			// Wait for close channel and check err
//			//<-done
//			its.Error(err)
//			its.True(errors.Is(err, lxDb.ErrNotFound))
//		})
//	})
//}
//
//func TestMongoBaseRepo_FindOneAndUpdate(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	t.Run("without audits", func(t *testing.T) {
//		t.Run("with options after", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and update
//			err = base.FindOneAndUpdate(filter, update, &result, opts, time.Second*10)
//			its.NoError(err)
//
//			expected := TestUser{
//				Id:       testUsers[0].Id,
//				Name:     "updated-name",
//				Gender:   testUsers[0].Gender,
//				Email:    testUsers[0].Email,
//				IsActive: testUsers[0].IsActive,
//			}
//
//			its.Equal(expected, result)
//		})
//		t.Run("with options before", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			err = base.FindOneAndUpdate(filter, update, &result, opts)
//			its.NoError(err)
//			its.Equal(testUsers[0], result)
//		})
//		t.Run("without options", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			//opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			err = base.FindOneAndUpdate(filter, update, &result, opts)
//			its.NoError(err)
//
//			// Should be equal with before option
//			its.Equal(testUsers[0], result)
//		})
//		t.Run("not found error", func(t *testing.T) {
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection)
//
//			// testUsers Sorted by name
//			setupData(db)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Empty filter for find first entry
//			filter := bson.D{{"_id", primitive.NewObjectID()}}
//			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
//			var result TestUser
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			err = base.FindOneAndUpdate(filter, update, &result, opts)
//			its.Error(err)
//			its.True(errors.Is(err, lxDb.ErrNotFound))
//		})
//	})
//	t.Run("with audits", func(t *testing.T) {
//		t.Run("SetReturnDocument before", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// New name for update
//			updatedName := "updatedName"
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				tu := testUsers[0]
//				tu.Name = updatedName
//				chkUser, err := lxDb.ToBsonMap(tu)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//			done := make(chan bool)
//			var result TestUser
//			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(au), done)
//
//			// Wait for close channel and check err
//			<-done
//			its.NoError(err)
//			its.Equal(testUsers[0], result)
//		})
//		t.Run("SetReturnDocument before audit error", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// New name for update
//			updatedName := "updatedName"
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				tu := testUsers[0]
//				tu.Name = updatedName
//				chkUser, err := lxDb.ToBsonMap(tu)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(errors.New("test-error")).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//			done := make(chan bool)
//			chanErr := make(chan error)
//			var result TestUser
//			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(au), done, chanErr)
//
//			// Wait for close and error channel from audit thread
//			its.Error(<-chanErr)
//			its.True(<-done)
//
//			its.NoError(err)
//			its.Equal(testUsers[0], result)
//		})
//		t.Run("SetReturnDocument before not found error", func(t *testing.T) {
//			// testUsers Sorted by name
//			setupData(db)
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// New name for update
//			updatedName := "updatedName"
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.Before)
//
//			// Find first entry sort by name and update
//			// Empty filter for find first entry
//			filter := bson.D{{"_id", primitive.NewObjectID()}}
//			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//			var result TestUser
//			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(au))
//
//			its.Error(err)
//			its.True(errors.Is(err, lxDb.ErrNotFound))
//		})
//		t.Run("without SetReturnDocument", func(t *testing.T) {
//			// When SetReturnDocument not set, use options.After by default
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// New name for update
//			updatedName := "updatedName"
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				tu := testUsers[0]
//				tu.Name = updatedName
//				chkUser, err := lxDb.ToBsonMap(tu)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			// When SetReturnDocument not set, use options.After by default
//			//opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and update
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//			done := make(chan bool)
//			var result TestUser
//			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(au), done)
//
//			// Wait for close channel and check err
//			<-done
//			its.NoError(err)
//
//			// Test user for check
//			tu := testUsers[0]
//			tu.Name = updatedName
//
//			its.Equal(tu, result)
//		})
//		t.Run("SetReturnDocument after audit error", func(t *testing.T) {
//			// testUsers Sorted by name
//			testUsers := setupData(db)
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// New name for update
//			updatedName := "updatedName"
//
//			// Check params with doAction
//			doAction := func(action string, user, data interface{}, elem ...interface{}) {
//				its.Equal(lxDb.Update, action)
//				its.Equal(au, user)
//
//				// Map for check
//				tu := testUsers[0]
//				tu.Name = updatedName
//				chkUser, err := lxDb.ToBsonMap(tu)
//
//				// Check
//				its.NoError(err)
//				its.Equal(&chkUser, data)
//			}
//
//			// Configure mock
//			mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).
//				Return(errors.New("test-error")).Do(doAction).Times(1)
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and update
//			// Empty filter for find first entry
//			filter := bson.D{}
//			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//			done := make(chan bool)
//			chanErr := make(chan error)
//			var result TestUser
//			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(au), done, chanErr)
//
//			// Wait for close and error channel from audit thread
//			its.Error(<-chanErr)
//			its.True(<-done)
//
//			its.NoError(err)
//
//			tu := testUsers[0]
//			tu.Name = updatedName
//			its.Equal(tu, result)
//		})
//		t.Run("SetReturnDocument after not found error", func(t *testing.T) {
//			// testUsers Sorted by name
//			setupData(db)
//
//			// Test the base repo with mock
//			mockCtrl := gomock.NewController(t)
//			defer mockCtrl.Finish()
//			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//			// Test the base repo
//			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//			// AuthUser
//			au := &bson.M{"name": "Timo Liebetrau"}
//
//			// New name for update
//			updatedName := "updatedName"
//
//			// Find options in other format
//			// Sort by name as option
//			fo := lxHelper.FindOptions{
//				Sort: map[string]int{"name": 1},
//			}
//
//			// Set options and return document
//			opts := fo.ToMongoFindOneAndUpdateOptions()
//			opts.SetReturnDocument(options.After)
//
//			// Find first entry sort by name and update
//			// Empty filter for find first entry
//			filter := bson.D{{"_id", primitive.NewObjectID()}}
//			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//			var result TestUser
//			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(au))
//
//			its.Error(err)
//			its.True(errors.Is(err, lxDb.ErrNotFound))
//		})
//	})
//}
//
//func TestMongoDbBaseRepo_UpdateOne(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//	testUsers := setupData(db)
//
//	t.Run("update", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//		// Update testUser
//		tu := testUsers[5]
//		newName := "Is Updated"
//
//		filter := bson.D{{"_id", tu.Id}}
//		update := bson.D{{"$set", bson.D{{"name", newName}}}}
//		err := base.UpdateOne(filter, update, time.Second*10, options.Update())
//		its.NoError(err)
//
//		// Check with Find
//		var check TestUser
//		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
//		its.Equal(newName, check.Name)
//	})
//	t.Run("with audit", func(t *testing.T) {
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// Update testUser
//		tu := testUsers[6]
//		tu.Name = "Is Updated"
//
//		// TestUser to map for tests
//		doc, err := lxDb.ToBsonMap(&tu)
//		its.NoError(err)
//
//		// AuthUser
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check params with doAction
//		doAction := func(action string, user, data interface{}, elem ...interface{}) {
//			its.Equal(lxDb.Update, action)
//			its.Equal(au, user)
//			its.Equal(&doc, data)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//		filter := bson.D{{"_id", tu.Id}}
//		update := bson.D{{"$set", bson.D{{"name", tu.Name}}}}
//		done := make(chan bool)
//		err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(au), done)
//
//		// Wait for close channel and check err
//		<-done
//		its.NoError(err)
//
//		// Check with Find
//		var check TestUser
//		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
//		its.Equal(tu.Name, check.Name)
//	})
//	t.Run("with audit error", func(t *testing.T) {
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// Update testUser
//		tu := testUsers[7]
//		tu.Name = "Is Updated"
//
//		// TestUser to map for tests
//		doc, err := lxDb.ToBsonMap(&tu)
//		its.NoError(err)
//
//		// AuthUser
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check params with doAction
//		doAction := func(action string, user, data interface{}, elem ...interface{}) {
//			its.Equal(lxDb.Update, action)
//			its.Equal(au, user)
//			its.Equal(&doc, data)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test-error")).Do(doAction).Times(1)
//
//		filter := bson.D{{"_id", tu.Id}}
//		update := bson.D{{"$set", bson.D{{"name", tu.Name}}}}
//		done := make(chan bool)
//		chanErr := make(chan error)
//		err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(au), done, chanErr)
//
//		// Wait for close and error channel from audit thread
//		its.Error(<-chanErr)
//		its.True(<-done)
//
//		// Check update return
//		its.NoError(err)
//	})
//	t.Run("with not found error", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//		// Update testUser
//		newName := "Is Updated"
//
//		filter := bson.D{{"_id", primitive.NewObjectID()}}
//		update := bson.D{{"$set", bson.D{{"name", newName}}}}
//		err := base.UpdateOne(filter, update)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//}
//
//func TestMongoDbBaseRepo_UpdateMany(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	t.Run("without audit", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		// Setup test data
//		setupData(db)
//
//		// All females to active, should be 13
//		filter := bson.D{{"gender", "Female"}}
//		update := bson.D{{"$set",
//			bson.D{{"is_active", true}},
//		}}
//		chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}
//
//		res, err := base.UpdateMany(filter, update, time.Second*10, options.Update())
//		its.NoError(err)
//
//		// 13 female in db
//		its.Equal(int64(13), res.MatchedCount)
//		// 8 inactive female
//		its.Equal(int64(8), res.ModifiedCount)
//		// 0 fails and upserted
//		its.Equal(int64(0), res.FailedCount)
//		its.Empty(res.FailedIDs)
//		its.Equal(int64(0), res.UpsertedCount)
//		its.Nil(res.UpsertedID)
//
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(13), count)
//	})
//	t.Run("with audit", func(t *testing.T) {
//		// testUsers test data Sorted by name
//		setupData(db)
//
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// AuthUser
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check mock params
//		doAction := func(entries []interface{}, elem ...interface{}) {
//			// Log entries for audit should be 8
//			its.Len(entries, 8)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//		// All females to active,
//		// All females should be 13
//		// All inactive females should be 8
//		filter := bson.D{{"gender", "Female"}}
//		chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}
//		update := bson.D{{"$set",
//			bson.D{{"is_active", true}},
//		}}
//		done := make(chan bool)
//		sidName := &lxDb.SubIdName{Name: "_id"}
//		res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(au), done, sidName)
//
//		// Wait for close channel and check err
//		<-done
//		its.NoError(err)
//
//		// 13 female in db
//		its.Equal(int64(13), res.MatchedCount)
//		// 8 inactive female
//		its.Equal(int64(8), res.ModifiedCount)
//		// 0 fails and upserted
//		its.Equal(int64(0), res.FailedCount)
//		its.Empty(res.FailedIDs)
//		its.Equal(int64(0), res.UpsertedCount)
//		its.Nil(res.UpsertedID)
//
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(13), count)
//	})
//	t.Run("with audit error", func(t *testing.T) {
//		// testUsers test data Sorted by name
//		setupData(db)
//
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// AuthUser
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check mock params
//		doAction := func(entries []interface{}, elem ...interface{}) {
//			// Log entries for audit should be 8
//			its.Len(entries, 8)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(
//			errors.New("test-error")).Do(doAction).Times(1)
//
//		// All females to active,
//		// All females should be 13
//		// All inactive females should be 8
//		filter := bson.D{{"gender", "Female"}}
//		chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}
//		update := bson.D{{"$set",
//			bson.D{{"is_active", true}},
//		}}
//		done := make(chan bool)
//		chanErr := make(chan error)
//		res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(au), done, chanErr)
//
//		// Wait for close and error channel from audit thread
//		its.Error(<-chanErr)
//		its.True(<-done)
//		its.NoError(err)
//
//		// 13 female in db
//		its.Equal(int64(13), res.MatchedCount)
//		// 8 inactive female
//		its.Equal(int64(8), res.ModifiedCount)
//		// 0 fails and upserted
//		its.Equal(int64(0), res.FailedCount)
//		its.Empty(res.FailedIDs)
//		its.Equal(int64(0), res.UpsertedCount)
//		its.Nil(res.UpsertedID)
//
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(13), count)
//	})
//	t.Run("with audit and empty updates", func(t *testing.T) {
//		// testUsers test data Sorted by name
//		setupData(db)
//
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// AuthUser
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Wrong filter for 0 matches
//		filter := bson.D{{"gender", "Animal"}}
//		update := bson.D{{"$set",
//			bson.D{{"is_active", true}},
//		}}
//
//		res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(au))
//		its.NoError(err)
//
//		// 0 animals in db
//		its.Equal(int64(0), res.MatchedCount)
//		// 0 modified
//		its.Equal(int64(0), res.ModifiedCount)
//		// 0 fails and upserted
//		its.Equal(int64(0), res.FailedCount)
//		its.Empty(res.FailedIDs)
//		its.Equal(int64(0), res.UpsertedCount)
//		its.Nil(res.UpsertedID)
//	})
//}
//
//func TestMongoDbBaseRepo_DeleteOne(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//	testUsers := setupData(db)
//
//	t.Run("delete", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		// Test user
//		tu := testUsers[5]
//
//		filter := bson.D{{"_id", tu.Id}}
//		err := base.DeleteOne(filter)
//		its.NoError(err)
//
//		// Check with Find
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//	t.Run("with audit", func(t *testing.T) {
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// Test user
//		tu := testUsers[6]
//
//		// AuditAuth
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check mock params
//		doAction := func(action string, authUser, data interface{}, elem ...interface{}) {
//			its.Equal(lxDb.Delete, action)
//			its.Equal(au, authUser)
//
//			expectedData := bson.M{"_id": tu.Id}
//			its.Equal(expectedData, data)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//		filter := bson.D{{"_id", tu.Id}}
//		done := make(chan bool)
//		err = base.DeleteOne(filter, lxDb.SetAuditAuth(au), done)
//
//		// Wait for close channel and check err
//		<-done
//		its.NoError(err)
//
//		//// Check with Find
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//	t.Run("with audit error", func(t *testing.T) {
//		// Test the base repo with mock
//		mockCtrl := gomock.NewController(t)
//		defer mockCtrl.Finish()
//		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)
//
//		// Test user
//		tu := testUsers[7]
//
//		// AuditAuth
//		au := &bson.M{"name": "Timo Liebetrau"}
//
//		// Check mock params
//		doAction := func(act string, usr, data interface{}, elem ...interface{}) {
//			its.Equal(lxDb.Delete, act)
//			its.Equal(au, usr)
//
//			expectedData := bson.M{"_id": tu.Id}
//			its.Equal(expectedData, data)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(
//			errors.New("test-error")).Do(doAction).Times(1)
//
//		filter := bson.D{{"_id", tu.Id}}
//		done := make(chan bool)
//		chanErr := make(chan error)
//		err = base.DeleteOne(filter, lxDb.SetAuditAuth(au), done, chanErr)
//
//		// Wait for close and error channel from audit thread
//		its.Error(<-chanErr)
//		its.True(<-done)
//
//		its.NoError(err)
//
//		//// Check with Find
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//	t.Run("not found error", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		filter := bson.D{{"_id", primitive.NewObjectID()}}
//		err := base.DeleteOne(filter)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//	t.Run("delete with options", func(t *testing.T) {
//		// Test the base repo
//		base := lxDb.NewMongoBaseRepo(collection)
//
//		// Test user
//		tu := testUsers[9]
//
//		filter := bson.D{{"_id", tu.Id}}
//		err := base.DeleteOne(filter, time.Second*10, options.FindOneAndDelete())
//		its.NoError(err)
//
//		// Check with Find
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//}
//
//func TestMongoDbBaseRepo_DeleteMany(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//
//	// Mock for base repo
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//	mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection), mockIBaseRepoAudit)
//
//	// All Males, should be 12
//	filter := bson.D{{"gender", "Male"}}
//	chkFilter := bson.D{{"gender", "Male"}}
//
//	// AuditAuth user
//	au := &bson.M{"name": "Timo Liebetrau"}
//
//	t.Run("delete", func(t *testing.T) {
//		setupData(db)
//
//		res, err := base.DeleteMany(filter, time.Second*10, options.Delete())
//		its.NoError(err)
//
//		// 12 male in db
//		its.Equal(int64(12), res.DeletedCount)
//
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(0), count)
//	})
//	t.Run("with audit", func(t *testing.T) {
//		setupData(db)
//
//		// Check mock params
//		doAction := func(entries []interface{}, elem ...interface{}) {
//			// Log entries for audit should be 12
//			its.Len(entries, 12)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(nil).Do(doAction).Times(1)
//
//		done := make(chan bool)
//		sidName := &lxDb.SubIdName{Name: "_id"}
//		res, err := base.DeleteMany(filter, lxDb.SetAuditAuth(au), done, sidName)
//
//		// Wait for close channel and check err
//		<-done
//		its.NoError(err)
//
//		// 12 male in db
//		its.Equal(int64(12), res.DeletedCount)
//
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(0), count)
//	})
//	t.Run("with audit error", func(t *testing.T) {
//		setupData(db)
//
//		// Check mock params
//		doAction := func(entries []interface{}, elem ...interface{}) {
//			// Log entries for audit should be 12
//			its.Len(entries, 12)
//		}
//
//		// Configure mock
//		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(
//			errors.New("test-error")).Do(doAction).Times(1)
//
//		done := make(chan bool)
//		chanErr := make(chan error)
//		res, err := base.DeleteMany(filter, lxDb.SetAuditAuth(au), done, chanErr)
//
//		// Wait for close and error channel from audit thread
//		its.Error(<-chanErr)
//		its.True(<-done)
//
//		its.NoError(err)
//
//		// 12 male in db
//		its.Equal(int64(12), res.DeletedCount)
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(0), count)
//	})
//	t.Run("with audit and empty filter", func(t *testing.T) {
//		setupData(db)
//
//		res, err := base.DeleteMany(bson.D{{"gender", "Animals"}}, lxDb.SetAuditAuth(au))
//		its.NoError(err)
//		// 0 animals in db
//		its.Equal(int64(0), res.DeletedCount)
//	})
//}
//
//func TestMongoBaseRepo_GetCollection(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(collection)
//
//	its.Equal(collection, base.GetCollection())
//}
//
//func TestMongoBaseRepo_GetDb(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//	its.Equal(db, base.GetDb())
//}
//
//func TestMongoBaseRepo_GetRepoName(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(collection)
//
//	expect := db.Name() + "/" + collection.Name()
//
//	its.Equal(expect, base.GetRepoName())
//}
//
//func TestMongoDbBaseRepo_Aggregate(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	testUsers := setupData(db)
//
//	// Create expectUsers for compare result
//	skip := 5
//	limit := 5
//	y := 0
//	var (
//		expectUsers  []TestUser
//		expectFemale []TestUser
//	)
//	for i := range testUsers {
//		if testUsers[i].Gender == "Female" {
//			expectFemale = append(expectFemale, testUsers[i])
//		}
//		if i > (skip-1) && y < limit {
//			expectUsers = append(expectUsers, testUsers[i])
//			y++
//		}
//	}
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
//
//	t.Run("with $skip/$limit", func(t *testing.T) {
//		// Build mongo pipeline
//		pipeline := mongo.Pipeline{
//			bson.D{{
//				"$match", bson.M{},
//			}},
//			bson.D{{
//				"$sort", bson.M{
//					"name": 1,
//				},
//			}},
//			bson.D{{
//				"$skip", 5,
//			}},
//			bson.D{{
//				"$limit", 5,
//			}},
//		}
//
//		var result []TestUser
//		err = base.Aggregate(pipeline, &result, time.Second*10)
//
//		its.NoError(err)
//		its.Equal(expectUsers, result)
//	})
//	t.Run("with $match", func(t *testing.T) {
//		pipeline := mongo.Pipeline{
//			bson.D{{
//				"$match", bson.D{{"gender", "Female"}},
//			}},
//			bson.D{{
//				"$sort", bson.M{
//					"name": 1,
//				},
//			}},
//		}
//
//		var result []TestUser
//		err = base.Aggregate(pipeline, &result)
//		its.NoError(err)
//		its.Equal(expectFemale, result)
//	})
//	t.Run("without result", func(t *testing.T) {
//		pipeline := mongo.Pipeline{
//			bson.D{{
//				"$match", bson.D{{"gender", "Female"}},
//			}},
//			bson.D{{
//				"$sort", bson.M{
//					"name": 1,
//				},
//			}},
//		}
//
//		err = base.Aggregate(pipeline, nil)
//		its.NoError(err)
//	})
//}
//
//func TestMongoDbBaseRepo_LocaleTest(t *testing.T) {
//	its := assert.New(t)
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	its.NoError(err)
//
//	db := client.Database(TestDbName)
//	testUsers := setupDataLocale(db)
//
//	// sort by index that's will be ordered by username
//	sort.Slice(testUsers[:], func(i, j int) bool {
//		return testUsers[i].Index < testUsers[j].Index
//	})
//
//	// Test the base repo
//	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollectionLocale))
//	base.SetLocale("de")
//
//	t.Run("with find and locale set to 'de' and ordered by 'username'", func(t *testing.T) {
//		// Find options in other format
//		fo := lxHelper.FindOptions{
//			Sort: map[string]int{"username": 1},
//		}
//
//		filter := bson.M{}
//
//		var result []TestUserLocale
//		err = base.Find(filter, &result, fo.ToMongoFindOptions(), time.Second*10)
//
//		its.NoError(err)
//		its.Equal(testUsers, result)
//	})
//}
