package lxDb_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func getTestUsers() []interface{} {
	testUsers := make([]interface{}, 0)
	for i := 0; i < 10; i++ {
		testUsers = append(testUsers, TestUser{
			Name:  fmt.Sprintf("TestName_%d", i),
			Email: fmt.Sprintf("test_%d@test.de", i)})
	}

	return testUsers
}

func findIDInTestUserArr(vs []TestUser, t interface{}) int {
	for i, v := range vs {
		if v.Id == t {
			return i
		}
	}
	return -1
}

func findUserInTestUserArr(vs []TestUser, email, name string) int {
	for i, v := range vs {
		if v.Email == email && v.Name == name {
			return i
		}
	}
	return -1
}

func findUserInBsonMArr(vs []bson.M, email, name string) int {
	for i, v := range vs {
		chkEmail, ok1 := v["data"].(bson.M)["email"].(string)
		chkName, ok2 := v["data"].(bson.M)["name"].(string)

		if !ok1 || !ok2 {
			return -1
		}

		if chkEmail == email && chkName == name {
			return i
		}
	}
	return -1
}

func getTestAuditUser() *bson.M {
	return &bson.M{"name": "TestAuditUser"}
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

func TestMongoBaseRepo_InsertOne(t *testing.T) {
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

	t.Run("without_audit", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))

		// Test InsertOne
		testUser := getTestUsers()[1].(TestUser)
		res, err := base.InsertOne(&testUser, time.Second*10, options.InsertOne())
		its.NoError(err)

		// Add InsertedID to TestUser
		testUser.Id = res.(primitive.ObjectID)

		// Check insert result id
		var checkUser TestUser
		filter := bson.D{{"_id", testUser.Id}}
		its.NoError(base.FindOne(filter, &checkUser))

		its.Equal(testUser, checkUser)
	})
	t.Run("with_audit", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))

		// Test data
		testUser := getTestUsers()[1].(TestUser)
		auditUser := getTestAuditUser()

		// Check mock params
		var chkAuditDataUserId primitive.ObjectID
		doAction := func(elem interface{}) {
			switch val := elem.(type) {
			case bson.M:
				// Check user id
				chkAuditDataUserId = val["data"].(bson.M)["_id"].(primitive.ObjectID)

				// Check
				its.Equal(TestCollection, val["collection"])
				its.Equal(lxDb.Insert, val["action"].(string))
				its.Equal(auditUser, val["user"])
				its.Equal(testUser.Email, val["data"].(bson.M)["email"])
				its.Equal(testUser.Name, val["data"].(bson.M)["name"])
			default:
				t.Fail()
			}
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
		mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

		// Test InsertOne
		res, err := base.InsertOne(&testUser, lxDb.SetAuditAuth(auditUser))
		its.NoError(err)

		// Add InsertedID to TestUser and check with audit
		testUser.Id = res.(primitive.ObjectID)
		its.Equal(testUser.Id, chkAuditDataUserId)

		// Check insert user
		var checkUser TestUser
		filter := bson.D{{"_id", testUser.Id}}
		its.NoError(base.FindOne(filter, &checkUser))
		its.Equal(testUser, checkUser)
	})
}

func TestMongoBaseRepo_InsertMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Mock for base repo
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

	// Test the base repo with mock
	base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

	t.Run("without_audit", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))
		testUsers := getTestUsers()

		// Test the base repo
		res, err := base.InsertMany(testUsers, time.Second*10, options.InsertMany())
		its.NoError(err)
		its.Equal(10, len(res.InsertedIDs))
		its.Equal(int64(0), res.FailedCount)

		// Find and compare
		var checkUsers []TestUser
		err = base.Find(bson.D{}, &checkUsers)
		its.NoError(err)

		// Check users, find test user in checkUsers
		for _, v := range testUsers {
			its.True(findUserInTestUserArr(checkUsers, v.(TestUser).Email, v.(TestUser).Name) > -1)
		}

		// Check ids
		for _, id := range res.InsertedIDs {
			its.True(findIDInTestUserArr(checkUsers, id) > -1)
		}
	})

	t.Run("with_audit", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))

		// Test data
		testUsers := getTestUsers()
		auditUser := getTestAuditUser()

		doAction := func(elem interface{}) {
			switch val := elem.(type) {
			case []bson.M:
				its.Len(val, 10)
				its.Equal(TestCollection, val[1]["collection"])
				its.Equal(lxDb.Insert, val[1]["action"])
				its.Equal(auditUser, val[1]["user"])

				// Check users
				// Find user in test users arr
				for _, v := range testUsers {
					its.True(findUserInBsonMArr(val, v.(TestUser).Email, v.(TestUser).Name) > -1)
				}
			default:
				t.Fail()
			}
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
		mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

		// Test InsertMany
		res, err := base.InsertMany(testUsers, lxDb.SetAuditAuth(auditUser))
		its.NoError(err)
		its.Equal(10, len(res.InsertedIDs))
		its.Equal(int64(0), res.FailedCount)

		// Find users after insert and compare
		var checkUsers []TestUser
		err = base.Find(bson.D{}, &checkUsers)
		its.NoError(err)

		// Check users
		// Find check user in test users arr
		for _, v := range testUsers {
			its.True(findUserInTestUserArr(checkUsers, v.(TestUser).Email, v.(TestUser).Name) > -1)
		}

		// Find result id in check users
		for _, id := range res.InsertedIDs {
			its.True(findIDInTestUserArr(checkUsers, id) > -1)
		}
	})

	t.Run("empty_audits", func(t *testing.T) {
		// Drop for test
		its.NoError(collection.Drop(context.Background()))

		testUsers := make([]interface{}, 0)
		auditUser := getTestAuditUser()

		// Configure mock
		mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

		// Test the base repo
		res, err := base.InsertMany(testUsers, lxDb.SetAuditAuth(auditUser))

		its.NoError(err)
		its.Equal(0, len(res.InsertedIDs))
		its.Equal(int64(0), res.FailedCount)
	})
}

func TestMongoBaseRepo_CountDocuments(t *testing.T) {
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

func TestMongoBaseRepo_EstimatedDocumentCount(t *testing.T) {
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

func TestMongoBaseRepo_Find(t *testing.T) {
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

func TestMongoBaseRepo_FindOne(t *testing.T) {
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
		its.True(errors.Is(err, lxDb.ErrNotFound))
	})
}

func TestMongoBaseRepo_FindOneAndDelete(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	// Sorted by name
	testUsers := setupData(db)

	t.Run("with_options", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		// Find options in other format
		// Sort by name as option
		fo := lxHelper.FindOptions{
			Sort: map[string]int{"name": 1},
		}

		// Empty filter for find first entry
		filter := bson.D{}
		var result TestUser

		// Find first entry sort by name and delete
		err = base.FindOneAndDelete(filter, &result, fo.ToMongoFindOneAndDeleteOptions(), time.Second*15)
		its.NoError(err)
		its.Equal(testUsers[0], result)
	})
	t.Run("with_filter", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		filter := bson.D{{"email", testUsers[5].Email}}
		var result TestUser
		err = base.FindOneAndDelete(filter, &result)
		its.NoError(err)
		its.Equal(testUsers[5], result)
	})
	t.Run("not_found", func(t *testing.T) {
		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection)

		filter := bson.D{{"email", "unknown@email"}}
		var result TestUser
		err = base.FindOneAndDelete(filter, &result)
		its.Error(err)
		its.True(errors.Is(err, lxDb.ErrNotFound))
	})
	t.Run("with_audit", func(t *testing.T) {
		// Test the base repo with mock
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

		// Test the base repo
		base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

		// Test data
		testUser := testUsers[6]
		auditUser := getTestAuditUser()

		// Check mock params
		doAction := func(elem interface{}) {
			switch val := elem.(type) {
			case bson.M:
				its.Equal(TestCollection, val["collection"])
				its.Equal(lxDb.Delete, val["action"].(string))
				its.Equal(auditUser, val["user"])
				its.Equal(testUser.Id, val["data"].(bson.M)["_id"].(primitive.ObjectID))
			default:
				t.Fail()
			}
		}

		// Configure mock
		mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
		mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

		// Find and delete
		var result TestUser
		filter := bson.D{{"_id", testUser.Id}}
		err = base.FindOneAndDelete(filter, &result, lxDb.SetAuditAuth(auditUser))
		its.NoError(err)

		// Check result
		its.Equal(testUser, result)

		// Check with find, user should be deleted
		var check TestUser
		err = base.FindOne(bson.D{{"_id", testUser.Id}}, &check)
		its.Error(err)
		its.True(errors.Is(err, lxDb.ErrNotFound))
	})
}

func TestMongoBaseRepo_FindOneAndReplace(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	t.Run("without_audit", func(t *testing.T) {
		t.Run("options_after", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)
			tu := testUsers[0]

			// without is_active
			replaceUser := TestUser{
				Id:     tu.Id,
				Name:   "NewReplacementUser",
				Gender: tu.Gender,
				Email:  tu.Email,
			}

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			opts.SetReturnDocument(options.After)

			// Find first entry sort by name and replace it
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, time.Second*10)
			its.NoError(err)
			its.Equal(replaceUser, result)
		})
		t.Run("options_before", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)
			tu := testUsers[0]

			// without is_active
			replaceUser := TestUser{
				Id:     tu.Id,
				Name:   "NewReplacementUser",
				Gender: tu.Gender,
				Email:  tu.Email,
			}

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts)
			its.NoError(err)
			its.Equal(tu, result)
		})
		t.Run("without_options", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)
			tu := testUsers[0]

			// without is_active
			replaceUser := TestUser{
				Id:     tu.Id,
				Name:   "NewReplacementUser",
				Gender: tu.Gender,
				Email:  tu.Email,
			}

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			//opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and replace
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts)
			its.NoError(err)
			its.Equal(tu, result)
		})
		t.Run("error_not_found", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)
			tu := testUsers[0]

			// without is_active
			replaceUser := TestUser{
				Id:     tu.Id,
				Name:   "NewReplacementUser",
				Gender: tu.Gender,
				Email:  tu.Email,
			}

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{{"_id", primitive.NewObjectID()}}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
	t.Run("with_audit", func(t *testing.T) {
		t.Run("SetReturnDocument_before", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			testUser := testUsers[0]
			auditUser := getTestAuditUser()

			// without is_active
			replaceUser := TestUser{
				Id:     testUser.Id,
				Name:   "NewReplacementUser",
				Gender: testUser.Gender,
				Email:  testUser.Email,
			}

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Check mock params
			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case bson.M:
					its.Equal(TestCollection, val["collection"])
					its.Equal(lxDb.Update, val["action"].(string))
					its.Equal(auditUser, val["user"])

					// Map for check
					chkUser, err := lxDb.ToBsonMap(replaceUser)
					its.NoError(err)
					its.Equal(chkUser, val["data"])
				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and replace
			// Empty filter for find first entry
			var result TestUser
			filter := bson.D{}
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(auditUser))

			// Check
			its.NoError(err)
			its.Equal(testUser, result)
		})
		t.Run("SetReturnDocument_before_error_not_found", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			testUser := testUsers[0]
			auditUser := getTestAuditUser()

			// without is_active
			fakeId := primitive.NewObjectID()
			replaceUser := TestUser{
				Id:     fakeId,
				Name:   "NewReplacementUser",
				Gender: testUser.Gender,
				Email:  testUser.Email,
			}

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and replace
			// Empty filter for find first entry
			var result TestUser
			filter := bson.D{{"_id", fakeId}}
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(auditUser))

			// Check err
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
		// Should be use options.After by default
		t.Run("SetReturnDocument_without_options", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			testUser := testUsers[0]
			auditUser := getTestAuditUser()

			// without is_active
			replaceUser := TestUser{
				Id:     testUser.Id,
				Name:   "NewReplacementUser",
				Gender: testUser.Gender,
				Email:  testUser.Email,
			}

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Check mock params
			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case bson.M:
					its.Equal(TestCollection, val["collection"])
					its.Equal(lxDb.Update, val["action"].(string))
					its.Equal(auditUser, val["user"])

					// Map for check
					chkUser, err := lxDb.ToBsonMap(replaceUser)
					its.NoError(err)
					its.Equal(chkUser, val["data"])
				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			// Should be use options.After by default
			//opts.SetReturnDocument(options.After)

			// Find first entry sort by name and replace
			// Empty filter for find first entry
			var result TestUser
			filter := bson.D{}
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(auditUser))

			// Check err
			its.NoError(err)
			its.Equal(replaceUser, result)
		})
		t.Run("SetReturnDocument_after_error_not_found", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			testUser := testUsers[0]
			auditUser := getTestAuditUser()

			// without is_active
			fakeId := primitive.NewObjectID()
			replaceUser := TestUser{
				Id:     fakeId,
				Name:   "NewReplacementUser",
				Gender: testUser.Gender,
				Email:  testUser.Email,
			}

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndReplaceOptions()
			opts.SetReturnDocument(options.After)

			// Find first entry sort by name and replace
			// Empty filter for find first entry
			filter := bson.D{{"_id", fakeId}}
			//done := make(chan bool)
			var result TestUser
			err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(auditUser))

			// Wait for close channel and check err
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
}

func TestMongoBaseRepo_FindOneAndUpdate(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	t.Run("without_audits", func(t *testing.T) {
		t.Run("with options after", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{}
			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			opts.SetReturnDocument(options.After)

			// Find first entry sort by name and update
			err = base.FindOneAndUpdate(filter, update, &result, opts, time.Second*10)
			its.NoError(err)

			expected := TestUser{
				Id:       testUsers[0].Id,
				Name:     "updated-name",
				Gender:   testUsers[0].Gender,
				Email:    testUsers[0].Email,
				IsActive: testUsers[0].IsActive,
			}

			its.Equal(expected, result)
		})
		t.Run("with options before", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{}
			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			err = base.FindOneAndUpdate(filter, update, &result, opts)
			its.NoError(err)
			its.Equal(testUsers[0], result)
		})
		t.Run("without options", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{}
			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			//opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			err = base.FindOneAndUpdate(filter, update, &result, opts)
			its.NoError(err)

			// Should be equal with before option
			its.Equal(testUsers[0], result)
		})
		t.Run("not found error", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			setupData(db)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Empty filter for find first entry
			filter := bson.D{{"_id", primitive.NewObjectID()}}
			update := bson.D{{"$set", bson.D{{"name", "updated-name"}}}}
			var result TestUser

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			err = base.FindOneAndUpdate(filter, update, &result, opts)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
	t.Run("with_audits", func(t *testing.T) {
		t.Run("SetReturnDocument_before", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// New name for update
			updatedName := "updatedName"

			// Check mock params
			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case bson.M:
					its.Equal(TestCollection, val["collection"])
					its.Equal(lxDb.Update, val["action"].(string))
					its.Equal(auditUser, val["user"])

					// Map for check
					testUser := testUsers[0]
					testUser.Name = updatedName
					chkUser, err := lxDb.ToBsonMap(testUser)
					its.NoError(err)
					its.Equal(chkUser, val["data"])
				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			// Empty filter for find first entry
			var result TestUser
			filter := bson.D{}
			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(auditUser))

			// Check
			its.NoError(err)
			its.Equal(testUsers[0], result)
		})
		t.Run("SetReturnDocument_before_error_not_found", func(t *testing.T) {
			// testUsers Sorted by name
			setupData(db)
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			// New name for update
			updatedName := "updatedName"

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			opts.SetReturnDocument(options.Before)

			// Find first entry sort by name and update
			// Empty filter for find first entry
			filter := bson.D{{"_id", primitive.NewObjectID()}}
			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
			var result TestUser
			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(auditUser))

			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
		t.Run("without_SetReturnDocument", func(t *testing.T) {
			// When SetReturnDocument not set, use options.After by default
			// testUsers Sorted by name
			testUsers := setupData(db)
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// New name for update
			updatedName := "updatedName"

			// Check mock params
			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case bson.M:
					its.Equal(TestCollection, val["collection"])
					its.Equal(lxDb.Update, val["action"].(string))
					its.Equal(auditUser, val["user"])

					// Map for check
					testUser := testUsers[0]
					testUser.Name = updatedName
					chkUser, err := lxDb.ToBsonMap(testUser)
					its.NoError(err)
					its.Equal(chkUser, val["data"])
				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			// When SetReturnDocument not set, use options.After by default
			//opts.SetReturnDocument(options.After)

			// Find first entry sort by name and update
			// Empty filter for find first entry
			var result TestUser
			filter := bson.D{}
			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(auditUser))

			// Check
			its.NoError(err)

			// Test user for check
			tu := testUsers[0]
			tu.Name = updatedName

			its.Equal(tu, result)
		})
		t.Run("SetReturnDocument_after_error_not_found", func(t *testing.T) {
			// testUsers Sorted by name
			setupData(db)
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			// New name for update
			updatedName := "updatedName"

			// Find options in other format
			// Sort by name as option
			fo := lxHelper.FindOptions{
				Sort: map[string]int{"name": 1},
			}

			// Set options and return document
			opts := fo.ToMongoFindOneAndUpdateOptions()
			opts.SetReturnDocument(options.After)

			// Find first entry sort by name and update
			// filter with new id for not found error
			filter := bson.D{{"_id", primitive.NewObjectID()}}
			update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
			var result TestUser
			err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(auditUser))

			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
}

func TestMongoBaseRepo_UpdateOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	t.Run("without_audit", func(t *testing.T) {
		t.Run("update", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// testUsers Sorted by name
			testUsers := setupData(db)

			// Update testUser
			testUser := testUsers[5]
			newName := "Is Updated"

			filter := bson.D{{"_id", testUser.Id}}
			update := bson.D{{"$set", bson.D{{"name", newName}}}}
			err := base.UpdateOne(filter, update, time.Second*10, options.Update())
			its.NoError(err)

			// Check with Find
			var check TestUser
			its.NoError(base.FindOne(bson.D{{"_id", testUser.Id}}, &check))
			its.Equal(newName, check.Name)
		})
		t.Run("error_not_found", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection))

			// Update testUser
			newName := "Is Updated"

			// New id for not found error
			filter := bson.D{{"_id", primitive.NewObjectID()}}
			update := bson.D{{"$set", bson.D{{"name", newName}}}}
			err := base.UpdateOne(filter, update)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
	t.Run("with_audit", func(t *testing.T) {
		t.Run("update", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			auditUser := getTestAuditUser()
			testUser := testUsers[6]

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Update testUser
			testUser.Name = "Is Updated"

			// Check mock params
			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case bson.M:
					its.Equal(TestCollection, val["collection"])
					its.Equal(lxDb.Update, val["action"].(string))
					its.Equal(auditUser, val["user"])

					// Map for check
					chkUser, err := lxDb.ToBsonMap(testUser)
					its.NoError(err)
					its.Equal(chkUser, val["data"])
				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			filter := bson.D{{"_id", testUser.Id}}
			update := bson.D{{"$set", bson.D{{"name", testUser.Name}}}}
			err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(auditUser))

			// Check
			its.NoError(err)

			// Check with Find
			var check TestUser
			its.NoError(base.FindOne(bson.D{{"_id", testUser.Id}}, &check))
			its.Equal(testUser.Name, check.Name)
		})
		t.Run("error_not_found", func(t *testing.T) {
			// testUsers Sorted by name
			testUsers := setupData(db)
			auditUser := getTestAuditUser()
			testUser := testUsers[6]

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Update testUser
			testUser.Name = "Is Updated"

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			// New id for not found
			filter := bson.D{{"_id", primitive.NewObjectID()}}
			update := bson.D{{"$set", bson.D{{"name", testUser.Name}}}}
			err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(auditUser))
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
}

func TestMongoDbBaseRepo_UpdateMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	t.Run("without_audit", func(t *testing.T) {
		t.Run("update_many", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// Setup test data
			setupData(db)

			// All females to active, should be 13
			filter := bson.D{{"gender", "Female"}}
			update := bson.D{{"$set",
				bson.D{{"is_active", true}},
			}}
			chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}

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
	})
	t.Run("with_audit", func(t *testing.T) {
		t.Run("update_many", func(t *testing.T) {
			// testUsers test data Sorted by name
			setupData(db)
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case []bson.M:
					its.Len(val, 8)
					its.Equal(TestCollection, val[1]["collection"])
					its.Equal(lxDb.Update, val[1]["action"])
					its.Equal(auditUser, val[1]["user"])

					// Check val arr should only female with active true
					for _, v := range val {
						its.Equal("Female", v["data"].(bson.M)["gender"].(string))
						its.True(v["data"].(bson.M)["is_active"].(bool))
					}

				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			// All females to active,
			// All females should be 13
			// All inactive females should be 8
			filter := bson.D{{"gender", "Female"}}
			chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}
			update := bson.D{{"$set",
				bson.D{{"is_active", true}},
			}}
			res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(auditUser))

			// Check
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
		t.Run("empty_updates", func(t *testing.T) {
			// testUsers test data Sorted by name
			setupData(db)
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			// Wrong filter for 0 matches
			filter := bson.D{{"gender", "Animal"}}
			update := bson.D{{"$set",
				bson.D{{"is_active", true}},
			}}

			res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(auditUser))
			its.NoError(err)

			// 0 animals in db
			its.Equal(int64(0), res.MatchedCount)
			// 0 modified
			its.Equal(int64(0), res.ModifiedCount)
			// 0 fails and upserted
			its.Equal(int64(0), res.FailedCount)
			its.Empty(res.FailedIDs)
			its.Equal(int64(0), res.UpsertedCount)
			its.Nil(res.UpsertedID)
		})
	})
}

func TestMongoBaseRepo_DeleteOne(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	t.Run("without_audit", func(t *testing.T) {
		t.Run("delete_one", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// Setup test data
			testUsers := setupData(db)
			testUser := testUsers[5]

			filter := bson.D{{"_id", testUser.Id}}
			err := base.DeleteOne(filter)
			its.NoError(err)

			// Check with Find
			var check TestUser
			err = base.FindOne(bson.D{{"_id", testUser.Id}}, &check)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
		t.Run("error_not_found", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			filter := bson.D{{"_id", primitive.NewObjectID()}}
			err := base.DeleteOne(filter)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
		t.Run("with_options", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)

			// Test user
			testUsers := setupData(db)
			testUser := testUsers[9]

			filter := bson.D{{"_id", testUser.Id}}
			err := base.DeleteOne(filter, time.Second*10, options.FindOneAndDelete())
			its.NoError(err)

			// Check with Find
			var check TestUser
			err = base.FindOne(bson.D{{"_id", testUser.Id}}, &check)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
	t.Run("with_audit", func(t *testing.T) {
		t.Run("delete_one", func(t *testing.T) {
			// Setup test data
			testUsers := setupData(db)
			testUser := testUsers[6]
			auditUser := getTestAuditUser()

			// Test the base repo with mock
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection, mockIBaseRepoAudit)

			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case bson.M:
					its.Equal(TestCollection, val["collection"])
					its.Equal(lxDb.Delete, val["action"])
					its.Equal(auditUser, val["user"])
					its.Equal(testUser.Id, val["data"].(bson.M)["_id"].(primitive.ObjectID))

				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			filter := bson.D{{"_id", testUser.Id}}
			err = base.DeleteOne(filter, lxDb.SetAuditAuth(auditUser))

			// Check
			its.NoError(err)

			// Check with Find
			var check TestUser
			err = base.FindOne(bson.D{{"_id", testUser.Id}}, &check)
			its.Error(err)
			its.True(errors.Is(err, lxDb.ErrNotFound))
		})
	})
}

func TestMongoDbBaseRepo_DeleteMany(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	collection := db.Collection(TestCollection)

	t.Run("without_audit", func(t *testing.T) {
		t.Run("delete_many", func(t *testing.T) {
			// Test the base repo
			base := lxDb.NewMongoBaseRepo(collection)
			setupData(db)

			// All Males, should be 12
			filter := bson.D{{"gender", "Male"}}
			chkFilter := bson.D{{"gender", "Male"}}
			res, err := base.DeleteMany(filter, time.Second*10, options.Delete())
			its.NoError(err)

			// 12 male in db
			its.Equal(int64(12), res.DeletedCount)

			// Check with Count
			count, err := base.CountDocuments(chkFilter)
			its.NoError(err)
			its.Equal(int64(0), count)
		})
	})
	t.Run("with_audit", func(t *testing.T) {
		t.Run("delete_many", func(t *testing.T) {
			setupData(db)
			auditUser := getTestAuditUser()

			// Mock for base repo
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo with mock
			base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection), mockIBaseRepoAudit)

			doAction := func(elem interface{}) {
				switch val := elem.(type) {
				case []bson.M:
					// Log entries for audit should be 12
					its.Len(val, 12)
					its.Equal(TestCollection, val[1]["collection"])
					its.Equal(lxDb.Delete, val[1]["action"])
					its.Equal(auditUser, val[1]["user"])
				default:
					t.Fail()
				}
			}

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)
			mockIBaseRepoAudit.EXPECT().Send(gomock.Any()).Return().Do(doAction).Times(1)

			// All Males, should be 12
			filter := bson.D{{"gender", "Male"}}
			chkFilter := bson.D{{"gender", "Male"}}
			res, err := base.DeleteMany(filter, lxDb.SetAuditAuth(auditUser))

			// Check
			its.NoError(err)

			// 12 male in db
			its.Equal(int64(12), res.DeletedCount)

			// Check with Count
			count, err := base.CountDocuments(chkFilter)
			its.NoError(err)
			its.Equal(int64(0), count)
		})
		t.Run("empty_filter", func(t *testing.T) {
			setupData(db)
			auditUser := getTestAuditUser()

			// Mock for base repo
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			mockIBaseRepoAudit := lxDbMocks.NewMockIBaseRepoAudit(mockCtrl)

			// Test the base repo with mock
			base := lxDb.NewMongoBaseRepo(db.Collection(TestCollection), mockIBaseRepoAudit)

			// Configure mock
			mockIBaseRepoAudit.EXPECT().IsActive().Return(true).Times(1)

			res, err := base.DeleteMany(bson.D{{"gender", "Animals"}}, lxDb.SetAuditAuth(auditUser))
			its.NoError(err)
			// 0 animals in db
			its.Equal(int64(0), res.DeletedCount)
		})
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

func TestMongoDbBaseRepo_Aggregate(t *testing.T) {
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

	t.Run("with $skip/$limit", func(t *testing.T) {
		// Build mongo pipeline
		pipeline := mongo.Pipeline{
			bson.D{{
				"$match", bson.M{},
			}},
			bson.D{{
				"$sort", bson.M{
					"name": 1,
				},
			}},
			bson.D{{
				"$skip", 5,
			}},
			bson.D{{
				"$limit", 5,
			}},
		}

		var result []TestUser
		err = base.Aggregate(pipeline, &result, time.Second*10)

		its.NoError(err)
		its.Equal(expectUsers, result)
	})
	t.Run("with $match", func(t *testing.T) {
		pipeline := mongo.Pipeline{
			bson.D{{
				"$match", bson.D{{"gender", "Female"}},
			}},
			bson.D{{
				"$sort", bson.M{
					"name": 1,
				},
			}},
		}

		var result []TestUser
		err = base.Aggregate(pipeline, &result)
		its.NoError(err)
		its.Equal(expectFemale, result)
	})
	t.Run("without result", func(t *testing.T) {
		pipeline := mongo.Pipeline{
			bson.D{{
				"$match", bson.D{{"gender", "Female"}},
			}},
			bson.D{{
				"$sort", bson.M{
					"name": 1,
				},
			}},
		}

		err = base.Aggregate(pipeline, nil)
		its.NoError(err)
	})
}

func TestMongoDbBaseRepo_LocaleTest(t *testing.T) {
	its := assert.New(t)

	client, err := lxDb.GetMongoDbClient(dbHost)
	its.NoError(err)

	db := client.Database(TestDbName)
	testUsers := setupDataLocale(db)

	// sort by index that's will be ordered by username
	sort.Slice(testUsers[:], func(i, j int) bool {
		return testUsers[i].Index < testUsers[j].Index
	})

	// Test the base repo
	base := lxDb.NewMongoBaseRepo(db.Collection(TestCollectionLocale))
	base.SetLocale("de")

	t.Run("with find and locale set to 'de' and ordered by 'username'", func(t *testing.T) {
		// Find options in other format
		fo := lxHelper.FindOptions{
			Sort: map[string]int{"username": 1},
		}

		filter := bson.M{}

		var result []TestUserLocale
		err = base.Find(filter, &result, fo.ToMongoFindOptions(), time.Second*10)

		its.NoError(err)
		its.Equal(testUsers, result)
	})
}

/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

// TODO TestMongoBaseRepo_InsertOne2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_InsertOne2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("with_audit", func(t *testing.T) {
//		// Drop for test
//		its.NoError(collection.Drop(context.Background()))
//
//		// Test data
//		testUser := getTestUsers()[1].(TestUser)
//		auditUser := getTestAuditUser()
//
//		// Test InsertOne
//		res, err := base.InsertOne(&testUser, lxDb.SetAuditAuth(auditUser))
//
//		// wait for worker
//		its.NoError(<-audit.ErrChan)
//		its.NoError(err)
//
//		// Add InsertedID to TestUser and check with audit
//		testUser.Id = res.(primitive.ObjectID)
//
//		// Check insert user
//		var checkUser TestUser
//		filter := bson.D{{"_id", testUser.Id}}
//		its.NoError(base.FindOne(filter, &checkUser))
//		its.Equal(testUser, checkUser)
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_InsertMany2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_InsertMany2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("with_audit", func(t *testing.T) {
//		// Drop for test
//		its.NoError(collection.Drop(context.Background()))
//
//		// Test data
//		testUsers := getTestUsers()
//		auditUser := getTestAuditUser()
//
//		// Test InsertMany
//		res, err := base.InsertMany(testUsers, lxDb.SetAuditAuth(auditUser))
//
//		// wait for worker
//		its.NoError(<-audit.ErrChan)
//
//		// check
//		its.NoError(err)
//		its.Equal(10, len(res.InsertedIDs))
//		its.Equal(int64(0), res.FailedCount)
//
//		// Find users after insert and compare
//		var checkUsers []TestUser
//		err = base.Find(bson.D{}, &checkUsers)
//		its.NoError(err)
//
//		// Check users
//		// Find check user in test users arr
//		for _, v := range testUsers {
//			its.True(findUserInTestUserArr(checkUsers, v.(TestUser).Email, v.(TestUser).Name) > -1)
//		}
//
//		// Find result id in check users
//		for _, id := range res.InsertedIDs {
//			its.True(findIDInTestUserArr(checkUsers, id) > -1)
//		}
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_FindOneAndDelete2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_FindOneAndDelete2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Sorted by name
//	testUsers := setupData(db)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("with_audit", func(t *testing.T) {
//		// Test data
//		testUser := testUsers[6]
//		t.Log("should be deleted", testUser)
//		auditUser := getTestAuditUser()
//
//		// Test find and delete
//		var result TestUser
//		filter := bson.D{{"_id", testUser.Id}}
//		err = base.FindOneAndDelete(filter, &result, lxDb.SetAuditAuth(auditUser))
//
//		// wait for worker
//		its.NoError(<-audit.ErrChan)
//		its.NoError(err)
//
//		// Check result
//		its.Equal(testUser, result)
//
//		// Check with find, user should be deleted
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", testUser.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_FindOneAndReplace2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_FindOneAndReplace2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Sorted by name
//	testUsers := setupData(db)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("options_before", func(t *testing.T) {
//		// Test data
//		testUser := testUsers[6]
//		auditUser := getTestAuditUser()
//
//		// replaced user
//		replaceUser := TestUser{
//			Id: testUser.Id,
//			Name:   "NewReplacementUser",
//			Gender: testUser.Gender,
//			Email:  testUser.Email,
//			IsActive: testUser.IsActive,
//		}
//
//		// Find options in other format
//		// Sort by name as option
//		fo := lxHelper.FindOptions{
//			Sort: map[string]int{"name": 1},
//		}
//
//		// Set options and return document
//		opts := fo.ToMongoFindOneAndReplaceOptions()
//		opts.SetReturnDocument(options.After)
//
//		// Find first entry sort by name and replace
//		// Empty filter for find first entry
//		var result TestUser
//		filter := bson.D{{"email", testUser.Email}}
//		err = base.FindOneAndReplace(filter, replaceUser, &result, opts, lxDb.SetAuditAuth(auditUser))
//		its.NoError(err)
//		if err == nil {
//			// wait for worker
//			its.NoError(<-audit.ErrChan)
//		}
//
//		// Check
//		its.Equal(replaceUser, result)
//
//		t.Log("old: ", testUser)
//		t.Log("new: ", replaceUser)
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_FindOneAndUpdate2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_FindOneAndUpdate2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Sorted by name
//	testUsers := setupData(db)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("SetReturnDocument_before", func(t *testing.T) {
//		auditUser := getTestAuditUser()
//
//		// New name for update
//		updatedName := "updatedName"
//
//		// Find options in other format
//		// Sort by name as option
//		fo := lxHelper.FindOptions{
//			Sort: map[string]int{"name": 1},
//		}
//
//		// Set options and return document
//		opts := fo.ToMongoFindOneAndUpdateOptions()
//		opts.SetReturnDocument(options.Before)
//
//		// Find first entry sort by name and update
//		// Empty filter for find first entry
//		var result TestUser
//		filter := bson.D{}
//		update := bson.D{{"$set", bson.D{{"name", updatedName}}}}
//		err = base.FindOneAndUpdate(filter, update, &result, opts, lxDb.SetAuditAuth(auditUser))
//		its.NoError(err)
//		if err == nil {
//			// wait for worker
//			its.NoError(<-audit.ErrChan)
//		}
//
//		// Check
//		its.Equal(testUsers[0], result)
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_UpdateOne2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_UpdateOne2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Sorted by name
//	testUsers := setupData(db)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("update", func(t *testing.T) {
//		// testUsers Sorted by name
//		auditUser := getTestAuditUser()
//		testUser := testUsers[6]
//
//		// Update testUser
//		testUser.Name = "Is Updated"
//
//		filter := bson.D{{"_id", testUser.Id}}
//		update := bson.D{{"$set", bson.D{{"name", testUser.Name}}}}
//		err = base.UpdateOne(filter, update, lxDb.SetAuditAuth(auditUser))
//		its.NoError(err)
//		if err == nil {
//			// wait for worker
//			its.NoError(<-audit.ErrChan)
//		}
//
//		// Check with Find
//		var check TestUser
//		its.NoError(base.FindOne(bson.D{{"_id", testUser.Id}}, &check))
//		its.Equal(testUser.Name, check.Name)
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_UpdateMany2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_UpdateMany2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Sorted by name
//	setupData(db)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("update_many", func(t *testing.T) {
//		// testUsers test data Sorted by name
//		auditUser := getTestAuditUser()
//
//		// All females to active,
//		// All females should be 13
//		// All inactive females should be 8
//		filter := bson.D{{"gender", "Female"}}
//		chkFilter := bson.D{{"gender", "Female"}, {"is_active", true}}
//		update := bson.D{{"$set",
//			bson.D{{"is_active", true}},
//		}}
//		res, err := base.UpdateMany(filter, update, lxDb.SetAuditAuth(auditUser))
//		its.NoError(err)
//		if err == nil {
//			// wait for worker
//			its.NoError(<-audit.ErrChan)
//		}
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
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_DeleteOne2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_DeleteOne2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("delete_one", func(t *testing.T) {
//		// Setup test data
//		testUsers := setupData(db)
//		testUser := testUsers[6]
//		auditUser := getTestAuditUser()
//
//		filter := bson.D{{"_id", testUser.Id}}
//		err = base.DeleteOne(filter, lxDb.SetAuditAuth(auditUser))
//		its.NoError(err)
//		if err == nil {
//			// wait for worker
//			its.NoError(<-audit.ErrChan)
//		}
//
//		// Check with Find
//		var check TestUser
//		err = base.FindOne(bson.D{{"_id", testUser.Id}}, &check)
//		its.Error(err)
//		its.True(errors.Is(err, lxDb.ErrNotFound))
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}

// TODO TestMongoBaseRepo_DeleteMany2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestMongoBaseRepo_DeleteMany2(t *testing.T) {
//	its := assert.New(t)
//
//	dbHost := "mongodb://127.0.0.1"
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	client, err := lxDb.GetMongoDbClient(dbHost)
//	lxHelper.HandlePanicErr(err)
//
//	db := client.Database(TestDbName)
//	collection := db.Collection(TestCollection)
//
//	// Start audit worker
//	logEntry := lxLog.GetLogger().WithField("client", "test")
//	audit := lxAudit.NewQueue("test-client", auditHost, authKey, logEntry)
//
//	// StartWorker with ErrChan param for testing
//	numWorkers := 2
//	for i := 0; i < numWorkers; i++ {
//		audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//	}
//
//	// Test the base repo with mock
//	base := lxDb.NewMongoBaseRepo(collection, audit)
//
//	t.Run("delete_many", func(t *testing.T) {
//		setupData(db)
//		auditUser := getTestAuditUser()
//
//		// All Males, should be 12
//		filter := bson.D{{"gender", "Male"}}
//		chkFilter := bson.D{{"gender", "Male"}}
//		res, err := base.DeleteMany(filter, lxDb.SetAuditAuth(auditUser))
//		its.NoError(err)
//		if err == nil {
//			// wait for worker
//			its.NoError(<-audit.ErrChan)
//		}
//
//		// 12 male in db
//		its.Equal(int64(12), res.DeletedCount)
//
//		// Check with Count
//		count, err := base.CountDocuments(chkFilter)
//		its.NoError(err)
//		its.Equal(int64(0), count)
//	})
//
//	// Stop the worker
//	close(audit.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//}
