package lxDb_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	lxDb "github.com/litixsoft/lxgo/db"
	lxDbMocks "github.com/litixsoft/lxgo/db/mocks"
	"github.com/litixsoft/lxgo/helper"
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

func TestGetMongoDbClient(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)
	its.IsType(&mongo.Client{}, client)
}

func TestMongoDbBaseRepo_InsertOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	lxHelper.HandlePanicErr(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Drop for test
	its.NoError(collection.Drop(context.Background()))

	testUser := TestUser{
		Name:  "TestName",
		Email: "test@test.de",
	}

	t.Run("insert", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		// Channel for close
		res, err := base.InsertOne(&testUser)
		its.NoError(err)

		// Check insert result id
		var checkUser TestUser
		filter := bson.D{{"_id", res.(primitive.ObjectID)}}
		its.NoError(base.FindOne(filter, &checkUser))

		its.Equal(testUser.Name, checkUser.Name)
		its.Equal(testUser.Email, checkUser.Email)
		its.Equal(testUser.IsActive, checkUser.IsActive)
	})
	t.Run("with audit", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Repo with audit mocks
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// AuthUser
		au := &bson.M{"name": "Timo Liebetrau"}

		// Check mock params
		doAction := func(act string, usr, data interface{}, elem ...interface{}) {
			its.Equal(lxDb.Insert, act)
			its.Equal(au, usr)
			its.Equal(&testUser, data)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)

		// Channel for close and run test
		done := make(chan bool)
		res, err := base.InsertOne(&testUser, lxDb.SetAuditAuth(au), done)

		// Wait for close channel and check err
		<-done
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
	t.Run("with audit error", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Repo with audit mocks
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// AuthUser
		au := &bson.M{"name": "Timo Liebetrau"}

		// Check mock params
		doAction := func(act string, usr, data interface{}, elem ...interface{}) {
			its.Equal(lxDb.Insert, act)
			its.Equal(au, usr)
			its.Equal(&testUser, data)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test-error")).Do(doAction).Times(1)

		// Channel for close and run test
		done := make(chan bool)
		chanErr := make(chan error)
		_, err := base.InsertOne(&testUser, lxDb.SetAuditAuth(au), done, chanErr)

		// Wait for close and error channel from audit thread
		its.Error(<-chanErr)
		its.True(<-done)

		// Check update return
		its.NoError(err)
	})
	t.Run("with timeout and options", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)
		au := &bson.M{"name": "Timo Liebetrau"}

		// Channel for close
		res, err := base.InsertOne(&testUser, &au, time.Second*10, options.InsertOne().SetBypassDocumentValidation(false))
		its.NoError(err)

		// Check insert result id
		var checkUser TestUser
		filter := bson.D{{"_id", res.(primitive.ObjectID)}}
		its.NoError(base.FindOne(filter, &checkUser))

		its.Equal(testUser.Name, checkUser.Name)
		its.Equal(testUser.Email, checkUser.Email)
		its.Equal(testUser.IsActive, checkUser.IsActive)
	})
}

func TestMongoDbBaseRepo_InsertMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	lxHelper.HandlePanicErr(err)

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
	res, err := base.InsertMany(testUsers, time.Second*10, options.InsertMany())
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
	setupData(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))
	// is 13
	filter := bson.D{{"gender", "Female"}}

	// Find options in other format
	// is 9
	fo := lxHelper.FindOptions{
		Skip:  int64(4),
		Limit: int64(10),
	}

	// Get count
	res, err := base.CountDocuments(filter, fo.ToMongoCountOptions(), time.Second*10)
	its.NoError(err)
	its.Equal(int64(9), res)
}

func TestMongoDbBaseRepo_EstimatedDocumentCount(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupData(db)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	// Get count
	res, err := base.EstimatedDocumentCount(time.Second*10, options.EstimatedDocumentCount())
	its.NoError(err)
	its.Equal(int64(len(testUsers)), res)
}

func TestMongoDbBaseRepo_Find(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupData(db)

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
		fo := lxHelper.FindOptions{
			Sort:  map[string]int{"name": 1},
			Skip:  int64(5),
			Limit: int64(5),
		}

		// Find and compare with converted find options
		filter := bson.D{}
		var result []TestUser
		err = base.Find(filter, &result, fo.ToMongoFindOptions(), time.Second*10)
		its.NoError(err)
		its.Equal(expectUsers, result)
	})
	t.Run("with filter and options", func(t *testing.T) {
		// Find options in other format
		fo := lxHelper.FindOptions{
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
	testUsers := setupData(db)
	testSkip := int64(5)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	t.Run("with options", func(t *testing.T) {
		// Find options in other format
		fo := lxHelper.FindOptions{
			Sort: map[string]int{"name": 1},
			Skip: testSkip,
		}

		// Find and compare with converted find options
		filter := bson.D{}
		var result TestUser
		err = base.FindOne(filter, &result, fo.ToMongoFindOneOptions(), time.Second*10)
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
		its.IsType(&lxDb.NotFoundError{}, err)
	})
}

func TestMongoDbBaseRepo_UpdateOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)
	testUsers := setupData(db)

	t.Run("update", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

		// Update testUser
		tu := testUsers[5]
		newName := "Is Updated"

		filter := bson.D{{"_id", tu.Id}}
		update := bson.D{{"$set", bson.D{{"name", newName}}}}
		err := base.UpdateOne(filter, update)
		its.NoError(err)

		// Check with Find
		var check TestUser
		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
		its.Equal(newName, check.Name)
	})
	t.Run("with timeout and options", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

		// Update testUser
		tu := testUsers[5]
		newName := "Is Updated"

		filter := bson.D{{"_id", tu.Id}}
		update := bson.D{{"$set", bson.D{{"name", newName}}}}
		err := base.UpdateOne(filter, update, time.Second*10, options.FindOneAndUpdate().SetMaxTime(time.Second*10))
		its.NoError(err)

		// Check with Find
		var check TestUser
		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
		its.Equal(newName, check.Name)
	})
	t.Run("with audit", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// Update testUser
		tu := testUsers[5]
		newName := "Is Updated"
		tu.Name = newName

		// TestUser to map for tests
		doc, err := lxDb.ToBsonDoc(&tu)
		its.NoError(err)

		// AuthUser
		au := &bson.M{"name": "Timo Liebetrau"}

		// Check params with doAction
		doAction := func(action string, user, data interface{}, elem ...interface{}) {
			its.Equal(lxDb.Update, action)
			its.Equal(au, user)
			its.Equal(doc, data)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)

		filter := bson.D{{"_id", tu.Id}}
		update := bson.D{{"$set", bson.D{{"name", newName}}}}
		done := make(chan bool)
		err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(au), done)

		// Wait for close channel and check err
		<-done
		its.NoError(err)

		// Check with Find
		var check TestUser
		its.NoError(base.FindOne(bson.D{{"_id", tu.Id}}, &check))
		its.Equal(newName, check.Name)
	})
	t.Run("with audit error", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// Update testUser
		tu := testUsers[5]
		tu.Name = "Is Updated"

		// TestUser to map for tests
		doc, err := lxDb.ToBsonDoc(&tu)
		its.NoError(err)

		// AuthUser
		au := &bson.M{"name": "Timo Liebetrau"}

		// Check params with doAction
		doAction := func(action string, user, data interface{}, elem ...interface{}) {
			its.Equal(lxDb.Update, action)
			its.Equal(au, user)
			its.Equal(doc, data)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("test-error")).Do(doAction).Times(1)

		filter := bson.D{{"_id", tu.Id}}
		update := bson.D{{"$set", bson.D{{"name", tu.Name}}}}
		done := make(chan bool)
		chanErr := make(chan error)
		err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(au), done, chanErr)

		// Wait for close and error channel from audit thread
		its.Error(<-chanErr)
		its.True(<-done)

		// Check update return
		its.NoError(err)
	})
	t.Run("with not found error", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

		// Update testUser
		newName := "Is Updated"

		filter := bson.D{{"_id", primitive.NewObjectID()}}
		update := bson.D{{"$set", bson.D{{"name", newName}}}}
		err := base.UpdateOne(filter, update)
		its.Error(err)
		its.IsType(&lxDb.NotFoundError{}, err)
	})
}

func TestMongoDbBaseRepo_UpdateMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)

	// Mock for base repo
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

	// Test the base repo with mock
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection), mockIBaseRepoAudit)

	// All females to active, should be 13
	filter := bson.D{{"gender", "Female"}}
	update := bson.D{{"$set",
		bson.D{{"is_active", true}},
	}}
	chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}

	// AuditAuth user
	au := &bson.M{"name": "Timo Liebetrau"}

	t.Run("without audit", func(t *testing.T) {
		setupData(db)

		res, err := base.UpdateMany(filter, update, time.Second*10, options.Update())
		its.NoError(err)

		// 13 female in db
		its.Equal(int64(13), res.MatchedCount)
		// 8 inactive female
		its.Equal(int64(8), res.ModifiedCount)
		// 0 fails and upserted
		its.Equal(int64(0), res.FailedCount)
		its.Empty(res.FailedIDs)
		its.Equal(int64(0), res.UpsertedCount)
		its.Nil(res.UpsertedID)

		// Check with Count
		count, err := base.CountDocuments(chkFilter)
		its.NoError(err)
		its.Equal(int64(13), count)
	})
	t.Run("with audit", func(t *testing.T) {
		setupData(db)

		// Check mock params
		doAction := func(entries []interface{}, elem ...interface{}) {
			// Log entries for audit should be 8
			its.Len(entries, 8)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(nil).Do(doAction).Times(1)

		done := make(chan bool)
		sidName := &lxDb.SubIdName{Name: "_id"}
		res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(au), done, sidName)

		// Wait for close channel and check err
		<-done
		its.NoError(err)

		// 13 female in db
		its.Equal(int64(13), res.MatchedCount)
		// 8 inactive female
		its.Equal(int64(8), res.ModifiedCount)
		// 0 fails and upserted
		its.Equal(int64(0), res.FailedCount)
		its.Empty(res.FailedIDs)
		its.Equal(int64(0), res.UpsertedCount)
		its.Nil(res.UpsertedID)

		// Check with Count
		count, err := base.CountDocuments(chkFilter)
		its.NoError(err)
		its.Equal(int64(13), count)
	})
	t.Run("with audit error", func(t *testing.T) {
		setupData(db)

		// Check mock params
		doAction := func(entries []interface{}, elem ...interface{}) {
			// Log entries for audit should be 8
			its.Len(entries, 8)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(
			errors.New("test-error")).Do(doAction).Times(1)

		done := make(chan bool)
		chanErr := make(chan error)
		res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(au), done, chanErr)

		// Wait for close and error channel from audit thread
		its.Error(<-chanErr)
		its.True(<-done)
		its.NoError(err)

		// 13 female in db
		its.Equal(int64(13), res.MatchedCount)
		// 8 inactive female
		its.Equal(int64(8), res.ModifiedCount)
		// 0 fails and upserted
		its.Equal(int64(0), res.FailedCount)
		its.Empty(res.FailedIDs)
		its.Equal(int64(0), res.UpsertedCount)
		its.Nil(res.UpsertedID)

		// Check with Count
		count, err := base.CountDocuments(chkFilter)
		its.NoError(err)
		its.Equal(int64(13), count)
	})
}

func TestMongoDbBaseRepo_DeleteOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)
	testUsers := setupData(db)

	t.Run("delete", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		// Test user
		tu := testUsers[5]

		filter := bson.D{{"_id", tu.Id}}
		err := base.DeleteOne(filter)
		its.NoError(err)

		// Check with Find
		var check TestUser
		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
		its.Error(err)
		its.IsType(&lxDb.NotFoundError{}, err)
	})
	t.Run("with audit", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// Test user
		tu := testUsers[6]

		// TestUser to map for tests
		doc, err := lxDb.ToBsonDoc(&tu)
		its.NoError(err)

		// AuditAuth
		au := &bson.M{"name": "Timo Liebetrau"}

		// Check mock params
		doAction := func(action string, authUser, data interface{}, elem ...interface{}) {
			its.Equal(lxDb.Delete, action)
			its.Equal(au, authUser)
			its.Equal(doc, data)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Do(doAction).Times(1)

		filter := bson.D{{"_id", tu.Id}}
		done := make(chan bool)
		err = base.DeleteOne(filter, lxDb.SetAuditAuth(au), done)

		// Wait for close channel and check err
		<-done
		its.NoError(err)

		//// Check with Find
		var check TestUser
		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
		its.Error(err)
		its.IsType(&lxDb.NotFoundError{}, err)
	})
	t.Run("with audit error", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// Test user
		tu := testUsers[7]

		// TestUser to map for tests
		doc, err := lxDb.ToBsonDoc(&tu)
		its.NoError(err)

		// AuditAuth
		au := &bson.M{"name": "Timo Liebetrau"}

		// Check mock params
		doAction := func(act string, usr, data interface{}, elem ...interface{}) {
			its.Equal(lxDb.Delete, act)
			its.Equal(au, usr)
			its.Equal(doc, data)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntry(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			errors.New("test-error")).Do(doAction).Times(1)

		filter := bson.D{{"_id", tu.Id}}
		done := make(chan bool)
		chanErr := make(chan error)
		err = base.DeleteOne(filter, lxDb.SetAuditAuth(au), done, chanErr)

		// Wait for close and error channel from audit thread
		its.Error(<-chanErr)
		its.True(<-done)

		its.NoError(err)

		//// Check with Find
		var check TestUser
		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
		its.Error(err)
		its.IsType(&lxDb.NotFoundError{}, err)
	})
	t.Run("not found error", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		filter := bson.D{{"_id", primitive.NewObjectID()}}
		err := base.DeleteOne(filter)
		its.Error(err)
		its.IsType(&lxDb.NotFoundError{}, err)
	})
	t.Run("delete with options", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		// Test user
		tu := testUsers[9]

		filter := bson.D{{"_id", tu.Id}}
		err := base.DeleteOne(filter, time.Second*10, options.FindOneAndDelete())
		its.NoError(err)

		// Check with Find
		var check TestUser
		err = base.FindOne(bson.D{{"_id", tu.Id}}, &check)
		its.Error(err)
		its.IsType(&lxDb.NotFoundError{}, err)
	})
}

func TestMongoDbBaseRepo_DeleteMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)

	// Mock for base repo
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

	// Test the base repo with mock
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection), mockIBaseRepoAudit)

	// All Males, should be 12
	filter := bson.D{{"gender", "Male"}}
	chkFilter := bson.D{{"gender", "Male"}}

	// AuditAuth user
	au := &bson.M{"name": "Timo Liebetrau"}

	t.Run("delete", func(t *testing.T) {
		setupData(db)

		res, err := base.DeleteMany(filter, time.Second*10, options.Delete())
		its.NoError(err)

		// 12 male in db
		its.Equal(int64(12), res.DeletedCount)
		// 0 fails
		its.Equal(int64(0), res.FailedCount)
		its.Empty(res.FailedIDs)

		// Check with Count
		count, err := base.CountDocuments(chkFilter)
		its.NoError(err)
		its.Equal(int64(0), count)
	})
	t.Run("with audit", func(t *testing.T) {
		setupData(db)

		// Check mock params
		doAction := func(entries []interface{}, elem ...interface{}) {
			// Log entries for audit should be 12
			its.Len(entries, 12)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(nil).Do(doAction).Times(1)

		done := make(chan bool)
		sidName := &lxDb.SubIdName{Name: "_id"}
		res, err := base.DeleteMany(filter, lxDb.SetAuditAuth(au), done, sidName)

		// Wait for close channel and check err
		<-done
		its.NoError(err)

		// 12 male in db
		its.Equal(int64(12), res.DeletedCount)
		// 0 fails
		its.Equal(int64(0), res.FailedCount)
		its.Empty(res.FailedIDs)

		// Check with Count
		count, err := base.CountDocuments(chkFilter)
		its.NoError(err)
		its.Equal(int64(0), count)
	})
	t.Run("with audit error", func(t *testing.T) {
		setupData(db)

		// Check mock params
		doAction := func(entries []interface{}, elem ...interface{}) {
			// Log entries for audit should be 12
			its.Len(entries, 12)
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().LogEntries(gomock.Any()).Return(
			errors.New("test-error")).Do(doAction).Times(1)

		done := make(chan bool)
		chanErr := make(chan error)
		res, err := base.DeleteMany(filter, lxDb.SetAuditAuth(au), done, chanErr)

		// Wait for close and error channel from audit thread
		its.Error(<-chanErr)
		its.True(<-done)

		its.NoError(err)

		// 12 male in db
		its.Equal(int64(12), res.DeletedCount)
		// 0 fails
		its.Equal(int64(0), res.FailedCount)
		its.Empty(res.FailedIDs)

		// Check with Count
		count, err := base.CountDocuments(chkFilter)
		its.NoError(err)
		its.Equal(int64(0), count)
	})
}

func TestMongoBaseRepo_GetCollection(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(collection)

	its.Equal(collection, base.GetCollection())
}

func TestMongoBaseRepo_GetDb(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

	its.Equal(db, base.GetDb())
}

func TestMongoBaseRepo_GetRepoName(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(collection)

	expect := db.Name() + "/" + collection.Name()

	its.Equal(expect, base.GetRepoName())
}
