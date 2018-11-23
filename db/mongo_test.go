package lxDb_test

import (
	"encoding/json"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/litixsoft/lxgo/db"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"sort"
	"testing"
)

// TestUser, struct for test users
type TestUser struct {
	Id       bson.ObjectId `json:"id" bson:"_id"`
	Name     string        `json:"name" bson:"name"`
	Gender   string        `json:"gender" bson:"gender"`
	Email    string        `json:"email" bson:"email"`
	IsActive bool          `json:"is_active" bson:"is_active"`
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
		dbHost = "localhost:27017"
	}

	log.Println("DB_HOST:", dbHost)
}

// setupData, create the test data and prepare the database
func setupData(conn *mgo.Session) []TestUser {
	// Delete collection if exists
	conn.DB(TestDbName).C(TestCollection).DropCollection()

	// Setup indexes
	indexes := []mgo.Index{
		{Key: []string{"name"}},
		{Key: []string{"email"}, Unique: true},
	}

	// Ensure indexes
	col := conn.DB(TestDbName).C(TestCollection)
	for _, i := range indexes {
		if err := col.EnsureIndex(i); err != nil {
			log.Fatal(err)
		}
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
		users[i].Id = bson.NewObjectId()

		// Insert user
		if err := conn.DB(TestDbName).C(TestCollection).Insert(users[i]); err != nil {
			log.Fatal(err)
		}
	}

	// Sort test users with id for compare
	sort.Slice(users[:], func(i, j int) bool {
		return users[i].Id < users[j].Id
	})

	// Return mongo connection
	return users
}

func TestNewMongoDb(t *testing.T) {
	conn := lxDb.GetMongoDbConnection(dbHost)
	defer conn.Close()

	t.Run("correct type", func(t *testing.T) {
		db := lxDb.NewMongoDb(conn, TestDbName, TestCollection)
		chkT := reflect.TypeOf(db)
		assert.Equal(t, "*lxDb.MongoDb", chkT.String())
	})

	t.Run("test query", func(t *testing.T) {
		db := lxDb.NewMongoDb(conn, TestDbName, TestCollection)

		// setup data
		testUsers := setupData(conn)

		var result []TestUser
		assert.NoError(t, db.Conn.DB(db.Name).C(db.Collection).Find(nil).All(&result))

		// Sort result for compare
		sort.Slice(result[:], func(i, j int) bool {
			return result[i].Id < result[j].Id
		})

		assert.Equal(t, testUsers, result)
	})
}

func TestMongoDb_Setup(t *testing.T) {
	conn := lxDb.GetMongoDbConnection(dbHost)
	defer conn.Close()

	// Delete collection if exists
	conn.DB(TestDbName).C(TestCollection).DropCollection()

	db := lxDb.NewMongoDb(conn, TestDbName, TestCollection)

	// setup indexes
	assert.NoError(t, db.Setup([]mgo.Index{
		{Key: []string{"name"}},
		{Key: []string{"email"}, Unique: true},
	}))

	// index should be correct set
	idx, err := db.Conn.DB(db.Name).C(db.Collection).Indexes()
	assert.NoError(t, err)

	assert.Equal(t, "email_1", idx[1].Name)
	assert.True(t, idx[1].Unique)
	assert.Equal(t, "name_1", idx[2].Name)
}
