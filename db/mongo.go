package lxDb

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"log"
	"time"
)

// Db struct for mongodb
type MongoDb struct {
	Conn       *mgo.Session
	Name       string
	Collection string
}

// NewMongoDbConn, get a new db connection
func GetMongoDbConnection(url string) *mgo.Session {
	// dial info
	dialInfo, err := mgo.ParseURL(url)
	if err != nil {
		log.Fatal(err)
	}

	// Connection timeout
	dialInfo.Timeout = 30 * time.Second

	// Create new connection
	conn, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMode(mgo.Primary, true)

	return conn
}

func NewMongoDb(connection *mgo.Session, dbName, collection string) *MongoDb {
	return &MongoDb{
		Conn:       connection,
		Name:       dbName,
		Collection: collection,
	}
}

// Setup create indexes for user collection.
func (db *MongoDb) Setup(indexes []mgo.Index) error {
	// Copy mongo session (thread safe) and close after function
	conn := db.Conn.Copy()
	defer conn.Close()

	// Ensure indexes
	col := conn.DB(db.Name).C(db.Collection)

	for _, i := range indexes {
		if err := col.EnsureIndex(i); err != nil {
			return err
		}
	}

	return nil
}

// AuditLog log db audit
func (db *MongoDb) AuditLog(action string, user, data interface{}) error {
	var auditErr error

	conn := db.Conn.Copy()
	defer conn.Close()

	// New entry
	entry := new(AuditModel)
	entry.TimeStamp = time.Now()
	entry.Collection = db.Collection
	entry.Action = action
	entry.User = user
	entry.Data = data

	if err := conn.DB(db.Name + "_audit").C("audit").Insert(entry); err != nil {
		jEntry := entry.ToJson()
		auditErr = errors.New(fmt.Sprintf("Could not save audit entry: %v", jEntry))
	}

	return auditErr
}

// ToJson, convert model to json string for log
func (am *AuditModel) ToJson() string {
	conEntry := bson.M{
		"timestamp":  bson.M{"$date": am.TimeStamp.UTC().Format(time.RFC3339)},
		"collection": am.Collection,
		"action":     am.Action,
		"user":       am.User,
		"data":       am.Data,
	}

	jentry, _ := json.Marshal(conEntry)

	return string(jentry)
}
