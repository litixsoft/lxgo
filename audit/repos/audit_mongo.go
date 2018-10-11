package lxAuditRepos

import (
	"github.com/globalsign/mgo"
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/db"
	"time"
)

type IAuditLogger interface {
	Errorf(format string, args ...interface{})
}

// auditMongo, mongo repository
type auditMongo struct {
	serviceName string
	serviceHost string
	db          *lxDb.MongoDb
	logger      IAuditLogger
}

// NewAuditMongo, return instance of auditMongo repository
func NewAuditMongo(db *lxDb.MongoDb, logger IAuditLogger, serviceName, serviceHost string) lxAudit.IAudit {
	return &auditMongo{db: db, logger: logger, serviceName: serviceName, serviceHost: serviceHost}
}

// SetupAudit, set the indexes for mongoDb
func (repo *auditMongo) SetupAudit() error {
	// Copy mongo session (thread safe) and close after function
	conn := repo.db.Conn.Copy()
	defer conn.Close()

	// Setup indexes
	return repo.db.Setup([]mgo.Index{
		{Key: []string{"timestamp"}},
	})
}

// Log, save log entry to mongoDb
func (repo *auditMongo) Log(action int, user, message, data interface{}) chan bool {
	// channel for done
	done := make(chan bool, 1)

	go func() {
		// Copy mongo session (thread safe) and close after function
		conn := repo.db.Conn.Copy()
		defer conn.Close()

		// Log entry
		entry := &lxAudit.AuditModel{
			TimeStamp:   time.Now(),
			ServiceName: repo.serviceName,
			ServiceHost: repo.serviceHost,
			Action:      action,
			User:        user,
			Message:     message,
			Data:        data,
		}

		// Insert entry
		if err := conn.DB(repo.db.Name).C(repo.db.Collection).Insert(entry); err != nil {
			// log error
			repo.logger.Errorf("Error by audit log: %v", err)

			// convert entry to json and log
			jEntry := entry.ToJson()
			repo.logger.Errorf("Could not save audit entry: %v", jEntry)
		}

		done <- true
	}()

	return done
}
