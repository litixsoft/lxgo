package lxAuditRepos_test

import (
	"github.com/globalsign/mgo/bson"
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/audit/repos"
	"github.com/litixsoft/lxgo/db"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	TestDbName      = "lxgo_test"
	ServiceName     = "LxGo_Test_Service"
	ServiceHost     = "localhost:3101"
	AuditCollection = "audit"
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

func TestNewAuditMongo(t *testing.T) {
	conn := lxDb.GetMongoDbConnection(dbHost, "", "", "")
	defer conn.Close()

	// Db base repo
	db := lxDb.NewMongoDb(conn, TestDbName, AuditCollection)

	// create audit mongo repo
	repo := lxAuditRepos.NewAuditMongo(db, ServiceName, ServiceHost)

	//  should be *lxAuditRepos.auditMongo type
	chkT := reflect.TypeOf(repo)
	assert.Equal(t, "*lxAuditRepos.auditMongo", chkT.String())
}

func TestAuditMongo_SetupAudit(t *testing.T) {
	conn := lxDb.GetMongoDbConnection(dbHost, "", "", "")
	defer conn.Close()

	// Delete collection
	conn.DB(TestDbName).C(AuditCollection).DropCollection()

	// Test audit entry for create db and check indexes
	testEntry := lxAudit.AuditModel{
		TimeStamp:   time.Now(),
		ServiceName: ServiceName,
		ServiceHost: ServiceHost,
		Action:      lxAudit.Create,
		User:        "TestUser",
		Message:     "TestMessage",
		Data: struct {
			Name  string
			Age   int
			Login time.Time
		}{
			Name:  "Test User",
			Age:   44,
			Login: time.Now(),
		},
	}
	if err := conn.DB(TestDbName).C(AuditCollection).Insert(testEntry); err != nil {
		log.Fatal(err)
	}

	// Db base, repo
	db := lxDb.NewMongoDb(conn, TestDbName, AuditCollection)
	repo := lxAuditRepos.NewAuditMongo(db, "TestService", "localhost:3101")

	t.Run("index before setup", func(t *testing.T) {
		idx, err := conn.DB(TestDbName).C(AuditCollection).Indexes()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(idx))
		assert.Equal(t, "_id_", idx[0].Name)
	})

	t.Run("index setup", func(t *testing.T) {
		// setup audit
		assert.NoError(t, repo.SetupAudit())

		// check indexes
		idx, err := conn.DB(TestDbName).C(AuditCollection).Indexes()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(idx))
		assert.Equal(t, "timestamp_1", idx[1].Name)
	})
}

func TestAuditMongo_Log(t *testing.T) {
	conn := lxDb.GetMongoDbConnection(dbHost, "", "", "")
	defer conn.Close()

	// Delete collection
	conn.DB(TestDbName).C(AuditCollection).DropCollection()

	// Repo and setup
	// Db base, repo
	db := lxDb.NewMongoDb(conn, TestDbName, AuditCollection)
	repo := lxAuditRepos.NewAuditMongo(db, "TestService", "localhost:3101")

	// setup repo
	assert.NoError(t, repo.SetupAudit())

	// log a new entry
	done, errs := repo.Log(lxAudit.Log, "test_user", "a audit message", lxHelper.M{"name": "test_name"})

	// wait for go routines and check
	assert.NoError(t, <-errs)
	assert.True(t, <-done)

	// audit entry should be found in db
	var result lxAudit.AuditModel
	assert.NoError(t, conn.DB(db.Name).C(db.Collection).Find(lxHelper.M{"user": "test_user"}).One(&result))
	assert.Equal(t, "TestService", result.ServiceName)
	assert.Equal(t, "localhost:3101", result.ServiceHost)
	assert.Equal(t, lxAudit.Log, result.Action)
	assert.Equal(t, "test_user", result.User)
	assert.Equal(t, "a audit message", result.Message)
	assert.Equal(t, bson.M{"name": "test_name"}, result.Data)
}
